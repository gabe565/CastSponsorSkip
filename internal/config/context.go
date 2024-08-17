package config

import "context"

type ctxKey uint8

const configKey ctxKey = iota

func NewContext(ctx context.Context, conf *Config) context.Context {
	return context.WithValue(ctx, configKey, conf)
}

func FromContext(ctx context.Context) *Config {
	conf, ok := ctx.Value(configKey).(*Config)
	if !ok {
		panic("config not found in context")
	}
	return conf
}
