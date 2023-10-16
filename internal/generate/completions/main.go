package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gabe565/castsponsorskip/cmd"
	flag "github.com/spf13/pflag"
)

var Shells = []string{"bash", "zsh", "fish"}

func main() {
	flags := flag.NewFlagSet("", flag.ContinueOnError)

	var dateParam string
	flags.StringVar(&dateParam, "date", time.Now().Format(time.RFC3339), "Build date")

	if err := flags.Parse(os.Args); err != nil {
		panic(err)
	}

	date, err := time.Parse(time.RFC3339, dateParam)
	if err != nil {
		panic(err)
	}

	if err := os.RemoveAll("completions"); err != nil {
		panic(err)
	}

	if err := os.MkdirAll("completions", 0o777); err != nil {
		panic(err)
	}

	rootCmd := cmd.NewCommand("latest", "")
	name := rootCmd.Name()
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	for _, shell := range Shells {
		rootCmd.SetArgs([]string{"--completion=" + shell})
		if err := rootCmd.Execute(); err != nil {
			panic(err)
		}

		f, err := os.Create(filepath.Join("completions", name+"."+shell))
		if err != nil {
			panic(err)
		}

		if _, err := io.Copy(f, &buf); err != nil {
			panic(err)
		}

		if err := f.Close(); err != nil {
			panic(err)
		}

		if err := os.Chtimes(f.Name(), date, date); err != nil {
			panic(err)
		}
	}
}
