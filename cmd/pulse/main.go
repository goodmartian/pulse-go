package main

import (
	"fmt"
	"os"

	"github.com/goodmartian/pulse-go/internal/config"
	"github.com/goodmartian/pulse-go/internal/i18n"
	"github.com/goodmartian/pulse-go/internal/tui"
)

func main() {
	config.Init()
	i18n.Init()
	args := os.Args[1:]

	if config.NeedsSetup() {
		if len(args) == 0 || args[0] != "setup" {
			fi, _ := os.Stdin.Stat()
			if fi == nil || fi.Mode()&os.ModeCharDevice == 0 {
				fmt.Fprintln(os.Stderr, "pulse: run 'pulse setup' interactively first, or set PULSE_DIR env var.")
				os.Exit(1)
			}
			tui.CmdSetup()
			i18n.Init()
		}
	}

	tui.Run(args)
}
