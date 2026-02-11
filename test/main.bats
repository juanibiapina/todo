#!/usr/bin/env bats

load test_helper

@test "version: --version flag shows version" {
  run todo --version
  assert_success
  assert_output --partial "todo version"
}

@test "help: --help flag shows usage" {
  run todo --help
  assert_success
  assert_output --partial "A CLI for managing tickets stored as markdown"
  assert_output --partial "Available Commands"
}

@test "unknown command: returns error" {
  run todo nonexistent
  assert_failure
  assert_output --partial "unknown command"
}

@test "help: shows all subcommands" {
  run todo --help
  assert_success
  assert_output --partial "add"
  assert_output --partial "list"
  assert_output --partial "show"
  assert_output --partial "done"
  assert_output --partial "set-state"
  assert_output --partial "set-description"
  assert_output --partial "move-up"
  assert_output --partial "move-down"
}
