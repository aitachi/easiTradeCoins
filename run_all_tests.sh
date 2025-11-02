#!/bin/bash

# ================================
# EasiTradeCoins - Master Test Runner
# 全面测试运行脚本
# ================================

set -e

echo "======================================"
echo "EasiTradeCoins - Master Test Runner"
echo "======================================"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Test results
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Report directory
REPORT_DIR="./test-reports"
mkdir -p $REPORT_DIR
MASTER_REPORT="$REPORT_DIR/master-report-$(date +%Y%m%d-%H%M%S).md"

# Start master report
cat > $MASTER_REPORT << EOF
# EasiTradeCoins - Master Test Report

**Date**: $(date)
**Version**: $(git describe --tags --always 2>/dev/null || echo "unknown")
**Branch**: $(git branch --show-current 2>/dev/null || echo "unknown")
**Commit**: $(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

---

## Test Execution Summary

EOF

print_section() {
    echo -e "\n${BLUE}======================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}======================================${NC}\n"
    echo -e "\n## $1\n" >> $MASTER_REPORT
}

run_test_suite() {
    local name=$1
    local command=$2

    echo -e "${YELLOW}Running: $name${NC}"
    ((TOTAL_TESTS++))

    if eval $command; then
        echo -e "${GREEN}✓ PASSED${NC}: $name\n"
        echo "- ✅ **PASSED**: $name" >> $MASTER_REPORT
        ((PASSED_TESTS++))
    else
        echo -e "${RED}✗ FAILED${NC}: $name\n"
        echo "- ❌ **FAILED**: $name" >> $MASTER_REPORT
        ((FAILED_TESTS++))
    fi
}

# ================================
# 1. Unit Tests
# ================================
print_section "1. Unit Tests"

echo "Running Go backend unit tests..."
run_test_suite "Go Backend Services Tests" \
    "cd go-backend && go test -v ./internal/services/... -coverprofile=coverage.out 2>&1 | tee $REPORT_DIR/go-unit-tests.log"

echo "Running Smart Contract tests..."
run_test_suite "Smart Contract Tests" \
    "cd contracts && npx hardhat test 2>&1 | tee $REPORT_DIR/contract-tests.log"

# ================================
# 2. Integration Tests
# ================================
print_section "2. Integration Tests"

echo "Running integration tests..."
run_test_suite "Integration Tests" \
    "bash ./tests/integration_test.sh 2>&1 | tee $REPORT_DIR/integration-tests.log"

# ================================
# 3. API Tests
# ================================
print_section "3. API Tests"

echo "Starting backend for API tests..."
if ! pgrep -f "easitrade-backend" > /dev/null; then
    cd go-backend && go run cmd/server/main.go &
    BACKEND_PID=$!
    echo "Backend started with PID: $BACKEND_PID"
    sleep 5
    cd ..
fi

echo "Testing API endpoints..."
run_test_suite "API Health Check" \
    "curl -f http://localhost:8080/health"

run_test_suite "API Swagger Documentation" \
    "curl -f http://localhost:8080/swagger/index.html"

run_test_suite "API Metrics Endpoint" \
    "curl -f http://localhost:8081/metrics"

# ================================
# 4. Database Tests
# ================================
print_section "4. Database Tests"

echo "Testing database connections..."
run_test_suite "PostgreSQL Connection" \
    "psql -h localhost -U socialfi -d socialfi -c 'SELECT 1;' > /dev/null 2>&1"

run_test_suite "MySQL Connection" \
    "mysql -h localhost -u root -e 'SELECT 1;' > /dev/null 2>&1"

run_test_suite "Redis Connection" \
    "redis-cli -h localhost ping | grep -q PONG"

# ================================
# 5. Performance Tests
# ================================
print_section "5. Performance Tests"

echo "Running performance tests..."
run_test_suite "Performance Test Suite" \
    "bash ./tests/performance_test.sh 2>&1 | tee $REPORT_DIR/performance-tests.log"

# ================================
# 6. Security Audit
# ================================
print_section "6. Security Audit"

echo "Running security audit..."
run_test_suite "Security Audit" \
    "bash ./tests/security_audit.sh 2>&1 | tee $REPORT_DIR/security-audit.log"

# ================================
# 7. Contract Deployment Tests (Sepolia)
# ================================
print_section "7. Sepolia Testnet Deployment"

if [ -f .env ] && grep -q "ETHEREUM_RPC_URL" .env; then
    echo "Deploying to Sepolia testnet..."
    run_test_suite "Sepolia Deployment" \
        "cd contracts && npx hardhat run scripts/deploy-sepolia.js --network sepolia 2>&1 | tee $REPORT_DIR/sepolia-deployment.log"
else
    echo -e "${YELLOW}⚠ Skipping Sepolia deployment (no RPC configured)${NC}"
    echo "- ⚠️ **SKIPPED**: Sepolia Deployment (no RPC configured)" >> $MASTER_REPORT
fi

# ================================
# 8. Code Coverage
# ================================
print_section "8. Code Coverage Analysis"

echo "Generating Go coverage report..."
if [ -f go-backend/coverage.out ]; then
    cd go-backend
    go tool cover -html=coverage.out -o $REPORT_DIR/go-coverage.html
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
    echo "Go Code Coverage: $COVERAGE"
    echo "- **Go Coverage**: $COVERAGE" >> ../$MASTER_REPORT
    cd ..
fi

echo "Generating Solidity coverage report..."
run_test_suite "Solidity Coverage" \
    "cd contracts && npx hardhat coverage 2>&1 | tee $REPORT_DIR/solidity-coverage.log"

# ================================
# 9. Linting and Code Quality
# ================================
print_section "9. Code Quality Checks"

echo "Running Go linter..."
if command -v golangci-lint &> /dev/null; then
    run_test_suite "Go Lint" \
        "cd go-backend && golangci-lint run ./... 2>&1 | tee $REPORT_DIR/go-lint.log"
else
    echo -e "${YELLOW}⚠ golangci-lint not installed${NC}"
    echo "- ⚠️ **SKIPPED**: Go Lint (not installed)" >> $MASTER_REPORT
fi

echo "Running Solidity linter..."
run_test_suite "Solidity Lint" \
    "cd contracts && npx solhint 'src/**/*.sol' 2>&1 | tee $REPORT_DIR/solidity-lint.log"

# ================================
# 10. Build Tests
# ================================
print_section "10. Build Tests"

echo "Testing Go build..."
run_test_suite "Go Build" \
    "cd go-backend && go build -o bin/server cmd/server/main.go"

echo "Testing contract compilation..."
run_test_suite "Contract Compilation" \
    "cd contracts && npx hardhat compile"

# ================================
# Cleanup
# ================================
if [ ! -z "$BACKEND_PID" ]; then
    echo "Stopping backend server..."
    kill $BACKEND_PID 2>/dev/null || true
fi

# ================================
# Final Report
# ================================
echo ""
echo "======================================"
echo "Test Execution Complete"
echo "======================================"
echo -e "Total Test Suites: $TOTAL_TESTS"
echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"
echo -e "${RED}Failed: $FAILED_TESTS${NC}"
echo ""

# Add summary to master report
cat >> $MASTER_REPORT << EOF

---

## Overall Results

| Metric | Value |
|--------|-------|
| Total Test Suites | $TOTAL_TESTS |
| Passed | $PASSED_TESTS |
| Failed | $FAILED_TESTS |
| Success Rate | $(echo "scale=2; $PASSED_TESTS * 100 / $TOTAL_TESTS" | bc)% |

## Test Reports

All detailed test reports are available in the \`$REPORT_DIR\` directory:

- \`go-unit-tests.log\` - Go backend unit tests
- \`contract-tests.log\` - Smart contract tests
- \`integration-tests.log\` - Integration tests
- \`performance-tests.log\` - Performance benchmarks
- \`security-audit.log\` - Security audit results
- \`sepolia-deployment.log\` - Sepolia testnet deployment
- \`go-coverage.html\` - Go code coverage visualization
- \`solidity-coverage.log\` - Solidity coverage report

---

## Recommendations

### If tests passed (${PASSED_TESTS}/${TOTAL_TESTS}):
1. ✅ Review test coverage reports
2. ✅ Check performance benchmarks meet requirements
3. ✅ Address any security warnings
4. ✅ Proceed with deployment preparation

### If tests failed:
1. ❌ Review failed test logs in detail
2. ❌ Fix identified issues
3. ❌ Re-run test suite
4. ❌ Do NOT deploy until all tests pass

---

## Next Steps

1. **Review all test reports** in \`$REPORT_DIR\`
2. **Address any failures or warnings**
3. **Update documentation** with test results
4. **Prepare deployment** if all tests pass
5. **Schedule regular testing** (CI/CD pipeline)

---

*Master report generated on $(date)*
EOF

echo "Master report saved to: $MASTER_REPORT"
echo "All test logs saved to: $REPORT_DIR"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "\n${GREEN}========================================${NC}"
    echo -e "${GREEN}All tests passed! Ready for deployment ✓${NC}"
    echo -e "${GREEN}========================================${NC}\n"
    exit 0
else
    echo -e "\n${RED}========================================${NC}"
    echo -e "${RED}Some tests failed! Review logs ✗${NC}"
    echo -e "${RED}========================================${NC}\n"
    exit 1
fi
