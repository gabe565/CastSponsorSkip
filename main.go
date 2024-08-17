package main

import (
	"os"

	"github.com/gabe565/castsponsorskip/cmd"
)

var version = "beta"

func main() {
	rootCmd := cmd.New(cmd.WithVersion(version))
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
