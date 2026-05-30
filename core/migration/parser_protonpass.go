package migration

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"passquantum/core/model"
)

// ---------------- JSON schema ----------------

// protonPassExport is the top-level structure of Proton Pass's export. The
// vaults field is an *object* keyed by vault ID, not an array.
type protonPassExport struct {
	Version string                      `json:"version"`
	Vaults  map[string]protonPassVault  `json:"vaults"`
}

type protonPassVault struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Items       []protonPassItem   `json:"items"`
}

type protonPassItem struct {
	ItemID     string                 `json:"itemId"`
	CreateTime int64                  `json:"createTime"`
	ModifyTime int64                  `json:"modifyTime"`
	State      int                    `json:"state"` // 1 = active, 2 = trashed
	Data       protonPassItemData     `json:"data"`
}

type protonPassItemData struct {
	Type     string                 `json:"type"` // "login", "note", "alias", "creditCard"
	Metadata protonPassMetadata     `json:"metadata"`
	Content  json.RawMessage        `json:"content"`
	Extra    []protonPassExtraField `json:"extraFields"`
}

type protonPassMetadata struct {
	Name string `json:"name"`
	Note string `json:"note"`
}

type protonPassExtraField struct {
	FieldName string                 `json:"fieldName"`
	Type      string                 `json:"type"`
	Data      map[string]json.RawMessage `json:"data"`
}

// Per-type content payloads.

type protonPassLoginContent struct {
	Username  string   `json:"username"`
	Password  string   `json:"password"`
	URLs      []string `json:"urls"`
	TOTPURI   string   `json:"totpUri"`
	ItemEmail string   `json:"itemEmail"`
}

type protonPassCardContent struct {
	CardholderName string `json:"cardholderName"`
	Number         string `json:"number"`
	ExpirationDate string `json:"expirationDate"` // "MM/YY"
	VerificationNumber string `json:"verificationNumber"`
	CardType       string `json:"cardType"`
	PIN            string `json:"pin"`
}

type protonPassAliasContent struct {
	AliasEmail string `json:"aliasEmail"`
}

// ---------------- Importer ----------------

// ProtonPassImporter parses Proton Pass exports. Two shapes are accepted:
//
//   - A bare data.json file (unencrypted export, JSON content type).
//   - A ZIP archive that contains data.json plus optional attachments.
//
// The encrypted .pgp variant is not yet supported (Phase 3).
type ProtonPassImporter struct{}

func init() {
	DefaultRegistry.Register(&ProtonPassImporter{})
}

func (ProtonPassImporter) ID() string           { return "protonpass" }
func (ProtonPassImporter) DisplayName() string  { return "Proton Pass" }
func (ProtonPassImporter) Extensions() []string { return []string{".json", ".zip"} }

func (ProtonPassImporter) Detect(filename string, head []byte) float64 {
	lower := strings.ToLower(filename)

	// .zip archive: look for the signature and for "proton" in the filename
	// (Proton Pass exports are usually named like proton-pass-export-*.zip).
	if looksLikeZIP(head) {
		if strings.Contains(lower, "proton") || strings.Contains(lower, "pass-export") {
			return 0.85
		}
		// A bare ZIP without a hint is uncertain — let other importers win.
		return 0.2
	}

	// .json: look for the unambiguous "vaults" object + nested "metadata"
	// shape that Proton emits.
	h := string(head)
	if strings.Contains(h, "\"vaults\"") && strings.Contains(h, "\"metadata\"") &&
		strings.Contains(h, "\"itemId\"") {
		return 0.95
	}
	return 0
}

func (ProtonPassImporter) Parse(r io.Reader, _ ParseOptions) (*ImportResult, error) {
	raw, err := readWholeReader(r)
	if err != nil {
		return nil, err
	}

	// Sniff: ZIP or bare JSON?
	var dataBytes []byte
	if looksLikeZIP(raw) {
		zr, _, err := readWholeZIP(bytes.NewReader(raw))
		if err != nil {
			return nil, err
		}
		dataBytes, _, err = findZIPFile(zr, "data.json")
		if err != nil {
			return nil, err
		}
		if len(dataBytes) == 0 {
			return nil, errors.New("proton pass: data.json not found inside ZIP")
		}
	} else {
		dataBytes = raw
	}

	var export protonPassExport
	if err := json.Unmarshal(dataBytes, &export); err != nil {
		return nil, fmt.Errorf("proton pass json: %w", err)
	}

	result := &ImportResult{}
	for _, vault := range export.Vaults {
		// Filter out the trash vault by convention.
		if strings.EqualFold(strings.TrimSpace(vault.Name), "Recycle Bin") ||
			strings.EqualFold(strings.TrimSpace(vault.Name), "Trash") {
			continue
		}
		for i := range vault.Items {
			item := &vault.Items[i]
			// Trashed items (state == 2) are skipped.
			if item.State == 2 {
				result.Skipped++
				continue
			}
			entry, ok := protonPassConvertItem(item, vault.Name)
			if !ok {
				result.Skipped++
				continue
			}
			result.Entries = append(result.Entries, entry)
		}
	}
	return result, nil
}

func protonPassConvertItem(item *protonPassItem, folder string) (ImportedEntry, bool) {
	common := ImportedEntry{
		Title:  strings.TrimSpace(item.Data.Metadata.Name),
		Notes:  strings.TrimSpace(item.Data.Metadata.Note),
		Folder: folder,
		Source: "protonpass",
	}

	switch item.Data.Type {
	case "login":
		var c protonPassLoginContent
		if err := json.Unmarshal(item.Data.Content, &c); err != nil {
			return ImportedEntry{}, false
		}
		// Proton's quirk: the actual login is in `itemEmail` when `username`
		// is empty (which is the common case for new accounts).
		username := strings.TrimSpace(c.Username)
		if username == "" {
			username = strings.TrimSpace(c.ItemEmail)
		}
		common.Type = model.EntryTypePassword
		common.Username = username
		common.Password = []byte(c.Password)
		common.URLs = c.URLs
		common.TOTP = c.TOTPURI

		// Custom fields are stored in extraFields with various types.
		if extras := flattenProtonExtraFields(item.Data.Extra); len(extras) > 0 {
			common.Fields = extras
		}
		return common, true

	case "note":
		common.Type = model.EntryTypeNote
		return common, true

	case "alias":
		var c protonPassAliasContent
		_ = json.Unmarshal(item.Data.Content, &c)
		// Aliases are e-mail forwarding addresses; record them as notes so
		// the user does not lose track of which alias was associated with
		// which service.
		common.Type = model.EntryTypeNote
		alias := strings.TrimSpace(c.AliasEmail)
		if alias == "" {
			return ImportedEntry{}, false
		}
		if common.Notes != "" {
			common.Notes += "\n\n"
		}
		common.Notes += "Alias e-mail: " + alias
		return common, true

	case "creditCard", "card":
		var c protonPassCardContent
		if err := json.Unmarshal(item.Data.Content, &c); err != nil {
			return ImportedEntry{}, false
		}
		subtype := "credit"
		if strings.EqualFold(c.CardType, "debit") {
			subtype = "debit"
		}
		month, year := splitProtonExpiry(c.ExpirationDate)
		common.Type = model.EntryTypeCard
		common.Card = &CardData{
			Subtype:  subtype,
			Holder:   c.CardholderName,
			Number:   []byte(c.Number),
			ExpMonth: month,
			ExpYear:  year,
			CVV:      []byte(c.VerificationNumber),
		}
		return common, true

	default:
		return ImportedEntry{}, false
	}
}

// flattenProtonExtraFields walks Proton's extraFields[] and copies every
// non-secret value into a flat map[string]string so the mapper can fold it
// into the encrypted notes payload. Hidden fields (type == "hidden") are
// preserved as "(hidden value)" rather than copied verbatim, mirroring the
// Bitwarden handling — we do not want to make a second plaintext copy of a
// secret in a non-secret field.
func flattenProtonExtraFields(extras []protonPassExtraField) map[string]string {
	if len(extras) == 0 {
		return nil
	}
	out := make(map[string]string, len(extras))
	for _, f := range extras {
		name := strings.TrimSpace(f.FieldName)
		if name == "" {
			continue
		}
		var raw json.RawMessage
		if v, ok := f.Data["content"]; ok {
			raw = v
		} else if v, ok := f.Data["value"]; ok {
			raw = v
		}
		var value string
		if len(raw) > 0 {
			_ = json.Unmarshal(raw, &value)
		}
		switch strings.ToLower(f.Type) {
		case "hidden", "totp":
			out[name] = "(hidden value)"
		default:
			if value != "" {
				out[name] = value
			}
		}
	}
	return out
}

// splitProtonExpiry handles Proton's "MM/YY", "MM/YYYY" and "YYYY-MM" forms.
func splitProtonExpiry(raw string) (month, year string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ""
	}
	for _, sep := range []string{"/", "-"} {
		if i := strings.Index(raw, sep); i > 0 {
			a := strings.TrimSpace(raw[:i])
			b := strings.TrimSpace(raw[i+1:])
			// If the first component is 4 digits it is the year (YYYY-MM).
			if len(a) == 4 {
				return b, a
			}
			return a, b
		}
	}
	return raw, ""
}
