# Makefile for EasiTradeCoins

.PHONY: help install build test deploy clean

help:
	@echo "EasiTradeCoins - Makefile Commands"
	@echo ""
	@echo "  install        Install dependencies"
	@echo "  build          Build all components"
	@echo "  test           Run all tests"
	@echo "  deploy-dev     Deploy for local development"
	@echo "  deploy-docker  Deploy with Docker"
	@echo "  deploy-prod    Deploy to production"
	@echo "  clean          Clean build artifacts"
	@echo ""

install:
	@echo "Installing dependencies..."
	@cd contracts && forge install
	@cd go-backend && go mod download
	@echo "Dependencies installed!"

build:
	@echo "Building smart contracts..."
	@cd contracts && forge build
	@echo "Building Go backend..."
	@cd go-backend && go build -o server ./cmd/server
	@echo "Build complete!"

test:
	@echo "Testing smart contracts..."
	@cd contracts && forge test -vv
	@echo "Testing Go backend..."
	@cd go-backend && go test ./... -v
	@echo "All tests passed!"

deploy-dev:
	@echo "Deploying for local development..."
	@docker-compose up -d postgres redis
	@sleep 5
	@psql $(DATABASE_URL) -f deployment/init.sql || true
	@cd go-backend && go run cmd/server/main.go

deploy-docker:
	@echo "Deploying with Docker..."
	@docker-compose down
	@docker-compose build
	@docker-compose up -d
	@echo "Deployment complete!"
	@docker-compose ps

deploy-contracts:
	@echo "Deploying smart contracts..."
	@cd contracts && forge script script/Deploy.s.sol:DeployAll \
		--rpc-url $(SEPOLIA_RPC_URL) \
		--private-key $(PRIVATE_KEY) \
		--broadcast
	@echo "Contracts deployed!"

clean:
	@echo "Cleaning build artifacts..."
	@cd contracts && forge clean
	@cd go-backend && rm -f server
	@docker-compose down -v
	@echo "Clean complete!"

logs:
	@docker-compose logs -f backend

restart:
	@docker-compose restart backend

stop:
	@docker-compose stop

down:
	@docker-compose down
