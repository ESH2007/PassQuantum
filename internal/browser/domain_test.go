package browser

import "testing"

func TestNormalizeDomain(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"https://mail.google.com/inbox", "google.com"},
		{"https://www.google.com", "google.com"},
		{"github.com", "github.com"},
		{"https://github.com/login", "github.com"},
		{"https://www.amazon.co.uk/dp/B08N5WRWNW", "amazon.co.uk"},
		{"https://accounts.google.com/signin/v2/identifier", "google.com"},
		{"https://login.microsoftonline.com/common/oauth2", "microsoftonline.com"},
		{"localhost:3000", "localhost"},
		{"localhost", "localhost"},
		{"192.168.1.1", "192.168.1.1"},
		{"http://192.168.1.1:8080/admin", "192.168.1.1"},
		{"https://app.example.com", "example.com"},
		{"", ""},
		{"   ", ""},
		{"https://subdomain.example.org/path?query=1", "example.org"},
		{"https://bbc.co.uk/news", "bbc.co.uk"},
		{"notion.so", "notion.so"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := NormalizeDomain(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeDomain(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
