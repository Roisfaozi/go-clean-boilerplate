package model_test

import (
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/model"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestCreateRoleRequest_Validation(t *testing.T) {
	v := validator.New()
	_ = validation.RegisterCustomValidations(v)

	tests := []struct {
		name    string
		req     model.CreateRoleRequest
		wantErr bool
	}{
		{
			name: "Valid Role",
			req: model.CreateRoleRequest{
				Name:        "Admin",
				Description: "Administrator role",
			},
			wantErr: false,
		},
		{
			name: "XSS in Name (Vulnerability Check)",
			req: model.CreateRoleRequest{
				Name:        "<script>alert(1)</script>",
				Description: "Malicious role",
			},
			wantErr: true,
		},
		{
			name: "XSS in Description",
			req: model.CreateRoleRequest{
				Name:        "Admin",
				Description: "<script>alert(1)</script>",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Struct(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
