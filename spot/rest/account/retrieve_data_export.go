package account

// RetrieveDataExportParameters contains Retrieve Data Export required parameters.
type RetrieveDataExportParameters struct {
	// Report ID to retrieve
	Id string
}

// RetrieveDataExportResponse contains Retrieve Data Export response data.
type RetrieveDataExportResponse struct {
	// Binary zip archive containing the report
	Report []byte
}
