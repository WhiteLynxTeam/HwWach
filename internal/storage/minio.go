package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Storage interface {
	Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (minio.UploadInfo, error)
	PresignedGetURL(ctx context.Context, objectName string, expiry time.Duration) (string, error)
	PresignedPutURL(ctx context.Context, objectName, contentType string, expiry time.Duration) (string, error)
	Delete(ctx context.Context, objectName string) error
}

type minioStorage struct {
	client        *minio.Client
	presignClient *minio.Client // отдельный клиент для presigned URL (подключён к public endpoint)
	bucket        string
}

func NewMinioClient(endpoint, accessKey, secretKey string, useSSL bool, bucket string) (*minio.Client, error) {
	cli, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio init failed: %w", err)
	}

	ctx := context.Background()
	exists, err := cli.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("bucketExists check failed: %w", err)
	}
	if !exists {
		if err := cli.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("makeBucket failed: %w", err)
		}
	}

	return cli, nil
}

func NewMinioStorage(cli *minio.Client, bucket string, publicURL string, accessKey, secretKey string) Storage {
	var presignClient *minio.Client
	if publicURL != "" {
		parsedURL, err := url.Parse(publicURL)
		if err == nil {
			// Определяем useSSL на основе схемы
			useSSL := parsedURL.Scheme == "https"
			// Создаём отдельный клиент для presigned URL, подключённый к public endpoint
			presignClient, err = minio.New(parsedURL.Host, &minio.Options{
				Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
				Secure: useSSL,
			})
			if err != nil {
				// Если не получилось — просто не используем presignClient
				presignClient = nil
			}
		}
	}
	return &minioStorage{client: cli, presignClient: presignClient, bucket: bucket}
}

func (s *minioStorage) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (minio.UploadInfo, error) {
	return s.client.PutObject(ctx, s.bucket, objectName, reader, size, minio.PutObjectOptions{ContentType: contentType})
}

func (s *minioStorage) PresignedGetURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	reqParams := make(url.Values)
	u, err := s.client.PresignedGetObject(ctx, s.bucket, objectName, expiry, reqParams)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (s *minioStorage) PresignedPutURL(ctx context.Context, objectName, contentType string, expiry time.Duration) (string, error) {
	// Если есть отдельный клиент для presigned URL, используем его
	if s.presignClient != nil {
		u, err := s.presignClient.PresignedPutObject(ctx, s.bucket, objectName, expiry)
		if err != nil {
			return "", err
		}
		return u.String(), nil
	}

	// Фолбэк на основной клиент
	u, err := s.client.PresignedPutObject(ctx, s.bucket, objectName, expiry)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (s *minioStorage) Delete(ctx context.Context, objectName string) error {
	return s.client.RemoveObject(ctx, s.bucket, objectName, minio.RemoveObjectOptions{})
}
