#!/usr/bin/env bats

load test_helper

@test "displays a list of tags" {
  create_command todo-file "echo FILE"
  create_todo_file "FILE" "item 1 #tag2
item 2 #tag2
item 3 #tag1"

  run todo-tags
  assert_success
  assert_output "tag1
tag2"
}
