package handler

import (
	"context"
	"fmt"
	"mime"
	"net/http"
	"path/filepath"

	"minio-photo-service/internal/minio"

	"github.com/gorilla/mux"
)

// Register регистрирует HTTP обработчики
func Register(r *mux.Router, minioClient *minio.Client) {
	r.HandleFunc("/upload", uploadHandler(minioClient)).Methods("POST")
	r.HandleFunc("/download/{objectName}", downloadHandler(minioClient)).Methods("GET")
}

func uploadHandler(minioClient *minio.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(32 << 20) // Лимит 32MB
		if err != nil {
			http.Error(w, "Ошибка парсинга формы", http.StatusBadRequest)
			return
		}

		file, handler, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Ошибка получения файла", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Проверяем расширение
		ext := filepath.Ext(handler.Filename)
		if ext != ".jpg" && ext != ".png" && ext != ".jpeg" {
			http.Error(w, "Поддерживаются только JPG/PNG", http.StatusBadRequest)
			return
		}

		// Загружаем в MinIO
		ctx := context.Background()
		objectName, url, err := minioClient.Upload(ctx, file, handler.Size, handler.Filename, handler.Header.Get("Content-Type"))
		if err != nil {
			http.Error(w, "Ошибка загрузки", http.StatusInternalServerError)
			return
		}

		// Возвращаем JSON
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"objectName": "%s", "url": "%s"}`, objectName, url)
	}
}

func downloadHandler(minioClient *minio.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		objectName := vars["objectName"]

		ctx := context.Background()
		data, err := minioClient.GetObject(ctx, objectName)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка получения объекта: %v", err), http.StatusInternalServerError)
			return
		}

		// Определяем MIME-тип на основе расширения файла
		mimeType := mime.TypeByExtension(filepath.Ext(objectName))
		if mimeType == "" {
			mimeType = "application/octet-stream" // Фallback для неизвестных типов
		}

		// Устанавливаем заголовки
		w.Header().Set("Content-Type", mimeType)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
		w.WriteHeader(http.StatusOK)

		// Отправляем содержимое объекта в ответ
		if _, err := w.Write(data); err != nil {
			fmt.Printf("Ошибка записи ответа: %v", err)
		}
	}
}
