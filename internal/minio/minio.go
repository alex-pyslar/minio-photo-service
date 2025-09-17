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
	fmt.Printf("Инициализация MinIO клиента: endpoint=%s, useSSL=%v\n", endpoint, useSSL)
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к MinIO: %v", err)
	}

	ctx := context.Background()
	fmt.Printf("Проверка существования бакета: %s\n", bucket)
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки бакета: %v", err)
	}
	if !exists {
		fmt.Printf("Бакет %s не существует, создаём...\n", bucket)
		err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("ошибка создания бакета: %v", err)
		}
		fmt.Printf("Бакет %s успешно создан\n", bucket)
	}

	return &Client{client: client, bucketName: bucket}, nil
}

// Upload загружает файл в MinIO и возвращает objectName и presigned URL
func (c *Client) Upload(ctx context.Context, file io.Reader, size int64, filename, contentType string) (string, string, error) {
	// Генерируем уникальное имя объекта
	ext := filepath.Ext(filename)
	objectName := uuid.New().String() + ext
	fmt.Printf("Загрузка объекта: %s, размер: %d\n", objectName, size)

	// Загружаем с явным указанием времени в GMT
	_, err := c.client.PutObject(ctx, c.bucketName, objectName, file, size, minio.PutObjectOptions{
		ContentType: contentType,
		UserMetadata: map[string]string{
			"Last-Modified": time.Now().UTC().Format(time.RFC1123), // Для совместимости
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

// GetPresignedURL генерирует presigned URL для объекта
func (c *Client) GetPresignedURL(ctx context.Context, objectName string) (string, error) {
	// Проверяем существование объекта
	_, err := c.client.StatObject(ctx, c.bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return "", fmt.Errorf("ошибка проверки объекта: %v", err)
	}

	// Генерируем presigned URL (7 дней)
	url, err := c.client.PresignedGetObject(ctx, c.bucketName, objectName, time.Hour*24*7, nil)
	if err != nil {
		return "", fmt.Errorf("ошибка генерации URL: %v", err)
	}

	fmt.Printf("Сгенерирована ссылка для %s: %s\n", objectName, url.String())
	return url.String(), nil
}
