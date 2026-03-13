sed -i 's/defer resp.Body.Close()/defer func() { _ = resp.Body.Close() }()/' internal/worker/handlers/webhook_handler.go
sed -i 's/dbSQL.Close()/defer func() { _ = dbSQL.Close() }()/' internal/modules/user/repository/user_repository_test.go
