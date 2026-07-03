package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRunBuildsReleaseBundleWithCoreCLI(t *testing.T) {
	var out bytes.Buffer
	if err := run(context.Background(), &out); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"release-bundle: core CLI",
		"report: passed",
		"bundle: completed verification_reports=1",
		"gate: passed",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output = %q, want %q", got, want)
		}
	}
}
