#!/usr/bin/env bats

load test_helper

@test "add: creates a ticket with new state" {
  run todo add "Fix the bug"
  assert_success
  assert_output --partial "(new)"
  assert_output --partial "Fix the bug"
}

@test "add: ticket appears in list" {
  todo add "Fix the bug"
  run todo list
  assert_success
  assert_output --partial "Fix the bug"
  assert_output --partial "new"
}

@test "add: with description sets state to refined" {
  run todo add "Fix auth" "The auth handler needs work"
  assert_success
  assert_output --partial "(refined)"
  assert_output --partial "Fix auth"
}

@test "add: with description via stdin sets state to refined" {
  run bash -c 'echo "Description from stdin" | todo add "Stdin ticket"'
  assert_success
  assert_output --partial "(refined)"
}

@test "add: generates a 3-character ID" {
  run todo add "Test ticket"
  assert_success
  # Output format: "Added ticket XXX (new): Test ticket"
  [[ "${output}" =~ Added\ ticket\ [A-Za-z0-9]{3} ]]
}

@test "add: duplicate titles are allowed" {
  run todo add "Same title"
  assert_success
  run todo add "Same title"
  assert_success

  count="$(ticket_count)"
  [[ "${count}" -eq 2 ]]
}

@test "add: creates TODO.md if it doesn't exist" {
  [[ ! -f TODO.md ]]
  todo add "First ticket"
  [[ -f TODO.md ]]
}

@test "add: multiple tickets get unique IDs" {
  todo add "Ticket one"
  todo add "Ticket two"
  todo add "Ticket three"

  # Extract IDs from list output
  local ids
  ids="$(todo list | awk '{print $1}' | sort -u)"
  local count
  count="$(echo "${ids}" | wc -l | tr -d ' ')"
  [[ "${count}" -eq 3 ]]
}
