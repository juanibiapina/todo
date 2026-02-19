#!/usr/bin/env bats

load test_helper

@test "add: creates a ticket" {
  run todo add "Fix the bug"
  assert_success
  assert_output --partial "Added"
  assert_output --partial "Fix the bug"
}

@test "add: ticket appears in list" {
  todo add "Fix the bug"
  run todo list
  assert_success
  assert_output --partial "Fix the bug"
}

@test "add: with description as positional arg" {
  run todo add "Fix auth" "The auth handler needs work"
  assert_success
  assert_output --partial "Added"
  assert_output --partial "Fix auth"

  # Verify the description was stored
  local id
  id="$(extract_id_from_add "${output}")"
  run todo show "${id}"
  assert_output --partial "The auth handler needs work"
}

@test "add: with description via stdin" {
  run bash -c 'echo "Description from stdin" | todo add "Stdin ticket"'
  assert_success
  assert_output --partial "Added"

  local id
  id="$(extract_id_from_add "${output}")"
  run todo show "${id}"
  assert_output --partial "Description from stdin"
}

@test "add: generates a 3-character ID" {
  run todo add "Test ticket"
  assert_success
  local id
  id="$(extract_id_from_add "${output}")"
  [[ "${#id}" -eq 3 ]]
  [[ "${id}" =~ ^[A-Za-z0-9]{3}$ ]]
}

@test "add: duplicate titles are allowed" {
  run todo add "Same title"
  assert_success
  run todo add "Same title"
  assert_success

  count="$(ticket_count)"
  [[ "${count}" -eq 2 ]]
}

@test "add: creates tickets directory if it doesn't exist" {
  [[ ! -d docs/tickets ]]
  todo add "First ticket"
  [[ -d docs/tickets ]]
}

@test "add: multiple tickets get unique IDs" {
  todo add "Ticket one"
  todo add "Ticket two"
  todo add "Ticket three"

  # Extract IDs from list output (field 1 is the ID)
  local ids
  ids="$(todo list | awk '{print $1}' | sort -u)"
  local count
  count="$(echo "${ids}" | wc -l | tr -d ' ')"
  [[ "${count}" -eq 3 ]]
}

# --- Default title ---

@test "add: defaults to Untitled when no args" {
  run todo add
  assert_success
  assert_output --partial "Added"
  assert_output --partial "Untitled"
}

# --- Description flag ---

@test "add: -d flag sets description" {
  run todo add "Flag desc ticket" -d "Description via flag"
  assert_success

  local id
  id="$(extract_id_from_add "${output}")"
  run todo show "${id}"
  assert_output --partial "Description via flag"
}

@test "add: -d flag takes priority over positional description" {
  run todo add "Priority test" "positional desc" -d "flag desc"
  assert_success

  local id
  id="$(extract_id_from_add "${output}")"
  run todo show "${id}"
  assert_output --partial "flag desc"
  refute_output --partial "positional desc"
}

# --- Type flag ---

@test "add: default type is task" {
  run todo add "Default type ticket"
  assert_success

  local id
  id="$(extract_id_from_add "${output}")"
  run todo show "${id}"
  assert_output --partial "type: task"
}

@test "add: -t flag sets type" {
  run todo add "Bug ticket" -t bug
  assert_success

  local id
  id="$(extract_id_from_add "${output}")"
  run todo show "${id}"
  assert_output --partial "type: bug"
}

@test "add: all valid types accepted" {
  for type in bug feature task epic chore; do
    run todo add "Type ${type}" -t "${type}"
    assert_success
  done
}

@test "add: invalid type returns error" {
  run todo add "Bad type" -t invalid
  assert_failure
  assert_output --partial "invalid type"
}

# --- Priority flag ---

@test "add: default priority is 2" {
  run todo add "Default priority ticket"
  assert_success

  local id
  id="$(extract_id_from_add "${output}")"
  run todo show "${id}"
  assert_output --partial "priority: 2"
}

@test "add: -p flag sets priority" {
  run todo add "High priority" -p 4
  assert_success

  local id
  id="$(extract_id_from_add "${output}")"
  run todo show "${id}"
  assert_output --partial "priority: 4"
}

@test "add: priority 0 is valid" {
  run todo add "Zero priority" -p 0
  assert_success
}

@test "add: priority out of range returns error" {
  run todo add "Bad priority" -p 5
  assert_failure
  assert_output --partial "invalid priority"

  run todo add "Bad priority" -p -1
  assert_failure
  assert_output --partial "invalid priority"
}

# --- Assignee flag ---

@test "add: default assignee is git user.name" {
  run todo add "Assigned ticket"
  assert_success

  local id
  id="$(extract_id_from_add "${output}")"
  run todo show "${id}"
  assert_output --partial "assignee: Test User"
}

@test "add: -a flag overrides default assignee" {
  run todo add "Custom assignee" -a "bob"
  assert_success

  local id
  id="$(extract_id_from_add "${output}")"
  run todo show "${id}"
  assert_output --partial "assignee: bob"
}

# --- External ref flag ---

@test "add: --external-ref sets external reference" {
  run todo add "With ref" --external-ref "JIRA-456"
  assert_success

  local id
  id="$(extract_id_from_add "${output}")"
  run todo show "${id}"
  assert_output --partial "external_ref: JIRA-456"
}

# --- Parent flag ---

@test "add: --parent sets parent ticket" {
  todo add "Parent ticket"
  local parent_id
  parent_id="$(todo list | awk '{print $1}')"

  run todo add "Child ticket" --parent "${parent_id}"
  assert_success

  local id
  id="$(extract_id_from_add "${output}")"
  run todo show "${id}"
  assert_output --partial "parent: ${parent_id}"
}

@test "add: --parent with non-existent ID returns error" {
  run todo add "Bad parent" --parent "zzz"
  assert_failure
  assert_output --partial "parent ticket not found"
}

# --- Design flag ---

@test "add: --design sets design notes" {
  run todo add "With design" --design "Use microservices architecture"
  assert_success

  local id
  id="$(extract_id_from_add "${output}")"
  run todo show "${id}"
  assert_output --partial "design: Use microservices architecture"
}

# --- Acceptance flag ---

@test "add: --acceptance sets acceptance criteria" {
  run todo add "With acceptance" --acceptance "All tests pass"
  assert_success

  local id
  id="$(extract_id_from_add "${output}")"
  run todo show "${id}"
  assert_output --partial "acceptance: All tests pass"
}

# --- Tags flag ---

@test "add: --tags sets comma-separated tags" {
  run todo add "Tagged ticket" --tags "backend,urgent,v2"
  assert_success

  local id
  id="$(extract_id_from_add "${output}")"
  run todo show "${id}"
  assert_output --partial "backend"
  assert_output --partial "urgent"
  assert_output --partial "v2"
}

@test "add: --tags handles spaces around commas" {
  run todo add "Spaced tags" --tags "tag1 , tag2 , tag3"
  assert_success

  local id
  id="$(extract_id_from_add "${output}")"
  run todo show "${id}"
  assert_output --partial "tag1"
  assert_output --partial "tag2"
  assert_output --partial "tag3"
}

# --- Combined flags ---

@test "add: multiple flags together" {
  todo add "Parent for combo"
  local parent_id
  parent_id="$(todo list | awk '{print $1}')"

  run todo add "Full ticket" \
    -d "Full description" \
    -t feature \
    -p 3 \
    -a alice \
    --external-ref "JIRA-789" \
    --parent "${parent_id}" \
    --design "REST API" \
    --acceptance "200 OK responses" \
    --tags "api,backend"
  assert_success

  local id
  id="$(extract_id_from_add "${output}")"
  run todo show "${id}"
  assert_output --partial "type: feature"
  assert_output --partial "priority: 3"
  assert_output --partial "assignee: alice"
  assert_output --partial "external_ref: JIRA-789"
  assert_output --partial "parent: ${parent_id}"
  assert_output --partial "design: REST API"
  assert_output --partial "acceptance: 200 OK responses"
  assert_output --partial "Full description"
  assert_output --partial "api"
  assert_output --partial "backend"
}
