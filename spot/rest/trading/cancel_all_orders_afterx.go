package trading

import (
	"time"

	"github.com/gbdevw/purple-goctopus/spot/rest/common"
)

// CancelAllOrdersAfterX request parameters
type CancelCancelAllOrdersAfterXRequestParameters struct {
	// Duration (in seconds) to set/extend the timer by
	Timeout int64 `json:"timeout"`
}

// CancelAllOrdersAfterX result
type CancelAllOrdersAfterXResult struct {
	// Timestamp (RFC3339 format) at which the request was received
	CurrentTime time.Time `json:"currentTime"`
	// Timestamp (RFC3339 format) after which all orders will be cancelled, unless the timer is extended or disabled
	TriggerTime time.Time `json:"triggerTime"`
}

// CancelAllOrdersAfterX response
type CancelAllOrdersAfterXResponse struct {
	common.KrakenSpotRESTResponse
	Result *CancelAllOrdersAfterXResult `json:"result,omitempty"`
}
