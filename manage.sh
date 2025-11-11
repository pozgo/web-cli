#!/bin/bash

# Management script for web-cli
# Usage:
#   ./manage.sh start   - Start the server
#   ./manage.sh stop    - Stop the server
#   ./manage.sh restart - Restart the server
#   ./manage.sh status  - Check if server is running

set -e

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

# Function to start the server
start() {
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

    # Start the server in background
    ./$APP_NAME &
    PID=$!
    echo $PID > "$PID_FILE"

    sleep 1

    if is_running; then
        echo -e "${GREEN}$APP_NAME started successfully (PID: $PID)${NC}"
        echo -e "${GREEN}Access the application at http://localhost:$PORT${NC}"
    else
        echo -e "${RED}Failed to start $APP_NAME${NC}"
        rm -f "$PID_FILE"
        exit 1
    fi
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

    kill "$PID"
    rm -f "$PID_FILE"

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
    else
        echo -e "${RED}$APP_NAME is not running${NC}"
    fi
}

# Main script
case "$1" in
    start)
        start
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
    *)
        echo "Usage: $0 {start|stop|restart|status}"
        exit 1
        ;;
esac
