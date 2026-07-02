# OpenAI Chat Quickstart

[![CI](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/gopact-ai/gopact-examples)](../../LICENSE)

<!-- gopact:doc-language: en -->

Chinese documentation: [README_zh.md](README_zh.md)

This example makes one OpenAI-compatible chat completion call through `gopact-ext/models/openai`. Default tests use a fake server; real provider runs read `GOPACT_LLM_BASEURL`, `GOPACT_LLM_TOKEN`, and `GOPACT_LLM_MODEL` from `.env`.
