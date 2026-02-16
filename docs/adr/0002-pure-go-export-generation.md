# ADR-002: Pure Go Export Generation (No External Binaries)

## Status

Accepted

## Date

2026-02-16

## Context

The dashboard export system needs to generate PDF, PNG, JPEG, and PPTX files from dashboard data. The initial approach considered was `chromedp` (headless Chrome) for PDF/image rendering, but this introduces several deployment constraints:

1. **Binary dependency**: Requires Chrome/Chromium installed in Docker images (~400MB+ image size increase).
2. **Resource consumption**: Headless Chrome uses 200-500MB RAM per render context.
3. **Startup latency**: Chrome process initialization adds 2-5 seconds to each export.
4. **CI/CD complexity**: Requires Chrome installation in all build/test environments.
5. **Security surface**: Running a full browser engine introduces potential sandbox escape vectors.

## Decision

Use **pure Go standard library** for all export generation:

- **PDF**: Raw PDF 1.4 stream construction (`bytes.Buffer` + cross-reference table). Uses built-in Helvetica (Type1 font, no embedding required).
- **PNG/JPEG**: Go `image`, `image/draw`, `image/png`, `image/jpeg` packages. Canvas-based rendering with resolution scaling.
- **PPTX**: `archive/zip` + OOXML templates. Custom `PPTXGenerator` building valid `.pptx` from `SlideDeck` model.

**Zero external dependencies.** The same Go binary that serves the API also generates all exports.

## Consequences

### Positive

- **Zero deployment dependencies**: No Chrome, no wkhtmltopdf, no LibreOffice. Just the Go binary.
- **Deterministic output**: Same input always produces the same output. No browser rendering non-determinism.
- **Low resource usage**: Export generation uses ~10-50MB RAM vs. 200-500MB for headless Chrome.
- **Fast cold start**: No Chrome process startup. Export generation begins immediately.
- **Smaller Docker images**: No Chrome installation (~400MB saved).

### Negative

- **Limited rendering fidelity**: Cannot render actual dashboard HTML/CSS/JS. Exports show metadata + structural layout rather than pixel-perfect dashboard screenshots.
- **No font rendering in images**: Text rendering in PNG/JPEG requires a font rasterizer (not included). Images show layout/structure without text.
- **Future enhancement path**: If pixel-perfect rendering is required, `chromedp` can be added as an optional enhancement layer on top of this foundation.
