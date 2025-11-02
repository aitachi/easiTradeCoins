#!/bin/bash

# ================================
# EasiTradeCoins - Integration Test Script
# 集成测试脚本
# ================================

set -e

echo "======================================"
echo "EasiTradeCoins Integration Tests"
echo "======================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test results
TESTS_PASSED=0
TESTS_FAILED=0

# Function to print test result
print_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ PASSED${NC}: $2"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗ FAILED${NC}: $2"
        ((TESTS_FAILED++))
    fi
}

# Function to run test
run_test() {
    echo -e "${YELLOW}Running:${NC} $1"
    eval $2
    print_result $? "$1"
    echo ""
}

# ================================
# 1. Database Connection Tests
# ================================
echo "1. Testing Database Connections..."
echo "===================================="

# PostgreSQL
run_test "PostgreSQL Connection" \
    "psql -h localhost -U socialfi -d socialfi -p 5432 -c 'SELECT 1;' > /dev/null 2>&1"

# MySQL
run_test "MySQL Connection" \
    "mysql -h localhost -u root -e 'SELECT 1;' > /dev/null 2>&1"

# Redis
run_test "Redis Connection" \
    "redis-cli -h localhost -p 6379 ping | grep -q PONG"

# ================================
# 2. Message Queue Tests
# ================================
echo "2. Testing Message Queue..."
echo "===================================="

# Kafka
run_test "Kafka Broker Connectivity" \
    "kafka-broker-api-versions --bootstrap-server localhost:9092 > /dev/null 2>&1"

# ================================
# 3. Search Engine Tests
# ================================
echo "3. Testing Search Engine..."
echo "===================================="

# Elasticsearch
run_test "Elasticsearch Health Check" \
    "curl -s http://localhost:9200/_cluster/health | grep -q '\"status\":\"green\\|yellow\"'"

# ================================
# 4. API Endpoint Tests
# ================================
echo "4. Testing API Endpoints..."
echo "===================================="

# Start backend if not running
if ! pgrep -f "easitrade-backend" > /dev/null; then
    echo "Starting backend server..."
    cd go-backend && go run cmd/server/main.go &
    BACKEND_PID=$!
    sleep 5
fi

# Health check
run_test "API Health Endpoint" \
    "curl -s http://localhost:8080/health | grep -q 'ok\\|healthy'"

# Swagger docs
run_test "Swagger Documentation" \
    "curl -s http://localhost:8080/swagger/index.html | grep -q 'swagger'"

# Metrics endpoint
run_test "Prometheus Metrics" \
    "curl -s http://localhost:8081/metrics | grep -q 'go_'"

# ================================
# 5. WebSocket Tests
# ================================
echo "5. Testing WebSocket..."
echo "===================================="

run_test "WebSocket Connection" \
    "node -e \"const ws = require('ws'); const socket = new ws('ws://localhost:8080/ws'); socket.on('open', () => { console.log('connected'); process.exit(0); }); socket.on('error', () => process.exit(1));\" > /dev/null 2>&1"

# ================================
# 6. Trading Function Tests
# ================================
echo "6. Testing Trading Functions..."
echo "===================================="

# Run Go unit tests
run_test "Go Backend Unit Tests" \
    "cd go-backend && go test -v ./internal/services/... -count=1"

# ================================
# 7. Smart Contract Tests
# ================================
echo "7. Testing Smart Contracts..."
echo "===================================="

# Run Hardhat tests
run_test "Smart Contract Unit Tests" \
    "cd contracts && npx hardhat test"

# ================================
# 8. Performance Tests
# ================================
echo "8. Performance Tests..."
echo "===================================="

# API response time test
run_test "API Response Time < 100ms" \
    "curl -w '%{time_total}' -o /dev/null -s http://localhost:8080/health | awk '{exit !($1 < 0.1)}'"

# ================================
# 9. Security Tests
# ================================
echo "9. Security Tests..."
echo "===================================="

# Check for SQL injection protection
run_test "SQL Injection Protection" \
    "curl -s 'http://localhost:8080/api/v1/orders?symbol=BTC_USDT%27%20OR%201=1--' | grep -qv 'error\\|exception'"

# Check for XSS protection
run_test "XSS Protection Headers" \
    "curl -I -s http://localhost:8080 | grep -q 'X-Content-Type-Options'"

# ================================
# Test Summary
# ================================
echo ""
echo "======================================"
echo "Test Summary"
echo "======================================"
echo -e "Total Tests: $((TESTS_PASSED + TESTS_FAILED))"
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
echo -e "${RED}Failed: $TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed! ✓${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed! ✗${NC}"
    exit 1
fi
