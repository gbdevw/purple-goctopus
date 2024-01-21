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

// Unit test suite for Ping
type PingUnitTestSuite struct {
	suite.Suite
}

// Run the unit test suite
func TestPingUnitTestSuite(t *testing.T) {
	suite.Run(t, new(PingUnitTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test marshalling an example Ping message from documentation into the same payload.
func (suite *PingUnitTestSuite) TestPingMarshalJson() {
	// Payload to unmarshal
	payload := `{
		"event": "ping",
		"reqid": 42
	}`
	// Remove whitespaces
	payload = matchesWhitespacesRegex.ReplaceAllString(payload, "")
	// Unmarshal payload into target struct
	target := new(Ping)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Marshal target
	actual, err := json.Marshal(target)
	require.NoError(suite.T(), err)
	// Compare
	require.Equal(suite.T(), payload, string(actual))
}
