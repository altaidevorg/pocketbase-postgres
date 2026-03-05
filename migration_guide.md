# PocketBase SQLite to PostgreSQL Migration Guide

## Overview

PocketBase was originally designed with SQLite as its primary database, deeply coupling certain core features and query generation logic to SQLite-specific SQL functions and behaviors. Moving to PostgreSQL requires addressing these incompatibilities without breaking the existing SQLite support.

This document details the issues encountered, the solutions implemented, and guidance for future development and debugging.

## Key Challenges & Solutions

### 1. Sorting by `@rowid`
**Issue**: The Admin UI requests sorting by `@rowid` ("newest first"), which the backend translated to `ORDER BY _rowid_ DESC`. `_rowid_` is an internal column specific to SQLite and does not exist in PostgreSQL.

**Fix**:
- **Location**: `apis/logs.go`
- **Solution**: Intercept the `sort` query parameter in the logs API. If sorting by `@rowid`, replace it with `created`.
- **Reasoning**: `created` is a standard timestamp column present in all tables, providing the same "chronological order" capability as `rowid`, but in a database-agnostic way.

### 2. JSON Extraction Syntax
**Issue**: The logs filter functionality relied on SQLite's `JSON_EXTRACT` function (e.g., `JSON_EXTRACT(data, '$.auth')`). PostgreSQL uses different operators (`->`, `->>`) for JSON path traversal.

**Fix**:
- **Location**: `tools/search/simple_field_resolver.go`
- **Solution**:
    - Added `isPostgres` configuration to `SimpleFieldResolver`.
    - Implemented a PostgreSQL-specific path generation logic using `->` (traverse) and `->>` (extract text/end).
    - Checks `app.IsPostgres()` in `apis/logs.go` to configure the resolver correctly for each request.

### 3. Date Formatting Functions
**Issue**: The logs statistics query used `strftime('%Y-%m-%d %H:00:00', created)`, which is SQLite-specific. PostgreSQL uses `to_char(timestamp, format)`.

**Fix**:
- **Location**: `core/log_query.go`
- **Solution**:
    - Added a check for `app.IsPostgres()`.
    - Used `to_char(created::timestamp, 'YYYY-MM-DD HH24:00:00')` for PostgreSQL.
    - **Crucial Detail**: The `created` column in `_logs` is text/string in the DB schema (for SQLite compatibility). `to_char` in Postgres requires a timestamp type, so we explicitly cast it (`created::timestamp`).

### 4. Inequality Operator (`!=`)
**Issue**: The search filter logic translated `!=` to `IS NOT`. In SQLite, `col IS NOT 'value'` works for value comparison. In PostgreSQL (and standard SQL), `IS NOT` is strictly for `NULL` / `TRUE` / `FALSE` checks. `col IS NOT 'value'` is a syntax error.

**Fix**:
- **Location**: `tools/search/filter.go`
- **Solution**:
    - Updated `FilterData` logic to check if the field resolver is Postgres-aware.
    - If Postgres is detected, `!=` is translated to `IS DISTINCT FROM`, which is the standard SQL equivalent for null-safe inequality checks.

## Developer Guidelines

### 1. Database Abstraction via `dbx`
- Avoid writing raw SQL strings whenever possible. Use `dbx.Expression` builders.
- If you must write raw SQL (e.g., complex reporting queries), use `app.IsPostgres()` to branch logic. Do not try to find a "common denominator" SQL if it compromises performance or correctness on either DB.

### 2. Field Resolvers
- The `tools/search` package is the central place for translating API query parameters (sort, filter) into SQL.
- If adding new filter capabilities, ensure `FieldResolver` implementations are aware of the underlying database dialect if the syntax differs.

### 3. Type Safety
- Be aware of implicit type coercions. SQLite is very lenient (e.g., treating text execution as dates). PostgreSQL is strict.
- Always explicit cast columns if the Go struct type (`types.DateTime`) implies a certain DB type but the underlying schema might be different.

## Debugging Tips

### Common PostgreSQL Error Codes
- **42601 (Syntax Error)**: Usually means invalid SQL syntax like `IS NOT 'value'` or correct SQL keywords used in the wrong place.
- **42883 (Undefined Function)**: Function does not exist or *signature mismatch*. e.g., calling `to_char(text, text)` instead of `to_char(timestamp, text)`. This often means you need a cast.
- **42703 (Undefined Column)**: Often happens if `_rowid_` or other internal SQLite columns are referenced.

### Troubleshooting Steps
1. **Check the Logs**: The `pq` driver errors usually print the failed query and the character position. Use an online SQL formatter to find the exact syntax error location.
2. **Reproduce in CLI**: Run the generated SQL directly in a `psql` shell connected to the dev DB to iterate quickly on syntax fixes.
3. **Trace the Builder**: `tools/search/provider.go`, `filter.go`, and `sort.go` are the engines converting URL params to SQL. Look there first for query generation issues.
