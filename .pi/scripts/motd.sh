#!/usr/bin/env bash
echo "Current date: $(date '+%Y-%m-%d')"
echo "Current working directory: $(pwd)"
echo
echo "Go project environment:
$(go env | grep -E 'GOMOD=|GOOS=|GOVERSION=')"
echo "Go module: $(go list -m)"
