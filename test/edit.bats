#!/usr/bin/env bats

load test_helper

@test "edit: updates item text" {
  local out
  out="$(todo add Buy milk)"
  local id
  id="$(extract_id "${out}")"

  run todo edit "${id}" Buy eggs
  assert_success
  assert_output --partial "Buy eggs"

  run todo list
  assert_output --partial "Buy eggs"
  refute_output --partial "Buy milk"
}

@test "edit: output shows id, check state, and new text" {
  local out
  out="$(todo add Original)"
  local id
  id="$(extract_id "${out}")"

  run todo edit "${id}" Updated
  assert_success
  assert_output "${id} [ ] Updated"
}

@test "edit: preserves checked state" {
  local out
  out="$(todo add Task)"
  local id
  id="$(extract_id "${out}")"
  todo check "${id}"

  run todo edit "${id}" Edited task
  assert_success
  assert_output "${id} [x] Edited task"
}

@test "edit: non-existent ID fails" {
  run todo edit 999 New text
  assert_failure
  assert_output --partial "not found"
}

@test "edit: invalid ID format fails" {
  run todo edit abc New text
  assert_failure
  assert_output --partial "invalid"
}

@test "edit: requires at least two arguments" {
  run todo edit
  assert_failure

  run todo edit 1
  assert_failure
}

@test "edit: cannot edit item from another directory" {
  local out
  out="$(todo add Secret item)"
  local id
  id="$(extract_id "${out}")"

  local other_dir="${BATS_TEST_TMPDIR}/otherdir"
  mkdir -p "${other_dir}"
  cd "${other_dir}"

  run todo edit "${id}" Hacked
  assert_failure
  assert_output --partial "not found"
}
