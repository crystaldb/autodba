#!/bin/bash

# Colors for different services
AUTODBA_COLOR='\033[0;36m'      # Cyan for autodba
COLLECTOR_COLOR='\033[0;35m'    # Magenta for collector
TIMESTAMP_COLOR='\033[0;33m'    # Yellow for timestamp
NC='\033[0m'                    # No Color

# Function to format and colorize the log output
format_log() {
    while read -r line; do
        # Extract unit name and rest of the message
        if [[ $line =~ autodba-collector ]]; then
            service="collector"
            color=$COLLECTOR_COLOR
        else
            service="autodba"
            color=$AUTODBA_COLOR
        fi
        
        # Extract timestamp and message
        if [[ $line =~ ^([A-Za-z]+[[:space:]]+[0-9]+[[:space:]]+[0-9:]+)[[:space:]](.*)$ ]]; then
            timestamp="${BASH_REMATCH[1]}"
            message="${BASH_REMATCH[2]}"
            echo -e "${TIMESTAMP_COLOR}${timestamp}${NC} ${color}[${service}]${NC} ${message}"
        else
            echo -e "${color}[${service}]${NC} ${line}"
        fi
    done
}

# Run journalctl with combined unit filter
journalctl -xef -u autodba.service -u autodba-collector.service | format_log
