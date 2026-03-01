Output demo.gif

Set FontSize 18
Set Width 1200
Set Height 600
Set Theme "Catppuccin Mocha"
Set TypingSpeed 120ms
Set Padding 20
Set Shell "bash"

# Bootstrap demo environment (hidden, fast)
Hide
Type "source demo/setup.sh"
Enter
Sleep 1s
# Remove existing config so setup wizard triggers
Type "rm -f $XDG_CONFIG_HOME/pulse/config.json"
Enter
Sleep 0.3s
Show

# 1. First launch — setup wizard
Type "pulse setup"
Enter
Sleep 2s
# Type the projects directory path into the huh form
Set TypingSpeed 40ms
Type "./demo/projects"
Set TypingSpeed 120ms
Sleep 1s
Enter
Sleep 3s

# Re-apply env so pulse uses demo dir going forward
Hide
Type "source demo/setup.sh 2>/dev/null"
Enter
Sleep 0.5s
Show

# 2. Dashboard — focus, streak, contribution graph
Type "pulse"
Enter
Sleep 4s

# 3. Project map — statuses and last commits
Type "pulse map"
Enter
Sleep 4s

# 4. Project details — stack, done criteria
Type "pulse info my-awesome-api"
Enter
Sleep 4s

# 5. Today's log
Type "pulse log"
Enter
Sleep 4s

# 6. Weekly log with commit hashes
Type "pulse log week"
Enter
Sleep 4s

# 7. Park an idea without losing focus
Type 'pulse park "Add GraphQL layer on top of REST"'
Enter
Sleep 3s

# 8. All parked ideas
Type "pulse ideas list"
Enter
Sleep 4s

# 9. Switch history — every context switch is tracked
Type "pulse switches"
Enter
Sleep 4s

# 10. All available commands
Type "pulse help"
Enter
Sleep 5s
