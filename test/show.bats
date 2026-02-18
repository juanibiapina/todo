#!/usr/bin/env bats

load test_helper

@test "show: displays ticket by ID" {
  local out
  out="$(todo add "Show me")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"

  run todo show "${id}"
  assert_success
  assert_output --partial "Show me"
  assert_output --partial "id: ${id}"
}

@test "show: displays description" {
  local out
  out="$(todo add "Described ticket" "This is the description")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"

  run todo show "${id}"
  assert_success
  assert_output --partial "Described ticket"
  assert_output --partial "This is the description"
}

@test "show: nonexistent ID returns error" {
  run todo show "ZZZ"
  assert_failure
  assert_output --partial "ticket not found"
}

@test "show: outputs full markdown format" {
  local out
  out="$(todo add "Markdown ticket")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"

  run todo show "${id}"
  assert_success
  assert_output --partial "# Markdown ticket"
  assert_output --partial "---"
}

@test "show: displays Blockers section for unclosed deps" {
  local out1 out2
  out1="$(todo add "Blocker task")"
  local dep_id
  dep_id="$(echo "${out1}" | awk '{print $2}')"

  out2="$(todo add "Blocked task")"
  local id
  id="$(echo "${out2}" | awk '{print $2}')"

  todo dep "${id}" "${dep_id}"

  run todo show "${id}"
  assert_success
  assert_output --partial "## Blockers"
  assert_output --partial "${dep_id}"
  assert_output --partial "Blocker task"
}

@test "show: no Blockers section when deps are closed" {
  local out1 out2
  out1="$(todo add "Closed dep")"
  local dep_id
  dep_id="$(echo "${out1}" | awk '{print $2}')"

  out2="$(todo add "Main task")"
  local id
  id="$(echo "${out2}" | awk '{print $2}')"

  todo dep "${id}" "${dep_id}"
  todo done "${dep_id}"

  run todo show "${id}"
  assert_success
  refute_output --partial "## Blockers"
}

@test "show: displays Blocking section for reverse deps" {
  local out1 out2
  out1="$(todo add "Upstream task")"
  local id
  id="$(echo "${out1}" | awk '{print $2}')"

  out2="$(todo add "Downstream task")"
  local blocked_id
  blocked_id="$(echo "${out2}" | awk '{print $2}')"

  todo dep "${blocked_id}" "${id}"

  run todo show "${id}"
  assert_success
  assert_output --partial "## Blocking"
  assert_output --partial "${blocked_id}"
  assert_output --partial "Downstream task"
}

@test "show: no Blocking section when ticket is closed" {
  local out1 out2
  out1="$(todo add "Done task")"
  local id
  id="$(echo "${out1}" | awk '{print $2}')"

  out2="$(todo add "Depends on done")"
  local blocked_id
  blocked_id="$(echo "${out2}" | awk '{print $2}')"

  todo dep "${blocked_id}" "${id}"
  todo done "${id}"

  run todo show "${id}"
  assert_success
  refute_output --partial "## Blocking"
}

@test "show: displays Children section" {
  local out1
  out1="$(todo add "Parent epic")"
  local parent_id
  parent_id="$(echo "${out1}" | awk '{print $2}')"

  local out2
  out2="$(todo add "Child task" --parent "${parent_id}")"
  local child_id
  child_id="$(echo "${out2}" | awk '{print $2}')"

  run todo show "${parent_id}"
  assert_success
  assert_output --partial "## Children"
  assert_output --partial "${child_id}"
  assert_output --partial "Child task"
}

@test "show: displays Linked section" {
  local out1 out2
  out1="$(todo add "First ticket")"
  local id1
  id1="$(echo "${out1}" | awk '{print $2}')"

  out2="$(todo add "Second ticket")"
  local id2
  id2="$(echo "${out2}" | awk '{print $2}')"

  todo link "${id1}" "${id2}"

  run todo show "${id1}"
  assert_success
  assert_output --partial "## Linked"
  assert_output --partial "${id2}"
  assert_output --partial "Second ticket"
}

@test "show: enhances parent line with title" {
  local out1
  out1="$(todo add "Parent ticket")"
  local parent_id
  parent_id="$(echo "${out1}" | awk '{print $2}')"

  local out2
  out2="$(todo add "Child ticket" --parent "${parent_id}")"
  local child_id
  child_id="$(echo "${out2}" | awk '{print $2}')"

  run todo show "${child_id}"
  assert_success
  assert_output --partial "parent: ${parent_id} (Parent ticket)"
}

@test "show: no relation sections when ticket has none" {
  local out
  out="$(todo add "Lonely ticket")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"

  run todo show "${id}"
  assert_success
  refute_output --partial "## Blockers"
  refute_output --partial "## Blocking"
  refute_output --partial "## Children"
  refute_output --partial "## Linked"
}

@test "show: TODO_PAGER pipes through pager" {
  local out
  out="$(todo add "Pager test")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"

  # Use cat as the pager — output should be identical
  run env TODO_PAGER=cat todo show "${id}"
  assert_success
  assert_output --partial "Pager test"
  assert_output --partial "id: ${id}"
}

@test "show: no pager when stdout is not a TTY" {
  local out
  out="$(todo add "Pipe test")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"

  # Piping to cat means stdout is not a TTY — pager should not be used
  local result
  result="$(TODO_PAGER=cat todo show "${id}" | cat)"
  [[ "${result}" == *"Pipe test"* ]]
  [[ "${result}" == *"id: ${id}"* ]]
}
