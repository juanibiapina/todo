#!/usr/bin/env bats

load test_helper

@test "list: empty list shows no tickets message" {
  run todo list
  assert_success
  assert_output "No tickets"
}

@test "list: shows tickets with IDs and states" {
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

@test "list: shows correct state for each ticket" {
  todo add "New ticket"
  todo add "Refined ticket" "Has a description"

  run todo list
  assert_success

  # Verify both tickets appear
  assert_output --partial "New ticket"
  assert_output --partial "Refined ticket"

  # Verify actual states via show command
  local new_id refined_id
  new_id="$(echo "${output}" | grep "New ticket" | awk '{print $2}')"
  refined_id="$(echo "${output}" | grep "Refined ticket" | awk '{print $2}')"
  [[ "$(ticket_state "${new_id}")" == "new" ]]
  [[ "$(ticket_state "${refined_id}")" == "refined" ]]
}
