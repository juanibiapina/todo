#!/usr/bin/env bats

load test_helper

@test "set-description: sets description on ticket" {
  local out
  out="$(todo add "Describe me")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"

  run todo set-description "${id}" "New description"
  assert_success
  assert_output --partial "Updated description"

  run todo show "${id}"
  assert_output --partial "New description"
}

@test "set-description: replaces existing description" {
  local out
  out="$(todo add "Has desc" "Old description")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"

  todo set-description "${id}" "Replaced description"

  run todo show "${id}"
  assert_output --partial "Replaced description"
  refute_output --partial "Old description"
}

@test "set-description: via stdin" {
  local out
  out="$(todo add "Stdin desc")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"

  run bash -c "echo 'From stdin' | todo set-description ${id}"
  assert_success

  run todo show "${id}"
  assert_output --partial "From stdin"
}

@test "set-description: empty description returns error" {
  local out
  out="$(todo add "No desc")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"

  run todo set-description "${id}"
  assert_failure
  assert_output --partial "no description provided"
}

@test "set-description: nonexistent ticket returns error" {
  run todo set-description "ZZZ" "Some description"
  assert_failure
  assert_output --partial "ticket not found"
}
