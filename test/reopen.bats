#!/usr/bin/env bats

load test_helper

@test "reopen: sets status to open" {
  local out
  out="$(todo add "Reopen me")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"

  todo close "${id}"
  run todo show "${id}"
  assert_output --partial "status: closed"

  run todo reopen "${id}"
  assert_success
  assert_output --partial "Reopened ticket"
  assert_output --partial "Reopen me"

  run todo show "${id}"
  assert_success
  assert_output --partial "status: open"
}

@test "reopen: ticket reappears in list" {
  local out
  out="$(todo add "Come back")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"

  todo close "${id}"
  run todo list
  refute_output --partial "Come back"

  todo reopen "${id}"
  run todo list
  assert_output --partial "Come back"
}

@test "reopen: fails for nonexistent ticket" {
  run todo reopen "ZZZ"
  assert_failure
  assert_output --partial "ticket not found"
}

@test "reopen: works after done command" {
  local out
  out="$(todo add "Done then reopen")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"

  todo done "${id}"
  run todo show "${id}"
  assert_output --partial "status: closed"

  run todo reopen "${id}"
  assert_success

  run todo show "${id}"
  assert_output --partial "status: open"

  run todo list
  assert_output --partial "Done then reopen"
}
