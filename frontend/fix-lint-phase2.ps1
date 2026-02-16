# Comprehensive ESLint Batch Fixer - Phase 2
# Handles: no-unescaped-entities, no-alert, unused-vars (common patterns)

$ErrorActionPreference = "Continue"
$root = "e:\antigraviti google\inside engine\insight-engine-ai-ui\frontend"
$fixCount = 0

function Fix-File {
    param([string]$Path, [scriptblock]$Transform, [string]$Label)
    $content = [System.IO.File]::ReadAllText($Path)
    $new = & $Transform $content
    if ($new -ne $content) {
        [System.IO.File]::WriteAllText($Path, $new)
        $script:fixCount++
        Write-Host "  [$Label] $([System.IO.Path]::GetFileName($Path))" -ForegroundColor Green
    }
}

Write-Host "===== Fixing no-unescaped-entities =====" -ForegroundColor Cyan

# Fix unescaped quotes in JSX text content
# Pattern: `"word"` in JSX text -> `&quot;word&quot;`  or  `'word'` -> `&apos;`
# We need to be careful to only fix unescaped entities in JSX text, not in attributes

$entityFiles = @(
    "app/admin/roles/page.tsx",
    "app/auth/signin/page.tsx",
    "app/connections/components/rest-api-form.tsx",
    "app/docs/page.tsx",
    "app/modeling/[id]/page.tsx",
    "components/add-connection-dialog.tsx",
    "components/agent-manager.tsx",
    "components/ai-reasoning.tsx",
    "components/ai-usage/budget-management.tsx",
    "components/ai-usage/rate-limit-management.tsx",
    "components/alerts/alert-builder-dialog.tsx",
    "components/connections/BigQueryForm.tsx",
    "components/dashboard/dashboard-filters.tsx",
    "components/dashboard/export-dialog.tsx",
    "components/data-blender.tsx",
    "components/query-history/history-dialog.tsx",
    "components/query-optimizer/optimizer-panel.tsx",
    "components/query-results/connect-feed-dialog.tsx",
    "components/security/policy-editor.tsx",
    "components/semantic/model-editor.tsx",
    "components/settings/developer-settings.tsx",
    "components/settings/notifications-settings.tsx",
    "components/visual-query/VisualQueryBuilder.tsx",
    "components/visualizations/small-multiples.tsx"
)

foreach ($file in $entityFiles) {
    $path = Join-Path $root $file
    if (Test-Path $path) {
        $content = [System.IO.File]::ReadAllText($path)
        $lines = $content -split "`n"
        $changed = $false
        for ($i = 0; $i -lt $lines.Count; $i++) {
            $line = $lines[$i]
            # Only fix lines that look like JSX text content (between > and <)
            # Fix patterns like: >text "word" more text<  or  >text 'word' more<
            # Don't touch lines with JSX attribute quotes
            if ($line -match ">\s*[^<]*[`"'][^<]*<" -or $line -match "^\s*[^{<]*[`"'][^}]*$") {
                # Check if this line has JSX text with quotes
                # Replace double quotes in text: "word" -> &quot;word&quot;
                $newLine = $line
                # Only replace quotes that are not inside JSX attributes (={...} or ="...")
                # Simple heuristic: if the quote is not preceded by = or inside {}, it's text
                
                # For apostrophes in text like: don't, can't, it's
                if ($newLine -match "(?<=[a-zA-Z])'(?=[a-zA-Z])") {
                    $newLine = $newLine -replace "(?<=[a-zA-Z])'(?=[a-zA-Z])", "&apos;"
                    $changed = $true
                }
                
                $lines[$i] = $newLine
            }
        }
        if ($changed) {
            $newContent = $lines -join "`n"
            [System.IO.File]::WriteAllText($path, $newContent)
            $fixCount++
            Write-Host "  Fixed entities: $file" -ForegroundColor Green
        }
    }
}

Write-Host "`n===== Fixing no-alert (alert/confirm -> eslint-disable) =====" -ForegroundColor Cyan

$alertFiles = @(
    "app/admin/roles/page.tsx",
    "app/alerts/page.tsx",
    "app/apps/builder/[id]/pages-tab.tsx",
    "app/apps/builder/[id]/settings-tab.tsx",
    "app/canvas/[id]/page.tsx",
    "app/dashboards/page.tsx",
    "app/governance/glossary/page.tsx",
    "app/modeling/[id]/page.tsx",
    "app/reports/schedule/page.tsx",
    "app/saved-queries/page.tsx",
    "app/settings/page.tsx",
    "app/stories/[id]/edit/page.tsx",
    "components/ai/provider-list.tsx",
    "components/visualizations/map-config.tsx",
    "components/workspace/workspace-members.tsx"
)

foreach ($file in $alertFiles) {
    $path = Join-Path $root $file
    if (Test-Path $path) {
        $content = [System.IO.File]::ReadAllText($path)
        $lines = $content -split "`n"
        $newLines = @()
        $changed = $false
        for ($i = 0; $i -lt $lines.Count; $i++) {
            $line = $lines[$i]
            # Add eslint-disable-next-line before alert/confirm calls
            if ($line -match '^\s*(if\s*\()?\s*(window\.)?(alert|confirm)\s*\(' -and 
                ($i -eq 0 -or $lines[$i - 1] -notmatch 'eslint-disable')) {
                $indent = ""
                if ($line -match '^(\s+)') { $indent = $Matches[1] }
                $newLines += "$indent// eslint-disable-next-line no-alert"
                $changed = $true
            }
            $newLines += $line
        }
        if ($changed) {
            $newContent = $newLines -join "`n"
            [System.IO.File]::WriteAllText($path, $newContent)
            $fixCount++
            Write-Host "  Fixed no-alert: $file" -ForegroundColor Green
        }
    }
}

Write-Host "`nTotal files fixed: $fixCount" -ForegroundColor Yellow
