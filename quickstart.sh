#!/bin/bash

# Quick Start Script for EasiTradeCoins
# This script helps you get started quickly

set -e

echo "=========================================="
echo "  EasiTradeCoins - Quick Start"
echo "=========================================="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if .env exists
if [ ! -f .env ]; then
    echo -e "${YELLOW}Creating .env file from .env.example...${NC}"
    cp .env.example .env
    echo -e "${GREEN}✓ .env file created${NC}"
    echo -e "${YELLOW}⚠ Please edit .env file with your configuration${NC}"
    echo ""
fi

# Choose installation mode
echo "Choose installation mode:"
echo "1) Docker (Recommended - Full stack with one command)"
echo "2) Local Development (Requires PostgreSQL and Redis)"
echo "3) Contracts Only (Deploy smart contracts only)"
echo ""
read -p "Enter your choice (1-3): " choice

case $choice in
    1)
        echo ""
        echo "=========================================="
        echo "  Docker Installation"
        echo "=========================================="
        echo ""

        # Check Docker
        if ! command -v docker &> /dev/null; then
            echo -e "${RED}✗ Docker not found. Please install Docker first.${NC}"
            echo "  Visit: https://docs.docker.com/get-docker/"
            exit 1
        fi

        if ! command -v docker-compose &> /dev/null; then
            echo -e "${RED}✗ Docker Compose not found. Please install Docker Compose first.${NC}"
            exit 1
        fi

        echo -e "${GREEN}✓ Docker is installed${NC}"
        echo ""

        # Build and start services
        echo "Building and starting services..."
        docker-compose down -v
        docker-compose up -d --build

        echo ""
        echo "Waiting for services to start (30 seconds)..."
        sleep 30

        echo ""
        echo -e "${GREEN}=========================================="
        echo "  Installation Complete! 🎉"
        echo "==========================================${NC}"
        echo ""
        echo "Services:"
        docker-compose ps
        echo ""
        echo "API Endpoints:"
        echo "  - Health Check: http://localhost:8080/health"
        echo "  - API Base URL: http://localhost:8080/api/v1"
        echo "  - WebSocket: ws://localhost:8080/ws"
        echo ""
        echo "Next Steps:"
        echo "  1. Test API: curl http://localhost:8080/health"
        echo "  2. View logs: docker-compose logs -f backend"
        echo "  3. Stop services: docker-compose down"
        echo ""
        ;;

    2)
        echo ""
        echo "=========================================="
        echo "  Local Development Installation"
        echo "=========================================="
        echo ""

        # Check dependencies
        echo "Checking dependencies..."

        if ! command -v go &> /dev/null; then
            echo -e "${RED}✗ Go not found. Please install Go 1.21+${NC}"
            echo "  Visit: https://golang.org/dl/"
            exit 1
        fi
        echo -e "${GREEN}✓ Go is installed${NC}"

        if ! command -v psql &> /dev/null; then
            echo -e "${YELLOW}⚠ PostgreSQL client not found${NC}"
            echo "  Please ensure PostgreSQL is installed and running"
        fi

        # Start databases with Docker
        echo ""
        echo "Starting databases (PostgreSQL + Redis)..."
        docker-compose up -d postgres redis

        echo "Waiting for databases (10 seconds)..."
        sleep 10

        # Initialize database
        echo ""
        echo "Initializing database..."
        if command -v psql &> /dev/null; then
            psql "${DATABASE_URL:-postgresql://postgres:postgres@localhost:5432/easitradecoins}" -f deployment/init.sql || true
            echo -e "${GREEN}✓ Database initialized${NC}"
        else
            echo -e "${YELLOW}⚠ Please run manually: psql \$DATABASE_URL -f deployment/init.sql${NC}"
        fi

        # Install Go dependencies
        echo ""
        echo "Installing Go dependencies..."
        cd go-backend
        go mod download
        echo -e "${GREEN}✓ Dependencies installed${NC}"

        echo ""
        echo -e "${GREEN}=========================================="
        echo "  Setup Complete! 🎉"
        echo "==========================================${NC}"
        echo ""
        echo "To start the backend server:"
        echo "  cd go-backend"
        echo "  go run cmd/server/main.go"
        echo ""
        echo "Or use:"
        echo "  make deploy-dev"
        echo ""
        ;;

    3)
        echo ""
        echo "=========================================="
        echo "  Smart Contracts Deployment"
        echo "=========================================="
        echo ""

        # Check Foundry
        if ! command -v forge &> /dev/null; then
            echo -e "${RED}✗ Foundry not found${NC}"
            echo "  Install with: curl -L https://foundry.paradigm.xyz | bash"
            exit 1
        fi
        echo -e "${GREEN}✓ Foundry is installed${NC}"

        # Check environment
        if [ -z "$PRIVATE_KEY" ]; then
            echo -e "${YELLOW}⚠ PRIVATE_KEY not set in environment${NC}"
            echo "  Please set PRIVATE_KEY in .env file"
            exit 1
        fi

        if [ -z "$SEPOLIA_RPC_URL" ]; then
            echo -e "${YELLOW}⚠ SEPOLIA_RPC_URL not set${NC}"
            echo "  Please set SEPOLIA_RPC_URL in .env file"
            exit 1
        fi

        # Load .env
        export $(cat .env | grep -v '#' | awk '/=/ {print $1}')

        # Install dependencies
        echo ""
        echo "Installing OpenZeppelin contracts..."
        cd contracts
        forge install OpenZeppelin/openzeppelin-contracts --no-git || true

        # Build contracts
        echo ""
        echo "Building contracts..."
        forge build

        # Deploy
        echo ""
        read -p "Deploy to Sepolia testnet? (y/n): " deploy_choice
        if [ "$deploy_choice" = "y" ]; then
            echo "Deploying contracts..."
            forge script script/Deploy.s.sol:DeployAll \
                --rpc-url $SEPOLIA_RPC_URL \
                --private-key $PRIVATE_KEY \
                --broadcast

            echo ""
            echo -e "${GREEN}✓ Contracts deployed!${NC}"
        fi

        cd ..
        echo ""
        echo -e "${GREEN}=========================================="
        echo "  Contracts Setup Complete! 🎉"
        echo "==========================================${NC}"
        ;;

    *)
        echo -e "${RED}Invalid choice${NC}"
        exit 1
        ;;
esac

echo ""
echo "For more information, see README.md"
echo ""
