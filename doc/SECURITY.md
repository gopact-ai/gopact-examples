# Security Policy

<!-- gopact:doc-language: zh,en -->

## 中文

本文档是 gopact 开源文档集的一部分，中文内容用于说明当前仓库约束、能力或维护流程。

## English

This document is part of the gopact open-source documentation set. The English section gives an entry point for readers who prefer English, while the remaining sections preserve the maintained technical details.


## Supported Versions

`gopact-examples` follows the latest `main` branch and the latest released
`gopact` / `gopact-ext` versions used by the examples.

## Reporting a Vulnerability

Do not open a public issue for suspected vulnerabilities. Report privately to
the maintainers through the gopact-ai organization owner channel until a
dedicated security advisory process is enabled.

Include:

- affected example path
- reproduction steps
- impact and trust boundary
- whether provider credentials, prompts, tool payloads, artifacts, or external
  tokens may be exposed

## Handling Guidelines

- Do not include secrets, tokens, raw prompts, raw model responses, raw tool
  args/results, or private customer data in issues, tests, examples, or logs.
- Keep `.env` local and use `.env.example` for placeholders only.
- CI must use mock services and must not require real provider credentials.
