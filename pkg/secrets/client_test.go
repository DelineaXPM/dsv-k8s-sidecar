package secrets_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/DelineaXPM/dsv-k8s-sidecar/pkg/secrets"

	"github.com/stretchr/testify/suite"
)

type ClientTestSuite struct {
	suite.Suite
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestClientSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

func (suite *ClientTestSuite) TestGetSecret() {
	authTs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := &secrets.AuthResponse{
			Token: "foo",
		}

		body, _ := json.Marshal(b)
		w.Write(body)
	}))
	defer authTs.Close()

	secretTs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := &secrets.SecretResponseData{
			ID:   "abc",
			Path: "foo",
			Value: map[string]interface{}{
				"username": "bar",
			},
			Attributes: map[string]interface{}{
				"foo": "bar",
			},
		}

		body, _ := json.Marshal(b)
		w.Write(body)
	}))

	os.Setenv("DSV_API_URL", authTs.URL+"/%s")
	client := secrets.CreateSecretClient("foo", "id", "secret", "client_credentials")
	client.Close()

	client.SetSecretURL(secretTs.URL + "/%s/%s")
	result, err := client.GetSecret("foo")

	suite.Nil(err)
	suite.Equal("bar", result.Value.(map[string]interface{})["username"])

	// Check Cache
	secretTs.Close()
	result, err = client.GetSecret("foo")

	suite.Nil(err)
	suite.Equal("bar", result.Value.(map[string]interface{})["username"])
}
