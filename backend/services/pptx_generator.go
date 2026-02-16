package services

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"insight-engine-backend/models"
	"strings"
	"time"
)

// PPTXGenerator creates valid .pptx files from SlideDeck models.
// Uses pure Go — no external binaries or libraries required.
// A .pptx file is a ZIP archive with an OPC (Open Packaging Convention) structure.
type PPTXGenerator struct{}

// NewPPTXGenerator creates a new PPTX generator
func NewPPTXGenerator() *PPTXGenerator {
	return &PPTXGenerator{}
}

// GeneratePPTX creates a valid .pptx binary from a SlideDeck model
func (g *PPTXGenerator) GeneratePPTX(deck *models.SlideDeck) ([]byte, error) {
	if deck == nil {
		return nil, fmt.Errorf("slide deck cannot be nil")
	}
	if len(deck.Slides) == 0 {
		return nil, fmt.Errorf("slide deck must contain at least one slide")
	}

	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	// Total slides = 1 title slide + N content slides
	totalSlides := 1 + len(deck.Slides)

	// 1. Write [Content_Types].xml
	if err := g.writeContentTypes(zw, totalSlides); err != nil {
		return nil, fmt.Errorf("failed to write content types: %w", err)
	}

	// 2. Write _rels/.rels
	if err := g.writeRootRels(zw); err != nil {
		return nil, fmt.Errorf("failed to write root rels: %w", err)
	}

	// 3. Write docProps/app.xml
	if err := g.writeAppProps(zw, deck); err != nil {
		return nil, fmt.Errorf("failed to write app props: %w", err)
	}

	// 4. Write docProps/core.xml
	if err := g.writeCoreProps(zw, deck); err != nil {
		return nil, fmt.Errorf("failed to write core props: %w", err)
	}

	// 5. Write ppt/presentation.xml
	if err := g.writePresentation(zw, totalSlides); err != nil {
		return nil, fmt.Errorf("failed to write presentation: %w", err)
	}

	// 6. Write ppt/_rels/presentation.xml.rels
	if err := g.writePresentationRels(zw, totalSlides); err != nil {
		return nil, fmt.Errorf("failed to write presentation rels: %w", err)
	}

	// 7. Write ppt/slideMasters/slideMaster1.xml
	if err := g.writeSlideMaster(zw); err != nil {
		return nil, fmt.Errorf("failed to write slide master: %w", err)
	}

	// 8. Write ppt/slideMasters/_rels/slideMaster1.xml.rels
	if err := g.writeSlideMasterRels(zw, totalSlides); err != nil {
		return nil, fmt.Errorf("failed to write slide master rels: %w", err)
	}

	// 9. Write ppt/slideLayouts/slideLayout1.xml
	if err := g.writeSlideLayout(zw); err != nil {
		return nil, fmt.Errorf("failed to write slide layout: %w", err)
	}

	// 10. Write ppt/slideLayouts/_rels/slideLayout1.xml.rels
	if err := g.writeSlideLayoutRels(zw); err != nil {
		return nil, fmt.Errorf("failed to write slide layout rels: %w", err)
	}

	// 11. Write ppt/theme/theme1.xml
	if err := g.writeTheme(zw); err != nil {
		return nil, fmt.Errorf("failed to write theme: %w", err)
	}

	// 12. Write title slide (slide1)
	if err := g.writeTitleSlide(zw, deck); err != nil {
		return nil, fmt.Errorf("failed to write title slide: %w", err)
	}
	if err := g.writeSlideRels(zw, 1); err != nil {
		return nil, fmt.Errorf("failed to write slide1 rels: %w", err)
	}

	// 13. Write content slides (slide2, slide3, ...)
	for i, slide := range deck.Slides {
		slideNum := i + 2
		if err := g.writeContentSlide(zw, slideNum, &slide); err != nil {
			return nil, fmt.Errorf("failed to write slide %d: %w", slideNum, err)
		}
		if err := g.writeSlideRels(zw, slideNum); err != nil {
			return nil, fmt.Errorf("failed to write slide%d rels: %w", slideNum, err)
		}
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zip: %w", err)
	}

	return buf.Bytes(), nil
}

// EMU helpers — PowerPoint uses English Metric Units (1 inch = 914400 EMU)
const (
	emuPerInch = 914400
	slideW     = 12192000 // 13.333 inches (16:9 widescreen)
	slideH     = 6858000  // 7.5 inches
)

func inchToEMU(inches float64) int64 {
	return int64(inches * float64(emuPerInch))
}

// ============================================================================
// Content_Types
// ============================================================================

func (g *PPTXGenerator) writeContentTypes(zw *zip.Writer, totalSlides int) error {
	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	sb.WriteString(`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">`)
	sb.WriteString(`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>`)
	sb.WriteString(`<Default Extension="xml" ContentType="application/xml"/>`)
	sb.WriteString(`<Override PartName="/ppt/presentation.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml"/>`)
	sb.WriteString(`<Override PartName="/ppt/slideMasters/slideMaster1.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slideMaster+xml"/>`)
	sb.WriteString(`<Override PartName="/ppt/slideLayouts/slideLayout1.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slideLayout+xml"/>`)
	sb.WriteString(`<Override PartName="/ppt/theme/theme1.xml" ContentType="application/vnd.openxmlformats-officedocument.theme+xml"/>`)
	sb.WriteString(`<Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/>`)
	sb.WriteString(`<Override PartName="/docProps/app.xml" ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml"/>`)
	for i := 1; i <= totalSlides; i++ {
		sb.WriteString(fmt.Sprintf(`<Override PartName="/ppt/slides/slide%d.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slide+xml"/>`, i))
	}
	sb.WriteString(`</Types>`)
	return g.writeZipEntry(zw, "[Content_Types].xml", sb.String())
}

// ============================================================================
// Relationships
// ============================================================================

func (g *PPTXGenerator) writeRootRels(zw *zip.Writer) error {
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="ppt/presentation.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/>
  <Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties" Target="docProps/app.xml"/>
</Relationships>`
	return g.writeZipEntry(zw, "_rels/.rels", content)
}

func (g *PPTXGenerator) writePresentationRels(zw *zip.Writer, totalSlides int) error {
	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	sb.WriteString(`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)
	sb.WriteString(`<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideMaster" Target="slideMasters/slideMaster1.xml"/>`)
	sb.WriteString(`<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme" Target="theme/theme1.xml"/>`)
	for i := 1; i <= totalSlides; i++ {
		rId := fmt.Sprintf("rId%d", i+2)
		sb.WriteString(fmt.Sprintf(`<Relationship Id="%s" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide" Target="slides/slide%d.xml"/>`, rId, i))
	}
	sb.WriteString(`</Relationships>`)
	return g.writeZipEntry(zw, "ppt/_rels/presentation.xml.rels", sb.String())
}

func (g *PPTXGenerator) writeSlideRels(zw *zip.Writer, slideNum int) error {
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout" Target="../slideLayouts/slideLayout1.xml"/>
</Relationships>`
	path := fmt.Sprintf("ppt/slides/_rels/slide%d.xml.rels", slideNum)
	return g.writeZipEntry(zw, path, content)
}

func (g *PPTXGenerator) writeSlideMasterRels(zw *zip.Writer, totalSlides int) error {
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout" Target="../slideLayouts/slideLayout1.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme" Target="../theme/theme1.xml"/>
</Relationships>`
	return g.writeZipEntry(zw, "ppt/slideMasters/_rels/slideMaster1.xml.rels", content)
}

func (g *PPTXGenerator) writeSlideLayoutRels(zw *zip.Writer) error {
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideMaster" Target="../slideMasters/slideMaster1.xml"/>
</Relationships>`
	return g.writeZipEntry(zw, "ppt/slideLayouts/_rels/slideLayout1.xml.rels", content)
}

// ============================================================================
// Document Properties
// ============================================================================

func (g *PPTXGenerator) writeAppProps(zw *zip.Writer, deck *models.SlideDeck) error {
	totalSlides := 1 + len(deck.Slides)
	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties" xmlns:vt="http://schemas.openxmlformats.org/officeDocument/2006/docPropsVTypes">
  <Application>InsightEngine AI</Application>
  <Slides>%d</Slides>
  <ScaleCrop>false</ScaleCrop>
  <LinksUpToDate>false</LinksUpToDate>
  <SharedDoc>false</SharedDoc>
  <HyperlinksChanged>false</HyperlinksChanged>
</Properties>`, totalSlides)
	return g.writeZipEntry(zw, "docProps/app.xml", content)
}

func (g *PPTXGenerator) writeCoreProps(zw *zip.Writer, deck *models.SlideDeck) error {
	now := time.Now().UTC().Format(time.RFC3339)
	escapedTitle := xmlEscape(deck.Title)
	escapedDesc := xmlEscape(deck.Description)
	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:dcmitype="http://purl.org/dc/dcmitype/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <dc:title>%s</dc:title>
  <dc:subject>%s</dc:subject>
  <dc:creator>InsightEngine AI</dc:creator>
  <dcterms:created xsi:type="dcterms:W3CDTF">%s</dcterms:created>
  <dcterms:modified xsi:type="dcterms:W3CDTF">%s</dcterms:modified>
</cp:coreProperties>`, escapedTitle, escapedDesc, now, now)
	return g.writeZipEntry(zw, "docProps/core.xml", content)
}

// ============================================================================
// Presentation.xml
// ============================================================================

func (g *PPTXGenerator) writePresentation(zw *zip.Writer, totalSlides int) error {
	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	sb.WriteString(`<p:presentation xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">`)
	sb.WriteString(fmt.Sprintf(`<p:sldMasterIdLst><p:sldMasterId id="2147483648" r:id="rId1"/></p:sldMasterIdLst>`))
	sb.WriteString(`<p:sldIdLst>`)
	for i := 1; i <= totalSlides; i++ {
		sldID := 255 + i
		rId := fmt.Sprintf("rId%d", i+2)
		sb.WriteString(fmt.Sprintf(`<p:sldId id="%d" r:id="%s"/>`, sldID, rId))
	}
	sb.WriteString(`</p:sldIdLst>`)
	sb.WriteString(fmt.Sprintf(`<p:sldSz cx="%d" cy="%d"/>`, slideW, slideH))
	sb.WriteString(fmt.Sprintf(`<p:notesSz cx="%d" cy="%d"/>`, slideH, slideW))
	sb.WriteString(`</p:presentation>`)
	return g.writeZipEntry(zw, "ppt/presentation.xml", sb.String())
}

// ============================================================================
// Slide Master + Layout + Theme (minimal but valid)
// ============================================================================

func (g *PPTXGenerator) writeSlideMaster(zw *zip.Writer) error {
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sldMaster xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
  <p:cSld>
    <p:bg>
      <p:bgPr>
        <a:solidFill><a:srgbClr val="FFFFFF"/></a:solidFill>
        <a:effectLst/>
      </p:bgPr>
    </p:bg>
    <p:spTree>
      <p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr>
      <p:grpSpPr/>
    </p:spTree>
  </p:cSld>
  <p:clrMap bg1="lt1" tx1="dk1" bg2="lt2" tx2="dk2" accent1="accent1" accent2="accent2" accent3="accent3" accent4="accent4" accent5="accent5" accent6="accent6" hlink="hlink" folHlink="folHlink"/>
  <p:sldLayoutIdLst>
    <p:sldLayoutId id="2147483649" r:id="rId1"/>
  </p:sldLayoutIdLst>
</p:sldMaster>`
	return g.writeZipEntry(zw, "ppt/slideMasters/slideMaster1.xml", content)
}

func (g *PPTXGenerator) writeSlideLayout(zw *zip.Writer) error {
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sldLayout xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" type="blank" preserve="1">
  <p:cSld name="Blank">
    <p:spTree>
      <p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr>
      <p:grpSpPr/>
    </p:spTree>
  </p:cSld>
</p:sldLayout>`
	return g.writeZipEntry(zw, "ppt/slideLayouts/slideLayout1.xml", content)
}

func (g *PPTXGenerator) writeTheme(zw *zip.Writer) error {
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<a:theme xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" name="InsightEngine">
  <a:themeElements>
    <a:clrScheme name="InsightEngine">
      <a:dk1><a:srgbClr val="1A1A2E"/></a:dk1>
      <a:lt1><a:srgbClr val="FFFFFF"/></a:lt1>
      <a:dk2><a:srgbClr val="16213E"/></a:dk2>
      <a:lt2><a:srgbClr val="F0F0F5"/></a:lt2>
      <a:accent1><a:srgbClr val="6366F1"/></a:accent1>
      <a:accent2><a:srgbClr val="8B5CF6"/></a:accent2>
      <a:accent3><a:srgbClr val="06B6D4"/></a:accent3>
      <a:accent4><a:srgbClr val="10B981"/></a:accent4>
      <a:accent5><a:srgbClr val="F59E0B"/></a:accent5>
      <a:accent6><a:srgbClr val="EF4444"/></a:accent6>
      <a:hlink><a:srgbClr val="6366F1"/></a:hlink>
      <a:folHlink><a:srgbClr val="8B5CF6"/></a:folHlink>
    </a:clrScheme>
    <a:fontScheme name="InsightEngine">
      <a:majorFont><a:latin typeface="Inter"/><a:ea typeface=""/><a:cs typeface=""/></a:majorFont>
      <a:minorFont><a:latin typeface="Inter"/><a:ea typeface=""/><a:cs typeface=""/></a:minorFont>
    </a:fontScheme>
    <a:fmtScheme name="InsightEngine">
      <a:fillStyleLst>
        <a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
        <a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
        <a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
      </a:fillStyleLst>
      <a:lnStyleLst>
        <a:ln w="9525"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill></a:ln>
        <a:ln w="9525"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill></a:ln>
        <a:ln w="9525"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill></a:ln>
      </a:lnStyleLst>
      <a:effectStyleLst>
        <a:effectStyle><a:effectLst/></a:effectStyle>
        <a:effectStyle><a:effectLst/></a:effectStyle>
        <a:effectStyle><a:effectLst/></a:effectStyle>
      </a:effectStyleLst>
      <a:bgFillStyleLst>
        <a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
        <a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
        <a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
      </a:bgFillStyleLst>
    </a:fmtScheme>
  </a:themeElements>
</a:theme>`
	return g.writeZipEntry(zw, "ppt/theme/theme1.xml", content)
}

// ============================================================================
// Slide Generation
// ============================================================================

func (g *PPTXGenerator) writeTitleSlide(zw *zip.Writer, deck *models.SlideDeck) error {
	escapedTitle := xmlEscape(deck.Title)
	escapedDesc := xmlEscape(deck.Description)

	titleX := inchToEMU(1.0)
	titleY := inchToEMU(2.0)
	titleW := inchToEMU(11.333)
	titleH := inchToEMU(1.5)

	subtitleX := inchToEMU(2.0)
	subtitleY := inchToEMU(3.8)
	subtitleW := inchToEMU(9.333)
	subtitleH := inchToEMU(1.0)

	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sld xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
  <p:cSld>
    <p:bg>
      <p:bgPr>
        <a:gradFill>
          <a:gsLst>
            <a:gs pos="0"><a:srgbClr val="1A1A2E"/></a:gs>
            <a:gs pos="100000"><a:srgbClr val="16213E"/></a:gs>
          </a:gsLst>
          <a:lin ang="5400000" scaled="1"/>
        </a:gradFill>
        <a:effectLst/>
      </p:bgPr>
    </p:bg>
    <p:spTree>
      <p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr>
      <p:grpSpPr/>
      <p:sp>
        <p:nvSpPr><p:cNvPr id="2" name="Title"/><p:cNvSpPr/><p:nvPr/></p:nvSpPr>
        <p:spPr>
          <a:xfrm><a:off x="%d" y="%d"/><a:ext cx="%d" cy="%d"/></a:xfrm>
          <a:prstGeom prst="rect"><a:avLst/></a:prstGeom>
        </p:spPr>
        <p:txBody>
          <a:bodyPr anchor="ctr"/>
          <a:p>
            <a:pPr algn="ctr"/>
            <a:r>
              <a:rPr lang="en-US" sz="3600" b="1" dirty="0"><a:solidFill><a:srgbClr val="FFFFFF"/></a:solidFill><a:latin typeface="Inter"/></a:rPr>
              <a:t>%s</a:t>
            </a:r>
          </a:p>
        </p:txBody>
      </p:sp>
      <p:sp>
        <p:nvSpPr><p:cNvPr id="3" name="Subtitle"/><p:cNvSpPr/><p:nvPr/></p:nvSpPr>
        <p:spPr>
          <a:xfrm><a:off x="%d" y="%d"/><a:ext cx="%d" cy="%d"/></a:xfrm>
          <a:prstGeom prst="rect"><a:avLst/></a:prstGeom>
        </p:spPr>
        <p:txBody>
          <a:bodyPr anchor="ctr"/>
          <a:p>
            <a:pPr algn="ctr"/>
            <a:r>
              <a:rPr lang="en-US" sz="1800" dirty="0"><a:solidFill><a:srgbClr val="A0A0C0"/></a:solidFill><a:latin typeface="Inter"/></a:rPr>
              <a:t>%s</a:t>
            </a:r>
          </a:p>
        </p:txBody>
      </p:sp>
    </p:spTree>
  </p:cSld>
</p:sld>`,
		titleX, titleY, titleW, titleH, escapedTitle,
		subtitleX, subtitleY, subtitleW, subtitleH, escapedDesc)

	return g.writeZipEntry(zw, "ppt/slides/slide1.xml", content)
}

func (g *PPTXGenerator) writeContentSlide(zw *zip.Writer, slideNum int, slide *models.Slide) error {
	escapedTitle := xmlEscape(slide.Title)

	titleX := inchToEMU(0.5)
	titleY := inchToEMU(0.3)
	titleW := inchToEMU(12.333)
	titleH := inchToEMU(0.8)

	bodyX := inchToEMU(0.8)
	bodyY := inchToEMU(1.4)
	bodyW := inchToEMU(11.733)
	bodyH := inchToEMU(5.5)

	var bodyXML string

	switch slide.Layout {
	case "chart_focus":
		bodyXML = g.buildChartPlaceholder(bodyX, bodyY, bodyW, bodyH, slide.ChartID)
	case "bullet_points", "title_and_body", "two_columns":
		bodyXML = g.buildBulletPoints(bodyX, bodyY, bodyW, bodyH, slide.BulletPoints)
	case "title_only", "blank":
		bodyXML = ""
	default:
		bodyXML = g.buildBulletPoints(bodyX, bodyY, bodyW, bodyH, slide.BulletPoints)
	}

	var notesXML string
	if slide.SpeakerNotes != "" {
		notesXML = g.buildSpeakerNotes(slide.SpeakerNotes)
	}

	// Accent bar — thin colored rectangle at the top
	accentBarX := int64(0)
	accentBarY := int64(0)
	accentBarW := int64(slideW)
	accentBarH := inchToEMU(0.06)

	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sld xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
  <p:cSld>
    <p:spTree>
      <p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr>
      <p:grpSpPr/>
      <p:sp>
        <p:nvSpPr><p:cNvPr id="2" name="AccentBar"/><p:cNvSpPr/><p:nvPr/></p:nvSpPr>
        <p:spPr>
          <a:xfrm><a:off x="%d" y="%d"/><a:ext cx="%d" cy="%d"/></a:xfrm>
          <a:prstGeom prst="rect"><a:avLst/></a:prstGeom>
          <a:solidFill><a:srgbClr val="6366F1"/></a:solidFill>
          <a:ln><a:noFill/></a:ln>
        </p:spPr>
      </p:sp>
      <p:sp>
        <p:nvSpPr><p:cNvPr id="3" name="Title"/><p:cNvSpPr/><p:nvPr/></p:nvSpPr>
        <p:spPr>
          <a:xfrm><a:off x="%d" y="%d"/><a:ext cx="%d" cy="%d"/></a:xfrm>
          <a:prstGeom prst="rect"><a:avLst/></a:prstGeom>
        </p:spPr>
        <p:txBody>
          <a:bodyPr anchor="b"/>
          <a:p>
            <a:r>
              <a:rPr lang="en-US" sz="2800" b="1" dirty="0"><a:solidFill><a:srgbClr val="1A1A2E"/></a:solidFill><a:latin typeface="Inter"/></a:rPr>
              <a:t>%s</a:t>
            </a:r>
          </a:p>
        </p:txBody>
      </p:sp>
      %s
    </p:spTree>
  </p:cSld>
  %s
</p:sld>`,
		accentBarX, accentBarY, accentBarW, accentBarH,
		titleX, titleY, titleW, titleH, escapedTitle,
		bodyXML,
		notesXML)

	path := fmt.Sprintf("ppt/slides/slide%d.xml", slideNum)
	return g.writeZipEntry(zw, path, content)
}

func (g *PPTXGenerator) buildBulletPoints(x, y, w, h int64, bullets []string) string {
	if len(bullets) == 0 {
		return ""
	}

	var paragraphs strings.Builder
	for _, bullet := range bullets {
		escaped := xmlEscape(bullet)
		paragraphs.WriteString(fmt.Sprintf(`
      <a:p>
        <a:pPr marL="457200" indent="-228600">
          <a:buFont typeface="Arial"/>
          <a:buChar char="•"/>
        </a:pPr>
        <a:r>
          <a:rPr lang="en-US" sz="1800" dirty="0"><a:solidFill><a:srgbClr val="333333"/></a:solidFill><a:latin typeface="Inter"/></a:rPr>
          <a:t>%s</a:t>
        </a:r>
      </a:p>`, escaped))
	}

	return fmt.Sprintf(`
      <p:sp>
        <p:nvSpPr><p:cNvPr id="4" name="Body"/><p:cNvSpPr/><p:nvPr/></p:nvSpPr>
        <p:spPr>
          <a:xfrm><a:off x="%d" y="%d"/><a:ext cx="%d" cy="%d"/></a:xfrm>
          <a:prstGeom prst="rect"><a:avLst/></a:prstGeom>
        </p:spPr>
        <p:txBody>
          <a:bodyPr anchor="t"/>
          %s
        </p:txBody>
      </p:sp>`, x, y, w, h, paragraphs.String())
}

func (g *PPTXGenerator) buildChartPlaceholder(x, y, w, h int64, chartID string) string {
	msg := fmt.Sprintf("Chart: %s (Render chart images server-side for embedding)", chartID)
	escaped := xmlEscape(msg)
	return fmt.Sprintf(`
      <p:sp>
        <p:nvSpPr><p:cNvPr id="4" name="ChartPlaceholder"/><p:cNvSpPr/><p:nvPr/></p:nvSpPr>
        <p:spPr>
          <a:xfrm><a:off x="%d" y="%d"/><a:ext cx="%d" cy="%d"/></a:xfrm>
          <a:prstGeom prst="rect"><a:avLst/></a:prstGeom>
          <a:solidFill><a:srgbClr val="F0F0F5"/></a:solidFill>
          <a:ln w="12700"><a:solidFill><a:srgbClr val="CCCCCC"/></a:solidFill><a:prstDash val="dash"/></a:ln>
        </p:spPr>
        <p:txBody>
          <a:bodyPr anchor="ctr"/>
          <a:p>
            <a:pPr algn="ctr"/>
            <a:r>
              <a:rPr lang="en-US" sz="1400" i="1" dirty="0"><a:solidFill><a:srgbClr val="666666"/></a:solidFill><a:latin typeface="Inter"/></a:rPr>
              <a:t>%s</a:t>
            </a:r>
          </a:p>
        </p:txBody>
      </p:sp>`, x, y, w, h, escaped)
}

func (g *PPTXGenerator) buildSpeakerNotes(notes string) string {
	escaped := xmlEscape(notes)
	return fmt.Sprintf(`
  <p:notes>
    <p:cSld>
      <p:spTree>
        <p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr>
        <p:grpSpPr/>
        <p:sp>
          <p:nvSpPr><p:cNvPr id="2" name="Notes"/><p:cNvSpPr/><p:nvPr/></p:nvSpPr>
          <p:spPr/>
          <p:txBody>
            <a:bodyPr/>
            <a:p>
              <a:r>
                <a:rPr lang="en-US" dirty="0"/>
                <a:t>%s</a:t>
              </a:r>
            </a:p>
          </p:txBody>
        </p:sp>
      </p:spTree>
    </p:cSld>
  </p:notes>`, escaped)
}

// ============================================================================
// Helpers
// ============================================================================

func (g *PPTXGenerator) writeZipEntry(zw *zip.Writer, name, content string) error {
	w, err := zw.Create(name)
	if err != nil {
		return fmt.Errorf("failed to create zip entry %s: %w", name, err)
	}
	_, err = w.Write([]byte(content))
	if err != nil {
		return fmt.Errorf("failed to write zip entry %s: %w", name, err)
	}
	return nil
}

func xmlEscape(s string) string {
	var buf bytes.Buffer
	if err := xml.EscapeText(&buf, []byte(s)); err != nil {
		return s
	}
	return buf.String()
}
