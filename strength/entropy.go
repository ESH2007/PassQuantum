package strength

import (
	"fmt"
	"math"
	"unicode"
)

// CalcEntropy estimates password entropy in bits using the higher of Shannon
// entropy and a charset-size estimate.
func CalcEntropy(password string) float64 {
	runes := []rune(password)
	if len(runes) == 0 {
		return 0
	}

	freq := make(map[rune]int)
	for _, r := range runes {
		freq[r]++
	}

	shannonPerRune := 0.0
	length := float64(len(runes))
	for _, count := range freq {
		p := float64(count) / length
		shannonPerRune -= p * math.Log2(p)
	}
	shannonBits := shannonPerRune * length

	charsetSize := estimatedCharsetSize(runes)
	charsetBits := math.Log2(float64(charsetSize)) * length

	if charsetBits > shannonBits {
		return charsetBits
	}
	return shannonBits
}

// EstimateCrackTime returns a human-readable brute-force estimate assuming
// 10 billion guesses per second.
func EstimateCrackTime(entropy float64) string {
	if entropy <= 0 {
		return "Instantly"
	}

	seconds := math.Exp2(math.Min(entropy, 1024)) / 1e10
	if seconds < 1 {
		return "Instantly"
	}
	if seconds < 60 {
		return formatUnit(seconds, 1, "second")
	}
	if seconds < 3600 {
		return formatUnit(seconds/60, 1, "minute")
	}
	if seconds < 86400 {
		return formatUnit(seconds/3600, 1, "hour")
	}
	years := seconds / (365 * 24 * 3600)
	if years < 1 {
		return formatUnit(seconds/86400, 1, "day")
	}
	if years >= 1000 {
		return "Practically unbreakable"
	}
	if years >= 100 {
		return formatUnit(years/100, 1, "century")
	}
	return formatUnit(years, 1, "year")
}

func estimatedCharsetSize(runes []rune) int {
	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSpecial := false
	extra := make(map[rune]struct{})

	for _, r := range runes {
		switch {
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsLetter(r):
			extra[r] = struct{}{}
		default:
			hasSpecial = true
		}
	}

	size := 0
	if hasLower {
		size += 26
	}
	if hasUpper {
		size += 26
	}
	if hasDigit {
		size += 10
	}
	if hasSpecial {
		size += 33
	}
	size += len(extra)
	if size == 0 {
		size = len(runes)
	}
	return size
}

func formatUnit(value float64, min float64, singular string) string {
	count := int(math.Ceil(math.Max(value, min)))
	if count == 1 {
		return fmt.Sprintf("%d %s", count, singular)
	}
	return fmt.Sprintf("%d %ss", count, singular)
}
