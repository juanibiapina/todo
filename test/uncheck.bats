#!/usr/bin/env bats

load test_helper

@test "uncheck: marks item as unchecked" {
  local out
  out="$(todo add Buy milk)"
  local id
  id="$(extract_id "${out}")"

  todo check "${id}"
  run todo uncheck "${id}"
  assert_success
  assert_output --partial "[ ]"
  assert_output --partial "Buy milk"

  run todo list
  assert_output --partial "[ ] Buy milk"
}

@test "uncheck: idempotent" {
  local out
  out="$(todo add Task)"
  local id
  id="$(extract_id "${out}")"

  run todo uncheck "${id}"
  assert_success
}

@test "uncheck: non-existent ID fails" {
  run todo uncheck 999
  assert_failure
  assert_output --partial "not found"
}

@test "uncheck: invalid ID format fails" {
  run todo uncheck abc
  assert_failure
  assert_output --partial "invalid"
}

@test "uncheck: requires an argument" {
  run todo uncheck
  assert_failure
}
