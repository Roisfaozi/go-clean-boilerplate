#!/bin/bash
sed -i 's/if err.Error() == "user not found" {\n\t\t\t\treturn nil\n\t\t\t}/if err.Error() == "user not found" {\n\t\t\t\treturn exception.ErrNotFound\n\t\t\t}/' internal/modules/user/usecase/user_usecase.go
