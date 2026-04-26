package storage

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Storage interface {
	PresignedGetURL(ctx context.Context, objectName string, expiry time.Duration) (string, error)
	PresignedPutURL(ctx context.Context, objectName, contentType string, expiry time.Duration) (string, error)
	Delete(ctx context.Context, objectName string) error
	GetPublicURL(objectName string) string
}

type MinioStorage struct {
	internalClient *minio.Client // minio:9000 — для операций
	externalClient *minio.Client // 149.154.65.57:9000 — для presigned URL
	bucket         string
	externalBaseURL *url.URL     // http://149.154.65.57:9000 — для публичных URL
}

func NewMinioStorage(internalEndpoint, externalEndpoint, accessKey, secretKey string, useSSL bool, bucket string) (Storage, error) {
	// Создаём внутренний клиент (для операций внутри Docker)
	internalClient, err := minio.New(internalEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
		Region: "us-east-1",
	})
	if err != nil {
		return nil, fmt.Errorf("internal client init failed: %w", err)
	}

	// Проверяем/создаём бакет через внутренний клиент
	ctx := context.Background()
	exists, err := internalClient.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("bucketExists check failed: %w", err)
	}
	if !exists {
		if err := internalClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("makeBucket failed: %w", err)
		}
	}

	// Создаём внешний клиент (для presigned URL)
	externalUseSSL := useSSL
	var parsedURL *url.URL
	if externalEndpoint != "" {
		var err error
		parsedURL, err = url.Parse(externalEndpoint)
		if err == nil {
			externalUseSSL = parsedURL.Scheme == "https"
			externalEndpoint = parsedURL.Host
		}
	}

	externalClient, err := minio.New(externalEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: externalUseSSL,
		Region: "us-east-1", // Явно указываем регион, чтобы избежать запроса GetBucketLocation
	})
	if err != nil {
		return nil, fmt.Errorf("external client init failed: %w", err)
	}

	return &MinioStorage{
		internalClient: internalClient,
		externalClient: externalClient,
		bucket:         bucket,
		externalBaseURL: parsedURL,
	}, nil
}

func (s *MinioStorage) PresignedGetURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	reqParams := make(url.Values)
	u, err := s.internalClient.PresignedGetObject(ctx, s.bucket, objectName, expiry, reqParams)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (s *MinioStorage) PresignedPutURL(ctx context.Context, objectName, contentType string, expiry time.Duration) (string, error) {
	u, err := s.externalClient.PresignedPutObject(ctx, s.bucket, objectName, expiry)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (s *MinioStorage) Delete(ctx context.Context, objectName string) error {
	return s.internalClient.RemoveObject(ctx, s.bucket, objectName, minio.RemoveObjectOptions{})
}

func (s *MinioStorage) GetPublicURL(objectName string) string {
	if s.externalBaseURL == nil {
		return objectName
	}
	return fmt.Sprintf("%s/%s/%s", s.externalBaseURL.String(), s.bucket, objectName)
}
