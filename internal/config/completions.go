package config

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"gabe565.com/castsponsorskip/internal/config/names"
	"gabe565.com/utils/must"
	"github.com/spf13/cobra"
)

func RegisterCompletions(cmd *cobra.Command) {
	must.Must(cmd.RegisterFlagCompletionFunc(names.FlagLogLevel, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"debug", "info", "warn", "error", "none"}, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
	}))
	must.Must(cmd.RegisterFlagCompletionFunc(names.FlagNetworkInterface, completeNetworkInterface))
	must.Must(cmd.RegisterFlagCompletionFunc(names.FlagDiscoverInterval, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"5m", "10m", "15m"}, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
	}))
	must.Must(cmd.RegisterFlagCompletionFunc(names.FlagPausedInterval, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"1m", "2m", "5m", "10m", "30m", "1h"}, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
	}))
	must.Must(cmd.RegisterFlagCompletionFunc(names.FlagPlayingInterval, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"1s", "2s"}, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
	}))
	must.Must(cmd.RegisterFlagCompletionFunc(names.FlagSkipDelay, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"500ms", "1s", "2s", "3s", "5s", "10s"}, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
	}))
	must.Must(cmd.RegisterFlagCompletionFunc(names.FlagIgnoreSegmentDuration, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"30s", "1m", "2m", "5m"}, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
	}))
	must.Must(cmd.RegisterFlagCompletionFunc(names.FlagCategories, completeCategories))
	must.Must(cmd.RegisterFlagCompletionFunc(names.FlagActionTypes, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"skip", "mute"}, cobra.ShellCompDirectiveNoFileComp
	}))
}

func completeNetworkInterface(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	completions := make([]string, 0, len(interfaces))
	for _, iface := range interfaces {
		addrs, _ := iface.Addrs()
		addrStr := make([]string, 0, len(addrs))
		for _, addr := range addrs {
			addrStr = append(addrStr, addr.String())
		}
		completions = append(completions, iface.Name+"\t"+strings.Join(addrStr, ", "))
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

func completeCategories(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const srcURL = "https://github.com/ajayyy/SponsorBlock/raw/master/config.json.example"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srcURL, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	config := make(map[string]any)
	if err := json.NewDecoder(response.Body).Decode(&config); err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	_ = response.Body.Close()

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
