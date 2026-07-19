package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func main() {
	filePath := "internal/modules/user/test/user_usecase_test.go"
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	oldStr := `t.Run("Error - User Not Found", func(t *testing.T) {
		deps, uc := setupUserTest()
		deps.Repo.On("FindByID", mock.Anything, deleteReq.ID).Return(nil, errors.New("user not found"))

		err := uc.DeleteUser(context.Background(), actorUserID, deleteReq)

		assert.NoError(t, err)
		deps.Repo.AssertExpectations(t)
		deps.AuditUC.AssertNotCalled(t, "LogActivity", mock.Anything, mock.Anything)
	})`

	newStr := `t.Run("Error - User Not Found", func(t *testing.T) {
		deps, uc := setupUserTest()
		deps.Repo.On("FindByID", mock.Anything, deleteReq.ID).Return(nil, errors.New("user not found"))

		err := uc.DeleteUser(context.Background(), actorUserID, deleteReq)

		assert.ErrorIs(t, err, exception.ErrNotFound)
		deps.Repo.AssertExpectations(t)
		deps.AuditUC.AssertNotCalled(t, "LogActivity", mock.Anything, mock.Anything)
	})`

	newContent := strings.Replace(string(content), oldStr, newStr, 1)

	err = ioutil.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		fmt.Printf("Error writing: %v\n", err)
		return
	}
	fmt.Println("Fixed user_usecase_test.go")
}
