#!/usr/bin/env bats

load test_helper

@test "closed: empty list returns no output" {
  run todo closed
  assert_success
  assert_output ""
}

@test "closed: shows closed tickets" {
  local out
  out="$(todo add "Closed ticket")"
  local id
  id="$(extract_id_from_add "${out}")"

  todo close "${id}"

  run todo closed
  assert_success
  assert_output --partial "Closed ticket"
}

@test "closed: hides open tickets" {
  todo add "Open ticket"

  run todo closed
  assert_success
  assert_output ""
}

@test "closed: hides in_progress tickets" {
  local out
  out="$(todo add "Started ticket")"
  local id
  id="$(extract_id_from_add "${out}")"

  todo start "${id}"

  run todo closed
  assert_success
  assert_output ""
}

@test "closed: sorted by mtime descending" {
  local out1 out2 out3
  out1="$(todo add "First closed")"
  out2="$(todo add "Second closed")"
  out3="$(todo add "Third closed")"
  local id1 id2 id3
  id1="$(extract_id_from_add "${out1}")"
  id2="$(extract_id_from_add "${out2}")"
  id3="$(extract_id_from_add "${out3}")"

  # Close in order, using touch to set specific mtimes
  todo close "${id1}"
  touch -t 202501010100 "${TODO_TEST_DIR}/docs/tickets/${id1}.md"

  todo close "${id2}"
  touch -t 202501010200 "${TODO_TEST_DIR}/docs/tickets/${id2}.md"

  todo close "${id3}"
  touch -t 202501010300 "${TODO_TEST_DIR}/docs/tickets/${id3}.md"

  run todo closed
  assert_success

  # Most recent (id3) should be first, oldest (id1) last
  local line1 line2 line3
  line1="$(echo "${output}" | sed -n '1p')"
  line2="$(echo "${output}" | sed -n '2p')"
  line3="$(echo "${output}" | sed -n '3p')"

  [[ "${line1}" == *"Third closed"* ]]
  [[ "${line2}" == *"Second closed"* ]]
  [[ "${line3}" == *"First closed"* ]]
}

@test "closed: --limit restricts output count" {
  local out1 out2 out3
  out1="$(todo add "Ticket A")"
  out2="$(todo add "Ticket B")"
  out3="$(todo add "Ticket C")"
  local id1 id2 id3
  id1="$(extract_id_from_add "${out1}")"
  id2="$(extract_id_from_add "${out2}")"
  id3="$(extract_id_from_add "${out3}")"

  todo close "${id1}"
  todo close "${id2}"
  todo close "${id3}"

  run todo closed --limit 2
  assert_success
  local line_count
  line_count="$(echo "${output}" | wc -l | tr -d ' ')"
  [[ "${line_count}" -eq 2 ]]
}

@test "closed: -n shorthand for --limit" {
  local out1 out2 out3
  out1="$(todo add "Ticket A")"
  out2="$(todo add "Ticket B")"
  out3="$(todo add "Ticket C")"
  local id1 id2 id3
  id1="$(extract_id_from_add "${out1}")"
  id2="$(extract_id_from_add "${out2}")"
  id3="$(extract_id_from_add "${out3}")"

  todo close "${id1}"
  todo close "${id2}"
  todo close "${id3}"

  run todo closed -n 1
  assert_success
  local line_count
  line_count="$(echo "${output}" | wc -l | tr -d ' ')"
  [[ "${line_count}" -eq 1 ]]
}

@test "closed: -a filters by assignee" {
  local out1 out2
  out1="$(todo add "Alice task" -a alice)"
  out2="$(todo add "Bob task" -a bob)"
  local id1 id2
  id1="$(extract_id_from_add "${out1}")"
  id2="$(extract_id_from_add "${out2}")"

  todo close "${id1}"
  todo close "${id2}"

  run todo closed -a alice
  assert_success
  assert_output --partial "Alice task"
  refute_output --partial "Bob task"
}

@test "closed: -T filters by tag" {
  local out1 out2
  out1="$(todo add "Backend task" --tags "backend,urgent")"
  out2="$(todo add "Frontend task" --tags "frontend")"
  local id1 id2
  id1="$(extract_id_from_add "${out1}")"
  id2="$(extract_id_from_add "${out2}")"

  todo close "${id1}"
  todo close "${id2}"

  run todo closed -T backend
  assert_success
  assert_output --partial "Backend task"
  refute_output --partial "Frontend task"
}

@test "closed: output format is id - Title without status" {
  local out
  out="$(todo add "Format test")"
  local id
  id="$(extract_id_from_add "${out}")"

  todo close "${id}"

  run todo closed
  assert_success
  assert_output --partial "- Format test"
  # Status should NOT be shown since all displayed are closed
  refute_output --partial "[closed]"
}

@test "closed: combined filters" {
  local out1 out2 out3
  out1="$(todo add "Alice backend" -a alice --tags "backend")"
  out2="$(todo add "Alice frontend" -a alice --tags "frontend")"
  out3="$(todo add "Bob backend" -a bob --tags "backend")"
  local id1 id2 id3
  id1="$(extract_id_from_add "${out1}")"
  id2="$(extract_id_from_add "${out2}")"
  id3="$(extract_id_from_add "${out3}")"

  todo close "${id1}"
  todo close "${id2}"
  todo close "${id3}"

  run todo closed -a alice -T backend
  assert_success
  assert_output --partial "Alice backend"
  refute_output --partial "Alice frontend"
  refute_output --partial "Bob backend"
}
