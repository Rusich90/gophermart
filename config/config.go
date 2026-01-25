package config

import (
	"flag"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	RunAddress              string
	DatabaseURI             string
	AccrualSystemAddress    string
	AuthSecret              string
	MigrationsPath          string
	PollInterval            int // Интервал опроса внешних систем в секундах
	MaxConcurrentRequests   int // Максимальное количество параллельных запросов к системе начислений
}

func InitConfig() *Config {
	_ = godotenv.Load()

	config := &Config{}

	flag.StringVar(&config.RunAddress, "a", "localhost:8080", "Адрес и порт запуска сервиса")
	flag.StringVar(&config.DatabaseURI, "d", "", "Адрес подключения к базе данных")
	flag.StringVar(&config.AccrualSystemAddress, "r", "", "Адрес системы расчёта начислений")
	flag.StringVar(&config.MigrationsPath, "m", "file://migrations", "Путь до миграций")
	flag.IntVar(&config.PollInterval, "poll-interval", 5, "Интервал опроса внешних систем в секундах")
	flag.IntVar(&config.MaxConcurrentRequests, "max-concurrent-requests", 10, "Максимальное количество параллельных запросов к системе начислений")

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

	if envPollInterval, exists := os.LookupEnv("POLL_INTERVAL"); exists {
		if interval, err := strconv.Atoi(envPollInterval); err == nil {
			config.PollInterval = interval
		}
	}

	if envMaxConcurrentRequests, exists := os.LookupEnv("MAX_CONCURRENT_REQUESTS"); exists {
		if maxRequests, err := strconv.Atoi(envMaxConcurrentRequests); err == nil {
			config.MaxConcurrentRequests = maxRequests
		}
	}

	return config
}
