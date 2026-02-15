package services

import (
	"bytes"
	"fmt"
	"time"

	"insight-engine-backend/models"

	"github.com/xuri/excelize/v2"
)

type ReportingService struct{}

func NewReportingService() *ReportingService {
	return &ReportingService{}
}

func (s *ReportingService) GenerateExcelReport(req models.ReportRequest) (*bytes.Buffer, error) {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	// Create a new sheet.
	index, err := f.NewSheet("Report")
	if err != nil {
		return nil, err
	}

	// Set value of a cell.
	// Title
	f.SetCellValue("Report", "A1", req.Title)
	f.SetCellValue("Report", "A2", fmt.Sprintf("Generated at: %s", time.Now().Format(time.RFC1123)))

	// Headers
	// Start from Row 4
	for i, header := range req.Headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 4)
		f.SetCellValue("Report", cell, header)
	}

	// Data
	// Start from Row 5
	for i, row := range req.Data {
		rowNum := i + 5
		for j, header := range req.Headers {
			cell, _ := excelize.CoordinatesToCellName(j+1, rowNum)
			val := row[header]
			f.SetCellValue("Report", cell, val)
		}
	}

	// Set active sheet of the workbook.
	f.SetActiveSheet(index)

	// Save to buffer
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	return buf, nil
}
