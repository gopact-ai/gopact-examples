package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/gopact-ai/gopact"
	"github.com/gopact-ai/gopact-examples/internal/exampleenv"
	"github.com/gopact-ai/gopact-ext/agents/scheduler"
)

type demoResult struct {
	Drain              scheduler.DrainResult
	CompletedSnapshot  scheduler.MemoryQueueSnapshot
	DeadLetterResult   scheduler.WorkerResult
	DeadLetterSnapshot scheduler.MemoryQueueSnapshot
	Checks             []gopact.VerificationCheck
	LeaseReleased      bool
}

type zeroDelayDecider struct{}

func main() {
	if err := run(context.Background(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, out io.Writer) error {
	result, err := runDemo(ctx)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintln(out, "background scheduler: leased worker"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "drain: dequeued=%d completed=%d retried=%d dead_lettered=%d\n",
		result.Drain.Dequeued,
		result.Drain.Completed,
		result.Drain.Retried,
		result.Drain.DeadLettered,
	); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "completed: %s\n", completedJobIDs(result.CompletedSnapshot)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "retry evidence: %s\n", retryEvidence(result.CompletedSnapshot)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "dead-letter: %s action=%s\n",
		result.DeadLetterResult.Job.ID,
		result.DeadLetterResult.Decision.Action,
	); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "schedule evidence: %s\n", scheduleEvidenceSummary(result.Checks)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "lease: released=%t\n", result.LeaseReleased); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "summary: processed=%d retried=%d dead_lettered=%d\n",
		len(result.CompletedSnapshot.Completed),
		len(result.CompletedSnapshot.Retried),
		len(result.DeadLetterSnapshot.DeadLettered),
	)
	return err
}

func runDemo(ctx context.Context) (demoResult, error) {
	if err := exampleenv.LoadDotEnv(); err != nil {
		return demoResult{}, err
	}

	leases := gopact.NewMemoryLeaseBackend()
	recorder := gopact.NewVerificationRecorder()
	queue := scheduler.NewMemoryQueue(
		scheduler.Job{ID: "doc-scan", Payload: "scan docs", Attempt: 1, MaxAttempts: 3},
		scheduler.Job{ID: "flaky-index", Payload: "refresh index", Attempt: 1, MaxAttempts: 3},
	)
	seenFlaky := false
	worker, err := scheduler.NewWorker(
		queue,
		scheduler.HandlerFunc(func(_ context.Context, job scheduler.Job) (scheduler.Result, error) {
			if job.ID == "flaky-index" && !seenFlaky {
				seenFlaky = true
				return scheduler.Result{}, errors.New("temporary index failure")
			}
			return scheduler.Result{
				Status: scheduler.JobSucceeded,
				Output: "done " + job.ID,
			}, nil
		}),
		scheduler.WithScheduleDecider(zeroDelayDecider{}),
		scheduler.WithRecorder(recorder),
		scheduler.WithLease(leases, gopact.LeaseRequest{
			Key:   "scheduler/quickstart",
			Owner: "quickstart-worker",
			TTL:   time.Minute,
		}),
	)
	if err != nil {
		return demoResult{}, err
	}
	drain, err := worker.Drain(ctx, 5)
	if err != nil {
		return demoResult{}, err
	}

	deadLetterQueue := scheduler.NewMemoryQueue(
		scheduler.Job{ID: "poison-task", Payload: "cannot recover", Attempt: 2, MaxAttempts: 2},
	)
	deadLetterWorker, err := scheduler.NewWorker(
		deadLetterQueue,
		scheduler.HandlerFunc(func(context.Context, scheduler.Job) (scheduler.Result, error) {
			return scheduler.Result{}, errors.New("permanent failure")
		}),
		scheduler.WithScheduleDecider(zeroDelayDecider{}),
		scheduler.WithRecorder(recorder),
		scheduler.WithLease(leases, gopact.LeaseRequest{
			Key:   "scheduler/quickstart",
			Owner: "quickstart-worker",
			TTL:   time.Minute,
		}),
	)
	if err != nil {
		return demoResult{}, err
	}
	deadLetterResult, err := deadLetterWorker.RunOnce(ctx)
	if !errors.Is(err, scheduler.ErrJobDeadLettered) {
		return demoResult{}, err
	}

	_, leaseActive, err := leases.GetLease(ctx, "scheduler/quickstart")
	if err != nil {
		return demoResult{}, err
	}
	return demoResult{
		Drain:              drain,
		CompletedSnapshot:  queue.Snapshot(),
		DeadLetterResult:   deadLetterResult,
		DeadLetterSnapshot: deadLetterQueue.Snapshot(),
		Checks:             recorder.Checks(),
		LeaseReleased:      !leaseActive,
	}, nil
}

func (zeroDelayDecider) DecideSchedule(_ context.Context, request scheduler.ScheduleRequest) (scheduler.ScheduleDecision, error) {
	maxAttempts := request.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	decision := scheduler.ScheduleDecision{
		Attempt:     request.Attempt,
		MaxAttempts: maxAttempts,
		Reason:      "quickstart scheduler policy",
		Metadata:    map[string]any{"policy": "quickstart"},
	}
	if request.Attempt >= maxAttempts {
		decision.Action = scheduler.ScheduleDeadLetter
		return decision, nil
	}
	decision.Action = scheduler.ScheduleRetry
	decision.NextAttempt = request.Attempt + 1
	return decision, nil
}

func completedJobIDs(snapshot scheduler.MemoryQueueSnapshot) string {
	ids := make([]string, 0, len(snapshot.Completed))
	for _, record := range snapshot.Completed {
		ids = append(ids, record.Job.ID)
	}
	return strings.Join(ids, ", ")
}

func retryEvidence(snapshot scheduler.MemoryQueueSnapshot) string {
	parts := make([]string, 0, len(snapshot.Retried))
	for _, record := range snapshot.Retried {
		parts = append(parts, fmt.Sprintf("%s attempt=%d next=%d",
			record.Job.ID,
			record.Decision.Attempt,
			record.Decision.NextAttempt,
		))
	}
	return strings.Join(parts, ", ")
}

func scheduleEvidenceSummary(checks []gopact.VerificationCheck) string {
	var passed, failed int
	for _, check := range checks {
		switch check.Status {
		case gopact.VerificationStatusPassed:
			passed++
		case gopact.VerificationStatusFailed:
			failed++
		}
	}
	return fmt.Sprintf("passed=%d failed=%d checks=%d", passed, failed, len(checks))
}
