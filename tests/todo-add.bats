#!/usr/bin/env bats

load test_helper

@test "without a todo file, creates it and add the item" {
  create_command todo-file "echo FILE"

  todo-add new-item

  run todo-ls
  assert_success
  assert_output " 1 - new-item"
}

@test "with content, appends to the content" {
  create_command todo-file "echo FILE"
  create_todo_file "FILE" "item 1
item 2"

  todo-add new-item

  run todo-ls
  assert_success
  assert_output " 1 - item 1
 2 - item 2
 3 - new-item"
}
