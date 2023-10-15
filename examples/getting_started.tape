# Configure all our settings
Set FontSize 20
Set Width 1000
Set Height 600
Set Shell zsh
Set TypingSpeed 50ms
Set Theme "Catppuccin Mocha"
Set Padding 8

Output examples/getting_started.webm
Output examples/getting_started.gif

Type "changie init"
Sleep 1s
Enter
Sleep 1s

Type "changie new"
Sleep 1s
Enter
Sleep 1s

Down@250ms 4
Enter
Sleep 1s

Type "bugfix details"
Sleep 1s
Enter
Sleep 1s

Type "changie batch auto"
Sleep 1s
Enter
Sleep 1s

Type "cat .changes/v0"
Sleep 1s
Tab
Sleep 1s
Enter

Type "changie merge"
Sleep 1s
Enter
Sleep 1s

Type "cat CHANGELOG.md"
Sleep 1s
Enter
Sleep 1s

# Slight pause at the end
Sleep 2s

# Cleanup
Hide
Type "rm -rf .changes"
Enter
Type "rm CHANGELOG.md"
Enter
Type "rm .changie.yaml"
Enter
