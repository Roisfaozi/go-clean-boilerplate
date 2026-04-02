import re

with open('internal/worker/handlers/webhook_handler.go', 'r') as f:
    content = f.read()

content = content.replace(
    '\tdefer resp.Body.Close()\n',
    '\tdefer func() {\n\t\tif err := resp.Body.Close(); err != nil {\n\t\t\th.log.WithContext(ctx).WithError(err).Error("failed to close webhook response body")\n\t\t}\n\t}()\n'
)

with open('internal/worker/handlers/webhook_handler.go', 'w') as f:
    f.write(content)
