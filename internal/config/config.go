package config

import (
	"crypto/rsa"
	"fmt"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Config struct {
	Log   Log   `envPrefix:""`
	HTTP  HTTP  `envPrefix:""`
	GRPC  GRPC  `envPrefix:""`
	JWT   JWT   `envPrefix:""`
	Redis Redis `envPrefix:""`
}

type Log struct {
	Level string `env:"LOG_LEVEL" envDefault:"DEBUG"`
}

type HTTP struct {
	Address        string `env:"HTTP_ADDRESS" envDefault:"0.0.0.0:8080"`
	TrustedProxies []netip.Prefix
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
	TimeoutShort    time.Duration
	TimeoutDefault  time.Duration
	TimeoutLong     time.Duration
}

type JWT struct {
	PublicKeyPath string `env:"JWT_PUBLIC_KEY_PATH" envDefault:"/run/secrets/jwt_public.pem"`
	PublicKey     *rsa.PublicKey
}

type Redis struct {
	Addr     string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	Password string `env:"REDIS_PASSWORD"`
	DB       int    `env:"REDIS_DB" envDefault:"0"`
	Protocol int    `env:"REDIS_PROTOCOL" envDefault:"3"`
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

func envInt(env string, def int) (int, error) {
	value := envOrDefault(env, strconv.Itoa(def))
	result, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", env, err)
	}
	return result, nil
}

func envDuration(env string, def time.Duration) (time.Duration, error) {
	value := envOrDefault(env, def.String())
	result, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration: %w", env, err)
	}
	if result <= 0 {
		return 0, fmt.Errorf("%s must be positive", env)
	}
	return result, nil
}

func trustedProxies(value string) ([]netip.Prefix, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}

	values := strings.Split(value, ",")
	proxies := make([]netip.Prefix, 0, len(values))
	for _, value := range values {
		proxy, err := netip.ParsePrefix(strings.TrimSpace(value))
		if err != nil {
			return nil, fmt.Errorf("parse HTTP_TRUSTED_PROXIES %q: %w", value, err)
		}
		proxies = append(proxies, proxy)
	}
	return proxies, nil
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
	proxies, err := trustedProxies(envOrDefault("HTTP_TRUSTED_PROXIES", ""))
	if err != nil {
		return Config{}, err
	}

	redisDB, err := envInt("REDIS_DB", 0)
	if err != nil {
		return Config{}, err
	}

	redisProtocol, err := envInt("REDIS_PROTOCOL", 3)
	if err != nil {
		return Config{}, err
	}

	grpcTimeoutShort, err := envDuration("AUTH_GRPC_TIMEOUT_SHORT", 2*time.Second)
	if err != nil {
		return Config{}, err
	}

	grpcTimeoutDefault, err := envDuration("AUTH_GRPC_TIMEOUT_DEFAULT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}

	grpcTimeoutLong, err := envDuration("AUTH_GRPC_TIMEOUT_LONG", 8*time.Second)
	if err != nil {
		return Config{}, err
	}

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
			Address:        envOrDefault("HTTP_ADDRESS", "0.0.0.0:8080"),
			TrustedProxies: proxies,
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
			TimeoutShort:    grpcTimeoutShort,
			TimeoutDefault:  grpcTimeoutDefault,
			TimeoutLong:     grpcTimeoutLong,
		},
		JWT: JWT{
			PublicKeyPath: jwtPublicKeyPath,
			PublicKey:     jwtPublicKey,
		},
		Redis: Redis{
			Addr:     envOrDefault("REDIS_ADDR", "localhost:6379"),
			Password: envOrDefault("REDIS_PASSWORD", ""),
			DB:       redisDB,
			Protocol: redisProtocol,
		},
	}
	return cfg, nil
}
