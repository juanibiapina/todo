#!/usr/bin/env bats

load test_helper

@test "partial id: show with exact match" {
  local out
  out="$(todo add "Exact match")"
  local id
  id="$(extract_id_from_add "${out}")"

  run todo show "${id}"
  assert_success
  assert_output --partial "Exact match"
}

@test "partial id: show with prefix" {
  local out
  out="$(todo add "Prefix test")"
  local id
  id="$(extract_id_from_add "${out}")"

  # Use first 2 chars as prefix
  local prefix="${id:0:2}"

  run todo show "${prefix}"
  assert_success
  assert_output --partial "Prefix test"
}

@test "partial id: show with suffix" {
  local out
  out="$(todo add "Suffix test")"
  local id
  id="$(extract_id_from_add "${out}")"

  # Use last 2 chars as suffix
  local suffix="${id:1:2}"

  run todo show "${suffix}"
  assert_success
  assert_output --partial "Suffix test"
}

@test "partial id: done with partial id" {
  local out
  out="$(todo add "Done partial")"
  local id
  id="$(extract_id_from_add "${out}")"

  local prefix="${id:0:2}"

  run todo done "${prefix}"
  assert_success

  # Verify it was closed
  run todo show "${id}"
  assert_success
  assert_output --partial "status: closed"
}

@test "partial id: status with partial id" {
  local out
  out="$(todo add "Status partial")"
  local id
  id="$(extract_id_from_add "${out}")"

  local prefix="${id:0:2}"

  run todo status "${prefix}" in_progress
  assert_success

  run todo show "${id}"
  assert_success
  assert_output --partial "status: in_progress"
}

@test "partial id: start with partial id" {
  local out
  out="$(todo add "Start partial")"
  local id
  id="$(extract_id_from_add "${out}")"

  local prefix="${id:0:2}"

  run todo start "${prefix}"
  assert_success

  run todo show "${id}"
  assert_success
  assert_output --partial "status: in_progress"
}

@test "partial id: close with partial id" {
  local out
  out="$(todo add "Close partial")"
  local id
  id="$(extract_id_from_add "${out}")"

  local prefix="${id:0:2}"

  run todo close "${prefix}"
  assert_success

  run todo show "${id}"
  assert_success
  assert_output --partial "status: closed"
}

@test "partial id: reopen with partial id" {
  local out
  out="$(todo add "Reopen partial")"
  local id
  id="$(extract_id_from_add "${out}")"

  todo done "${id}"

  local prefix="${id:0:2}"

  run todo reopen "${prefix}"
  assert_success

  run todo show "${id}"
  assert_success
  assert_output --partial "status: open"
}

@test "partial id: set-description with partial id" {
  local out
  out="$(todo add "Desc partial")"
  local id
  id="$(extract_id_from_add "${out}")"

  local prefix="${id:0:2}"

  run todo set-description "${prefix}" "New description"
  assert_success

  run todo show "${id}"
  assert_success
  assert_output --partial "New description"
}

@test "partial id: not found" {
  run todo show "ZZZZZZ"
  assert_failure
  assert_output --partial "ticket not found"
}

@test "partial id: exact match takes precedence" {
  # Create two tickets with controlled IDs by writing files directly
  mkdir -p docs/tickets

  cat > docs/tickets/ab.md << 'EOF'
---
id: ab
---
# Exact
EOF

  cat > docs/tickets/abc.md << 'EOF'
---
id: abc
---
# Longer
EOF

  # "ab" should exact-match "ab", not ambiguously match both
  run todo show "ab"
  assert_success
  assert_output --partial "# Exact"
}
