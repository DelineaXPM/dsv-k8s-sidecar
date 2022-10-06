package auth_test

import (
	"testing"

	"github.com/DelineaXPM/dsv-k8s-sidecar/pkg/auth"
	"github.com/DelineaXPM/dsv-k8s-sidecar/pkg/mocks"
	"github.com/ericchiang/k8s/apis/core/v1"
	metaV1 "github.com/ericchiang/k8s/apis/meta/v1"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

type AuthTestSuite struct {
	suite.Suite
	registry  *mocks.MockPodRegistry
	underTest auth.AuthService
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCacheSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}

func (suite *AuthTestSuite) SetupTest() {
	mockCtrl := gomock.NewController(suite.T())
	defer mockCtrl.Finish()

	suite.registry = mocks.NewMockPodRegistry(mockCtrl)
	suite.underTest = auth.NewAuthService("foo", suite.registry)
}

func (suite *AuthTestSuite) TestGetTokenValidPod() {
	ip, uid, name := "12356", "abc", "name"
	pod := &v1.Pod{
		Status: &v1.PodStatus{
			PodIP: &ip,
		},
		Metadata: &metaV1.ObjectMeta{
			Uid:  &uid,
			Name: &name,
		},
	}

	suite.registry.EXPECT().Get(gomock.Eq("name")).Return(pod)

	request := &auth.TokenRequest{
		PodName: name,
		PodIp:   ip,
	}

	result := suite.underTest.GetToken(request)

	suite.NotNil(result)
	suite.NotEmpty(result.Token)
}

func (suite *AuthTestSuite) TestGetUnaryInterceptor() {
}
