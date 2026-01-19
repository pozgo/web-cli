#!/bin/bash

# Management script for web-cli
# Usage:
#   ./manage.sh start   - Start the server (with auth)
#   ./manage.sh dev     - Start the server WITHOUT auth (dev mode)
#   ./manage.sh stop    - Stop the server
#   ./manage.sh restart - Restart the server
#   ./manage.sh status  - Check if server is running
#   ./manage.sh fresh   - Start fresh (delete all data and restart)
#   ./manage.sh reset   - Delete all data (keeps server stopped)

set -euo pipefail

APP_NAME="web-cli"
PID_FILE=".web-cli.pid"
PORT=7777

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Function to check if server is running
is_running() {
    if [ -f "$PID_FILE" ]; then
        PID=$(cat "$PID_FILE")
        if ps -p "$PID" > /dev/null 2>&1; then
            return 0
        else
            rm -f "$PID_FILE"
            return 1
        fi
    fi
    return 1
}

# Function to start the server (with optional auth)
# Usage: start [noauth]
start() {
    local noauth="${1:-}"

    if is_running; then
        echo -e "${YELLOW}Server is already running (PID: $(cat $PID_FILE))${NC}"
        exit 1
    fi

    echo -e "${GREEN}Starting $APP_NAME...${NC}"

    # Check if binary exists
    if [ ! -f "./$APP_NAME" ]; then
        echo -e "${RED}Binary not found. Please run ./build.sh first${NC}"
        exit 1
    fi

    if [ "$noauth" = "noauth" ]; then
        # Start without authentication (dev mode)
        export AUTH_ENABLED=false
        echo -e "${YELLOW}Starting in DEV MODE (no authentication)${NC}"

        # Start the server in background without auth
        ./$APP_NAME &
        PID=$!
        echo $PID > "$PID_FILE"

        sleep 1

        if is_running; then
            echo -e "${GREEN}$APP_NAME started successfully (PID: $PID)${NC}"
            echo -e "${GREEN}Access the application at http://localhost:$PORT${NC}"
            echo -e "${RED}WARNING: Authentication is DISABLED${NC}"
        else
            echo -e "${RED}Failed to start $APP_NAME${NC}"
            rm -f "$PID_FILE"
            exit 1
        fi
    else
        # Generate random password for testing
        echo -e "${YELLOW}Generating test credentials...${NC}"

        # Check if openssl is available
        if command -v openssl &> /dev/null; then
            AUTH_PASSWORD=$(openssl rand -base64 16)
        else
            # Fallback to simple random generation
            AUTH_PASSWORD="test-$(date +%s)-$(( RANDOM % 10000 ))"
        fi

        AUTH_USERNAME="admin"

        # Export environment variables for authentication
        export AUTH_ENABLED=true
        export AUTH_USERNAME="$AUTH_USERNAME"
        export AUTH_PASSWORD="$AUTH_PASSWORD"

        # Save credentials to file for reference
        cat > .web-cli-credentials << EOF
# Web CLI Test Credentials
# Generated: $(date)
# ========================================
Username: $AUTH_USERNAME
Password: $AUTH_PASSWORD

# Basic Auth cURL Example:
curl -u $AUTH_USERNAME:$AUTH_PASSWORD http://localhost:$PORT/api/health

# Web Browser:
Visit http://localhost:$PORT and login with the credentials above
EOF

        echo ""
        echo -e "${GREEN}========================================${NC}"
        echo -e "${GREEN}  Test Credentials (saved to .web-cli-credentials)${NC}"
        echo -e "${GREEN}========================================${NC}"
        echo -e "${YELLOW}  Username: ${NC}$AUTH_USERNAME"
        echo -e "${YELLOW}  Password: ${NC}$AUTH_PASSWORD"
        echo -e "${GREEN}========================================${NC}"
        echo ""

        # Start the server in background with auth enabled
        ./$APP_NAME &
        PID=$!
        echo $PID > "$PID_FILE"

        sleep 1

        if is_running; then
            echo -e "${GREEN}$APP_NAME started successfully (PID: $PID)${NC}"
            echo -e "${GREEN}Access the application at http://localhost:$PORT${NC}"
            echo -e "${YELLOW}Authentication is ENABLED - use credentials above${NC}"
            echo ""
            echo -e "${YELLOW}Quick test:${NC}"
            echo -e "  curl -u $AUTH_USERNAME:$AUTH_PASSWORD http://localhost:$PORT/api/health"
            echo ""
        else
            echo -e "${RED}Failed to start $APP_NAME${NC}"
            rm -f "$PID_FILE"
            rm -f .web-cli-credentials
            exit 1
        fi
    fi
}

# Function to start in dev mode (no auth)
dev() {
    start noauth
}

# Function to stop the server
stop() {
    if ! is_running; then
        echo -e "${YELLOW}Server is not running${NC}"
        # Try to kill any process on port 7777 just in case
        if lsof -ti:$PORT > /dev/null 2>&1; then
            echo -e "${YELLOW}Found process on port $PORT, killing it...${NC}"
            lsof -ti:$PORT | xargs kill -9
            echo -e "${GREEN}Process killed${NC}"
        fi
        rm -f "$PID_FILE"
        return
    fi

    PID=$(cat "$PID_FILE")
    echo -e "${YELLOW}Stopping $APP_NAME (PID: $PID)...${NC}"

    # Try graceful shutdown first
    kill "$PID" 2>/dev/null || true
    
    # Wait up to 5 seconds for process to terminate
    for i in {1..5}; do
        if ! ps -p "$PID" > /dev/null 2>&1; then
            break
        fi
        sleep 1
    done
    
    # Force kill if still running
    if ps -p "$PID" > /dev/null 2>&1; then
        echo -e "${YELLOW}Process did not terminate gracefully, forcing...${NC}"
        kill -9 "$PID" 2>/dev/null || true
    fi
    
    rm -f "$PID_FILE"
    rm -f .web-cli-credentials

    echo -e "${GREEN}$APP_NAME stopped${NC}"
}

# Function to restart the server
restart() {
    echo -e "${YELLOW}Restarting $APP_NAME...${NC}"
    stop
    sleep 1
    start
}

# Function to check status
status() {
    if is_running; then
        PID=$(cat "$PID_FILE")
        echo -e "${GREEN}$APP_NAME is running (PID: $PID)${NC}"
        echo -e "${GREEN}Access the application at http://localhost:$PORT${NC}"
        
        # Show credentials if file exists
        if [ -f ".web-cli-credentials" ]; then
            echo ""
            echo -e "${YELLOW}Test credentials available in: .web-cli-credentials${NC}"
            echo -e "${YELLOW}To view: cat .web-cli-credentials${NC}"
        fi
    else
        echo -e "${RED}$APP_NAME is not running${NC}"
    fi
}

# Function to reset/delete all data
reset_data() {
    echo -e "${YELLOW}Deleting all data...${NC}"
    
    # Stop server if running
    if is_running; then
        echo -e "${YELLOW}Stopping server first...${NC}"
        stop
    fi
    
    # Delete data directory
    if [ -d "data" ]; then
        rm -rf data/
        echo -e "${GREEN}Data directory deleted${NC}"
    else
        echo -e "${YELLOW}Data directory does not exist${NC}"
    fi
    
    # Delete credentials file
    rm -f .web-cli-credentials
    
    echo -e "${GREEN}All data has been reset${NC}"
}

# Function to start fresh (reset and start)
fresh() {
    echo -e "${YELLOW}Starting fresh...${NC}"
    reset_data
    echo ""
    start
}

# Main script
case "${1:-}" in
    start)
        start
        ;;
    dev)
        dev
        ;;
    stop)
        stop
        ;;
    restart)
        restart
        ;;
    status)
        status
        ;;
    fresh)
        fresh
        ;;
    reset)
        reset_data
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|status|fresh|reset|dev}"
        echo ""
        echo "Commands:"
        echo "  start   - Start the server (with authentication)"
        echo "  dev     - Start the server WITHOUT authentication (dev mode)"
        echo "  stop    - Stop the server"
        echo "  restart - Restart the server"
        echo "  status  - Check if server is running"
        echo "  fresh   - Delete all data and start fresh"
        echo "  reset   - Delete all data (keeps server stopped)"
        exit 1
        ;;
esac
