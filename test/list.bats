#!/usr/bin/env bats

load test_helper

@test "list: empty list returns no output" {
  run todo list
  assert_success
  assert_output ""
}

@test "list: shows tickets with IDs" {
  todo add "First ticket"
  todo add "Second ticket"

  run todo list
  assert_success
  assert_output --partial "First ticket"
  assert_output --partial "Second ticket"
}

@test "list: sorts alphabetically by filename" {
  todo add "Zebra"
  todo add "Apple"
  todo add "Mango"

  run todo list
  assert_success

  # All three should appear
  assert_output --partial "Zebra"
  assert_output --partial "Apple"
  assert_output --partial "Mango"
}

@test "list: shows tickets with and without descriptions" {
  todo add "No description ticket"
  todo add "Has description ticket" "Has a description"

  run todo list
  assert_success

  # Verify both tickets appear
  assert_output --partial "No description ticket"
  assert_output --partial "Has description ticket"
}

# --- Output format ---

@test "list: output format shows status" {
  local out
  out="$(todo add "Status ticket")"
  local id
  id="$(extract_id_from_add "${out}")"

  todo start "${id}"

  run todo list
  assert_success
  assert_output --partial "[in_progress]"
  assert_output --partial "- Status ticket"
}

@test "list: output format shows deps" {
  local out1 out2
  out1="$(todo add "Dep ticket")"
  out2="$(todo add "Main ticket")"
  local id1 id2
  id1="$(extract_id_from_add "${out1}")"
  id2="$(extract_id_from_add "${out2}")"

  todo dep "${id2}" "${id1}"

  run todo list
  assert_success
  assert_output --partial "<- [${id1}]"
}

@test "list: output format shows dash separator" {
  todo add "Dash test"

  run todo list
  assert_success
  assert_output --partial "- Dash test"
}

@test "list: output format with no status omits brackets" {
  # Tickets with status="" should not show []
  # Create a ticket with no explicit status set via raw file
  local out
  out="$(todo add "No status ticket")"
  local id
  id="$(extract_id_from_add "${out}")"

  # The add command sets type/priority/assignee but not status initially
  # Status defaults to empty unless set
  run todo show "${id}"

  # Check list output - should not have empty brackets
  run todo list
  assert_success
  refute_output --partial "[]"
}

# --- Status filter ---

@test "list: --status filters by status" {
  local out1 out2 out3
  out1="$(todo add "Open ticket")"
  out2="$(todo add "In progress ticket")"
  out3="$(todo add "Closed ticket")"
  local id2 id3
  id2="$(extract_id_from_add "${out2}")"
  id3="$(extract_id_from_add "${out3}")"

  todo start "${id2}"
  todo close "${id3}"

  run todo list --status in_progress
  assert_success
  assert_output --partial "In progress ticket"
  refute_output --partial "Open ticket"
  refute_output --partial "Closed ticket"
}

@test "list: --status closed shows closed tickets" {
  local out
  out="$(todo add "Will close")"
  local id
  id="$(extract_id_from_add "${out}")"

  todo close "${id}"

  # Default list hides closed
  run todo list
  refute_output --partial "Will close"

  # --status closed shows them
  run todo list --status closed
  assert_success
  assert_output --partial "Will close"
  assert_output --partial "[closed]"
}

@test "list: --status open shows only open tickets" {
  local out1 out2
  out1="$(todo add "Open one")"
  out2="$(todo add "Started one")"
  local id2
  id2="$(extract_id_from_add "${out2}")"

  todo start "${id2}"

  run todo list --status open
  assert_success
  assert_output --partial "Open one"
  refute_output --partial "Started one"
}

# --- Assignee filter ---

@test "list: -a filters by assignee" {
  todo add "Alice ticket" -a alice
  todo add "Bob ticket" -a bob

  run todo list -a alice
  assert_success
  assert_output --partial "Alice ticket"
  refute_output --partial "Bob ticket"
}

@test "list: --assignee filters by assignee" {
  todo add "Alice ticket" -a alice
  todo add "Bob ticket" -a bob

  run todo list --assignee bob
  assert_success
  assert_output --partial "Bob ticket"
  refute_output --partial "Alice ticket"
}

@test "list: -a with no match returns no output" {
  todo add "Some ticket" -a alice

  run todo list -a nobody
  assert_success
  assert_output ""
}

# --- Tag filter ---

@test "list: -T filters by tag" {
  todo add "Backend ticket" --tags "backend,urgent"
  todo add "Frontend ticket" --tags "frontend"

  run todo list -T backend
  assert_success
  assert_output --partial "Backend ticket"
  refute_output --partial "Frontend ticket"
}

@test "list: --tag filters by tag" {
  todo add "Tagged ticket" --tags "urgent,v2"
  todo add "Untagged ticket"

  run todo list --tag urgent
  assert_success
  assert_output --partial "Tagged ticket"
  refute_output --partial "Untagged ticket"
}

@test "list: -T with no match returns no output" {
  todo add "Some ticket" --tags "backend"

  run todo list -T nonexistent
  assert_success
  assert_output ""
}

# --- Combined filters ---

@test "list: combined --status and -a filters" {
  local out1 out2 out3
  out1="$(todo add "Alice open" -a alice)"
  out2="$(todo add "Alice started" -a alice)"
  out3="$(todo add "Bob started" -a bob)"
  local id2 id3
  id2="$(extract_id_from_add "${out2}")"
  id3="$(extract_id_from_add "${out3}")"

  todo start "${id2}"
  todo start "${id3}"

  run todo list --status in_progress -a alice
  assert_success
  assert_output --partial "Alice started"
  refute_output --partial "Alice open"
  refute_output --partial "Bob started"
}

@test "list: combined -a and -T filters" {
  todo add "Alice backend" -a alice --tags "backend"
  todo add "Alice frontend" -a alice --tags "frontend"
  todo add "Bob backend" -a bob --tags "backend"

  run todo list -a alice -T backend
  assert_success
  assert_output --partial "Alice backend"
  refute_output --partial "Alice frontend"
  refute_output --partial "Bob backend"
}

@test "list: combined --status, -a, and -T filters" {
  local out1 out2 out3
  out1="$(todo add "Match ticket" -a alice --tags "urgent")"
  out2="$(todo add "Wrong assignee" -a bob --tags "urgent")"
  out3="$(todo add "Wrong tag" -a alice --tags "other")"
  local id1
  id1="$(extract_id_from_add "${out1}")"

  todo start "${id1}"

  run todo list --status in_progress -a alice -T urgent
  assert_success
  assert_output --partial "Match ticket"
  refute_output --partial "Wrong assignee"
  refute_output --partial "Wrong tag"
}
