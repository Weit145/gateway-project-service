package config

import (
	"os"
	"strings"
)

type Config struct {
	Log  Log  `envPrefix:""`
	HTTP HTTP `envPrefix:""`
	GRPC GRPC `envPrefix:""`
}

type Log struct {
	Level string `env:"LOG_LEVEL" envDefault:"DEBUG"`
}

type HTTP struct {
	Address string `env:"HTTP_ADDRESS" envDefault:"0.0.0.0:8080"`
}

type GRPC struct {
	AuthAddr string `env:"AUTH_GRPC_ADDR" envDefault:"auth:50051"`
}

func envOrDefault(env string, def string) string {
	value, ok := os.LookupEnv(env)
	if !ok {
		return strings.TrimSpace(def)
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return strings.TrimSpace(def)
	}
	return value
}

func InitConfig() Config {
	cfg := Config{
		Log: Log{
			Level: envOrDefault("LOG_LEVEL", "DEBUG"),
		},
		HTTP: HTTP{
			Address: envOrDefault("HTTP_ADDRESS", "0.0.0.0:8080"),
		},
		GRPC: GRPC{
			AuthAddr: envOrDefault("AUTH_GRPC_ADDR", "auth:50051"),
		},
	}
	return cfg
}
