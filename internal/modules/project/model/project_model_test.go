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

	testCases := []struct {
		name      string
		request   interface{}
		expectErr bool
		msg       string
	}{
		{
			name: "CreateProjectRequest REJECTS XSS in Name",
			request: model.CreateProjectRequest{
				Name:   "<script>alert('xss')</script>",
				Domain: "example.com",
			},
			expectErr: true,
			msg:       "Validation must fail for XSS payload in Name",
		},
		{
			name: "CreateProjectRequest REJECTS XSS in Domain",
			request: model.CreateProjectRequest{
				Name:   "Safe Project",
				Domain: "<script>alert('xss')</script>",
			},
			expectErr: true,
			msg:       "Validation must fail for XSS payload in Domain",
		},
		{
			name: "UpdateProjectRequest REJECTS XSS in Name",
			request: model.UpdateProjectRequest{
				Name:   "<script>alert('xss')</script>",
				Domain: "example.com",
			},
			expectErr: true,
			msg:       "Validation must fail for XSS payload in Name",
		},
		{
			name: "UpdateProjectRequest REJECTS XSS in Domain",
			request: model.UpdateProjectRequest{
				Name:   "Safe Project",
				Domain: "<script>alert('xss')</script>",
			},
			expectErr: true,
			msg:       "Validation must fail for XSS payload in Domain",
		},
		{
			name: "CreateProjectRequest ACCEPTS safe input",
			request: model.CreateProjectRequest{
				Name:   "Safe Project Name",
				Domain: "example.com",
			},
			expectErr: false,
			msg:       "Validation should pass for safe input",
		},
		{
			name: "UpdateProjectRequest ACCEPTS safe input",
			request: model.UpdateProjectRequest{
				Name:   "Safe Project Name",
				Domain: "example.com",
			},
			expectErr: false,
			msg:       "Validation should pass for safe input",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Struct(tc.request)
			if tc.expectErr {
				assert.Error(t, err, tc.msg)
			} else {
				assert.NoError(t, err, tc.msg)
			}
		})
	}
}
