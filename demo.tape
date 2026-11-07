# VHS tape for assets/demo.gif and assets/help.png
#
# Prerequisites: brew install vhs
# Usage: make demo

Output assets/demo.gif

Set Shell zsh
Set FontFamily "SauceCodePro Nerd Font"
Set FontSize 14
Set Width 1400
Set Height 500
Set Padding 24
Set TypingSpeed 60ms
Set Theme { "name": "TokyoNight Moon", "black": "#1b1d2b", "red": "#ff757f", "green": "#c3e88d", "yellow": "#ffc777", "blue": "#82aaff", "purple": "#c099ff", "cyan": "#86e1fc", "white": "#828bb8", "brightBlack": "#444a73", "brightRed": "#ff757f", "brightGreen": "#c3e88d", "brightYellow": "#ffc777", "brightBlue": "#82aaff", "brightPurple": "#c099ff", "brightCyan": "#86e1fc", "brightWhite": "#c8d3f5", "background": "#222436", "foreground": "#c8d3f5", "selection": "#2d3f76", "cursor": "#c8d3f5" }

Hide
Type "export PROMPT='$ ' TODO_FILE=/tmp/todo-demo.md"
Enter
Type "rm -rf /tmp/todo-demo && mkdir -p /tmp/todo-demo && cd /tmp/todo-demo"
Enter
Type "printf '## /tmp/todo-demo\n\n- [ ] Ship the Rust rewrite @active\n- [ ] Fix the login bug\n  - [ ] Reproduce locally\n  - [x] Check server logs\n- [ ] Write release notes\n--- Later\n- [ ] Update the homepage\n- [ ] Plan the next sprint\n' > $TODO_FILE"
Enter
Type "clear"
Enter
Sleep 300ms
Show

Type "todo tui"
Sleep 400ms
Enter
Sleep 1.5s

# Move to "Reproduce locally" and mark it done
Type "jj"
Sleep 600ms
Type " "
Sleep 1s

# Add an item below "Write release notes"
Type "jj"
Sleep 600ms
Type "o"
Sleep 400ms
Type "Draft the announcement"
Sleep 400ms
Enter
Sleep 1s

# Mark it active: it moves to the top
Type "x"
Sleep 1.2s

# Filter and jump
Type "/"
Sleep 400ms
Type "login"
Sleep 1.2s
Enter
Sleep 1s

# Show the keymap
Type "?"
Sleep 1.5s
Screenshot assets/help.png
Sleep 1s
Type "?"
Sleep 800ms

Type "q"
Sleep 500ms
