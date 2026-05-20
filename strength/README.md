# strength/

Password strength analysis engine. No UI dependency — all functions operate on
strings and return typed result structs consumed by the UI layer.

| File | Description |
|---|---|
| `result.go` | Defines `AnalysisResult` and `PasswordIssue` types returned by the analyzer. |
| `analyzer.go` | `Analyze(password, stored)`: orchestrates all checks and returns a consolidated `AnalysisResult` with a numeric score, issues list, and crack-time estimate. |
| `entropy.go` | Entropy calculation and crack-time estimation based on character-set size and password length. |
| `patterns.go` | Detects structural weaknesses: repeated characters, keyboard walks, date patterns, common words, missing character classes, and short length. |
| `similarity.go` | Levenshtein and Jaccard similarity checks against the set of already-stored passwords to detect password reuse. |
| `wordlists.go` | Embeds common password and name wordlists used by `patterns.go` for dictionary matching. |
| `easter_egg.go` | "neal.fun password game" easter egg: generates fun strength-challenge messages for specific password patterns. Called from the UI strength widget. |
