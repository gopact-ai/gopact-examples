# release-bundle

<!-- gopact:doc-language: zh -->

[英文文档](./README.md)

## 中文

这个 quickstart 演示无凭据 self-bootstrap release bundle 路径。示例会写入已记录的 `RunExport` 和已观察的 `VerificationReport`，调用 core `gopact release-bundle -run-export <file> -report <file>` CLI，解析 JSON bundle，并校验 self-bootstrap release gate。

```bash
go run ./quickstart/release-bundle
```

不需要 provider 凭据。
