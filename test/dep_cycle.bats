#!/usr/bin/env bats

load test_helper

@test "dep cycle reports no output when no cycles" {
  run todo add "A"
  local id_a
  id_a="$(extract_id_from_add "$output")"

  run todo add "B"
  local id_b
  id_b="$(extract_id_from_add "$output")"

  run todo dep "${id_a}" "${id_b}"

  run todo dep cycle
  assert_success
  assert_output ""
}

@test "dep cycle detects simple two-node cycle" {
  run todo add "First"
  local id_a
  id_a="$(extract_id_from_add "$output")"

  run todo add "Second"
  local id_b
  id_b="$(extract_id_from_add "$output")"

  run todo dep "${id_a}" "${id_b}"
  run todo dep "${id_b}" "${id_a}"

  run todo dep cycle
  assert_success
  assert_output --partial "Cycle:"
  assert_output --partial "${id_a}"
  assert_output --partial "${id_b}"
  assert_output --partial "First"
  assert_output --partial "Second"
}

@test "dep cycle detects three-node cycle" {
  run todo add "A"
  local id_a
  id_a="$(extract_id_from_add "$output")"

  run todo add "B"
  local id_b
  id_b="$(extract_id_from_add "$output")"

  run todo add "C"
  local id_c
  id_c="$(extract_id_from_add "$output")"

  run todo dep "${id_a}" "${id_b}"
  run todo dep "${id_b}" "${id_c}"
  run todo dep "${id_c}" "${id_a}"

  run todo dep cycle
  assert_success
  assert_output --partial "Cycle:"
  assert_output --partial "${id_a}"
  assert_output --partial "${id_b}"
  assert_output --partial "${id_c}"
}

@test "dep cycle excludes closed tickets" {
  run todo add "Open"
  local id_a
  id_a="$(extract_id_from_add "$output")"

  run todo add "Will close"
  local id_b
  id_b="$(extract_id_from_add "$output")"

  run todo dep "${id_a}" "${id_b}"
  run todo dep "${id_b}" "${id_a}"

  # Close one ticket — should break the cycle
  run todo close "${id_b}"

  run todo dep cycle
  assert_success
  assert_output ""
}

@test "dep cycle shows status in member details" {
  run todo add "Started"
  local id_a
  id_a="$(extract_id_from_add "$output")"

  run todo add "Open"
  local id_b
  id_b="$(extract_id_from_add "$output")"

  run todo start "${id_a}"
  run todo status "${id_b}" "open"

  run todo dep "${id_a}" "${id_b}"
  run todo dep "${id_b}" "${id_a}"

  run todo dep cycle
  assert_success
  assert_output --partial "[in_progress]"
  assert_output --partial "[open]"
}

@test "dep cycle reports no output when no tickets" {
  run todo dep cycle
  assert_success
  assert_output ""
}

@test "dep cycle reports no output when tickets have no deps" {
  run todo add "Solo A"
  run todo add "Solo B"

  run todo dep cycle
  assert_success
  assert_output ""
}
