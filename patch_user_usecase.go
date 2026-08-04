package main

import (
	"io/ioutil"
	"strings"
)

func main() {
	content, _ := ioutil.ReadFile("internal/modules/user/usecase/user_usecase.go")
	s := string(content)

	s = strings.Replace(s, "if err.Error() == \"user not found\" {\n\t\t\treturn nil\n\t\t}", "if err.Error() == \"user not found\" {\n\t\t\treturn exception.ErrNotFound\n\t\t}", 1)

	ioutil.WriteFile("internal/modules/user/usecase/user_usecase.go", []byte(s), 0644)
}
