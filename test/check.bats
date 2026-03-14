#!/usr/bin/env bats

load test_helper

@test "check: marks item as checked" {
  local out
  out="$(todo add Buy milk)"
  local id
  id="$(extract_id "${out}")"

  run todo check "${id}"
  assert_success
  assert_output --partial "[x]"
  assert_output --partial "Buy milk"

  run todo list
  assert_output --partial "[x] Buy milk"
}

@test "check: idempotent" {
  local out
  out="$(todo add Task)"
  local id
  id="$(extract_id "${out}")"

  todo check "${id}"
  run todo check "${id}"
  assert_success
}

@test "check: non-existent ID fails" {
  run todo check 999
  assert_failure
  assert_output --partial "not found"
}

@test "check: invalid ID format fails" {
  run todo check abc
  assert_failure
  assert_output --partial "invalid"
}

@test "check: requires an argument" {
  run todo check
  assert_failure
}

@test "check: cannot check item from another directory" {
  local out
  out="$(todo add Secret item)"
  local id
  id="$(extract_id "${out}")"

  local other_dir="${BATS_TEST_TMPDIR}/otherdir"
  mkdir -p "${other_dir}"
  cd "${other_dir}"

  run todo check "${id}"
  assert_failure
  assert_output --partial "not found"
}

@test "check: clears active status" {
  local out
  out="$(todo add Active task)"
  local id
  id="$(extract_id "${out}")"

  sqlite3 "${TODO_DB}" "UPDATE items SET is_active = 1 WHERE id = ${id}"

  # Verify it's active
  run todo list
  assert_output --partial "(active)"

  # Check should clear active
  todo check "${id}"
  run todo list
  refute_output --partial "(active)"
}
