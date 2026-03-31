# PocketBase PostgreSQL Fork Usage Guide

This guide is intended for developers who want to use the PostgreSQL-compatible fork of PocketBase as a framework in their Go applications. 

By default, the upstream PocketBase is heavily coupled with SQLite. This fork modifies core logic (such as `@rowid` sorting, `JSON_EXTRACT`, and `strftime` operations) to run natively on PostgreSQL.

---

## 1. Setup & Installation

When using this fork as a Go module dependency for your application, Go's module resolver will naturally try to fetch from the official upstream repository (`github.com/pocketbase/pocketbase`), which completely ignores the PostgreSQL patches.

To force your project to use your remote GitHub fork instead, you must use the `replace` directive in your project's `go.mod`.

### Terminal Command:
Inside your own Go project directory (where your app's `main.go` lives), run this to override the module:

```bash
go mod edit -replace github.com/pocketbase/pocketbase=github.com/altaidevorg/pocketbase-postgres@main
```
*(Note: You can replace `@master` with a specific commit hash or version tag like `@v0.22.x` for stability).*

### Verifying your `go.mod`:
This command automatically updates your `go.mod` to look something like this:

```go
module my-cool-app

go 1.23

require github.com/pocketbase/pocketbase v0.22.0

// Redirects the official import path to our postgres fork!
replace github.com/pocketbase/pocketbase => github.com/altaidevorg/pocketbase-postgresql master
```

Finally, tidy up the dependencies to download the fork:
```bash
go mod tidy
```

---

## 2. Configuring PostgreSQL Credentials

To start your PocketBase application backed by PostgreSQL, you must provide your database credentials via a Data Source Name (DSN). PocketBase configurations expect this URL connection string format:

```text
postgres://[username]:[password]@[host]:[port]/[database_name]?sslmode=disable
```

### Initializing the App
When creating your `main.go`, you no longer rely on the default SQLite initialization. Provide the `DB_URL` via environment variables or pass it during initialization.

#### Option A: Using Environment Variables (Recommended)
If your app is already set up to read from database environment variables, you simply export `DB_URL` or `PB_DB_URL` before running your server:

```bash
# Export the credentials string
export DB_URL="postgres://postgres:mysecretpassword@localhost:5432/pocketbase_db?sslmode=disable"

# Run your app
go run main.go serve
```

#### Option B: Hardcoding / Passing Custom Config 
If you initialize the App with a custom configuration struct directly in Go:

```go
package main

import (
	"log"
	"os"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func main() {
    // Determine your PostgreSQL DSN (from env or secret manager)
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        dsn = "postgres://postgres:password@localhost:5432/pb_data?sslmode=disable"
    }

	// Initialize pocketbase
	app := pocketbase.New()

    // ... attach custom logic, PB hooks, run migrations ...

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
```

---

## 3. Supported PostgreSQL Features
This fork patches the following compatibility blockers across standard endpoints:
- **Lists / Searches**: Sorts containing SQLite's internal `-@rowid` are safely translated to `-created`.
- **JSON Queries**: PocketBase list filters mapped with `JSON_EXTRACT` correctly use `->` and `->>` when filtering JSON payload data.
- **Log Stats**: Query aggregates dynamically use standard Postgres `to_char(created::timestamp)` in place of `strftime`.
- **Inequality Operators**: Standard PocketBase `!=` queries now successfully map to `IS DISTINCT FROM` rather than `IS NOT`, allowing safe non-null value filtering.

## 4. Troubleshooting
If you receive the following database panics/errors at runtime:
- `pq: syntax error at or near "$1"`: Usually means an operator translation missed the Postgres check. Ensure `app.IsPostgres()` is returning `true`.
- `function to_char(text, unknown) does not exist`: Occurs if a datetime column (like `-created`) was not properly cast to `::timestamp` in the query builder.
- `database "pb_data" does not exist`: Ensure you have explicitly created the Postgres database prior to starting the application; PocketBase will attempt to auto-run its system migrations (`pb_migrations`) on the existing database, but it cannot create the initial logical database itself.
