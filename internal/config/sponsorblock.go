package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func (c *Config) RegisterSkipSponsors(cmd *cobra.Command) {
	key := "skip-sponsors"
	cmd.PersistentFlags().Bool(key, Default.SkipSponsors, "Skip sponsored segments with SponsorBlock")
	if err := c.viper.BindPFlag(key, cmd.PersistentFlags().Lookup(key)); err != nil {
		panic(err)
	}
}

func (c *Config) RegisterCategories(cmd *cobra.Command) {
	key := "categories"
	cmd.PersistentFlags().StringSliceP(key, "c", Default.Categories, "Comma-separated list of SponsorBlock categories to skip")
	if err := c.viper.BindPFlag(key, cmd.PersistentFlags().Lookup(key)); err != nil {
		panic(err)
	}
	if err := cmd.RegisterFlagCompletionFunc(key, completeCategories); err != nil {
		panic(err)
	}

	if env := os.Getenv("SBCCATEGORIES"); env != "" {
		val := strings.Split(env, " ")
		slog.Warn(fmt.Sprintf(`SBCCATEGORIES is deprecated. Please set %q instead.`, "CSS_CATEGORIES="+strings.Join(val, ",")))
		c.viper.SetDefault(key, val)
	}
}

func (c *Config) RegisterActionTypes(cmd *cobra.Command) {
	key := "action-types"
	cmd.PersistentFlags().StringSlice(key, Default.ActionTypes, "SponsorBlock action types to handle. Shorter segments that overlap with content can be muted instead of skipped.")
	if err := c.viper.BindPFlag(key, cmd.PersistentFlags().Lookup(key)); err != nil {
		panic(err)
	}
	if err := cmd.RegisterFlagCompletionFunc(key, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"skip", "mute"}, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		panic(err)
	}
}

func completeCategories(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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
}
