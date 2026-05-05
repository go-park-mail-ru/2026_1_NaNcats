package s3

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
)

//go:generate mockgen -destination=mocks/s3_client_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/s3 S3ClientAPI
type S3ClientAPI interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

type s3Storage struct {
	client S3ClientAPI
	bucket string
	region string
}

func NewS3Storage(ctx context.Context, keyID, secretKey, bucket, region string) (FileStorage, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(keyID, secretKey, "")),
		config.WithBaseEndpoint("https://storage.yandexcloud.net"),
	)
	if err != nil {
		return nil, err
	}

	otelaws.AppendMiddlewares(&cfg.APIOptions)

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &s3Storage{
		client: client,
		bucket: bucket,
		region: region,
	}, nil
}

func (s *s3Storage) UploadFile(ctx context.Context, file io.Reader, filename, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(filename),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	return ("https://" + s.bucket + ".storage.yandexcloud.net/" + filename), nil
}

func (s *s3Storage) DeleteFile(ctx context.Context, fileURL string) error {
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return fmt.Errorf("invalid file URL: %w", err)
	}

	key := strings.TrimPrefix(parsedURL.Path, "/")

	if key == "" {
		return fmt.Errorf("empty object key in URL")
	}

	_, err = s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}
