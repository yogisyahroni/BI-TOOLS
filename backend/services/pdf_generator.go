package services

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"time"
)

// PDFGenerator builds production-quality multi-page PDF 1.4 documents.
// Pure Go — no external binaries (chromedp, wkhtmltopdf) required.
// Supports: multi-page, data tables, headers, footers, page numbers,
// watermarks, orientation, page sizes, Helvetica + Helvetica-Bold fonts.
type PDFGenerator struct {
	pageW        float64
	pageH        float64
	marginLeft   float64
	marginRight  float64
	marginTop    float64
	marginBottom float64
	cursorY      float64
	currentPage  int
	totalPages   int

	// Content state
	pages        []string // content stream per page
	currentBuf   *bytes.Buffer
	headerText   string
	footerText   string
	watermark    string
	brandingText string
	timestamp    string

	// Metrics (approximate Helvetica widths)
	charWidths map[byte]float64
}

// PDFSection represents a titled section with tabular data
type PDFSection struct {
	Title   string
	Headers []string
	Rows    [][]string
}

// PDFContent holds all the data to render in the PDF
type PDFContent struct {
	Title     string
	Subtitle  string
	Timestamp string
	Footer    string
	Watermark string
	Metadata  map[string]string
	Sections  []PDFSection
	Branding  string
}

// NewPDFGenerator creates a generator with the given page dimensions (in points).
func NewPDFGenerator(pageW, pageH float64) *PDFGenerator {
	gen := &PDFGenerator{
		pageW:        pageW,
		pageH:        pageH,
		marginLeft:   50,
		marginRight:  50,
		marginTop:    60,
		marginBottom: 60,
		currentPage:  0,
		totalPages:   0,
		pages:        make([]string, 0, 8),
		currentBuf:   &bytes.Buffer{},
		brandingText: "Powered by InsightEngine AI",
	}
	gen.initCharWidths()
	return gen
}

// PageDimensions returns common page sizes in points (1 inch = 72 pt).
func PageDimensions(size PageSize, orientation PageOrientation, customW, customH *int) (float64, float64) {
	var w, h float64

	switch size {
	case PageSizeLetter:
		w, h = 612, 792
	case PageSizeLegal:
		w, h = 612, 1008
	case PageSizeTabloid:
		w, h = 792, 1224
	case PageSizeCustom:
		if customW != nil && customH != nil {
			w = float64(*customW)
			h = float64(*customH)
		} else {
			w, h = 595, 842 // fallback A4
		}
	default: // A4
		w, h = 595, 842
	}

	if orientation == OrientationLandscape {
		w, h = h, w
	}
	return w, h
}

// Generate builds the complete PDF and returns the raw bytes.
func (g *PDFGenerator) Generate(content *PDFContent) []byte {
	g.headerText = content.Title
	g.footerText = content.Footer
	g.watermark = content.Watermark
	g.timestamp = content.Timestamp
	if content.Branding != "" {
		g.brandingText = content.Branding
	}

	// Start first page
	g.newPage()

	// === TITLE PAGE ===
	g.renderTitlePage(content)

	// === DATA SECTIONS ===
	for _, section := range content.Sections {
		g.renderSection(&section)
	}

	// Finalize last page
	g.finalizePage()

	// Build the raw PDF binary
	return g.buildPDFBinary()
}

// ============================================================
// RENDERING METHODS
// ============================================================

func (g *PDFGenerator) renderTitlePage(content *PDFContent) {
	// Title (24pt bold)
	g.setFont("F2", 24) // Helvetica-Bold
	g.setColor(0.1, 0.1, 0.18)
	g.drawText(g.marginLeft, g.cursorY, content.Title)
	g.cursorY -= 32

	// Accent bar
	g.drawLine(g.marginLeft, g.cursorY, g.pageW-g.marginRight, g.cursorY, 2, 0.388, 0.4, 0.945)
	g.cursorY -= 20

	// Subtitle (14pt)
	if content.Subtitle != "" {
		g.setFont("F1", 14)
		g.setColor(0.4, 0.4, 0.4)
		g.drawText(g.marginLeft, g.cursorY, content.Subtitle)
		g.cursorY -= 22
	}

	// Timestamp
	if content.Timestamp != "" {
		g.setFont("F1", 10)
		g.setColor(0.5, 0.5, 0.5)
		g.drawText(g.marginLeft, g.cursorY, content.Timestamp)
		g.cursorY -= 16
	}

	// Metadata block
	if len(content.Metadata) > 0 {
		g.cursorY -= 8
		g.setFont("F1", 10)
		g.setColor(0.45, 0.45, 0.45)
		for key, val := range content.Metadata {
			if g.needsNewPage(16) {
				g.pageBreak()
			}
			g.drawText(g.marginLeft, g.cursorY, fmt.Sprintf("%s: %s", key, val))
			g.cursorY -= 16
		}
	}

	g.cursorY -= 20
}

func (g *PDFGenerator) renderSection(section *PDFSection) {
	// Section title bar
	if g.needsNewPage(50) {
		g.pageBreak()
	}

	// Draw section title with accent background
	barH := 24.0
	barW := g.pageW - g.marginLeft - g.marginRight
	barY := g.cursorY - barH + 4

	// Solid accent background
	g.currentBuf.WriteString(fmt.Sprintf("0.94 0.94 0.98 rg\n%.2f %.2f %.2f %.2f re f\n",
		g.marginLeft, barY, barW, barH))

	// Left accent stripe
	g.currentBuf.WriteString(fmt.Sprintf("0.388 0.4 0.945 rg\n%.2f %.2f 3 %.2f re f\n",
		g.marginLeft, barY, barH))

	// Section title text
	g.setFont("F2", 13)
	g.setColor(0.15, 0.15, 0.2)
	g.drawText(g.marginLeft+12, barY+7, section.Title)
	g.cursorY = barY - 10

	// Data table
	if len(section.Headers) > 0 {
		g.renderTable(section.Headers, section.Rows)
	} else if len(section.Rows) > 0 {
		// Render as simple text lines
		g.setFont("F1", 10)
		g.setColor(0.2, 0.2, 0.2)
		for _, row := range section.Rows {
			if g.needsNewPage(14) {
				g.pageBreak()
			}
			text := strings.Join(row, "  |  ")
			g.drawText(g.marginLeft+8, g.cursorY, text)
			g.cursorY -= 14
		}
	}

	g.cursorY -= 16 // spacing between sections
}

func (g *PDFGenerator) renderTable(headers []string, rows [][]string) {
	tableW := g.pageW - g.marginLeft - g.marginRight
	colCount := len(headers)
	if colCount == 0 {
		return
	}

	// Calculate column widths based on content
	colWidths := g.calculateColumnWidths(headers, rows, tableW)

	rowH := 18.0
	headerH := 22.0
	fontSize := 9.0

	// Adaptive font size for many columns
	if colCount > 8 {
		fontSize = 7.0
		rowH = 14.0
		headerH = 18.0
	} else if colCount > 5 {
		fontSize = 8.0
		rowH = 16.0
		headerH = 20.0
	}

	// === Header row ===
	if g.needsNewPage(headerH + rowH) {
		g.pageBreak()
	}

	headerY := g.cursorY - headerH + 4

	// Header background
	g.currentBuf.WriteString(fmt.Sprintf("0.22 0.22 0.28 rg\n%.2f %.2f %.2f %.2f re f\n",
		g.marginLeft, headerY, tableW, headerH))

	// Header text (white)
	g.setFont("F2", fontSize)
	g.setColor(1, 1, 1)
	xPos := g.marginLeft
	for i, header := range headers {
		cellW := colWidths[i]
		truncated := g.truncateText(header, cellW-8, fontSize)
		g.drawText(xPos+4, headerY+6, truncated)
		xPos += cellW
	}

	g.cursorY = headerY - 1

	// === Data rows ===
	maxRows := len(rows)
	if maxRows > 500 {
		maxRows = 500 // safety cap
	}

	for rowIdx := 0; rowIdx < maxRows; rowIdx++ {
		if g.needsNewPage(rowH + 4) {
			g.pageBreak()
			// Re-draw header on new page
			headerY = g.cursorY - headerH + 4
			g.currentBuf.WriteString(fmt.Sprintf("0.22 0.22 0.28 rg\n%.2f %.2f %.2f %.2f re f\n",
				g.marginLeft, headerY, tableW, headerH))
			g.setFont("F2", fontSize)
			g.setColor(1, 1, 1)
			xPos = g.marginLeft
			for i, header := range headers {
				cellW := colWidths[i]
				truncated := g.truncateText(header, cellW-8, fontSize)
				g.drawText(xPos+4, headerY+6, truncated)
				xPos += cellW
			}
			g.cursorY = headerY - 1
		}

		rowY := g.cursorY - rowH + 2

		// Alternating row color
		if rowIdx%2 == 0 {
			g.currentBuf.WriteString(fmt.Sprintf("0.97 0.97 0.99 rg\n%.2f %.2f %.2f %.2f re f\n",
				g.marginLeft, rowY, tableW, rowH))
		} else {
			g.currentBuf.WriteString(fmt.Sprintf("1 1 1 rg\n%.2f %.2f %.2f %.2f re f\n",
				g.marginLeft, rowY, tableW, rowH))
		}

		// Row bottom border
		g.currentBuf.WriteString(fmt.Sprintf("0.88 0.88 0.9 RG\n0.5 w\n%.2f %.2f m %.2f %.2f l S\n",
			g.marginLeft, rowY, g.marginLeft+tableW, rowY))

		// Cell values
		row := rows[rowIdx]
		g.setFont("F1", fontSize)
		g.setColor(0.2, 0.2, 0.22)
		xPos = g.marginLeft
		for i := 0; i < colCount; i++ {
			cellW := colWidths[i]
			cellValue := ""
			if i < len(row) {
				cellValue = row[i]
			}
			truncated := g.truncateText(cellValue, cellW-8, fontSize)
			g.drawText(xPos+4, rowY+5, truncated)
			xPos += cellW
		}

		g.cursorY = rowY - 1
	}

	// Row count footer
	if len(rows) > maxRows {
		g.cursorY -= 4
		g.setFont("F1", 8)
		g.setColor(0.5, 0.5, 0.5)
		g.drawText(g.marginLeft+4, g.cursorY, fmt.Sprintf("... and %d more rows (truncated for PDF)", len(rows)-maxRows))
		g.cursorY -= 12
	}

	// Total rows indicator
	g.setFont("F1", 8)
	g.setColor(0.6, 0.6, 0.6)
	rowCountText := fmt.Sprintf("Total rows: %d", len(rows))
	g.drawText(g.marginLeft+4, g.cursorY, rowCountText)
	g.cursorY -= 14
}

func (g *PDFGenerator) calculateColumnWidths(headers []string, rows [][]string, totalW float64) []float64 {
	colCount := len(headers)
	widths := make([]float64, colCount)

	// Measure header widths
	for i, h := range headers {
		widths[i] = g.measureText(h, 9) + 16 // padding
	}

	// Measure first N data rows for width estimation
	sampleSize := 20
	if sampleSize > len(rows) {
		sampleSize = len(rows)
	}
	for _, row := range rows[:sampleSize] {
		for i := 0; i < colCount && i < len(row); i++ {
			w := g.measureText(row[i], 9) + 16
			if w > widths[i] {
				widths[i] = w
			}
		}
	}

	// Cap individual column max
	maxColW := totalW * 0.4
	for i := range widths {
		if widths[i] > maxColW {
			widths[i] = maxColW
		}
	}

	// Scale to fit total width
	sum := 0.0
	for _, w := range widths {
		sum += w
	}
	if sum > 0 {
		scale := totalW / sum
		for i := range widths {
			widths[i] = math.Floor(widths[i] * scale)
		}
	}

	// Redistribute rounding errors to last column
	actual := 0.0
	for _, w := range widths {
		actual += w
	}
	if diff := totalW - actual; diff != 0 && colCount > 0 {
		widths[colCount-1] += diff
	}

	return widths
}

// ============================================================
// PAGE MANAGEMENT
// ============================================================

func (g *PDFGenerator) newPage() {
	g.currentPage++
	g.cursorY = g.pageH - g.marginTop
	g.currentBuf = &bytes.Buffer{}
}

func (g *PDFGenerator) finalizePage() {
	// Render header
	g.renderPageHeader()

	// Render footer
	g.renderPageFooter()

	// Render watermark
	if g.watermark != "" {
		g.renderWatermark()
	}

	g.pages = append(g.pages, g.currentBuf.String())
}

func (g *PDFGenerator) pageBreak() {
	g.finalizePage()
	g.newPage()
}

func (g *PDFGenerator) needsNewPage(requiredHeight float64) bool {
	return g.cursorY-requiredHeight < g.marginBottom
}

func (g *PDFGenerator) renderPageHeader() {
	if g.headerText == "" {
		return
	}

	headerY := g.pageH - 35

	// Thin top accent bar
	g.currentBuf.WriteString(fmt.Sprintf("0.388 0.4 0.945 rg\n0 %.2f %.2f 3 re f\n", g.pageH-3, g.pageW))

	// Header text
	g.currentBuf.WriteString("BT\n")
	g.currentBuf.WriteString(fmt.Sprintf("/F1 8 Tf\n0.5 0.5 0.5 rg\n%.2f %.2f Td\n(%s) Tj\n",
		g.marginLeft, headerY, pdfEscapeString(g.headerText)))
	g.currentBuf.WriteString("ET\n")

	// Page number (right aligned) — placeholder, replaced in post-processing
	pageNumText := fmt.Sprintf("Page %d", g.currentPage)
	pageNumW := g.measureText(pageNumText, 8)
	g.currentBuf.WriteString("BT\n")
	g.currentBuf.WriteString(fmt.Sprintf("/F1 8 Tf\n0.5 0.5 0.5 rg\n%.2f %.2f Td\n(%s) Tj\n",
		g.pageW-g.marginRight-pageNumW, headerY, pdfEscapeString(pageNumText)))
	g.currentBuf.WriteString("ET\n")

	// Separator line
	g.currentBuf.WriteString(fmt.Sprintf("0.85 0.85 0.88 RG\n0.5 w\n%.2f %.2f m %.2f %.2f l S\n",
		g.marginLeft, headerY-6, g.pageW-g.marginRight, headerY-6))
}

func (g *PDFGenerator) renderPageFooter() {
	footerY := 30.0

	// Separator line
	g.currentBuf.WriteString(fmt.Sprintf("0.85 0.85 0.88 RG\n0.5 w\n%.2f %.2f m %.2f %.2f l S\n",
		g.marginLeft, footerY+12, g.pageW-g.marginRight, footerY+12))

	// Footer text (left)
	if g.footerText != "" {
		g.currentBuf.WriteString("BT\n")
		g.currentBuf.WriteString(fmt.Sprintf("/F1 8 Tf\n0.6 0.6 0.6 rg\n%.2f %.2f Td\n(%s) Tj\n",
			g.marginLeft, footerY, pdfEscapeString(g.footerText)))
		g.currentBuf.WriteString("ET\n")
	}

	// Branding (right)
	brandW := g.measureText(g.brandingText, 8)
	g.currentBuf.WriteString("BT\n")
	g.currentBuf.WriteString(fmt.Sprintf("/F1 8 Tf\n0.388 0.4 0.945 rg\n%.2f %.2f Td\n(%s) Tj\n",
		g.pageW-g.marginRight-brandW, footerY, pdfEscapeString(g.brandingText)))
	g.currentBuf.WriteString("ET\n")
}

func (g *PDFGenerator) renderWatermark() {
	// Render large diagonal watermark text with transparency
	// Using ExtGState for alpha
	centerX := g.pageW / 2
	centerY := g.pageH / 2

	// Diagonal angle: ~35 degrees
	angle := 35.0 * math.Pi / 180.0
	cosA := math.Cos(angle)
	sinA := math.Sin(angle)

	// Save graphics state, apply transform, draw text, restore
	g.currentBuf.WriteString("q\n")
	// Semi-transparent gray
	g.currentBuf.WriteString("0.85 0.85 0.85 rg\n")
	g.currentBuf.WriteString("BT\n")
	g.currentBuf.WriteString(fmt.Sprintf("/F2 48 Tf\n"))
	// Text matrix: rotation + translation to center
	textW := g.measureText(g.watermark, 48) / 2
	g.currentBuf.WriteString(fmt.Sprintf("%.4f %.4f %.4f %.4f %.2f %.2f Tm\n",
		cosA, sinA, -sinA, cosA, centerX-textW*cosA, centerY-textW*sinA))
	g.currentBuf.WriteString(fmt.Sprintf("(%s) Tj\n", pdfEscapeString(g.watermark)))
	g.currentBuf.WriteString("ET\n")
	g.currentBuf.WriteString("Q\n")
}

// ============================================================
// LOW-LEVEL PDF DRAWING PRIMITIVES
// ============================================================

func (g *PDFGenerator) setFont(fontName string, size float64) {
	g.currentBuf.WriteString(fmt.Sprintf("/%s %.1f Tf\n", fontName, size))
}

func (g *PDFGenerator) setColor(r, g2, b float64) {
	g.currentBuf.WriteString(fmt.Sprintf("%.3f %.3f %.3f rg\n", r, g2, b))
}

func (g *PDFGenerator) drawText(x, y float64, text string) {
	g.currentBuf.WriteString("BT\n")
	g.currentBuf.WriteString(fmt.Sprintf("%.2f %.2f Td\n(%s) Tj\n", x, y, pdfEscapeString(text)))
	g.currentBuf.WriteString("ET\n")
}

func (g *PDFGenerator) drawLine(x1, y1, x2, y2, width, r, green, b float64) {
	g.currentBuf.WriteString(fmt.Sprintf("%.3f %.3f %.3f RG\n%.1f w\n%.2f %.2f m %.2f %.2f l S\n",
		r, green, b, width, x1, y1, x2, y2))
}

// ============================================================
// TEXT MEASUREMENT (approximate Helvetica metrics)
// ============================================================

func (g *PDFGenerator) initCharWidths() {
	// Approximate Helvetica glyph widths normalized to 1000 units per em.
	// Common subset covering ASCII printable range (32-126).
	g.charWidths = map[byte]float64{
		' ': 278, '!': 278, '"': 355, '#': 556, '$': 556, '%': 889, '&': 667, '\'': 191,
		'(': 333, ')': 333, '*': 389, '+': 584, ',': 278, '-': 333, '.': 278, '/': 278,
		'0': 556, '1': 556, '2': 556, '3': 556, '4': 556, '5': 556, '6': 556, '7': 556,
		'8': 556, '9': 556, ':': 278, ';': 278, '<': 584, '=': 584, '>': 584, '?': 556,
		'@': 1015, 'A': 667, 'B': 667, 'C': 722, 'D': 722, 'E': 667, 'F': 611, 'G': 778,
		'H': 722, 'I': 278, 'J': 500, 'K': 667, 'L': 556, 'M': 833, 'N': 722, 'O': 778,
		'P': 667, 'Q': 778, 'R': 722, 'S': 667, 'T': 611, 'U': 722, 'V': 667, 'W': 944,
		'X': 667, 'Y': 667, 'Z': 611, '[': 278, '\\': 278, ']': 278, '^': 469, '_': 556,
		'`': 333, 'a': 556, 'b': 556, 'c': 500, 'd': 556, 'e': 556, 'f': 278, 'g': 556,
		'h': 556, 'i': 222, 'j': 222, 'k': 500, 'l': 222, 'm': 833, 'n': 556, 'o': 556,
		'p': 556, 'q': 556, 'r': 333, 's': 500, 't': 278, 'u': 556, 'v': 500, 'w': 722,
		'x': 500, 'y': 500, 'z': 500, '{': 334, '|': 260, '}': 334, '~': 584,
	}
}

func (g *PDFGenerator) measureText(text string, fontSize float64) float64 {
	total := 0.0
	for i := 0; i < len(text); i++ {
		w, ok := g.charWidths[text[i]]
		if !ok {
			w = 500 // default width for unknown glyphs
		}
		total += w
	}
	return total * fontSize / 1000.0
}

func (g *PDFGenerator) truncateText(text string, maxW, fontSize float64) string {
	if g.measureText(text, fontSize) <= maxW {
		return text
	}

	ellipsis := "..."
	ellipsisW := g.measureText(ellipsis, fontSize)
	targetW := maxW - ellipsisW

	consumed := 0.0
	cutIdx := 0
	for i := 0; i < len(text); i++ {
		w, ok := g.charWidths[text[i]]
		if !ok {
			w = 500
		}
		charW := w * fontSize / 1000.0
		if consumed+charW > targetW {
			break
		}
		consumed += charW
		cutIdx = i + 1
	}

	if cutIdx == 0 {
		return ellipsis
	}
	return text[:cutIdx] + ellipsis
}

// ============================================================
// PDF BINARY CONSTRUCTION
// ============================================================

func (g *PDFGenerator) buildPDFBinary() []byte {
	g.totalPages = len(g.pages)
	var buf bytes.Buffer
	offsets := make([]int, 0, g.totalPages*2+5)

	// PDF Header
	buf.WriteString("%PDF-1.4\n")
	buf.Write([]byte{'%', 0xE2, 0xE3, 0xCF, 0xD3, '\n'})

	// Object 1: Catalog
	offsets = append(offsets, buf.Len())
	buf.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")

	// Object 2: Pages tree — build Kids array
	kidsArr := make([]string, g.totalPages)
	for i := 0; i < g.totalPages; i++ {
		pageObjNum := 5 + i*2 // Page objects start at obj 5, alternating with streams
		kidsArr[i] = fmt.Sprintf("%d 0 R", pageObjNum)
	}
	offsets = append(offsets, buf.Len())
	buf.WriteString(fmt.Sprintf("2 0 obj\n<< /Type /Pages /Kids [%s] /Count %d >>\nendobj\n",
		strings.Join(kidsArr, " "), g.totalPages))

	// Object 3: Font Helvetica
	offsets = append(offsets, buf.Len())
	buf.WriteString("3 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >>\nendobj\n")

	// Object 4: Font Helvetica-Bold
	offsets = append(offsets, buf.Len())
	buf.WriteString("4 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold /Encoding /WinAnsiEncoding >>\nendobj\n")

	// Page objects + content streams (interleaved)
	for i, pageContent := range g.pages {
		pageObjNum := 5 + i*2
		streamObjNum := pageObjNum + 1

		// Replace page number placeholder with actual total
		finalContent := strings.ReplaceAll(pageContent,
			fmt.Sprintf("Page %d", i+1),
			fmt.Sprintf("Page %d of %d", i+1, g.totalPages))

		// Page object
		offsets = append(offsets, buf.Len())
		buf.WriteString(fmt.Sprintf("%d 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %.0f %.0f] "+
			"/Contents %d 0 R /Resources << /Font << /F1 3 0 R /F2 4 0 R >> >> >>\nendobj\n",
			pageObjNum, g.pageW, g.pageH, streamObjNum))

		// Content stream object
		offsets = append(offsets, buf.Len())
		buf.WriteString(fmt.Sprintf("%d 0 obj\n<< /Length %d >>\nstream\n%s\nendstream\nendobj\n",
			streamObjNum, len(finalContent), finalContent))
	}

	// Cross-reference table
	totalObjs := len(offsets)
	xrefOffset := buf.Len()
	buf.WriteString("xref\n")
	buf.WriteString(fmt.Sprintf("0 %d\n", totalObjs+1))
	buf.WriteString("0000000000 65535 f \n")
	for _, off := range offsets {
		buf.WriteString(fmt.Sprintf("%010d 00000 n \n", off))
	}

	// Trailer
	buf.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R >>\n", totalObjs+1))
	buf.WriteString(fmt.Sprintf("startxref\n%d\n%%%%EOF\n", xrefOffset))

	return buf.Bytes()
}

// ============================================================
// HELPER — Format timestamp for PDF
// ============================================================

// FormatPDFTimestamp returns a formatted timestamp string for PDF metadata.
func FormatPDFTimestamp() string {
	return fmt.Sprintf("Generated: %s", time.Now().UTC().Format("2006-01-02 15:04:05 UTC"))
}
