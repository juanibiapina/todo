#!/usr/bin/env bats

load test_helper

@test "ready: empty list returns no output" {
  run todo ready
  assert_success
  assert_output ""
}

@test "ready: shows ticket with no deps" {
  todo add "No deps ticket"

  run todo ready
  assert_success
  assert_output --partial "No deps ticket"
}

@test "ready: shows ticket when all deps are closed" {
  local out1 out2
  out1="$(todo add "Dep ticket")"
  out2="$(todo add "Main ticket")"
  local id1 id2
  id1="$(extract_id_from_add "${out1}")"
  id2="$(extract_id_from_add "${out2}")"

  todo dep "${id2}" "${id1}"
  todo close "${id1}"

  run todo ready
  assert_success
  assert_output --partial "Main ticket"
}

@test "ready: hides ticket with unclosed deps" {
  local out1 out2
  out1="$(todo add "Blocker")"
  out2="$(todo add "Blocked ticket")"
  local id1 id2
  id1="$(extract_id_from_add "${out1}")"
  id2="$(extract_id_from_add "${out2}")"

  todo dep "${id2}" "${id1}"

  run todo ready
  assert_success
  # Blocked ticket should NOT appear (has unclosed dep)
  refute_output --partial "Blocked ticket"
  # Blocker itself has no deps, so it should appear
  assert_output --partial "Blocker"
}

@test "ready: hides closed tickets" {
  local out
  out="$(todo add "Closed ticket")"
  local id
  id="$(extract_id_from_add "${out}")"

  todo close "${id}"

  run todo ready
  assert_success
  refute_output --partial "Closed ticket"
}

@test "ready: shows in_progress tickets" {
  local out
  out="$(todo add "In progress ticket")"
  local id
  id="$(extract_id_from_add "${out}")"

  todo start "${id}"

  run todo ready
  assert_success
  assert_output --partial "In progress ticket"
  assert_output --partial "[in_progress]"
}

@test "ready: sorted by priority ascending then ID" {
  local out1 out2 out3
  out1="$(todo add "Low priority" -p 4)"
  out2="$(todo add "High priority" -p 0)"
  out3="$(todo add "Medium priority" -p 2)"

  run todo ready
  assert_success

  # All three should appear
  assert_output --partial "High priority"
  assert_output --partial "Medium priority"
  assert_output --partial "Low priority"

  # Check ordering: high priority (P0) should come before medium (P2) before low (P4)
  local line1 line2 line3
  line1="$(echo "${output}" | sed -n '1p')"
  line2="$(echo "${output}" | sed -n '2p')"
  line3="$(echo "${output}" | sed -n '3p')"

  [[ "${line1}" == *"[P0]"* ]]
  [[ "${line2}" == *"[P2]"* ]]
  [[ "${line3}" == *"[P4]"* ]]
}

@test "ready: output format shows priority and status" {
  local out
  out="$(todo add "Format test" -p 3)"
  local id
  id="$(extract_id_from_add "${out}")"

  run todo ready
  assert_success
  assert_output --partial "[P3]"
  assert_output --partial "[open]"
  assert_output --partial "- Format test"
}

@test "ready: -a filters by assignee" {
  todo add "Alice task" -a alice
  todo add "Bob task" -a bob

  run todo ready -a alice
  assert_success
  assert_output --partial "Alice task"
  refute_output --partial "Bob task"
}

@test "ready: -T filters by tag" {
  todo add "Backend task" --tags "backend,urgent"
  todo add "Frontend task" --tags "frontend"

  run todo ready -T backend
  assert_success
  assert_output --partial "Backend task"
  refute_output --partial "Frontend task"
}

@test "ready: empty status displayed as open" {
  # New tickets have empty status, should show as [open]
  todo add "New ticket"

  run todo ready
  assert_success
  assert_output --partial "[open]"
  refute_output --partial "[]"
}
