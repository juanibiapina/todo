eval "$(basher init-bundle -)"

require basherpm/bats-assertions

export TODO_TEST_DIR="${BATS_TMPDIR}/todo"
export TODO_CWD="${TODO_TEST_DIR}/cwd"

export PATH="${BATS_TEST_DIRNAME}/../libexec:$PATH"

mkdir -p "${TODO_CWD}"

setup() {
  cd "${TODO_CWD}"
}

teardown() {
  rm -rf "$TODO_TEST_DIR"
}

load lib/helpers
