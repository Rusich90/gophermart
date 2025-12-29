package config

import (
	"flag"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
	AuthSecret           string
	MigrationsPath       string
}

func InitConfig() *Config {
	_ = godotenv.Load()

	config := &Config{}

	flag.StringVar(&config.RunAddress, "a", "localhost:8080", "Адрес и порт запуска сервиса")
	flag.StringVar(&config.DatabaseURI, "d", "", "Адрес подключения к базе данных")
	flag.StringVar(&config.AccrualSystemAddress, "r", "", "Адрес системы расчёта начислений")
	flag.StringVar(&config.MigrationsPath, "m", "file://migrations", "Path to migrations")

	flag.Parse()

	if envRunAddr, exists := os.LookupEnv("RUN_ADDRESS"); exists {
		config.RunAddress = envRunAddr
	}

	if envDBURI, exists := os.LookupEnv("DATABASE_URI"); exists {
		config.DatabaseURI = envDBURI
	}

	if envAccrualAddr, exists := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); exists {
		config.AccrualSystemAddress = envAccrualAddr
	}

	if envAuthSecret, exists := os.LookupEnv("AUTH_SECRET"); exists {
		config.AuthSecret = envAuthSecret
	}

	if envMigrationsPath, exists := os.LookupEnv("MIGRATIONS_PATH"); exists {
		config.MigrationsPath = envMigrationsPath
	}

	return config
}
