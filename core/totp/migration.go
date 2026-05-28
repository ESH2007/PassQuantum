package totp

import (
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
)

// ParseGoogleAuthExport decodes a Google Authenticator migration payload
// (otpauth-migration://offline?data=...) and returns the contained TOTP entries.
//
// The data parameter is a base64-encoded protocol buffer with a well-known
// schema. This implementation uses a minimal hand-written protobuf decoder
// to avoid pulling in the full protobuf dependency.
func ParseGoogleAuthExport(migrationURI string) ([]*TOTPParams, error) {
	u, err := url.Parse(migrationURI)
	if err != nil {
		return nil, fmt.Errorf("parse migration URI: %w", err)
	}

	data := u.Query().Get("data")
	if data == "" {
		return nil, fmt.Errorf("migration URI missing 'data' parameter")
	}

	raw, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		raw, err = base64.RawStdEncoding.DecodeString(data)
		if err != nil {
			return nil, fmt.Errorf("base64 decode: %w", err)
		}
	}

	return parseMigrationPayload(raw)
}

// parseMigrationPayload decodes the protobuf wire format used by Google
// Authenticator's export feature.
//
// Wire format (simplified, field numbers from the known schema):
//
//	message MigrationPayload {
//	  repeated OtpParameters otp_parameters = 1;
//	}
//
//	message OtpParameters {
//	  bytes  secret   = 1;
//	  string name     = 2;
//	  string issuer   = 3;
//	  int32  algorithm = 4;  // 0=UNSPECIFIED, 1=SHA1, 2=SHA256, 3=SHA512
//	  int32  digits    = 5;  // 0=UNSPECIFIED, 1=SIX, 2=EIGHT
//	  int32  type      = 6;  // 0=UNSPECIFIED, 1=HOTP, 2=TOTP
//	}
func parseMigrationPayload(data []byte) ([]*TOTPParams, error) {
	var results []*TOTPParams
	pos := 0

	for pos < len(data) {
		fieldNum, wireType, n, err := readTag(data[pos:])
		if err != nil {
			return results, err
		}
		pos += n

		if fieldNum == 1 && wireType == 2 {
			msgLen, n, err := readVarint(data[pos:])
			if err != nil {
				return results, err
			}
			pos += n

			if pos+int(msgLen) > len(data) {
				return results, fmt.Errorf("truncated otp_parameters message")
			}

			params, err := parseOTPParameters(data[pos : pos+int(msgLen)])
			if err != nil {
				return results, fmt.Errorf("parse otp_parameters: %w", err)
			}
			if params != nil {
				results = append(results, params)
			}
			pos += int(msgLen)
		} else {
			n, err := skipField(data[pos:], wireType)
			if err != nil {
				return results, err
			}
			pos += n
		}
	}

	return results, nil
}

func parseOTPParameters(data []byte) (*TOTPParams, error) {
	p := DefaultParams()
	pos := 0

	for pos < len(data) {
		fieldNum, wireType, n, err := readTag(data[pos:])
		if err != nil {
			return nil, err
		}
		pos += n

		switch {
		case fieldNum == 1 && wireType == 2: // secret (bytes)
			bLen, n, err := readVarint(data[pos:])
			if err != nil {
				return nil, err
			}
			pos += n
			if pos+int(bLen) > len(data) {
				return nil, fmt.Errorf("truncated secret")
			}
			p.Secret = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(data[pos : pos+int(bLen)])
			pos += int(bLen)

		case fieldNum == 2 && wireType == 2: // name (string)
			sLen, n, err := readVarint(data[pos:])
			if err != nil {
				return nil, err
			}
			pos += n
			if pos+int(sLen) > len(data) {
				return nil, fmt.Errorf("truncated name")
			}
			name := string(data[pos : pos+int(sLen)])
			if idx := strings.LastIndex(name, ":"); idx >= 0 {
				p.Account = strings.TrimSpace(name[idx+1:])
				if p.Issuer == "" {
					p.Issuer = strings.TrimSpace(name[:idx])
				}
			} else {
				p.Account = name
			}
			pos += int(sLen)

		case fieldNum == 3 && wireType == 2: // issuer (string)
			sLen, n, err := readVarint(data[pos:])
			if err != nil {
				return nil, err
			}
			pos += n
			if pos+int(sLen) > len(data) {
				return nil, fmt.Errorf("truncated issuer")
			}
			p.Issuer = string(data[pos : pos+int(sLen)])
			pos += int(sLen)

		case fieldNum == 4 && wireType == 0: // algorithm (varint)
			v, n, err := readVarint(data[pos:])
			if err != nil {
				return nil, err
			}
			pos += n
			switch v {
			case 2:
				p.Algorithm = AlgorithmSHA256
			case 3:
				p.Algorithm = AlgorithmSHA512
			default:
				p.Algorithm = AlgorithmSHA1
			}

		case fieldNum == 5 && wireType == 0: // digits (varint)
			v, n, err := readVarint(data[pos:])
			if err != nil {
				return nil, err
			}
			pos += n
			switch v {
			case 2:
				p.Digits = 8
			default:
				p.Digits = 6
			}

		case fieldNum == 6 && wireType == 0: // type (varint, 2=TOTP)
			_, n, err := readVarint(data[pos:])
			if err != nil {
				return nil, err
			}
			pos += n

		default:
			n, err := skipField(data[pos:], wireType)
			if err != nil {
				return nil, err
			}
			pos += n
		}
	}

	if p.Secret == "" {
		return nil, nil
	}
	return p, nil
}

func readVarint(data []byte) (uint64, int, error) {
	var val uint64
	for i := 0; i < len(data) && i < 10; i++ {
		b := data[i]
		val |= uint64(b&0x7F) << (7 * i)
		if b&0x80 == 0 {
			return val, i + 1, nil
		}
	}
	return 0, 0, fmt.Errorf("varint overflow or truncated")
}

func readTag(data []byte) (fieldNum int, wireType int, n int, err error) {
	v, n, err := readVarint(data)
	if err != nil {
		return 0, 0, 0, err
	}
	return int(v >> 3), int(v & 0x07), n, nil
}

func skipField(data []byte, wireType int) (int, error) {
	switch wireType {
	case 0: // varint
		_, n, err := readVarint(data)
		return n, err
	case 1: // 64-bit
		if len(data) < 8 {
			return 0, fmt.Errorf("truncated 64-bit field")
		}
		return 8, nil
	case 2: // length-delimited
		bLen, n, err := readVarint(data)
		if err != nil {
			return 0, err
		}
		total := n + int(bLen)
		if total > len(data) {
			return 0, fmt.Errorf("truncated length-delimited field")
		}
		return total, nil
	case 5: // 32-bit
		if len(data) < 4 {
			return 0, fmt.Errorf("truncated 32-bit field")
		}
		return 4, nil
	default:
		return 0, fmt.Errorf("unsupported wire type %d", wireType)
	}
}
