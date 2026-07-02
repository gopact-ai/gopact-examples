# Security Policy

<!-- gopact:doc-language: en -->

Chinese documentation: [SECURITY_zh.md](SECURITY_zh.md)

`gopact-examples` demonstrates provider, tool, A2A, and engineering-evidence flows. The security baseline is credential isolation and public-safe logs. Default examples must run in CI without real provider credentials.

## Supported Versions

The `main` branch receives security fixes. Released example snapshots are best-effort unless a maintainer explicitly marks a tag as supported.

## Reporting a Vulnerability

Do not open public issues for suspected vulnerabilities. Report privately through the `gopact-ai` maintainer channel until GitHub Security Advisory handling is enabled.

Include the affected quickstart, reproduction steps, expected impact, and whether any credential, endpoint ID, model response, or user data may have appeared in logs, issues, pull requests, or commit messages.

## Secret Handling

Do not commit `.env`, provider credentials, endpoint IDs, captured model payloads with private data, or generated artifacts that contain secrets. If a secret reaches git history, rotate the credential before publishing any cleanup commit.
