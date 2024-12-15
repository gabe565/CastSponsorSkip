package main

import (
	"log/slog"
	"os"

	"gabe565.com/castsponsorskip/cmd"
	"gabe565.com/utils/cobrax"
)

var version = "beta"

func main() {
	rootCmd := cmd.New(cobrax.WithVersion(version))
	if err := rootCmd.Execute(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
