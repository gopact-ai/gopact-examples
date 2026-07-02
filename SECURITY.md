# Security Policy

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
