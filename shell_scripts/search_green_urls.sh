#!/bin/bash
#
# Script: search_green_urls.sh
# Purpose: Search GitHub org for Istio green hostname references
# Usage: ./search_green_urls.sh [url_list_file]
#
# If no file provided, searches for all green URLs matching the pattern
# If file provided, searches for each URL in the file
#

ORG=<org>
PATTERN=<whatever pattern you want to search up>

# Function to search for a specific URL
search_url() {
    local url="$1"
    echo "=== Searching: $url ===" >&2
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
    echo "Searching for URLs from file: $1"
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
    echo "Searching for all green URL references in $ORG org"
    echo "==================================================="
    echo ""
    search_all | filter_results
fi

echo ""
echo "Done!"
