import re

with open("internal/modules/user/test/user_usecase_test.go", "r") as f:
    content = f.read()

# Replace assert.NoError with assert.ErrorIs for the "Error - User Not Found" block
new_content = re.sub(
    r'(t\.Run\("Error - User Not Found", func\(t \*testing\.T\) \{.*?)assert\.NoError\(t, err\)',
    r'\1assert.ErrorIs(t, err, exception.ErrNotFound)',
    content,
    flags=re.DOTALL | re.MULTILINE
)

with open("internal/modules/user/test/user_usecase_test.go", "w") as f:
    f.write(new_content)
