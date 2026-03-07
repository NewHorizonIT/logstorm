#!/bin/bash

# LogStorm Test Runner
# Usage: ./run_tests.sh [test_name]

set -e

BASE_URL="${BASE_URL:-http://localhost:3123}"
RESULTS_DIR="./results"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Create results directory
mkdir -p "$RESULTS_DIR"

# Check if k6 is installed
check_k6() {
    if ! command -v k6 &> /dev/null; then
        echo -e "${RED}Error: k6 is not installed${NC}"
        echo "Install k6: https://k6.io/docs/getting-started/installation/"
        exit 1
    fi
}

# Check if server is running
check_server() {
    echo "Checking server at $BASE_URL..."
    if curl -s "$BASE_URL/health" > /dev/null; then
        echo -e "${GREEN}Server is running${NC}"
    else
        echo -e "${RED}Error: Server not responding at $BASE_URL${NC}"
        exit 1
    fi
}

# Run test
run_test() {
    local test_name=$1
    local test_file="${test_name}_test.js"
    
    if [ ! -f "$test_file" ]; then
        echo -e "${RED}Error: Test file $test_file not found${NC}"
        exit 1
    fi
    
    echo -e "${YELLOW}Running $test_name test...${NC}"
    echo "Output: $RESULTS_DIR/${test_name}_$(date +%Y%m%d_%H%M%S).json"
    
    k6 run \
        --out json="$RESULTS_DIR/${test_name}_$(date +%Y%m%d_%H%M%S).json" \
        -e BASE_URL="$BASE_URL" \
        "$test_file"
}

# Run all tests
run_all() {
    echo "Running all tests..."
    echo ""
    
    for test in smoke load stress spike; do
        run_test $test
        echo ""
        sleep 5  # Brief pause between tests
    done
}

# Help
show_help() {
    echo "LogStorm Test Runner"
    echo ""
    echo "Usage: ./run_tests.sh [command]"
    echo ""
    echo "Commands:"
    echo "  smoke       Run smoke test (quick validation)"
    echo "  load        Run load test (comprehensive)"
    echo "  stress      Run stress test (find limits)"
    echo "  spike       Run spike test (burst traffic)"
    echo "  soak        Run soak test (long running)"
    echo "  all         Run all tests except soak"
    echo "  generate    Generate test logs"
    echo "  help        Show this help"
    echo ""
    echo "Examples:"
    echo "  ./run_tests.sh smoke"
    echo "  BASE_URL=http://prod:3123 ./run_tests.sh load"
    echo "  ./run_tests.sh generate --count 1000"
}

# Main
check_k6

case "${1:-help}" in
    smoke|load|stress|spike|soak)
        check_server
        run_test "$1"
        ;;
    all)
        check_server
        run_all
        ;;
    generate)
        check_server
        shift
        node generate_logs.js "$@"
        ;;
    help|*)
        show_help
        ;;
esac
