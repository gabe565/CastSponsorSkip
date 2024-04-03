package main

import (
	"os"

	"github.com/gabe565/castsponsorskip/cmd"
)

//nolint:gochecknoglobals
var (
	version = "beta"
	commit  = ""
)

func main() {
	rootCmd := cmd.NewCommand(version, commit)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
