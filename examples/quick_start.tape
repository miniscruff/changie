# Configure all our settings
Set FontSize 20
Set Width 800
Set Height 600
Set TypingSpeed 50ms
Set Theme "Catppuccin Mocha"
Set Padding 8

Output quick_start.webm
Output quick_start.gif

Type "changie init"
Sleep 1s
Enter
Sleep 1s

Type "changie new"
Sleep 1s
Enter
Sleep 1s

Down@250ms 4
Sleep 1s
Enter

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
Sleep 1s

Type "changie merge"
Sleep 1s
Enter
Sleep 1s

Type "cat CHANGELOG.md"
Sleep 1s
Enter
Sleep 1s

Hide
Type "rm -rf .changes"
Enter
Type "rm CHANGELOG.md"
Enter
Type "rm .changie.yaml"
Enter
