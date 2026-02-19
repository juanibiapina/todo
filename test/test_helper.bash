#!/usr/bin/env bash

# Load bats helpers
load "${BATS_TEST_DIRNAME}/bats-support/load"
load "${BATS_TEST_DIRNAME}/bats-assert/load"

setup() {
  # Build binary into test tmpdir
  go build -o "${BATS_TEST_TMPDIR}/todo" "${BATS_TEST_DIRNAME}/.."

  # Put our binary first in PATH
  export PATH="${BATS_TEST_TMPDIR}:${PATH}"

  # Create and cd into a temp working directory with git identity
  export TODO_TEST_DIR="${BATS_TEST_TMPDIR}/workdir"
  mkdir -p "${TODO_TEST_DIR}"
  cd "${TODO_TEST_DIR}"
  git init --quiet
  git config user.name "Test User"
  git config user.email "test@example.com"
}

# Helper: count tickets
ticket_count() {
  local output
  output="$(todo list)"
  if [[ -z "${output}" ]]; then
    echo 0
  else
    echo "${output}" | wc -l | tr -d ' '
  fi
}

# Helper: check if a ticket with given title exists
has_ticket() {
  local title="$1"
  todo list | grep -qF "${title}"
}

# Helper: get the ID from an "Added ticket XXX" message (legacy)
extract_id() {
  echo "$1" | grep -oE '[A-Za-z0-9]{3}' | head -1
}

# Helper: extract ID from add command output "Added <ID> <title>"
extract_id_from_add() {
  echo "$1" | awk '{print $2}'
}
