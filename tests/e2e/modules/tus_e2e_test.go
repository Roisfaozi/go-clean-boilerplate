//go:build e2e
// +build e2e

package modules

import (
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/tests/e2e/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTUS_E2E_Lifecycle(t *testing.T) {
	server := setup.SetupTusTestServer(t)
	defer server.Cleanup()

	adminToken := setup.CreateAdminAndLogin(t, server)

	t.Run("Auth Guard - Enforce JWT", func(t *testing.T) {
		resp := server.Client.POST("/api/v1/upload/files/", nil)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Resumable Upload - Full Flow", func(t *testing.T) {
		reqHeaders := []setup.RequestOption{
			setup.WithAuth(adminToken),
			setup.WithHeader("Tus-Resumable", "1.0.0"),
			setup.WithHeader("Upload-Length", "10"),
			setup.WithHeader("Upload-Metadata", "filename dGVzdC50eHQ=,type YXZhdGFy"), // filename test.txt, type avatar
		}

		resp := server.Client.POST("/api/v1/upload/files/", nil, reqHeaders...)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		location := resp.Header.Get("Location")
		require.NotEmpty(t, location)

		client := &http.Client{}

		patchReq, _ := http.NewRequest("PATCH", location, strings.NewReader("hello"))
		patchReq.Header.Set("Authorization", "Bearer "+adminToken)
		patchReq.Header.Set("Tus-Resumable", "1.0.0")
		patchReq.Header.Set("Upload-Offset", "0")
		patchReq.Header.Set("Content-Type", "application/offset+octet-stream")

		patchResp, err := client.Do(patchReq)
		require.NoError(t, err)
		defer patchResp.Body.Close()
		require.Equal(t, http.StatusNoContent, patchResp.StatusCode)

		newOffset := patchResp.Header.Get("Upload-Offset")
		assert.Equal(t, "5", newOffset)

		patchReq2, _ := http.NewRequest("PATCH", location, strings.NewReader("world"))
		patchReq2.Header.Set("Authorization", "Bearer "+adminToken)
		patchReq2.Header.Set("Tus-Resumable", "1.0.0")
		patchReq2.Header.Set("Upload-Offset", "5")
		patchReq2.Header.Set("Content-Type", "application/offset+octet-stream")

		patchResp2, err := client.Do(patchReq2)
		require.NoError(t, err)
		defer patchResp2.Body.Close()
		require.Equal(t, http.StatusNoContent, patchResp2.StatusCode)

		finalOffset := patchResp2.Header.Get("Upload-Offset")
		assert.Equal(t, "10", finalOffset)
	})

	t.Run("Concurrency - Multiple Users", func(t *testing.T) {
		adminToken2 := setup.CreateAdminAndLogin(t, server)
		var wg sync.WaitGroup
		wg.Add(2)

		// Upload 1
		go func() {
			defer wg.Done()
			reqHeaders := []setup.RequestOption{
				setup.WithAuth(adminToken),
				setup.WithHeader("Tus-Resumable", "1.0.0"),
				setup.WithHeader("Upload-Length", "5"),
				setup.WithHeader("Upload-Metadata", "filename file1.txt,type avatar"),
			}
			resp := server.Client.POST("/api/v1/upload/files/", nil, reqHeaders...)
			assert.Equal(t, http.StatusCreated, resp.StatusCode)
		}()

		// Upload 2
		go func() {
			defer wg.Done()
			reqHeaders := []setup.RequestOption{
				setup.WithAuth(adminToken2),
				setup.WithHeader("Tus-Resumable", "1.0.0"),
				setup.WithHeader("Upload-Length", "5"),
				setup.WithHeader("Upload-Metadata", "filename file2.txt,type avatar"),
			}
			resp := server.Client.POST("/api/v1/upload/files/", nil, reqHeaders...)
			assert.Equal(t, http.StatusCreated, resp.StatusCode)
		}()

		wg.Wait()
	})

}
