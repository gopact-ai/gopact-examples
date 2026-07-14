package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/gopact-ai/gopact"
	"github.com/gopact-ai/gopact/workflow"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

const (
	sessionID = "session-otel"
	runID     = "run-otel"
)

type spanEventSink struct{}

func (spanEventSink) Emit(ctx context.Context, event gopact.Event) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("gen_ai.conversation.id", event.SessionID),
		attribute.String("gopact.run.id", event.RunID),
		attribute.String("gopact.workflow.name", event.DefinitionID),
	)
	span.AddEvent(event.Type)
	return nil
}

type staticLookup struct{}

func (staticLookup) Lookup(_ context.Context, input string) (string, error) {
	return "observed:" + input, nil
}

// tracedLookup is application-owned infrastructure instrumentation. It wraps
// the adapter call directly instead of manufacturing a Workflow domain event.
type tracedLookup struct {
	next   staticLookup
	tracer trace.Tracer
}

func (adapter tracedLookup) Lookup(ctx context.Context, input string) (string, error) {
	ctx, span := adapter.tracer.Start(ctx, "adapter.lookup")
	defer span.End()
	return adapter.next.Lookup(ctx, input)
}

func runExample(ctx context.Context, provider trace.TracerProvider) (string, []gopact.Event, error) {
	tracer := provider.Tracer("gopact-examples/integrations/otel")
	ctx, span := tracer.Start(ctx, "workflow.run")
	defer span.End()
	lookup := tracedLookup{next: staticLookup{}, tracer: tracer}

	wf := workflow.New[string, string]("observed-workflow")
	observe := wf.Node("observe", func(ctx context.Context, input string) (string, error) {
		return lookup.Lookup(ctx, input)
	})
	wf.Entry(observe)
	wf.Exit(observe)

	var events []gopact.Event
	output, invokeErr := wf.Invoke(
		ctx,
		"input",
		gopact.WithSessionID(sessionID),
		gopact.WithRunID(runID),
		gopact.WithEventSink(spanEventSink{}),
		gopact.WithEventHandler(func(_ context.Context, event gopact.Event) error {
			events = append(events, event)
			return nil
		}),
	)
	if invokeErr != nil {
		return "", nil, fmt.Errorf("invoke workflow: %w", invokeErr)
	}
	return output, events, nil
}

func runProgram(ctx context.Context, provider *sdktrace.TracerProvider, exporter *tracetest.InMemoryExporter) (string, int, bool, error) {
	output, events, runErr := runExample(ctx, provider)
	spans := exporter.GetSpans()
	traceValid := false
	for _, span := range spans {
		if span.Name == "workflow.run" {
			traceValid = span.SpanContext.TraceID().IsValid()
		}
	}
	shutdownErr := provider.Shutdown(context.WithoutCancel(ctx))
	return output, len(events), traceValid, errors.Join(runErr, shutdownErr)
}

func main() {
	ctx := context.Background()
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	output, eventCount, traceValid, err := runProgram(ctx, provider, exporter)
	if err != nil {
		panic(err)
	}
	fmt.Printf("output=%s events=%d trace-valid=%t\n", output, eventCount, traceValid)
}
