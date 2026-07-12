package main

import (
	"context"
	"testing"

	"github.com/gopact-ai/gopact/workflow"
)

func TestRunExample(t *testing.T) {
	result, err := runExample(context.Background())
	if err != nil {
		t.Fatalf("runExample() error = %v", err)
	}
	if result.sessionID != "customer-case-42" {
		t.Fatalf("session ID = %q, want customer-case-42", result.sessionID)
	}
	if len(result.beforeRuns) != 2 || len(result.afterRuns) != 2 {
		t.Fatalf("run counts before/after = %d/%d, want 2/2", len(result.beforeRuns), len(result.afterRuns))
	}
	if result.beforeRuns[0].SessionID != result.sessionID || result.beforeRuns[0].RunID != "run-intake" ||
		result.beforeRuns[0].DefinitionID != "intake" ||
		result.beforeRuns[0].Status != workflow.CheckpointCompleted {
		t.Fatalf("first run before resume = %+v, want completed intake", result.beforeRuns[0])
	}
	if result.beforeRuns[1].SessionID != result.sessionID || result.beforeRuns[1].RunID != "run-review" ||
		result.beforeRuns[1].DefinitionID != "review" ||
		result.beforeRuns[1].Status != workflow.CheckpointInterrupted {
		t.Fatalf("second run before resume = %+v, want interrupted review", result.beforeRuns[1])
	}
	if result.snapshot.RunMeta.RunID != "run-review" || result.snapshot.RunMeta.SessionID != result.sessionID {
		t.Fatalf("snapshot run metadata = %+v, want run-review/%s", result.snapshot.RunMeta, result.sessionID)
	}
	if len(result.snapshot.Timeline) == 0 || len(result.snapshot.Checkpoints) == 0 {
		t.Fatalf("snapshot timeline/checkpoints = %d/%d, want both non-empty", len(result.snapshot.Timeline), len(result.snapshot.Checkpoints))
	}
	if result.afterRuns[0].RunID != "run-intake" || result.afterRuns[0].Status != workflow.CheckpointCompleted {
		t.Fatalf("first run after resume = %+v, want completed intake", result.afterRuns[0])
	}
	if result.afterRuns[1].RunID != "run-review" || result.afterRuns[1].DefinitionID != "review" ||
		result.afterRuns[1].Status != workflow.CheckpointCompleted {
		t.Fatalf("second run after resume = %+v, want completed review", result.afterRuns[1])
	}
	if result.output != "approved:request" {
		t.Fatalf("output = %q, want approved:request", result.output)
	}
}
