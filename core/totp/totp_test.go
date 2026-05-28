package totp

import (
	"encoding/base32"
	"encoding/base64"
	"testing"
)

func TestGenerateCode_SHA1_6digits_30s(t *testing.T) {
	params := &TOTPParams{
		Secret:    base32.StdEncoding.EncodeToString([]byte("12345678901234567890")),
		Algorithm: AlgorithmSHA1,
		Digits:    6,
		Period:    30,
	}
	code, remaining, err := GenerateCode(params)
	if err != nil {
		t.Fatalf("GenerateCode: %v", err)
	}
	if len(code) != 6 {
		t.Errorf("expected 6-digit code, got %q", code)
	}
	if remaining <= 0 || remaining > 30 {
		t.Errorf("remaining %d out of range [1,30]", remaining)
	}
}

func TestGenerateCode_SHA256(t *testing.T) {
	params := &TOTPParams{
		Secret:    base32.StdEncoding.EncodeToString([]byte("12345678901234567890")),
		Algorithm: AlgorithmSHA256,
		Digits:    6,
		Period:    30,
	}
	code, _, err := GenerateCode(params)
	if err != nil {
		t.Fatalf("GenerateCode SHA256: %v", err)
	}
	if len(code) != 6 {
		t.Errorf("expected 6-digit code, got %q", code)
	}
}

func TestGenerateCode_SHA512_8digits(t *testing.T) {
	params := &TOTPParams{
		Secret:    base32.StdEncoding.EncodeToString([]byte("12345678901234567890")),
		Algorithm: AlgorithmSHA512,
		Digits:    8,
		Period:    60,
	}
	code, remaining, err := GenerateCode(params)
	if err != nil {
		t.Fatalf("GenerateCode SHA512/8: %v", err)
	}
	if len(code) != 8 {
		t.Errorf("expected 8-digit code, got %q", code)
	}
	if remaining <= 0 || remaining > 60 {
		t.Errorf("remaining %d out of range [1,60]", remaining)
	}
}

func TestParseOTPAuthURI_Full(t *testing.T) {
	uri := "otpauth://totp/GitHub:user%40example.com?secret=JBSWY3DPEHPK3PXP&issuer=GitHub&algorithm=SHA1&digits=6&period=30"
	params, err := ParseOTPAuthURI(uri)
	if err != nil {
		t.Fatalf("ParseOTPAuthURI: %v", err)
	}
	if params.Issuer != "GitHub" {
		t.Errorf("issuer = %q, want GitHub", params.Issuer)
	}
	if params.Account != "user@example.com" {
		t.Errorf("account = %q, want user@example.com", params.Account)
	}
	if params.Secret != "JBSWY3DPEHPK3PXP" {
		t.Errorf("secret = %q, want JBSWY3DPEHPK3PXP", params.Secret)
	}
	if params.Algorithm != AlgorithmSHA1 {
		t.Errorf("algorithm = %q, want SHA1", params.Algorithm)
	}
	if params.Digits != 6 {
		t.Errorf("digits = %d, want 6", params.Digits)
	}
	if params.Period != 30 {
		t.Errorf("period = %d, want 30", params.Period)
	}
}

func TestParseOTPAuthURI_Minimal(t *testing.T) {
	uri := "otpauth://totp/MyService?secret=JBSWY3DPEHPK3PXP"
	params, err := ParseOTPAuthURI(uri)
	if err != nil {
		t.Fatalf("ParseOTPAuthURI minimal: %v", err)
	}
	if params.Secret != "JBSWY3DPEHPK3PXP" {
		t.Errorf("secret = %q", params.Secret)
	}
	if params.Digits != 6 {
		t.Errorf("digits = %d, want 6 (default)", params.Digits)
	}
	if params.Period != 30 {
		t.Errorf("period = %d, want 30 (default)", params.Period)
	}
	if params.Algorithm != AlgorithmSHA1 {
		t.Errorf("algorithm = %q, want SHA1 (default)", params.Algorithm)
	}
}

func TestSerializeDeserializeRoundTrip(t *testing.T) {
	original := &TOTPParams{
		Secret:    "JBSWY3DPEHPK3PXP",
		Algorithm: AlgorithmSHA256,
		Digits:    8,
		Period:    60,
		Issuer:    "TestIssuer",
		Account:   "test@example.com",
	}
	data, err := Serialize(original)
	if err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	restored, err := Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if restored.Secret != original.Secret ||
		restored.Algorithm != original.Algorithm ||
		restored.Digits != original.Digits ||
		restored.Period != original.Period ||
		restored.Issuer != original.Issuer ||
		restored.Account != original.Account {
		t.Errorf("round-trip mismatch:\n  got  %+v\n  want %+v", restored, original)
	}
}

func TestValidate_EmptySecret(t *testing.T) {
	p := &TOTPParams{Secret: "", Digits: 6, Period: 30}
	if err := Validate(p); err == nil {
		t.Error("expected error for empty secret")
	}
}

func TestValidate_BadDigits(t *testing.T) {
	p := &TOTPParams{Secret: "JBSWY3DPEHPK3PXP", Digits: 4, Period: 30}
	if err := Validate(p); err == nil {
		t.Error("expected error for digits=4")
	}
}

func TestValidate_BadPeriod(t *testing.T) {
	p := &TOTPParams{Secret: "JBSWY3DPEHPK3PXP", Digits: 6, Period: 0}
	if err := Validate(p); err == nil {
		t.Error("expected error for period=0")
	}
}

func TestValidate_BadAlgorithm(t *testing.T) {
	p := &TOTPParams{Secret: "JBSWY3DPEHPK3PXP", Digits: 6, Period: 30, Algorithm: "MD5"}
	if err := Validate(p); err == nil {
		t.Error("expected error for algorithm=MD5")
	}
}

func TestFormatCode(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"123456", "123 456"},
		{"12345678", "1234 5678"},
		{"", ""},
		{"1", "1"},
	}
	for _, tc := range tests {
		if got := FormatCode(tc.in); got != tc.want {
			t.Errorf("FormatCode(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestParseMigrationPayload(t *testing.T) {
	// Build a minimal protobuf payload matching Google Auth export format.
	// Field 1 (otp_parameters), wire type 2 (length-delimited).
	//   Inner fields: 1=secret(bytes), 2=name(string), 3=issuer(string),
	//                 4=algo(varint), 5=digits(varint), 6=type(varint)
	secret := []byte("TestSecret1234")
	name := "GitHub:user@example.com"
	issuer := "GitHub"

	inner := encodeBytes(1, secret)
	inner = append(inner, encodeString(2, name)...)
	inner = append(inner, encodeString(3, issuer)...)
	inner = append(inner, encodeVarintField(4, 1)...) // SHA1
	inner = append(inner, encodeVarintField(5, 1)...) // SIX digits
	inner = append(inner, encodeVarintField(6, 2)...) // TOTP

	outer := encodeBytes(1, inner)

	data := base64.StdEncoding.EncodeToString(outer)
	uri := "otpauth-migration://offline?data=" + data

	results, err := ParseGoogleAuthExport(uri)
	if err != nil {
		t.Fatalf("ParseGoogleAuthExport: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	p := results[0]
	expectedSecret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secret)
	if p.Secret != expectedSecret {
		t.Errorf("secret = %q, want %q", p.Secret, expectedSecret)
	}
	if p.Issuer != "GitHub" {
		t.Errorf("issuer = %q, want GitHub", p.Issuer)
	}
	if p.Account != "user@example.com" {
		t.Errorf("account = %q, want user@example.com", p.Account)
	}
	if p.Algorithm != AlgorithmSHA1 {
		t.Errorf("algorithm = %q, want SHA1", p.Algorithm)
	}
	if p.Digits != 6 {
		t.Errorf("digits = %d, want 6", p.Digits)
	}
}

// protobuf encoding helpers for tests

func encodeVarint(v uint64) []byte {
	var buf []byte
	for v >= 0x80 {
		buf = append(buf, byte(v&0x7F)|0x80)
		v >>= 7
	}
	buf = append(buf, byte(v))
	return buf
}

func encodeVarintField(fieldNum int, v uint64) []byte {
	tag := encodeVarint(uint64(fieldNum<<3 | 0))
	return append(tag, encodeVarint(v)...)
}

func encodeBytes(fieldNum int, data []byte) []byte {
	tag := encodeVarint(uint64(fieldNum<<3 | 2))
	length := encodeVarint(uint64(len(data)))
	result := append(tag, length...)
	return append(result, data...)
}

func encodeString(fieldNum int, s string) []byte {
	return encodeBytes(fieldNum, []byte(s))
}
