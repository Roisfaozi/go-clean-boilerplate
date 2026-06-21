package tus

import "context"

type UploadEvent struct {
	UploadID string            // The Tus ID (e.g., abc-123)
	FileURL  string            // The Full S3 URL
	Metadata map[string]string // Metadata sent by client
}

// UploadHook must be implemented by Feature Modules (User, Project, etc)
type UploadHook interface {
	HandleUpload(ctx context.Context, event UploadEvent) error
}

type Registry struct {
	hooks map[string]UploadHook
}

func NewRegistry() *Registry {
	return &Registry{hooks: make(map[string]UploadHook)}
}

func (r *Registry) Register(uploadType string, hook UploadHook) {
	r.hooks[uploadType] = hook
}

func (r *Registry) Get(uploadType string) UploadHook {
	return r.hooks[uploadType]
}
