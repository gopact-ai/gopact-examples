package eval_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/gopact-ai/gopact"
	"github.com/gopact-ai/gopact-examples/eval"
	"github.com/gopact-ai/gopact/workflow"
)

// faultResumeScenario is the corpus's first entry and its reference shape: a
// guard interrupts a workflow (a human-in-the-loop pause), the run suspends on
// a durable checkpoint, and a second invocation resumes it to completion. It
// exercises gopact's most load-bearing guarantee — durable recovery across an
// interrupt — end to end, asserting only over the trajectory.
//
// The scenario drives two invocations against one Recorder so the resume event
// lands in the same trajectory as the initial interrupt.
func faultResumeScenario() eval.Scenario {
	const runID = "eval-fault-resume"
	return eval.Scenario{
		Name: "guard interrupt then durable resume",
		Metadata: eval.Metadata{
			Model:   "none",
			Harness: "gopact/workflow",
			Config:  "guard-before-run/approval",
		},
		Run: func(ctx context.Context, rec *eval.Recorder) (eval.Result, error) {
			store := workflow.NewMemoryStore()
			interrupt := true
			build := func() *workflow.Workflow[string, string] {
				wf := workflow.New[string, string]("eval-fault-resume", workflow.WithStore(store))
				process := wf.Node("process", func(_ context.Context, in string) (string, error) {
					return "processed:" + in, nil
				})
				process.Guard(workflow.BeforeRun("approval", workflow.GuardFunc[string, string](
					func(context.Context, workflow.GuardContext[string, string]) (workflow.GuardDecision[string, string], error) {
						if !interrupt {
							return workflow.GuardAllow[string, string]{}, nil
						}
						interrupt = false
						return workflow.GuardInterrupt[string, string]{
							Request: workflow.InterruptRequest{ID: "approval", Subject: "approval"},
						}, nil
					},
				)))
				wf.Entry(process)
				wf.Exit(process)
				return wf
			}

			opts := append([]gopact.RunOption{gopact.WithRunID(runID)}, rec.RunOptions()...)
			_, err := build().Invoke(ctx, "order-42", opts...)
			var interrupted workflow.InterruptError
			if !errors.As(err, &interrupted) {
				return eval.Result{}, fmt.Errorf("initial invoke: got %v, want interrupt", err)
			}

			resumeOpts := append([]gopact.RunOption{
				workflow.WithResume(workflow.ResumeRequest{
					RunID:        runID,
					CheckpointID: interrupted.CheckpointID,
					Resolutions: []workflow.InterruptResolution{{
						InterruptID: "approval",
						PayloadRef:  "artifact://approved",
					}},
				}),
			}, rec.RunOptions()...)
			output, err := build().Invoke(ctx, "", resumeOpts...)
			return eval.Result{Output: output, Err: err}, nil
		},
		Expect: []eval.Matcher{
			eval.GuardInterrupted("approval"),
			eval.Resumed(),
			eval.Completed(),
			eval.FinalText("processed:order-42"),
		},
	}
}
