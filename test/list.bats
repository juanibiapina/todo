#!/usr/bin/env bats

load test_helper

@test "list: empty list shows no tickets message" {
  run todo list
  assert_success
  assert_output "No tickets"
}

@test "list: shows tickets with IDs" {
  todo add "First ticket"
  todo add "Second ticket"

  run todo list
  assert_success
  assert_output --partial "First ticket"
  assert_output --partial "Second ticket"
}

@test "list: preserves insertion order" {
  todo add "Alpha"
  todo add "Beta"
  todo add "Gamma"

  run todo list
  assert_success

  # Check that Alpha appears before Beta, and Beta before Gamma
  local alpha_line beta_line gamma_line
  alpha_line="$(echo "${output}" | grep -n "Alpha" | cut -d: -f1)"
  beta_line="$(echo "${output}" | grep -n "Beta" | cut -d: -f1)"
  gamma_line="$(echo "${output}" | grep -n "Gamma" | cut -d: -f1)"

  [[ "${alpha_line}" -lt "${beta_line}" ]]
  [[ "${beta_line}" -lt "${gamma_line}" ]]
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
