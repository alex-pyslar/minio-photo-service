package minio

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Client обёртка для minio.Client
type Client struct {
	client     *minio.Client
	bucketName string
}

// NewClient создаёт новый MinIO клиент и проверяет/создаёт бакет
func NewClient(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*Client, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к MinIO: %v", err)
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки бакета: %v", err)
	}
	if !exists {
		err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("ошибка создания бакета: %v", err)
		}
	}

	return &Client{client: client, bucketName: bucket}, nil
}

// Upload загружает файл в MinIO и возвращает objectName и presigned URL
func (c *Client) Upload(ctx context.Context, file io.Reader, size int64, filename, contentType string) (string, string, error) {
	// Генерируем уникальное имя объекта
	ext := filepath.Ext(filename)
	objectName := uuid.New().String() + ext

	// Загружаем с явным указанием времени в GMT
	_, err := c.client.PutObject(ctx, c.bucketName, objectName, file, size, minio.PutObjectOptions{
		ContentType: contentType,
		UserMetadata: map[string]string{
			"Last-Modified": time.Now().UTC().Format(time.RFC1123), // Явно задаём корректный формат
		},
	})
	if err != nil {
		return "", "", fmt.Errorf("ошибка загрузки в MinIO: %v", err)
	}

	// Генерируем presigned URL (7 дней)
	url, err := c.client.PresignedGetObject(ctx, c.bucketName, objectName, time.Hour*24*7, nil)
	if err != nil {
		return "", "", fmt.Errorf("ошибка генерации URL: %v", err)
	}

	return objectName, url.String(), nil
}

// GetObject получает объект из MinIO
func (c *Client) GetObject(ctx context.Context, objectName string) ([]byte, error) {
	// Сначала проверяем метаданные объекта
	objInfo, err := c.client.StatObject(ctx, c.bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки объекта: %v", err)
	}
	// Логируем Last-Modified для отладки
	fmt.Printf("Объект: %s, Last-Modified: %s\n", objectName, objInfo.LastModified.Format(time.RFC1123))

	// Получаем объект
	object, err := c.client.GetObject(ctx, c.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("ошибка получения объекта: %v", err)
	}
	defer object.Close()

	// Читаем содержимое
	data, err := io.ReadAll(object)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения содержимого объекта: %v", err)
	}

	return data, nil
}
