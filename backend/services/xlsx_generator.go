package services

import (
	"archive/zip"
	"bytes"
	"fmt"
	"strings"
	"time"
)

// XLSXGenerator builds OOXML .xlsx files without external dependencies.
// An XLSX is a ZIP containing XML parts conforming to the SpreadsheetML spec.
type XLSXGenerator struct{}

// NewXLSXGenerator creates a new XLSXGenerator instance
func NewXLSXGenerator() *XLSXGenerator {
	return &XLSXGenerator{}
}

// XLSXSheet represents a single worksheet with tabular data
type XLSXSheet struct {
	Name    string     // Tab name (max 31 chars, no special chars)
	Headers []string   // Column headers
	Rows    [][]string // Data rows
}

// GenerateXLSX builds a complete .xlsx file from one or more sheets and returns the bytes.
func (g *XLSXGenerator) GenerateXLSX(sheets []XLSXSheet, title string) ([]byte, error) {
	if len(sheets) == 0 {
		return nil, fmt.Errorf("at least one sheet is required")
	}

	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	// ---- [Content_Types].xml ----
	if err := g.writeContentTypes(zw, len(sheets)); err != nil {
		return nil, fmt.Errorf("content types: %w", err)
	}

	// ---- _rels/.rels ----
	if err := g.writeRootRels(zw); err != nil {
		return nil, fmt.Errorf("root rels: %w", err)
	}

	// ---- xl/_rels/workbook.xml.rels ----
	if err := g.writeWorkbookRels(zw, len(sheets)); err != nil {
		return nil, fmt.Errorf("workbook rels: %w", err)
	}

	// ---- xl/workbook.xml ----
	if err := g.writeWorkbook(zw, sheets); err != nil {
		return nil, fmt.Errorf("workbook: %w", err)
	}

	// ---- xl/styles.xml ----
	if err := g.writeStyles(zw); err != nil {
		return nil, fmt.Errorf("styles: %w", err)
	}

	// ---- xl/sharedStrings.xml ----
	sst, sstIndex := g.buildSharedStrings(sheets)
	if err := g.writeSharedStrings(zw, sst); err != nil {
		return nil, fmt.Errorf("shared strings: %w", err)
	}

	// ---- xl/worksheets/sheet{N}.xml ----
	for i, sheet := range sheets {
		if err := g.writeSheet(zw, i+1, sheet, sstIndex); err != nil {
			return nil, fmt.Errorf("sheet %d: %w", i+1, err)
		}
	}

	// ---- docProps/core.xml ----
	if err := g.writeCoreProps(zw, title); err != nil {
		return nil, fmt.Errorf("core props: %w", err)
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("close zip: %w", err)
	}

	return buf.Bytes(), nil
}

// writeZipEntry writes a single entry into the zip archive
func (g *XLSXGenerator) writeZipEntry(zw *zip.Writer, path string, content string) error {
	w, err := zw.Create(path)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(content))
	return err
}

// writeContentTypes writes [Content_Types].xml
func (g *XLSXGenerator) writeContentTypes(zw *zip.Writer, sheetCount int) error {
	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
  <Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>
  <Override PartName="/xl/sharedStrings.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sharedStrings+xml"/>
  <Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/>`)

	for i := 1; i <= sheetCount; i++ {
		sb.WriteString(fmt.Sprintf(`
  <Override PartName="/xl/worksheets/sheet%d.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>`, i))
	}

	sb.WriteString(`
</Types>`)
	return g.writeZipEntry(zw, "[Content_Types].xml", sb.String())
}

// writeRootRels writes _rels/.rels
func (g *XLSXGenerator) writeRootRels(zw *zip.Writer) error {
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/>
</Relationships>`
	return g.writeZipEntry(zw, "_rels/.rels", content)
}

// writeWorkbookRels writes xl/_rels/workbook.xml.rels
func (g *XLSXGenerator) writeWorkbookRels(zw *zip.Writer, sheetCount int) error {
	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)

	for i := 1; i <= sheetCount; i++ {
		sb.WriteString(fmt.Sprintf(`
  <Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet%d.xml"/>`, i, i))
	}

	// Shared strings and styles after sheets
	ssID := sheetCount + 1
	stID := sheetCount + 2
	sb.WriteString(fmt.Sprintf(`
  <Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/sharedStrings" Target="sharedStrings.xml"/>
  <Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>`, ssID, stID))

	sb.WriteString(`
</Relationships>`)
	return g.writeZipEntry(zw, "xl/_rels/workbook.xml.rels", sb.String())
}

// writeWorkbook writes xl/workbook.xml
func (g *XLSXGenerator) writeWorkbook(zw *zip.Writer, sheets []XLSXSheet) error {
	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets>`)

	for i, sheet := range sheets {
		name := sanitizeSheetName(sheet.Name)
		sb.WriteString(fmt.Sprintf(`
    <sheet name="%s" sheetId="%d" r:id="rId%d"/>`, xmlEscapeXLSX(name), i+1, i+1))
	}

	sb.WriteString(`
  </sheets>
</workbook>`)
	return g.writeZipEntry(zw, "xl/workbook.xml", sb.String())
}

// writeStyles writes xl/styles.xml with header styling
func (g *XLSXGenerator) writeStyles(zw *zip.Writer) error {
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <fonts count="2">
    <font>
      <sz val="11"/>
      <name val="Inter"/>
      <color rgb="FF333333"/>
    </font>
    <font>
      <b/>
      <sz val="11"/>
      <name val="Inter"/>
      <color rgb="FFFFFFFF"/>
    </font>
  </fonts>
  <fills count="3">
    <fill><patternFill patternType="none"/></fill>
    <fill><patternFill patternType="gray125"/></fill>
    <fill>
      <patternFill patternType="solid">
        <fgColor rgb="FF6366F1"/>
        <bgColor indexed="64"/>
      </patternFill>
    </fill>
  </fills>
  <borders count="1">
    <border>
      <left/><right/><top/><bottom/><diagonal/>
    </border>
  </borders>
  <cellStyleXfs count="1">
    <xf numFmtId="0" fontId="0" fillId="0" borderId="0"/>
  </cellStyleXfs>
  <cellXfs count="2">
    <xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/>
    <xf numFmtId="0" fontId="1" fillId="2" borderId="0" xfId="0" applyFont="1" applyFill="1" applyAlignment="1">
      <alignment horizontal="center" vertical="center"/>
    </xf>
  </cellXfs>
  <cellStyles count="1">
    <cellStyle name="Normal" xfId="0" builtinId="0"/>
  </cellStyles>
</styleSheet>`
	return g.writeZipEntry(zw, "xl/styles.xml", content)
}

// buildSharedStrings collects all unique strings across all sheets
func (g *XLSXGenerator) buildSharedStrings(sheets []XLSXSheet) ([]string, map[string]int) {
	index := make(map[string]int)
	var sst []string

	addString := func(s string) {
		if _, exists := index[s]; !exists {
			index[s] = len(sst)
			sst = append(sst, s)
		}
	}

	for _, sheet := range sheets {
		for _, h := range sheet.Headers {
			addString(h)
		}
		for _, row := range sheet.Rows {
			for _, cell := range row {
				addString(cell)
			}
		}
	}

	return sst, index
}

// writeSharedStrings writes xl/sharedStrings.xml
func (g *XLSXGenerator) writeSharedStrings(zw *zip.Writer, sst []string) error {
	var sb strings.Builder
	count := len(sst)
	sb.WriteString(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" count="%d" uniqueCount="%d">`, count, count))

	for _, s := range sst {
		sb.WriteString(fmt.Sprintf(`
  <si><t>%s</t></si>`, xmlEscapeXLSX(s)))
	}

	sb.WriteString(`
</sst>`)
	return g.writeZipEntry(zw, "xl/sharedStrings.xml", sb.String())
}

// writeSheet writes xl/worksheets/sheet{N}.xml
func (g *XLSXGenerator) writeSheet(zw *zip.Writer, sheetNum int, sheet XLSXSheet, sstIndex map[string]int) error {
	var sb strings.Builder

	colCount := len(sheet.Headers)
	if colCount == 0 {
		colCount = 1
	}
	lastCol := colName(colCount - 1)
	totalRows := 1 + len(sheet.Rows)

	sb.WriteString(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <dimension ref="A1:%s%d"/>
  <sheetViews>
    <sheetView tabSelected="%d" workbookViewId="0">
      <pane ySplit="1" topLeftCell="A2" activePane="bottomLeft" state="frozen"/>
    </sheetView>
  </sheetViews>
  <cols>`, lastCol, totalRows, boolToInt(sheetNum == 1)))

	// Auto-width columns (approximate: 15 chars)
	for i := 0; i < colCount; i++ {
		width := 15.0
		if i < len(sheet.Headers) && len(sheet.Headers[i]) > 15 {
			width = float64(len(sheet.Headers[i])) * 1.2
		}
		if width > 50 {
			width = 50
		}
		sb.WriteString(fmt.Sprintf(`
    <col min="%d" max="%d" width="%.1f" bestFit="1" customWidth="1"/>`, i+1, i+1, width))
	}

	sb.WriteString(`
  </cols>
  <sheetData>`)

	// Header row (style index 1 = bold white on indigo)
	sb.WriteString(`
    <row r="1" spans="1:` + fmt.Sprintf("%d", colCount) + `">`)
	for ci, hdr := range sheet.Headers {
		ref := colName(ci) + "1"
		idx := sstIndex[hdr]
		sb.WriteString(fmt.Sprintf(`
      <c r="%s" t="s" s="1"><v>%d</v></c>`, ref, idx))
	}
	sb.WriteString(`
    </row>`)

	// Data rows
	for ri, row := range sheet.Rows {
		rowNum := ri + 2
		sb.WriteString(fmt.Sprintf(`
    <row r="%d" spans="1:%d">`, rowNum, colCount))

		for ci := 0; ci < colCount; ci++ {
			ref := colName(ci) + fmt.Sprintf("%d", rowNum)
			if ci < len(row) {
				cellVal := row[ci]
				idx, ok := sstIndex[cellVal]
				if ok {
					sb.WriteString(fmt.Sprintf(`
      <c r="%s" t="s"><v>%d</v></c>`, ref, idx))
				} else {
					sb.WriteString(fmt.Sprintf(`
      <c r="%s"><v>0</v></c>`, ref))
				}
			}
		}

		sb.WriteString(`
    </row>`)
	}

	sb.WriteString(`
  </sheetData>
  <autoFilter ref="A1:` + lastCol + fmt.Sprintf("%d", totalRows) + `"/>
</worksheet>`)

	path := fmt.Sprintf("xl/worksheets/sheet%d.xml", sheetNum)
	return g.writeZipEntry(zw, path, sb.String())
}

// writeCoreProps writes docProps/core.xml
func (g *XLSXGenerator) writeCoreProps(zw *zip.Writer, title string) error {
	now := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <dc:title>%s</dc:title>
  <dc:creator>Insight Engine</dc:creator>
  <dcterms:created xsi:type="dcterms:W3CDTF">%s</dcterms:created>
  <dcterms:modified xsi:type="dcterms:W3CDTF">%s</dcterms:modified>
</cp:coreProperties>`, xmlEscapeXLSX(title), now, now)
	return g.writeZipEntry(zw, "docProps/core.xml", content)
}

// ---- Helpers ----

// colName converts a 0-based column index to an Excel column name (A, B, ..., Z, AA, AB, ...)
func colName(index int) string {
	name := ""
	for {
		name = string(rune('A'+index%26)) + name
		index = index/26 - 1
		if index < 0 {
			break
		}
	}
	return name
}

// sanitizeSheetName ensures the sheet name conforms to Excel rules
func sanitizeSheetName(name string) string {
	if name == "" {
		return "Sheet1"
	}
	// Remove invalid characters
	invalid := []string{"\\", "/", "?", "*", "[", "]", ":"}
	cleanName := name
	for _, ch := range invalid {
		cleanName = strings.ReplaceAll(cleanName, ch, "_")
	}
	// Truncate to 31 characters
	if len(cleanName) > 31 {
		cleanName = cleanName[:31]
	}
	return cleanName
}

// xmlEscapeXLSX escapes special XML characters
func xmlEscapeXLSX(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// boolToInt converts a boolean to 1 or 0
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
