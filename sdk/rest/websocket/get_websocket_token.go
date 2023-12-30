package websocket

import (
	"github.com/gbdevw/purple-goctopus/spot/rest/common"
)

// Result for GetWebsocketToken
type GetWebsocketTokenResult struct {
	// Websockets token.
	Token string `json:"token"`
	// Time (in seconds) after which the token expires.
	Expires int64 `json:"expires"`
}

// Response for GetWebsocketToken
type GetWebsocketTokenResponse struct {
	common.KrakenSpotRESTResponse
	Result *GetWebsocketTokenResult `json:"result,omitempty"`
}
