package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/gopact-ai/gopact"
	"github.com/gopact-ai/gopact-ext/agents/scheduler"
)

func TestRunShowsBackgroundSchedulerFlow(t *testing.T) {
	var out bytes.Buffer
	if err := run(context.Background(), &out); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"background scheduler: leased worker",
		"drain: dequeued=3 completed=2 retried=1 dead_lettered=0",
		"completed: doc-scan, flaky-index",
		"retry evidence: flaky-index attempt=1 next=2",
		"dead-letter: poison-task action=dead_letter",
		"schedule evidence: passed=1 failed=1 checks=2",
		"lease: released=true",
		"summary: processed=2 retried=1 dead_lettered=1",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output = %q, want %q", got, want)
		}
	}
}

func TestRunDemoCapturesSchedulerStateTransitions(t *testing.T) {
	result, err := runDemo(context.Background())
	if err != nil {
		t.Fatalf("runDemo() error = %v", err)
	}

	if result.Drain.Dequeued != 3 || result.Drain.Completed != 2 || result.Drain.Retried != 1 {
		t.Fatalf("drain = %+v, want completed and retried jobs", result.Drain)
	}
	if len(result.CompletedSnapshot.Completed) != 2 || len(result.CompletedSnapshot.Retried) != 1 {
		t.Fatalf("completed snapshot = %+v, want two completed and one retry", result.CompletedSnapshot)
	}
	retry := result.CompletedSnapshot.Retried[0]
	if retry.Job.ID != "flaky-index" ||
		retry.Decision.Action != scheduler.ScheduleRetry ||
		retry.Decision.Attempt != 1 ||
		retry.Decision.NextAttempt != 2 {
		t.Fatalf("retry record = %+v, want flaky-index attempt 1 -> 2", retry)
	}
	if len(result.DeadLetterSnapshot.DeadLettered) != 1 ||
		result.DeadLetterSnapshot.DeadLettered[0].Job.ID != "poison-task" {
		t.Fatalf("dead-letter snapshot = %+v, want poison-task", result.DeadLetterSnapshot)
	}
	if !result.LeaseReleased {
		t.Fatal("lease released = false, want true")
	}
}

func TestRunDemoRecordsScheduleVerificationEvidence(t *testing.T) {
	result, err := runDemo(context.Background())
	if err != nil {
		t.Fatalf("runDemo() error = %v", err)
	}
	if len(result.Checks) != 2 {
		t.Fatalf("checks = %+v, want retry and dead-letter evidence", result.Checks)
	}

	requireScheduleCheck(t, result.Checks[0], gopact.VerificationStatusPassed, "flaky-index")
	requireScheduleCheck(t, result.Checks[1], gopact.VerificationStatusFailed, "poison-task")
}

func requireScheduleCheck(t *testing.T, check gopact.VerificationCheck, status gopact.VerificationStatus, jobID string) {
	t.Helper()

	if check.Status != status {
		t.Fatalf("check = %+v, want status %s", check, status)
	}
	if len(check.Evidence) != 1 || check.Evidence[0].Type != scheduler.VerificationEvidenceTypeSchedule {
		t.Fatalf("check evidence = %+v, want scheduler evidence", check.Evidence)
	}
	if check.Evidence[0].Metadata["job_id"] != jobID {
		t.Fatalf("evidence metadata = %+v, want job_id %s", check.Evidence[0].Metadata, jobID)
	}
}
