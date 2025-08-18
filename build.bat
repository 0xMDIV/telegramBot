@echo off
echo Building Telegram Security Bot...

if not exist bin mkdir bin

echo Building for Windows...
set CGO_ENABLED=1
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -o bin/telegram-security-bot-windows-amd64.exe .

echo Building for Linux...
set CGO_ENABLED=1
set GOOS=linux
set GOARCH=amd64
go build -ldflags="-s -w" -o bin/telegram-security-bot-linux-amd64 .

echo Build completed!
echo Windows binary: bin/telegram-security-bot-windows-amd64.exe
echo Linux binary: bin/telegram-security-bot-linux-amd64