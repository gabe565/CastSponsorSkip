package main

import (
	"log/slog"
	"os"

	"gabe565.com/castsponsorskip/cmd"
)

var version = "beta"

func main() {
	rootCmd := cmd.New(cmd.WithVersion(version))
	if err := rootCmd.Execute(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
