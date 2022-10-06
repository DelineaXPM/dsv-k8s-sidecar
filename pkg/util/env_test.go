package util_test

import (
	"os"
	"testing"

	"github.com/DelineaXPM/dsv-k8s-sidecar/pkg/util"
	"github.com/stretchr/testify/suite"
)

type EnvironmentTestSuite struct {
	suite.Suite
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCacheSuite(t *testing.T) {
	suite.Run(t, new(EnvironmentTestSuite))
}

func (suite *EnvironmentTestSuite) TestEnvDefault() {
	result := util.EnvString("ASBSDCET", "test")
	suite.Equal("test", result)
}

func (suite *EnvironmentTestSuite) TestEnv() {
	os.Setenv("ASBSDCET", "foo")
	result := util.EnvString("ASBSDCET", "test")
	suite.Equal("foo", result)
	os.Setenv("ASBSDCET", "")
}
