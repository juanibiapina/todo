#!/usr/bin/env bats

load test_helper

@test "undep removes a dependency" {
  run todo add "Ticket A"
  local id_a
  id_a="$(extract_id_from_add "$output")"

  run todo add "Ticket B"
  local id_b
  id_b="$(extract_id_from_add "$output")"

  run todo dep "${id_a}" "${id_b}"
  assert_success

  run todo undep "${id_a}" "${id_b}"
  assert_success
  assert_output "Removed dependency ${id_b} from ${id_a}"

  run todo show "${id_a}"
  assert_success
  refute_output --partial "deps:"
}

@test "undep is idempotent" {
  run todo add "Ticket A"
  local id_a
  id_a="$(extract_id_from_add "$output")"

  run todo add "Ticket B"
  local id_b
  id_b="$(extract_id_from_add "$output")"

  run todo dep "${id_a}" "${id_b}"
  assert_success

  run todo undep "${id_a}" "${id_b}"
  assert_success

  # Removing again should still succeed
  run todo undep "${id_a}" "${id_b}"
  assert_success
}

@test "undep fails when ticket does not exist" {
  run todo add "Ticket B"
  local id_b
  id_b="$(extract_id_from_add "$output")"

  run todo undep "zzz" "${id_b}"
  assert_failure
  assert_output --partial "ticket not found"
}

@test "undep works with partial IDs" {
  run todo add "Ticket A"
  local id_a
  id_a="$(extract_id_from_add "$output")"

  run todo add "Ticket B"
  local id_b
  id_b="$(extract_id_from_add "$output")"

  run todo dep "${id_a}" "${id_b}"
  assert_success

  # Use first 2 chars as partial IDs
  local partial_a="${id_a:0:2}"
  local partial_b="${id_b:0:2}"

  run todo undep "${partial_a}" "${partial_b}"
  assert_success

  run todo show "${id_a}"
  assert_success
  refute_output --partial "deps:"
}

@test "undep requires exactly 2 arguments" {
  run todo undep
  assert_failure

  run todo add "Ticket A"
  local id_a
  id_a="$(extract_id_from_add "$output")"

  run todo undep "${id_a}"
  assert_failure
}
