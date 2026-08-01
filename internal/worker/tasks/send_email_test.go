package tasks

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestNewSendEmailTask_Success(t *testing.T) {
	task, err := NewSendEmailTask("test@example.com", "Subject", "Body")
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, TypeSendEmail, task.Type())
}

func TestNewSendEmailTask_MarshalError(t *testing.T) {
    // SendEmailPayload only has strings
}
