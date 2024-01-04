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

// Unit test suite for ErrorMessage
type ErrorMessageUnitTestSuite struct {
	suite.Suite
}

// Run the unit test suite
func TestErrorMessageUnitTestSuite(t *testing.T) {
	suite.Run(t, new(ErrorMessageUnitTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test marshalling an example ErrorMessage message from documentation into the same payload.
func (suite *ErrorMessageUnitTestSuite) TestErrorMessageMarshalJson() {
	// Payload to unmarshal
	payload := `{
		"event": "error",
		"errorMessage":"Exceeded msg rate",
		"reqid": 42
	}`
	// Remove whitespaces
	payload = matchesWhitespacesRegex.ReplaceAllString(payload, "")
	// Unmarshal payload into target struct
	target := new(ErrorMessage)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Marshal target
	actual, err := json.Marshal(target)
	require.NoError(suite.T(), err)
	// Compare
	require.Equal(suite.T(), payload, string(actual))
}
