package screens

import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"passquantum/core/model"
	"passquantum/strength"
	"passquantum/app"
	"passquantum/theme"
)

const (
	strengthBarWidth  float32 = 200
	strengthBarHeight float32 = 14
)

type StrengthBarWidget struct {
	*fyne.Container
	track           *canvas.Rectangle
	fill            *canvas.Rectangle
	titleLabel      *canvas.Text
	scoreLabel      *canvas.Text
	crackLabel      *canvas.Text
	issuesContainer *fyne.Container
	normalContainer *fyne.Container
	easterContainer *fyne.Container
	lastSatisfied   int
}

// NewStrengthBar creates a reusable live password strength widget.
func NewStrengthBar() fyne.CanvasObject {
	track := canvas.NewRectangle(color.NRGBA{R: 31, G: 41, B: 55, A: 255})
	track.Resize(fyne.NewSize(strengthBarWidth, strengthBarHeight))
	track.CornerRadius = 6

	fill := canvas.NewRectangle(scoreColor(strength.ScoreVeryWeak))
	fill.Resize(fyne.NewSize(strengthBarWidth/5, strengthBarHeight))
	fill.CornerRadius = 6

	bar := container.NewGridWrap(fyne.NewSize(strengthBarWidth, strengthBarHeight), container.NewWithoutLayout(track, fill))

	titleLabel := canvas.NewText("STRENGTH ANALYSIS", theme.ColorAccentCyan)
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	titleLabel.TextSize = 11

	scoreLabel := canvas.NewText("Very Weak  🔓", theme.ColorTextPrimary)
	scoreLabel.TextStyle = fyne.TextStyle{Bold: true}
	scoreLabel.TextSize = 12

	crackLabel := canvas.NewText("Crack time: Instantly", theme.ColorTextSecondary)
	crackLabel.TextSize = 11

	issuesContainer := container.NewVBox(newStrengthText("Start typing to analyze password strength.", theme.ColorTextSecondary, 11, false))

	normalContainer := container.NewVBox(
		titleLabel,
		container.NewHBox(bar, widget.NewLabel("  "), scoreLabel),
		crackLabel,
		issuesContainer,
	)
	easterContainer := container.NewStack()
	easterContainer.Hide()

	root := container.NewVBox(normalContainer, easterContainer)
	return &StrengthBarWidget{
		Container:       root,
		track:           track,
		fill:            fill,
		titleLabel:      titleLabel,
		scoreLabel:      scoreLabel,
		crackLabel:      crackLabel,
		issuesContainer: issuesContainer,
		normalContainer: normalContainer,
		easterContainer: easterContainer,
	}
}

// BindStrengthBar binds a password entry to live strength analysis updates.
func BindStrengthBar(bar fyne.CanvasObject, passwordEntry *widget.Entry, getStored func() []string) {
	strengthBar, ok := bar.(*StrengthBarWidget)
	if !ok || passwordEntry == nil {
		return
	}

	previous := passwordEntry.OnChanged
	passwordEntry.OnChanged = func(value string) {
		if previous != nil {
			previous(value)
		}

		stored := []string(nil)
		if getStored != nil {
			stored = getStored()
		}

		result := strength.Analyze(value, stored)
		strengthBar.apply(result)
	}

	initialStored := []string(nil)
	if getStored != nil {
		initialStored = getStored()
	}
	strengthBar.apply(strength.Analyze(passwordEntry.Text, initialStored))
}

// NewIssuesList creates a stacked list of issue labels with severity indicators.
func NewIssuesList(issues []strength.Issue) fyne.CanvasObject {
	if len(issues) == 0 {
		return newStrengthText("No major issues detected.", theme.ColorTextSecondary, 11, false)
	}

	items := make([]fyne.CanvasObject, 0, len(issues))
	for _, issue := range issues {
		indicator := "💡"
		if issue.Penalty >= 20 {
			indicator = "❌"
		} else if issue.Penalty >= 10 {
			indicator = "⚠️"
		}
		label := newStrengthText(indicator+" "+issue.Message, theme.ColorTextSecondary, 11, false)
		items = append(items, label)
	}
	return container.NewVBox(items...)
}

// NewEasterEggPanel creates the PassQuantum Password Game rule list panel.
func NewEasterEggPanel(rules []strength.EasterEggRule) fyne.CanvasObject {
	intro := newStrengthText(strength.EasterEggIntroMessage(), theme.ColorTextPrimary, 11, true)

	items := []fyne.CanvasObject{intro, widget.NewLabel("")}
	for _, rule := range rules {
		rowColor := color.NRGBA{R: 226, G: 75, B: 74, A: 255}
		prefix := "❌ "
		if rule.Satisfied {
			rowColor = color.NRGBA{R: 29, G: 158, B: 117, A: 255}
			prefix = "✅ "
		}
		row := canvas.NewText(prefix+rule.Description, rowColor)
		row.TextSize = 11
		items = append(items, row)
	}

	scroll := container.NewVScroll(container.NewVBox(items...))
	scroll.SetMinSize(fyne.NewSize(0, 170))
	return scroll
}

func (sb *StrengthBarWidget) apply(result strength.AnalysisResult) {
	if result.EasterEggMode {
		sb.normalContainer.Hide()
		satisfied := countSatisfied(result.EasterEggRules)
		bg := canvas.NewRectangle(color.NRGBA{R: 83, G: 74, B: 183, A: 30})
		bg.CornerRadius = theme.BorderRadius
		panel := NewEasterEggPanel(result.EasterEggRules)
		sb.easterContainer.Objects = []fyne.CanvasObject{container.NewStack(bg, container.NewPadded(panel))}
		sb.easterContainer.Show()
		sb.easterContainer.Refresh()
		if satisfied > sb.lastSatisfied {
			pulseBackground(bg)
		}
		sb.lastSatisfied = satisfied
		return
	}

	sb.lastSatisfied = 0
	sb.easterContainer.Hide()
	sb.normalContainer.Show()

	fillWidth := strengthBarWidth * float32(int(result.Score)+1) / 5
	if fillWidth < 1 {
		fillWidth = 1
	}
	sb.fill.FillColor = scoreColor(result.Score)
	sb.fill.Resize(fyne.NewSize(fillWidth, strengthBarHeight))
	sb.fill.Refresh()

	icon := "🔓"
	if result.Score >= strength.ScoreStrong {
		icon = "🔒"
	}
	sb.titleLabel.Text = "STRENGTH ANALYSIS"
	sb.titleLabel.Refresh()
	sb.scoreLabel.Text = fmt.Sprintf("%s  %s", result.ScoreLabel, icon)
	sb.scoreLabel.Refresh()
	sb.crackLabel.Text = "Crack time: " + result.CrackTime
	sb.crackLabel.Refresh()
	sb.issuesContainer.Objects = []fyne.CanvasObject{NewIssuesList(result.Issues)}
	sb.issuesContainer.Refresh()
}

func newStrengthText(value string, clr color.Color, size float32, bold bool) *canvas.Text {
	t := canvas.NewText(value, clr)
	t.TextSize = size
	t.TextStyle = fyne.TextStyle{Bold: bold}
	return t
}

func scoreColor(score strength.Score) color.NRGBA {
	switch score {
	case strength.ScoreVeryStrong:
		return color.NRGBA{R: 0x53, G: 0x4A, B: 0xB7, A: 255}
	case strength.ScoreStrong:
		return color.NRGBA{R: 0x1D, G: 0x9E, B: 0x75, A: 255}
	case strength.ScoreFair:
		return color.NRGBA{R: 0x97, G: 0xC4, B: 0x59, A: 255}
	case strength.ScoreWeak:
		return color.NRGBA{R: 0xEF, G: 0x9F, B: 0x27, A: 255}
	default:
		return color.NRGBA{R: 0xE2, G: 0x4B, B: 0x4A, A: 255}
	}
}

func pulseBackground(bg *canvas.Rectangle) {
	base := bg.FillColor.(color.NRGBA)
	go func() {
		alphas := []uint8{22, 40, 65, 40, 22}
		for _, alpha := range alphas {
			current := alpha
			fyne.Do(func() {
				bg.FillColor = color.NRGBA{R: base.R, G: base.G, B: base.B, A: current}
				bg.Refresh()
			})
			time.Sleep(80 * time.Millisecond)
		}
	}()
}

func countSatisfied(rules []strength.EasterEggRule) int {
	total := 0
	for _, rule := range rules {
		if rule.Satisfied {
			total++
		}
	}
	return total
}

func storedVaultPasswords(appState *app.AppState) []string {
	if appState == nil || appState.CurrentVault == "" || appState.PrivateKey == nil {
		return nil
	}

	appState.Mu.Lock()
	defer appState.Mu.Unlock()

	entries, err := app.ReadVault(app.GetVaultPath(appState.CurrentVault), appState.MasterPassword)
	if err != nil {
		return nil
	}

	stored := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.Type != model.EntryTypePassword {
			continue
		}
		sharedSecret, err := app.Decapsulate(entry.KyberCiphertext, appState.PrivateKey)
		if err != nil {
			continue
		}
		plaintext, err := app.DecryptAES256GCM(entry.Nonce, entry.Ciphertext, sharedSecret)
		if err != nil {
			continue
		}
		stored = append(stored, plaintext)
	}
	return stored
}
