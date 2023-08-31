package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var completionFlag string

func CompletionFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&completionFlag, "completion", "", "Output command-line completion code for the specified shell. Can be 'bash', 'zsh', 'fish', or 'powershell'.")
	err := cmd.RegisterFlagCompletionFunc("completion", completionCompletion)
	if err != nil {
		panic(err)
	}
}

func completionCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"bash", "zsh", "fish", "powershell"}, cobra.ShellCompDirectiveNoFileComp
}

func completion(cmd *cobra.Command) error {
	switch completionFlag {
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
		return fmt.Errorf("%v: invalid shell", completionFlag)
	}
	return nil
}
