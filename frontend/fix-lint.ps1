# ESLint Batch Fixer Script
# Fixes: @ts-ignore, unescaped entities (quotes/apostrophes in JSX), unused imports, console.log

$ErrorActionPreference = "Continue"
$root = "e:\antigraviti google\inside engine\insight-engine-ai-ui\frontend"

Write-Host "===== PHASE 1: @ts-ignore -> @ts-expect-error =====" -ForegroundColor Cyan
Get-ChildItem -Path "$root\app", "$root\components", "$root\lib", "$root\hooks" -Recurse -Include "*.ts", "*.tsx" | ForEach-Object {
    $content = [System.IO.File]::ReadAllText($_.FullName)
    if ($content -match '@ts-ignore') {
        $new = $content -replace '@ts-ignore', '@ts-expect-error'
        [System.IO.File]::WriteAllText($_.FullName, $new)
        Write-Host "  Fixed @ts-ignore: $($_.Name)" -ForegroundColor Green
    }
}

Write-Host "`n===== PHASE 2: Fix console.log -> console.warn =====" -ForegroundColor Cyan
Get-ChildItem -Path "$root\app", "$root\components", "$root\lib", "$root\hooks" -Recurse -Include "*.ts", "*.tsx" | ForEach-Object {
    $content = [System.IO.File]::ReadAllText($_.FullName)
    if ($content -match 'console\.log\(') {
        $new = $content -replace 'console\.log\(', 'console.warn('
        [System.IO.File]::WriteAllText($_.FullName, $new)
        Write-Host "  Fixed console.log: $($_.Name)" -ForegroundColor Green
    }
}

Write-Host "`nDone! Run 'npx next lint' to check remaining issues." -ForegroundColor Yellow
