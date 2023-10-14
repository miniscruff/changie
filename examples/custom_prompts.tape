# Configure all our settings
Set FontSize 20
Set Width 1200
Set Height 600
Set Shell zsh
Set TypingSpeed 50ms
Set Theme "Catppuccin Mocha"
Set Padding 8

Output custom_prompts.webm
Output custom_prompts.gif

Hide
Type "changie init"
Enter
Type "export CHANGIE_CONFIG_PATH=custom_prompts_config.yaml"
Enter
Type "clear"
Enter
Show

Type "changie new"
Sleep 1s
Enter
Sleep 1s

# Use the added kind
Enter
Sleep 1s

Type "custom links for issue and authors"
Sleep 1s
Enter
Sleep 1s

# Issue number
Type "26"
Sleep 1s
Enter
Sleep 1s

# Author
Type "miniscruff"
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
