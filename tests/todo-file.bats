#!/usr/bin/env bats

load test_helper

@test "when there are no files, chooses .todo file" {
  run todo-file
  assert_success
  assert_output ".todo"
}

@test "chooses .todo file if it exists" {
  create_todo_file .todo
  run todo-file
  assert_success
  assert_output ".todo"
}

@test "uses global todo file if local .todo doesn't exist" {
  export TODO_DIR=/global-path

  run todo-file

  assert_success
  assert_output "/global-path/todo.txt"
}
