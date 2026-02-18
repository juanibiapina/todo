#!/usr/bin/env bats

load test_helper

@test "blocked: empty list returns no output" {
  run todo blocked
  assert_success
  assert_output ""
}

@test "blocked: ticket with no deps is not blocked" {
  todo add "No deps ticket"

  run todo blocked
  assert_success
  assert_output ""
}

@test "blocked: shows ticket with unclosed dep" {
  local out1 out2
  out1="$(todo add "Blocker")"
  out2="$(todo add "Blocked ticket")"
  local id1 id2
  id1="$(extract_id_from_add "${out1}")"
  id2="$(extract_id_from_add "${out2}")"

  todo dep "${id2}" "${id1}"

  run todo blocked
  assert_success
  assert_output --partial "Blocked ticket"
  # Only one line of output — the blocker itself has no deps so it's not blocked
  local line_count
  line_count="$(echo "${output}" | wc -l | tr -d ' ')"
  [[ "${line_count}" -eq 1 ]]
}

@test "blocked: hides ticket when all deps are closed" {
  local out1 out2
  out1="$(todo add "Dep ticket")"
  out2="$(todo add "Main ticket")"
  local id1 id2
  id1="$(extract_id_from_add "${out1}")"
  id2="$(extract_id_from_add "${out2}")"

  todo dep "${id2}" "${id1}"
  todo close "${id1}"

  run todo blocked
  assert_success
  # Main ticket has all deps closed, so it should not be blocked
  assert_output ""
}

@test "blocked: only shows unclosed blockers in output" {
  local out1 out2 out3
  out1="$(todo add "Closed dep")"
  out2="$(todo add "Open dep")"
  out3="$(todo add "Blocked ticket")"
  local id1 id2 id3
  id1="$(extract_id_from_add "${out1}")"
  id2="$(extract_id_from_add "${out2}")"
  id3="$(extract_id_from_add "${out3}")"

  todo dep "${id3}" "${id1}"
  todo dep "${id3}" "${id2}"
  todo close "${id1}"

  run todo blocked
  assert_success
  assert_output --partial "Blocked ticket"
  # Only the open dep should be shown as a blocker
  assert_output --partial "${id2}"
  refute_output --partial "${id1}"
}

@test "blocked: hides closed tickets" {
  local out1 out2
  out1="$(todo add "Dep")"
  out2="$(todo add "Ticket to close")"
  local id1 id2
  id1="$(extract_id_from_add "${out1}")"
  id2="$(extract_id_from_add "${out2}")"

  todo dep "${id2}" "${id1}"
  todo close "${id2}"

  run todo blocked
  assert_success
  # Closed ticket should not appear even if it has unclosed deps
  refute_output --partial "Ticket to close"
}

@test "blocked: shows in_progress tickets" {
  local out1 out2
  out1="$(todo add "Blocker")"
  out2="$(todo add "In progress blocked")"
  local id1 id2
  id1="$(extract_id_from_add "${out1}")"
  id2="$(extract_id_from_add "${out2}")"

  todo dep "${id2}" "${id1}"
  todo start "${id2}"

  run todo blocked
  assert_success
  assert_output --partial "In progress blocked"
  assert_output --partial "[in_progress]"
}

@test "blocked: sorted by priority ascending then ID" {
  local out_blocker out1 out2 out3
  out_blocker="$(todo add "Common blocker")"
  out1="$(todo add "Low priority" -p 4)"
  out2="$(todo add "High priority" -p 0)"
  out3="$(todo add "Medium priority" -p 2)"
  local blocker_id id1 id2 id3
  blocker_id="$(extract_id_from_add "${out_blocker}")"
  id1="$(extract_id_from_add "${out1}")"
  id2="$(extract_id_from_add "${out2}")"
  id3="$(extract_id_from_add "${out3}")"

  todo dep "${id1}" "${blocker_id}"
  todo dep "${id2}" "${blocker_id}"
  todo dep "${id3}" "${blocker_id}"

  run todo blocked
  assert_success

  # All three should appear
  assert_output --partial "High priority"
  assert_output --partial "Medium priority"
  assert_output --partial "Low priority"

  # Check ordering: P0 before P2 before P4
  local line1 line2 line3
  line1="$(echo "${output}" | sed -n '1p')"
  line2="$(echo "${output}" | sed -n '2p')"
  line3="$(echo "${output}" | sed -n '3p')"

  [[ "${line1}" == *"[P0]"* ]]
  [[ "${line2}" == *"[P2]"* ]]
  [[ "${line3}" == *"[P4]"* ]]
}

@test "blocked: output format shows priority status and blockers" {
  local out1 out2
  out1="$(todo add "Blocker task")"
  out2="$(todo add "Format test" -p 3)"
  local id1 id2
  id1="$(extract_id_from_add "${out1}")"
  id2="$(extract_id_from_add "${out2}")"

  todo dep "${id2}" "${id1}"

  run todo blocked
  assert_success
  assert_output --partial "[P3]"
  assert_output --partial "[open]"
  assert_output --partial "- Format test"
  assert_output --partial "<- [${id1}]"
}

@test "blocked: -a filters by assignee" {
  local out_blocker out1 out2
  out_blocker="$(todo add "Blocker")"
  out1="$(todo add "Alice task" -a alice)"
  out2="$(todo add "Bob task" -a bob)"
  local blocker_id id1 id2
  blocker_id="$(extract_id_from_add "${out_blocker}")"
  id1="$(extract_id_from_add "${out1}")"
  id2="$(extract_id_from_add "${out2}")"

  todo dep "${id1}" "${blocker_id}"
  todo dep "${id2}" "${blocker_id}"

  run todo blocked -a alice
  assert_success
  assert_output --partial "Alice task"
  refute_output --partial "Bob task"
}

@test "blocked: -T filters by tag" {
  local out_blocker out1 out2
  out_blocker="$(todo add "Blocker")"
  out1="$(todo add "Backend task" --tags "backend,urgent")"
  out2="$(todo add "Frontend task" --tags "frontend")"
  local blocker_id id1 id2
  blocker_id="$(extract_id_from_add "${out_blocker}")"
  id1="$(extract_id_from_add "${out1}")"
  id2="$(extract_id_from_add "${out2}")"

  todo dep "${id1}" "${blocker_id}"
  todo dep "${id2}" "${blocker_id}"

  run todo blocked -T backend
  assert_success
  assert_output --partial "Backend task"
  refute_output --partial "Frontend task"
}
