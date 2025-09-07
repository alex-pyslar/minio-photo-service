package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config хранит настройки сервиса
type Config struct {
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	MinioUseSSL    bool
	Port           string
}

// Load загружает конфигурацию из .env или переменных окружения
func Load() (*Config, error) {
	// Загружаем .env файл, если он существует
	_ = godotenv.Load() // Игнорируем ошибку, если .env отсутствует

	cfg := &Config{
		MinioEndpoint:  os.Getenv("MINIO_ENDPOINT"),
		MinioAccessKey: os.Getenv("MINIO_ACCESS_KEY"),
		MinioSecretKey: os.Getenv("MINIO_SECRET_KEY"),
		MinioBucket:    os.Getenv("MINIO_BUCKET"),
		MinioUseSSL:    os.Getenv("MINIO_USE_SSL") == "true",
		Port:           os.Getenv("PORT"),
	}

	// Устанавливаем порт по умолчанию
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	// Проверяем обязательные поля
	if cfg.MinioEndpoint == "" || cfg.MinioAccessKey == "" || cfg.MinioSecretKey == "" || cfg.MinioBucket == "" {
		return nil, fmt.Errorf("отсутствуют обязательные переменные окружения: MINIO_ENDPOINT, MINIO_ACCESS_KEY, MINIO_SECRET_KEY, MINIO_BUCKET")
	}

	return cfg, nil
}
