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

// Test unmarshalling an example Ping message from documentation into the corresponding struct.
func (suite *PingUnitTestSuite) TestPingUnmarshalJson() {
	// Payload to unmarshal
	payload := `{
		"event": "ping",
		"reqid": 42
	}`
	// Expectations
	expectedEvent := string(EventTypePing)
	expectedReqId := 42
	// Unmarshal payload into target struct
	target := new(Ping)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Check data
	require.Equal(suite.T(), expectedEvent, target.Event)
	require.Equal(suite.T(), expectedReqId, target.ReqId)
}
