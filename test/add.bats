#!/usr/bin/env bats

load test_helper

@test "add: creates an item" {
  run todo add Buy groceries
  assert_success
  assert_output --partial "Buy groceries"
}

@test "add: item appears in list" {
  todo add Buy groceries
  run todo list
  assert_success
  assert_output --partial "Buy groceries"
}

@test "add: multiple words joined as text" {
  run todo add Fix the login bug
  assert_success
  assert_output --partial "Fix the login bug"
}

@test "add: multiple items get unique IDs" {
  todo add First
  todo add Second
  todo add Third

  run todo list
  assert_success

  local ids
  ids="$(echo "${output}" | awk '{print $1}' | sort -un)"
  local count
  count="$(echo "${ids}" | wc -l | tr -d ' ')"
  [[ "${count}" -eq 3 ]]
}

@test "add: requires at least one argument" {
  run todo add
  assert_failure
}

@test "add: new items are unchecked" {
  todo add Test item
  run todo list
  assert_success
  assert_output --partial "[ ]"
}
