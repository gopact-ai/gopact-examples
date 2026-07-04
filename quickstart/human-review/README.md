# Human Review

<!-- gopact:doc-language: en -->

This quickstart shows a credential-free approval gate using `gopact-ext/agents/humanreview`.

```bash
go run ./quickstart/human-review
```

It builds a typed graph with `draft -> review -> publish`. The review node emits a `humanreview` approval interrupt, then the example resumes the same gate through both step export/import and checkpoint resume.
