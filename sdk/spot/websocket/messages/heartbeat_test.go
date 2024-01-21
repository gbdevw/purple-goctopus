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

// Unit test suite for Heartbeat
type HeartbeatUnitTestSuite struct {
	suite.Suite
}

// Run the unit test suite
func TestHeartbeatUnitTestSuite(t *testing.T) {
	suite.Run(t, new(HeartbeatUnitTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test unmarshalling an example Heartbeat message from documentation into the corresponding struct.
func (suite *HeartbeatUnitTestSuite) TestHeartbeatUnmarshalJson() {
	// Payload to unmarshal
	payload := `{
		"event": "heartbeat"
	}`
	// Expectations
	expectedEvent := string(EventTypeHeartbeat)
	// Unmarshal payload into target struct
	target := new(Heartbeat)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Check data
	require.Equal(suite.T(), expectedEvent, target.Event)
}
