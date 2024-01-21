package market

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

/*************************************************************************************************/
/* TEST SUITE                                                                                    */
/*************************************************************************************************/

// Unit test suite for GetOHLCData DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type GetOHLCDataTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestGetOHLCDataTestSuite(t *testing.T) {
	suite.Run(t, new(GetOHLCDataTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of GetOHLCDataResponse.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding GetOHLCDataResponse struct.
func (suite *GetOHLCDataTestSuite) TestGetOHLCDataResponseUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "XXBTZUSD": [
			[
			  1688671200,
			  "30306.14242424242",
			  "30306.2",
			  "30305.7",
			  "30305.7",
			  "30306.1",
			  "3.39243896",
			  23
			],
			[
			  1688671260,
			  "30304.5",
			  "30304.5",
			  "30300.0",
			  "30300.0",
			  "30300.0",
			  "4.42996871",
			  18
			],
			[
			  1688671320,
			  "30300.3",
			  "30300.4",
			  "30291.4",
			  "30291.4",
			  "30294.7",
			  "2.13024789",
			  25
			],
			[
			  1688671380,
			  "30291.8",
			  "30295.1",
			  "30291.8",
			  "30295.0",
			  "30293.8",
			  "1.01836275",
			  9
			]
		  ],
		  "last": 1688672160
		}
	}`
	expectedResultCount := 4
	expectedLast := int64(1688672160)
	expectedItem1Open := "30306.14242424242"    // Ensure all decimals from source are OK
	expectedItem1Timestamp := int64(1688671200) // Ensure parsong does not loose precision
	expectedItem1Count := int64(23)
	// Unmarshal payload into struct
	response := new(GetOHLCDataResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Len(suite.T(), response.Result.Data, expectedResultCount)
	require.Equal(suite.T(), expectedLast, response.Result.Last)
	require.Equal(suite.T(), expectedItem1Open, response.Result.Data[0].Open)
	require.Equal(suite.T(), expectedItem1Timestamp, response.Result.Data[0].Timestamp)
	require.Equal(suite.T(), expectedItem1Count, response.Result.Data[0].TradesCount)
}

// Test the JSON marshaller of GetOHLCDataResponse.
//
// The test will ensure:
//   - The same JSON payload as the API response is generated when marshalling GetOHLCDataResponse.
func (suite *GetOHLCDataTestSuite) TestGetOHLCDataMarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "XXBTZUSD": [
			[
			  1688671200,
			  "30306.14242424242",
			  "30306.2",
			  "30305.7",
			  "30305.7",
			  "30306.1",
			  "3.39243896",
			  23
			],
			[
			  1688671260,
			  "30304.5",
			  "30304.5",
			  "30300.0",
			  "30300.0",
			  "30300.0",
			  "4.42996871",
			  18
			],
			[
			  1688671320,
			  "30300.3",
			  "30300.4",
			  "30291.4",
			  "30291.4",
			  "30294.7",
			  "2.13024789",
			  25
			],
			[
			  1688671380,
			  "30291.8",
			  "30295.1",
			  "30291.8",
			  "30295.0",
			  "30293.8",
			  "1.01836275",
			  9
			]
		  ],
		  "last": 1688672160
		}
	}`
	// Compact payload so we can compare it with the output of Marshal
	target := &bytes.Buffer{}
	err := json.Compact(target, []byte(payload))
	require.NoError(suite.T(), err)
	expectedCompactPayload, err := io.ReadAll(target)
	require.NoError(suite.T(), err)
	// Unmarshal payload into struct
	response := new(GetOHLCDataResponse)
	err = json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	// Marshal and compare
	result, err := json.Marshal(response)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), string(expectedCompactPayload), string(result))
}
