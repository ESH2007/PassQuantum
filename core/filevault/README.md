# core/filevault/

Encrypted file storage. Lets the user keep arbitrary files (documents, images,
keys, …) inside a vault, each encrypted individually with the same post-quantum
primitives used for vault entries (Kyber768 key exchange + AES-256-GCM). A
per-vault JSON manifest tracks the stored files; the encrypted blobs live next
to it on disk.

| File | Description |
|---|---|
| `store.go` | `Store` — the high-level API. `NewStore` binds a vault name and the Kyber keypair; `StoreFile`/`RetrieveFile` encrypt-in / decrypt-out with progress callbacks; `OpenFile` decrypts to a tracked temp file for viewing; `DecryptToMemory` returns plaintext bytes; `DeleteFile`, `ListFiles`, and `LoadManifest`/`SaveManifest` round out CRUD. |
| `manifest.go` | `FileManifest` and `FileMetadata` — the on-disk index (UUID, original name, size, timestamps, per-file Kyber ciphertext) describing every stored file. |
| `crypto.go` | Streaming file encryption: `EncryptFile`/`DecryptFile` and their `…WithProgress` variants chunk through an `io.Reader`/`io.Writer` using AES-256-GCM, so large files never need to be fully buffered in memory. |
| `tempfiles.go` | `TempTracker` registers and cleans up temp files created by `OpenFile`; `SecureDelete` overwrites before unlinking; `CleanupOrphans` removes leftovers from a previous run; `TempFilePath` builds a tracked temp path. |
| `crypto_test.go` | Round-trip tests for the streaming encrypt/decrypt helpers. |
