package migration

import (
	"strings"
	"testing"
)

func TestDomainOf(t *testing.T) {
	cases := map[string]string{
		"https://github.com":             "github.com",
		"https://www.github.com/login":   "github.com",
		"http://example.com:8080/x":      "example.com",
		"chase.com":                      "chase.com",
		"www.example.com":                "example.com",
		"":                               "",
		"not a url at all":               "",
	}
	for in, want := range cases {
		if got := DomainOf(in); got != want {
			t.Errorf("DomainOf(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDeriveServiceName_AppendsDomainWhenDifferent(t *testing.T) {
	got := DeriveServiceName("My Bank", []string{"https://chase.com/login"})
	if got != "My Bank (chase.com)" {
		t.Errorf("got %q", got)
	}
}

func TestDeriveServiceName_NoAppendWhenTitleContainsDomain(t *testing.T) {
	got := DeriveServiceName("github.com personal", []string{"https://github.com"})
	if strings.Contains(got, "(") {
		t.Errorf("should not append domain when title contains it: %q", got)
	}
}

func TestDeriveTitle_FallsBackToHost(t *testing.T) {
	got := DeriveTitle("", []string{"https://accounts.google.com/login"}, "")
	if got != "accounts.google.com" {
		t.Errorf("got %q", got)
	}
}

func TestDedupURLs(t *testing.T) {
	in := []string{
		"https://github.com",
		"github.com",
		"  ",
		"https://github.com",
		"https://other.com",
	}
	out := DedupURLs(in)
	if len(out) != 2 {
		t.Fatalf("expected 2 unique URLs, got %d: %v", len(out), out)
	}
}

func TestNormalizeTOTP_BareSecretWrappedToURI(t *testing.T) {
	got := NormalizeTOTP("JBSWY3DPEHPK3PXP", "Gmail", "alice@example.com")
	if !strings.HasPrefix(got, "otpauth://totp/") {
		t.Errorf("not an otpauth URI: %q", got)
	}
	if !strings.Contains(got, "secret=JBSWY3DPEHPK3PXP") {
		t.Errorf("missing secret param: %q", got)
	}
	if !strings.Contains(got, "issuer=Gmail") {
		t.Errorf("missing issuer param: %q", got)
	}
}

func TestNormalizeTOTP_OTPAuthURIPassedThrough(t *testing.T) {
	uri := "otpauth://totp/Gmail:a@b?secret=XYZ"
	if got := NormalizeTOTP(uri, "", ""); got != uri {
		t.Errorf("otpauth URI should be passed through: got %q", got)
	}
}

func TestBuildNotesPayload(t *testing.T) {
	out := BuildNotesPayload(
		"main notes here",
		[]string{"https://x.com", "https://y.com"},
		"My Folder",
		map[string]string{"recovery": "abc-def"},
	)
	for _, want := range []string{
		"main notes here",
		"https://x.com",
		"https://y.com",
		"Folder: My Folder",
		"recovery: abc-def",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in payload:\n%s", want, out)
		}
	}
}
