#!/usr/bin/env bats

load test_helper

@test "list: empty list returns no output" {
  run todo list
  assert_success
  assert_output ""
}

@test "list: shows items with IDs and checkboxes" {
  todo add First item
  todo add Second item

  run todo list
  assert_success
  assert_output --partial "[ ] First item"
  assert_output --partial "[ ] Second item"
}

@test "list: shows checked and unchecked items" {
  todo add Unchecked item
  local out
  out="$(todo add Checked item)"
  local id
  id="$(extract_id "${out}")"
  todo check "${id}"

  run todo list
  assert_success
  assert_output --partial "[ ] Unchecked item"
  assert_output --partial "[x] Checked item"
}

@test "list: items ordered by ID" {
  todo add First
  todo add Second
  todo add Third

  run todo list
  assert_success

  # Extract IDs and verify ascending order
  local id1 id2 id3
  id1="$(echo "${output}" | sed -n '1p' | awk '{print $1}')"
  id2="$(echo "${output}" | sed -n '2p' | awk '{print $1}')"
  id3="$(echo "${output}" | sed -n '3p' | awk '{print $1}')"
  [[ "${id1}" -lt "${id2}" ]]
  [[ "${id2}" -lt "${id3}" ]]
}

@test "list: scoped to current directory" {
  todo add Item in workdir

  local other_dir="${BATS_TEST_TMPDIR}/otherdir"
  mkdir -p "${other_dir}"
  cd "${other_dir}"

  run todo list
  assert_success
  assert_output ""
}

@test "list: bare todo runs list" {
  todo add Hello world
  run todo
  assert_success
  assert_output --partial "Hello world"
}

@test "list: shows active marker" {
  local out1
  out1="$(todo add Buy groceries)"
  todo add Fix bug
  local id1
  id1="$(extract_id "${out1}")"

  # Make first item active via database directly (no CLI subcommand yet)
  sqlite3 "${TODO_DB}" "UPDATE items SET is_active = 1 WHERE id = ${id1}"

  run todo list
  assert_success
  assert_output --partial "Buy groceries  (active)"
  refute_output --partial "Fix bug  (active)"
}
