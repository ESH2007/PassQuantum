package screens

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"passquantum/strength"
	"passquantum/theme"
	"passquantum/ui/widgets"
)

func (ns *NavigationState) createGeneratorView() fyne.CanvasObject {
	generatedPasswordDisplay := widget.NewEntry()
	generatedPasswordDisplay.PlaceHolder = "Generated password will appear here"
	generatedPasswordDisplay.MultiLine = false

	gc := newGeneratorControls(ns.window,
		func() string { return generatedPasswordDisplay.Text },
		func(s string) { generatedPasswordDisplay.SetText(s) },
	)

	// Auto-generate on load
	if initial, err := GeneratePassword(*gc.Settings); err == nil {
		generatedPasswordDisplay.SetText(initial)
	}

	header := theme.PageHeader(
		"PASSQUANTUM / GENERATOR",
		"Password generator",
		"Cryptographically secure random password generation.",
		theme.StatusPill("CSPRNG: crypto/rand", theme.PillAccent),
	)

	// Output card
	regenerateBtn := theme.CreateGhostButton("Regenerate", func() {
		password, err := GeneratePassword(*gc.Settings)
		if err != nil {
			widgets.ShowAppError(err, ns.window)
			return
		}
		generatedPasswordDisplay.SetText(password)
	})

	passwordDisplay := canvas.NewText("", theme.ColorTextPrimary)
	passwordDisplay.TextSize = 22
	passwordDisplay.TextStyle = fyne.TextStyle{Monospace: true}

	generatedPasswordDisplay.OnChanged = func(s string) {
		passwordDisplay.Text = s
		passwordDisplay.Refresh()
	}
	// Trigger initial display update
	passwordDisplay.Text = generatedPasswordDisplay.Text

	displayBg := canvas.NewRectangle(theme.ColorSidebarBg)
	displayBg.CornerRadius = theme.Space2
	displayBorder := canvas.NewRectangle(color.Transparent)
	displayBorder.CornerRadius = theme.Space2
	displayBorder.StrokeWidth = 1
	displayBorder.StrokeColor = theme.ColorLine2
	displayBorder.FillColor = color.Transparent

	displayBox := container.NewStack(
		displayBg, displayBorder,
		container.NewPadded(container.NewPadded(passwordDisplay)),
	)

	copyBtn := theme.CreateDefaultButton("Copy", func() {
		if generatedPasswordDisplay.Text != "" {
			ns.window.Clipboard().SetContent(generatedPasswordDisplay.Text)
			widgets.ShowAppInformation("Copied", "Password copied to clipboard!", ns.window)
		}
	})

	saveToVaultBtn := theme.CreatePrimaryButton("Save to vault", func() {
		password := generatedPasswordDisplay.Text
		if password == "" {
			widgets.ShowAppInformation("Empty", "Generate a password first!", ns.window)
			return
		}
		showSaveGeneratedPasswordDialog(ns.window, ns.app, ns.appState, password)
	})

	outputBody := container.NewVBox(
		displayBox,
		container.NewHBox(copyBtn, saveToVaultBtn),
	)

	outputCard := theme.CardWithHeader("OUTPUT", "Generated password", regenerateBtn, outputBody)

	// Length row: slider fills available space, entry shows numeric value on right
	lengthRow := container.NewBorder(nil, nil, nil,
		container.NewGridWrap(fyne.NewSize(52, 0), gc.LengthInput),
		gc.LengthSlider,
	)

	// Options card
	optionsBody := container.NewVBox(
		theme.FieldLabel("LENGTH", nil),
		lengthRow,
		theme.FieldLabel("CHARACTER CLASSES", nil),
		container.NewGridWithColumns(2,
			gc.UppercaseCheck,
			gc.LowercaseCheck,
			gc.NumbersCheck,
			gc.SpecialCharsCheck,
		),
		gc.AmbiguousCheck,
	)

	optionsCard := theme.CardWithHeader("CONFIGURATION", "Parameters", nil, optionsBody)

	return container.NewVBox(header, outputCard, optionsCard)
}

func (ns *NavigationState) createCheckerView() fyne.CanvasObject {
	header := theme.PageHeader(
		"PASSQUANTUM / ANALYZER",
		"Password strength analyzer",
		"Offline analysis. Nothing leaves your machine.",
		theme.StatusPill("Offline: local only", theme.PillOk),
	)

	passwordInput := widget.NewPasswordEntry()
	passwordInput.PlaceHolder = "Enter password to check"

	strengthMeter := theme.SegmentedStrengthMeterFlex(0)
	strengthLabel := canvas.NewText("", theme.ColorTextSecondary)
	strengthLabel.TextSize = 11
	strengthLabel.TextStyle = fyne.TextStyle{Monospace: true}

	lengthVal := canvas.NewText("-", theme.ColorTextPrimary)
	lengthVal.TextSize = 13
	charsetVal := canvas.NewText("-", theme.ColorTextPrimary)
	charsetVal.TextSize = 13
	entropyVal := canvas.NewText("-", theme.ColorTextPrimary)
	entropyVal.TextSize = 13
	crackVal := canvas.NewText("-", theme.ColorTextPrimary)
	crackVal.TextSize = 13

	issuesBox := container.NewVBox()

	analysisContainer := container.NewVBox(
		container.NewBorder(nil, nil, nil, strengthLabel, strengthMeter),
	)

	divider := canvas.NewRectangle(theme.ColorLine1)
	divider.SetMinSize(fyne.NewSize(0, 1))

	kvTable := theme.KeyValueTable([]theme.KVItem{
		{Key: "Length", Value: "-"},
		{Key: "Character set", Value: "-"},
		{Key: "Estimated entropy", Value: "-"},
		{Key: "Brute-force estimate", Value: "-"},
	})

	hintText := canvas.NewText("Start typing to analyze password strength.", theme.ColorFg2)
	hintText.TextSize = 12
	hintText.TextStyle = fyne.TextStyle{Monospace: true}

	updateStrength := func(value string) {
		result := strength.Analyze(value, storedVaultPasswords(ns.appState))

		if result.EasterEggMode {
			level := int(result.Score) + 1
			if level > 5 {
				level = 5
			}
			newMeter := theme.SegmentedStrengthMeterFlex(level)
			strengthLabel.Text = "Password Game Mode"
			strengthLabel.Refresh()

			analysisContainer.Objects = []fyne.CanvasObject{
				container.NewBorder(nil, nil, nil, strengthLabel, newMeter),
				divider,
				NewEasterEggPanel(result.EasterEggRules),
			}
			analysisContainer.Refresh()
			return
		}

		level := 0
		if value != "" {
			level = int(result.Score) + 1
			if level > 5 {
				level = 5
			}
		}

		newMeter := theme.SegmentedStrengthMeterFlex(level)

		if value == "" {
			strengthLabel.Text = ""
			strengthLabel.Refresh()
			analysisContainer.Objects = []fyne.CanvasObject{
				container.NewBorder(nil, nil, nil, strengthLabel, newMeter),
				hintText,
			}
		} else {
			strengthLabel.Text = result.ScoreLabel
			strengthLabel.Refresh()
			// Build character set description
			var sets []string
			for _, r := range value {
				if r >= 'A' && r <= 'Z' {
					sets = append(sets, "A-Z")
					break
				}
			}
			for _, r := range value {
				if r >= 'a' && r <= 'z' {
					sets = append(sets, "a-z")
					break
				}
			}
			for _, r := range value {
				if r >= '0' && r <= '9' {
					sets = append(sets, "0-9")
					break
				}
			}
			hasSpecial := false
			for _, r := range value {
				if !((r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')) {
					hasSpecial = true
					break
				}
			}
			if hasSpecial {
				sets = append(sets, "symbols")
			}

			charsetStr := "-"
			if len(sets) > 0 {
				charsetStr = ""
				for i, s := range sets {
					if i > 0 {
						charsetStr += " : "
					}
					charsetStr += s
				}
			}

			lengthVal.Text = fmt.Sprintf("%d characters", len(value))
			lengthVal.Refresh()
			charsetVal.Text = charsetStr
			charsetVal.Refresh()
			entropyVal.Text = result.CrackTime
			entropyVal.Refresh()
			crackVal.Text = result.CrackTime
			crackVal.Refresh()

			kvTable = theme.KeyValueTable([]theme.KVItem{
				{Key: "Length", Value: fmt.Sprintf("%d characters", len(value))},
				{Key: "Character set", Value: charsetStr},
				{Key: "Estimated entropy", Value: fmt.Sprintf("%.0f bits", result.Entropy)},
				{Key: "Brute-force estimate", Value: result.CrackTime},
			})

			issuesBox.Objects = nil
			if len(result.Issues) > 0 {
				issuesBox.Objects = []fyne.CanvasObject{NewIssuesList(result.Issues)}
			}
			issuesBox.Refresh()

			analysisContainer.Objects = []fyne.CanvasObject{
				container.NewBorder(nil, nil, nil, strengthLabel, newMeter),
				divider,
				kvTable,
				issuesBox,
			}
		}
		analysisContainer.Refresh()
	}

	passwordInput.OnChanged = updateStrength
	updateStrength(passwordInput.Text)

	inputCard := theme.CardWithHeader("INPUT", "Password to analyze", nil,
		container.NewVBox(passwordInput),
	)

	analysisCard := theme.CardWithHeader("ANALYSIS", "Strength result", nil, analysisContainer)

	return container.NewVBox(header, inputCard, analysisCard)
}
