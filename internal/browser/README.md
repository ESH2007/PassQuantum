# internal/browser/

Localhost autofill server that backs the PassQuantum browser extension. It runs
an HTTP server bound to `127.0.0.1:8765`, gated by a one-time pairing handshake,
and answers domain-matched credential lookups and save/update requests from the
extension. Not importable from outside this module. The extension client lives in
the repo-root `extension/` directory.

## Trust boundary

- Binds to loopback only and rejects any request whose host is not localhost.
- Requires pairing first: the desktop app shows a short-lived token, the user
  enters it in the extension, and the extension thereafter authenticates with the
  paired secret. Unpaired requests are refused.
- Includes a small dependency-free token-bucket rate limiter.
- Only ever exposes credentials for an **already-unlocked** vault.

| File | Description |
|---|---|
| `server.go` | `Server`: builds the route mux, applies the localhost-only/CORS middleware and rate limiter, and manages start/stop on `127.0.0.1:8765`. Routes: `/vault/pair`, `/vault/status`, `/vault/exists`, `/vault/save`, `/vault/update/`, `/vault/never-save`. |
| `handlers.go` | Request/response JSON types and the `handle*` methods for each route. |
| `pairing.go` | `PairingState`: starts a pairing window, surfaces the token to the UI, and validates the token the extension submits. |
| `config.go` | Persisted extension config: paired secret, the per-domain "never save" list, and load/save to disk. |
| `domain.go` | `NormalizeDomain` — canonicalizes a hostname for matching (strips `www.`, lowercases, etc.). |
| `domain_map.go` | `DomainMap`: persistent association between a domain and the vault entry IDs that apply to it (`Lookup`, `Associate`, `Dissociate`). |
| `vault_service.go` | `VaultService` interface (`IsReady`, `Status`, `FindCredentials`, `SaveCredential`, `UpdatePassword`) — the abstraction the server depends on, keeping it decoupled from `app`. |
| `vault_service_impl.go` | `appVaultService`: the concrete implementation backed by `app.AppState` and a `DomainMap`. |

The desktop side wires pairing through `ui/screens/pairing_dialog.go`.
