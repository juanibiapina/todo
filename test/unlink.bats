#!/usr/bin/env bats

load test_helper

@test "unlink removes bidirectional link" {
  run todo add "Ticket A"
  local id_a
  id_a="$(extract_id_from_add "$output")"

  run todo add "Ticket B"
  local id_b
  id_b="$(extract_id_from_add "$output")"

  run todo link "${id_a}" "${id_b}"
  assert_success

  run todo unlink "${id_a}" "${id_b}"
  assert_success
  assert_output "Unlinked ${id_a} and ${id_b}"

  # Neither should have links
  run todo show "${id_a}"
  assert_success
  refute_output --partial "links:"

  run todo show "${id_b}"
  assert_success
  refute_output --partial "links:"
}

@test "unlink is idempotent" {
  run todo add "Ticket A"
  local id_a
  id_a="$(extract_id_from_add "$output")"

  run todo add "Ticket B"
  local id_b
  id_b="$(extract_id_from_add "$output")"

  # Unlink tickets that were never linked — should succeed
  run todo unlink "${id_a}" "${id_b}"
  assert_success
}

@test "unlink fails when ticket does not exist" {
  run todo add "Ticket A"
  local id_a
  id_a="$(extract_id_from_add "$output")"

  run todo unlink "${id_a}" "zzz"
  assert_failure
  assert_output --partial "ticket not found"
}

@test "unlink works with partial IDs" {
  run todo add "Ticket A"
  local id_a
  id_a="$(extract_id_from_add "$output")"

  run todo add "Ticket B"
  local id_b
  id_b="$(extract_id_from_add "$output")"

  run todo link "${id_a}" "${id_b}"
  assert_success

  # Use first 2 chars as partial IDs
  local partial_a="${id_a:0:2}"
  local partial_b="${id_b:0:2}"

  run todo unlink "${partial_a}" "${partial_b}"
  assert_success

  # Verify links are removed from both
  run todo show "${id_a}"
  assert_success
  refute_output --partial "links:"

  run todo show "${id_b}"
  assert_success
  refute_output --partial "links:"
}

@test "unlink only removes the specified link" {
  run todo add "Ticket A"
  local id_a
  id_a="$(extract_id_from_add "$output")"

  run todo add "Ticket B"
  local id_b
  id_b="$(extract_id_from_add "$output")"

  run todo add "Ticket C"
  local id_c
  id_c="$(extract_id_from_add "$output")"

  # Link A to both B and C
  run todo link "${id_a}" "${id_b}"
  assert_success
  run todo link "${id_a}" "${id_c}"
  assert_success

  # Unlink A and B only
  run todo unlink "${id_a}" "${id_b}"
  assert_success

  # A should still link to C
  run todo show "${id_a}"
  assert_success
  assert_output --partial "${id_c}"
  refute_output --partial "- ${id_b}"
}

@test "unlink requires exactly 2 arguments" {
  run todo unlink
  assert_failure

  run todo add "Ticket A"
  local id_a
  id_a="$(extract_id_from_add "$output")"

  run todo unlink "${id_a}"
  assert_failure
}
