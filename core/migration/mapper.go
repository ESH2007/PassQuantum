package migration

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudflare/circl/kem/kyber/kyber768"

	"passquantum/core/crypto"
	"passquantum/core/model"
	"passquantum/core/totp"
)

// notePayload matches the JSON format produced by ui/screens/main_screen.go
// when the user creates a note manually, so imported notes are readable by
// the existing UI without any decoder changes.
type notePayload struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// cardPayload matches the JSON format used by the manual card form.
type cardPayload struct {
	Subtype string `json:"subtype"`
	Holder  string `json:"holder"`
	Number  string `json:"number"`
	Expiry  string `json:"expiry"`
	CVV     string `json:"cvv"`
}

// MapAndEncrypt converts a slice of parsed ImportedEntry values into
// model.VaultEntry values, encrypting each payload with a fresh Kyber +
// AES-256-GCM envelope identical to the one used by manual entry creation.
//
// existing is the current vault contents; it is used both for duplicate
// detection and for in-place rewrites when dupAction == DupReplace.
//
// The function wipes every secret byte slice it touches in entries once the
// corresponding VaultEntry has been built, regardless of whether the entry
// was kept, skipped or replaced.
//
// result.Warnings collects human-readable, non-secret hints about data that
// could not be stored (e.g. password entries whose notes were diverted into
// companion note entries).
func MapAndEncrypt(
	entries []ImportedEntry,
	pubKey *kyber768.PublicKey,
	existing []*model.VaultEntry,
	dupAction DuplicateAction,
) (*MapResult, error) {
	if pubKey == nil {
		return nil, fmt.Errorf("migration: nil public key")
	}
	result := &MapResult{}

	for i := range entries {
		built, err := buildVaultEntries(&entries[i], pubKey, result)

		// Unconditionally wipe secrets we copied or touched.
		wipeSecrets(&entries[i])

		if err != nil {
			// Do not include any field contents in the error string.
			result.Errors = append(result.Errors,
				fmt.Sprintf("entry %d (%s): %s", i, entries[i].Source, err.Error()))
			continue
		}

		for _, ve := range built {
			dup := findDuplicate(existing, ve.Type, ve.Service, ve.Username)
			if dup != nil {
				switch dupAction {
				case DupSkip:
					result.Skipped++
					continue
				case DupReplace:
					dup.KyberCiphertext = ve.KyberCiphertext
					dup.Nonce = ve.Nonce
					dup.Ciphertext = ve.Ciphertext
					result.Replaced++
					continue
				case DupKeepBoth:
					// fall through and append.
				}
			}
			result.NewEntries = append(result.NewEntries, ve)
		}
	}

	return result, nil
}

// buildVaultEntries produces one or more VaultEntries from an ImportedEntry.
// A password entry that also carries extras (notes/extra URLs/custom fields)
// or a TOTP secret yields multiple entries: the password, an optional
// companion note, and an optional TOTP entry.
func buildVaultEntries(entry *ImportedEntry, pubKey *kyber768.PublicKey, result *MapResult) ([]*model.VaultEntry, error) {
	entry.URLs = DedupURLs(entry.URLs)

	switch entry.Type {
	case model.EntryTypePassword, model.EntryTypeUnknown:
		return buildPasswordWithExtras(entry, pubKey, result)
	case model.EntryTypeNote:
		ve, err := buildNote(entry, pubKey)
		if err != nil {
			return nil, err
		}
		return []*model.VaultEntry{ve}, nil
	case model.EntryTypeCard:
		ve, err := buildCard(entry, pubKey)
		if err != nil {
			return nil, err
		}
		return []*model.VaultEntry{ve}, nil
	case model.EntryTypeTOTP:
		ve, err := buildTOTP(entry, pubKey)
		if err != nil {
			return nil, err
		}
		return []*model.VaultEntry{ve}, nil
	default:
		return nil, fmt.Errorf("unsupported entry type %d", entry.Type)
	}
}

// buildPasswordWithExtras encrypts the password as a raw string payload
// (preserving compatibility with the existing UI decoder, which treats the
// plaintext as the password itself) and then emits companion entries when
// the imported record carries additional data:
//
//   - A separate EntryTypeNote when there are notes, extra URLs, a folder
//     name or custom fields. The note title is "<service> — notes" so it
//     groups visually next to its password.
//   - A separate EntryTypeTOTP when a TOTP secret was attached to the login.
func buildPasswordWithExtras(entry *ImportedEntry, pubKey *kyber768.PublicKey, result *MapResult) ([]*model.VaultEntry, error) {
	out := make([]*model.VaultEntry, 0, 3)

	title := DeriveTitle(entry.Title, entry.URLs, entry.Username)
	service := DeriveServiceName(title, entry.URLs)

	if len(entry.Password) > 0 {
		// Keep the payload as the raw password string so the existing UI
		// can render it without any decoder changes.
		ve, err := encryptEntry(pubKey, entry.Password)
		if err != nil {
			return nil, err
		}
		ve.Type = model.EntryTypePassword
		ve.Service = service
		ve.Username = strings.TrimSpace(entry.Username)
		out = append(out, ve)
	}

	// Collect everything we cannot store on the password itself into a
	// companion note. The primary URL is already represented in the service
	// name, so only the *extra* URLs are folded in.
	extraURLs := entry.URLs
	if len(extraURLs) > 0 {
		extraURLs = extraURLs[1:]
	}
	notesPlain := BuildNotesPayload(entry.Notes, extraURLs, entry.Folder, entry.Fields)

	if notesPlain != "" {
		noteTitle := title + " — notes"
		payload, err := json.Marshal(notePayload{
			Type:    "note",
			Title:   noteTitle,
			Content: notesPlain,
		})
		if err == nil {
			ve, err := encryptEntry(pubKey, payload)
			crypto.WipeBytes(payload)
			if err == nil {
				ve.Type = model.EntryTypeNote
				ve.Service = "NOTE:" + noteTitle
				ve.Username = "note"
				out = append(out, ve)
				if result != nil {
					result.Warnings = append(result.Warnings,
						"created companion note for entry: "+title)
				}
			}
		}
	}

	// Embedded TOTP becomes its own EntryTypeTOTP record.
	if strings.TrimSpace(entry.TOTP) != "" {
		totpEntry := *entry
		totpEntry.Type = model.EntryTypeTOTP
		ve, err := buildTOTP(&totpEntry, pubKey)
		if err != nil {
			if result != nil {
				result.Warnings = append(result.Warnings,
					"could not import TOTP for entry: "+title)
			}
		} else {
			out = append(out, ve)
		}
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("no payload to encrypt")
	}
	return out, nil
}

func buildNote(entry *ImportedEntry, pubKey *kyber768.PublicKey) (*model.VaultEntry, error) {
	title := DeriveTitle(entry.Title, entry.URLs, entry.Username)

	// Append extras (URLs, folder, custom fields) to the visible content so
	// the importer never silently drops information.
	content := BuildNotesPayload(entry.Notes, entry.URLs, entry.Folder, entry.Fields)
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("empty note content")
	}

	payload, err := json.Marshal(notePayload{
		Type:    "note",
		Title:   title,
		Content: content,
	})
	if err != nil {
		return nil, fmt.Errorf("note marshal: %w", err)
	}

	ve, err := encryptEntry(pubKey, payload)
	crypto.WipeBytes(payload)
	if err != nil {
		return nil, err
	}
	ve.Type = model.EntryTypeNote
	ve.Service = "NOTE:" + title
	ve.Username = "note"
	return ve, nil
}

func buildCard(entry *ImportedEntry, pubKey *kyber768.PublicKey) (*model.VaultEntry, error) {
	if entry.Card == nil {
		return nil, fmt.Errorf("card data missing")
	}
	title := DeriveTitle(entry.Title, entry.URLs, entry.Username)

	subtype := strings.ToLower(strings.TrimSpace(entry.Card.Subtype))
	if subtype != "credit" && subtype != "debit" {
		subtype = "credit"
	}

	cp := cardPayload{
		Subtype: subtype,
		Holder:  entry.Card.Holder,
		Number:  string(entry.Card.Number),
		Expiry:  joinExpiry(entry.Card.ExpMonth, entry.Card.ExpYear),
		CVV:     string(entry.Card.CVV),
	}
	payload, err := json.Marshal(cp)
	if err != nil {
		return nil, fmt.Errorf("card marshal: %w", err)
	}

	ve, err := encryptEntry(pubKey, payload)
	crypto.WipeBytes(payload)
	crypto.WipeBytes(entry.Card.Number)
	crypto.WipeBytes(entry.Card.CVV)
	if err != nil {
		return nil, err
	}
	ve.Type = model.EntryTypeCard
	ve.CardSubtype = subtype
	ve.Service = "CARD:" + title
	ve.Username = subtype
	return ve, nil
}

func buildTOTP(entry *ImportedEntry, pubKey *kyber768.PublicKey) (*model.VaultEntry, error) {
	raw := strings.TrimSpace(entry.TOTP)
	if raw == "" {
		return nil, fmt.Errorf("missing TOTP secret")
	}

	var params *totp.TOTPParams
	issuer := strings.TrimSpace(firstNonEmpty(entry.Title, entry.Source))
	account := strings.TrimSpace(entry.Username)

	if IsOTPAuthURI(raw) {
		p, err := totp.ParseOTPAuthURI(raw)
		if err != nil {
			return nil, fmt.Errorf("parse otpauth: %w", err)
		}
		params = p
		// Override empty fields from the URI with values from the entry.
		if params.Issuer == "" {
			params.Issuer = issuer
		}
		if params.Account == "" {
			params.Account = account
		}
	} else {
		params = totp.DefaultParams()
		params.Secret = sanitizeBase32(raw)
		params.Issuer = issuer
		params.Account = account
	}

	if err := totp.Validate(params); err != nil {
		return nil, fmt.Errorf("totp validate: %w", err)
	}

	payload, err := totp.Serialize(params)
	if err != nil {
		return nil, fmt.Errorf("totp serialize: %w", err)
	}

	ve, err := encryptEntry(pubKey, payload)
	crypto.WipeBytes(payload)
	if err != nil {
		return nil, err
	}
	ve.Type = model.EntryTypeTOTP
	ve.Service = "TOTP:" + params.Issuer
	ve.Username = params.Account
	return ve, nil
}

// encryptEntry performs the standard per-entry Kyber + AES-256-GCM envelope
// and returns a VaultEntry populated with the resulting crypto fields. The
// caller is responsible for setting Type, Service, Username and CardSubtype.
func encryptEntry(pubKey *kyber768.PublicKey, plaintext []byte) (*model.VaultEntry, error) {
	ct, ss, err := crypto.Encapsulate(pubKey)
	if err != nil {
		return nil, fmt.Errorf("kyber encapsulate: %w", err)
	}
	nonce, ciphertext, err := crypto.EncryptAES256GCM(string(plaintext), ss)
	crypto.WipeBytes(ss)
	if err != nil {
		return nil, fmt.Errorf("aes encrypt: %w", err)
	}

	ve := model.NewVaultEntry()
	ve.KyberCiphertext = ct
	ve.Nonce = nonce
	ve.Ciphertext = ciphertext
	return ve, nil
}

// findDuplicate is a local copy of app.FindDuplicateEntry's logic, kept here
// so the migration package does not depend on the app package (avoiding an
// import cycle when app/import.go consumes the mapper).
//
// Service strings are normalized into a small set of comparison candidates —
// the title alone, the domain in parentheses, and the bare lowercased value —
// so an imported "GitHub (github.com)" entry collides with either a manual
// "github.com" entry or a manual "GitHub" entry.
func findDuplicate(entries []*model.VaultEntry, entryType model.EntryType, service, username string) *model.VaultEntry {
	if entryType != model.EntryTypePassword && entryType != model.EntryTypeTOTP {
		return nil
	}
	wantKeys := serviceCompareKeys(service, entryType)
	wantUser := strings.ToLower(strings.TrimSpace(username))
	for _, e := range entries {
		if e == nil || e.Type != entryType {
			continue
		}
		gotKeys := serviceCompareKeys(e.Service, entryType)
		gotUser := strings.ToLower(strings.TrimSpace(e.Username))
		if gotUser != wantUser {
			continue
		}
		for _, gk := range gotKeys {
			for _, wk := range wantKeys {
				if gk != "" && gk == wk {
					return e
				}
			}
		}
	}
	return nil
}

// serviceCompareKeys produces every form of the service string that should be
// treated as the same identity for dedup purposes:
//
//   - The whole string, lowercased and stripped of TOTP: prefix.
//   - The text portion before any " (...)" annotation.
//   - The text portion inside the trailing parentheses, treated as a domain.
//   - A domain extracted via DomainOf if the string itself is URL-shaped.
//
// Empty strings are omitted.
func serviceCompareKeys(service string, entryType model.EntryType) []string {
	s := strings.TrimSpace(service)
	if entryType == model.EntryTypeTOTP {
		s = strings.TrimPrefix(s, "TOTP:")
		s = strings.TrimPrefix(s, "totp:")
	}
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return nil
	}

	keys := []string{s}
	if d := DomainOf(s); d != "" && d != s {
		keys = append(keys, d)
	}
	if idx := strings.LastIndex(s, " ("); idx > 0 && strings.HasSuffix(s, ")") {
		base := strings.TrimSpace(s[:idx])
		inner := strings.TrimSpace(s[idx+2 : len(s)-1])
		if base != "" {
			keys = append(keys, base)
		}
		if inner != "" {
			keys = append(keys, inner)
		}
	}
	return keys
}

func wipeSecrets(e *ImportedEntry) {
	crypto.WipeBytes(e.Password)
	e.Password = nil
	if e.Card != nil {
		crypto.WipeBytes(e.Card.Number)
		crypto.WipeBytes(e.Card.CVV)
		e.Card.Number = nil
		e.Card.CVV = nil
	}
}

func joinExpiry(month, year string) string {
	m := strings.TrimSpace(month)
	y := strings.TrimSpace(year)
	switch {
	case m != "" && y != "":
		return fmt.Sprintf("%s/%s", m, y)
	case m != "":
		return m
	case y != "":
		return y
	default:
		return ""
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}

// sanitizeBase32 strips whitespace and dashes that some exporters embed in
// base32 secrets for human readability.
func sanitizeBase32(raw string) string {
	var b strings.Builder
	for _, r := range raw {
		if r == ' ' || r == '\t' || r == '-' || r == '\n' || r == '\r' {
			continue
		}
		b.WriteRune(r)
	}
	return strings.ToUpper(b.String())
}
