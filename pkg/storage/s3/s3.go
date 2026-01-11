package s3

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Storage struct {
	Client     *s3.Client
	Bucket     string
	Endpoint   string
	IsPublic   bool // If true, generates public URLs instead of presigned
	Region     string
	BaseURL    string // Optional override for public URL (e.g. CDN)
}

type S3Config struct {
	Endpoint       string
	Region         string
	Bucket         string
	AccessKey      string
	SecretKey      string
	UseSSL         bool
	ForcePathStyle bool
}

func NewS3Storage(cfg S3Config) (*S3Storage, error) {
	ctx := context.TODO()

	// Configure Custom Endpoint Resolver if specific endpoint is provided (MinIO, R2)
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if cfg.Endpoint != "" {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           cfg.Endpoint,
				SigningRegion: cfg.Region,
			}, nil
		}
		// Returning EndpointNotFoundError allows the service to fallback to its default resolution
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")),
		awsconfig.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.ForcePathStyle // Required for MinIO
	})

	return &S3Storage{
		Client:   client,
		Bucket:   cfg.Bucket,
		Endpoint: cfg.Endpoint,
		Region:   cfg.Region,
	}, nil
}

func (s *S3Storage) UploadFile(ctx context.Context, file io.Reader, filename string, contentType string) (string, error) {
	_, err := s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.Bucket),
		Key:         aws.String(filename),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to s3: %w", err)
	}

	// For S3-compatible, usually we return key or construct URL
	return s.GetFileUrl(filename)
}

func (s *S3Storage) DeleteFile(ctx context.Context, filename string) error {
	_, err := s.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(filename),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from s3: %w", err)
	}
	return nil
}

func (s *S3Storage) GetFileUrl(filename string) (string, error) {
	// Generate Presigned URL (valid for 1 hour)
	presignClient := s3.NewPresignClient(s.Client)
	req, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(filename),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Hour
	})
	if err != nil {
		return "", fmt.Errorf("failed to presign url: %w", err)
	}
	return req.URL, nil
}
