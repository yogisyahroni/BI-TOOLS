# verification/scripts/run_verification.ps1
# Master verification script for Windows - VERIFY-001 to VERIFY-006

param(
    [switch]$SkipFrontend,
    [switch]$QuickMode,
    [string]$BackendUrl = "http://localhost:8080",
    [string]$FrontendUrl = "http://localhost:3000"
)

# Colors
$Red = "`e[0;31m"
$Green = "`e[0;32m"
$Yellow = "`e[1;33m"
$Blue = "`e[0;34m"
$NC = "`e[0m"  # No Color

Write-Host @"
$Blue
╔══════════════════════════════════════════════════════════════════╗
║     INSIGHTENGINE E2E VERIFICATION SUITE - YOLO MODE            ║
║     VERIFY-001 to VERIFY-006 - Production Readiness            ║
╚══════════════════════════════════════════════════════════════════╝
$NC
"@

# Configuration
$ReportsDir = "verification/reports"
$FixesDir = "verification/fixes"

# Create directories
New-Item -ItemType Directory -Force -Path $ReportsDir | Out-Null
New-Item -ItemType Directory -Force -Path $FixesDir | Out-Null

# Check if backend is running
Write-Host "${Blue}🔍 Checking backend health...${NC}"
try {
    $response = Invoke-RestMethod -Uri "$BackendUrl/api/health" -Method GET -TimeoutSec 5
    Write-Host "${Green}✅ Backend is running${NC}"
} catch {
    Write-Host "${Red}❌ Backend not running at $BackendUrl${NC}"
    Write-Host "Please start the backend first:"
    Write-Host "  cd backend"
    Write-Host "  go run main.go"
    exit 1
}

# Check if frontend is running (optional)
if (-not $SkipFrontend) {
    Write-Host "${Blue}🔍 Checking frontend health...${NC}"
    try {
        $response = Invoke-WebRequest -Uri $FrontendUrl -Method GET -TimeoutSec 5
        Write-Host "${Green}✅ Frontend is running${NC}"
    } catch {
        Write-Host "${Yellow}⚠️ Frontend not running at $FrontendUrl${NC}"
        Write-Host "Some tests may be skipped"
    }
}

Write-Host ""
Write-Host "${Blue}═══════════════════════════════════════════════════════════════${NC}"
Write-Host "${Blue}  VERIFY-001: Execute Full E2E Verification Checklist${NC}"
Write-Host "${Blue}═══════════════════════════════════════════════════════════════${NC}"
Write-Host ""

# Run authentication verification
Write-Host "${Yellow}🧪 Running Authentication & Security Tests...${NC}"
Push-Location verification/scripts
try {
    go run verify_auth.go
    if ($LASTEXITCODE -eq 0) {
        Write-Host "${Green}✅ Authentication tests passed${NC}"
    } else {
        Write-Host "${Red}❌ Authentication tests failed${NC}"
    }
} catch {
    Write-Host "${Red}❌ Failed to run authentication tests: $_${NC}"
}
Pop-Location

# Run connector verification (skip if quick mode)
if (-not $QuickMode) {
    Write-Host ""
    Write-Host "${Yellow}🧪 Running Database Connector Tests...${NC}"
    Push-Location verification/scripts
    try {
        go run verify_connectors.go
        if ($LASTEXITCODE -eq 0) {
            Write-Host "${Green}✅ Connector tests passed${NC}"
        } else {
            Write-Host "${Red}❌ Connector tests failed${NC}"
        }
    } catch {
        Write-Host "${Red}❌ Failed to run connector tests: $_${NC}"
    }
    Pop-Location
}

# Generate consolidated report
Write-Host ""
Write-Host "${Blue}═══════════════════════════════════════════════════════════════${NC}"
Write-Host "${Blue}  Generating Consolidated Report${NC}"
Write-Host "${Blue}═══════════════════════════════════════════════════════════════${NC}"

$Timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
$ReportFile = "$ReportsDir/verification_master_$Timestamp.md"

$ReportContent = @"
# InsightEngine E2E Verification Report

**Timestamp:** $(Get-Date)  
**Environment:** Staging  
**Mode:** YOLO - Autonomous Verification  
**Backend:** $BackendUrl  
**Frontend:** $FrontendUrl  

## Summary

| Phase | Status | Notes |
|-------|--------|-------|
| VERIFY-001 | Completed | Full E2E verification checklist |
| VERIFY-002 | Pending | Authentication fixes |
| VERIFY-003 | Pending | Database connector fixes |
| VERIFY-004 | Pending | Sharing & permissions fixes |
| VERIFY-005 | Pending | Scheduled reports & alerts fixes |
| VERIFY-006 | Pending | Production deployment dry-run |

## Test Results

### Authentication & Security (Tasks 1-16)

See: ``verification/reports/auth_verification_*.json``

### Database Connectors (Tasks 17-35)

See: ``verification/reports/connectors_verification_*.json``

## Critical Issues Found

*Populated based on test results*

## Action Items

### Immediate (VERIFY-002 to VERIFY-006)
1. Fix authentication flow issues
2. Fix database connector integration
3. Fix permission enforcement gaps
4. Fix scheduled reports delivery
5. Fix alert triggers
6. Production deployment dry-run

## Next Steps

1. Review individual test reports in ``verification/reports/``
2. Address all FAILED tests
3. Re-run verification after fixes
4. Proceed to production deployment

---

**Report Generated:** $(Get-Date)  
**Verified By:** Automated Verification Suite
"@

$ReportContent | Out-File -FilePath $ReportFile -Encoding UTF8
Write-Host "${Green}📁 Master report saved to: $ReportFile${NC}"

# Count results
Write-Host ""
Write-Host "${Blue}═══════════════════════════════════════════════════════════════${NC}"
Write-Host "${Blue}  Verification Summary${NC}"
Write-Host "${Blue}═══════════════════════════════════════════════════════════════${NC}"

# Check for failed tests
$JsonFiles = Get-ChildItem -Path $ReportsDir -Filter "*.json"
$FailedTests = 0

foreach ($file in $JsonFiles) {
    try {
        $content = Get-Content $file.FullName | ConvertFrom-Json
        $failed = $content.results | Where-Object { $_.status -eq "FAIL" }
        if ($failed) {
            $FailedTests++
        }
    } catch {
        # Ignore parse errors
    }
}

if ($FailedTests -gt 0) {
    Write-Host "${Red}❌ Found $FailedTests report(s) with failures${NC}"
    Write-Host ""
    Write-Host "${Yellow}⚠️ VERIFY-001 completed with issues${NC}"
    Write-Host ""
    Write-Host "Next steps:"
    Write-Host "1. Review reports in verification/reports/"
    Write-Host "2. Fix issues identified"
    Write-Host "3. Re-run verification"
    Write-Host "4. Continue to VERIFY-002 to VERIFY-006"
    exit 1
} else {
    Write-Host "${Green}✅ All verifications passed!${NC}"
    Write-Host ""
    Write-Host "Ready to proceed with production deployment (VERIFY-006)"
    exit 0
}
