package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func CompletionFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().String("completion", "", "Output command-line completion code for the specified shell. Can be 'bash', 'zsh', 'fish', or 'powershell'.")
	err := cmd.RegisterFlagCompletionFunc("completion", completionCompletion)
	if err != nil {
		panic(err)
	}
}

func completionCompletion(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string{"bash", "zsh", "fish", "powershell"}, cobra.ShellCompDirectiveNoFileComp
}

var ErrInvalidShell = errors.New("invalid shell")

func completion(cmd *cobra.Command, shell string) error {
	switch shell {
	case "bash":
		if err := cmd.Root().GenBashCompletion(cmd.OutOrStdout()); err != nil {
			return err
		}
	case "zsh":
		if err := cmd.Root().GenZshCompletion(cmd.OutOrStdout()); err != nil {
			return err
		}
	case "fish":
		if err := cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true); err != nil {
			return err
		}
	case "powershell":
		if err := cmd.Root().GenPowerShellCompletionWithDesc(cmd.OutOrStdout()); err != nil {
			return err
		}
	default:
		return fmt.Errorf("%w: %s", ErrInvalidShell, shell)
	}
	return nil
}
