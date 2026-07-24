package eval

import (
	"context"
	"testing"

	"github.com/gopact-ai/gopact"
)

// Result is the terminal outcome a scenario's run produced. The runner folds it
// into the captured [Trajectory] before matchers see it.
//
// Err is the run's own terminal error and is part of the trajectory: a scenario
// that asserts [Interrupted] or [Failed] expects it to be non-nil. It is not a
// harness failure. Report harness failures (bad setup, missing fixtures) as the
// second return value of [Scenario.Run] instead.
type Result struct {
	// Response is the agent response message for an agent run, if any.
	Response *gopact.Message
	// Output is the final output rendered to text, if any.
	Output string
	// Err is the run's terminal error, if any.
	Err error
}

// Scenario is one evaluation case: a run to drive and the expectations its
// trajectory must satisfy.
//
// Run receives a [Recorder] and is responsible for invoking the system under
// test with rec.RunOptions() attached (and, for a multi-step scenario such as
// interrupt-then-resume, for driving every invocation against the same
// recorder). It returns the terminal [Result] and, separately, any harness
// error that should fail the scenario outright.
type Scenario struct {
	// Name identifies the scenario in suite output; it becomes the subtest name.
	Name string
	// Metadata pins the model, harness, and config versions this case runs
	// under. It is stamped onto the captured trajectory for regression tracking.
	Metadata Metadata
	// Run drives the system under test and returns its terminal result.
	Run func(ctx context.Context, rec *Recorder) (Result, error)
	// Expect lists the deterministic matchers the trajectory must satisfy.
	Expect []Matcher
	// Judges lists optional, possibly non-deterministic evaluators (LLM-as-judge,
	// task-reward functions). They run only when set; the default path stays
	// deterministic.
	Judges []Judge
}

// Judge is an opt-in, possibly non-deterministic evaluator of a trajectory —
// an LLM-as-judge, a task-reward function that inspects environment terminal
// state, or any scorer that a deterministic [Matcher] cannot express. Judges
// run only when a [Scenario] lists them.
type Judge interface {
	// Judge scores the trajectory. A non-nil error fails the scenario; the
	// score is recorded for reporting and thresholding by the caller.
	Judge(ctx context.Context, tr Trajectory) (Verdict, error)
	// Name identifies the judge in output.
	Name() string
}

// Verdict is a judge's structured decision.
type Verdict struct {
	// Pass reports whether the judge accepted the trajectory.
	Pass bool
	// Score is an optional numeric score in [0,1] for ranking or thresholding.
	Score float64
	// Reason explains the verdict for reporting.
	Reason string
}

// Evaluate drives one scenario and returns its captured [Trajectory] without
// asserting anything. Advanced callers use it to run custom checks; [Run] wraps
// it with matcher and judge assertions against a *testing.T.
func Evaluate(ctx context.Context, sc Scenario) (Trajectory, error) {
	rec := NewRecorder(sc.Metadata)
	if sc.Run == nil {
		return Trajectory{}, errNoRun
	}
	result, err := sc.Run(ctx, rec)
	if err != nil {
		return Trajectory{}, err
	}
	tr := rec.Trajectory()
	tr.Response = result.Response
	tr.Output = result.Output
	tr.Err = result.Err
	return tr, nil
}

// Run drives one scenario against t, applying every matcher and judge. Matcher
// failures and judge failures are reported with t.Errorf so all expectations
// are checked; a harness error from the run is fatal.
func Run(t *testing.T, sc Scenario) Trajectory {
	t.Helper()
	tr, err := Evaluate(t.Context(), sc)
	if err != nil {
		t.Fatalf("scenario %q: run: %v", sc.Name, err)
	}
	for _, matcher := range sc.Expect {
		if matcher == nil {
			continue
		}
		if err := matcher.Match(tr); err != nil {
			t.Errorf("scenario %q: expect [%s]: %v", sc.Name, matcher.Describe(), err)
		}
	}
	for _, judge := range sc.Judges {
		if judge == nil {
			continue
		}
		verdict, err := judge.Judge(t.Context(), tr)
		if err != nil {
			t.Errorf("scenario %q: judge %q: %v", sc.Name, judge.Name(), err)
			continue
		}
		if !verdict.Pass {
			t.Errorf("scenario %q: judge %q failed (score %.3f): %s",
				sc.Name, judge.Name(), verdict.Score, verdict.Reason)
		}
	}
	return tr
}

// RunSuite runs each scenario as a subtest named by its Name.
func RunSuite(t *testing.T, scenarios ...Scenario) {
	t.Helper()
	for _, sc := range scenarios {
		sc := sc
		t.Run(sc.Name, func(t *testing.T) {
			Run(t, sc)
		})
	}
}

// errNoRun is returned by Evaluate when a scenario has no Run function.
var errNoRun = &scenarioError{"scenario has no Run function"}

type scenarioError struct{ msg string }

func (e *scenarioError) Error() string { return e.msg }
