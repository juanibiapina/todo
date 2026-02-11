#!/usr/bin/env bats

load test_helper

@test "add: creates a ticket" {
  run todo add "Fix the bug"
  assert_success
  assert_output --partial "Added"
  assert_output --partial "Fix the bug"
}

@test "add: ticket appears in list" {
  todo add "Fix the bug"
  run todo list
  assert_success
  assert_output --partial "Fix the bug"
}

@test "add: with description" {
  run todo add "Fix auth" "The auth handler needs work"
  assert_success
  assert_output --partial "Added"
  assert_output --partial "Fix auth"

  # Verify the description was stored
  local id
  id="$(extract_id_from_add "${output}")"
  run todo show "${id}"
  assert_output --partial "The auth handler needs work"
}

@test "add: with description via stdin" {
  run bash -c 'echo "Description from stdin" | todo add "Stdin ticket"'
  assert_success
  assert_output --partial "Added"

  local id
  id="$(extract_id_from_add "${output}")"
  run todo show "${id}"
  assert_output --partial "Description from stdin"
}

@test "add: generates a 3-character ID" {
  run todo add "Test ticket"
  assert_success
  local id
  id="$(extract_id_from_add "${output}")"
  [[ "${#id}" -eq 3 ]]
  [[ "${id}" =~ ^[A-Za-z0-9]{3}$ ]]
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

  # Extract IDs from list output (field 1 is the ID)
  local ids
  ids="$(todo list | awk '{print $1}' | sort -u)"
  local count
  count="$(echo "${ids}" | wc -l | tr -d ' ')"
  [[ "${count}" -eq 3 ]]
}
