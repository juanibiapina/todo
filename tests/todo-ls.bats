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
  assert_output " 1 - item 1
 2 - item 2"
}

@test "with a filter, applies the filter" {
  create_command todo-file "echo FILE"
  create_todo_file "FILE" "banana 1
apple 3
banana 2
carrot 4"

  run todo-ls banana
  assert_success
  assert_output " 1 - banana 1
 3 - banana 2"
}
