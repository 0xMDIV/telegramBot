#!/bin/bash

echo "Building Telegram Security Bot..."

mkdir -p bin

echo "Building for Linux..."
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/telegram-security-bot-linux-amd64 .

echo "Building for Windows..."
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o bin/telegram-security-bot-windows-amd64.exe .

echo "Build completed!"
echo "Linux binary: bin/telegram-security-bot-linux-amd64"
echo "Windows binary: bin/telegram-security-bot-windows-amd64.exe"