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

// Unit test suite for OwnTrades
type OwnTradesUnitTestSuite struct {
	suite.Suite
}

// Run the unit test suite
func TestOwnTradesUnitTestSuite(t *testing.T) {
	suite.Run(t, new(OwnTradesUnitTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test unmarshalling an example OwnTrades message from documentation into the corresponding struct.
func (suite *OwnTradesUnitTestSuite) TestOwnTradesUnmarshalJson() {
	// Payload to unmarshal
	payload := `[
		[
		  {
			"TDLH43-DVQXD-2KHVYY": {
			  "ordertxid": "TDLH43-DVQXD-2KHVYY",
			  "postxid": "OGTT3Y-C6I3P-XRI6HX",
			  "pair": "XBT/EUR",
			  "time": "1560516023.070651",
			  "type": "sell",
			  "ordertype": "limit",
			  "price": "100000.00000",
			  "cost": "1000000.00000",
			  "fee": "1600.00000",
			  "vol": "1000000000.00000000",
			  "margin": "0.00000"
			}
		  },
		  {
			"TDLH43-DVQXD-2KHVYY": {
				"ordertxid": "TDLH43-DVQXD-2KHVYY",
				"postxid": "OGTT3Y-C6I3P-XRI6HX",
				"pair": "XBT/EUR",
				"time": "1560516023.070651",
				"type": "sell",
				"ordertype": "limit",
				"price": "100000.00000",
				"cost": "1000000.00000",
				"fee": "1600.00000",
				"vol": "1000000000.00000000",
				"margin": "0.00000"
			}
		  },
		  {
			"TDLH43-DVQXD-2KHVYY": {
				"ordertxid": "TDLH43-DVQXD-2KHVYY",
				"postxid": "OGTT3Y-C6I3P-XRI6HX",
				"pair": "XBT/EUR",
				"time": "1560516023.070651",
				"type": "sell",
				"ordertype": "limit",
				"price": "100000.00000",
				"cost": "1000000.00000",
				"fee": "1600.00000",
				"vol": "1000000000.00000000",
				"margin": "0.00000"
			}
		  },
		  {
			"TDLH43-DVQXD-2KHVYY": {
				"ordertxid": "TDLH43-DVQXD-2KHVYY",
				"postxid": "OGTT3Y-C6I3P-XRI6HX",
				"pair": "XBT/EUR",
				"time": "1560516023.070651",
				"type": "sell",
				"ordertype": "limit",
				"price": "100000.00000",
				"cost": "1000000.00000",
				"fee": "1600.00000",
				"vol": "1000000000.00000000",
				"margin": "0.00000"
			}
		  }
		],
		"ownTrades",
		{
		  "sequence": 2948
		}
	]`
	// Expectations
	expectedChannelName := string(ChannelOwnTrades)
	expectedSeqId := int64(2948)
	expectedCount := 4
	expectedTradeId := "TDLH43-DVQXD-2KHVYY"
	expectedVolume := "1000000000.00000000"
	// Unmarshal payload into target struct
	target := new(OwnTrades)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Check data
	require.Equal(suite.T(), expectedChannelName, target.ChannelName)
	require.Equal(suite.T(), expectedSeqId, target.SequenceId.Sequence)
	require.Len(suite.T(), target.Data, expectedCount)
	require.Equal(suite.T(), expectedVolume, target.Data[0][expectedTradeId].Volume)
}

// Test marshalling an example OwnTrades message to the same paylaod as documentation.
func (suite *OwnTradesUnitTestSuite) TestOwnTradesMarshalJson() {
	// Payload to marshal
	payload := `[
		[
		  {
			"TDLH43-DVQXD-2KHVYY": {
			  "ordertxid": "TDLH43-DVQXD-2KHVYY",
			  "postxid": "OGTT3Y-C6I3P-XRI6HX",
			  "pair": "XBT/EUR",
			  "time": "1560516023.070651",
			  "type": "sell",
			  "ordertype": "limit",
			  "price": "100000.00000",
			  "cost": "1000000.00000",
			  "fee": "1600.00000",
			  "vol": "1000000000.00000000",
			  "margin": "0.00000"
			}
		  },
		  {
			"TDLH43-DVQXD-2KHVYY": {
				"ordertxid": "TDLH43-DVQXD-2KHVYY",
				"postxid": "OGTT3Y-C6I3P-XRI6HX",
				"pair": "XBT/EUR",
				"time": "1560516023.070651",
				"type": "sell",
				"ordertype": "limit",
				"price": "100000.00000",
				"cost": "1000000.00000",
				"fee": "1600.00000",
				"vol": "1000000000.00000000",
				"margin": "0.00000"
			}
		  },
		  {
			"TDLH43-DVQXD-2KHVYY": {
				"ordertxid": "TDLH43-DVQXD-2KHVYY",
				"postxid": "OGTT3Y-C6I3P-XRI6HX",
				"pair": "XBT/EUR",
				"time": "1560516023.070651",
				"type": "sell",
				"ordertype": "limit",
				"price": "100000.00000",
				"cost": "1000000.00000",
				"fee": "1600.00000",
				"vol": "1000000000.00000000",
				"margin": "0.00000"
			}
		  },
		  {
			"TDLH43-DVQXD-2KHVYY": {
				"ordertxid": "TDLH43-DVQXD-2KHVYY",
				"postxid": "OGTT3Y-C6I3P-XRI6HX",
				"pair": "XBT/EUR",
				"time": "1560516023.070651",
				"type": "sell",
				"ordertype": "limit",
				"price": "100000.00000",
				"cost": "1000000.00000",
				"fee": "1600.00000",
				"vol": "1000000000.00000000",
				"margin": "0.00000"
			}
		  }
		],
		"ownTrades",
		{
		  "sequence": 2948
		}
	]`
	// Remove whitespace
	payload = matchesWhitespacesRegex.ReplaceAllString(payload, "")
	// Unmarshal payload into target struct
	target := new(OwnTrades)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Marshal target
	actual, err := json.Marshal(target)
	require.NoError(suite.T(), err)
	// Check data
	require.Equal(suite.T(), payload, string(actual))
}
