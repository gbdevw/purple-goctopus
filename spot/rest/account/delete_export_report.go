package account

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// Enum for report deletion type
type ReportDeletionEnum string

// Values for ReportDeletionEnum
const (
	DeleteReport ReportDeletionEnum = "delete"
	CancelReport ReportDeletionEnum = "cancel"
)

// DeleteExportReport request parameters.
type DeleteExportReportRequestParameters struct {
	// Report ID to delete or cancel.
	Id string `json:"id"`
	// Type of deletion. 'delete' can only be used for reports that have already been processed.
	// Use 'cancel' for queued or processing reports.
	//
	// Cf. ReportDeletionEnum for values.
	Type string `json:"type"`
}

// DeleteExportReport result
type DeleteExportReportResult struct {
	// Whether deletion was successful
	Delete bool `json:"delete"`
	// Whether cancellation was successful
	Cancel bool `json:"cancel"`
}

// DeleteExportReport response
type DeleteExportReportResponse struct {
	common.KrakenSpotRESTResponse
	Result *DeleteExportReportResult `json:"result,omitempty"`
}
