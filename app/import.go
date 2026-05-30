package app

import (
	"fmt"
	"io"
	"os"

	"passquantum/core/migration"
)

// ImportSummary is returned by BatchImport and carries everything the UI
// needs to show a final result screen. All fields are safe to display
// (no secret material).
type ImportSummary struct {
	ParseWarnings []string // warnings produced by the parser
	MapWarnings   []string // warnings produced by the mapper
	MapErrors     []string // per-entry mapping errors

	TotalParsed int
	Skipped     int // rows the parser skipped (empty/invalid)
	NewEntries  int // entries appended to the vault
	Replaced    int // existing entries whose crypto was rewritten in place
	DupSkipped  int // entries that collided and were dropped per DupSkip
}

// ParseImportFile is a thin wrapper around the migration package that opens
// the file, enforces the size cap and delegates to the chosen importer.
func ParseImportFile(path, importerID string, opts migration.ParseOptions) (*migration.ImportResult, migration.Importer, error) {
	if err := migration.ValidateSize(path); err != nil {
		return nil, nil, err
	}
	imp, ok := migration.DefaultRegistry.ByID(importerID)
	if !ok {
		return nil, nil, fmt.Errorf("unknown importer: %s", importerID)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	limited := io.LimitReader(f, migration.MaxFileSize+1)
	res, err := imp.Parse(limited, opts)
	if err != nil {
		return nil, imp, err
	}
	return res, imp, nil
}

// BatchImport applies a parsed ImportResult to the currently-unlocked vault.
// It is the *only* code path that writes imported data: the encryption
// envelope, dedup logic and atomic write all happen here so the import
// converges on the same crypto pipeline as manual entry creation.
//
// The caller must already hold a parse result (typically from ParseImportFile)
// and have a vault unlocked (appState.IsUnlocked == true).
func BatchImport(
	appState *AppState,
	parsed *migration.ImportResult,
	dupAction migration.DuplicateAction,
) (*ImportSummary, error) {
	if appState == nil || !appState.IsUnlocked {
		return nil, fmt.Errorf("vault is not unlocked")
	}
	if appState.CurrentVault == "" {
		return nil, fmt.Errorf("no vault selected")
	}
	if parsed == nil {
		return nil, fmt.Errorf("nil parse result")
	}

	appState.Mu.Lock()
	defer appState.Mu.Unlock()

	vaultFile := GetVaultPath(appState.CurrentVault)
	existing, err := ReadVault(vaultFile, appState.MasterPassword)
	if err != nil {
		return nil, fmt.Errorf("read vault: %w", err)
	}

	mapped, err := migration.MapAndEncrypt(parsed.Entries, appState.PublicKey, existing, dupAction)
	if err != nil {
		return nil, fmt.Errorf("map and encrypt: %w", err)
	}

	// existing is mutated in place by DupReplace; the new entries are
	// appended afterwards so collisions never duplicate rows.
	combined := append(existing, mapped.NewEntries...)

	if err := WriteVault(combined, vaultFile, appState.MasterPassword); err != nil {
		return nil, fmt.Errorf("write vault: %w", err)
	}

	return &ImportSummary{
		ParseWarnings: parsed.Warnings,
		MapWarnings:   mapped.Warnings,
		MapErrors:     mapped.Errors,
		TotalParsed:   len(parsed.Entries),
		Skipped:       parsed.Skipped,
		NewEntries:    len(mapped.NewEntries),
		Replaced:      mapped.Replaced,
		DupSkipped:    mapped.Skipped,
	}, nil
}
