import re
with open('tests/integration/modules/user_integration_test.go', 'r') as f:
    content = f.read()

content = content.replace("assert.Error(t, err)", "assert.ErrorIs(t, err, exception.ErrNotFound)")

if "github.com/Roisfaozi/go-clean-boilerplate/pkg/exception" not in content:
    content = content.replace('import (', 'import (\n\t"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"')

with open('tests/integration/modules/user_integration_test.go', 'w') as f:
    f.write(content)

with open('internal/modules/user/usecase/user_usecase.go', 'r') as f:
    lines = f.readlines()

in_delete = False
for i, line in enumerate(lines):
    if "func (u *userUseCaseImpl) DeleteUser(" in line:
        in_delete = True
    if in_delete and 'if err.Error() == "user not found" {' in line:
        lines[i+1] = '			return exception.ErrNotFound\n'
        break

with open('internal/modules/user/usecase/user_usecase.go', 'w') as f:
    f.writelines(lines)

with open('internal/modules/user/test/user_usecase_test.go', 'r') as f:
    lines = f.readlines()

in_delete_test = False
for i, line in enumerate(lines):
    if 't.Run("Error - User Not Found", func(t *testing.T) {' in line:
        in_delete_test = True
    if in_delete_test and 'err := uc.DeleteUser(' in line:
        pass
    if in_delete_test and 'assert.NoError(t, err)' in line:
        # Check if we are inside DeleteUser's tests by looking back
        for prev in reversed(lines[:i]):
            if 'func TestUserUseCase_DeleteUser(' in prev:
                lines[i] = '			assert.ErrorIs(t, err, exception.ErrNotFound)\n'
                in_delete_test = False
                break
            elif 'func TestUserUseCase_' in prev:
                in_delete_test = False
                break

with open('internal/modules/user/test/user_usecase_test.go', 'w') as f:
    f.writelines(lines)
