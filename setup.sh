#!/bin/bash

set -e

# Create font target directory
FONT_TARGET_DIR="./generate/font"
mkdir -p "$FONT_TARGET_DIR"

# Download Noto Sans Thai font
echo "Downloading Noto Sans Thai font..."
THAI_FONT_URL="https://github.com/notofonts/thai/releases/download/NotoSansThaiLooped-v2.000/NotoSansThaiLooped-v2.000.zip"
THAI_DOWNLOAD_DIR=$(mktemp -d)
THAI_DOWNLOAD_ZIP="$THAI_DOWNLOAD_DIR/NotoSansThaiLooped-v2.000.zip"

curl -L -o "$THAI_DOWNLOAD_ZIP" "$THAI_FONT_URL"
unzip -o "$THAI_DOWNLOAD_ZIP" -d "$THAI_DOWNLOAD_DIR"
find "$THAI_DOWNLOAD_DIR" -name "*.ttf" ! -name "*wght]*" -exec cp {} "$FONT_TARGET_DIR/" \;
rm -rf "$THAI_DOWNLOAD_DIR"

# Download Roboto font
echo "Downloading Roboto font..."
ROBOTO_FONT_URL="https://github.com/googlefonts/roboto-2/releases/download/v2.138/roboto-android.zip"
ROBOTO_DOWNLOAD_DIR=$(mktemp -d)
ROBOTO_DOWNLOAD_ZIP="$ROBOTO_DOWNLOAD_DIR/roboto-android.zip"

curl -L -o "$ROBOTO_DOWNLOAD_ZIP" "$ROBOTO_FONT_URL"
unzip -o "$ROBOTO_DOWNLOAD_ZIP" -d "$ROBOTO_DOWNLOAD_DIR"
# Copy only the regular weight Roboto font
find "$ROBOTO_DOWNLOAD_DIR" -name "Roboto-Regular.ttf" -exec cp {} "$FONT_TARGET_DIR/" \;
rm -rf "$ROBOTO_DOWNLOAD_DIR"

echo "Font installation complete!"