#!/bin/bash

# ABOUTME: Build performance monitoring script
# ABOUTME: Tracks build times and performance metrics for regression detection

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
METRICS_DIR="$PROJECT_ROOT/.build-metrics"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
METRICS_FILE="$METRICS_DIR/build-$TIMESTAMP.json"

# Create metrics directory if it doesn't exist
mkdir -p "$METRICS_DIR"

# Function to log build metrics
log_metric() {
    local component="$1"
    local start_time="$2"
    local end_time="$3"
    local exit_code="$4"

    local duration=$((end_time - start_time))

    # Create JSON entry
    local json_entry=$(cat <<EOF
{
    "timestamp": "$TIMESTAMP",
    "component": "$component",
    "duration_seconds": $duration,
    "exit_code": $exit_code,
    "success": $([ $exit_code -eq 0 ] && echo "true" || echo "false")
}
EOF
)

    # Append to metrics file
    if [ ! -f "$METRICS_FILE" ]; then
        echo "[$json_entry" > "$METRICS_FILE"
    else
        # Remove the last ] and add comma and new entry
        sed -i.bak '$ s/\]/,/' "$METRICS_FILE" && rm "$METRICS_FILE.bak" 2>/dev/null || true
        echo "$json_entry" >> "$METRICS_FILE"
    fi
}

# Function to finalize metrics file
finalize_metrics() {
    if [ -f "$METRICS_FILE" ]; then
        echo "]" >> "$METRICS_FILE"
    fi
}

# Function to monitor a command
monitor_command() {
    local component="$1"
    shift
    local command="$@"

    echo "🔍 Monitoring build: $component"
    echo "📝 Command: $command"

    local start_time=$(date +%s)
    local exit_code=0

    # Run the command and capture exit code
    $command || exit_code=$?

    local end_time=$(date +%s)
    local duration=$((end_time - start_time))

    log_metric "$component" "$start_time" "$end_time" "$exit_code"

    if [ $exit_code -eq 0 ]; then
        echo "✅ $component completed successfully in ${duration}s"
    else
        echo "❌ $component failed after ${duration}s (exit code: $exit_code)"
    fi

    return $exit_code
}

# Function to display performance summary
show_performance_summary() {
    echo ""
    echo "📊 Build Performance Summary:"
    echo "=============================="

    if [ ! -f "$METRICS_FILE" ]; then
        echo "No metrics file found"
        return
    fi

    # Read the metrics and display summary
    local total_time=0
    local successful_builds=0
    local failed_builds=0

    # Parse JSON manually (avoiding dependencies)
    while read -r line; do
        if [[ $line =~ \"duration_seconds\":[[:space:]]*([0-9]+) ]]; then
            local duration=${BASH_REMATCH[1]}
            total_time=$((total_time + duration))
        fi

        if [[ $line =~ \"success\":[[:space:]]*true ]]; then
            successful_builds=$((successful_builds + 1))
        elif [[ $line =~ \"success\":[[:space:]]*false ]]; then
            failed_builds=$((failed_builds + 1))
        fi
    done < "$METRICS_FILE"

    local total_builds=$((successful_builds + failed_builds))

    echo "Total build time: ${total_time}s"
    echo "Successful builds: $successful_builds"
    echo "Failed builds: $failed_builds"
    echo "Total builds: $total_builds"

    if [ $total_builds -gt 0 ]; then
        local avg_time=$((total_time / total_builds))
        echo "Average build time: ${avg_time}s"
    fi

    echo ""
    echo "📁 Metrics saved to: $METRICS_FILE"
}

# Main execution
main() {
    local target="${1:-all}"

    echo "🚀 Starting monitored build for target: $target"
    echo "📅 Timestamp: $TIMESTAMP"
    echo ""

    # Initialize metrics file
    echo "[" > "$METRICS_FILE"

    # Trap to ensure metrics file is finalized
    trap finalize_metrics EXIT

    case "$target" in
        "all")
            monitor_command "tree-sitter" make tree-sitter
            monitor_command "pvm" make pvm
            monitor_command "pvx" make pvx
            monitor_command "pvi" make pvi
            monitor_command "psc" make psc
            ;;
        "tree-sitter")
            monitor_command "tree-sitter" make tree-sitter
            ;;
        "pvm"|"pvx"|"pvi"|"psc")
            monitor_command "$target" make "$target"
            ;;
        "test")
            monitor_command "test" make test
            ;;
        *)
            echo "Unknown target: $target"
            echo "Usage: $0 [all|tree-sitter|pvm|pvx|pvi|psc|test]"
            exit 1
            ;;
    esac

    show_performance_summary
}

# Run main function with all arguments
main "$@"
