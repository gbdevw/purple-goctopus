package market

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// Enum for system status
type SystemStatus string

// Values for SystemStatus
const (
	Online      SystemStatus = "online"
	Maintenance SystemStatus = "maintenance"
	CancelOnly  SystemStatus = "cancel_only"
	PostOnly    SystemStatus = "post_only"
)

// GetSystemStatus result
type GetSystemStatusResult struct {
	// System status
	Status string `json:"status"`
	// Current timestamp (RFC3339)
	Timestamp string `json:"timestamp"`
}

// GetSystemStatus response
type GetSystemStatusResponse struct {
	common.KrakenSpotRESTResponse
	Result *GetSystemStatusResult `json:"result,omitempty"`
}
