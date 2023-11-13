package account

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// Enum for report deletion type
type ReportDeletion string

// Values for ReportDeletion
const (
	DeleteReport ReportDeletion = "delete"
	CancelReport ReportDeletion = "cancel"
)

// DeleteExportReportParameters contains Delete Data Export required parameters.
type DeleteExportReportParameters struct {
	// Report ID to delete or cancel
	Id string
	// Type of deletion.
	// delete can only be used for reports that have already been processed. Use cancel for queued or processing reports.
	// Values: "delete" "cancel"
	Type string
}

// DeleteExportReport Result
type DeleteExportReportResult struct {
	// Whether deletion was successful
	Delete bool `json:"delete,omitempty"`
	// Whether cancellation was successful
	Cancel bool `json:"cancel,omitempty"`
}

// Delete Export Report Response
type DeleteExportReportResponse struct {
	common.KrakenSpotRESTResponse
	Result *DeleteExportReportResult `json:"result,omitempty"`
}
