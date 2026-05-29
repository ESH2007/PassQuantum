package app

import (
	"testing"

	"passquantum/core/model"
)

func makePasswordEntry(service, username string) *model.VaultEntry {
	return &model.VaultEntry{
		ID:       1,
		Type:     model.EntryTypePassword,
		Service:  service,
		Username: username,
	}
}

func makeTOTPEntry(issuer, account string) *model.VaultEntry {
	return &model.VaultEntry{
		ID:       1,
		Type:     model.EntryTypeTOTP,
		Service:  "TOTP:" + issuer,
		Username: account,
	}
}

func TestFindDuplicateEntry_PasswordMatch(t *testing.T) {
	entries := []*model.VaultEntry{
		makePasswordEntry("github.com", "alice@example.com"),
	}

	tests := []struct {
		name             string
		queryService     string
		queryUsername    string
		wantMatch        bool
	}{
		{"exact match", "github.com", "alice@example.com", true},
		{"case-insensitive service", "GitHub.com", "alice@example.com", true},
		{"case-insensitive username", "github.com", "ALICE@example.com", true},
		{"normalized URL → host", "https://github.com/login", "alice@example.com", true},
		{"www. prefix stripped", "www.github.com", "alice@example.com", true},
		{"different username", "github.com", "bob@example.com", false},
		{"different service", "gitlab.com", "alice@example.com", false},
		{"trimmed whitespace", "  github.com  ", "  alice@example.com  ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindDuplicateEntry(entries, model.EntryTypePassword, tt.queryService, tt.queryUsername)
			if tt.wantMatch && got == nil {
				t.Errorf("expected match for %q/%q, got nil", tt.queryService, tt.queryUsername)
			}
			if !tt.wantMatch && got != nil {
				t.Errorf("expected no match for %q/%q, got entry", tt.queryService, tt.queryUsername)
			}
		})
	}
}

func TestFindDuplicateEntry_TOTPMatch(t *testing.T) {
	entries := []*model.VaultEntry{
		makeTOTPEntry("GitHub", "alice@example.com"),
	}

	tests := []struct {
		name           string
		queryService   string
		queryUsername  string
		wantMatch      bool
	}{
		{"without TOTP: prefix", "GitHub", "alice@example.com", true},
		{"with TOTP: prefix", "TOTP:GitHub", "alice@example.com", true},
		{"case-insensitive issuer", "github", "alice@example.com", true},
		{"case-insensitive account", "GitHub", "ALICE@example.com", true},
		{"different issuer", "GitLab", "alice@example.com", false},
		{"different account", "GitHub", "bob@example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindDuplicateEntry(entries, model.EntryTypeTOTP, tt.queryService, tt.queryUsername)
			if tt.wantMatch && got == nil {
				t.Errorf("expected match for %q/%q, got nil", tt.queryService, tt.queryUsername)
			}
			if !tt.wantMatch && got != nil {
				t.Errorf("expected no match for %q/%q, got entry", tt.queryService, tt.queryUsername)
			}
		})
	}
}

func TestFindDuplicateEntry_TypeMismatch(t *testing.T) {
	// A password entry with service "GitHub" should NOT match a TOTP query
	// for "GitHub" (different types are independent namespaces).
	entries := []*model.VaultEntry{
		makePasswordEntry("GitHub", "alice@example.com"),
	}

	got := FindDuplicateEntry(entries, model.EntryTypeTOTP, "GitHub", "alice@example.com")
	if got != nil {
		t.Error("password entry incorrectly matched TOTP query")
	}
}

func TestFindDuplicateEntry_UnsupportedTypes(t *testing.T) {
	entries := []*model.VaultEntry{
		{ID: 1, Type: model.EntryTypeNote, Service: "Server IPs", Username: ""},
		{ID: 2, Type: model.EntryTypeCard, Service: "Visa", Username: ""},
		{ID: 3, Type: model.EntryTypeFile, Service: "FILE:resume.pdf", Username: ""},
	}

	for _, tp := range []model.EntryType{model.EntryTypeNote, model.EntryTypeCard, model.EntryTypeFile} {
		got := FindDuplicateEntry(entries, tp, "anything", "")
		if got != nil {
			t.Errorf("type %v should return nil (no natural dedup key), got entry", tp)
		}
	}
}

func TestFindDuplicateEntry_EmptyEntries(t *testing.T) {
	got := FindDuplicateEntry(nil, model.EntryTypePassword, "github.com", "alice")
	if got != nil {
		t.Error("expected nil for empty entries slice")
	}

	got = FindDuplicateEntry([]*model.VaultEntry{}, model.EntryTypePassword, "github.com", "alice")
	if got != nil {
		t.Error("expected nil for empty entries slice")
	}
}

func TestFindDuplicateEntry_NilEntryInSlice(t *testing.T) {
	entries := []*model.VaultEntry{
		nil,
		makePasswordEntry("github.com", "alice"),
	}
	got := FindDuplicateEntry(entries, model.EntryTypePassword, "github.com", "alice")
	if got == nil {
		t.Error("expected match despite nil entry in slice")
	}
}
