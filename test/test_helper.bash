#!/usr/bin/env bash

# Load bats helpers
load "${BATS_TEST_DIRNAME}/bats-support/load"
load "${BATS_TEST_DIRNAME}/bats-assert/load"

setup() {
  # Build binary into test tmpdir
  go build -o "${BATS_TEST_TMPDIR}/todo" "${BATS_TEST_DIRNAME}/.."

  # Put our binary first in PATH
  export PATH="${BATS_TEST_TMPDIR}:${PATH}"

  # Create and cd into a temp working directory
  export TODO_TEST_DIR="${BATS_TEST_TMPDIR}/workdir"
  mkdir -p "${TODO_TEST_DIR}"
  cd "${TODO_TEST_DIR}"
}

# Helper: count tickets
ticket_count() {
  local output
  output="$(todo list)"
  if [[ "${output}" == "No tickets" ]]; then
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

# Helper: get the state of a ticket by ID
ticket_state() {
  local id="$1"
  todo show "${id}" | grep "^state:" | awk '{print $2}'
}

# Helper: get the ID from an "Added ticket XXX" message
extract_id() {
  echo "$1" | grep -oE '[A-Za-z0-9]{3}' | head -1
}
