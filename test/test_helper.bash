#!/usr/bin/env bash

# Load bats helpers
load "${BATS_TEST_DIRNAME}/bats-support/load"
load "${BATS_TEST_DIRNAME}/bats-assert/load"

setup() {
  # Build binary into test tmpdir
  go build -o "${BATS_TEST_TMPDIR}/todo" "${BATS_TEST_DIRNAME}/.."

  # Put our binary first in PATH
  export PATH="${BATS_TEST_TMPDIR}:${PATH}"

  # Use a per-test database so tests are isolated
  export TODO_DB="${BATS_TEST_TMPDIR}/todo.db"

  # Create and cd into a temp working directory
  export TODO_TEST_DIR="${BATS_TEST_TMPDIR}/workdir"
  mkdir -p "${TODO_TEST_DIR}"
  cd "${TODO_TEST_DIR}"
}

# Helper: count items in list output
item_count() {
  local output
  output="$(todo list)"
  if [[ -z "${output}" ]]; then
    echo 0
  else
    echo "${output}" | wc -l | tr -d ' '
  fi
}

# Helper: extract ID from add output "ID text"
extract_id() {
  echo "$1" | awk '{print $1}'
}
