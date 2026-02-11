#!/usr/bin/env bats

load test_helper

@test "show: displays ticket by ID" {
  local out
  out="$(todo add "Show me")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"

  run todo show "${id}"
  assert_success
  assert_output --partial "Show me"
  assert_output --partial "id: ${id}"
}

@test "show: displays description" {
  local out
  out="$(todo add "Described ticket" "This is the description")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"

  run todo show "${id}"
  assert_success
  assert_output --partial "Described ticket"
  assert_output --partial "This is the description"
}

@test "show: nonexistent ID returns error" {
  run todo show "ZZZ"
  assert_failure
  assert_output --partial "ticket not found"
}

@test "show: outputs full markdown format" {
  local out
  out="$(todo add "Markdown ticket")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"

  run todo show "${id}"
  assert_success
  assert_output --partial "## Markdown ticket"
  assert_output --partial "---"
}
