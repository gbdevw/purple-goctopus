package account

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

/*************************************************************************************************/
/* TEST SUITE                                                                                    */
/*************************************************************************************************/

// Unit test suite for GetTradeVolume DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type GetTradeVolumeTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestGetTradeVolumeTestSuite(t *testing.T) {
	suite.Run(t, new(GetTradeVolumeTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of GetTradeVolume.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding GetTradeVolumeResponse struct.
func (suite *GetTradeVolumeTestSuite) TestGetTradeVolumeUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "currency": "ZUSD",
		  "volume": "200709587.4223",
		  "fees": {
			"XXBTZUSD": {
			  "fee": "0.1000",
			  "minfee": "0.1000",
			  "maxfee": "0.2600",
			  "nextfee": null,
			  "nextvolume": null,
			  "tiervolume": "10000000.0000"
			}
		  },
		  "fees_maker": {
			"XXBTZUSD": {
			  "fee": "0.0000",
			  "minfee": "0.0000",
			  "maxfee": "0.1600",
			  "nextfee": null,
			  "nextvolume": null,
			  "tiervolume": "10000000.0000"
			}
		  }
		}
	}`
	expectedCurrency := "ZUSD"
	expectedVolume := "200709587.4223"
	expectedTargetPair := "XXBTZUSD"
	expectedFeesBTCFee := "0.1000"
	expectedFeesBTCMinFee := "0.1000"
	expectedFeesBTCMaxFee := "0.2600"
	expectedFeesBTCNextFee := ""    // Null will be mapped to an empty string
	expectedFeesBTCNextVolume := "" // Null will be mapped to an empty string
	expectedFeesBTCTierVolume := "10000000.0000"
	expectedFeesMakerBTCFee := "0.0000"
	expectedFeesMakerBTCMinFee := "0.0000"
	expectedFeesMakerBTCMaxFee := "0.1600"
	expectedFeesMakerBTCNextFee := ""    // Null will be mapped to an empty string
	expectedFeesMakerBTCNextVolume := "" // Null will be mapped to an empty string
	expectedFeesMakerBTCTierVolume := "10000000.0000"
	// Unmarshal payload into struct
	response := new(GetTradeVolumeResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Equal(suite.T(), expectedCurrency, response.Result.Currency)
	require.Equal(suite.T(), expectedVolume, response.Result.Volume.String())
	require.Equal(suite.T(), expectedFeesBTCFee, response.Result.Fees[expectedTargetPair].Fee.String())
	require.Equal(suite.T(), expectedFeesBTCMinFee, response.Result.Fees[expectedTargetPair].MinimumFee.String())
	require.Equal(suite.T(), expectedFeesBTCMaxFee, response.Result.Fees[expectedTargetPair].MaximumFee.String())
	require.Equal(suite.T(), expectedFeesBTCNextFee, response.Result.Fees[expectedTargetPair].NextFee.String())
	require.Equal(suite.T(), expectedFeesBTCNextVolume, response.Result.Fees[expectedTargetPair].NextTierVolume.String())
	require.Equal(suite.T(), expectedFeesBTCTierVolume, response.Result.Fees[expectedTargetPair].TierVolume.String())
	require.Equal(suite.T(), expectedFeesMakerBTCFee, response.Result.FeesMaker[expectedTargetPair].Fee.String())
	require.Equal(suite.T(), expectedFeesMakerBTCMinFee, response.Result.FeesMaker[expectedTargetPair].MinimumFee.String())
	require.Equal(suite.T(), expectedFeesMakerBTCMaxFee, response.Result.FeesMaker[expectedTargetPair].MaximumFee.String())
	require.Equal(suite.T(), expectedFeesMakerBTCNextFee, response.Result.FeesMaker[expectedTargetPair].NextFee.String())
	require.Equal(suite.T(), expectedFeesMakerBTCNextVolume, response.Result.FeesMaker[expectedTargetPair].NextTierVolume.String())
	require.Equal(suite.T(), expectedFeesMakerBTCTierVolume, response.Result.FeesMaker[expectedTargetPair].TierVolume.String())
}
