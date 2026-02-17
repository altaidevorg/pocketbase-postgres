@echo off
REM Set the Postgres connection string
set PB_DB_CONNECT=postgres://user:pass@localhost:5432/db?sslmode=disable

REM Run the PocketBase app in dev mode using go run (compiles and runs the latest code)
echo Starting PocketBase with Postgres in DEV mode...
go run examples/base/main.go serve --http=127.0.0.1:8091 --dev
#examples\base\base serve --http=127.0.0.1:8091 --dev
REM Pause so the window stays open if there is an error
if %ERRORLEVEL% NEQ 0 (
    echo.
    echo PocketBase exited with error code %ERRORLEVEL%.
    pause
)
