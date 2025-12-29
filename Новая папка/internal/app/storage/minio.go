package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	Client  *minio.Client
	Bucket  string
	BaseURL string
}

// NewMinioClient создаёт подключение к MinIO
func NewMinioClient() *MinioClient {
	endpoint := "localhost:9000"
	accessKeyID := "minioadmin"
	secretAccessKey := "minioadmin"
	useSSL := false
	bucket := "planets"

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalf("Ошибка подключения к MinIO: %v", err)
	}

	exists, err := client.BucketExists(context.Background(), bucket)
	if err != nil {
		log.Fatalf("Ошибка проверки бакета: %v", err)
	}
	if !exists {
		err = client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{})
		if err != nil {
			log.Fatalf("Ошибка создания бакета: %v", err)
		}
		log.Printf("Создан новый бакет: %s", bucket)
	}

	return &MinioClient{
		Client:  client,
		Bucket:  bucket,
		BaseURL: "http://localhost:9000/" + bucket,
	}
}

// GetFileURL возвращает полный URL файла
func (m *MinioClient) GetFileURL(objectName string) string {
	return fmt.Sprintf("%s/%s", m.BaseURL, objectName)
}
