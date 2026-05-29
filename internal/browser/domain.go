package browser

import (
	"net"
	"net/url"
	"strings"
)

var knownSecondLevelDomains = map[string]bool{
	"co": true, "com": true, "org": true, "net": true,
	"ac": true, "gov": true, "edu": true, "mil": true,
}

// NormalizeDomain extracts the registrable domain from a URL or hostname.
// "https://mail.google.com/inbox" → "google.com"
// "github.com" → "github.com"
// "https://www.amazon.co.uk/dp/..." → "amazon.co.uk"
// "localhost:3000" → "localhost"
// "192.168.1.1" → "192.168.1.1"
func NormalizeDomain(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}

	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}

	host := u.Hostname()
	if host == "" {
		return raw
	}

	if net.ParseIP(host) != nil {
		return host
	}

	if host == "localhost" {
		return host
	}

	parts := strings.Split(host, ".")
	if len(parts) <= 2 {
		return host
	}

	last := parts[len(parts)-1]
	secondLast := parts[len(parts)-2]

	if len(last) == 2 && knownSecondLevelDomains[secondLast] {
		if len(parts) >= 3 {
			return strings.Join(parts[len(parts)-3:], ".")
		}
		return host
	}

	return strings.Join(parts[len(parts)-2:], ".")
}
