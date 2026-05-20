package strength

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

var datePatternRegexps = []*regexp.Regexp{
	regexp.MustCompile(`\b\d{2}[/-]\d{2}[/-]\d{2,4}\b`),
	regexp.MustCompile(`\b\d{8}\b`),
	regexp.MustCompile(`\b\d{4}\b`),
}

// DetectRepeatedChars flags repeated characters or substrings appearing at least three times.
func DetectRepeatedChars(p string) []Issue {
	runes := []rune(p)
	if len(runes) == 0 {
		return nil
	}

	charCounts := make(map[rune]int)
	for _, r := range runes {
		charCounts[r]++
		if charCounts[r] >= 3 {
			return []Issue{{Code: "repeated_chars", Message: "Contains heavily repeated characters or fragments.", Penalty: 15}}
		}
	}

	maxLen := len(runes) / 3
	if maxLen > 4 {
		maxLen = 4
	}
	for size := 2; size <= maxLen; size++ {
		counts := make(map[string]int)
		for i := 0; i+size <= len(runes); i++ {
			key := string(runes[i : i+size])
			counts[key]++
			if counts[key] >= 3 {
				return []Issue{{Code: "repeated_chars", Message: "Contains heavily repeated characters or fragments.", Penalty: 15}}
			}
		}
	}

	return nil
}

// DetectKeyboardWalks flags common keyboard walks of length four or more.
func DetectKeyboardWalks(p string) []Issue {
	lower := strings.ToLower(p)
	sequences := []string{
		"qwertyuiop", "asdfghjkl", "zxcvbnm", "1234567890", "0987654321",
		"poiuytrewq", "lkjhgfdsa", "mnbvcxz",
	}
	for _, seq := range sequences {
		for size := 4; size <= len(seq); size++ {
			for i := 0; i+size <= len(seq); i++ {
				if strings.Contains(lower, seq[i:i+size]) {
					return []Issue{{Code: "keyboard_walk", Message: "Contains an easy keyboard sequence.", Penalty: 20}}
				}
			}
		}
	}
	return nil
}

// DetectLeetSpeak flags dictionary-like words that become obvious after common leet substitutions.
func DetectLeetSpeak(p string) []Issue {
	replacer := strings.NewReplacer(
		"@", "a", "3", "e", "0", "o", "1", "i", "$", "s", "4", "a",
	)
	normalized := strings.ToLower(replacer.Replace(p))
	for _, word := range CommonPasswords() {
		if len(word) < 4 {
			continue
		}
		if strings.Contains(normalized, strings.ToLower(word)) {
			return []Issue{{Code: "leet_speak", Message: "Looks like a common word disguised with leet substitutions.", Penalty: 18}}
		}
	}
	return nil
}

// DetectDatePatterns flags common date-like fragments and standalone years.
func DetectDatePatterns(p string) []Issue {
	for _, rx := range datePatternRegexps {
		matches := rx.FindAllString(p, -1)
		for _, match := range matches {
			if looksLikeDate(match) {
				return []Issue{{Code: "date_pattern", Message: "Contains a recognizable date or year pattern.", Penalty: 12}}
			}
		}
	}
	return nil
}

// DetectCommonNames flags common first-name substrings.
func DetectCommonNames(p string, names []string) []Issue {
	lower := strings.ToLower(p)
	matches := 0
	seen := make(map[string]struct{})
	for _, name := range names {
		name = strings.TrimSpace(strings.ToLower(name))
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		if strings.Contains(lower, name) {
			seen[name] = struct{}{}
			matches++
		}
	}
	if matches == 0 {
		return nil
	}
	penalty := matches * 15
	if penalty > 25 {
		penalty = 25
	}
	return []Issue{{Code: "common_name", Message: "Contains one or more common names.", Penalty: penalty}}
}

// DetectCommonWords flags common dictionary or password-base words.
func DetectCommonWords(p string, words []string) []Issue {
	lower := strings.ToLower(p)
	matches := 0
	seen := make(map[string]struct{})
	for _, word := range words {
		word = strings.TrimSpace(strings.ToLower(word))
		if len(word) < 4 {
			continue
		}
		if _, ok := seen[word]; ok {
			continue
		}
		if strings.Contains(lower, word) {
			seen[word] = struct{}{}
			matches++
		}
	}
	if matches == 0 {
		return nil
	}
	penalty := matches * 20
	if penalty > 30 {
		penalty = 30
	}
	return []Issue{{Code: "common_word", Message: "Contains common dictionary or password words.", Penalty: penalty}}
}

// DetectRepeatedPatterns flags structural repetition such as abcabc or abababab.
func DetectRepeatedPatterns(p string) []Issue {
	runes := []rune(p)
	if len(runes) < 4 {
		return nil
	}

	for size := 1; size <= len(runes)/2; size++ {
		for start := 0; start+2*size <= len(runes); start++ {
			first := string(runes[start : start+size])
			count := 1
			idx := start + size
			for idx+size <= len(runes) && string(runes[idx:idx+size]) == first {
				count++
				idx += size
			}
			if count >= 2 && size >= 2 {
				return []Issue{{Code: "repeated_pattern", Message: "Contains a repeated structural pattern.", Penalty: 15}}
			}
		}
	}

	for i := 0; i+5 < len(runes); i += 2 {
		if runes[i] == runes[i+1] && runes[i+2] == runes[i+3] && runes[i+4] == runes[i+5] {
			return []Issue{{Code: "repeated_pattern", Message: "Contains a repeated structural pattern.", Penalty: 15}}
		}
	}

	return nil
}

// DetectMissingCharClasses flags each missing character class independently.
func DetectMissingCharClasses(p string) []Issue {
	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSpecial := false

	for _, r := range p {
		switch {
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	issues := make([]Issue, 0, 4)
	if !hasLower {
		issues = append(issues, Issue{Code: "missing_lowercase", Message: "Missing a lowercase letter.", Penalty: 10})
	}
	if !hasUpper {
		issues = append(issues, Issue{Code: "missing_uppercase", Message: "Missing an uppercase letter.", Penalty: 10})
	}
	if !hasDigit {
		issues = append(issues, Issue{Code: "missing_digit", Message: "Missing a digit.", Penalty: 10})
	}
	if !hasSpecial {
		issues = append(issues, Issue{Code: "missing_special", Message: "Missing a special character.", Penalty: 10})
	}
	return issues
}

// DetectShortLength flags passwords shorter than current recommended lengths.
func DetectShortLength(p string) []Issue {
	length := len([]rune(p))
	if length < 8 {
		return []Issue{{Code: "short_length", Message: "Length is below the minimum recommended threshold.", Penalty: 30}}
	}
	if length < 12 {
		return []Issue{{Code: "short_length", Message: "Length is acceptable but still on the short side.", Penalty: 10}}
	}
	return nil
}

func looksLikeDate(match string) bool {
	compact := strings.NewReplacer("/", "", "-", "", ".", "").Replace(match)
	if len(compact) == 4 {
		year, err := strconv.Atoi(compact)
		return err == nil && year >= 1900 && year <= 2100
	}
	if len(compact) == 8 {
		return true
	}
	if strings.ContainsAny(match, "/-") {
		return true
	}
	return false
}

func sortedIssueCodes(issues []Issue) []string {
	codes := make([]string, 0, len(issues))
	for _, issue := range issues {
		codes = append(codes, issue.Code)
	}
	sort.Strings(codes)
	return codes
}

func issueMessage(code string, format string, values ...any) Issue {
	return Issue{Code: code, Message: fmt.Sprintf(format, values...)}
}
