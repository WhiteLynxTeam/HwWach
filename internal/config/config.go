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

	// Сервер
	v.SetDefault("serveraddress", ":8080")

	// PostgreSQL — дефолт для docker-compose (контейнер postgres)
	v.SetDefault("databasedsn", "host=postgres user=postgres dbname=hwwach_db sslmode=disable password=postgres")

	// JWT — безопасный дефолт (но лучше менять в production)
	v.SetDefault("jwtsecret", "your-super-secret-jwt-key-here-make-it-long-and-random")

	// MinIO — дефолт для docker-compose (контейнер minio)
	v.SetDefault("minioendpoint", "minio:9000")
	v.SetDefault("minioaccesskey", "Q3AM3UQ867SPQQA43P2F")
	v.SetDefault("miniosecretkey", "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG")
	v.SetDefault("miniousessl", false)
	v.SetDefault("miniobucket", "photos")

	v.SetConfigName("config")
	v.SetConfigType("toml")
	v.AddConfigPath(".")
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	return &cfg, nil
}
