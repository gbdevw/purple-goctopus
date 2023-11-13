package market

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// Reponse for GetServerTime
type GetServerTimeResponse struct {
	common.KrakenSpotRESTResponse
	Result struct {
		// Unix timestamp
		Unixtime int64 `json:"unixtime"`
		// RFC 1123 time format
		Rfc1123 string `json:"rfc1123"`
	} `json:"result,omitempty"`
}
