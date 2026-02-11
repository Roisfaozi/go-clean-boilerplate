package tus

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockHook for testing
type MockHook struct {
	called bool
}

func (m *MockHook) HandleUpload(ctx context.Context, event UploadEvent) error {
	m.called = true
	return nil
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	// Positive Case
	registry := NewRegistry()
	hook := &MockHook{}
	uploadType := "avatar"

	registry.Register(uploadType, hook)
	retrieved := registry.Get(uploadType)

	assert.NotNil(t, retrieved)
	assert.Equal(t, hook, retrieved)
}

func TestRegistry_GetNonExistent(t *testing.T) {
	// Negative Case
	registry := NewRegistry()
	retrieved := registry.Get("non-existent")

	assert.Nil(t, retrieved)
}

func TestRegistry_OverwriteHook(t *testing.T) {
	// Edge Case
	registry := NewRegistry()
	hook1 := &MockHook{}
	hook2 := &MockHook{}
	uploadType := "avatar"

	registry.Register(uploadType, hook1)
	registry.Register(uploadType, hook2) // Overwrite

	retrieved := registry.Get(uploadType)
	assert.Equal(t, hook2, retrieved)
}

func TestRegistry_EmptyType(t *testing.T) {
	// Edge Case
	registry := NewRegistry()
	hook := &MockHook{}
	
	registry.Register("", hook)
	assert.Equal(t, hook, registry.Get(""))
}

func TestRegistry_TypeInjection(t *testing.T) {
	// Security Case: Ensure special characters don't crash the registry (standard map behavior)
	registry := NewRegistry()
	hook := &MockHook{}
	injectionKey := "../../etc/passwd%20' OR 1=1"

	registry.Register(injectionKey, hook)
	assert.Equal(t, hook, registry.Get(injectionKey))
}
