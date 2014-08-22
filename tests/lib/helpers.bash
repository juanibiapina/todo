create_todo_file() {
  local todofile="$1"
  local content="$2"

  echo "$content" > "$todofile"
}

create_command() {
  local command="$1"
  shift
  mkdir -p "${TODO_TEST_DIR}/path"
  cat > "${TODO_TEST_DIR}/path/$command" <<SH
#!/usr/bin/env bash

$@
SH
  chmod +x "${TODO_TEST_DIR}/path/$command"
  export PATH="${TODO_TEST_DIR}/path:$PATH"
}
