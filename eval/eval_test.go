package eval_test

import (
	"context"
	"testing"

	"github.com/gopact-ai/gopact"
	"github.com/gopact-ai/gopact-examples/eval"
	"github.com/gopact-ai/gopact-ext/agents/react"
	"github.com/gopact-ai/gopact-ext/models/fake"
	"github.com/gopact-ai/gopact/agent"
	"github.com/gopact-ai/gopact/workflow"
)

// TestCorpus runs the shipped scenario corpus. These are the out-of-the-box
// cases; each drives a real gopact run and asserts only over its trajectory.
func TestCorpus(t *testing.T) {
	eval.RunSuite(t,
		faultResumeScenario(),
		reactFinalScenario(),
	)
}

// reactFinalScenario proves the same Recorder captures an agent SUT: a ReAct
// agent backed by a fake model runs to a final answer, and the trajectory
// carries the agent's lifecycle plus its final response.
func reactFinalScenario() eval.Scenario {
	return eval.Scenario{
		Name: "react agent reaches final answer",
		Metadata: eval.Metadata{
			Model:   "fake",
			Harness: "gopact-ext/agents/react",
			Config:  "no-tools",
		},
		Run: func(ctx context.Context, rec *eval.Recorder) (eval.Result, error) {
			target, err := react.New(
				agent.Identity{Name: "eval-react", Description: "eval", Version: "v1"},
				fake.New(fake.WithResponse("done")),
			)
			if err != nil {
				return eval.Result{}, err
			}
			opts := append([]gopact.RunOption{gopact.WithRunID("eval-react-final")}, rec.RunOptions()...)
			resp, err := target.Invoke(ctx, agent.Request{
				Messages: []gopact.Message{gopact.UserMessage("work")},
			}, opts...)
			return eval.Result{Response: &resp.Message, Err: err}, nil
		},
		Expect: []eval.Matcher{
			eval.Completed(),
			eval.NoError(),
			eval.ModelTurns(1),
			eval.FinalText("done"),
		},
	}
}

// TestEvaluateReturnsTrajectory shows the Tier-2 path: drive a scenario and
// inspect the raw trajectory without matcher assertions.
func TestEvaluateReturnsTrajectory(t *testing.T) {
	tr, err := eval.Evaluate(t.Context(), faultResumeScenario())
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if tr.RunID != "eval-fault-resume" {
		t.Fatalf("trajectory RunID = %q, want eval-fault-resume", tr.RunID)
	}
	if len(tr.Lifecycle) == 0 {
		t.Fatal("trajectory has no lifecycle events")
	}
	if tr.Metadata.Harness != "gopact/workflow" {
		t.Fatalf("trajectory metadata harness = %q, want gopact/workflow", tr.Metadata.Harness)
	}
}

// The remaining tests exercise matcher logic over synthetic trajectories so
// tool-order, guard, and failure matchers are covered without a scripted model.

func lifecycle(types ...string) []eval.LifecycleEvent {
	events := make([]eval.LifecycleEvent, len(types))
	for i, ty := range types {
		events[i] = eval.LifecycleEvent{Sequence: int64(i + 1), Type: ty}
	}
	return events
}

func toolFinished(names ...string) []eval.ToolStep {
	steps := make([]eval.ToolStep, len(names))
	for i, name := range names {
		steps[i] = eval.ToolStep{Type: gopact.ToolEventCallFinished, CallID: name, Name: name}
	}
	return steps
}

func TestToolOrderFollowsTranscriptNotCompletion(t *testing.T) {
	// Transcript records a, b, c in declaration order even if they finished
	// c, b, a: the matcher asserts the recorded order.
	tr := eval.Trajectory{Tool: toolFinished("search", "read", "write")}
	if err := eval.ToolOrder("search", "read", "write").Match(tr); err != nil {
		t.Fatalf("ToolOrder in-order = %v, want nil", err)
	}
	if err := eval.ToolOrder("read", "search").Match(tr); err == nil {
		t.Fatal("ToolOrder wrong-order = nil, want mismatch")
	}
	if err := eval.ToolCalled("search").Match(tr); err != nil {
		t.Fatalf("ToolCalled(search) = %v, want nil", err)
	}
	if err := eval.ToolCalled("absent").Match(tr); err == nil {
		t.Fatal("ToolCalled(absent) = nil, want mismatch")
	}
}

func TestGuardMatchers(t *testing.T) {
	denied := eval.Trajectory{Lifecycle: []eval.LifecycleEvent{
		{Type: workflow.EventGuardRejected, Summary: "budget-guard blocked spend"},
	}}
	if err := eval.GuardDenied("budget-guard").Match(denied); err != nil {
		t.Fatalf("GuardDenied match = %v, want nil", err)
	}
	if err := eval.GuardDenied("other").Match(denied); err == nil {
		t.Fatal("GuardDenied(other) = nil, want mismatch")
	}
	if err := eval.GuardInterrupted("").Match(denied); err == nil {
		t.Fatal("GuardInterrupted over a rejection = nil, want mismatch")
	}
}

func TestFailedMatcher(t *testing.T) {
	failed := eval.Trajectory{Err: context.DeadlineExceeded}
	if err := eval.Failed("deadline").Match(failed); err != nil {
		t.Fatalf("Failed(deadline) = %v, want nil", err)
	}
	if err := eval.Failed("").Match(failed); err != nil {
		t.Fatalf("Failed() = %v, want nil", err)
	}
	if err := eval.Failed("").Match(eval.Trajectory{}); err == nil {
		t.Fatal("Failed() over success = nil, want mismatch")
	}
	if err := eval.Completed().Match(failed); err == nil {
		t.Fatal("Completed() over failure = nil, want mismatch")
	}
}

func TestCompletedRequiresCompletionEvent(t *testing.T) {
	tr := eval.Trajectory{Lifecycle: lifecycle(workflow.EventWorkflowStarted)}
	if err := eval.Completed().Match(tr); err == nil {
		t.Fatal("Completed() without completion event = nil, want mismatch")
	}
	tr.Lifecycle = append(tr.Lifecycle, eval.LifecycleEvent{Type: workflow.EventWorkflowCompleted})
	if err := eval.Completed().Match(tr); err != nil {
		t.Fatalf("Completed() with completion event = %v, want nil", err)
	}
}

// recordingJudge is a deterministic Judge used to prove the extension point
// runs and can fail a scenario.
type recordingJudge struct {
	pass bool
}

func (j recordingJudge) Name() string { return "recording" }

func (j recordingJudge) Judge(_ context.Context, _ eval.Trajectory) (eval.Verdict, error) {
	return eval.Verdict{Pass: j.pass, Score: 1, Reason: "deterministic stub"}, nil
}

func TestJudgeRunsInScenario(t *testing.T) {
	sc := reactFinalScenario()
	sc.Judges = []eval.Judge{recordingJudge{pass: true}}
	eval.Run(t, sc)
}
