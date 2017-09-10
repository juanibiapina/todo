#!/usr/bin/env bats

load test_helper

@test "when there are no files, chooses .todo file" {
  run todo-file
  assert_success ".todo"
}

@test "chooses .todo file if it exists" {
  create_todo_file .todo
  run todo-file
  assert_success ".todo"
}
