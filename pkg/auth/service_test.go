package auth_test

import (
	"testing"

	"github.com/DelineaXPM/dsv-k8s-sidecar/pkg/auth"
	"github.com/DelineaXPM/dsv-k8s-sidecar/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func testPod(name, ip string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metaV1.ObjectMeta{UID: "abc", Name: name},
		Status:     v1.PodStatus{PodIP: ip},
	}
}

func (suite *AuthTestSuite) TestGetTokenValidPod() {
	suite.registry.EXPECT().Get(gomock.Eq("name")).Return(testPod("name", "12356"))

	result := suite.underTest.GetToken(&auth.TokenRequest{PodName: "name", PodIp: "12356"})

	suite.NotNil(result)
	suite.NotEmpty(result.Token)
}

func (suite *AuthTestSuite) TestGetTokenRejects() {
	cases := map[string]struct {
		pod     *v1.Pod
		request auth.TokenRequest
	}{
		"unknown pod":        {nil, auth.TokenRequest{PodName: "name", PodIp: "12356"}},
		"IP mismatch":        {testPod("name", "12356"), auth.TokenRequest{PodName: "name", PodIp: "10.0.0.1"}},
		"unassigned pod IP":  {testPod("name", ""), auth.TokenRequest{PodName: "name", PodIp: ""}},
		"request without IP": {testPod("name", "12356"), auth.TokenRequest{PodName: "name"}},
	}

	for name, tc := range cases {
		suite.Run(name, func() {
			suite.registry.EXPECT().Get(gomock.Eq("name")).Return(tc.pod)
			suite.Nil(suite.underTest.GetToken(&tc.request))
		})
	}
}

func (suite *AuthTestSuite) TestGetUnaryInterceptor() {
}
