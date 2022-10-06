package auth_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/DelineaXPM/dsv-k8s-sidecar/pkg/auth"

	"github.com/stretchr/testify/suite"
)

type ClientTestSuite struct {
	suite.Suite
}

func TestClientSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

func (suite *ClientTestSuite) TestClient() {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := &auth.TokenResponse{
			Token: "foo",
		}

		body, _ := json.Marshal(b)
		w.Write(body)
	}))
	defer ts.Close()

	os.Setenv("AUTH_URL", ts.URL)
	creds, err := auth.GetToken("test", "123", "")

	suite.NoError(err)
	authMap, err := creds.GetRequestMetadata(context.TODO(), "doesn't matter")

	token, ok := authMap["Authorization"]
	suite.True(ok)
	suite.Equal("foo", token)
}

func (suite *ClientTestSuite) TestClientError() {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := &auth.TokenResponse{
			Token: "foo",
		}

		body, _ := json.Marshal(b)
		w.WriteHeader(500)
		w.Write(body)
	}))
	defer ts.Close()

	os.Setenv("AUTH_URL", ts.URL)
	_, err := auth.GetToken("test", "123", "")

	suite.Error(err)
}
