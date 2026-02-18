#!/usr/bin/env bats

load test_helper

@test "done: closes ticket" {
  local out
  out="$(todo add "To be done")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"

  run todo done "${id}"
  assert_success
  assert_output --partial "Completed ticket"
  assert_output --partial "To be done"

  # Ticket should have status: closed
  run todo show "${id}"
  assert_success
  assert_output --partial "status: closed"

  # Ticket should no longer appear in list
  run todo list
  refute_output --partial "To be done"
}

@test "done: nonexistent ID returns error" {
  run todo done "ZZZ"
  assert_failure
  assert_output --partial "ticket not found"
}

@test "done: only closes the targeted ticket" {
  local out1 out2
  out1="$(todo add "Keep me")"
  out2="$(todo add "Remove me")"
  local id2
  id2="$(echo "${out2}" | awk '{print $2}')"

  todo done "${id2}"

  run todo list
  assert_success
  assert_output --partial "Keep me"
  refute_output --partial "Remove me"
}

@test "done: reduces visible ticket count by one" {
  todo add "One"
  local out
  out="$(todo add "Two")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"
  todo add "Three"

  [[ "$(ticket_count)" -eq 3 ]]

  todo done "${id}"

  [[ "$(ticket_count)" -eq 2 ]]
}

@test "done: ticket file still exists after close" {
  local out
  out="$(todo add "Persisted")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"

  todo done "${id}"

  # File should still be on disk
  [[ -f "docs/tickets/${id}.md" ]]
}
