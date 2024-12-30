#!/bin/bash

function ccff() {
    local original_dir=$(pwd)
    local project_dir="$HOME/code/github.com/quinn/go-astro"
    local cmd_name=$1
    local temp_bin=$(mktemp /tmp/gorun_bin_XXXXXX)

    if [ -z "$cmd_name" ]; then
        echo "Usage: ccff <command>"
        echo "Example: ccff generate/pages"
        return 1
    fi

    echo "Building temporary binary for cmd/$cmd_name in $temp_bin"

    # Change to the project directory and build the project
    cd "$project_dir" || return
    go build -o "$temp_bin" "./cmd/$cmd_name"

    if [ $? -ne 0 ]; then
        echo "Build failed"
        cd "$original_dir"
        rm -f "$temp_bin"
        return 1
    fi

    # Shift the first argument (command name) and pass the rest to the binary
    shift
    # Change back to the original directory and run the temporary binary
    cd "$original_dir" || return
    "$temp_bin" "$@"

    # Clean up the temporary binary
    rm "$temp_bin"
}

# If this script is being sourced, don't execute ccff
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    ccff "$@"
fi
