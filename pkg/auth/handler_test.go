package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DelineaXPM/dsv-k8s-sidecar/pkg/auth"
	"github.com/DelineaXPM/dsv-k8s-sidecar/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

type AuthHandlerTestSuite struct {
	suite.Suite
	underTest   auth.AuthHandler
	authService *mocks.MockAuthService
}

func TestAuthHandlerSuite(t *testing.T) {
	suite.Run(t, new(AuthHandlerTestSuite))
}

func (suite *AuthHandlerTestSuite) SetupTest() {
	mockCtrl := gomock.NewController(suite.T())
	defer mockCtrl.Finish()

	suite.authService = mocks.NewMockAuthService(mockCtrl)
	suite.underTest = auth.NewAuthHandler(suite.authService)
}

func (suite *AuthHandlerTestSuite) TestHandleAuthGet() {
	// arrange
	tokenReq := &auth.TokenRequest{
		PodIp:   "1234",
		PodName: "abc",
	}

	tokenResp := &auth.TokenResponse{
		Token: "foo",
	}

	body, _ := json.Marshal(tokenReq)
	req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(body))

	rr := httptest.NewRecorder()

	suite.authService.EXPECT().GetToken(gomock.Any()).Return(tokenResp)

	// act
	suite.underTest.GetToken(rr, req)

	// assert
	suite.Equal(http.StatusOK, rr.Code)
	result := new(auth.TokenResponse)
	json.NewDecoder(rr.Body).Decode(result)

	suite.NotNil(result)
	suite.Equal("foo", result.Token)
}
