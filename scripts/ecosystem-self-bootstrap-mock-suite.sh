#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "$0")" && pwd)"
examples_dir="$(cd "${script_dir}/.." && pwd)"
default_root="$(cd "${examples_dir}/.." && pwd)"
ecosystem_root="${GOPACT_ECOSYSTEM_ROOT:-${default_root}}"
ecosystem_fetch="${GOPACT_ECOSYSTEM_FETCH:-0}"

export GOPRIVATE="${GOPRIVATE:-github.com/gopact-ai/*}"
export GONOSUMDB="${GONOSUMDB:-github.com/gopact-ai/*}"

resolve_repo() {
  local name="$1"
  local url="$2"
  local fetch_tags="$3"
  local local_path="${ecosystem_root}/${name}"

  if [[ "${name}" == "gopact-examples" ]]; then
    printf '%s\n' "${examples_dir}"
    return
  fi

  if [[ -d "${local_path}/.git" ]]; then
    printf '%s\n' "${local_path}"
    return
  fi

  if [[ "${ecosystem_fetch}" == "1" ]]; then
    mkdir -p "${ecosystem_root}"
    git clone --depth=1 "${url}" "${local_path}"
    if [[ "${fetch_tags}" == "1" ]]; then
      git -C "${local_path}" fetch --force --tags origin "refs/tags/*:refs/tags/*"
    fi
    printf '%s\n' "${local_path}"
    return
  fi

  echo "ecosystem-smoke: missing ${local_path}" >&2
  echo "ecosystem-smoke: set GOPACT_ECOSYSTEM_FETCH=1 to clone missing repositories" >&2
  return 1
}

run_repo_suite() {
  local name="$1"
  local url="$2"
  local fetch_tags="$3"
  local path
  path="$(resolve_repo "${name}" "${url}" "${fetch_tags}")"

  echo "ecosystem-smoke: ${name}"
  (cd "${path}" && ./scripts/self-bootstrap-mock-suite.sh)
}

run_repo_suite "gopact" "https://github.com/gopact-ai/gopact.git" "1"
run_repo_suite "gopact-ext" "https://github.com/gopact-ai/gopact-ext.git" "0"
run_repo_suite "gopact-examples" "https://github.com/gopact-ai/gopact-examples.git" "0"
