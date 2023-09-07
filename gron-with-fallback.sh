#!/usr/bin/env bash

if [ "$#" -eq 0 ]; then
    input_file="$(mktemp)"
    cat - >"$input_file"
    gron "$input_file" 2>/dev/null || cat "$input_file"
    rm "$input_file"
else
    gron "$1" 2>/dev/null || cat "$1"
fi
