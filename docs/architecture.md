# Architecture

This document outlines the high-level module boundaries for the markdown server. Populate this as the refactor progresses.

## Packages

- `cmd/server`: binary entrypoint responsible for wiring configuration and starting services.
- `internal/server`: HTTP server setup, routing, and middleware.
- `internal/handler`: request handlers and websocket lifecycle management.
- `internal/storage`: document persistence and filesystem operations.
- `internal/diff`: diff/patch conversion helpers shared across handlers.
- `config`: configuration loading and defaults for the application.
