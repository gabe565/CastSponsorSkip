package config

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func (c *Config) RegisterCategories(cmd *cobra.Command) {
	key := "categories"
	cmd.PersistentFlags().StringSliceP(key, "c", []string{"sponsor"}, "Comma-separated list of SponsorBlock categories to skip")
	if err := viper.BindPFlag(key, cmd.PersistentFlags().Lookup(key)); err != nil {
		panic(err)
	}
	if err := cmd.RegisterFlagCompletionFunc(key, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		response, err := http.Get("https://github.com/ajayyy/SponsorBlock/raw/master/config.json.example")
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		config := make(map[string]any)
		if err := json.NewDecoder(response.Body).Decode(&config); err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		wikiLinks, ok := config["wikiLinks"].(map[string]any)
		if !ok {
			return nil, cobra.ShellCompDirectiveError
		}

		completions := make([]string, 0, len(wikiLinks))
		for category, url := range wikiLinks {
			completions = append(completions, category+"\t"+url.(string))
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		panic(err)
	}

	if env := os.Getenv("SBCCATEGORIES"); env != "" {
		viper.SetDefault(key, env)
	}
}
