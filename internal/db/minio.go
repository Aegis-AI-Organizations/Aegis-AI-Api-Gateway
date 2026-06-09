package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const PublicMinioEndpoint = "storage.aegis-ai.fr"
const defaultMinioRegion = "us-east-1"

type MinioClient struct {
	Client        *minio.Client
	PresignClient *minio.Client
	Bucket        string
}

func NewMinioClient() (*MinioClient, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	bucket := os.Getenv("MINIO_BUCKET")
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"

	if endpoint == "" || accessKey == "" || secretKey == "" || bucket == "" {
		return nil, fmt.Errorf("missing MinIO configuration (MINIO_ENDPOINT, ACCESS_KEY, SECRET_KEY, BUCKET)")
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	presignClient, err := minio.New(PublicMinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: true,
		Region: defaultMinioRegion,
	})
	if err != nil {
		return nil, err
	}

	return &MinioClient{
		Client:        client,
		PresignClient: presignClient,
		Bucket:        bucket,
	}, nil
}

func (m *MinioClient) GeneratePresignedPutURL(ctx context.Context, objectName string) (string, error) {
	// URL expires after 15 minutes as per requirements
	expiry := time.Duration(15) * time.Minute

	presignClient := m.PresignClient
	if presignClient == nil {
		presignClient = m.Client
	}

	presignedURL, err := presignClient.PresignedPutObject(ctx, m.Bucket, objectName, expiry)
	if err != nil {
		return "", err
	}

	return presignedURL.String(), nil
}
