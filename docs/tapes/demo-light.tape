# gh-infra demo (light theme variant)
#
# Run with: make demos
# Requirements: Docker

Output demo-light.gif
Output demo-light.mp4

# Light theme (Tokyo Night Day inspired)
Set Theme { "name": "gh-infra-light", "black": "#0f0f14", "red": "#f52a65", "green": "#587539", "yellow": "#8c6c3e", "blue": "#2e7de9", "magenta": "#9854f1", "cyan": "#007197", "white": "#6172b0", "brightBlack": "#a1a6c5", "brightRed": "#f52a65", "brightGreen": "#587539", "brightYellow": "#8c6c3e", "brightBlue": "#2e7de9", "brightMagenta": "#9854f1", "brightCyan": "#007197", "brightWhite": "#3760bf", "background": "#e1e2e7", "foreground": "#3760bf", "selection": "#b6bfe2", "cursor": "#3760bf" }

Set Shell "bash"
Set FontSize 22
Set Width 1200
Set Height 800
Set Padding 20
Set WindowBar ""
Set TypingSpeed 50ms

# --- Hidden setup ---
Hide
Type "source /data/setup.sh"
Enter
Sleep 2s
Type "clear"
Enter
Sleep 500ms
Show

# --- Demo starts ---

# 1. Show the YAML file
Sleep 1s
Type "cat /tmp/demo/repos.yaml"
Sleep 500ms
Enter
Sleep 4s

# 2. Run plan
Type "gh infra plan /tmp/demo/"
Sleep 500ms
Enter
Sleep 6s

# 3. Run apply
Type "gh infra apply /tmp/demo/ --auto-approve"
Sleep 500ms
Enter
Sleep 8s

# End
Sleep 3s
