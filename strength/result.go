package strength

// Score represents the final password strength bucket.
type Score int

const (
	ScoreVeryWeak   Score = 0
	ScoreWeak       Score = 1
	ScoreFair       Score = 2
	ScoreStrong     Score = 3
	ScoreVeryStrong Score = 4
)

// Issue describes a weakness discovered during password analysis.
type Issue struct {
	Code    string
	Message string
	Penalty int
}

// EasterEggRule represents a single rule in the PassQuantum Password Game.
type EasterEggRule struct {
	Index       int
	Description string
	Satisfied   bool
}

// AnalysisResult contains the complete result of a password analysis run.
type AnalysisResult struct {
	Score          Score
	ScoreLabel     string
	Entropy        float64
	CrackTime      string
	Issues         []Issue
	Suggestions    []string
	EasterEggMode  bool
	EasterEggRules []EasterEggRule
}

func scoreLabel(score Score) string {
	switch score {
	case ScoreVeryStrong:
		return "Very Strong"
	case ScoreStrong:
		return "Strong"
	case ScoreFair:
		return "Fair"
	case ScoreWeak:
		return "Weak"
	default:
		return "Very Weak"
	}
}
