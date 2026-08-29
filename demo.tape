Output demo.gif

Set Shell zsh
Set FontSize 14
Set Width 900
Set Height 500
Set Padding 20

# Interactive TUI
Type "chbedcl -r us-east-1"
Enter
Sleep 1s

Down
Down
Sleep 500ms
Down
Sleep 500ms

Type "opus"
Sleep 1s

Enter
Sleep 1s

# Headless / agent-friendly
Type "# For agents:"
Enter
Sleep 500ms

Type "chbedcl --list --json -r us-east-1"
Enter
Sleep 2s

Type "chbedcl --set us.anthropic.claude-opus-5 --json"
Enter
Sleep 1s
