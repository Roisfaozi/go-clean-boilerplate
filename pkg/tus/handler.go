package tus

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sirupsen/logrus"
	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/s3store"
)

type Config struct {
    S3Bucket   string
    S3Endpoint string
    BasePath   string
}

func NewHandler(cfg Config, registry *Registry, s3Client *s3.Client, log *logrus.Logger) (*handler.Handler, error) {
    store := s3store.New(cfg.S3Bucket, s3Client)

    // Create Composer
    composer := handler.NewStoreComposer()
    composer.UseCore(store)

    // Create Handler with Notifications Enabled
    tusHandler, err := handler.NewHandler(handler.Config{
        BasePath:              cfg.BasePath,
        StoreComposer:         composer,
        NotifyCompleteUploads: true,
    })
    if err != nil {
        return nil, err
    }

    // Background Dispatcher
    go func() {
        for {
            event := <-tusHandler.CompleteUploads
            meta := event.Upload.MetaData
            uploadType := meta["type"]

            if hook := registry.Get(uploadType); hook != nil {
                fileURL := fmt.Sprintf("%s/%s/%s", cfg.S3Endpoint, cfg.S3Bucket, event.Upload.ID)
                
                // Dispatch to specific module
                err := hook.HandleUpload(context.Background(), UploadEvent{
                    UploadID: event.Upload.ID,
                    FileURL:  fileURL,
                    Metadata: meta,
                })
                if err != nil {
                    if log != nil {
                        log.Errorf("Hook error for %s: %v", uploadType, err)
                    } else {
                         fmt.Printf("Hook error for %s: %v\n", uploadType, err)
                    }
                }
            }
        }
    }()

    return tusHandler, nil
}
