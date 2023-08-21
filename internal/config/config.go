package config

import (
	"regexp"

	"github.com/spf13/viper"
)

func Load() {
	InitViper()
	DiscoverIntervalValue = viper.GetDuration(DiscoverIntervalKey)
	PausedIntervalValue = viper.GetDuration(PausedIntervalKey)
	PlayingIntervalValue = viper.GetDuration(PlayingIntervalKey)
	InterfaceValue = viper.GetString(InterfaceKey)
	CategoriesValue = regexp.MustCompile("[, ]+").Split(viper.GetString(CategoriesKey), -1)
	YouTubeAPIKeyValue = viper.GetString(YouTubeAPIKeyKey)
}
