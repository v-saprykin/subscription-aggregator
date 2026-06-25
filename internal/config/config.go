package config

import (
	"errors"

	"github.com/spf13/viper"
)

type Config struct {
	AppEnv      string
	HTTPAddr    string
	LogLevel    string
	DatabaseURL string
}

func Load() (Config, error) {
	v := viper.New()
	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AddConfigPath(".")

	v.SetDefault("APP_ENV", "local")
	v.SetDefault("HTTP_ADDR", ":8080")
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("DATABASE_URL", "")

	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return Config{}, err
		}
	}

	return Config{
		AppEnv:      v.GetString("APP_ENV"),
		HTTPAddr:    v.GetString("HTTP_ADDR"),
		LogLevel:    v.GetString("LOG_LEVEL"),
		DatabaseURL: v.GetString("DATABASE_URL"),
	}, nil
}
