#!/usr/bin/env bats

load test_helper

@test "link creates bidirectional links" {
  run todo add "Ticket A"
  local id_a
  id_a="$(extract_id_from_add "$output")"

  run todo add "Ticket B"
  local id_b
  id_b="$(extract_id_from_add "$output")"

  run todo link "${id_a}" "${id_b}"
  assert_success
  assert_output "Linked tickets: ${id_a}, ${id_b}"

  # A should link to B
  run todo show "${id_a}"
  assert_success
  assert_output --partial "links:"
  assert_output --partial "${id_b}"

  # B should link to A
  run todo show "${id_b}"
  assert_success
  assert_output --partial "links:"
  assert_output --partial "${id_a}"
}

@test "link supports three or more tickets" {
  run todo add "Ticket A"
  local id_a
  id_a="$(extract_id_from_add "$output")"

  run todo add "Ticket B"
  local id_b
  id_b="$(extract_id_from_add "$output")"

  run todo add "Ticket C"
  local id_c
  id_c="$(extract_id_from_add "$output")"

  run todo link "${id_a}" "${id_b}" "${id_c}"
  assert_success
  assert_output "Linked tickets: ${id_a}, ${id_b}, ${id_c}"

  # A should link to B and C
  run todo show "${id_a}"
  assert_success
  assert_output --partial "${id_b}"
  assert_output --partial "${id_c}"

  # B should link to A and C
  run todo show "${id_b}"
  assert_success
  assert_output --partial "${id_a}"
  assert_output --partial "${id_c}"

  # C should link to A and B
  run todo show "${id_c}"
  assert_success
  assert_output --partial "${id_a}"
  assert_output --partial "${id_b}"
}

@test "link is idempotent" {
  run todo add "Ticket A"
  local id_a
  id_a="$(extract_id_from_add "$output")"

  run todo add "Ticket B"
  local id_b
  id_b="$(extract_id_from_add "$output")"

  run todo link "${id_a}" "${id_b}"
  assert_success

  run todo link "${id_a}" "${id_b}"
  assert_success

  # A should have B only once — count in the YAML frontmatter section only
  run todo show "${id_a}"
  assert_success
  # Extract frontmatter (between --- delimiters) and count link ID occurrences
  local fm_count
  fm_count=$(echo "$output" | sed -n '/^---$/,/^---$/p' | grep -c "${id_b}" || true)
  [ "$fm_count" -eq 1 ]
}

@test "link fails when ticket does not exist" {
  run todo add "Ticket A"
  local id_a
  id_a="$(extract_id_from_add "$output")"

  run todo link "${id_a}" "zzz"
  assert_failure
  assert_output --partial "ticket not found"
}

@test "link works with partial IDs" {
  run todo add "Ticket A"
  local id_a
  id_a="$(extract_id_from_add "$output")"

  run todo add "Ticket B"
  local id_b
  id_b="$(extract_id_from_add "$output")"

  # Use first 2 chars as partial IDs
  local partial_a="${id_a:0:2}"
  local partial_b="${id_b:0:2}"

  run todo link "${partial_a}" "${partial_b}"
  assert_success

  # Verify the full resolved IDs are stored
  run todo show "${id_a}"
  assert_success
  assert_output --partial "${id_b}"

  run todo show "${id_b}"
  assert_success
  assert_output --partial "${id_a}"
}

@test "link does not create self-links" {
  run todo add "Ticket A"
  local id_a
  id_a="$(extract_id_from_add "$output")"

  run todo add "Ticket B"
  local id_b
  id_b="$(extract_id_from_add "$output")"

  run todo link "${id_a}" "${id_b}"
  assert_success

  # A should not link to itself
  run todo show "${id_a}"
  assert_success
  local count
  count=$(echo "$output" | grep -c "${id_a}" || true)
  # id_a appears in the "id:" line but should NOT appear in links
  [ "$count" -eq 1 ]
}

@test "link requires at least 2 arguments" {
  run todo link
  assert_failure

  run todo add "Ticket A"
  local id_a
  id_a="$(extract_id_from_add "$output")"

  run todo link "${id_a}"
  assert_failure
}

@test "link preserves existing links" {
  run todo add "Ticket A"
  local id_a
  id_a="$(extract_id_from_add "$output")"

  run todo add "Ticket B"
  local id_b
  id_b="$(extract_id_from_add "$output")"

  run todo add "Ticket C"
  local id_c
  id_c="$(extract_id_from_add "$output")"

  # Link A and B first
  run todo link "${id_a}" "${id_b}"
  assert_success

  # Then link A and C
  run todo link "${id_a}" "${id_c}"
  assert_success

  # A should have both B and C
  run todo show "${id_a}"
  assert_success
  assert_output --partial "${id_b}"
  assert_output --partial "${id_c}"
}
