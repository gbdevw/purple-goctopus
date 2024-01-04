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

// Unit test suite for CancelAllOrdersAfterX
type CancelAllOrdersAfterXUnitTestSuite struct {
	suite.Suite
}

// Run the unit test suite
func TestCancelAllOrdersAfterXUnitTestSuite(t *testing.T) {
	suite.Run(t, new(CancelAllOrdersAfterXUnitTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test marshalling an example CancelAllOrdersAfterXRequest message to the same payload as documentation
func (suite *CancelAllOrdersAfterXUnitTestSuite) TestCancelAllOrdersAfterXRequestMarshalJson() {
	// Payload to marshal
	payload := `{
		"event": "cancelAllOrdersAfter",
		"token": "0000000000000000000000000000000000000000",
		"reqid": 1608543428050,
		"timeout": 60
	}`
	// Remove whitespaces from payload
	payload = matchesWhitespacesRegex.ReplaceAllString(payload, "")
	// Unmarshal to target
	target := new(CancelAllOrdersAfterXRequest)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Marshal target
	actual, err := json.Marshal(target)
	require.NoError(suite.T(), err)
	// Check data
	require.Equal(suite.T(), payload, string(actual))
}

// Test unmarshalling an example of a successfull CancelAllOrdersAfterXResponse and then test marshalling it to get the same
// payload as the API.
func (suite *CancelAllOrdersAfterXUnitTestSuite) TestCancelAllOrdersAfterXResponseMarshalJson() {
	// Payload to marshal
	payload := `{
		"event": "cancelAllOrdersAfterStatus",
		"reqid": 1608543428051,
		"status": "ok",
		"currentTime": "2020-12-21T09:37:09Z",
		"triggerTime": "0"
	}`
	// Remove whitespaces from payload
	payload = matchesWhitespacesRegex.ReplaceAllString(payload, "")
	// Unmarshal to target
	target := new(CancelAllOrdersAfterXResponse)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Marshal target
	actual, err := json.Marshal(target)
	require.NoError(suite.T(), err)
	// Check data
	require.Equal(suite.T(), payload, string(actual))
}
