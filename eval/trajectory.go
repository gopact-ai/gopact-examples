// Package eval is a trajectory-based evaluation harness for gopact agents and
// workflows.
//
// The evaluation ground truth is the run trajectory: the ordered stream of
// lifecycle, model, and tool events a run emits. gopact produces this stream
// deterministically and replayably, so eval asserts against it rather than
// against final assistant text (which drifts) or an LLM judge (which is
// non-deterministic by default).
//
// # Two tiers
//
// Tier 1 — out of the box. Describe a [Scenario], list declarative [Matcher]
// expectations over its trajectory, and run the table with [Run] or [RunSuite]
// against any *testing.T. Every built-in matcher is deterministic, so a green
// suite stays green without a live model.
//
// Tier 2 — advanced. The captured [Trajectory] and the [Recorder] that builds
// it are exported. Compose custom matchers over the raw event slices, plug an
// LLM-as-judge or a task-reward function through the [Judge] interface, or
// generate scenarios programmatically. Judges are opt-in; nothing in the
// default path is non-deterministic.
package eval

import (
	"context"
	"sync"

	"github.com/gopact-ai/gopact"
)

// EventClass partitions a trajectory into the three event streams a gopact run
// emits so matchers can scope their assertions.
type EventClass string

// Event classes.
const (
	// ClassLifecycle marks a workflow lifecycle event (node/guard/workflow).
	ClassLifecycle EventClass = "lifecycle"
	// ClassModel marks a model-call event.
	ClassModel EventClass = "model"
	// ClassTool marks a tool-call event.
	ClassTool EventClass = "tool"
)

// LifecycleEvent is one captured workflow process event. It preserves the
// stable fields matchers assert on without retaining the full event payload.
type LifecycleEvent struct {
	Sequence int64
	Type     string
	NodeID   string
	RunID    string
	Summary  string
}

// ModelStep is one captured model-call event.
type ModelStep struct {
	Type     gopact.ModelEventType
	Request  *gopact.ModelRequest
	Response *gopact.ModelResponse
	Err      error
}

// ToolStep is one captured tool-call event.
type ToolStep struct {
	Type    gopact.ToolEventType
	CallID  string
	Name    string
	Outcome gopact.ToolOutcome
	Err     error
}

// Trajectory is the ordered record of everything one run emitted, plus its
// final result. It is the ground truth every [Matcher] asserts against.
//
// Lifecycle, Model, and Tool hold their respective streams in capture order.
// Lifecycle events carry a monotonic Sequence; component events are ordered by
// arrival, which for a single run is their emission order.
type Trajectory struct {
	// RunID is the run these events belong to.
	RunID string
	// Lifecycle holds workflow process events in sequence order.
	Lifecycle []LifecycleEvent
	// Model holds model-call events in emission order.
	Model []ModelStep
	// Tool holds tool-call events in emission order.
	Tool []ToolStep
	// Response is the agent response for an agent run, when the SUT returned one.
	Response *gopact.Message
	// Output is the final output rendered to text, when available.
	Output string
	// Err is the terminal error the run returned, if any.
	Err error
	// Metadata records the model, harness, and config versions this trajectory
	// was produced under so a corpus can be replayed across versions and used
	// for regression detection.
	Metadata Metadata
}

// Metadata pins the versions a trajectory was produced under. Recording it on
// every trajectory is what lets one corpus run across model/harness/config
// revisions and surface regressions.
type Metadata struct {
	Model   string
	Harness string
	Config  string
	Extra   map[string]string
}

// Recorder captures a run's three event streams into a single ordered
// [Trajectory]. A Recorder satisfies gopact's EventSink, ModelEventSink, and
// ToolEventSink at once, so one value attached with [Recorder.RunOptions]
// captures the whole run.
//
// A Recorder is safe for concurrent event emission: parallel tool nodes may
// emit simultaneously. Read [Recorder.Trajectory] only after the run returns.
type Recorder struct {
	mu        sync.Mutex
	runID     string
	lifecycle []LifecycleEvent
	model     []ModelStep
	tool      []ToolStep
	metadata  Metadata
}

// NewRecorder creates a Recorder that stamps captured trajectories with meta.
func NewRecorder(meta Metadata) *Recorder {
	return &Recorder{metadata: meta}
}

// Emit implements gopact.EventSink for workflow lifecycle events.
func (r *Recorder) Emit(_ context.Context, event gopact.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.runID == "" {
		r.runID = event.RunID
	}
	r.lifecycle = append(r.lifecycle, LifecycleEvent{
		Sequence: event.Sequence,
		Type:     event.Type,
		NodeID:   event.NodeID,
		RunID:    event.RunID,
		Summary:  event.Summary,
	})
	return nil
}

// EmitModelEvent implements gopact.ModelEventSink.
func (r *Recorder) EmitModelEvent(_ context.Context, event gopact.ModelEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.model = append(r.model, ModelStep{
		Type:     event.Type,
		Request:  event.Request,
		Response: event.Response,
		Err:      event.Err,
	})
	return nil
}

// EmitToolEvent implements gopact.ToolEventSink.
func (r *Recorder) EmitToolEvent(_ context.Context, event gopact.ToolEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	step := ToolStep{Type: event.Type, CallID: event.Call.ID, Name: event.Call.Name, Outcome: event.Outcome, Err: event.Err}
	if event.Outcome != nil {
		if id := event.Outcome.ToolCallID(); id != "" {
			step.CallID = id
		}
		if step.Name == "" {
			step.Name = event.Outcome.ToolName()
		}
	}
	r.tool = append(r.tool, step)
	return nil
}

// RunOptions returns the gopact.RunOption values that attach this Recorder to a
// run so it captures the lifecycle, model, and tool streams together. Pass the
// result to Invoke alongside gopact.WithRunID.
func (r *Recorder) RunOptions() []gopact.RunOption {
	return []gopact.RunOption{gopact.WithEventSink(r)}
}

// Trajectory returns the captured trajectory. Call it after the run returns.
// The caller supplies the terminal result; capture cannot observe the return
// value of the SUT's Invoke.
func (r *Recorder) Trajectory() Trajectory {
	r.mu.Lock()
	defer r.mu.Unlock()
	return Trajectory{
		RunID:     r.runID,
		Lifecycle: append([]LifecycleEvent(nil), r.lifecycle...),
		Model:     append([]ModelStep(nil), r.model...),
		Tool:      append([]ToolStep(nil), r.tool...),
		Metadata:  r.metadata,
	}
}
