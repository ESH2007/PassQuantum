package totp

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type Algorithm string

const (
	AlgorithmSHA1   Algorithm = "SHA1"
	AlgorithmSHA256 Algorithm = "SHA256"
	AlgorithmSHA512 Algorithm = "SHA512"
)

type TOTPParams struct {
	Secret    string    `json:"secret"`
	Algorithm Algorithm `json:"algorithm"`
	Digits    int       `json:"digits"`
	Period    int       `json:"period"`
	Issuer    string    `json:"issuer"`
	Account   string    `json:"account"`
}

func (a Algorithm) toOTP() otp.Algorithm {
	switch a {
	case AlgorithmSHA256:
		return otp.AlgorithmSHA256
	case AlgorithmSHA512:
		return otp.AlgorithmSHA512
	default:
		return otp.AlgorithmSHA1
	}
}

func algorithmFromOTP(a otp.Algorithm) Algorithm {
	switch a {
	case otp.AlgorithmSHA256:
		return AlgorithmSHA256
	case otp.AlgorithmSHA512:
		return AlgorithmSHA512
	default:
		return AlgorithmSHA1
	}
}

func digitsFromOTP(d otp.Digits) int {
	switch d {
	case otp.DigitsEight:
		return 8
	default:
		return 6
	}
}

func (p *TOTPParams) otpDigits() otp.Digits {
	switch p.Digits {
	case 8:
		return otp.DigitsEight
	default:
		return otp.DigitsSix
	}
}

// GenerateCode produces the current TOTP code and the seconds remaining
// until the code expires.
func GenerateCode(params *TOTPParams) (code string, remaining int, err error) {
	if err := Validate(params); err != nil {
		return "", 0, err
	}

	now := time.Now()
	code, err = totp.GenerateCodeCustom(params.Secret, now, totp.ValidateOpts{
		Period:    uint(params.Period),
		Digits:    params.otpDigits(),
		Algorithm: params.Algorithm.toOTP(),
	})
	if err != nil {
		return "", 0, fmt.Errorf("totp generate: %w", err)
	}

	remaining = params.Period - int(now.Unix()%int64(params.Period))
	return code, remaining, nil
}

// ParseOTPAuthURI parses an otpauth://totp/... URI into TOTPParams.
func ParseOTPAuthURI(uri string) (*TOTPParams, error) {
	key, err := otp.NewKeyFromURL(uri)
	if err != nil {
		return nil, fmt.Errorf("parse otpauth URI: %w", err)
	}

	period := 30
	if key.Period() > 0 {
		period = int(key.Period())
	}

	digits := digitsFromOTP(key.Digits())
	// The library only returns 6 or 8; handle 7 from raw URL query if needed.
	if u, err := url.Parse(key.URL()); err == nil {
		if v := u.Query().Get("digits"); v == "7" {
			digits = 7
		}
	}

	return &TOTPParams{
		Secret:    key.Secret(),
		Algorithm: algorithmFromOTP(key.Algorithm()),
		Digits:    digits,
		Period:    period,
		Issuer:    key.Issuer(),
		Account:   key.AccountName(),
	}, nil
}

// Validate checks that a TOTPParams struct has valid fields.
func Validate(params *TOTPParams) error {
	if params.Secret == "" {
		return fmt.Errorf("TOTP secret is required")
	}
	if params.Digits < 6 || params.Digits > 8 {
		return fmt.Errorf("TOTP digits must be 6, 7, or 8 (got %d)", params.Digits)
	}
	if params.Period <= 0 {
		return fmt.Errorf("TOTP period must be positive (got %d)", params.Period)
	}
	switch params.Algorithm {
	case AlgorithmSHA1, AlgorithmSHA256, AlgorithmSHA512, "":
	default:
		return fmt.Errorf("unsupported TOTP algorithm: %s", params.Algorithm)
	}
	return nil
}

// DefaultParams returns sensible defaults for a new TOTP entry.
func DefaultParams() *TOTPParams {
	return &TOTPParams{
		Algorithm: AlgorithmSHA1,
		Digits:    6,
		Period:    30,
	}
}

// Serialize encodes TOTPParams to JSON for encrypted storage.
func Serialize(params *TOTPParams) ([]byte, error) {
	return json.Marshal(params)
}

// Deserialize decodes JSON back to TOTPParams.
func Deserialize(data []byte) (*TOTPParams, error) {
	var p TOTPParams
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("deserialize TOTP params: %w", err)
	}
	return &p, nil
}

// FormatCode inserts a space in the middle of the code for readability.
// "123456" -> "123 456", "12345678" -> "1234 5678".
func FormatCode(code string) string {
	mid := len(code) / 2
	if mid == 0 {
		return code
	}
	return code[:mid] + " " + code[mid:]
}
