#!/usr/bin/env bats

load test_helper

@test "add --section: creates a section" {
  run todo add --section
  assert_success
}

@test "add -s: creates a section (short flag)" {
  run todo add -s
  assert_success
}

@test "add --section: accepts a title" {
  run todo add --section Planning
  assert_success
}

@test "add --section: section appears in list as divider" {
  todo add -s
  run todo list
  assert_success
  assert_output --partial "──────────"
}

@test "add --section: titled section shows the title in list output" {
  todo add -s Planning
  run todo list
  assert_success
  assert_output --partial "── Planning ─"
}

@test "list: sections render without checkboxes" {
  todo add -s
  todo add Fix the bug
  run todo list
  assert_success
  assert_output --partial "──────────"
  assert_output --partial "[ ] Fix the bug"
  # Section line should not contain checkbox
  local section_line
  section_line="$(echo "${output}" | grep "─────")"
  refute [ "$(echo "${section_line}" | grep -c '\[.\]')" -gt 0 ]
}

@test "list: sections and items coexist in order" {
  todo add Fix the login bug
  todo add -s
  todo add Deploy to staging
  todo add Buy groceries

  run todo list
  assert_success

  local line1 line2 line3 line4
  line1="$(echo "${output}" | sed -n '1p')"
  line2="$(echo "${output}" | sed -n '2p')"
  line3="$(echo "${output}" | sed -n '3p')"
  line4="$(echo "${output}" | sed -n '4p')"
  [[ "${line1}" == *"Fix the login bug"* ]]
  [[ "${line2}" == *"──────────"* ]]
  [[ "${line3}" == *"Deploy to staging"* ]]
  [[ "${line4}" == *"Buy groceries"* ]]
}

@test "check: rejects section" {
  todo add -s

  # Get the section ID from list (sections don't show IDs, so use the db directly)
  # Add a regular item first so we know the section ID
  local out
  out="$(todo add Task)"
  local task_id
  task_id="$(extract_id "${out}")"
  # Section was added first, so its ID is task_id - 1
  local section_id=$((task_id - 1))

  run todo check "${section_id}"
  assert_failure
  assert_output --partial "is a section"
}

@test "uncheck: rejects section" {
  todo add -s

  local out
  out="$(todo add Task)"
  local task_id
  task_id="$(extract_id "${out}")"
  local section_id=$((task_id - 1))

  run todo uncheck "${section_id}"
  assert_failure
  assert_output --partial "is a section"
}

@test "edit: renames a section" {
  todo add -s

  local out
  out="$(todo add Task)"
  local task_id
  task_id="$(extract_id "${out}")"
  local section_id=$((task_id - 1))

  run todo edit "${section_id}" "Renamed"
  assert_success

  run todo list
  assert_output --partial "── Renamed ─"
}

@test "clean: does not remove sections" {
  todo add -s
  local out
  out="$(todo add Remove this)"
  local id
  id="$(extract_id "${out}")"
  todo check "${id}"

  run todo clean
  assert_success
  assert_output "Deleted 1 item(s)"

  run todo list
  assert_output --partial "──────────"
}
