# internal/

Internal packages that are not part of the public API and cannot be imported
by code outside this module.

| Package | Description |
|---|---|
| [`storage/`](storage/README.md) | Low-level secure file I/O: vault file reads/writes with strict OS permissions, OS keyring integration, and Windows DPAPI wrapping. |
| [`browser/`](browser/README.md) | Localhost-only HTTP autofill server (`127.0.0.1:8765`) that pairs with the browser extension and serves domain-matched vault lookups. |
