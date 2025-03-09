@echo off
start cmd /k "cd %cd% && go run cmd/server/main.go"
start cmd /k "cd %cd% && go run cmd/client/main.go" 