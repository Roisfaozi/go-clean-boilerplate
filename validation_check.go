package main

import (
	"fmt"
	"github.com/go-playground/validator/v10"
)

type Config struct {
	Port int `validate:"required"`
}

func main() {
	v := validator.New()
	c := Config{Port: 0}
	err := v.Struct(c)
	if err != nil {
		fmt.Println("Validation failed:", err)
	} else {
		fmt.Println("Validation passed")
	}
}
