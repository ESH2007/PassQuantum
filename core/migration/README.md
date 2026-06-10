# core/migration/

Import framework for migrating from other password managers. Auto-detects an
export file's format, parses it into a normalized intermediate model, then maps
and encrypts the entries into PassQuantum vault entries. Parsers never log secret
field contents, and every secret byte slice touched during mapping is wiped.

## Framework

| File | Description |
|---|---|
| `importer.go` | The `Importer` interface (`ID`, `DisplayName`, `Extensions`, `Detect`, `Parse`) and the `Registry`. Each `parser_*.go` registers itself with the package-level `DefaultRegistry` in its `init()`. `Registry.Detect` scores all importers that accept the file's extension and returns them sorted by confidence. |
| `detect.go` | `DetectFile` reads the header bytes and runs detection; `ValidateSize` enforces `MaxFileSize`; `OpenLimited` wraps the file in an `io.LimitedReader` so a parser can never read past the cap. |
| `model.go` | The normalized intermediate types: `ImportedEntry` (with `CardData`/`IdentityData`), `ImportResult`, `ParseOptions`, and `DuplicateAction`. |
| `mapper.go` | `MapAndEncrypt` converts `[]ImportedEntry` into encrypted `*model.VaultEntry` values (Kyber768 + AES-GCM), de-duplicates against existing entries per the chosen `DuplicateAction`, wipes secrets as it goes, and returns warnings/errors without leaking field contents. |
| `normalize.go` | URL/title/service helpers: `NormalizeURL`, `DomainOf`, `DedupURLs`, `DeriveTitle`, `DeriveServiceName`, `BuildNotesPayload`. |
| `csv_common.go` / `zip_common.go` | Shared helpers for CSV-based and ZIP-archive-based exports. |

## Parsers

One `parser_*.go` per supported source format:

| File | Source |
|---|---|
| `parser_chromium.go` | Chrome / Brave / Edge / Opera / Vivaldi |
| `parser_firefox.go` | Mozilla Firefox |
| `parser_1password.go` | 1Password (1PUX) |
| `parser_bitwarden.go` | Bitwarden (CSV and JSON) |
| `parser_keepass.go` | KeePass / KeePassXC |
| `parser_dashlane.go` | Dashlane |
| `parser_kaspersky.go` | Kaspersky Password Manager (TXT) |
| `parser_nordpass.go` | NordPass |
| `parser_protonpass.go` | Proton Pass |
| `parser_lastpass.go` | LastPass |
| `parser_generic.go` | Generic CSV (auto-maps columns); the fallback when nothing else matches |

The import UI lives in `ui/screens/import_wizard.go`.
