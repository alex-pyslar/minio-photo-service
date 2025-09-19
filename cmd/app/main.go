package main

import (
	"log"
	"net/http"

	"minio-photo-service/internal/config"
	"minio-photo-service/internal/handler"
	"minio-photo-service/internal/minio"

	"github.com/gorilla/mux"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Инициализируем MinIO клиент
	minioClient, err := minio.NewClient(cfg.MinioEndpoint, cfg.MinioAccessKey, cfg.MinioSecretKey, cfg.MinioBucket, cfg.MinioUseSSL)
	if err != nil {
		log.Fatalf("Ошибка инициализации MinIO: %v", err)
	}

	// Настраиваем роутер
	r := mux.NewRouter()
	r.Use(handler.CorsMiddleware) // Применяем CORS ко всем запросам

	handler.Register(r, minioClient)

	// Запускаем сервер
	log.Printf("Сервис запущен на :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
