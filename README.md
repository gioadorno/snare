# snare 🪤

**The Local Webhook Interceptor & Replayer**

## Overview
Clients often have undocumented webhooks, APIs, or legacy systems that fire payloads unpredictably. `snare` is a localized "black hole" server that captures incoming HTTP requests, logs them to a local database, and allows you to perfectly replay them against your own development environment.

## Features
*   **Catch-All Interceptor:** Binds to a local port and captures 100% of incoming HTTP traffic (Headers, Body, Params, Method).
*   **Local Persistence:** Saves all trapped requests to a lightweight, embedded SQLite database.
*   **One-Click Replay:** Forwards any trapped request to a target URL exactly as it was received.
*   **TUI Viewer:** Inspect trapped payloads directly in the terminal before replaying.

## Usage
```bash
# Start the trap on port 8080
snare listen -port 8080

# List trapped requests
snare list

# Replay request ID 45 against your local dev server
snare replay -id 45 -target http://localhost:3000/api/v1/webhook
```

## FDE Philosophy
**Decouple testing from the client.** You cannot rely on a client to "fire a test webhook" every time you need to debug. Capture the chaotic network traffic once, and replay it locally a hundred times.
