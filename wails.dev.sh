#!/bin/bash
# wails.dev.sh - Platform-aware Wails development wrapper
# This script automatically applies the correct build tags based on the platform

set -e

echo "🚀 Starting Wails development server..."

# Detect platform
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo "📱 Platform: macOS"
    echo "⚙️  Using -tags nosystray to avoid AppDelegate conflicts"
    exec wails dev -tags nosystray "$@"
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    echo "🐧 Platform: Linux"
    exec wails dev "$@"
else
    echo "🪟 Platform: Other (assuming Windows-like)"
    exec wails dev "$@"
fi
