package main

import (
	"log/slog"
	"os"

	"github.com/gabe565/castsponsorskip/cmd"
)

var version = "beta"

func main() {
	rootCmd := cmd.New(cmd.WithVersion(version))
	if err := rootCmd.Execute(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
