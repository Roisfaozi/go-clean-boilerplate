package tus

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/telemetry"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sirupsen/logrus"
	"github.com/tus/tusd/v2/pkg/filestore"
	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/s3store"
)

type Config struct {
	StorageDriver string
	LocalRootPath string
	S3Bucket      string
	S3Endpoint    string
	BasePath      string
}

func NewHandler(cfg Config, registry *Registry, s3Client *s3.Client, log *logrus.Logger) (*handler.Handler, error) {
	var store handler.DataStore
	if cfg.StorageDriver == "s3" {
		s3Store := s3store.New(cfg.S3Bucket, s3Client)
		store = s3Store
	} else {
		// Default to local file store
		err := os.MkdirAll(cfg.LocalRootPath, 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to create tus directory: %w", err)
		}
		fileStore := filestore.New(cfg.LocalRootPath)
		store = fileStore
	}

	// Create Composer
	composer := handler.NewStoreComposer()
	if extendedStore, ok := store.(interface{ UseIn(*handler.StoreComposer) }); ok {
		extendedStore.UseIn(composer)
	} else {
		composer.UseCore(store)
	}

	// Create Handler with Notifications Enabled
	tusHandler, err := handler.NewHandler(handler.Config{
		BasePath:              cfg.BasePath,
		StoreComposer:         composer,
		NotifyCompleteUploads: true,
		PreUploadCreateCallback: func(hook handler.HookEvent) (handler.HTTPResponse, handler.FileInfoChanges, error) {
			resp, changes, err := BindAuthenticatedMetadata(hook)
			if err != nil {
				return resp, changes, err
			}

			if resp, _, err := ValidateUploadMetadata(changes.MetaData, registry); err != nil {
				return resp, handler.FileInfoChanges{}, err
			}

			return resp, changes, nil
		},
	})
	if err != nil {
		return nil, err
	}

	// Background Dispatcher
	go func() {
		for {
			event := <-tusHandler.CompleteUploads
			dispatchCompletedUpload(context.Background(), cfg, registry, store, event, log)
		}
	}()

	return tusHandler, nil
}

func dispatchCompletedUpload(ctx context.Context, cfg Config, registry *Registry, store handler.DataStore, event handler.HookEvent, log *logrus.Logger) {
	if registry == nil {
		return
	}
	meta := event.Upload.MetaData
	uploadType := meta["type"]

	hook := registry.Get(uploadType)
	if hook == nil {
		return
	}

	var fileURL string
	if cfg.StorageDriver == "s3" {
		fileURL = fmt.Sprintf("%s/%s/%s", cfg.S3Endpoint, cfg.S3Bucket, event.Upload.ID)
	} else {
		fileURL = fmt.Sprintf("%s/%s", cfg.BasePath, event.Upload.ID)
	}

	start := time.Now()
	// Dispatch to specific module
	err := hook.HandleUpload(ctx, UploadEvent{
		UploadID: event.Upload.ID,
		FileURL:  fileURL,
		Metadata: meta,
	})
	duration := time.Since(start)

	status := "success"
	if err != nil {
		status = "error"
		if log != nil {
			log.Errorf("Hook error for %s: %v", uploadType, err)
		} else {
			fmt.Printf("Hook error for %s: %v\n", uploadType, err)
		}
		cleanupFailedCompletedUpload(ctx, store, event.Upload.ID, log)
	}

	telemetry.RecordTUSUpload(cfg.StorageDriver, status, event.Upload.Size, duration)
}

func cleanupFailedCompletedUpload(ctx context.Context, store handler.DataStore, uploadID string, log *logrus.Logger) {
	terminater, ok := store.(handler.TerminaterDataStore)
	if !ok {
		if log != nil {
			log.Warnf("TUS store does not support termination after hook failure for upload %s", uploadID)
		}
		return
	}

	upload, err := store.GetUpload(ctx, uploadID)
	if err != nil {
		if log != nil {
			log.WithError(err).Warnf("Failed to load completed upload %s for cleanup after hook failure", uploadID)
		}
		return
	}

	if err := terminater.AsTerminatableUpload(upload).Terminate(ctx); err != nil {
		if log != nil {
			log.WithError(err).Warnf("Failed to terminate completed upload %s after hook failure", uploadID)
		}
		return
	}

	if log != nil {
		log.Warnf("Terminated completed upload %s after hook failure", uploadID)
	}
}
