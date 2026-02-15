#!/bin/bash
# verification/scripts/run_verification.sh
# Master verification script - VERIFY-001 to VERIFY-006

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}"
echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║     INSIGHTENGINE E2E VERIFICATION SUITE - YOLO MODE            ║"
echo "║     VERIFY-001 to VERIFY-006 - Production Readiness            ║"
echo "╚══════════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Configuration
BACKEND_URL="http://localhost:8080"
FRONTEND_URL="http://localhost:3000"
REPORTS_DIR="verification/reports"
FIXES_DIR="verification/fixes"

# Create directories
mkdir -p "$REPORTS_DIR"
mkdir -p "$FIXES_DIR"

# Check if backend is running
echo -e "${BLUE}🔍 Checking backend health...${NC}"
if ! curl -s "$BACKEND_URL/api/health" > /dev/null 2>&1; then
    echo -e "${RED}❌ Backend not running at $BACKEND_URL${NC}"
    echo "Please start the backend first:"
    echo "  cd backend && go run main.go"
    exit 1
fi
echo -e "${GREEN}✅ Backend is running${NC}"

# Check if frontend is running
echo -e "${BLUE}🔍 Checking frontend health...${NC}"
if ! curl -s "$FRONTEND_URL" > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  Frontend not running at $FRONTEND_URL${NC}"
    echo "Some tests may be skipped"
else
    echo -e "${GREEN}✅ Frontend is running${NC}"
fi

# Start verification
echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  VERIFY-001: Execute Full E2E Verification Checklist${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo ""

# Run authentication verification
echo -e "${YELLOW}🧪 Running Authentication & Security Tests...${NC}"
cd verification/scripts
go run verify_auth.go
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Authentication tests passed${NC}"
else
    echo -e "${RED}❌ Authentication tests failed${NC}"
fi
cd ../..

# Run connector verification
echo ""
echo -e "${YELLOW}🧪 Running Database Connector Tests...${NC}"
cd verification/scripts
go run verify_connectors.go
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Connector tests passed${NC}"
else
    echo -e "${RED}❌ Connector tests failed${NC}"
fi
cd ../..

# Generate consolidated report
echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  Generating Consolidated Report${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
REPORT_FILE="$REPORTS_DIR/verification_master_$TIMESTAMP.md"

cat > "$REPORT_FILE" << EOF
# InsightEngine E2E Verification Report

**Timestamp:** $(date)  
**Environment:** Staging  
**Mode:** YOLO - Autonomous Verification  

## Summary

| Phase | Status | Notes |
|-------|--------|-------|
| VERIFY-001 | In Progress | Full E2E verification checklist |
| VERIFY-002 | Pending | Authentication fixes |
| VERIFY-003 | Pending | Database connector fixes |
| VERIFY-004 | Pending | Sharing & permissions fixes |
| VERIFY-005 | Pending | Scheduled reports & alerts fixes |
| VERIFY-006 | Pending | Production deployment dry-run |

## Test Results

### Authentication & Security (Tasks 1-16)

See: \`verification/reports/auth_verification_*.json\`

### Database Connectors (Tasks 17-35)

See: \`verification/reports/connectors_verification_*.json\`

## Critical Issues Found

*To be populated after full verification*

## Action Items

### Immediate (VERIFY-002 to VERIFY-006)
1. Fix authentication flow issues
2. Fix database connector integration
3. Fix permission enforcement gaps
4. Fix scheduled reports delivery
5. Fix alert triggers
6. Production deployment dry-run

## Next Steps

1. Review individual test reports in \`verification/reports/\`
2. Address all FAILED tests
3. Re-run verification after fixes
4. Proceed to production deployment

---

**Report Generated:** $(date)  
**Verified By:** Automated Verification Suite  
EOF

echo -e "${GREEN}📁 Master report saved to: $REPORT_FILE${NC}"

# Count results
echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  Verification Summary${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"

# Check for any failed tests
FAILED_TESTS=$(find "$REPORTS_DIR" -name "*.json" -exec grep -l '"status": "FAIL"' {} \; | wc -l)

if [ "$FAILED_TESTS" -gt 0 ]; then
    echo -e "${RED}❌ Found $FAILED_TESTS report(s) with failures${NC}"
    echo ""
    echo -e "${YELLOW}⚠️  VERIFY-001 completed with issues${NC}"
    echo ""
    echo "Next steps:"
    echo "1. Review reports in verification/reports/"
    echo "2. Fix issues identified"
    echo "3. Re-run verification"
    echo "4. Continue to VERIFY-002 to VERIFY-006"
    exit 1
else
    echo -e "${GREEN}✅ All verifications passed!${NC}"
    echo ""
    echo "Ready to proceed with production deployment (VERIFY-006)"
    exit 0
fi
