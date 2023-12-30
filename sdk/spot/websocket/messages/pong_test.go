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

// Test unmarshalling an example Pong message from documentation into the corresponding struct.
func (suite *PongUnitTestSuite) TestPongUnmarshalJson() {
	// Payload to unmarshal
	payload := `{
		"event": "pong",
		"reqid": 42
	}`
	// Expectations
	expectedEvent := string(EventTypePong)
	expectedReqId := 42
	// Unmarshal payload into target struct
	target := new(Pong)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Check data
	require.Equal(suite.T(), expectedEvent, target.Event)
	require.Equal(suite.T(), expectedReqId, target.ReqId)
}
