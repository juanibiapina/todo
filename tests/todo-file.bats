#!/usr/bin/env bats

load test_helper

@test "chooses .todo file" {
  run todo-file
  assert_success ".todo"
}
