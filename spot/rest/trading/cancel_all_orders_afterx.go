package trading

import (
	"time"

	"github.com/gbdevw/purple-goctopus/spot/rest/common"
)

// CancelAllOrdersAfterXParameters
type CancelCancelAllOrdersAfterXParameters struct {
	// Duration (in seconds) to set/extend the timer by
	Timeout int64
}

// Response for Cancel All Orders After X
type CancelAllOrdersAfterXResponse struct {
	common.KrakenSpotRESTResponse
	Result *struct {
		// Timestamp (RFC3339 format) at which the request was received
		CurrentTime time.Time `json:"currentTime"`
		// Timestamp (RFC3339 format) after which all orders will be cancelled, unless the timer is extended or disabled
		TriggerTime time.Time `json:"triggerTime"`
	} `json:"result,omitempty"`
}
