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

// Unit test suite for SystemStatus
type SystemStatusUnitTestSuite struct {
	suite.Suite
}

// Run the unit test suite
func TestSystemStatusUnitTestSuite(t *testing.T) {
	suite.Run(t, new(SystemStatusUnitTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test unmarshalling an example SystemStatus message from documentation into the corresponding struct.
func (suite *SystemStatusUnitTestSuite) TestSystemStatusUnmarshalJson() {
	// Payload to unmarshal
	payload := `{
		"connectionID": 8628615390848610000,
		"event": "systemStatus",
		"status": "online",
		"version": "1.0.0"
	}`
	// Expectations
	expectedEvent := string(EventTypeSystemStatus)
	expectedConnectionId := int64(8628615390848610000)
	expectedStatus := string(StatusOnline)
	expectedVersion := "1.0.0"
	// Unmarshal payload into target struct
	target := new(SystemStatus)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Check data
	require.Equal(suite.T(), expectedEvent, target.Event)
	require.Equal(suite.T(), expectedConnectionId, target.ConnectionId)
	require.Equal(suite.T(), expectedStatus, target.Status)
	require.Equal(suite.T(), expectedVersion, target.Version)
}
