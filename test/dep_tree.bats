#!/usr/bin/env bats

load test_helper

@test "dep tree shows single ticket with no deps" {
  run todo add "Root ticket"
  local id
  id="$(extract_id_from_add "$output")"

  run todo dep tree "${id}"
  assert_success
  assert_output --partial "${id}"
  assert_output --partial "Root ticket"
}

@test "dep tree shows simple dependency" {
  run todo add "Root"
  local root_id
  root_id="$(extract_id_from_add "$output")"

  run todo add "Child"
  local child_id
  child_id="$(extract_id_from_add "$output")"

  run todo dep "${root_id}" "${child_id}"

  run todo dep tree "${root_id}"
  assert_success
  assert_output --partial "└── ${child_id}"
  assert_output --partial "Child"
}

@test "dep tree shows nested dependencies" {
  run todo add "Root"
  local root_id
  root_id="$(extract_id_from_add "$output")"

  run todo add "Mid"
  local mid_id
  mid_id="$(extract_id_from_add "$output")"

  run todo add "Leaf"
  local leaf_id
  leaf_id="$(extract_id_from_add "$output")"

  run todo dep "${root_id}" "${mid_id}"
  run todo dep "${mid_id}" "${leaf_id}"

  run todo dep tree "${root_id}"
  assert_success
  assert_output --partial "${root_id}"
  assert_output --partial "└── ${mid_id}"
  assert_output --partial "    └── ${leaf_id}"
}

@test "dep tree handles cycles" {
  run todo add "A"
  local id_a
  id_a="$(extract_id_from_add "$output")"

  run todo add "B"
  local id_b
  id_b="$(extract_id_from_add "$output")"

  run todo dep "${id_a}" "${id_b}"
  run todo dep "${id_b}" "${id_a}"

  run todo dep tree "${id_a}"
  assert_success
  assert_output --partial "(cycle)"
}

@test "dep tree deduplicates by default" {
  run todo add "Root"
  local root_id
  root_id="$(extract_id_from_add "$output")"

  run todo add "A"
  local id_a
  id_a="$(extract_id_from_add "$output")"

  run todo add "Shared"
  local shared_id
  shared_id="$(extract_id_from_add "$output")"

  run todo dep "${root_id}" "${id_a}"
  run todo dep "${root_id}" "${shared_id}"
  run todo dep "${id_a}" "${shared_id}"

  run todo dep tree "${root_id}"
  assert_success
  assert_output --partial "(dup)"
}

@test "dep tree --full disables deduplication" {
  run todo add "Root"
  local root_id
  root_id="$(extract_id_from_add "$output")"

  run todo add "A"
  local id_a
  id_a="$(extract_id_from_add "$output")"

  run todo add "Shared"
  local shared_id
  shared_id="$(extract_id_from_add "$output")"

  run todo dep "${root_id}" "${id_a}"
  run todo dep "${root_id}" "${shared_id}"
  run todo dep "${id_a}" "${shared_id}"

  run todo dep tree --full "${root_id}"
  assert_success
  refute_output --partial "(dup)"
}

@test "dep tree shows status" {
  run todo add "Root"
  local id
  id="$(extract_id_from_add "$output")"

  run todo start "${id}"

  run todo dep tree "${id}"
  assert_success
  assert_output --partial "[in_progress]"
}

@test "dep tree fails for nonexistent ticket" {
  run todo dep tree "zzz"
  assert_failure
  assert_output --partial "ticket not found"
}

@test "dep tree supports partial IDs" {
  run todo add "Root"
  local root_id
  root_id="$(extract_id_from_add "$output")"

  local partial="${root_id:0:2}"

  run todo dep tree "${partial}"
  assert_success
  assert_output --partial "${root_id}"
}
