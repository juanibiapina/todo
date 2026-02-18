#!/usr/bin/env bats

load test_helper

@test "query: empty output when no tickets" {
  run todo query
  assert_success
  assert_output ""
}

@test "query: outputs valid JSONL" {
  run todo add "First ticket"
  assert_success
  run todo add "Second ticket"
  assert_success

  run todo query
  assert_success

  # Each line must be valid JSON
  local line_count=0
  while IFS= read -r line; do
    echo "${line}" | jq . > /dev/null 2>&1
    [ $? -eq 0 ]
    line_count=$((line_count + 1))
  done <<< "${output}"
  [ "${line_count}" -eq 2 ]
}

@test "query: includes id, title, and priority fields" {
  run todo add "My ticket"
  assert_success
  local id
  id=$(extract_id_from_add "${output}")

  run todo query
  assert_success

  local tid title priority
  tid=$(echo "${output}" | jq -r '.id')
  title=$(echo "${output}" | jq -r '.title')
  priority=$(echo "${output}" | jq -r '.priority')

  [ "${tid}" = "${id}" ]
  [ "${title}" = "My ticket" ]
  [ "${priority}" = "2" ]
}

@test "query: includes all frontmatter fields" {
  run todo add "Full ticket" -t bug -p 1 -a Alice --external-ref "EXT-1" --design "Design doc" --acceptance "AC list" --tags "urgent,backend"
  assert_success
  local id
  id=$(extract_id_from_add "${output}")

  run todo start "${id}"
  assert_success

  run todo query
  assert_success

  [ "$(echo "${output}" | jq -r '.id')" = "${id}" ]
  [ "$(echo "${output}" | jq -r '.title')" = "Full ticket" ]
  [ "$(echo "${output}" | jq -r '.status')" = "in_progress" ]
  [ "$(echo "${output}" | jq -r '.type')" = "bug" ]
  [ "$(echo "${output}" | jq -r '.priority')" = "1" ]
  [ "$(echo "${output}" | jq -r '.assignee')" = "Alice" ]
  [ "$(echo "${output}" | jq -r '.external_ref')" = "EXT-1" ]
  [ "$(echo "${output}" | jq -r '.design')" = "Design doc" ]
  [ "$(echo "${output}" | jq -r '.acceptance')" = "AC list" ]
  [ "$(echo "${output}" | jq -r '.tags | length')" = "2" ]
  [ "$(echo "${output}" | jq -r '.tags[0]')" = "urgent" ]
  [ "$(echo "${output}" | jq -r '.tags[1]')" = "backend" ]
}

@test "query: slices are arrays not null when empty" {
  run todo add "Simple ticket"
  assert_success

  run todo query
  assert_success

  [ "$(echo "${output}" | jq -r '.deps | type')" = "array" ]
  [ "$(echo "${output}" | jq -r '.links | type')" = "array" ]
  [ "$(echo "${output}" | jq -r '.tags | type')" = "array" ]
  [ "$(echo "${output}" | jq -r '.deps | length')" = "0" ]
  [ "$(echo "${output}" | jq -r '.links | length')" = "0" ]
  [ "$(echo "${output}" | jq -r '.tags | length')" = "0" ]
}

@test "query: includes deps and links" {
  run todo add "Ticket A"
  assert_success
  local id_a
  id_a=$(extract_id_from_add "${output}")

  run todo add "Ticket B"
  assert_success
  local id_b
  id_b=$(extract_id_from_add "${output}")

  run todo dep "${id_a}" "${id_b}"
  assert_success

  run todo link "${id_a}" "${id_b}"
  assert_success

  # Query and check ticket A
  run todo query
  assert_success

  local a_line
  a_line=$(echo "${output}" | jq -r "select(.id == \"${id_a}\")")

  [ "$(echo "${a_line}" | jq -r '.deps | length')" = "1" ]
  [ "$(echo "${a_line}" | jq -r '.deps[0]')" = "${id_b}" ]
  [ "$(echo "${a_line}" | jq -r '.links | length')" = "1" ]
  [ "$(echo "${a_line}" | jq -r '.links[0]')" = "${id_b}" ]
}

@test "query: --status filter" {
  run todo add "Open ticket"
  assert_success

  run todo add "Closed ticket"
  assert_success
  local closed_id
  closed_id=$(extract_id_from_add "${output}")
  run todo done "${closed_id}"
  assert_success

  run todo query --status closed
  assert_success

  local count
  count=$(echo "${output}" | wc -l | tr -d ' ')
  [ "${count}" -eq 1 ]
  [ "$(echo "${output}" | jq -r '.title')" = "Closed ticket" ]
}

@test "query: --status open matches empty status" {
  run todo add "New ticket"
  assert_success

  run todo query --status open
  assert_success

  local count
  count=$(echo "${output}" | wc -l | tr -d ' ')
  [ "${count}" -eq 1 ]
  [ "$(echo "${output}" | jq -r '.title')" = "New ticket" ]
}

@test "query: --type filter" {
  run todo add "Bug ticket" -t bug
  assert_success
  run todo add "Feature ticket" -t feature
  assert_success

  run todo query --type bug
  assert_success

  local count
  count=$(echo "${output}" | wc -l | tr -d ' ')
  [ "${count}" -eq 1 ]
  [ "$(echo "${output}" | jq -r '.title')" = "Bug ticket" ]
}

@test "query: --assignee filter" {
  run todo add "Alice ticket" -a Alice
  assert_success
  run todo add "Bob ticket" -a Bob
  assert_success

  run todo query --assignee Alice
  assert_success

  local count
  count=$(echo "${output}" | wc -l | tr -d ' ')
  [ "${count}" -eq 1 ]
  [ "$(echo "${output}" | jq -r '.title')" = "Alice ticket" ]
}

@test "query: --tag filter" {
  run todo add "Tagged ticket" --tags "urgent,backend"
  assert_success
  run todo add "Other ticket" --tags "frontend"
  assert_success

  run todo query --tag urgent
  assert_success

  local count
  count=$(echo "${output}" | wc -l | tr -d ' ')
  [ "${count}" -eq 1 ]
  [ "$(echo "${output}" | jq -r '.title')" = "Tagged ticket" ]
}

@test "query: combined filters" {
  run todo add "Match ticket" -t bug -a Alice --tags "urgent"
  assert_success
  run todo add "Wrong type" -t feature -a Alice --tags "urgent"
  assert_success
  run todo add "Wrong assignee" -t bug -a Bob --tags "urgent"
  assert_success

  run todo query --type bug --assignee Alice --tag urgent
  assert_success

  local count
  count=$(echo "${output}" | wc -l | tr -d ' ')
  [ "${count}" -eq 1 ]
  [ "$(echo "${output}" | jq -r '.title')" = "Match ticket" ]
}

@test "query: includes description" {
  run todo add "With desc" -d "Some description text"
  assert_success

  run todo query
  assert_success

  [ "$(echo "${output}" | jq -r '.description')" = "Some description text" ]
}
