package main

import (
	"os"

	"github.com/gabe565/sponsorblockcast-go/cmd"
)

func main() {
	rootCmd := cmd.NewCommand()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
