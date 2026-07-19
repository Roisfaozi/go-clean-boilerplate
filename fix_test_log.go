package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func main() {
	filePath := "tests/integration/modules/role_integration_test.go"
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	oldStr := `t.Logf("KNOWN GAP: Casbin grouping still contains deleted role '%s': %v — cleanup not cascaded by usecase.Delete", created.Name, roleStillInCasbin)`
	newStr := `assert.False(t, roleStillInCasbin, "Role should be completely removed from Casbin grouping policies")`

	newContent := strings.Replace(string(content), oldStr, newStr, 1)

	err = ioutil.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		fmt.Printf("Error writing: %v\n", err)
		return
	}
	fmt.Println("Fixed role_integration_test.go")
}
