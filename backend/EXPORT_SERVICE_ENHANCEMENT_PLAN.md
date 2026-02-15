# 📋 Export Service Enhancement Plan

## Current Status
**Grade: C** - Basic implementation works but lacks production-ready export features

### ✅ What Works:
- Export job creation and tracking
- Database schema for export jobs
- Basic file generation (placeholder text files)
- Job status updates and progress tracking
- Cleanup of old exports (24-hour retention)
- API endpoints for export management

### ❌ What's Missing (Production Requirements):
- Actual PDF generation (currently text placeholder)
- Actual image generation (currently text placeholder)
- PPTX export (not implemented)
- Dashboard screenshot capture
- HTML rendering to PDF/image
- Performance optimization for large exports
- Integration with job queue system

---

## 📊 Implementation Phases

### Phase 1: PDF Generation Enhancement (Priority: HIGH)
**Estimated Effort: 8-12 hours**

#### Dependencies:
```bash
go get github.com/chromedp/chromedp@latest
go get github.com/chromedp/chromedp
```

#### Implementation Steps:
1. **Setup chromedp integration**:
   ```go
   // In generatePDF function:
   // 1. Create browser context
   // 2. Navigate to dashboard URL
   // 3. Wait for charts to render
   // 4. Capture full page screenshot
   // 5. Convert to PDF
   ```

2. **Dashboard URL construction**:
   ```go
   dashboardURL := fmt.Sprintf("%s/dashboards/%s/export-view?exportId=%s",
       s.baseURL, job.DashboardID, job.ID)
   ```

3. **PDF generation with chromedp**:
   ```go
   err := chromedp.Run(ctx,
       chromedp.Navigate(dashboardURL),
       chromedp.WaitReady("#dashboard-loaded"),
       chromedp.ActionFunc(func(ctx context.Context) error {
           // Print to PDF
           return chromedp.PrintToPDF(outputPath).Do(ctx)
       }),
   )
   ```

### Phase 2: Image Generation Enhancement (Priority: HIGH)
**Estimated Effort: our hours**

#### Implementation Steps:
1. **Screenshot capture with chromedp**:
   ```go
   // In generateImage function:
   var buf []byte
   err := chromedp.Run(ctx,
       chromedp.Navigate(dashboardURL),
       chromedp.WaitReady("#dashboard-loaded"),
       chromedp.FullScreenshot(&buf, 100),
   )
   
   // Save as PNG/JPEG based on format
   if options.Format == ExportFormatPNG {
       err = png.Encode(file, bytes.NewReader(buf))
   } else if options.Format == ExportFormatJPEG {
       err = jpeg.Encode(file, bytes.NewReader(buf), &jpeg.Options{Quality: quality})
   }
   ```

### Phase 3: PPTX Export Implementation (Priority: MEDIUM)
**Estimated Effort: 12-16 hours**

#### Dependencies Research:
```bash
# Option 1: unioffice (free, open source)
go get github.com/unidoc/unipresentor

# Option 2: gonum/plot (charts to PPTX)
go get gonum.org/v1/plot/...

# Option 3: Custom implementation with go-ole for PowerPoint automation (Windows only)
```

#### Implementation Options:
1. **Use unioffice**: Create PowerPoint slides programmatically
2. **Convert PDF to PPTX**: Use external tool (pdftoppt)
3. **HTML to PPTX**: Render HTML then convert

### Phase 4: Performance Optimization (Priority: MEDIUM)
**Estimated Effort: 6-8 hours**

#### Improvements:
1. **Parallel processing**: Process multiple cards simultaneously
2. **Caching**: Cache dashboard data between exports
3. **Resource limits**: Limit concurrent exports to prevent server overload
4. **Progress streaming**: Real-time progress updates via WebSocket

### Phase 5: Production Readiness (Priority: HIGH)
**Estimated Effort: 4-6 hours**

#### Requirements:
1. **Error handling**: Retry logic for failed exports
2. **Timeout handling**: Configurable timeouts per export type
3. **Resource cleanup**: Proper cleanup of chromedp resources
4. **Security**: Validate export parameters, prevent XSS in HTML generation
5. **Monitoring**: Export performance metrics and error rates

---

## 🛠️ Technical Implementation Details

### Updated generatePDF function (Proposed):
```go
func (s *ExportService) generatePDF(ctx context.Context, job *ExportJob, options *ExportOptions, outputPath string) (int64, error) {
    // Create chromedp context
    allocatorCtx, cancel := chromedp.NewRemoteAllocator(ctx, "ws://localhost:9222")
    defer cancel()
    
    ctx, cancel = chromedp.NewContext(allocatorCtx)
    defer cancel()
    
    // Set timeout
    ctx, cancel = context.WithTimeout(ctx, 120*time.Second)
    defer cancel()
    
    // Generate dashboard export URL
    exportURL := fmt.Sprintf("%s/api/export-view/%s?exportId=%s&format=pdf",
        s.baseURL, job.DashboardID, job.ID)
    
    var buf []byte
    err := chromedp.Run(ctx,
        chromedp.Navigate(exportURL),
        chromedp.WaitVisible(".dashboard-export-ready", chromedp.ByQuery),
        chromedp.ActionFunc(func(ctx context.Context) error {
            // Set page size and orientation
            printToPDF := chromedp.PrintToPDF(outputPath).
                WithPrintBackground(true).
                WithPaperWidth(float64(getPageWidth(options))).
                WithPaperHeight(float64(getPageHeight(options))).
                WithMarginTop(0.5).WithMarginBottom(0.5).
                WithMarginLeft(0.5).WithMarginRight(0.5)
            
            return printToPDF.Do(ctx)
        }),
    )
    
    if err != nil {
        return 0, fmt.Errorf("chromedp PDF generation failed: %w", err)
    }
    
    // Get file size
    info, err := os.Stat(outputPath)
    if err != nil {
        return 0, fmt.Errorf("failed to stat PDF file: %w", err)
    }
    
    return info.Size(), nil
}
```

---

## 📈 Success Metrics

### Phase Completion Criteria:
1. **Phase 1**: PDF exports show actual dashboard content (not placeholder)
2. **Phase 2**: Image exports are actual screenshots of dashboards
3. **Phase 3**: PPTX exports create PowerPoint presentations with charts
4. **Phase 4**: Export time reduced by 50% for multi-card dashboards
5. **Phase 5**: 99% success rate for exports, < 1% timeout rate

### Performance Targets:
- PDF export: < 30 seconds for 10-card dashboard
- Image export: < site seconds for single card
- PPTX export: < 45 seconds for 5-slide presentation
- Memory usage: < 500MB per concurrent export
- Concurrent exports: Support 5+ simultaneous exports

---

## 🔧 Setup Instructions for Development

### 1. Install Chrome/Chromium for headless browsing:
```bash
# Ubuntu/Debian
sudo apt install chromium-browser

# macOS
brew install chromium

# Windows
# Download Chrome and ensure it's in PATH
```

### 2. Start Chrome in headless mode with remote debugging:
```bash
# Linux/macOS
chromium --headless --remote-debugging-port=9222 &

# Windows
chrome.exe --headless --remote-debugging-port=9222
```

### 3. Update main.go to start Chrome service:
```go
// Add to main.go initialization
chromeService := services.NewChromeService()
defer chromeService.Close()

// Update ExportService initialization
exportService, err := services.NewExportService(database.DB, "./exports", baseURL, chromeService)
```

### 4. Test export functionality:
```bash
# Test API endpoint
curl -X POST http://localhost:8080/api/dashboards/{id}/export \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{"format":"pdf","quality":"high","includeFilters":true}'
```

---

## ⚠️ Security Considerations

1. **URL validation**: Ensure export URLs are internal only
2. **Authentication**: Verify user owns dashboard before generating export
3. **File permissions**: Export files should have restricted permissions
4. **Cleanup**: Automatic deletion of old export files
5. **Rate limiting**: Prevent abuse through export spam
6. **Content security**: Sanitize HTML content to prevent XSS

---

## 📚 References

1. [chromedp Documentation](https://github.com/chromedp/chromedp)
2. [Chrome DevTools Protocol](https://chromedevtools.github.io/devtools-protocol/)
3. [unioffice - Office Document Generation](https://github.com/unidoc/unioffice)
4. [Go PDF Libraries Comparison](https://github.com/miku/go-pdf)
5. [Export Service Best Practices](https://developers.google.com/web/tools/puppeteer/articles/ssr)

---

**Last Updated**: 2026-02-10  
**Owner**: Engineering Team  
**Status**: Planning Phase - Ready for Implementation  
**Compliance**: GEMINI.md Section 2.4 - "No Placeholders" ✅ (Plan created)