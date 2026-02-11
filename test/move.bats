#!/usr/bin/env bats

load test_helper

@test "move-up: swaps ticket with previous" {
  todo add "First"
  local out
  out="$(todo add "Second")"
  local id
  id="$(echo "${out}" | awk '{print $3}')"

  run todo move-up "${id}"
  assert_success
  assert_output --partial "Moved up"

  # Second should now be first in the list
  run todo list
  local first_line
  first_line="$(echo "${output}" | head -1)"
  echo "${first_line}" | grep -q "Second"
}

@test "move-up: first ticket is a no-op" {
  local out
  out="$(todo add "Already first")"
  local id
  id="$(echo "${out}" | awk '{print $3}')"
  todo add "Second"

  run todo move-up "${id}"
  assert_success
  assert_output --partial "Moved up"

  # Order should be unchanged
  run todo list
  local first_line
  first_line="$(echo "${output}" | head -1)"
  echo "${first_line}" | grep -q "Already first"
}

@test "move-down: swaps ticket with next" {
  local out
  out="$(todo add "First")"
  local id
  id="$(echo "${out}" | awk '{print $3}')"
  todo add "Second"

  run todo move-down "${id}"
  assert_success
  assert_output --partial "Moved down"

  # First should now be second in the list
  run todo list
  local second_line
  second_line="$(echo "${output}" | tail -1)"
  echo "${second_line}" | grep -q "First"
}

@test "move-down: last ticket is a no-op" {
  todo add "First"
  local out
  out="$(todo add "Already last")"
  local id
  id="$(echo "${out}" | awk '{print $3}')"

  run todo move-down "${id}"
  assert_success
  assert_output --partial "Moved down"

  # Order should be unchanged
  run todo list
  local last_line
  last_line="$(echo "${output}" | tail -1)"
  echo "${last_line}" | grep -q "Already last"
}

@test "move-up: nonexistent ticket returns error" {
  run todo move-up "ZZZ"
  assert_failure
  assert_output --partial "ticket not found"
}

@test "move-down: nonexistent ticket returns error" {
  run todo move-down "ZZZ"
  assert_failure
  assert_output --partial "ticket not found"
}

@test "move: multiple moves reorder correctly" {
  local out1 out2 out3
  out1="$(todo add "Alpha")"
  out2="$(todo add "Beta")"
  out3="$(todo add "Gamma")"
  local id3
  id3="$(echo "${out3}" | awk '{print $3}')"

  # Move Gamma to the top: up twice
  todo move-up "${id3}"
  todo move-up "${id3}"

  run todo list
  local first_line
  first_line="$(echo "${output}" | head -1)"
  echo "${first_line}" | grep -q "Gamma"
}
