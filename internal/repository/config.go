package repository

import (
	"fmt"
	"os"
)

type Config struct {
	Host              string
	Port              string
	User              string
	Password          string
	Name              string
	AgeAPIURL         string
	GenderAPIURL      string
	NationalityAPIURL string
	LogLevel          string
}

func LoadConfigFromEnv() *Config {
	return &Config{
		Host:              os.Getenv("DB_HOST"),
		Port:              os.Getenv("DB_PORT"),
		User:              os.Getenv("DB_USER"),
		Password:          os.Getenv("DB_PASSWORD"),
		Name:              os.Getenv("DB_NAME"),
		AgeAPIURL:         os.Getenv("AGE_API_URL"),
		GenderAPIURL:      os.Getenv("GENDER_API_URL"),
		NationalityAPIURL: os.Getenv("NATIONALITY_API_URL"),
		LogLevel:          os.Getenv("LOG_LEVEL"),
	}
}
func (cfg *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name,
	)
}
