on tab_no_ctx()
  tell application "System Events" to keystroke "t" using command down
  delay 1
end tab_no_ctx

on full_screen()
  tell application "System Events" to key code 36 using command down
end full_screen

tell application "iTerm"
    create window with default profile
      set project_dir to POSIX path of ((path to me as text) & "::")
      set db_dir to "/tmp/badger"
    tell first session of current tab of current window
      set name to "Ledger"
      write text "cd " & project_dir
      write text "rm -rf " & db_dir & " &&  ./bin/db -config ledger.yml"

      split horizontally with default profile
    end tell
    tell second session of current tab of current window
      set name to "Ledger"
      write text "cd " & project_dir
      write text "./bin/pulsar -config node.yml"

      split horizontally with default profile
    end tell
    tell third session of current tab of current window
      set name to "Ledger"
      write text "cd " & project_dir
      write text "./bin/pulsar -config node2.yml"

      split horizontally with default profile
    end tell
    tell fourth session of current tab of current window
      set name to "Ledger"
      write text "cd " & project_dir
      write text "./bin/pulsar -config node3.yml"

      split horizontally with default profile
    end tell
    tell fifth session of current tab of current window
      set name to "Ledger"
      write text "cd " & project_dir
      write text "./bin/pulsar -config node4.yml"

      split horizontally with default profile
    end tell
end tell

my full_screen()