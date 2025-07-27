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

	v.SetDefault("ServerAddress", ":8080")
	v.SetDefault("DatabaseDSN", "host=localhost user=app dbname=app sslmode=disable password=secret")
	v.SetDefault("JWTSecret", "your_jwt_secret_here")
	v.SetDefault("MinioEndpoint", "play.min.io:9000")
	v.SetDefault("MinioAccessKey", "Q3AM3UQ867SPQQA43P2F")
	v.SetDefault("MinioSecretKey", "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG")
	v.SetDefault("MinioUseSSL", true)
	v.SetDefault("MinioBucket", "photos")

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
