package strength

import (
	"math"
	"strings"
	"time"
	"unicode"
)

var (
	monthNames        = []string{"january", "february", "march", "april", "may", "june", "july", "august", "september", "october", "november", "december"}
	moonPhases        = []string{"🌑", "🌒", "🌓", "🌔", "🌕", "🌖", "🌗", "🌘"}
	twoLetterElements = []string{"He", "Li", "Be", "Ne", "Na", "Mg", "Al", "Si", "Cl", "Ar", "Ca", "Sc", "Ti", "Cr", "Mn", "Fe", "Co", "Ni", "Cu", "Zn", "Ga", "Ge", "As", "Se", "Br", "Kr", "Rb", "Sr", "Zr", "Nb", "Mo", "Tc", "Ru", "Rh", "Pd", "Ag", "Cd", "In", "Sn", "Sb", "Te", "Xe", "Cs", "Ba", "La", "Ce", "Pr", "Nd", "Pm", "Sm", "Eu", "Gd", "Tb", "Dy", "Ho", "Er", "Tm", "Yb", "Lu", "Hf", "Ta", "Re", "Os", "Ir", "Pt", "Au", "Hg", "Tl", "Pb", "Bi", "Po", "At", "Rn", "Fr", "Ra", "Ac", "Th", "Pa", "Np", "Pu", "Am", "Cm", "Bk", "Cf", "Es", "Fm", "Md", "No", "Lr", "Rf", "Db", "Sg", "Bh", "Hs", "Mt", "Ds", "Rg", "Cn", "Nh", "Fl", "Mc", "Lv", "Ts", "Og"}
)

// GenerateEasterEggRules evaluates the fixed PassQuantum Password Game rule list.
func GenerateEasterEggRules(password string) []EasterEggRule {
	lower := strings.ToLower(password)
	digitSum := sumDigits(password)
	currentPhase := currentMoonPhaseEmoji(time.Now().UTC())
	leapYearPresent := containsLeapYear(password)
	emojiCount := distinctEmojiCount(password)
	primeLength := isPrime(len([]rune(password)))

	rules := []EasterEggRule{
		{Index: 1, Description: "Your password must be at least 5 characters.", Satisfied: len([]rune(password)) >= 5},
		{Index: 2, Description: "Your password must include a number.", Satisfied: containsDigit(password)},
		{Index: 3, Description: "Your password must include an uppercase letter.", Satisfied: containsUpper(password)},
		{Index: 4, Description: "Your password must include a special character.", Satisfied: containsSpecial(password)},
		{Index: 5, Description: "The digits in your password must add up to 25.", Satisfied: digitSum == 25},
		{Index: 6, Description: "Your password must include a month of the year.", Satisfied: containsMonth(lower)},
		{Index: 7, Description: "Your password must include a Roman numeral.", Satisfied: containsRomanNumeral(password)},
		{Index: 8, Description: "Your password must include a two-letter element symbol from the periodic table.", Satisfied: containsElementSymbol(password)},
		{Index: 9, Description: "Your password must include a leap year.", Satisfied: leapYearPresent},
		{Index: 10, Description: "🥚 This is Paul. He hasn't hatched yet. Please include him in your password and keep him safe.", Satisfied: strings.Contains(password, "🥚")},
		{Index: 11, Description: "Your password must include the current phase of the moon as an emoji.", Satisfied: strings.Contains(password, currentPhase)},
		{Index: 12, Description: "Your password must contain at least 3 different emoji.", Satisfied: emojiCount >= 3},
		{Index: 13, Description: "Your password length must be a prime number.", Satisfied: primeLength},
		{Index: 14, Description: "Your password must include the word \"passquantum\" (case-insensitive).", Satisfied: strings.Contains(lower, "passquantum")},
		{Index: 15, Description: "🐣 Paul has hatched! He eats 🐛 every minute. Feed him 3 worms.", Satisfied: strings.Contains(password, "🐣") && strings.Count(password, "🐛") >= 3},
	}

	return rules
}

// EasterEggIntroMessage returns the intro text for the easter egg panel.
func EasterEggIntroMessage() string {
	return "🎮 You typed neal.fun — welcome to The PassQuantum Password Game!\nYour password must satisfy all rules simultaneously.\nGood luck. You'll need it."
}

func containsDigit(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func containsUpper(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) {
			return true
		}
	}
	return false
}

func containsSpecial(s string) bool {
	const specials = "!@#$%^&*()_+-=[]{}|;':\",.<>?/`~"
	for _, r := range s {
		if strings.ContainsRune(specials, r) {
			return true
		}
	}
	return false
}

func sumDigits(s string) int {
	total := 0
	for _, r := range s {
		if unicode.IsDigit(r) {
			total += int(r - '0')
		}
	}
	return total
}

func containsMonth(lower string) bool {
	for _, month := range monthNames {
		if strings.Contains(lower, month) {
			return true
		}
	}
	return false
}

func containsRomanNumeral(s string) bool {
	upper := strings.ToUpper(s)
	for _, token := range []string{"I", "V", "X", "L", "C", "D", "M"} {
		if strings.Contains(upper, token) {
			return true
		}
	}
	return false
}

func containsElementSymbol(s string) bool {
	for _, symbol := range twoLetterElements {
		if strings.Contains(s, symbol) {
			return true
		}
	}
	return false
}

func containsLeapYear(s string) bool {
	runes := []rune(s)
	for i := 0; i+4 <= len(runes); i++ {
		chunk := string(runes[i : i+4])
		year := 0
		ok := true
		for _, r := range chunk {
			if !unicode.IsDigit(r) {
				ok = false
				break
			}
			year = year*10 + int(r-'0')
		}
		if ok && isLeapYear(year) {
			return true
		}
	}
	return false
}

func isLeapYear(year int) bool {
	return (year%4 == 0 && year%100 != 0) || year%400 == 0
}

func currentMoonPhaseEmoji(now time.Time) string {
	knownNewMoon := time.Date(2000, time.January, 6, 18, 14, 0, 0, time.UTC)
	days := now.Sub(knownNewMoon).Hours() / 24
	phase := math.Mod(days, 29.53)
	if phase < 0 {
		phase += 29.53
	}
	index := int(math.Floor((phase/29.53)*8+0.5)) % len(moonPhases)
	return moonPhases[index]
}

func distinctEmojiCount(s string) int {
	seen := make(map[rune]struct{})
	for _, r := range s {
		if isEmojiRune(r) {
			seen[r] = struct{}{}
		}
	}
	return len(seen)
}

func isEmojiRune(r rune) bool {
	return (r >= 0x1F300 && r <= 0x1F9FF) || (r >= 0x2600 && r <= 0x27BF)
}

func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	if n == 2 {
		return true
	}
	if n%2 == 0 {
		return false
	}
	limit := int(math.Sqrt(float64(n)))
	for i := 3; i <= limit; i += 2 {
		if n%i == 0 {
			return false
		}
	}
	return true
}
