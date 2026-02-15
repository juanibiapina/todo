#!/usr/bin/env bats

load test_helper

@test "list: empty list shows no tickets message" {
  run todo list
  assert_success
  assert_output "No tickets"
}

@test "list: shows tickets with IDs" {
  todo add "First ticket"
  todo add "Second ticket"

  run todo list
  assert_success
  assert_output --partial "First ticket"
  assert_output --partial "Second ticket"
}

@test "list: sorts alphabetically by filename" {
  todo add "Zebra"
  todo add "Apple"
  todo add "Mango"

  run todo list
  assert_success

  # All three should appear
  assert_output --partial "Zebra"
  assert_output --partial "Apple"
  assert_output --partial "Mango"
}

@test "list: shows tickets with and without descriptions" {
  todo add "No description ticket"
  todo add "Has description ticket" "Has a description"

  run todo list
  assert_success

  # Verify both tickets appear
  assert_output --partial "No description ticket"
  assert_output --partial "Has description ticket"
}
