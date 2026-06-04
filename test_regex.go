package main

import (
	"fmt"
	"regexp"
)

func main() {
	pattern := regexp.QuoteMeta("SELECT * FROM `webhooks` WHERE organization_id = ? AND ((webhooks.organization_id IS NULL OR NOT EXISTS (SELECT 1 FROM organizations WHERE organizations.id = webhooks.organization_id AND organizations.deleted_at IS NOT NULL AND organizations.deleted_at <> 0))) AND `webhooks`.`deleted_at` IS NULL")
	fmt.Println(pattern)
}
