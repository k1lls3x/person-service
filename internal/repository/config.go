package repository

import (
	"fmt"
	"os"
)
type Config struct{
	Host string
	Port string
	User string
	Password string
	Name string
}
func LoadConfigFromEnv() *Config {
	return &Config{
		Host:				os.Getenv("DB_HOST"),
		Port: 			os.Getenv("DB_PORT"),
		User: 			os.Getenv("DB_USER"),
		Password: 	os.Getenv("DB_PASSWORD"),
		Name:				os.Getenv("DB_NAME") ,
	}
}
func (cfg *Config) DSN() string{
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name,
	)
}
