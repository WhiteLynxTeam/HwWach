package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	ServerAddress  string
	DatabaseDSN    string
	JWTSecret      string
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioUseSSL    bool
	MinioBucket    string
}

func Load() (*Config, error) {
	v := viper.New()

	v.AutomaticEnv()

	// регистрируем ключи. 
	// Без этого Unmarshal не увидит переменные окружения.
	// Ключи в BindEnv должны совпадать с названиями в структуре (mapstructure)
	keys := []string{
		"SERVERADDRESS",
		"DATABASEDSN",
		"JWTSECRET",
		"MINIOENDPOINT",
		"MINIOACCESSKEY",
		"MINIOSECRETKEY",
		"MINIOBUCKET",
		"MINIOUSESSL",
		"TZ",
	}

	for _, key := range keys {
		if err := v.BindEnv(key); err != nil {
			return nil, err
		}
	}

	v.SetDefault("serveraddress", ":8080")
	v.SetDefault("miniobucket", "photos")

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	// Валидация обязательных переменных окружения
	if cfg.DatabaseDSN == "" {
		return nil, fmt.Errorf("DATABASEDSN environment variable is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWTSECRET environment variable is required")
	}
	if cfg.MinioEndpoint == "" {
		return nil, fmt.Errorf("MINIOENDPOINT environment variable is required")
	}
	if cfg.MinioAccessKey == "" {
		return nil, fmt.Errorf("MINIOACCESSKEY environment variable is required")
	}
	if cfg.MinioSecretKey == "" {
		return nil, fmt.Errorf("MINIOSECRETKEY environment variable is required")
	}

	return &cfg, nil
}
