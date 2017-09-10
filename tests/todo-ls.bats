#!/usr/bin/env bats

load test_helper

@test "without a todo file, displays nothing" {
  create_command todo-file "echo NOFILE"

  run todo-ls
  assert_success
  assert_output ""
}

@test "displays the contents of a todo file" {
  create_command todo-file "echo FILE"
  create_todo_file "FILE" "item 1
item 2"

  run todo-ls
  assert_success
  assert_output "  1 - item 1
  2 - item 2"
}
