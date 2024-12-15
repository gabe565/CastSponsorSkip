package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gabe565.com/castsponsorskip/cmd"
	"gabe565.com/castsponsorskip/internal/config"
	"gabe565.com/castsponsorskip/internal/config/names"
	"gabe565.com/utils/cobrax"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/pflag"
)

func main() {
	output := "./docs"

	if err := os.RemoveAll(output); err != nil {
		slog.Error("failed to remove existing dir", "error", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(output, 0o755); err != nil {
		slog.Error("failed to mkdir", "error", err)
		os.Exit(1)
	}

	root := cmd.New(cobrax.WithVersion("beta"))

	if err := errors.Join(
		generateFlagDoc(root, output),
		generateEnvDoc(root, filepath.Join(output, "envs.md")),
	); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func generateFlagDoc(cmd *cobra.Command, output string) error {
	if err := doc.GenMarkdownTree(cmd, output); err != nil {
		return fmt.Errorf("failed to generate markdown: %w", err)
	}
	return nil
}

func generateEnvDoc(cmd *cobra.Command, output string) error {
	excludeNames := []string{"completion", names.FlagConfig, "help", "version"}
	var rows []table.Row
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if slices.Contains(excludeNames, flag.Name) {
			return
		}

		var value string
		switch fv := flag.Value.(type) {
		case pflag.SliceValue:
			value = strings.Join(flag.Value.(pflag.SliceValue).GetSlice(), ",")
		default:
			value = fv.String()
		}
		if value == "" {
			value = " "
		}

		rows = append(rows, table.Row{
			"`" + config.EnvPrefix + strings.ReplaceAll(strings.ToUpper(flag.Name), "-", "_") + "`",
			flag.Usage,
			"`" + value + "`",
		})
	})
	t := table.NewWriter()
	t.AppendHeader(table.Row{"Name", "Usage", "Default"})
	t.AppendRows(rows)

	var buf strings.Builder
	buf.WriteString("# Environment Variables\n\n")
	buf.WriteString(t.RenderMarkdown())

	f, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("failed to create env file: %w", err)
	}
	defer f.Close()

	if _, err := io.WriteString(f, buf.String()); err != nil {
		return fmt.Errorf("failed to write to env file: %w", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close env file: %w", err)
	}

	return nil
}
