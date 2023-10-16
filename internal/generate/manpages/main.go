package main

import (
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gabe565/castsponsorskip/cmd"
	"github.com/spf13/cobra/doc"
	flag "github.com/spf13/pflag"
)

func main() {
	flags := flag.NewFlagSet("", flag.ContinueOnError)

	var version string
	flags.StringVar(&version, "version", "beta", "Version")

	var dateParam string
	flags.StringVar(&dateParam, "date", time.Now().Format(time.RFC3339), "Build date")

	if err := flags.Parse(os.Args); err != nil {
		panic(err)
	}

	date, err := time.Parse(time.RFC3339, dateParam)
	if err != nil {
		panic(err)
	}

	if err := os.RemoveAll("manpages"); err != nil {
		panic(err)
	}

	if err := os.MkdirAll("manpages", 0o755); err != nil {
		panic(err)
	}

	rootCmd := cmd.NewCommand("beta", "")
	name := rootCmd.Name()

	header := doc.GenManHeader{
		Title:   strings.ToUpper(name),
		Section: "1",
		Date:    &date,
		Source:  name + " " + version,
		Manual:  "User Commands",
	}

	f, err := os.Create(filepath.Join("manpages", name+".1.gz"))
	if err != nil {
		panic(err)
	}
	gz := gzip.NewWriter(f)

	if err := doc.GenMan(rootCmd, &header, gz); err != nil {
		panic(err)
	}

	if err := gz.Close(); err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}

	if err := os.Chtimes(f.Name(), date, date); err != nil {
		panic(err)
	}
}
