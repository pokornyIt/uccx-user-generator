@echo off
echo Build Windows
set GOOS=windows
set GOARCH=amd64
:: -s - strip binary data from debug information
:: -w - remove DWARF table
go install
::go build -i -ldflags="-s -w" -o "./builds/ccx-user-generator.exe"
go build -i -o "./builds/ccx-user-generator.exe"

goto end
echo .
echo Build Linux amd64
set GOOS=linux
set GOARCH=amd64
go install
go build -i -ldflags="-s -w" -o "./builds/ccx-user-generator"

:end
set GOOS=
set GOARCH=
