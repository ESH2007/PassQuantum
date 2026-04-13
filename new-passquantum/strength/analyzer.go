package strength

import "strings"

// Analyze performs full password analysis against patterns and stored-password similarity.
func Analyze(password string, storedPasswords []string) AnalysisResult {
	issues := make([]Issue, 0)
	issues = append(issues, DetectRepeatedChars(password)...)
	issues = append(issues, DetectKeyboardWalks(password)...)
	issues = append(issues, DetectLeetSpeak(password)...)
	issues = append(issues, DetectDatePatterns(password)...)
	issues = append(issues, DetectCommonNames(password, CommonNames())...)
	issues = append(issues, DetectCommonWords(password, CommonPasswords())...)
	issues = append(issues, DetectRepeatedPatterns(password)...)
	issues = append(issues, DetectMissingCharClasses(password)...)
	issues = append(issues, DetectShortLength(password)...)
	issues = append(issues, CheckSimilarToStored(password, storedPasswords)...)

	entropy := CalcEntropy(password)
	crackTime := EstimateCrackTime(entropy)

	penalty := 0
	for _, issue := range issues {
		penalty += issue.Penalty
	}
	rawScore := 100 - penalty
	if rawScore < 0 {
		rawScore = 0
	}
	if rawScore > 100 {
		rawScore = 100
	}

	score := ScoreVeryWeak
	switch {
	case rawScore >= 90:
		score = ScoreVeryStrong
	case rawScore >= 70:
		score = ScoreStrong
	case rawScore >= 50:
		score = ScoreFair
	case rawScore >= 30:
		score = ScoreWeak
	}

	result := AnalysisResult{
		Score:       score,
		ScoreLabel:  scoreLabel(score),
		Entropy:     entropy,
		CrackTime:   crackTime,
		Issues:      issues,
		Suggestions: buildSuggestions(issues),
	}

	if strings.Contains(strings.ToLower(password), "neal.fun") {
		result.EasterEggMode = true
		result.EasterEggRules = GenerateEasterEggRules(password)
	}

	return result
}

func buildSuggestions(issues []Issue) []string {
	seen := make(map[string]struct{})
	suggestions := make([]string, 0, len(issues))
	for _, issue := range issues {
		if _, ok := seen[issue.Code]; ok {
			continue
		}
		seen[issue.Code] = struct{}{}
		suggestions = append(suggestions, suggestionForIssue(issue.Code))
	}
	return suggestions
}

func suggestionForIssue(code string) string {
	switch code {
	case "repeated_chars", "repeated_pattern":
		return "Try a less repetitive structure so each section of the password feels more unique."
	case "keyboard_walk":
		return "Swap out keyboard sequences for unrelated characters or words."
	case "leet_speak":
		return "Use an original phrase instead of a common word with substitutions."
	case "date_pattern":
		return "Avoid birthdays, years, and date-like patterns that are easy to guess."
	case "common_name":
		return "Replace familiar names with less predictable words or random fragments."
	case "common_word":
		return "Choose words that are less common or mix unrelated terms together."
	case "missing_lowercase":
		return "Add at least one lowercase letter for more variety."
	case "missing_uppercase":
		return "Add at least one uppercase letter for more variety."
	case "missing_digit":
		return "Add at least one digit to widen the search space."
	case "missing_special":
		return "Add at least one special character to increase complexity."
	case "short_length":
		return "Make the password longer so it resists brute-force attacks better."
	case "exact_reuse", "near_duplicate", "high_similarity", "incremental_variation":
		return "Pick something clearly different from your existing vault passwords."
	default:
		return "Add more unique length and character variety to strengthen the password."
	}
}
