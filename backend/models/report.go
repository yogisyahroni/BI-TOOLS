package models

type ReportRequest struct {
	Title   string                   `json:"title"`
	Headers []string                 `json:"headers"` // Ordered list of keys
	Data    []map[string]interface{} `json:"data"`
}

type ReportConfig struct {
	SheetName string
}
