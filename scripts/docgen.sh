#!/bin/bash

# Function to generate a unique anchor
generate_anchor() {
    echo "$1" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-zA-Z0-9]/-/g' | sed 's/--*/-/g' | sed 's/^-//;s/-$//'
}

# Function to process references
process_references() {
    echo "$1" | sed -E 's/#([a-zA-Z0-9_/]+)/[#\1](#\1)/g'
}

# Function to generate markdown for a category
generate_markdown() {
    local category="$1"
    local level="$2"
    local anchor=$(generate_anchor "$category")
    
    if [[ $level -eq 0 ]]; then
        echo "# API Documentation"
        echo
    fi

    if [[ $category != "root" ]]; then
        printf '%s %s {#%s}\n\n' "$(printf '#%.0s' $(seq 1 $((level+1))))" "$category" "$anchor"
    fi

    # Process items in this category
    grep -n "@template($category" "$input_file" | while IFS=':' read -r line_num content; do
        name=$(echo "$content" | sed -E 's/.*@template\([^)]*\)([^@]*).*/\1/' | tr -d ' ')
        anchor=$(generate_anchor "$category-$name")
        echo "### $name {#$anchor}"
        echo
        sed -n "$line_num,/^[[:space:]]*$/p" "$input_file" | sed '1d;$d' | sed 's/^[ \t]*//' | process_references
        echo
    done

    # Process subcategories
    grep "@template($category/" "$input_file" | sed -E 's/.*@template\(([^)]*)\).*/\1/' | cut -d'/' -f2 | sort -u | while read -r subcategory; do
        generate_markdown "$category/$subcategory" $((level + 1))
    done
}

# Main script
if [[ $# -lt 1 ]]; then
    echo "Please provide a file to parse"
    exit 1
fi

input_file="$1"

# Generate root level categories
grep "@template(" "$input_file" | sed -E 's/.*@template\(([^/)]*).*/\1/' | sort -u | while read -r category; do
    generate_markdown "$category" 0
done