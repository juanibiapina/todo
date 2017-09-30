#!/usr/bin/env bats

load test_helper

@test "removes the nth todo item" {
  create_command todo-file "echo FILE"
  create_todo_file "FILE" "item 1
item 2
item 3"

  run todo-do 2
  assert_success

  run todo-ls
  assert_success
  assert_output " 1 - item 1
 2 - item 3"
}
