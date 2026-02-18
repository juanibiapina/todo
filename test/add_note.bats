#!/usr/bin/env bats

load test_helper

@test "add-note: adds a note to a ticket" {
  local out
  out="$(todo add "Note target")"
  local id
  id="$(extract_id_from_add "${out}")"

  run todo add-note "${id}" "This is a note"
  assert_success
  assert_output --partial "Added note to"

  run todo show "${id}"
  assert_output --partial "## Notes"
  assert_output --partial "This is a note"
}

@test "add-note: adds note via positional argument" {
  local out
  out="$(todo add "Positional note")"
  local id
  id="$(extract_id_from_add "${out}")"

  run todo add-note "${id}" "Positional text"
  assert_success

  run todo show "${id}"
  assert_output --partial "Positional text"
}

@test "add-note: adds note via stdin" {
  local out
  out="$(todo add "Stdin note")"
  local id
  id="$(extract_id_from_add "${out}")"

  run bash -c "echo 'From stdin pipe' | todo add-note ${id}"
  assert_success

  run todo show "${id}"
  assert_output --partial "From stdin pipe"
}

@test "add-note: appends to existing description" {
  local out
  out="$(todo add "With desc" "Original description")"
  local id
  id="$(extract_id_from_add "${out}")"

  todo add-note "${id}" "Appended note"

  run todo show "${id}"
  assert_output --partial "Original description"
  assert_output --partial "## Notes"
  assert_output --partial "Appended note"
}

@test "add-note: does not duplicate Notes header" {
  local out
  out="$(todo add "Multi note")"
  local id
  id="$(extract_id_from_add "${out}")"

  todo add-note "${id}" "First note"
  todo add-note "${id}" "Second note"

  run todo show "${id}"
  assert_output --partial "First note"
  assert_output --partial "Second note"

  # Count ## Notes headers — should be exactly 1
  local count
  count="$(todo show "${id}" | grep -c '## Notes')"
  [ "${count}" -eq 1 ]
}

@test "add-note: includes timestamp" {
  local out
  out="$(todo add "Timestamp note")"
  local id
  id="$(extract_id_from_add "${out}")"

  todo add-note "${id}" "Timestamped text"

  run todo show "${id}"
  # Timestamp format: **YYYY-MM-DD HH:MM UTC**
  assert_output --partial "**20"
  assert_output --partial "UTC**"
}

@test "add-note: fails for nonexistent ticket" {
  run todo add-note "ZZZ" "Some note"
  assert_failure
  assert_output --partial "ticket not found"
}

@test "add-note: fails with no text" {
  local out
  out="$(todo add "No text note")"
  local id
  id="$(extract_id_from_add "${out}")"

  run todo add-note "${id}"
  assert_failure
  assert_output --partial "no note text provided"
}

@test "add-note: works with partial IDs" {
  local out
  out="$(todo add "Partial note")"
  local id
  id="$(extract_id_from_add "${out}")"
  local partial="${id:0:2}"

  run todo add-note "${partial}" "Partial ID note"
  assert_success

  run todo show "${id}"
  assert_output --partial "Partial ID note"
}
