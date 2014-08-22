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

@test "chooses TODO file if it exists" {
  create_todo_file TODO
  run todo-file
  assert_success "TODO"
}

@test "gives preference to TODO over .todo" {
  create_todo_file TODO
  create_todo_file .todo
  run todo-file
  assert_success "TODO"
}
