package account

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// ExportReportStatus contains export report status data.
type ExportReportStatus struct {
	// Report ID
	Id string `json:"id"`
	// Description
	Description string `json:"descr"`
	// Format
	Format string `json:"format"`
	// Report
	Report string `json:"report"`
	// Subtype
	SubType string `json:"subtype"`
	// Status of report. Enum: "Queued" "Processing" "Processed"
	Status string `json:"status"`
	// Fields
	Fields string `json:"fields"`
	// UNIX timestamp of report request
	RequestTimestamp string `json:"createdtm"`
	// UNIX timestamp report processing began
	StartTimestamp string `json:"starttm"`
	// UNIX timestamp report processing finished
	CompletedTimestamp string `json:"completedtm"`
	// UNIX timestamp of the report data start time
	DataStartTimestamp string `json:"datastarttm"`
	// UNIX timestamp of the report data end time
	DataEndTimestamp string `json:"dataendtm"`
	// Asset
	Asset string `json:"asset"`
}

// GetExportReportStatusParameters contains Get Export Report Status required parameters.
type GetExportReportStatusParameters struct {
	// Type of reports to inquire about
	// Values: "trades" "ledgers"
	Report string
}

// GetExportReportStatusResponse contains Get Export Report Status response data.
type GetExportReportStatusResponse struct {
	common.KrakenSpotRESTResponse
	// Export Report Statuses
	Result []ExportReportStatus `json:"result"`
}
