package screens

import (
	"fmt"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"passquantum/app"
	"passquantum/core/migration"
	"passquantum/theme"
	"passquantum/ui/widgets"
)

// importWizardState holds the per-render state of the import wizard.
// A new instance is built every time the user navigates to NavViewImport.
type importWizardState struct {
	ns *NavigationState

	root *fyne.Container

	step int // 0 = pick file, 1 = preview, 2 = done

	selectedImporterID string

	parseResult *migration.ImportResult
	parsedFrom  string // file path
	importer    migration.Importer

	dupAction migration.DuplicateAction

	summary *app.ImportSummary
	lastErr error
}

// createImportView is the entry point dispatched from updateContent.
func (ns *NavigationState) createImportView() fyne.CanvasObject {
	w := &importWizardState{
		ns:        ns,
		dupAction: migration.DupSkip,
	}
	w.root = container.NewMax()
	w.renderStep()
	return w.root
}

func (w *importWizardState) renderStep() {
	var content fyne.CanvasObject
	switch w.step {
	case 0:
		content = w.renderPickStep()
	case 1:
		content = w.renderPreviewStep()
	case 2:
		content = w.renderDoneStep()
	default:
		content = widget.NewLabel("invalid step")
	}
	w.root.Objects = []fyne.CanvasObject{content}
	w.root.Refresh()
}

// ---------------- Step 0: pick file ----------------

func (w *importWizardState) renderPickStep() fyne.CanvasObject {
	header := theme.PageHeader(
		"PASSQUANTUM / "+w.ns.appState.CurrentVault+" / IMPORT",
		"Import passwords",
		"Bring in credentials exported from another password manager or browser. Every imported entry is re-encrypted locally with the same post-quantum envelope as manual entries.",
		nil,
	)

	autoCard := w.formatCard(
		"Auto-detect format",
		"Pick a file and let PassQuantum figure out the format from its contents.",
		theme.IconWand,
		func() { w.pickFile("") },
	)

	// One card per registered importer.
	importers := migration.DefaultRegistry.All()
	cards := []fyne.CanvasObject{autoCard}
	for _, imp := range importers {
		impCopy := imp
		cards = append(cards, w.formatCard(
			imp.DisplayName(),
			"Accepted extensions: "+strings.Join(imp.Extensions(), ", "),
			theme.IconDownload,
			func() { w.pickFile(impCopy.ID()) },
		))
	}

	grid := container.NewGridWithColumns(2, cards...)

	warning := canvas.NewText(
		"⚠ Your source export file contains plaintext passwords. Delete it once the import completes.",
		theme.ColorWarning,
	)
	warning.TextSize = 12

	return container.NewVBox(
		header,
		container.NewPadded(theme.SectionEyebrow("CHOOSE A SOURCE")),
		grid,
		container.NewPadded(warning),
	)
}

func (w *importWizardState) formatCard(title, desc string, icon *fyne.StaticResource, onClick func()) fyne.CanvasObject {
	titleTxt := canvas.NewText(title, theme.ColorTextPrimary)
	titleTxt.TextSize = 13
	titleTxt.TextStyle = fyne.TextStyle{Bold: true}

	descTxt := canvas.NewText(desc, theme.ColorTextSecondary)
	descTxt.TextSize = 11

	iconBlock := theme.TypeIcon(icon, theme.ColorAccentCyan)

	body := container.NewBorder(nil, nil, iconBlock, nil,
		container.NewVBox(titleTxt, descTxt))

	btn := theme.CreateGhostButton("Choose file", onClick)

	return theme.CardWithHeader("", "", btn, body)
}

func (w *importWizardState) pickFile(importerID string) {
	w.selectedImporterID = importerID
	widgets.PickAnyFile("Select export file",
		func(path string) {
			w.parsedFrom = path
			w.loadAndParse(path)
		},
		func(err error) {
			widgets.ShowAppError(fmt.Errorf("file picker: %w", err), w.ns.window)
		},
	)
}

// loadAndParse runs the importer in a goroutine and transitions to the
// preview step on success.
func (w *importWizardState) loadAndParse(path string) {
	go func() {
		if err := migration.ValidateSize(path); err != nil {
			fyne.Do(func() {
				widgets.ShowAppError(fmt.Errorf("file too large (max 100 MB)"), w.ns.window)
			})
			return
		}

		importerID := w.selectedImporterID
		if importerID == "" {
			results, err := migration.DetectFile(path)
			if err != nil {
				fyne.Do(func() {
					widgets.ShowAppError(fmt.Errorf("detect format: %w", err), w.ns.window)
				})
				return
			}
			if len(results) == 0 {
				fyne.Do(func() {
					widgets.ShowAppError(fmt.Errorf("no importer recognized %q. Pick a format manually.", filepath.Base(path)), w.ns.window)
				})
				return
			}
			importerID = results[0].Importer.ID()
		}

		res, imp, err := app.ParseImportFile(path, importerID, migration.ParseOptions{})
		if err != nil {
			fyne.Do(func() {
				widgets.ShowAppError(fmt.Errorf("parse %s: %w", imp.DisplayName(), err), w.ns.window)
			})
			return
		}

		fyne.Do(func() {
			w.parseResult = res
			w.importer = imp
			w.step = 1
			w.renderStep()
		})
	}()
}

// ---------------- Step 1: preview + dedup policy ----------------

func (w *importWizardState) renderPreviewStep() fyne.CanvasObject {
	if w.parseResult == nil {
		return widget.NewLabel("no parse result")
	}

	subtitle := fmt.Sprintf("Detected as %s. Found %d entries (%d skipped).",
		w.importer.DisplayName(), len(w.parseResult.Entries), w.parseResult.Skipped)

	header := theme.PageHeader(
		"PASSQUANTUM / "+w.ns.appState.CurrentVault+" / IMPORT",
		"Review entries",
		subtitle,
		nil,
	)

	dupSelect := widget.NewSelect(
		[]string{"Skip duplicates", "Replace existing", "Keep both"},
		func(s string) {
			switch s {
			case "Skip duplicates":
				w.dupAction = migration.DupSkip
			case "Replace existing":
				w.dupAction = migration.DupReplace
			case "Keep both":
				w.dupAction = migration.DupKeepBoth
			}
		},
	)
	dupSelect.SetSelected("Skip duplicates")

	dupRow := container.NewBorder(nil, nil,
		theme.FieldLabel("WHEN A DUPLICATE IS FOUND", nil), nil, dupSelect)

	// Preview list — at most 50 rows, never showing passwords.
	previewRows := w.parseResult.Entries
	const previewLimit = 50
	if len(previewRows) > previewLimit {
		previewRows = previewRows[:previewLimit]
	}

	previewItems := make([]fyne.CanvasObject, 0, len(previewRows))
	for i := range previewRows {
		e := &previewRows[i]
		previewItems = append(previewItems, w.previewRow(e))
	}
	previewBox := container.NewVBox(previewItems...)

	previewScroll := container.NewVScroll(previewBox)
	previewScroll.SetMinSize(fyne.NewSize(0, 320))

	previewCard := theme.CardWithHeader(
		fmt.Sprintf("PREVIEW (first %d)", len(previewRows)),
		"",
		nil,
		previewScroll,
	)

	// Warnings — collapse into a single multi-line card so the layout stays compact.
	var warningCard fyne.CanvasObject
	if len(w.parseResult.Warnings) > 0 {
		text := strings.Join(w.parseResult.Warnings, "\n")
		warnTxt := canvas.NewText(text, theme.ColorWarning)
		warnTxt.TextSize = 11
		warningCard = theme.CardWithHeader("PARSER WARNINGS", "", nil, warnTxt)
	}

	backBtn := theme.CreateGhostButton("Back", func() {
		w.parseResult = nil
		w.step = 0
		w.renderStep()
	})

	importBtn := theme.CreatePrimaryButtonWithIcon(
		fmt.Sprintf("Import %d entries", len(w.parseResult.Entries)),
		theme.IconDownload,
		func() { w.runImport() },
	)

	actions := container.NewHBox(backBtn, importBtn)

	items := []fyne.CanvasObject{
		header,
		container.NewPadded(dupRow),
		previewCard,
	}
	if warningCard != nil {
		items = append(items, warningCard)
	}
	items = append(items, container.NewPadded(actions))

	return container.NewVBox(items...)
}

func (w *importWizardState) previewRow(e *migration.ImportedEntry) fyne.CanvasObject {
	title := migration.DeriveTitle(e.Title, e.URLs, e.Username)

	typeLabel := "Password"
	icon := theme.IconKey
	switch e.Type {
	case 2: // EntryTypeNote
		typeLabel = "Note"
		icon = theme.IconNote
	case 3: // EntryTypeCard
		typeLabel = "Card"
		icon = theme.IconCard
	case 4: // EntryTypeTOTP
		typeLabel = "TOTP"
		icon = theme.IconClock
	}

	titleTxt := canvas.NewText(title, theme.ColorTextPrimary)
	titleTxt.TextSize = 12
	titleTxt.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := e.Username
	if subtitle == "" && len(e.URLs) > 0 {
		subtitle = e.URLs[0]
	}
	if len(e.Password) > 0 {
		if subtitle != "" {
			subtitle += "  •  "
		}
		subtitle += "••••••"
	}
	subTxt := canvas.NewText(subtitle, theme.ColorTextSecondary)
	subTxt.TextSize = 11

	iconBlock := theme.TypeIcon(icon, theme.ColorAccentCyan)

	badge := theme.KindBadge(typeLabel)
	titleRow := container.NewHBox(titleTxt, badge)

	left := container.NewHBox(iconBlock, container.NewVBox(titleRow, subTxt))
	return theme.CardWithHeader("", "", nil, left)
}

// runImport invokes BatchImport in a goroutine, then transitions to the
// done step. Progress is communicated via an indeterminate progress bar.
func (w *importWizardState) runImport() {
	progress := widget.NewProgressBarInfinite()
	progressCard := theme.CardWithHeader("IMPORTING", "",
		nil,
		container.NewVBox(progress),
	)
	w.root.Objects = []fyne.CanvasObject{
		container.NewVBox(
			theme.PageHeader(
				"PASSQUANTUM / "+w.ns.appState.CurrentVault+" / IMPORT",
				"Encrypting entries",
				"Each entry is re-encrypted with a fresh Kyber-768 envelope.",
				nil,
			),
			progressCard,
		),
	}
	w.root.Refresh()

	parsed := w.parseResult
	dupAction := w.dupAction

	go func() {
		summary, err := app.BatchImport(w.ns.appState, parsed, dupAction)
		fyne.Do(func() {
			w.summary = summary
			w.lastErr = err
			w.step = 2
			w.renderStep()
		})
	}()
}

// ---------------- Step 2: done ----------------

func (w *importWizardState) renderDoneStep() fyne.CanvasObject {
	if w.lastErr != nil {
		header := theme.PageHeader(
			"PASSQUANTUM / "+w.ns.appState.CurrentVault+" / IMPORT",
			"Import failed",
			w.lastErr.Error(),
			nil,
		)
		retryBtn := theme.CreatePrimaryButton("Try another file", func() {
			w.parseResult = nil
			w.summary = nil
			w.lastErr = nil
			w.step = 0
			w.renderStep()
		})
		return container.NewVBox(header, container.NewPadded(retryBtn))
	}

	s := w.summary
	if s == nil {
		return widget.NewLabel("no summary")
	}

	header := theme.PageHeader(
		"PASSQUANTUM / "+w.ns.appState.CurrentVault+" / IMPORT",
		"Import complete",
		fmt.Sprintf("Imported %d new entries, replaced %d, skipped %d.",
			s.NewEntries, s.Replaced, s.DupSkipped),
		nil,
	)

	stats := container.NewVBox(
		statRow("Parsed rows", s.TotalParsed+s.Skipped),
		statRow("New entries written", s.NewEntries),
		statRow("Replaced in place", s.Replaced),
		statRow("Skipped (duplicate)", s.DupSkipped),
		statRow("Skipped (parser)", s.Skipped),
	)
	statsCard := theme.CardWithHeader("SUMMARY", "", nil, container.NewPadded(stats))

	cards := []fyne.CanvasObject{header, statsCard}

	if len(s.MapWarnings)+len(s.ParseWarnings) > 0 {
		all := append([]string{}, s.ParseWarnings...)
		all = append(all, s.MapWarnings...)
		text := strings.Join(all, "\n")
		warnTxt := canvas.NewText(text, theme.ColorWarning)
		warnTxt.TextSize = 11
		cards = append(cards, theme.CardWithHeader("WARNINGS", "", nil, container.NewPadded(warnTxt)))
	}

	if len(s.MapErrors) > 0 {
		text := strings.Join(s.MapErrors, "\n")
		errTxt := canvas.NewText(text, theme.ColorDanger)
		errTxt.TextSize = 11
		cards = append(cards, theme.CardWithHeader("ERRORS", "", nil, container.NewPadded(errTxt)))
	}

	warning := canvas.NewText(
		"⚠ Your source export file still contains plaintext passwords. Delete it now.",
		theme.ColorWarning,
	)
	warning.TextSize = 12
	cards = append(cards, container.NewPadded(warning))

	viewBtn := theme.CreatePrimaryButton("View items", func() {
		w.ns.switchView(NavViewItems)
	})
	doneBtn := theme.CreateGhostButton("Import another file", func() {
		w.parseResult = nil
		w.summary = nil
		w.step = 0
		w.renderStep()
	})
	cards = append(cards, container.NewPadded(container.NewHBox(viewBtn, doneBtn)))

	return container.NewVBox(cards...)
}

func statRow(label string, value int) fyne.CanvasObject {
	l := canvas.NewText(label, theme.ColorTextSecondary)
	l.TextSize = 12
	v := canvas.NewText(fmt.Sprintf("%d", value), theme.ColorTextPrimary)
	v.TextSize = 13
	v.TextStyle = fyne.TextStyle{Bold: true}
	return container.NewBorder(nil, nil, l, v)
}
