package env_test

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/DelineaXPM/dsv-k8s-sidecar/pkg/env"
	"github.com/DelineaXPM/dsv-k8s-sidecar/pkg/mocks"
	"github.com/DelineaXPM/dsv-k8s-sidecar/pkg/secrets"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

type EnvTestSuite struct {
	suite.Suite
	client    *mocks.MockDSVClient
	underTest env.EnvironmentAgent
}

func TestEnvSuite(t *testing.T) {
	env.SecretFile = "dsv.json"
	suite.Run(t, new(EnvTestSuite))
}

func (suite *EnvTestSuite) SetupTest() {
	mockCtrl := gomock.NewController(suite.T())
	defer mockCtrl.Finish()

	os.Setenv(env.SecretEnvName, "foo bar /us-east-1/baz")

	suite.client = mocks.NewMockDSVClient(mockCtrl)
	suite.underTest = env.CreateEnvironmentAgent(suite.client)
}

func (suite *EnvTestSuite) TearDownTest() {
	os.Setenv(env.SecretEnvName, "")
	os.Remove(env.SecretFile)
}

func (suite *EnvTestSuite) TestUpdateEnv() {
	secret := &secrets.Secret{
		Value: "test",
	}
	suite.client.EXPECT().GetSecret(gomock.Any(), gomock.Any()).Return(secret, nil).Times(3)

	suite.underTest.UpdateEnv()

	jsonFile, _ := os.Open(env.SecretFile)

	byteValue, _ := io.ReadAll(jsonFile)

	var result map[string]interface{}
	json.Unmarshal([]byte(byteValue), &result)

	keys := []string{"foo", "bar", "/us-east-1/baz"}
	for _, key := range keys {
		val, ok := result[key]
		suite.True(ok)
		suite.Equal("test", val)
	}

	jsonFile.Close()
}
