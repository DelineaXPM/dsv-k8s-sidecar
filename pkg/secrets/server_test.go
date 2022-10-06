package secrets_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DelineaXPM/dsv-k8s-sidecar/pkg/mocks"
	"github.com/DelineaXPM/dsv-k8s-sidecar/pkg/secrets"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

type SecretTestSuite struct {
	suite.Suite
	client    *mocks.MockSecretClient
	underTest secrets.DsvServer
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestSecretSuite(t *testing.T) {
	suite.Run(t, new(SecretTestSuite))
}

func (suite *SecretTestSuite) SetupTest() {
	mockCtrl := gomock.NewController(suite.T())
	defer mockCtrl.Finish()

	suite.client = mocks.NewMockSecretClient(mockCtrl)
	suite.underTest = secrets.NewSecretServer(suite.client)
}

func (suite *SecretTestSuite) TestGetSecret() {
	reqSecret := &secrets.Secret{
		Path: "foo",
	}

	respSecret := &secrets.SecretResponseData{
		ID:   "a",
		Path: "foo",
		Value: map[string]interface{}{
			"username": "bar",
		},
	}

	suite.client.EXPECT().GetSecret("foo").Return(respSecret, nil)

	result, err := suite.underTest.GetSecret(context.TODO(), reqSecret)

	suite.NoError(err)
	suite.Equal(respSecret.ID, result.Id)
	suite.Equal(respSecret.Path, result.Path)
	suite.Equal("{\"username\":\"bar\"}", result.Value)
}

func (suite *SecretTestSuite) TestGetSecretClientError() {
	reqSecret := &secrets.Secret{
		Path: "foo",
	}

	clientError := &secrets.SecretClientError{
		Status: 500,
		Error:  errors.New("test response"),
	}
	suite.client.EXPECT().GetSecret("foo").Return(nil, clientError)

	result, err := suite.underTest.GetSecret(context.TODO(), reqSecret)

	suite.Error(err)
	suite.Nil(result)
}
