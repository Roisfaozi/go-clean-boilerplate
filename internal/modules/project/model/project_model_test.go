package model_test

import (
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/model"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestProjectRequest_XSS_Protection(t *testing.T) {
	// Setup validator with custom rules (xss)
	v := validator.New()
	err := validation.RegisterCustomValidations(v)
	assert.NoError(t, err)

	t.Run("CreateProjectRequest REJECTS XSS", func(t *testing.T) {
		req := model.CreateProjectRequest{
			Name:   "<script>alert('xss')</script>",
			Domain: "example.com",
		}

		err := v.Struct(req)
		assert.Error(t, err, "Validation must fail for XSS payload in Name")
	})

	t.Run("CreateProjectRequest REJECTS XSS in Domain", func(t *testing.T) {
		req := model.CreateProjectRequest{
			Name:   "Safe Project",
			Domain: "<script>alert('xss')</script>",
		}

		err := v.Struct(req)
		assert.Error(t, err, "Validation must fail for XSS payload in Domain")
	})

	t.Run("UpdateProjectRequest REJECTS XSS", func(t *testing.T) {
		req := model.UpdateProjectRequest{
			Name:   "<script>alert('xss')</script>",
			Domain: "example.com",
		}

		err := v.Struct(req)
		assert.Error(t, err, "Validation must fail for XSS payload in Name")
	})

	t.Run("UpdateProjectRequest REJECTS XSS in Domain", func(t *testing.T) {
		req := model.UpdateProjectRequest{
			Name:   "Safe Project",
			Domain: "<script>alert('xss')</script>",
		}

		err := v.Struct(req)
		assert.Error(t, err, "Validation must fail for XSS payload in Domain")
	})

	t.Run("CreateProjectRequest ACCEPTS safe input", func(t *testing.T) {
		req := model.CreateProjectRequest{
			Name:   "Safe Project Name",
			Domain: "example.com",
		}

		err := v.Struct(req)
		assert.NoError(t, err)
	})
}
