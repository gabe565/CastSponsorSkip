package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/gabe565/castsponsorskip/cmd"
)

var Shells = []string{"bash", "zsh", "fish"}

func main() {
	if err := os.RemoveAll("manpages"); err != nil {
		panic(err)
	}

	if err := os.MkdirAll("completions", 0o777); err != nil {
		panic(err)
	}

	rootCmd := cmd.NewCommand()
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	for _, shell := range Shells {
		rootCmd.SetArgs([]string{"--completion=" + shell})
		if err := rootCmd.Execute(); err != nil {
			panic(err)
		}

		f, err := os.Create(filepath.Join("completions", "castsponsorskip."+shell))
		if err != nil {
			panic(err)
		}

		if _, err := io.Copy(f, &buf); err != nil {
			panic(err)
		}

		if err := f.Close(); err != nil {
			panic(err)
		}
	}
}
