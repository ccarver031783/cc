#!/bin/bash
#
# Script: search_patterns.sh
# Purpose: Search GitHub org for whatever pattern or hostname references
# Usage: ./search_patterns.sh [url_list_file] [pattern]
#
# If no file provided, searches for all pattern URLs matching the pattern
# If file provided, searches for each URL in the file
# If pattern provided, searches for all pattern URLs matching the pattern
# If no pattern provided, searches for all pattern URLs matching the pattern
#

ORG=<org>
PATTERN=<pattern>

# Function to search for a specific URL
search_url() {
    local url="$1"
    echo "=== Searching: $PATTERN in $url ===" >&2
    gh search code "$url" --owner "$ORG" --json repository,path 2>/dev/null | \
        jq -r '.[] | "\(.repository.nameWithOwner) - \(.path)"'
}

# Function to search for all green URLs
search_all() {
    echo "Searching for all *.$PATTERN references..." >&2
    gh search code "$PATTERN" --owner "$ORG" --json repository,path --limit 100 2>/dev/null | \
        jq -r '.[] | "\(.repository.nameWithOwner) - \(.path)"' | \
        sort | uniq
}

# Function to filter out documentation/backup files
filter_results() {
    grep -v "\.cursor/rules" | \
    grep -v "CHANGELOG" | \
    grep -v "README" | \
    grep -v "cloudflare-backup" | \
    grep -v "\.md$"
}

# Main
if [ -n "$1" ] && [ -f "$1" ]; then
    echo "Searching for $PATTERN from file: $1"
    echo "================================"
    while IFS= read -r url; do
        if [ -n "$url" ]; then
            result=$(search_url "$url")
            if [ -n "$result" ]; then
                echo ""
                echo "=== $url ==="
                echo "$result"
            fi
        fi
    done < "$1"
else
    echo "Searching for all references to $PATTERN in $ORG org"
    echo "==================================================="
    echo ""
    search_all | filter_results
fi

echo ""
echo "Done! Now go do some work!"