// Package migration implements import of credentials exported from other
// password managers and browsers into the PassQuantum vault.
//
// The package is split into three layers:
//
//  1. Parsers (parser_*.go) read a vendor-specific export format and produce
//     a slice of neutral ImportedEntry values. Parsers run entirely in memory
//     and never touch disk beyond reading the input stream.
//
//  2. The mapper (mapper.go) converts ImportedEntry values into the project's
//     real model.VaultEntry, encrypting each payload through the same Kyber +
//     AES-256-GCM pipeline used by manual entry creation. Plaintext secrets
//     are wiped from memory as soon as they have been re-encrypted.
//
//  3. A registry (importer.go) lets callers iterate over supported formats and
//     run auto-detection from a filename + a small header buffer.
//
// The package never writes plaintext secrets to disk, never opens a network
// connection, and never logs the contents of any field.
package migration

import (
	"errors"
	"time"

	"passquantum/core/model"
)

// MaxFileSize is the hard upper bound (in bytes) on an import file.
// Anything larger is rejected before parsing to avoid memory-pressure DoS.
const MaxFileSize = 100 * 1024 * 1024

// ErrFileTooLarge is returned when an input exceeds MaxFileSize.
var ErrFileTooLarge = errors.New("migration: input file exceeds maximum allowed size")

// ErrUnsupportedFormat is returned when no importer can handle a given file.
var ErrUnsupportedFormat = errors.New("migration: no importer can handle this file")

// CardData groups card-specific fields. Number and CVV are stored as []byte
// so they can be wiped after re-encryption.
type CardData struct {
	Subtype  string // "credit" / "debit" / "" if unknown
	Brand    string // "Visa", "Mastercard", ...
	Holder   string
	Number   []byte
	ExpMonth string
	ExpYear  string
	CVV      []byte
}

// IdentityData groups identity/profile fields. Currently stored as a note
// payload because the VaultEntry model has no dedicated identity type.
type IdentityData struct {
	FullName   string
	Email      string
	Phone      string
	Address    string
	City       string
	State      string
	PostalCode string
	Country    string
}

// ImportedEntry is the neutral intermediate representation produced by every
// parser. Secrets are stored as []byte so the mapper can wipe them after the
// re-encrypted VaultEntry has been built.
type ImportedEntry struct {
	Type model.EntryType

	// Display / lookup
	Title    string   // service or item name
	Username string   // login / email / account
	URLs     []string // 0..N URLs; index 0 is treated as primary

	// Secret (for password / TOTP / card)
	Password []byte // SECRET — wipe after use
	TOTP     string // otpauth:// URI or raw base32 secret; normalized later

	// Free-form
	Notes string

	// Optional extras that VaultEntry cannot natively hold; the mapper folds
	// these into the encrypted Notes payload or discards them with a warning.
	Folder       string
	Tags         []string
	Fields       map[string]string
	Created      time.Time
	Modified     time.Time
	Source       string // importer ID that produced this entry

	// Type-specific blocks
	Card     *CardData
	Identity *IdentityData
}

// ImportResult is what a parser returns. It carries diagnostics so the UI can
// surface "N rows skipped, K entries had data we cannot store" to the user.
type ImportResult struct {
	Entries  []ImportedEntry
	Skipped  int      // rows discarded as empty/invalid
	Warnings []string // human-readable, non-secret hints
}

// ParseOptions carries optional inputs supplied by the UI for formats that
// need them (encrypted files, generic CSV column mapping, ...).
type ParseOptions struct {
	// Password unlocks encrypted exports (Proton .pgp, Bitwarden encrypted
	// JSON, KeePass KDBX). Treated as a secret; the parser must not log it.
	Password []byte

	// ColumnMapping is consumed by the generic CSV importer. Keys are the
	// destination field names ("title", "username", "password", "url",
	// "notes", "totp", "folder"); values are the source column headers.
	ColumnMapping map[string]string
}

// DuplicateAction tells the mapper what to do when an imported entry collides
// with one already in the vault (matched by app.FindDuplicateEntry).
type DuplicateAction int

const (
	// DupSkip drops the imported entry; the existing one is left untouched.
	DupSkip DuplicateAction = iota
	// DupReplace overwrites the existing entry's crypto fields in place.
	DupReplace
	// DupKeepBoth always appends as a new entry; duplicates accumulate.
	DupKeepBoth
)

// MapResult summarizes a mapping pass.
type MapResult struct {
	NewEntries []*model.VaultEntry // entries to append to the vault
	Replaced   int                 // entries already present whose crypto was rewritten in place
	Skipped    int                 // entries dropped because of DupSkip
	Errors     []string            // per-entry mapping errors (no secrets)
	Warnings   []string            // non-fatal notices (no secrets)
}
