package model

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

type RoleResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type CreateRoleRequest struct {
	Name        string `json:"name" validate:"required,max=50"`
	Description string `json:"description,omitempty" validate:"omitempty,xss"`
}

type UpdateRoleRequest struct {
	Description string `json:"description" validate:"required,xss"`
}

var (
	htmlTagRegex = regexp.MustCompile(`<[^>]*>`)
)

func RegisterCustomValidations(v *validator.Validate) error {
	if err := v.RegisterValidation("xss", func(fl validator.FieldLevel) bool {
		if fl.Field().Kind() != reflect.String {
			return false
		}

		safeTags := []string{"b", "i", "em", "strong", "u"}
		desc := fl.Field().String()

		temp := desc
		for _, tag := range safeTags {
			temp = regexp.MustCompile(fmt.Sprintf("<[/]?%s[^>]*>", tag)).ReplaceAllString(temp, "")
		}

		return !htmlTagRegex.MatchString(temp)
	}); err != nil {
		return err
	}

	return nil
}

func (r *CreateRoleRequest) Sanitize() {
	r.Name = strings.TrimSpace(r.Name)
	r.Description = htmlTagRegex.ReplaceAllString(strings.TrimSpace(r.Description), "")
}

func (r *UpdateRoleRequest) Sanitize() {
	r.Description = htmlTagRegex.ReplaceAllString(strings.TrimSpace(r.Description), "")
}
