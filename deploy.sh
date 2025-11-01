#!/bin/bash

# EasiTradeCoins Deployment Script

set -e

echo "=========================================="
echo "EasiTradeCoins Deployment Script"
echo "=========================================="

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '#' | awk '/=/ {print $1}')
fi

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check dependencies
echo "Checking dependencies..."

if ! command_exists forge; then
    echo "Error: Foundry not installed. Please install from https://getfoundry.sh"
    exit 1
fi

if ! command_exists go; then
    echo "Error: Go not installed. Please install from https://golang.org"
    exit 1
fi

if ! command_exists docker; then
    echo "Error: Docker not installed. Please install from https://docker.com"
    exit 1
fi

# Build smart contracts
echo ""
echo "=========================================="
echo "Building Smart Contracts"
echo "=========================================="

cd contracts
forge build

if [ "$1" == "deploy-contracts" ]; then
    echo "Deploying contracts to ${NETWORK}..."

    if [ "$NETWORK" == "sepolia" ]; then
        RPC_URL=$SEPOLIA_RPC_URL
    else
        RPC_URL=$MAINNET_RPC_URL
    fi

    forge script script/Deploy.s.sol:DeployAll \
        --rpc-url $RPC_URL \
        --private-key $PRIVATE_KEY \
        --broadcast \
        --verify
fi

cd ..

# Build Go backend
echo ""
echo "=========================================="
echo "Building Go Backend"
echo "=========================================="

cd go-backend
go mod tidy
go build -o server ./cmd/server
cd ..

# Run tests
if [ "$1" == "test" ]; then
    echo ""
    echo "=========================================="
    echo "Running Tests"
    echo "=========================================="

    # Test smart contracts
    cd contracts
    forge test -vv
    cd ..

    # Test Go backend
    cd go-backend
    go test ./... -v
    cd ..
fi

# Deploy with Docker
if [ "$1" == "deploy-docker" ]; then
    echo ""
    echo "=========================================="
    echo "Deploying with Docker"
    echo "=========================================="

    docker-compose down
    docker-compose build
    docker-compose up -d

    echo ""
    echo "Waiting for services to start..."
    sleep 10

    echo "Services status:"
    docker-compose ps

    echo ""
    echo "Backend logs:"
    docker-compose logs backend | tail -20
fi

# Start local development
if [ "$1" == "dev" ]; then
    echo ""
    echo "=========================================="
    echo "Starting Local Development"
    echo "=========================================="

    # Start databases
    docker-compose up -d postgres redis

    echo "Waiting for databases..."
    sleep 5

    # Run migrations
    psql $DATABASE_URL -f deployment/init.sql

    # Start backend
    cd go-backend
    go run cmd/server/main.go
fi

echo ""
echo "=========================================="
echo "Deployment Complete!"
echo "=========================================="
echo ""
echo "API Endpoint: http://localhost:8080"
echo "WebSocket: ws://localhost:8080/ws"
echo "Health Check: http://localhost:8080/health"
echo ""
echo "Next steps:"
echo "1. Test the API: curl http://localhost:8080/health"
echo "2. Register a user: curl -X POST http://localhost:8080/api/v1/auth/register"
echo "3. View logs: docker-compose logs -f backend"
echo ""
