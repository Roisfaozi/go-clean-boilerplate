package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func main() {
	filePath := "internal/modules/permission/usecase/permission_usecase.go"
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	oldStr := `	_, err := enf.DeleteRole(roleName)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to delete role from Casbin: %v", err)
		return err
	}

	if _, inTx := tx.DBFromContext(ctx); !inTx {`

	// Restore explicitly removing grouping policies for the role to cascade properly
	newStr := `	_, err := enf.DeleteRole(roleName)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to delete role from Casbin: %v", err)
		return err
	}

	// Ensure we cascade grouping policy cleanup explicitly for all assignments of this role
	_, err = enf.RemoveFilteredGroupingPolicy(1, roleName)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to explicitly remove grouping policy for role: %v", err)
		return err
	}

	if _, inTx := tx.DBFromContext(ctx); !inTx {`

	newContent := strings.Replace(string(content), oldStr, newStr, 1)

	err = ioutil.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		fmt.Printf("Error writing: %v\n", err)
		return
	}
	fmt.Println("Fixed permission_usecase.go again")
}
