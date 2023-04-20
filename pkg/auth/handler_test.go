package auth_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DelineaXPM/dsv-k8s-sidecar/pkg/auth"
	"github.com/DelineaXPM/dsv-k8s-sidecar/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestAuthHandler_GetToken(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	authService := mocks.NewMockAuthService(mockCtrl)
	underTest := auth.NewAuthHandler(authService)

	// Arrange.
	tokenReq := `{"podName": "1234", "podIp": "abc"}`

	req, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(tokenReq))
	require.NoError(t, err)

	rr := httptest.NewRecorder()

	authService.EXPECT().GetToken(gomock.Any()).Return(
		&auth.TokenResponse{
			Token: "foo",
		},
	)

	// Act.
	underTest.GetToken(rr, req)

	// Assert.
	//nolint:bodyclose // The response body is ok not to close in this case.
	resp := rr.Result()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, `{"token":"foo"}`, string(respBody))

	result := &auth.TokenResponse{}
	err = json.Unmarshal(respBody, result)
	require.NoError(t, err)
	require.Equal(t, "foo", result.Token)
}
