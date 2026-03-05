#!/usr/bin/env bats

load test_helper

@test "clean: deletes checked items" {
  todo add Keep this
  local out
  out="$(todo add Remove this)"
  local id
  id="$(extract_id "${out}")"
  todo check "${id}"

  run todo clean
  assert_success
  assert_output "Deleted 1 item(s)"

  run todo list
  assert_output --partial "Keep this"
  refute_output --partial "Remove this"
}

@test "clean: no checked items" {
  todo add Unchecked

  run todo clean
  assert_success
  assert_output "Deleted 0 item(s)"

  count="$(item_count)"
  [[ "${count}" -eq 1 ]]
}

@test "clean: deletes multiple checked items" {
  local out1 out2
  out1="$(todo add First)"
  out2="$(todo add Second)"
  todo add Third
  todo check "$(extract_id "${out1}")"
  todo check "$(extract_id "${out2}")"

  run todo clean
  assert_success
  assert_output "Deleted 2 item(s)"

  count="$(item_count)"
  [[ "${count}" -eq 1 ]]
}

@test "clean: empty list" {
  run todo clean
  assert_success
  assert_output "Deleted 0 item(s)"
}

@test "clean: scoped to current directory" {
  local out
  out="$(todo add Workdir item)"
  todo check "$(extract_id "${out}")"

  local other_dir="${BATS_TEST_TMPDIR}/otherdir"
  mkdir -p "${other_dir}"
  cd "${other_dir}"

  local out2
  out2="$(todo add Other item)"
  todo check "$(extract_id "${out2}")"

  # Clean only affects otherdir
  run todo clean
  assert_output "Deleted 1 item(s)"

  # Go back — workdir item still there
  cd "${TODO_TEST_DIR}"
  count="$(item_count)"
  [[ "${count}" -eq 1 ]]
}
