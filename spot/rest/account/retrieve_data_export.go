package account

import "io"

// RetrieveDataExport request parameters.
type RetrieveDataExportParameters struct {
	// Report ID to retrieve
	Id string `json:"id"`
}

// RetrieveDataExport response.
type RetrieveDataExportResponse struct {
	// Reader (tied to http.Response body) which can be used to download the zip archive which contains data.
	Report io.Reader
}
