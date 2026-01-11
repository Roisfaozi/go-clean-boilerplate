package config

import (
	"fmt"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/storage"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/storage/local"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/storage/s3"
)

func NewStorageProvider(cfg *AppConfig) (storage.Provider, error) {
	switch cfg.Storage.Driver {
	case "local":
		return local.NewLocalStorage(cfg.Storage.Local.RootPath, cfg.Storage.Local.BaseURL)
	case "s3":
		return s3.NewS3Storage(s3.S3Config{
			Endpoint:       cfg.Storage.S3.Endpoint,
			Region:         cfg.Storage.S3.Region,
			Bucket:         cfg.Storage.S3.Bucket,
			AccessKey:      cfg.Storage.S3.AccessKey,
			SecretKey:      cfg.Storage.S3.SecretKey,
			UseSSL:         cfg.Storage.S3.UseSSL,
			ForcePathStyle: cfg.Storage.S3.ForcePathStyle,
		})
	default:
		return nil, fmt.Errorf("unsupported storage driver: %s", cfg.Storage.Driver)
	}
}
