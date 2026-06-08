package config

import (
	"crypto/rsa"
	"fmt"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type Config struct {
	Log  Log  `envPrefix:""`
	HTTP HTTP `envPrefix:""`
	GRPC GRPC `envPrefix:""`
	JWT  JWT  `envPrefix:""`
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

type JWT struct {
	PublicKeyPath string `env:"JWT_PUBLIC_KEY_PATH" envDefault:"/run/secrets/jwt_public.pem"`
	PublicKey     *rsa.PublicKey
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

func loadJWTPublicKey(path string) (*rsa.PublicKey, error) {
	keyPEM, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read JWT public key %q: %w", path, err)
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(keyPEM)
	if err != nil {
		return nil, fmt.Errorf("parse JWT public key %q: %w", path, err)
	}

	return publicKey, nil
}

func InitConfig() (Config, error) {
	jwtPublicKeyPath := envOrDefault("JWT_PUBLIC_KEY_PATH", "/run/secrets/jwt_public.pem")
	jwtPublicKey, err := loadJWTPublicKey(jwtPublicKeyPath)

	if err != nil {
		return Config{}, fmt.Errorf("Warning: failed to load JWT public key from %q: %v\n", jwtPublicKeyPath, err)
	}

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
		JWT: JWT{
			PublicKeyPath: jwtPublicKeyPath,
			PublicKey:     jwtPublicKey,
		},
	}
	return cfg, nil
}
