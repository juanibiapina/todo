#!/usr/bin/env bats

load test_helper

@test "start: sets status to in_progress" {
  local out
  out="$(todo add "Start test")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"

  run todo start "${id}"
  assert_success
  assert_output --partial "Started ticket"
  assert_output --partial "Start test"

  run todo show "${id}"
  assert_success
  assert_output --partial "status: in_progress"
}

@test "start: fails for nonexistent ticket" {
  run todo start "ZZZ"
  assert_failure
  assert_output --partial "ticket not found"
}

@test "start: ticket remains visible in list" {
  local out
  out="$(todo add "Keep visible")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"

  todo start "${id}"

  run todo list
  assert_output --partial "Keep visible"
}
