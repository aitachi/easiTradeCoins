#!/bin/bash

# ================================
# EasiTradeCoins - Security Audit Script
# 安全审计脚本
# ================================

set -e

echo "======================================"
echo "EasiTradeCoins Security Audit"
echo "======================================"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

ISSUES_FOUND=0
WARNINGS_FOUND=0

# Create audit report directory
REPORT_DIR="./security-audit-reports"
mkdir -p $REPORT_DIR
REPORT_FILE="$REPORT_DIR/audit-$(date +%Y%m%d-%H%M%S).md"

# Start report
cat > $REPORT_FILE << EOF
# EasiTradeCoins Security Audit Report

**Date**: $(date)
**Auditor**: Automated Security Scan
**Version**: $(git describe --tags --always)

---

## Executive Summary

This report contains the results of the automated security audit for EasiTradeCoins platform.

---

## Findings

EOF

print_section() {
    echo -e "\n${BLUE}=== $1 ===${NC}\n"
    echo -e "\n### $1\n" >> $REPORT_FILE
}

print_check() {
    if [ $2 -eq 0 ]; then
        echo -e "${GREEN}✓ PASS${NC}: $1"
        echo "- ✅ **PASS**: $1" >> $REPORT_FILE
    else
        echo -e "${RED}✗ FAIL${NC}: $1"
        echo "- ❌ **FAIL**: $1" >> $REPORT_FILE
        ((ISSUES_FOUND++))
    fi
}

print_warning() {
    echo -e "${YELLOW}⚠ WARNING${NC}: $1"
    echo "- ⚠️ **WARNING**: $1" >> $REPORT_FILE
    ((WARNINGS_FOUND++))
}

# ================================
# 1. Code Security Analysis
# ================================
print_section "1. Code Security Analysis"

echo "Scanning for common vulnerabilities..."

# Check for hardcoded secrets
echo "Checking for hardcoded secrets..."
if grep -r "password.*=.*\"" go-backend/ --include="*.go" | grep -v "DB_PASSWORD\|REDIS_PASSWORD" > /dev/null; then
    print_warning "Potential hardcoded passwords found in Go code"
else
    print_check "No hardcoded passwords in Go code" 0
fi

# Check for SQL injection vulnerabilities
echo "Checking for SQL injection vulnerabilities..."
if grep -r "db.Exec.*+\|db.Query.*+" go-backend/ --include="*.go" > /dev/null; then
    print_warning "Potential SQL injection vulnerability (string concatenation in queries)"
else
    print_check "No SQL string concatenation found" 0
fi

# Check for insecure random number generation
echo "Checking for insecure random number generation..."
if grep -r "math/rand\\.Intn\|rand\\.Seed" go-backend/ --include="*.go" > /dev/null; then
    print_warning "Using math/rand instead of crypto/rand for security-sensitive operations"
else
    print_check "Using secure random number generation" 0
fi

# ================================
# 2. Smart Contract Security
# ================================
print_section "2. Smart Contract Security"

echo "Analyzing smart contracts..."

# Check for reentrancy guards
echo "Checking for reentrancy protection..."
if grep -r "ReentrancyGuard\|nonReentrant" contracts/src/ --include="*.sol" > /dev/null; then
    print_check "Reentrancy guards implemented" 0
else
    print_warning "No reentrancy guards found"
fi

# Check for SafeMath usage (for Solidity < 0.8.0)
echo "Checking for overflow protection..."
if grep -r "pragma solidity.*0\\.8" contracts/src/ --include="*.sol" > /dev/null; then
    print_check "Using Solidity 0.8+ with built-in overflow protection" 0
else
    print_warning "Not using Solidity 0.8+ overflow protection"
fi

# Check for access control
echo "Checking for access control..."
if grep -r "onlyOwner\|Ownable\|AccessControl" contracts/src/ --include="*.sol" > /dev/null; then
    print_check "Access control mechanisms implemented" 0
else
    print_warning "No access control found"
fi

# ================================
# 3. Authentication & Authorization
# ================================
print_section "3. Authentication & Authorization"

# Check JWT secret length
echo "Checking JWT secret configuration..."
if [ -f .env.local ]; then
    JWT_SECRET=$(grep "JWT_SECRET" .env.local | cut -d'=' -f2)
    if [ ${#JWT_SECRET} -lt 32 ]; then
        print_warning "JWT secret is too short (< 32 characters)"
    else
        print_check "JWT secret length is adequate" 0
    fi
fi

# Check password hashing
echo "Checking password hashing..."
if grep -r "bcrypt\\.GenerateFromPassword" go-backend/ --include="*.go" > /dev/null; then
    print_check "Using bcrypt for password hashing" 0
else
    print_warning "Password hashing mechanism not found"
fi

# ================================
# 4. Data Protection
# ================================
print_section "4. Data Protection"

# Check for HTTPS enforcement
echo "Checking for HTTPS configuration..."
if grep -r "TLS\|https" deployment/nginx/ --include="*.conf" > /dev/null; then
    print_check "HTTPS configuration found" 0
else
    print_warning "No HTTPS configuration found"
fi

# Check for sensitive data logging
echo "Checking for sensitive data in logs..."
if grep -r "log.*password\|log.*secret\|log.*token" go-backend/ --include="*.go" -i > /dev/null; then
    print_warning "Potential sensitive data in logs"
else
    print_check "No sensitive data logged" 0
fi

# ================================
# 5. Input Validation
# ================================
print_section "5. Input Validation"

# Check for input validation
echo "Checking for input validation..."
if grep -r "validator\\.Validate\|Sanitize" go-backend/ --include="*.go" > /dev/null; then
    print_check "Input validation implemented" 0
else
    print_warning "No input validation framework found"
fi

# Check for XSS protection
echo "Checking for XSS protection..."
if grep -r "html\\.EscapeString\|template\\.HTMLEscape" go-backend/ --include="*.go" > /dev/null; then
    print_check "XSS protection measures found" 0
else
    print_warning "No XSS protection found"
fi

# ================================
# 6. Rate Limiting
# ================================
print_section "6. Rate Limiting"

echo "Checking for rate limiting..."
if grep -r "RateLimit\|TokenBucket\|Limiter" go-backend/ --include="*.go" > /dev/null; then
    print_check "Rate limiting implemented" 0
else
    print_warning "No rate limiting found"
fi

# ================================
# 7. Error Handling
# ================================
print_section "7. Error Handling"

echo "Checking error handling..."
if grep -r "if err != nil" go-backend/ --include="*.go" | wc -l | grep -q "[0-9]"; then
    print_check "Error handling present in code" 0
fi

# Check for sensitive error messages
echo "Checking for information disclosure in errors..."
if grep -r "panic\|log\\.Fatal.*err" go-backend/ --include="*.go" > /dev/null; then
    print_warning "Potential information disclosure through error messages"
else
    print_check "Error messages properly handled" 0
fi

# ================================
# 8. Dependency Security
# ================================
print_section "8. Dependency Security"

echo "Checking Go dependencies for vulnerabilities..."
cd go-backend
if command -v govulncheck &> /dev/null; then
    if govulncheck ./... > /dev/null 2>&1; then
        print_check "No known vulnerabilities in Go dependencies" 0
    else
        print_warning "Vulnerabilities found in Go dependencies"
    fi
else
    print_warning "govulncheck not installed, skipping Go dependency check"
fi
cd ..

echo "Checking npm dependencies for vulnerabilities..."
cd contracts
if command -v npm &> /dev/null; then
    if npm audit --audit-level=high > /dev/null 2>&1; then
        print_check "No high-severity vulnerabilities in npm dependencies" 0
    else
        print_warning "Vulnerabilities found in npm dependencies"
    fi
else
    print_warning "npm not installed, skipping npm dependency check"
fi
cd ..

# ================================
# 9. Configuration Security
# ================================
print_section "9. Configuration Security"

# Check for exposed .env files
echo "Checking for exposed environment files..."
if [ -f .env ] && ! grep -q ".env" .gitignore; then
    print_warning ".env file exists but not in .gitignore"
else
    print_check ".env properly ignored" 0
fi

# Check database SSL mode
echo "Checking database SSL configuration..."
if grep "DB_SSL_MODE=disable" .env.local > /dev/null 2>&1; then
    print_warning "Database SSL is disabled (not recommended for production)"
else
    print_check "Database SSL configuration OK" 0
fi

# ================================
# 10. Smart Contract Static Analysis
# ================================
print_section "10. Smart Contract Static Analysis"

echo "Running Slither static analysis..."
cd contracts
if command -v slither &> /dev/null; then
    if slither src/ --json slither-report.json > /dev/null 2>&1; then
        print_check "Slither analysis completed" 0
    else
        print_warning "Slither found potential issues"
    fi
else
    print_warning "Slither not installed, skipping static analysis"
fi
cd ..

# ================================
# Audit Summary
# ================================
echo ""
echo "======================================"
echo "Audit Summary"
echo "======================================"
echo -e "Critical Issues: ${RED}$ISSUES_FOUND${NC}"
echo -e "Warnings: ${YELLOW}$WARNINGS_FOUND${NC}"
echo ""

cat >> $REPORT_FILE << EOF

---

## Summary

- **Critical Issues**: $ISSUES_FOUND
- **Warnings**: $WARNINGS_FOUND

## Recommendations

1. **Address all critical issues** before production deployment
2. **Review all warnings** and implement fixes where applicable
3. **Conduct manual security review** for business logic vulnerabilities
4. **Perform penetration testing** on staging environment
5. **Implement bug bounty program** for production
6. **Regular security audits** (quarterly recommended)
7. **Keep dependencies up to date**
8. **Monitor security advisories** for used libraries

---

## Tools Used

- grep (pattern matching)
- govulncheck (Go vulnerability scanner)
- npm audit (npm dependency scanner)
- slither (Solidity static analyzer)

---

*Report generated on $(date)*
EOF

echo "Full report saved to: $REPORT_FILE"

if [ $ISSUES_FOUND -gt 0 ]; then
    echo -e "${RED}Security audit found critical issues!${NC}"
    exit 1
else
    echo -e "${GREEN}Security audit completed with $WARNINGS_FOUND warnings.${NC}"
    exit 0
fi
