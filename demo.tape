# VHS tape for generating assets/demo.gif
#
# Prerequisites:
#   brew install vhs
#   make build
#
# Usage:
#   vhs demo.tape

Output assets/demo.gif

Set Shell zsh
Set FontSize 16
Set Width 900
Set Height 500
Set Padding 20
Set Theme "Catppuccin Mocha"

# Set up demo environment (hidden from recording)
Hide
Type "export TODO_FILE=/tmp/todo-demo.md"
Enter
Sleep 200ms
Type "cd /tmp/todo-demo"
Enter
Sleep 200ms
Type "clear"
Enter
Sleep 300ms
Show

# Launch TUI
Type "todo tui"
Sleep 500ms
Enter
Sleep 1.5s

# Navigate down through items
Type "j"
Sleep 500ms
Type "j"
Sleep 500ms
Type "j"
Sleep 500ms

# Navigate back up
Type "k"
Sleep 500ms
Type "k"
Sleep 500ms

# Check an item with space
Type " "
Sleep 1s

# Move down and toggle active
Type "j"
Sleep 500ms
Type "x"
Sleep 1s

# Pause to show result
Sleep 1s

# Quit
Type "q"
Sleep 500ms
