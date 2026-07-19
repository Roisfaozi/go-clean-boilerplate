package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func main() {
	filePath := "internal/modules/user/usecase/user_usecase.go"
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	oldStr := `if err.Error() == "user not found" {
			return nil
		}`
	newStr := `if err.Error() == "user not found" {
			return exception.ErrNotFound
		}`

	newContent := strings.Replace(string(content), oldStr, newStr, 1)

	err = ioutil.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		fmt.Printf("Error writing: %v\n", err)
		return
	}
	fmt.Println("Fixed user_usecase.go")
}
