# gh-infra demo: plan → apply workflow
#
# Run with: make demos
# Requirements: Docker
#
# This tape uses a mock gh CLI (mock-gh) to simulate GitHub API responses.
# gh-infra runs for real — spinners, diffs, and colors are all genuine.

Output demo.gif
Output demo.mp4

# Dark theme (Tokyo Night inspired)
Set Theme { "name": "gh-infra-dark", "black": "#16161e", "red": "#f7768e", "green": "#9ece6a", "yellow": "#e0af68", "blue": "#7aa2f7", "magenta": "#bb9af7", "cyan": "#7dcfff", "white": "#a9b1d6", "brightBlack": "#565f89", "brightRed": "#f7768e", "brightGreen": "#9ece6a", "brightYellow": "#e0af68", "brightBlue": "#7aa2f7", "brightMagenta": "#bb9af7", "brightCyan": "#7dcfff", "brightWhite": "#c0caf5", "background": "#1a1b26", "foreground": "#a9b1d6", "selection": "#283457", "cursor": "#c0caf5" }

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
