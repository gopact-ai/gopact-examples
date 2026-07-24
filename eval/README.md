# eval — trajectory-based evaluation for gopact runs

<!-- gopact:doc-language: en -->

A small, dependency-light library for evaluating what a gopact run *did*, not what
a model *said*. It is a separate Go module: it pins released `gopact` and
`gopact-ext` tags so the evaluator stays independent of the system under test and
can be run across versions for regression.

## What it evaluates

The single ground truth is the **trajectory** — the deterministic, replayable event
stream a run emits (workflow lifecycle + model + tool events), plus the terminal
result. It does not assert over prompt or completion text, and it does not call an
LLM judge by default. This matches the mainstream consensus (tau-bench, Inspect):
score the *outcome and the path taken*, because that is what is reproducible.

Because `react.Agent` is workflow-backed, one `Recorder` captures both a raw
workflow and an agent through the same `gopact.WithEventSink` seam — the corpus
treats both kinds of system under test identically.

## Tier 1 — out of the box

Declare a `Scenario`, drive the run, assert with deterministic `Matcher`s. It runs
as a plain `go test` — there is no YAML DSL and no standalone CLI to learn.

```go
func TestCorpus(t *testing.T) {
    eval.RunSuite(t,
        faultResumeScenario(),
        reactFinalScenario(),
    )
}
```

A scenario's `Run` invokes the system under test with `rec.RunOptions()` attached
and returns the terminal `Result`; the runner folds that into the captured
trajectory and applies every `Expect` matcher via `t.Errorf`.

Built-in matchers (all deterministic — no model, no wall-clock, no map-order
dependence):

| Matcher | Asserts |
| --- | --- |
| `Completed()` | reached workflow completion with no terminal error |
| `NoError()` | no terminal error (use for interrupt scenarios) |
| `Failed(want)` | failed; terminal error text contains `want` |
| `Interrupted()` | suspended on an interrupt (durable-resume boundary) |
| `Resumed()` | recovered from a checkpoint — gopact's durability actually fired |
| `GuardInterrupted(id)` | a guard raised a HITL pause matching `id` |
| `GuardDenied(id)` | a guard rejected an action matching `id` |
| `ToolCalled(name)` | the named tool produced a finished outcome |
| `ToolOrder(names...)` | named tools finished in this relative transcript order |
| `FinalText(want)` / `FinalContains(want)` | final output text |
| `ModelTurns(n)` | model was invoked exactly `n` times |

## Tier 2 — for advanced users

Everything Tier 1 stands on is exported:

- `Evaluate(ctx, sc)` returns the raw `Trajectory` (lifecycle / model / tool event
  slices + result) with no assertions — inspect or diff it however you like.
- `Matcher` / `MatcherFunc` — write custom assertions over the raw event slices;
  compose them with the built-ins.
- `Judge` — an opt-in, possibly non-deterministic evaluator (LLM-as-judge, a
  task-reward function over environment terminal state). Judges run only when a
  scenario lists them, so the default path stays deterministic.
- `Scenario` is a plain struct, so a corpus can be generated programmatically.

```go
tr, err := eval.Evaluate(ctx, sc)
// tr.Lifecycle, tr.Model, tr.Tool, tr.Response, tr.Err, tr.Metadata ...
```

## Metadata and regression

Every scenario stamps `Metadata{Model, Harness, Config}` onto its trajectory so a
result is attributable to a specific model + harness + config triple. Combined with
the pinned dependency tags, this is what lets the same corpus run against multiple
released versions to catch regressions.

## Run

```bash
go test ./...
```

This module has its own `go.mod`, so it is not part of the repository-root
`go test ./...` sweep; run it from this directory.
