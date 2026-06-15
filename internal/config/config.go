package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	StorageType      string `mapstructure:"STORAGE_TYPE"`
	ServerAddress    string
	DatabaseDSN      string
	JWTSecret        string
	MinioEndpoint    string
	MinioPublicURL   string
	MinioAccessKey   string
	MinioSecretKey   string
	MinioUseSSL      bool
	MinioBucket      string
	YandexEndpoint   string `mapstructure:"YANDEX_ENDPOINT"`
	YandexDiskToken  string `mapstructure:"YANDEX_DISK_TOKEN"`
	YandexDiskBucket string `mapstructure:"YANDEX_DISK_BUCKET"`
}

func Load() (*Config, error) {
	v := viper.New()

	v.AutomaticEnv()

	// регистрируем ключи.
	// Без этого Unmarshal не увидит переменные окружения.
	// Ключи в BindEnv должны совпадать с названиями в структуре (mapstructure)
	keys := []string{
		"STORAGE_TYPE",
		"SERVERADDRESS",
		"DATABASEDSN",
		"JWTSECRET",
		"MINIOENDPOINT",
		"MINIOPUBLICURL",
		"MINIOACCESSKEY",
		"MINIOSECRETKEY",
		"MINIOBUCKET",
		"MINIOUSESSL",
		"YANDEX_ENDPOINT",
		"YANDEX_DISK_TOKEN",
		"YANDEX_DISK_BUCKET",
		"TZ",
	}

	for _, key := range keys {
		if err := v.BindEnv(key); err != nil {
			return nil, err
		}
	}

	v.SetDefault("serveraddress", ":8080")
	v.SetDefault("miniobucket", "photos")
	v.SetDefault("STORAGE_TYPE", "yandex")
	v.SetDefault("YANDEX_DISK_BUCKET", "photos")

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

	// Валидация переменных в зависимости от типа хранилища
	switch cfg.StorageType {
	case "minio":
		if cfg.MinioEndpoint == "" {
			return nil, fmt.Errorf("MINIOENDPOINT environment variable is required for minio storage")
		}
		if cfg.MinioAccessKey == "" {
			return nil, fmt.Errorf("MINIOACCESSKEY environment variable is required for minio storage")
		}
		if cfg.MinioSecretKey == "" {
			return nil, fmt.Errorf("MINIOSECRETKEY environment variable is required for minio storage")
		}
	case "yandex":
		if cfg.YandexEndpoint == "" {
			return nil, fmt.Errorf("YANDEX_ENDPOINT environment variable is required for yandex storage")
		}
		if cfg.YandexDiskToken == "" {
			return nil, fmt.Errorf("YANDEX_DISK_TOKEN environment variable is required for yandex storage")
		}
	default:
		return nil, fmt.Errorf("unsupported STORAGE_TYPE: %s (expected 'minio' or 'yandex')", cfg.StorageType)
	}

	return &cfg, nil
}
