# Background Scheduler

[![CI](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/gopact-ai/gopact-examples)](../../LICENSE)

<!-- gopact:doc-language: en -->

Chinese documentation: [README_zh.md](README_zh.md)

This example shows `agents/scheduler` running leased background work without provider credentials. It uses an in-memory queue and lease backend so the behavior is deterministic in CI.

```bash
go run ./quickstart/background-scheduler
```

It covers:

- Running a bounded background drain with a leased worker.
- Retrying a failed job and completing the next attempt.
- Dead-lettering a permanently failed job.
- Recording schedule verification evidence.
- Releasing worker ownership after each pass.
