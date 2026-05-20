# cmd/

Standalone command-line utilities. Each subdirectory is a separate `package main`
and can be built independently with `go build ./cmd/<name>`.

| Directory | Description |
|---|---|
| `test-vault/` | Manual vault smoke-test utility: creates a vault, writes a test entry, re-reads it, and prints the result. Useful for verifying the vault encryption pipeline end-to-end without launching the full UI. Run with `go run ./cmd/test-vault`. |
