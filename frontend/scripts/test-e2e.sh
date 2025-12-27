#!/bin/bash

# MrRSS E2E Test Runner
# This script helps you run E2E tests with the Wails backend

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🧪 MrRSS E2E Test Runner${NC}\n"

# Change to frontend directory
cd "$(dirname "$0")/../"

# Check if backend is running
check_backend() {
    if curl -s http://localhost:34115/api/feeds > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# Function to start backend
start_backend() {
    echo -e "${YELLOW}⚠️  Backend is not running${NC}"
    echo "Please start the backend in another terminal:"
    echo ""
    echo "  cd .. && wails3 dev"
    echo ""
    echo "Or press Ctrl+C to exit and start it manually"
    echo ""
    read -p "Press Enter when backend is ready..."
}

# Check backend status
if ! check_backend; then
    start_backend
    if ! check_backend; then
        echo -e "${RED}❌ Backend still not responding. Exiting.${NC}"
        exit 1
    fi
fi

echo -e "${GREEN}✅ Backend is running${NC}\n"

# Run tests
echo "Running Cypress tests..."
echo ""

# Parse arguments
MODE="interactive"
SPEC_FILE=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --headless)
            MODE="headless"
            shift
            ;;
        --spec)
            SPEC_FILE="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --headless    Run tests in headless mode (no GUI)"
            echo "  --spec FILE   Run specific test file"
            echo "  --help        Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0                    # Interactive mode"
            echo "  $0 --headless         # Headless mode"
            echo "  $0 --spec cypress/e2e/app-smoke.cy.ts  # Run specific test"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Run Cypress based on mode
if [ "$MODE" = "headless" ]; then
    if [ -n "$SPEC_FILE" ]; then
        npx cypress run --spec "$SPEC_FILE"
    else
        npx cypress run
    fi
else
    if [ -n "$SPEC_FILE" ]; then
        npx cypress open --spec "$SPEC_FILE"
    else
        npx cypress open
    fi
fi
