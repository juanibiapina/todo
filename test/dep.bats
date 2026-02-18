#!/usr/bin/env bats

load test_helper

@test "dep adds a dependency" {
  run todo add "Ticket A"
  local id_a
  id_a="$(extract_id_from_add "$output")"

  run todo add "Ticket B"
  local id_b
  id_b="$(extract_id_from_add "$output")"

  run todo dep "${id_a}" "${id_b}"
  assert_success
  assert_output "Added dependency ${id_b} to ${id_a}"

  run todo show "${id_a}"
  assert_success
  assert_output --partial "deps:"
  assert_output --partial "${id_b}"
}

@test "dep is idempotent" {
  run todo add "Ticket A"
  local id_a
  id_a="$(extract_id_from_add "$output")"

  run todo add "Ticket B"
  local id_b
  id_b="$(extract_id_from_add "$output")"

  run todo dep "${id_a}" "${id_b}"
  assert_success

  run todo dep "${id_a}" "${id_b}"
  assert_success

  # Should only have one dep entry, not two — count in the YAML frontmatter section only
  run todo show "${id_a}"
  assert_success
  # Extract frontmatter (between --- delimiters) and count dep ID occurrences
  local fm_count
  fm_count=$(echo "$output" | sed -n '/^---$/,/^---$/p' | grep -c "${id_b}" || true)
  [ "$fm_count" -eq 1 ]
}

@test "dep fails when ticket does not exist" {
  run todo add "Ticket B"
  local id_b
  id_b="$(extract_id_from_add "$output")"

  run todo dep "zzz" "${id_b}"
  assert_failure
  assert_output --partial "ticket not found"
}

@test "dep fails when dependency does not exist" {
  run todo add "Ticket A"
  local id_a
  id_a="$(extract_id_from_add "$output")"

  run todo dep "${id_a}" "zzz"
  assert_failure
  assert_output --partial "ticket not found"
}

@test "dep works with partial IDs" {
  run todo add "Ticket A"
  local id_a
  id_a="$(extract_id_from_add "$output")"

  run todo add "Ticket B"
  local id_b
  id_b="$(extract_id_from_add "$output")"

  # Use first 2 chars as partial IDs
  local partial_a="${id_a:0:2}"
  local partial_b="${id_b:0:2}"

  run todo dep "${partial_a}" "${partial_b}"
  assert_success

  # Verify the full resolved ID is stored
  run todo show "${id_a}"
  assert_success
  assert_output --partial "${id_b}"
}

@test "dep supports multiple dependencies" {
  run todo add "Ticket A"
  local id_a
  id_a="$(extract_id_from_add "$output")"

  run todo add "Ticket B"
  local id_b
  id_b="$(extract_id_from_add "$output")"

  run todo add "Ticket C"
  local id_c
  id_c="$(extract_id_from_add "$output")"

  run todo dep "${id_a}" "${id_b}"
  assert_success

  run todo dep "${id_a}" "${id_c}"
  assert_success

  run todo show "${id_a}"
  assert_success
  assert_output --partial "${id_b}"
  assert_output --partial "${id_c}"
}

@test "dep does not modify the dependency ticket" {
  run todo add "Ticket A"
  local id_a
  id_a="$(extract_id_from_add "$output")"

  run todo add "Ticket B"
  local id_b
  id_b="$(extract_id_from_add "$output")"

  run todo dep "${id_a}" "${id_b}"
  assert_success

  # The dependency ticket should have no deps
  run todo show "${id_b}"
  assert_success
  refute_output --partial "deps:"
}

@test "dep requires exactly 2 arguments" {
  run todo dep
  assert_failure

  run todo add "Ticket A"
  local id_a
  id_a="$(extract_id_from_add "$output")"

  run todo dep "${id_a}"
  assert_failure
}
