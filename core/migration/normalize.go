package migration

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// NormalizeURL ensures a URL string has a scheme and is syntactically valid.
// An empty or unparseable input returns "" (callers treat that as "no URL").
func NormalizeURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	// Schemes other than http/https (otpauth://, ftp://, ...) are preserved.
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return ""
	}
	return u.String()
}

// DomainOf returns the bare host (without scheme, port, www. prefix or path)
// of a URL string. Returns "" if the input has no recognizable host.
func DomainOf(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	host := u.Hostname()
	host = strings.TrimPrefix(strings.ToLower(host), "www.")
	return host
}

// DedupURLs normalizes and de-duplicates a URL list, preserving first-seen
// order. Empty entries are dropped.
func DedupURLs(urls []string) []string {
	seen := make(map[string]struct{}, len(urls))
	out := make([]string, 0, len(urls))
	for _, raw := range urls {
		u := NormalizeURL(raw)
		if u == "" {
			continue
		}
		if _, ok := seen[u]; ok {
			continue
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}
	return out
}

// DeriveTitle picks a reasonable display title. Preference order:
//
//  1. Explicit title if non-empty.
//  2. Domain of the first URL.
//  3. Username if present.
//  4. "(untitled)" as a last resort.
func DeriveTitle(title string, urls []string, username string) string {
	if t := strings.TrimSpace(title); t != "" {
		return t
	}
	for _, raw := range urls {
		if d := DomainOf(raw); d != "" {
			return d
		}
	}
	if u := strings.TrimSpace(username); u != "" {
		return u
	}
	return "(untitled)"
}

// DeriveServiceName produces the value stored in VaultEntry.Service for a
// password entry. The title is used as-is; if there is a primary URL whose
// domain is not already implied by the title, the domain is appended in
// parentheses for disambiguation, e.g. "MyBank (chase.com)".
func DeriveServiceName(title string, urls []string) string {
	t := strings.TrimSpace(title)
	if len(urls) == 0 {
		return t
	}
	domain := DomainOf(urls[0])
	if domain == "" {
		return t
	}
	if t == "" {
		return domain
	}
	// Suppress redundancy when the title already names the domain.
	lowerT := strings.ToLower(t)
	if lowerT == domain || strings.Contains(lowerT, domain) {
		return t
	}
	return fmt.Sprintf("%s (%s)", t, domain)
}

// BuildNotesPayload assembles the encrypted notes content for a password
// entry. It folds extra URLs, custom fields and the original folder name into
// the plaintext so the information is preserved even though VaultEntry has no
// dedicated columns for them. The order is stable so round-trips are clean.
func BuildNotesPayload(notes string, urls []string, folder string, fields map[string]string) string {
	var b strings.Builder
	if n := strings.TrimSpace(notes); n != "" {
		b.WriteString(n)
	}

	if len(urls) > 0 {
		if b.Len() > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString("URLs:\n")
		for _, u := range urls {
			b.WriteString("  ")
			b.WriteString(u)
			b.WriteByte('\n')
		}
	}

	if f := strings.TrimSpace(folder); f != "" {
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString("Folder: ")
		b.WriteString(f)
		b.WriteByte('\n')
	}

	if len(fields) > 0 {
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString("--- Custom Fields ---\n")
		keys := make([]string, 0, len(fields))
		for k := range fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			b.WriteString(k)
			b.WriteString(": ")
			b.WriteString(fields[k])
			b.WriteByte('\n')
		}
	}

	return strings.TrimRight(b.String(), "\n")
}

// IsOTPAuthURI reports whether s looks like an otpauth:// URI.
func IsOTPAuthURI(s string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(s)), "otpauth://")
}

// NormalizeTOTP turns either a raw base32 secret or an otpauth URI into a
// canonical otpauth://totp/... URI. Returns "" if the input is empty.
//
// Detailed TOTP parameter parsing happens in totp.ParseOTPAuthURI later in
// the pipeline; this function only ensures the value reaching the mapper
// is in URI form so the mapper can call ParseOTPAuthURI uniformly.
func NormalizeTOTP(raw, issuer, account string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if IsOTPAuthURI(raw) {
		return raw
	}
	// Bare base32 secret: synthesize a minimal otpauth URI.
	secret := strings.ReplaceAll(raw, " ", "")
	label := strings.TrimSpace(account)
	if label == "" {
		label = "imported"
	}
	if iss := strings.TrimSpace(issuer); iss != "" {
		label = iss + ":" + label
	}
	values := url.Values{}
	values.Set("secret", secret)
	if iss := strings.TrimSpace(issuer); iss != "" {
		values.Set("issuer", iss)
	}
	return "otpauth://totp/" + url.PathEscape(label) + "?" + values.Encode()
}
