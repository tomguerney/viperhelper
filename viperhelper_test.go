package viperhelper

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ViperHelperSuite struct {
	suite.Suite
}

func (suite *ViperHelperSuite) TestOne() {
	suite.Assert().Equal("this", "this")
}

func TestViperHelperSuite(t *testing.T) {
	suite.Run(t, new(ViperHelperSuite))
}
