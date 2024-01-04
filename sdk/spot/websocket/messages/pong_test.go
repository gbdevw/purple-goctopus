package messages

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

/*************************************************************************************************/
/* UNIT TEST SUITE                                                                               */
/*************************************************************************************************/

// Unit test suite for Pong
type PongUnitTestSuite struct {
	suite.Suite
}

// Run the unit test suite
func TestPongUnitTestSuite(t *testing.T) {
	suite.Run(t, new(PongUnitTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test marshalling an example Pong message from documentation into the same payload
func (suite *PongUnitTestSuite) TestPongMarshalJson() {
	// Payload to unmarshal
	payload := `{
		"event": "pong",
		"reqid": 42
	}`
	// Remove whitespaces
	payload = matchesWhitespacesRegex.ReplaceAllString(payload, "")
	// Unmarshal payload into target struct
	target := new(Pong)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Marshal target
	actual, err := json.Marshal(target)
	require.NoError(suite.T(), err)
	// Compare
	require.Equal(suite.T(), payload, string(actual))
}
