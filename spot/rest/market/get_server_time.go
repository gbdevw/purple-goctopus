package market

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// GetServerTime result
type GetServerTimeResult struct {
	// Unix timestamp
	Unixtime int64 `json:"unixtime"`
	// RFC 1123 time format
	Rfc1123 string `json:"rfc1123"`
}

// GetServerTime reponse
type GetServerTimeResponse struct {
	common.KrakenSpotRESTResponse
	Result *GetServerTimeResult `json:"result,omitempty"`
}
