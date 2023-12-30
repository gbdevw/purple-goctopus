package account

import (
	"github.com/gbdevw/purple-goctopus/spot/rest/common"
)

// RequestExportReport request parameters.
type RequestExportReportRequestParameters struct {
	// Type of data to export. Cf ReportTypeEnum for values.
	Report string `json:"report"`
	// Description for the export
	Description string `json:"description"`
}

// RequestExportReport request options.
type RequestExportReportRequestOptions struct {
	// File format to export. Cf. ReportFormatEnum for values.
	//
	// Defaults to "CSV". An empty string triggers the default behavior.
	Format string `json:"format,omitempty"`
	// List of fields to include. Cf ReportFieldsEnum for values and usage.
	//
	// Defaults to all. An empty value triggers the default behavior.
	Fields []string `json:"fields,omitempty"`
	// UNIX timestamp for report start time.
	//
	// Default 1st of the current month. A zero value triggers default behavior.
	StartTm int64 `json:"starttm,omitempty"`
	// UNIX timestamp for report end time.
	//
	// Defaults to now. A zero value triggers default behavior.
	EndTm int64 `json:"endtm,omitempty"`
}

// RequestExportReport Result
type RequestExportReportResult struct {
	// Request ID
	Id string `json:"id"`
}

// RequestExportReport response.
type RequestExportReportResponse struct {
	common.KrakenSpotRESTResponse
	Result *RequestExportReportResult `json:"result,omitempty"`
}
