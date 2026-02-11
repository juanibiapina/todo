#!/usr/bin/env bats

load test_helper

@test "set-state: change to refined" {
  local out
  out="$(todo add "State test")"
  local id
  id="$(extract_id_from_add "${out}")"

  run todo set-state "${id}" refined
  assert_success
  assert_output --partial "refined"

  [[ "$(ticket_state "${id}")" == "refined" ]]
}

@test "set-state: change to new" {
  local out
  out="$(todo add "Refined ticket" "Has description")"
  local id
  id="$(extract_id_from_add "${out}")"

  run todo set-state "${id}" new
  assert_success
  assert_output --partial "new"

  [[ "$(ticket_state "${id}")" == "new" ]]
}

@test "set-state: planned is no longer valid" {
  local out
  out="$(todo add "Plan me")"
  local id
  id="$(extract_id_from_add "${out}")"

  run todo set-state "${id}" planned
  assert_failure
  assert_output --partial "invalid state"
}

@test "set-state: invalid state returns error" {
  local out
  out="$(todo add "Bad state")"
  local id
  id="$(extract_id_from_add "${out}")"

  run todo set-state "${id}" invalid
  assert_failure
  assert_output --partial "invalid state"
}

@test "set-state: nonexistent ticket returns error" {
  run todo set-state "ZZZ" refined
  assert_failure
  assert_output --partial "ticket not found"
}
