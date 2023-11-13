package account

import (
	"time"

	"github.com/gbdevw/purple-goctopus/spot/rest/common"
)

// Enum for report types
type ReportType string

// Values for ReportType
const (
	Trades  ReportType = "trades"
	Ledgers ReportType = "ledgers"
)

// Enum for report formats
type ReportFormat string

// Values for report formats
const (
	CSV ReportFormat = "CSV"
	TSV ReportFormat = "TSV"
)

// RequestExportReportParameters contains Request Export Report required parameters.
type RequestExportReportParameters struct {
	// Type of data to export
	// Values: "trades", "ledgers"
	Report string
	// Description for the export
	Description string
}

// RequestExportReportOptions contains Request Export Report optional parameters.
type RequestExportReportOptions struct {
	// File format to export.
	// Defaults to "CSV".
	// Values: "CSV" "TSV"
	Format string
	// List of fields to include.
	// Defaults to all
	// Values for trades: ordertxid, time, ordertype, price, cost, fee, vol, margin, misc, ledgers
	// Values for ledgers: refid, time, type, aclass, asset, amount, fee, balance
	Fields []string
	// UNIX timestamp for report start time.
	// Default 1st of the current month
	StartTm *time.Time
	// UNIX timestamp for report end time.
	// Default: now
	EndTm *time.Time
}

// RequestExportReport Result
type RequestExportReportResult struct {
	// Request ID
	Id string `json:"id"`
}

// RequestExportReportResponse contains Request Export Report response data.
type RequestExportReportResponse struct {
	common.KrakenSpotRESTResponse
	Result *RequestExportReportResult `json:"result"`
}
