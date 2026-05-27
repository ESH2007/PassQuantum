package screens

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"passquantum/theme"
	"passquantum/ui/widgets"
)

type generatorControls struct {
	Settings          *PasswordGeneratorSettings
	LengthInput       *widget.Entry
	LengthSlider      *widget.Slider
	UppercaseCheck    *widget.Check
	LowercaseCheck    *widget.Check
	NumbersCheck      *widget.Check
	SpecialCharsCheck *widget.Check
	AmbiguousCheck    *widget.Check
	GenerateBtn       fyne.CanvasObject
	CopyBtn           fyne.CanvasObject
}

func newGeneratorControls(w fyne.Window, getText func() string, setText func(string)) *generatorControls {
	settings := DefaultPasswordGeneratorSettings()

	lengthInput := widget.NewEntry()
	lengthInput.SetText("16")

	lengthSlider := widget.NewSlider(4, 128)
	lengthSlider.Step = 1
	lengthSlider.SetValue(16)

	// Bidirectional sync: slider → entry
	lengthSlider.OnChanged = func(v float64) {
		l := int(v)
		settings.Length = l
		current := lengthInput.Text
		if current != strconv.Itoa(l) {
			lengthInput.SetText(strconv.Itoa(l))
		}
	}

	// Bidirectional sync: entry → slider
	lengthInput.OnChanged = func(s string) {
		if s != "" {
			fmt.Sscanf(s, "%d", &settings.Length)
			if settings.Length < 4 {
				settings.Length = 4
			}
			if settings.Length > 128 {
				settings.Length = 128
			}
			if lengthSlider.Value != float64(settings.Length) {
				lengthSlider.SetValue(float64(settings.Length))
			}
		}
	}

	uppercaseCheck := widget.NewCheck("Uppercase Letters (A-Z)", func(b bool) {
		settings.UseUppercase = b
	})
	uppercaseCheck.SetChecked(settings.UseUppercase)

	lowercaseCheck := widget.NewCheck("Lowercase Letters (a-z)", func(b bool) {
		settings.UseLowercase = b
	})
	lowercaseCheck.SetChecked(settings.UseLowercase)

	numbersCheck := widget.NewCheck("Numbers (0-9)", func(b bool) {
		settings.UseNumbers = b
	})
	numbersCheck.SetChecked(settings.UseNumbers)

	specialCharsCheck := widget.NewCheck("Special Characters (!@#$%^&*)", func(b bool) {
		settings.UseSpecialChars = b
	})
	specialCharsCheck.SetChecked(settings.UseSpecialChars)

	ambiguousCheck := widget.NewCheck("Exclude Ambiguous (i, l, 1, L, o, 0, O)", func(b bool) {
		settings.ExcludeAmbiguous = b
	})
	ambiguousCheck.SetChecked(settings.ExcludeAmbiguous)

	generateBtn := theme.CreatePrimaryButton("Generate", func() {
		password, err := GeneratePassword(settings)
		if err != nil {
			widgets.ShowAppError(err, w)
			return
		}
		setText(password)
	})

	copyBtn := theme.CreateDefaultButton("Copy", func() {
		if getText() != "" {
			w.Clipboard().SetContent(getText())
			widgets.ShowAppInformation("Copied", "Password copied to clipboard!", w)
		} else {
			widgets.ShowAppInformation("Empty", "Generate a password first!", w)
		}
	})

	return &generatorControls{
		Settings:          &settings,
		LengthInput:       lengthInput,
		LengthSlider:      lengthSlider,
		UppercaseCheck:    uppercaseCheck,
		LowercaseCheck:    lowercaseCheck,
		NumbersCheck:      numbersCheck,
		SpecialCharsCheck: specialCharsCheck,
		AmbiguousCheck:    ambiguousCheck,
		GenerateBtn:       generateBtn,
		CopyBtn:           copyBtn,
	}
}
