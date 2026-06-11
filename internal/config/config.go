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
	AuthAddr        string `env:"AUTH_GRPC_ADDR" envDefault:"auth:50051"`
	TLSCertPath     string `env:"AUTH_GRPC_TLS_CERT_PATH"`
	TLSKeyPath      string `env:"AUTH_GRPC_TLS_KEY_PATH"`
	TLSServerCAPath string `env:"AUTH_GRPC_TLS_CA_PATH"`
	TLSServerName   string `env:"AUTH_GRPC_TLS_SERVER_NAME" envDefault:"auth"`
	TLSCertPEM      []byte
	TLSKeyPEM       []byte
	TLSServerCAPEM  []byte
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

func requiredEnv(name string) (string, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return "", fmt.Errorf("%s is required", name)
	}

	return value, nil
}

func readRequiredFile(path, name string) ([]byte, error) {
	value, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s %q: %w", name, path, err)
	}
	if len(value) == 0 {
		return nil, fmt.Errorf("%s %q is empty", name, path)
	}

	return value, nil
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
	grpcTLSCertPath, err := requiredEnv("AUTH_GRPC_TLS_CERT_PATH")
	if err != nil {
		return Config{}, err
	}

	grpcTLSKeyPath, err := requiredEnv("AUTH_GRPC_TLS_KEY_PATH")
	if err != nil {
		return Config{}, err
	}

	grpcTLSServerCAPath, err := requiredEnv("AUTH_GRPC_TLS_CA_PATH")
	if err != nil {
		return Config{}, err
	}

	grpcTLSCertPEM, err := readRequiredFile(grpcTLSCertPath, "auth gRPC TLS client certificate")
	if err != nil {
		return Config{}, err
	}

	grpcTLSKeyPEM, err := readRequiredFile(grpcTLSKeyPath, "auth gRPC TLS client private key")
	if err != nil {
		return Config{}, err
	}

	grpcTLSServerCAPEM, err := readRequiredFile(grpcTLSServerCAPath, "auth gRPC TLS server CA certificate")
	if err != nil {
		return Config{}, err
	}

	jwtPublicKeyPath := envOrDefault("JWT_PUBLIC_KEY_PATH", "/run/secrets/jwt_public.pem")
	jwtPublicKey, err := loadJWTPublicKey(jwtPublicKeyPath)

	if err != nil {
		return Config{}, fmt.Errorf("load JWT public key from %q: %w", jwtPublicKeyPath, err)
	}

	cfg := Config{
		Log: Log{
			Level: envOrDefault("LOG_LEVEL", "DEBUG"),
		},
		HTTP: HTTP{
			Address: envOrDefault("HTTP_ADDRESS", "0.0.0.0:8080"),
		},
		GRPC: GRPC{
			AuthAddr:        envOrDefault("AUTH_GRPC_ADDR", "auth:50051"),
			TLSCertPath:     grpcTLSCertPath,
			TLSKeyPath:      grpcTLSKeyPath,
			TLSServerCAPath: grpcTLSServerCAPath,
			TLSServerName:   envOrDefault("AUTH_GRPC_TLS_SERVER_NAME", "auth"),
			TLSCertPEM:      grpcTLSCertPEM,
			TLSKeyPEM:       grpcTLSKeyPEM,
			TLSServerCAPEM:  grpcTLSServerCAPEM,
		},
		JWT: JWT{
			PublicKeyPath: jwtPublicKeyPath,
			PublicKey:     jwtPublicKey,
		},
	}
	return cfg, nil
}
