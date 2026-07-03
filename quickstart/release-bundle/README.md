# release-bundle

<!-- gopact:doc-language: en -->

Chinese documentation: [README_zh.md](README_zh.md)

This quickstart shows the credential-free self-bootstrap release bundle path. It writes a recorded `RunExport` and an observed `VerificationReport`, invokes the core `gopact release-bundle -run-export <file> -report <file>` CLI, parses the JSON bundle, and checks the self-bootstrap release gate.

```bash
go run ./quickstart/release-bundle
```

No provider credentials are required.
