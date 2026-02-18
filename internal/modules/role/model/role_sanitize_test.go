package model_test

import (
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/model"
	"github.com/stretchr/testify/assert"
)

func TestCreateRoleRequest_Sanitize(t *testing.T) {
	tests := []struct {
		name     string
		request  model.CreateRoleRequest
		expected model.CreateRoleRequest
	}{
		{
			name: "Sanitize Tags",
			request: model.CreateRoleRequest{
				Name:        "<script>alert(1)</script>",
				Description: "<b>Bold</b>",
			},
			expected: model.CreateRoleRequest{
				Name:        "&lt;script&gt;alert(1)&lt;/script&gt;",
				Description: "&lt;b&gt;Bold&lt;/b&gt;",
			},
		},
		{
			name: "Preserve Clean",
			request: model.CreateRoleRequest{
				Name:        "Admin",
				Description: "Administrator",
			},
			expected: model.CreateRoleRequest{
				Name:        "Admin",
				Description: "Administrator",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.request.Sanitize()
			assert.Equal(t, tt.expected.Name, tt.request.Name)
			assert.Equal(t, tt.expected.Description, tt.request.Description)
		})
	}
}
