#!/bin/bash
# verification/scripts/run_all_fixes.sh
# Master script to run VERIFY-002 through VERIFY-006

set -e

echo "╔════════════════════════════════════════════════════════════════╗"
echo "║     INSIGHTENGINE - VERIFY-002 TO VERIFY-006 RUNNER           ║"
echo "║     Production Readiness Fix Suite                            ║"
echo "╚════════════════════════════════════════════════════════════════╝"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

REPORTS_DIR="verification/reports"
mkdir -p "$REPORTS_DIR"

# Function to run a fix script
run_fix() {
    local script=$1
    local name=$2
    
    echo ""
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  Running $name${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo ""
    
    if [ -f "$script" ]; then
        cd verification/fixes
        go run $(basename "$script")
        local exit_code=$?
        cd ../..
        
        if [ $exit_code -eq 0 ]; then
            echo -e "${GREEN}✅ $name completed successfully${NC}"
        else
            echo -e "${RED}❌ $name completed with issues${NC}"
        fi
        
        return $exit_code
    else
        echo -e "${YELLOW}⚠️  Script not found: $script${NC}"
        return 1
    fi
}

# Check if we're in the right directory
if [ ! -d "verification/fixes" ]; then
    echo -e "${RED}❌ Error: Must run from project root directory${NC}"
    exit 1
fi

echo -e "${YELLOW}This will run all verification fixes (VERIFY-002 to VERIFY-006)${NC}"
echo -e "${YELLOW}Make sure your backend and database are running${NC}"
echo ""
read -p "Press Enter to continue or Ctrl+C to cancel..."

# Track results
SUCCESS=0
FAILED=0

# VERIFY-002: Authentication Fixes
echo ""
echo "🚀 Starting VERIFY-002: Authentication Flow Fixes"
if run_fix "verification/fixes/verify_002_fixes.go" "VERIFY-002"; then
    ((SUCCESS++))
else
    ((FAILED++))
fi

# VERIFY-003: Database Connector Fixes
echo ""
echo "🚀 Starting VERIFY-003: Database Connector Fixes"
if run_fix "verification/fixes/verify_003_fixes.go" "VERIFY-003"; then
    ((SUCCESS++))
else
    ((FAILED++))
fi

# VERIFY-004: Sharing & Permissions Fixes
echo ""
echo "🚀 Starting VERIFY-004: Sharing & Permissions Fixes"
if run_fix "verification/fixes/verify_004_fixes.go" "VERIFY-004"; then
    ((SUCCESS++))
else
    ((FAILED++))
fi

# VERIFY-005: Scheduled Reports & Alerts Fixes
echo ""
echo "🚀 Starting VERIFY-005: Scheduled Reports & Alerts Fixes"
if run_fix "verification/fixes/verify_005_fixes.go" "VERIFY-005"; then
    ((SUCCESS++))
else
    ((FAILED++))
fi

# VERIFY-006: Production Deployment Dry-Run
echo ""
echo "🚀 Starting VERIFY-006: Production Deployment Dry-Run"
if run_fix "verification/fixes/verify_006_production.go" "VERIFY-006"; then
    ((SUCCESS++))
else
    ((FAILED++))
fi

# Summary
echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  FINAL SUMMARY${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo ""
echo -e "Completed: $SUCCESS / 5 verification tasks"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All verifications completed successfully!${NC}"
    echo ""
    echo "Next steps:"
    echo "1. Review reports in verification/reports/"
    echo "2. Address any warnings"
    echo "3. Run full E2E tests"
    echo "4. Deploy to production"
    exit 0
else
    echo -e "${RED}❌ $FAILED verification(s) completed with issues${NC}"
    echo ""
    echo "Please review the output above and fix issues before production"
    exit 1
fi
