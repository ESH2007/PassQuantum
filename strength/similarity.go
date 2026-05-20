package strength

import (
	"strconv"
	"strings"
	"unicode"
)

// Levenshtein computes the edit distance between two strings.
func Levenshtein(a, b string) int {
	ra := []rune(a)
	rb := []rune(b)
	if len(ra) == 0 {
		return len(rb)
	}
	if len(rb) == 0 {
		return len(ra)
	}

	prev := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}

	for i := 1; i <= len(ra); i++ {
		curr := make([]int, len(rb)+1)
		curr[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 0
			if ra[i-1] != rb[j-1] {
				cost = 1
			}
			curr[j] = min3(
				prev[j]+1,
				curr[j-1]+1,
				prev[j-1]+cost,
			)
		}
		prev = curr
	}

	return prev[len(rb)]
}

// JaccardSimilarity returns the n-gram Jaccard similarity between two strings.
func JaccardSimilarity(a, b string, n int) float64 {
	if n <= 0 {
		return 0
	}
	setA := buildNGramSet(a, n)
	setB := buildNGramSet(b, n)
	if len(setA) == 0 && len(setB) == 0 {
		return 1
	}
	intersection := 0
	union := make(map[string]struct{}, len(setA)+len(setB))
	for k := range setA {
		union[k] = struct{}{}
		if _, ok := setB[k]; ok {
			intersection++
		}
	}
	for k := range setB {
		union[k] = struct{}{}
	}
	if len(union) == 0 {
		return 0
	}
	return float64(intersection) / float64(len(union))
}

// CheckSimilarToStored compares a candidate password against stored passwords.
func CheckSimilarToStored(candidate string, stored []string) []Issue {
	issues := make([]Issue, 0, 4)
	seen := map[string]bool{
		"exact_reuse":           false,
		"near_duplicate":        false,
		"high_similarity":       false,
		"incremental_variation": false,
	}

	for _, existing := range stored {
		if candidate == "" || existing == "" {
			continue
		}

		if !seen["incremental_variation"] && stripCommonAffixes(candidate) == stripCommonAffixes(existing) && candidate != existing {
			issues = append(issues, Issue{Code: "incremental_variation", Message: "Looks like an incremental variation of an existing password (e.g. Password1 → Password2).", Penalty: 25})
			seen["incremental_variation"] = true
		}

		if !seen["exact_reuse"] && candidate == existing {
			issues = append(issues, Issue{Code: "exact_reuse", Message: "Matches one of your existing vault passwords exactly.", Penalty: 40})
			seen["exact_reuse"] = true
		}

		if !seen["near_duplicate"] && Levenshtein(candidate, existing) <= 2 {
			issues = append(issues, Issue{Code: "near_duplicate", Message: "Very similar to one of your existing vault passwords.", Penalty: 30})
			seen["near_duplicate"] = true
		}

		if !seen["high_similarity"] && JaccardSimilarity(strings.ToLower(candidate), strings.ToLower(existing), 2) >= 0.75 {
			issues = append(issues, Issue{Code: "high_similarity", Message: "Very similar to one of your existing vault passwords.", Penalty: 20})
			seen["high_similarity"] = true
		}
	}

	return issues
}

func buildNGramSet(value string, n int) map[string]struct{} {
	runes := []rune(value)
	set := make(map[string]struct{})
	if len(runes) == 0 {
		return set
	}
	if len(runes) < n {
		set[string(runes)] = struct{}{}
		return set
	}
	for i := 0; i+n <= len(runes); i++ {
		set[string(runes[i:i+n])] = struct{}{}
	}
	return set
}

func stripCommonAffixes(value string) string {
	s := strings.ToLower(strings.TrimSpace(value))
	for {
		before := s
		s = strings.TrimFunc(s, func(r rune) bool {
			return unicode.IsPunct(r) || unicode.IsSpace(r)
		})
		s = stripNumericAffix(s, true)
		s = stripNumericAffix(s, false)
		s = strings.TrimFunc(s, func(r rune) bool {
			return unicode.IsPunct(r) || unicode.IsSpace(r)
		})
		if s == before {
			break
		}
	}
	return s
}

func stripNumericAffix(value string, prefix bool) string {
	if value == "" {
		return value
	}
	runes := []rune(value)
	grab := func() []rune {
		if prefix {
			i := 0
			for i < len(runes) && unicode.IsDigit(runes[i]) {
				i++
			}
			return runes[:i]
		}
		i := len(runes)
		for i > 0 && unicode.IsDigit(runes[i-1]) {
			i--
		}
		return runes[i:]
	}
	digits := grab()
	if len(digits) == 0 {
		return value
	}
	n, err := strconvAtoiRunes(digits)
	if err != nil {
		return value
	}
	if !(n >= 0 && n <= 99 || n >= 1900 && n <= 2100) {
		return value
	}
	if prefix {
		return string(runes[len(digits):])
	}
	return string(runes[:len(runes)-len(digits)])
}

func strconvAtoiRunes(runes []rune) (int, error) {
	return strconv.Atoi(string(runes))
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
