#!/usr/bin/env bats

load test_helper

@test "close: sets status to closed" {
  local out
  out="$(todo add "Close me")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"

  run todo close "${id}"
  assert_success
  assert_output --partial "Closed ticket"
  assert_output --partial "Close me"

  run todo show "${id}"
  assert_success
  assert_output --partial "status: closed"
}

@test "close: hides ticket from list" {
  local out
  out="$(todo add "Will be closed")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"

  todo close "${id}"

  run todo list
  refute_output --partial "Will be closed"
}

@test "close: fails for nonexistent ticket" {
  run todo close "ZZZ"
  assert_failure
  assert_output --partial "ticket not found"
}

@test "close: file still exists on disk" {
  local out
  out="$(todo add "Persisted close")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"

  todo close "${id}"

  [[ -f "docs/tickets/${id}.md" ]]
}
