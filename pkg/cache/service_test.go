package cache_test

import (
	"testing"

	"github.com/DelineaXPM/dsv-k8s-sidecar/pkg/cache"
	"github.com/stretchr/testify/suite"
)

type CacheTestSuite struct {
	suite.Suite
	underTest cache.Cache
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCacheSuite(t *testing.T) {
	suite.Run(t, new(CacheTestSuite))
}

func (suite *CacheTestSuite) SetupTest() {
	suite.underTest = cache.CreateMemoryCache()
}

func (suite *CacheTestSuite) TestAll() {
	suite.underTest.Set("foo", "bar")
	suite.underTest.Set("bar", 1)

	keys := suite.underTest.KeySet()
	suite.Len(keys, 2)

	val1 := suite.underTest.Get("foo").(string)
	suite.Equal("bar", val1)

	val2 := suite.underTest.Get("bar").(int)
	suite.Equal(1, val2)
}
