# Build Plan: snare

## Step 1: HTTP Interceptor Core
*   Scaffold standard `net/http` server with a catch-all route (`/`).
*   Extract headers, query parameters, method, and read the raw request body.

## Step 2: SQLite Persistence
*   Integrate `mattn/go-sqlite3` or `glebarez/go-sqlite` (cgo-free).
*   Create a schema to store trapped requests (`id`, `method`, `path`, `headers_json`, `body_blob`, `captured_at`).
*   Write incoming requests directly to the DB.

## Step 3: CLI Interface
*   Use Cobra to build the `listen`, `list`, and `replay` commands.
*   Implement a simple tabular view of recent traps for the `list` command.

## Step 4: The Replay Engine
*   Build the HTTP client logic to fetch a request from SQLite, reconstruct the `http.Request` struct, and fire it at the specified target URL.
*   Print the target server's response back to the terminal for debugging.
