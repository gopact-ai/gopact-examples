package main

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/gopact-ai/gopact"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// TestRunExampleExportsWorkflowIdentity verifies that workflow identity and events reach the caller-owned span.
func TestRunExampleExportsWorkflowIdentity(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() {
		if err := provider.Shutdown(context.Background()); err != nil {
			t.Errorf("shutdown tracer provider: %v", err)
		}
	})

	output, events, err := runExample(t.Context(), provider)
	if err != nil {
		t.Fatalf("runExample() error = %v", err)
	}
	if output != "observed:input" {
		t.Fatalf("output = %q, want observed:input", output)
	}
	if len(events) == 0 {
		t.Fatal("events = none, want workflow events")
	}
	for _, event := range events {
		if event.SessionID != "session-otel" || event.RunID != "run-otel" {
			t.Fatalf("event identity = %q/%q, want session-otel/run-otel", event.SessionID, event.RunID)
		}
	}

	// Inspect the exported span instead of relying on global tracer state.
	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("exported spans = %d, want 1", len(spans))
	}
	span := spans[0]
	if traceID := span.SpanContext.TraceID(); !traceID.IsValid() {
		t.Fatalf("trace ID = %s, want valid non-zero ID", traceID)
	}

	attributes := make(map[string]string, len(span.Attributes))
	for _, attr := range span.Attributes {
		attributes[string(attr.Key)] = attr.Value.AsString()
	}
	for key, want := range map[string]string{
		"gen_ai.conversation.id": "session-otel",
		"gopact.run.id":          "run-otel",
		"gopact.workflow.name":   "observed-workflow",
	} {
		if got := attributes[key]; got != want {
			t.Fatalf("span attribute %q = %q, want %q", key, got, want)
		}
	}
	if len(span.Events) != len(events) {
		t.Fatalf("span events = %d, workflow events = %d", len(span.Events), len(events))
	}
	for i, event := range events {
		if got := span.Events[i].Name; got != event.Type {
			t.Fatalf("span event %d name = %q, want %q", i, got, event.Type)
		}
	}
}

// TestRunProgramShutsDownProviderAfterWorkflowFailure verifies that workflow and provider shutdown errors are both retained.
func TestRunProgramShutsDownProviderAfterWorkflowFailure(t *testing.T) {
	// Cancel the workflow while the exporter contributes an independent shutdown failure.
	shutdownErr := errors.New("shutdown exporter")
	memory := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(shutdownErrorExporter{
		InMemoryExporter: memory,
		err:              shutdownErr,
	}))
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, _, _, err := runProgram(ctx, provider, memory)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("runProgram() error = %v, want context.Canceled", err)
	}
	if !errors.Is(err, shutdownErr) {
		t.Fatalf("runProgram() error = %v, want shutdown error", err)
	}
	if spans := memory.GetSpans(); len(spans) != 0 {
		t.Fatalf("exporter spans after shutdown = %d, want 0", len(spans))
	}
}

type shutdownErrorExporter struct {
	*tracetest.InMemoryExporter
	err error
}

func (e shutdownErrorExporter) Shutdown(ctx context.Context) error {
	return errors.Join(e.InMemoryExporter.Shutdown(ctx), e.err)
}

// TestEventEnvelopeDoesNotPersistTelemetryTraceIdentity verifies that telemetry trace IDs stay out of durable workflow events.
func TestEventEnvelopeDoesNotPersistTelemetryTraceIdentity(t *testing.T) {
	eventType := reflect.TypeFor[gopact.Event]()
	for _, field := range []string{"TraceID", "SpanID"} {
		if _, ok := eventType.FieldByName(field); ok {
			t.Fatalf("gopact.Event has telemetry field %q", field)
		}
	}
}
