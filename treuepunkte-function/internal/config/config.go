package config

import "os"

type Config struct {
	AppEnv  string
	AppPort string

	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string
}

func FromEnv() Config {
	return Config{
		AppEnv:  getenv("APP_ENV", "dev"),
		AppPort: getenv("APP_PORT", "8080"),
		DBHost:  getenv("DB_HOST", "127.0.0.1"),
		DBPort:  getenv("DB_PORT", "3306"),
		DBUser:  getenv("DB_USER", "treuepunkteuser"),
		DBPass:  getenv("DB_PASS", ""),
		DBName:  getenv("DB_NAME", "treuepunkteDB"),
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
