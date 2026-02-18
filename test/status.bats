#!/usr/bin/env bats

load test_helper

@test "status: sets status to open" {
  local out
  out="$(todo add "Status test")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"

  run todo status "${id}" open
  assert_success
  assert_output --partial "Status of"
  assert_output --partial "set to open"

  run todo show "${id}"
  assert_success
  assert_output --partial "status: open"
}

@test "status: sets status to in_progress" {
  local out
  out="$(todo add "Progress test")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"

  run todo status "${id}" in_progress
  assert_success
  assert_output --partial "set to in_progress"

  run todo show "${id}"
  assert_success
  assert_output --partial "status: in_progress"
}

@test "status: sets status to closed" {
  local out
  out="$(todo add "Close test")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"

  run todo status "${id}" closed
  assert_success
  assert_output --partial "set to closed"

  run todo show "${id}"
  assert_success
  assert_output --partial "status: closed"
}

@test "status: rejects invalid status" {
  local out
  out="$(todo add "Invalid test")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"

  run todo status "${id}" invalid
  assert_failure
  assert_output --partial "invalid status"
}

@test "status: fails for nonexistent ticket" {
  run todo status "ZZZ" open
  assert_failure
  assert_output --partial "ticket not found"
}

@test "status: closed ticket hidden from list" {
  local out
  out="$(todo add "Will close")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"

  todo status "${id}" closed

  run todo list
  refute_output --partial "Will close"
}

@test "status: reopened ticket visible in list" {
  local out
  out="$(todo add "Will reopen")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"

  todo status "${id}" closed

  run todo list
  refute_output --partial "Will reopen"

  todo status "${id}" open

  run todo list
  assert_output --partial "Will reopen"
}
