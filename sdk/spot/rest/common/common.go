package common

import (
	"fmt"
	"time"
)

/*************************************************************************************************/
/* COMMON STRUCTS                                                                                */
/*************************************************************************************************/

// Base layout for Kraken Spot REST API responses
type KrakenSpotRESTResponse struct {
	// Errors returned with the response.
	//
	// Please refer to https://support.kraken.com/hc/en-us/articles/360001491786-API-error-messages for details.
	Error []string `json:"error"`
	// Result for the request
	Result interface{} `json:"result,omitempty"`
}

// Container for security options to use during the API call (2FA, ...)
type SecurityOptions struct {
	// Second factor to use to sign request (authenticator app or password). An empty string can be used if 2FA is not enabled.
	//
	// Refer to https://support.kraken.com/hc/en-us/articles/360000714526-How-does-two-factor-authentication-2FA-for-API-keys-work- for details.
	SecondFactor string
}

/*************************************************************************************************/
/* COMMON HELPER FUNCTIONS                                                                       */
/*************************************************************************************************/

// Panic if input err is not nil. Return only V otherwise. The helper function works with all
// helper methods of OHLCDataResult and OHLCData which return an error.
//
// This helper function can be used to replace code blocks like this:
//
// ts, err := ohlc.GetStartTimestamp()
//
//	if err != nil {
//		 panic(err)
//	}
//
// By this:
//
// ts := Must(ohlc.GetStartTimestamp())
func Must[V string | []string | int64 | time.Time | interface{}](v V, err error) V {
	// Panic if error or return input
	if err != nil {
		panic(fmt.Errorf("unexpected error: %w", err))
	}
	return v
}
