package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	content, _ := os.ReadFile("tests/integration/scenarios/user_lifecycle_test.go")
	str := string(content)

	old := `	// Wait for any final async processing
	time.Sleep(500 * time.Millisecond)

	logs, _, err := auditUC.GetLogsDynamic(ctx, &querybuilder.DynamicFilter{
		Sort: &[]querybuilder.SortModel{{ColId: "CreatedAt", Sort: "asc"}},
	})
	require.NoError(t, err)

	var userLogs []auditModel.AuditLogResponse
	for _, l := range logs {
		if l.UserID == userID || l.EntityID == userID {
			userLogs = append(userLogs, l)
		}
	}

	require.GreaterOrEqual(t, len(userLogs), 4, "Should have at least 4 audit entries for this lifecycle")`

	new := `	// Wait for any final async processing
	// We use a polling mechanism to ensure the worker has processed all logs
	var userLogs []auditModel.AuditLogResponse
	maxRetries := 20

	for i := 0; i < maxRetries; i++ {
		logs, _, err := auditUC.GetLogsDynamic(ctx, &querybuilder.DynamicFilter{
			Sort: &[]querybuilder.SortModel{{ColId: "CreatedAt", Sort: "asc"}},
		})
		require.NoError(t, err)

		userLogs = nil
		for _, l := range logs {
			if l.UserID == userID || l.EntityID == userID {
				userLogs = append(userLogs, l)
			}
		}

		if len(userLogs) >= 4 {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	require.GreaterOrEqual(t, len(userLogs), 4, "Should have at least 4 audit entries for this lifecycle")`

	str = strings.Replace(str, old, new, 1)
	os.WriteFile("tests/integration/scenarios/user_lifecycle_test.go", []byte(str), 0644)
	fmt.Println("Patched successfully")
}
