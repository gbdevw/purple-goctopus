package account

import (
	"encoding/json"

	"github.com/gbdevw/purple-goctopus/spot/rest/common"
)

// GetExportReportStatus request parameters.
type GetExportReportStatusRequestParameters struct {
	// Type of reports to inquire about.
	//
	// Cf ReportTypeEnum for values
	Report string
}

// ExportReportStatus data.
type ExportReportStatusItem struct {
	// Report ID
	Id string `json:"id"`
	// Description
	Description string `json:"descr,omitempty"`
	// Format
	Format string `json:"format,omitempty"`
	// Report
	Report string `json:"report,omitempty"`
	// Subtype
	SubType string `json:"subtype,omitempty"`
	// Status of report. Enum: "Queued" "Processing" "Processed"
	Status string `json:"status,omitempty"`
	// Fields
	Fields string `json:"fields,omitempty"`
	// UNIX timestamp (seconds) of report request
	CreatedTimestamp json.Number `json:"createdtm,omitempty"`
	// UNIX timestamp (seconds) report processing began
	StartTimestamp json.Number `json:"starttm,omitempty"`
	// UNIX timestamp report processing finished
	CompletedTimestamp json.Number `json:"completedtm,omitempty"`
	// UNIX timestamp of the report data start time
	DataStartTimestamp json.Number `json:"datastarttm,omitempty"`
	// UNIX timestamp of the report data end time
	DataEndTimestamp json.Number `json:"dataendtm,omitempty"`
	// Asset
	Asset string `json:"asset,omitempty"`
}

// GetExportReportStatus response.
type GetExportReportStatusResponse struct {
	common.KrakenSpotRESTResponse
	Result []*ExportReportStatusItem `json:"result"`
}
