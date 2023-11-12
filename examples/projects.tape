# Configure all our settings
Set FontSize 20
Set Width 800
Set Height 600
Set TypingSpeed 50ms
Set Theme "Catppuccin Mocha"
Set Padding 8

Output projects.webm
Output projects.gif

Hide
Type "changie init"
Enter
Type "export CHANGIE_CONFIG_PATH=projects_config.yaml"
Enter
Type "mkdir -p .changes/landing .changes/user_feeds .changes/user_management"
Enter
Type "mkdir -p users/feeds users/management landing"
Enter
Type "clear"
Enter
Show

Type "changie new"
Sleep 1s
Enter
Sleep 1s

# Select user feeds and management
Space
Sleep 1s
Down
Sleep 1s
Space
Sleep 1s
Enter
Sleep 1s

# Use the added kind
Enter
Sleep 1s

Type "New feature in two projects"
Sleep 1s
Enter
Sleep 1s

Type "ls .changes/unreleased"
Sleep 1s
Enter
Sleep 1s

Type "changie batch -j user_feeds auto"
Sleep 1s
Enter
Sleep 1s

Type "cat .changes/user_feeds/v0"
Sleep 1s
Tab
Sleep 1s
Enter

Type "changie merge"
Sleep 1s
Enter
Sleep 1s

Type "cat users/feeds/CHANGELOG.md"
Sleep 1s
Enter

Type "changie latest -j user_feeds"
Sleep 1s
Enter
Enter

# Slight pause at the end
Sleep 3s

# Cleanup
Hide
Type "rm CHANGELOG.md"
Enter
Type "rm .changie.yaml"
Enter
Type "rm -rf .changes"
Enter
Type "rm -rf users"
Enter
Type "rm -rf landing"
Enter
