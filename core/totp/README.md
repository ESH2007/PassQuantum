# core/totp/

TOTP (time-based one-time password) support for 2FA vault entries. No UI
dependency — operates on parameter structs and images and returns codes/values
the UI renders. Built on top of `github.com/pquerna/otp`.

| File | Description |
|---|---|
| `totp.go` | Core type `TOTPParams` (secret, issuer, account, algorithm, digits, period) plus `GenerateCode` (current code + seconds remaining), `ParseOTPAuthURI` (`otpauth://totp/...` → params), `Validate`, `DefaultParams`, and JSON `Serialize`/`Deserialize` for encrypted storage in a vault entry. |
| `qr.go` | `DecodeQRFromImage` reads a QR code out of an `image.Image`; `DecodeQRToTOTP` decodes a QR and parses the embedded `otpauth://` URI into `TOTPParams` in one call. Uses `makiuchi-d/gozxing`. |
| `migration.go` | `ParseGoogleAuthExport` decodes a Google Authenticator export payload (`otpauth-migration://offline?data=...`) into a slice of `TOTPParams`. Includes a minimal hand-written protobuf decoder so the full protobuf dependency is not required. |
| `totp_test.go` | Unit tests for code generation, URI parsing, validation, and the Google Authenticator migration decoder. |

A TOTP vault entry stores the serialized `TOTPParams` as its encrypted payload;
the live 6/8-digit code is computed on demand and never persisted.
