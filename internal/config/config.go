package config

import "github.com/spf13/viper"

func Load() {
	PausedIntervalValue = viper.GetDuration(PausedIntervalKey)
	PlayingIntervalValue = viper.GetDuration(PlayingIntervalKey)
	CategoriesValue = viper.GetStringSlice(CategoriesKey)
}
