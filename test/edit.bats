#!/usr/bin/env bats

load test_helper

@test "edit: prints file path when stdout is not a TTY" {
  local out
  out="$(todo add "Edit me")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"

  # Piping through cat makes stdout non-TTY
  local result
  result="$(todo edit "${id}" | cat)"
  [[ "${result}" == *"docs/tickets/${id}.md"* ]]
}

@test "edit: resolves partial ID" {
  local out
  out="$(todo add "Partial edit")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"

  local partial="${id:0:2}"

  local result
  result="$(todo edit "${partial}" | cat)"
  [[ "${result}" == *"docs/tickets/${id}.md"* ]]
}

@test "edit: nonexistent ID returns error" {
  run todo edit "ZZZ"
  assert_failure
  assert_output --partial "ticket not found"
}

@test "edit: uses EDITOR env var in TTY mode" {
  local out
  out="$(todo add "Editor test")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"

  # Create a custom editor script that writes a marker file
  local editor_script="${BATS_TEST_TMPDIR}/test_editor.sh"
  cat > "${editor_script}" <<'SCRIPT'
#!/bin/bash
touch "${1}.edited"
SCRIPT
  chmod +x "${editor_script}"

  # Use script to simulate a TTY so the editor path is taken
  if [[ "$(uname)" == "Darwin" ]]; then
    script -q /dev/null env EDITOR="${editor_script}" todo edit "${id}" >/dev/null 2>&1 || true
  else
    script -q -c "EDITOR='${editor_script}' todo edit '${id}'" /dev/null >/dev/null 2>&1 || true
  fi

  # Verify the marker file was created next to the ticket file
  local path
  path="$(todo edit "${id}" | cat)"
  [[ -f "${path}.edited" ]]
}

@test "edit: file path points to correct ticket file" {
  local out
  out="$(todo add "Path check")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"

  local result
  result="$(todo edit "${id}" | cat)"

  # The path should end with <id>.md
  [[ "${result}" == *"/${id}.md" ]]
}

@test "edit: file is editable via the printed path" {
  local out
  out="$(todo add "Modify me")"
  local id
  id="$(echo "${out}" | awk '{print $2}')"

  # Get file path via non-TTY mode
  local path
  path="$(todo edit "${id}" | cat)"

  # Modify the file directly (simulating what an editor would do)
  sed 's/Modify/Modified/' "${path}" > "${path}.tmp" && mv "${path}.tmp" "${path}"

  run todo show "${id}"
  assert_success
  assert_output --partial "Modified me"
}

@test "edit: requires exactly one argument" {
  run todo edit
  assert_failure

  run todo edit "abc" "def"
  assert_failure
}
