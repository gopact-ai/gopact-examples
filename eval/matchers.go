package eval

import (
	"fmt"
	"strings"

	"github.com/gopact-ai/gopact"
	"github.com/gopact-ai/gopact/workflow"
)

// Matcher is one deterministic assertion over a captured [Trajectory]. It
// returns nil when the expectation holds, or an error describing the mismatch.
//
// Built-in matchers never call a model and never depend on wall-clock time or
// map iteration order, so a suite built from them is reproducible. Advanced
// users implement Matcher (or [MatcherFunc]) to assert over the raw event
// slices on Trajectory.
type Matcher interface {
	// Match reports whether the trajectory satisfies the expectation.
	Match(Trajectory) error
	// Describe returns a short human-readable form of the expectation, used in
	// failure messages and suite output.
	Describe() string
}

// MatcherFunc adapts a function and a description into a [Matcher].
type MatcherFunc struct {
	Desc string
	Fn   func(Trajectory) error
}

// Match implements [Matcher].
func (m MatcherFunc) Match(tr Trajectory) error { return m.Fn(tr) }

// Describe implements [Matcher].
func (m MatcherFunc) Describe() string { return m.Desc }

// Completed asserts the run reached workflow completion with no terminal error.
func Completed() Matcher {
	return MatcherFunc{
		Desc: "run completed",
		Fn: func(tr Trajectory) error {
			if tr.Err != nil {
				return fmt.Errorf("run returned error: %w", tr.Err)
			}
			if !hasLifecycle(tr, workflow.EventWorkflowCompleted) {
				return fmt.Errorf("no %s event in trajectory", workflow.EventWorkflowCompleted)
			}
			return nil
		},
	}
}

// NoError asserts the run returned no terminal error. Use it for scenarios that
// interrupt rather than complete, where [Completed] would not hold.
func NoError() Matcher {
	return MatcherFunc{
		Desc: "no terminal error",
		Fn: func(tr Trajectory) error {
			if tr.Err != nil {
				return fmt.Errorf("run returned error: %w", tr.Err)
			}
			return nil
		},
	}
}

// Failed asserts the run failed and, when want is non-empty, that the terminal
// error text contains want.
func Failed(want string) Matcher {
	desc := "run failed"
	if want != "" {
		desc = fmt.Sprintf("run failed with %q", want)
	}
	return MatcherFunc{
		Desc: desc,
		Fn: func(tr Trajectory) error {
			if tr.Err == nil {
				return fmt.Errorf("run succeeded, want failure")
			}
			if want != "" && !strings.Contains(tr.Err.Error(), want) {
				return fmt.Errorf("error %q does not contain %q", tr.Err.Error(), want)
			}
			return nil
		},
	}
}

// Interrupted asserts the run suspended on an interrupt (a durable-resume
// boundary) rather than completing.
func Interrupted() Matcher {
	return MatcherFunc{
		Desc: "run interrupted",
		Fn: func(tr Trajectory) error {
			if hasLifecycle(tr, workflow.EventWorkflowInterrupted) ||
				hasLifecycle(tr, workflow.EventGuardInterrupted) {
				return nil
			}
			return fmt.Errorf("no interrupt event in trajectory")
		},
	}
}

// Resumed asserts the run resumed from a checkpoint, i.e. it recovered rather
// than starting fresh. This is the signal that gopact's durable-lifecycle
// guarantee actually fired for the scenario.
func Resumed() Matcher {
	return MatcherFunc{
		Desc: "run resumed from checkpoint",
		Fn: func(tr Trajectory) error {
			if hasLifecycle(tr, workflow.EventWorkflowResumed) {
				return nil
			}
			return fmt.Errorf("no %s event in trajectory", workflow.EventWorkflowResumed)
		},
	}
}

// GuardInterrupted asserts a guard raised an interrupt (a HITL pause). When id
// is non-empty it must be contained in the interrupt's subject.
func GuardInterrupted(id string) Matcher {
	desc := "guard interrupted"
	if id != "" {
		desc = fmt.Sprintf("guard interrupted %q", id)
	}
	return MatcherFunc{
		Desc: desc,
		Fn: func(tr Trajectory) error {
			for _, ev := range tr.Lifecycle {
				if ev.Type != workflow.EventGuardInterrupted {
					continue
				}
				if id == "" || strings.Contains(ev.Summary, id) {
					return nil
				}
			}
			return fmt.Errorf("no guard interrupt matching %q", id)
		},
	}
}

// GuardDenied asserts a guard rejected an action. When id is non-empty it must
// appear in the rejection's summary.
func GuardDenied(id string) Matcher {
	desc := "guard denied"
	if id != "" {
		desc = fmt.Sprintf("guard denied %q", id)
	}
	return MatcherFunc{
		Desc: desc,
		Fn: func(tr Trajectory) error {
			for _, ev := range tr.Lifecycle {
				if ev.Type != workflow.EventGuardRejected {
					continue
				}
				if id == "" || strings.Contains(ev.Summary, id) {
					return nil
				}
			}
			return fmt.Errorf("no guard rejection matching %q", id)
		},
	}
}

// ToolCalled asserts the named tool produced at least one finished outcome.
func ToolCalled(name string) Matcher {
	return MatcherFunc{
		Desc: fmt.Sprintf("tool %q called", name),
		Fn: func(tr Trajectory) error {
			for _, step := range tr.Tool {
				if step.Type == gopact.ToolEventCallFinished && step.Name == name {
					return nil
				}
			}
			return fmt.Errorf("tool %q was not called", name)
		},
	}
}

// ToolOrder asserts the named tools finished in exactly the given relative
// order. Tools outside names are ignored; each name must appear. Ordering
// follows the transcript the harness assembled, not tool completion order, so
// this stays deterministic even when tools finish out of order.
func ToolOrder(names ...string) Matcher {
	return MatcherFunc{
		Desc: fmt.Sprintf("tools finish in order %v", names),
		Fn: func(tr Trajectory) error {
			want := make(map[string]struct{}, len(names))
			for _, n := range names {
				want[n] = struct{}{}
			}
			var got []string
			for _, step := range tr.Tool {
				if step.Type != gopact.ToolEventCallFinished {
					continue
				}
				if _, ok := want[step.Name]; ok {
					got = append(got, step.Name)
				}
			}
			if len(got) != len(names) {
				return fmt.Errorf("tool order = %v, want %v", got, names)
			}
			for i, n := range names {
				if got[i] != n {
					return fmt.Errorf("tool order = %v, want %v", got, names)
				}
			}
			return nil
		},
	}
}

// FinalText asserts the run's final output text equals want.
func FinalText(want string) Matcher {
	return MatcherFunc{
		Desc: fmt.Sprintf("final output = %q", want),
		Fn: func(tr Trajectory) error {
			if got := finalText(tr); got != want {
				return fmt.Errorf("final output = %q, want %q", got, want)
			}
			return nil
		},
	}
}

// FinalContains asserts the run's final output text contains want.
func FinalContains(want string) Matcher {
	return MatcherFunc{
		Desc: fmt.Sprintf("final output contains %q", want),
		Fn: func(tr Trajectory) error {
			if got := finalText(tr); !strings.Contains(got, want) {
				return fmt.Errorf("final output %q does not contain %q", got, want)
			}
			return nil
		},
	}
}

// ModelTurns asserts the model was invoked exactly n times (n finished
// model calls), a proxy for reasoning-loop length.
func ModelTurns(n int) Matcher {
	return MatcherFunc{
		Desc: fmt.Sprintf("model invoked %d times", n),
		Fn: func(tr Trajectory) error {
			got := 0
			for _, step := range tr.Model {
				if step.Type == gopact.ModelEventCallFinished {
					got++
				}
			}
			if got != n {
				return fmt.Errorf("model turns = %d, want %d", got, n)
			}
			return nil
		},
	}
}

func hasLifecycle(tr Trajectory, eventType string) bool {
	for _, ev := range tr.Lifecycle {
		if ev.Type == eventType {
			return true
		}
	}
	return false
}

func finalText(tr Trajectory) string {
	if tr.Response == nil {
		return tr.Output
	}
	if text := messageText(*tr.Response); text != "" {
		return text
	}
	return tr.Output
}

func messageText(msg gopact.Message) string {
	var b strings.Builder
	for _, part := range msg.Parts {
		if part.Type == gopact.MessagePartTypeText {
			b.WriteString(part.Text)
		}
	}
	return b.String()
}
