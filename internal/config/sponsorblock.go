package config

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	CategoriesKey   = "categories"
	CategoriesValue = []string{"sponsor"}
)

func Categories(cmd *cobra.Command) {
	cmd.PersistentFlags().StringSlice(CategoriesKey, CategoriesValue, "Sponsor Block categories to skip")
	if err := viper.BindPFlag(CategoriesKey, cmd.PersistentFlags().Lookup(CategoriesKey)); err != nil {
		panic(err)
	}

	if env := os.Getenv("SBCCATEGORIES"); env != "" {
		viper.SetDefault(CategoriesKey, env)
	}
}
