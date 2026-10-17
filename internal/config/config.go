package config

import (
	"errors"
	"os"
)

// переменные окружения с настройками
const (
	envDebug  string = "DEBUG"
	envDomain string = "DOMAIN"
	envPgDSN  string = "PG_DSN"
)

// Config - настройки сервиса из переменных окружения
type Config struct {
	// Debug - запуск для локальной разработки: задана переменная DEBUG, значение не важно
	Debug bool
	// Domain - домен куки входа: localhost локально, quiz-on.ru на проде
	Domain string
	// PgDSN - строка подключения к postgres
	PgDSN string
}

// Load - читает настройки из окружения. DOMAIN и PG_DSN обязательны
func Load() (Config, error) {
	_, debug := os.LookupEnv(envDebug)
	cfg := Config{
		Debug:  debug,
		Domain: os.Getenv(envDomain),
		PgDSN:  os.Getenv(envPgDSN),
	}

	if cfg.Domain == "" {
		return Config{}, errors.New("empty env variable " + envDomain)
	}
	if cfg.PgDSN == "" {
		return Config{}, errors.New("empty env variable " + envPgDSN)
	}

	return cfg, nil
}
