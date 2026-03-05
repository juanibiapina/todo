#!/usr/bin/env bats

load test_helper

@test "version: shows version" {
  run todo --version
  assert_success
  assert_output --partial "todo version"
}

@test "help: shows help" {
  run todo --help
  assert_success
  assert_output --partial "simple todo items"
}

@test "unknown command: fails" {
  run todo nonexistent
  assert_failure
}
