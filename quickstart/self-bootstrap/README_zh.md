# self-bootstrap

<!-- gopact:doc-language: zh -->

[英文文档](./README.md)

## 中文

这个 quickstart 使用 `gopact-ext/devagent/selfbootstrap` 运行一条无凭据 Dev Agent self-bootstrap workflow。示例注入本地 analyze、plan、write、test、review 阶段，并输出 run export 和 verification evidence 摘要。

```bash
go run ./quickstart/self-bootstrap
```

示例不会调用模型、执行命令或修改工作区。它展示宿主如何把已经观察到的 diff、file snapshot、command、CI gate 和 review result 交给可复用 self-bootstrap workflow。
