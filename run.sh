#!/bin/bash

# EduAccess API - Run Script
# This script runs the EduAccess API locally without Docker

set -e  # Exit on any error

echo "=========================================="
echo "EduAccess API - Starting Server"
echo "=========================================="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.25+ first."
    exit 1
fi

# Display Go version
GO_VERSION=$(go version)
echo "✓ $GO_VERSION"

# Check if .env file exists
if [ ! -f .env ]; then
    echo "⚠️  .env file not found. Copying from .env.example..."
    if [ -f .env.example ]; then
        cp .env.example .env
        echo "✓ Created .env. Please fill in your DATABASE_URL and JWT_SECRET"
        echo "  Then run this script again."
        exit 1
    else
        echo "❌ .env.example not found."
        exit 1
    fi
fi

echo "✓ .env file found"

# Load environment variables from .env
export $(cat .env | grep -v '^#' | xargs)

# Check for required environment variables
if [ -z "$JWT_SECRET" ]; then
    echo "❌ JWT_SECRET is not set in .env"
    exit 1
fi

if [ -z "$DATABASE_URL" ] && [ -z "$DB_HOST" ]; then
    echo "❌ Either DATABASE_URL or DB_HOST must be set in .env"
    exit 1
fi

echo "✓ Environment variables loaded"

# Tidy Go modules
echo ""
echo "📦 Tidying Go modules..."
go mod tidy
echo "✓ Go modules tidied"

# Generate Swagger docs
echo ""
echo "📚 Generating Swagger documentation..."
if ! command -v swag &> /dev/null; then
    echo "⚠️  swag not installed. Installing swag..."
    go install github.com/swaggo/swag/cmd/swag@latest
fi

swag init -g cmd/main.go --output docs
echo "✓ Swagger docs generated"

# Start the server
echo ""
echo "=========================================="
echo "🚀 Starting EduAccess API Server"
echo "=========================================="
echo "Server will be available at: http://localhost:${APP_PORT:-8080}"
echo "Swagger UI: http://localhost:${APP_PORT:-8080}/swagger/index.html"
echo "Press Ctrl+C to stop the server"
echo "=========================================="
echo ""

go run ./cmd/main.go
