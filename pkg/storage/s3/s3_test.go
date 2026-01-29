package s3_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/storage/s3"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockS3Client implements s3.S3ClientAPI
type MockS3Client struct {
	mock.Mock
}

func (m *MockS3Client) PutObject(ctx context.Context, params *awsS3.PutObjectInput, optFns ...func(*awsS3.Options)) (*awsS3.PutObjectOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*awsS3.PutObjectOutput), args.Error(1)
}

func (m *MockS3Client) DeleteObject(ctx context.Context, params *awsS3.DeleteObjectInput, optFns ...func(*awsS3.Options)) (*awsS3.DeleteObjectOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*awsS3.DeleteObjectOutput), args.Error(1)
}

// MockS3Presigner implements s3.S3PresignerAPI
type MockS3Presigner struct {
	mock.Mock
}

func (m *MockS3Presigner) PresignGetObject(ctx context.Context, params *awsS3.GetObjectInput, optFns ...func(*awsS3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v4.PresignedHTTPRequest), args.Error(1)
}

func TestUploadFile(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockClient := new(MockS3Client)
		mockPresigner := new(MockS3Presigner)
		storage := &s3.S3Storage{
			Client:    mockClient,
			Presigner: mockPresigner,
			Bucket:    "test-bucket",
		}
		filename := "test.txt"
		content := "content"

		mockClient.On("PutObject", mock.Anything, mock.MatchedBy(func(input *awsS3.PutObjectInput) bool {
			return *input.Bucket == "test-bucket" && *input.Key == filename
		}), mock.Anything).Return(&awsS3.PutObjectOutput{}, nil)

		mockPresigner.On("PresignGetObject", mock.Anything, mock.MatchedBy(func(input *awsS3.GetObjectInput) bool {
			return *input.Bucket == "test-bucket" && *input.Key == filename
		}), mock.Anything).Return(&v4.PresignedHTTPRequest{URL: "http://presigned.url/test.txt"}, nil)

		url, err := storage.UploadFile(context.Background(), strings.NewReader(content), filename, "text/plain")

		assert.NoError(t, err)
		assert.Equal(t, "http://presigned.url/test.txt", url)
		mockClient.AssertExpectations(t)
		mockPresigner.AssertExpectations(t)
	})

	t.Run("Upload Error", func(t *testing.T) {
		mockClient := new(MockS3Client)
		storage := &s3.S3Storage{
			Client: mockClient,
			Bucket: "test-bucket",
		}
		filename := "fail.txt"

		mockClient.On("PutObject", mock.Anything, mock.MatchedBy(func(input *awsS3.PutObjectInput) bool {
			return *input.Key == filename
		}), mock.Anything).Return(nil, errors.New("upload failed"))

		_, err := storage.UploadFile(context.Background(), strings.NewReader(""), filename, "text/plain")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to upload file to s3")
	})
}

func TestDeleteFile(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockClient := new(MockS3Client)
		storage := &s3.S3Storage{
			Client: mockClient,
			Bucket: "test-bucket",
		}
		mockClient.On("DeleteObject", mock.Anything, mock.MatchedBy(func(input *awsS3.DeleteObjectInput) bool {
			return *input.Bucket == "test-bucket" && *input.Key == "del.txt"
		}), mock.Anything).Return(&awsS3.DeleteObjectOutput{}, nil)

		err := storage.DeleteFile(context.Background(), "del.txt")
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		mockClient := new(MockS3Client)
		storage := &s3.S3Storage{
			Client: mockClient,
			Bucket: "test-bucket",
		}
		mockClient.On("DeleteObject", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("delete failed"))

		err := storage.DeleteFile(context.Background(), "del.txt")
		assert.Error(t, err)
	})
}

func TestGetFileUrl(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockPresigner := new(MockS3Presigner)
		storage := &s3.S3Storage{
			Presigner: mockPresigner,
			Bucket:    "test-bucket",
		}
		mockPresigner.On("PresignGetObject", mock.Anything, mock.MatchedBy(func(input *awsS3.GetObjectInput) bool {
			return *input.Bucket == "test-bucket" && *input.Key == "test.png"
		}), mock.Anything).Return(&v4.PresignedHTTPRequest{URL: "http://url.com/test.png"}, nil)

		url, err := storage.GetFileUrl("test.png")
		assert.NoError(t, err)
		assert.Equal(t, "http://url.com/test.png", url)
	})

	t.Run("Error", func(t *testing.T) {
		mockPresigner := new(MockS3Presigner)
		storage := &s3.S3Storage{
			Presigner: mockPresigner,
			Bucket:    "test-bucket",
		}
		mockPresigner.On("PresignGetObject", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("presign failed"))

		_, err := storage.GetFileUrl("test.png")
		assert.Error(t, err)
	})
}
