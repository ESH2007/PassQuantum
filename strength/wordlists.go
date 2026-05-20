package strength

import (
	_ "embed"
	"strings"
)

//go:embed data/common_names.txt
var namesData string

//go:embed data/common_passwords.txt
var passwordsData string

// CommonNames returns the embedded list of common English and Spanish names.
func CommonNames() []string {
	return splitWordList(namesData)
}

// CommonPasswords returns the embedded list of common password base words.
func CommonPasswords() []string {
	return splitWordList(passwordsData)
}

func splitWordList(data string) []string {
	lines := strings.Split(data, "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	return result
}
