source "$(basher package-path ztombol/bats-support)/load.bash"
source "$(basher package-path ztombol/bats-assert)/load.bash"

export TODO_TEST_DIR="${BATS_TMPDIR}/todo"
export TODO_CWD="${TODO_TEST_DIR}/cwd"

export TODO_FILE=""

export PATH="${BATS_TEST_DIRNAME}/../libexec:$PATH"

mkdir -p "${TODO_CWD}"

setup() {
  cd "${TODO_CWD}"
}

teardown() {
  rm -rf "$TODO_TEST_DIR"
}

load lib/helpers
