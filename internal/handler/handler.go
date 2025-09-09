package handler

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"

	"minio-photo-service/internal/minio"

	"github.com/gorilla/mux"
)

// Register регистрирует HTTP обработчики
func Register(r *mux.Router, minioClient *minio.Client) {
	r.PathPrefix("/api/photo").Subrouter()
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
		url, err := minioClient.GetPresignedURL(ctx, objectName)
		if err != nil {
			http.Error(w, "Ошибка получения URL", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}
