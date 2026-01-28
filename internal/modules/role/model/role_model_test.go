package model_test

import (
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/model"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateRoleRequest_Validation(t *testing.T) {
	validate := validator.New()
	err := validation.RegisterCustomValidations(validate)
	require.NoError(t, err)

	tests := []struct {
		name    string
		request model.CreateRoleRequest
		wantErr bool
	}{
		{
			name: "Valid Role",
			request: model.CreateRoleRequest{
				Name:        "Admin",
				Description: "Administrator role",
			},
			wantErr: false,
		},
		{
			name: "XSS in Name",
			request: model.CreateRoleRequest{
				Name:        "<script>alert(1)</script>",
				Description: "Malicious role",
			},
			wantErr: true,
		},
		{
			name: "XSS in Description",
			request: model.CreateRoleRequest{
				Name:        "Safe Role",
				Description: "<script>alert(1)</script>",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
