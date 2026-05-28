package screens

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"passquantum/app"
	"passquantum/core/model"
	"passquantum/core/totp"
	"passquantum/theme"
	"passquantum/ui/widgets"
)

type liveTOTPCard struct {
	issuer  string
	account string
	card    fyne.CanvasObject
	codeTxt *canvas.Text
	barRect *canvas.Rectangle
	params  *totp.TOTPParams
}

func (ns *NavigationState) createTOTPView() fyne.CanvasObject {
	addBtn := theme.CreatePrimaryButtonWithIcon("Add account", theme.IconPlus, func() {
		ns.showAddTOTPDialog()
	})

	header := theme.PageHeader(
		"PASSQUANTUM / "+ns.appState.CurrentVault+" / AUTHENTICATOR",
		"2FA Authenticator",
		"Time-based one-time passwords. All codes generated offline.",
		addBtn,
	)

	itemsContainer := container.NewVBox()
	loadingText := canvas.NewText("Loading TOTP entries...", theme.ColorFg2)
	loadingText.TextSize = 13
	itemsContainer.Objects = []fyne.CanvasObject{container.NewCenter(loadingText)}

	var allCards []liveTOTPCard

	searchEntry := widget.NewEntry()
	searchEntry.PlaceHolder = "Search TOTP accounts..."
	searchEntry.OnChanged = func(query string) {
		q := strings.ToLower(query)
		var filtered []fyne.CanvasObject
		for _, tc := range allCards {
			if q == "" || strings.Contains(strings.ToLower(tc.issuer), q) || strings.Contains(strings.ToLower(tc.account), q) {
				filtered = append(filtered, tc.card)
			}
		}
		if len(filtered) == 0 && query != "" {
			noMatch := canvas.NewText("No accounts match “"+query+"”", theme.ColorFg2)
			noMatch.TextSize = 13
			itemsContainer.Objects = []fyne.CanvasObject{container.NewCenter(noMatch)}
		} else {
			itemsContainer.Objects = filtered
		}
		itemsContainer.Refresh()
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		ns.appState.Mu.Lock()
		defer ns.appState.Mu.Unlock()

		vaultFile := app.GetVaultPath(ns.appState.CurrentVault)
		entries, err := app.ReadVault(vaultFile, ns.appState.MasterPassword)
		if err != nil {
			fyne.Do(func() {
				errText := canvas.NewText("Failed to read vault: "+err.Error(), theme.ColorDanger)
				errText.TextSize = 13
				itemsContainer.Objects = []fyne.CanvasObject{errText}
				itemsContainer.Refresh()
			})
			return
		}

		var totpEntries []*model.VaultEntry
		var payloads []string
		for _, entry := range entries {
			if entry.Type != model.EntryTypeTOTP && !strings.HasPrefix(entry.Service, "TOTP:") {
				continue
			}
			ss, err := app.Decapsulate(entry.KyberCiphertext, ns.appState.PrivateKey)
			if err != nil {
				continue
			}
			plaintext, err := app.DecryptAES256GCM(entry.Nonce, entry.Ciphertext, ss)
			if err != nil {
				continue
			}
			totpEntries = append(totpEntries, entry)
			payloads = append(payloads, plaintext)
		}

		fyne.Do(func() {
			if len(totpEntries) == 0 {
				vaultIco := canvas.NewImageFromResource(theme.IconClock)
				vaultIco.SetMinSize(fyne.NewSize(40, 40))
				emptyTitle := canvas.NewText("No TOTP accounts yet", theme.ColorTextPrimary)
				emptyTitle.TextSize = 15
				emptyTitle.TextStyle = fyne.TextStyle{Bold: true}
				emptySubtitle := canvas.NewText("Add your first 2FA account to this vault.", theme.ColorFg2)
				emptySubtitle.TextSize = 12
				addFirstBtn := theme.CreatePrimaryButton("Add account", func() {
					ns.showAddTOTPDialog()
				})
				emptyState := container.NewCenter(container.NewVBox(
					container.NewCenter(vaultIco),
					container.NewCenter(emptyTitle),
					container.NewCenter(emptySubtitle),
					container.NewCenter(addFirstBtn),
				))
				itemsContainer.Objects = []fyne.CanvasObject{emptyState}
				itemsContainer.Refresh()
				return
			}

			var cards []fyne.CanvasObject
			for i, entry := range totpEntries {
				params, err := totp.Deserialize([]byte(payloads[i]))
				if err != nil {
					continue
				}

				codeTxt := canvas.NewText("--- ---", theme.ColorTextPrimary)
				codeTxt.TextSize = 28
				codeTxt.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}

				barRect := canvas.NewRectangle(theme.ColorAccentCyan)
				barRect.SetMinSize(fyne.NewSize(0, 4))
				barRect.CornerRadius = 2

				barBg := canvas.NewRectangle(theme.ColorBg3)
				barBg.SetMinSize(fyne.NewSize(0, 4))
				barBg.CornerRadius = 2

				card := buildTOTPCard(entry, params, codeTxt, barRect, barBg, ns.window, ns.app, ns.appState)

				allCards = append(allCards, liveTOTPCard{
					issuer:  params.Issuer,
					account: params.Account,
					card:    card,
					codeTxt: codeTxt,
					barRect: barRect,
					params:  params,
				})
				cards = append(cards, card)
			}

			itemsContainer.Objects = cards
			itemsContainer.Refresh()

			go runTOTPTicker(ctx, allCards)
		})
	}()

	ns.onViewChange(cancel)

	searchBg := canvas.NewRectangle(theme.ColorSidebarBg)
	searchBg.CornerRadius = theme.RadiusInput
	searchBorder := canvas.NewRectangle(color.Transparent)
	searchBorder.CornerRadius = theme.RadiusInput
	searchBorder.StrokeWidth = 1
	searchBorder.StrokeColor = theme.ColorLine2
	searchBorder.FillColor = color.Transparent
	searchRow := container.NewStack(searchBg, searchBorder, container.NewPadded(searchEntry))

	return container.NewVBox(header, searchRow, itemsContainer)
}

func runTOTPTicker(ctx context.Context, cards []liveTOTPCard) {
	updateAll := func() {
		for i := range cards {
			tc := &cards[i]
			code, remaining, err := totp.GenerateCode(tc.params)
			if err != nil {
				continue
			}
			formatted := totp.FormatCode(code)
			fraction := float32(remaining) / float32(tc.params.Period)
			codeTxt := tc.codeTxt
			barRect := tc.barRect

			fyne.Do(func() {
				codeTxt.Text = formatted
				codeTxt.Refresh()
				barRect.SetMinSize(fyne.NewSize(200*fraction, 4))
				barRect.Refresh()
			})
		}
	}

	updateAll()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			updateAll()
		}
	}
}

func buildTOTPCard(entry *model.VaultEntry, params *totp.TOTPParams, codeTxt *canvas.Text, barRect, barBg *canvas.Rectangle, w fyne.Window, fyneApp fyne.App, appState *app.AppState) fyne.CanvasObject {
	icon := theme.TypeIcon(theme.IconClock, theme.ColorAccentCyan)

	issuer := params.Issuer
	if issuer == "" {
		issuer = strings.TrimPrefix(entry.Service, "TOTP:")
	}

	titleTxt := canvas.NewText(issuer, theme.ColorTextPrimary)
	titleTxt.TextSize = 13
	titleTxt.TextStyle = fyne.TextStyle{Bold: true}

	badge := theme.KindBadge("TOTP")
	titleRow := container.NewHBox(titleTxt, badge)

	accountTxt := canvas.NewText(params.Account, theme.ColorTextSecondary)
	accountTxt.TextSize = 11
	accountTxt.TextStyle = fyne.TextStyle{Monospace: true}

	progressBar := container.NewMax(barBg, container.NewHBox(barRect, layout.NewSpacer()))

	copyBtn := theme.CreateSmallIconButton(theme.IconCopy, func() {
		code, _, err := totp.GenerateCode(params)
		if err != nil {
			widgets.ShowAppError(err, w)
			return
		}
		clipboardAutoClear(w, code)
		widgets.ShowAppInformation("Copied", "TOTP code copied (auto-clears in 30s)", w)
	})

	deleteBtn := theme.CreateSmallIconButton(theme.IconTrash, func() {
		widgets.ShowAppConfirm("Delete", fmt.Sprintf("Delete TOTP for %s?", issuer), func(ok bool) {
			if ok {
				deleteEntryByID(entry.ID, "TOTP account", w, fyneApp, appState)
			}
		}, w)
	})

	left := container.NewHBox(icon, container.NewVBox(titleRow, accountTxt))
	codeAndBar := container.NewVBox(codeTxt, progressBar)
	buttons := container.NewHBox(copyBtn, deleteBtn)
	right := container.NewHBox(codeAndBar, buttons)

	content := container.NewBorder(nil, nil, left, right)

	return theme.CardWithHeader("", "", nil, content)
}

func (ns *NavigationState) showAddTOTPDialog() {
	manualIssuer := widget.NewEntry()
	manualIssuer.PlaceHolder = "e.g. GitHub"
	manualAccount := widget.NewEntry()
	manualAccount.PlaceHolder = "user@example.com"
	manualSecret := widget.NewEntry()
	manualSecret.PlaceHolder = "Base32 secret"
	manualAlgo := widget.NewSelect([]string{"SHA1", "SHA256", "SHA512"}, nil)
	manualAlgo.SetSelected("SHA1")
	manualDigits := widget.NewSelect([]string{"6", "7", "8"}, nil)
	manualDigits.SetSelected("6")
	manualPeriod := widget.NewSelect([]string{"30", "60", "90"}, nil)
	manualPeriod.SetSelected("30")

	manualForm := container.NewVBox(
		theme.FieldLabel("ISSUER", nil),
		manualIssuer,
		theme.FieldLabel("ACCOUNT", nil),
		manualAccount,
		theme.FieldLabel("SECRET (BASE32)", nil),
		manualSecret,
		container.NewGridWithColumns(3,
			container.NewVBox(theme.FieldLabel("ALGORITHM", nil), manualAlgo),
			container.NewVBox(theme.FieldLabel("DIGITS", nil), manualDigits),
			container.NewVBox(theme.FieldLabel("PERIOD", nil), manualPeriod),
		),
	)

	// QR Image tab
	qrStatus := canvas.NewText("Select a QR code image to import.", theme.ColorFg2)
	qrStatus.TextSize = 12
	var qrParams []*totp.TOTPParams

	qrSelectBtn := theme.CreateDefaultButton("Select image", func() {
		widgets.PickImageFile("Select QR Code", func(path string) {
			go func() {
				img, err := loadImageFromPath(path)
				if err != nil {
					fyne.Do(func() {
						qrStatus.Text = "Failed to load image: " + err.Error()
						qrStatus.Color = theme.ColorDanger
						qrStatus.Refresh()
					})
					return
				}

				uri, err := totp.DecodeQRFromImage(img)
				if err != nil {
					fyne.Do(func() {
						qrStatus.Text = "QR decode failed: " + err.Error()
						qrStatus.Color = theme.ColorDanger
						qrStatus.Refresh()
					})
					return
				}

				var params []*totp.TOTPParams
				if strings.HasPrefix(uri, "otpauth-migration://") {
					params, err = totp.ParseGoogleAuthExport(uri)
				} else {
					var p *totp.TOTPParams
					p, err = totp.ParseOTPAuthURI(uri)
					if err == nil {
						params = []*totp.TOTPParams{p}
					}
				}
				if err != nil || len(params) == 0 {
					fyne.Do(func() {
						msg := "No TOTP accounts found in QR code"
						if err != nil {
							msg = "QR parse failed: " + err.Error()
						}
						qrStatus.Text = msg
						qrStatus.Color = theme.ColorDanger
						qrStatus.Refresh()
					})
					return
				}

				qrParams = params
				fyne.Do(func() {
					if len(params) == 1 {
						qrStatus.Text = fmt.Sprintf("Found: %s (%s)", params[0].Issuer, params[0].Account)
					} else {
						qrStatus.Text = fmt.Sprintf("Found %d account(s)", len(params))
					}
					qrStatus.Color = theme.ColorSuccess
					qrStatus.Refresh()
				})
			}()
		}, func(err error) {
			widgets.ShowAppError(err, ns.window)
		})
	})

	qrForm := container.NewVBox(qrSelectBtn, qrStatus)

	// Paste URI tab
	uriEntry := widget.NewMultiLineEntry()
	uriEntry.PlaceHolder = "otpauth://totp/..."
	uriEntry.SetMinRowsVisible(3)

	uriStatus := canvas.NewText("", theme.ColorFg2)
	uriStatus.TextSize = 12
	var uriParams *totp.TOTPParams

	uriEntry.OnChanged = func(text string) {
		text = strings.TrimSpace(text)
		if text == "" {
			uriStatus.Text = ""
			uriStatus.Refresh()
			uriParams = nil
			return
		}
		params, err := totp.ParseOTPAuthURI(text)
		if err != nil {
			uriStatus.Text = "Invalid URI: " + err.Error()
			uriStatus.Color = theme.ColorDanger
			uriStatus.Refresh()
			uriParams = nil
			return
		}
		uriParams = params
		uriStatus.Text = fmt.Sprintf("Parsed: %s (%s)", params.Issuer, params.Account)
		uriStatus.Color = theme.ColorSuccess
		uriStatus.Refresh()
	}

	uriForm := container.NewVBox(
		theme.FieldLabel("OTPAUTH URI", nil),
		uriEntry,
		uriStatus,
	)

	// Google Authenticator migration tab
	migrationEntry := widget.NewMultiLineEntry()
	migrationEntry.PlaceHolder = "otpauth-migration://offline?data=..."
	migrationEntry.SetMinRowsVisible(3)

	migrationStatus := canvas.NewText("", theme.ColorFg2)
	migrationStatus.TextSize = 12
	var migrationParams []*totp.TOTPParams

	migrationEntry.OnChanged = func(text string) {
		text = strings.TrimSpace(text)
		if text == "" {
			migrationStatus.Text = ""
			migrationStatus.Refresh()
			migrationParams = nil
			return
		}
		params, err := totp.ParseGoogleAuthExport(text)
		if err != nil {
			migrationStatus.Text = "Parse error: " + err.Error()
			migrationStatus.Color = theme.ColorDanger
			migrationStatus.Refresh()
			migrationParams = nil
			return
		}
		if len(params) == 0 {
			migrationStatus.Text = "No TOTP accounts found in export"
			migrationStatus.Color = theme.ColorDanger
			migrationStatus.Refresh()
			migrationParams = nil
			return
		}
		migrationParams = params
		migrationStatus.Text = fmt.Sprintf("Found %d account(s)", len(params))
		migrationStatus.Color = theme.ColorSuccess
		migrationStatus.Refresh()
	}

	migrationForm := container.NewVBox(
		theme.FieldLabel("GOOGLE AUTHENTICATOR EXPORT URI", nil),
		migrationEntry,
		migrationStatus,
	)

	tabNames := []string{"Manual", "QR Image", "Paste URI", "Google Auth"}
	activeTab := 0

	manualForm.Show()
	qrForm.Hide()
	uriForm.Hide()
	migrationForm.Hide()

	var tabsWidget fyne.CanvasObject
	var dialogFormContent *fyne.Container
	var buildTabs func()

	refreshTabs := func(idx int) {
		activeTab = idx
		manualForm.Hide()
		qrForm.Hide()
		uriForm.Hide()
		migrationForm.Hide()
		switch idx {
		case 1:
			qrForm.Show()
		case 2:
			uriForm.Show()
		case 3:
			migrationForm.Show()
		default:
			manualForm.Show()
		}
		buildTabs()
		if dialogFormContent != nil {
			dialogFormContent.Objects[0] = tabsWidget
			dialogFormContent.Refresh()
		}
	}

	buildTabs = func() {
		tabsWidget = theme.UnderlineTabs(tabNames, activeTab, refreshTabs)
	}
	buildTabs()

	dialogFormContent = container.NewVBox(tabsWidget, manualForm, qrForm, uriForm, migrationForm)

	var customDialog *dialog.CustomDialog

	saveBtn := theme.CreatePrimaryButton("Save", func() {
		var paramsList []*totp.TOTPParams

		switch activeTab {
		case 0:
			if manualSecret.Text == "" || manualIssuer.Text == "" {
				widgets.ShowAppError(fmt.Errorf("issuer and secret are required"), ns.window)
				return
			}
			digits, _ := strconv.Atoi(manualDigits.Selected)
			period, _ := strconv.Atoi(manualPeriod.Selected)
			paramsList = []*totp.TOTPParams{{
				Secret:    manualSecret.Text,
				Algorithm: totp.Algorithm(manualAlgo.Selected),
				Digits:    digits,
				Period:    period,
				Issuer:    manualIssuer.Text,
				Account:   manualAccount.Text,
			}}
		case 1:
			if len(qrParams) == 0 {
				widgets.ShowAppError(fmt.Errorf("no QR code has been decoded yet"), ns.window)
				return
			}
			paramsList = qrParams
		case 2:
			if uriParams == nil {
				widgets.ShowAppError(fmt.Errorf("no valid URI has been parsed yet"), ns.window)
				return
			}
			paramsList = []*totp.TOTPParams{uriParams}
		case 3:
			if len(migrationParams) == 0 {
				widgets.ShowAppError(fmt.Errorf("no accounts found in Google Authenticator export"), ns.window)
				return
			}
			paramsList = migrationParams
		}

		for _, p := range paramsList {
			if err := totp.Validate(p); err != nil {
				widgets.ShowAppError(fmt.Errorf("invalid TOTP (%s): %w", p.Issuer, err), ns.window)
				return
			}
		}

		if customDialog != nil {
			customDialog.Hide()
		}

		go func() {
			ns.appState.Mu.Lock()
			defer ns.appState.Mu.Unlock()

			vaultFile := app.GetVaultPath(ns.appState.CurrentVault)
			entries, err := app.ReadVault(vaultFile, ns.appState.MasterPassword)
			if err != nil {
				fyne.Do(func() {
					widgets.ShowAppError(fmt.Errorf("failed to read vault: %w", err), ns.window)
				})
				return
			}

			for _, params := range paramsList {
				totpJSON, err := totp.Serialize(params)
				if err != nil {
					fyne.Do(func() {
						widgets.ShowAppError(err, ns.window)
					})
					return
				}

				ct, ss, err := app.Encapsulate(ns.appState.PublicKey)
				if err != nil {
					fyne.Do(func() {
						widgets.ShowAppError(fmt.Errorf("encapsulation failed: %v", err), ns.window)
					})
					return
				}

				nonce, ciphertext, err := app.EncryptAES256GCM(string(totpJSON), ss)
				if err != nil {
					fyne.Do(func() {
						widgets.ShowAppError(fmt.Errorf("encryption failed: %v", err), ns.window)
					})
					return
				}

				entry := model.NewVaultEntry()
				entry.Type = model.EntryTypeTOTP
				entry.KyberCiphertext = ct
				entry.Nonce = nonce
				entry.Ciphertext = ciphertext
				entry.Service = "TOTP:" + params.Issuer
				entry.Username = params.Account

				entries = append(entries, entry)
			}

			err = app.WriteVault(entries, vaultFile, ns.appState.MasterPassword)
			if err != nil {
				fyne.Do(func() {
					widgets.ShowAppError(fmt.Errorf("failed to save: %v", err), ns.window)
				})
				return
			}

			fyne.Do(func() {
				msg := "TOTP account added!"
				if len(paramsList) > 1 {
					msg = fmt.Sprintf("%d TOTP accounts imported!", len(paramsList))
				}
				widgets.ShowAppInformation("Success", msg, ns.window)
				ns.switchView(NavViewTOTP)
			})
		}()
	})

	cancelBtn := theme.CreateGhostButton("Cancel", func() {
		if customDialog != nil {
			customDialog.Hide()
		}
	})

	buttonBox := container.NewHBox(cancelBtn, saveBtn)
	dialogContent := container.NewVBox(dialogFormContent, container.NewCenter(buttonBox))
	customDialog = dialog.NewCustomWithoutButtons("Add TOTP Account", dialogContent, ns.window)
	customDialog.Show()
}

// createTOTPItemCard renders an inline TOTP card for the main Items list.
func createTOTPItemCard(index int, entry *model.VaultEntry, payload string, w fyne.Window, fyneApp fyne.App, appState *app.AppState) fyne.CanvasObject {
	params, err := totp.Deserialize([]byte(payload))
	if err != nil {
		errTxt := canvas.NewText("Invalid TOTP data", theme.ColorDanger)
		errTxt.TextSize = 13
		return theme.CardWithHeader("", "", nil, errTxt)
	}

	icon := theme.TypeIcon(theme.IconClock, theme.ColorAccentCyan)

	issuer := params.Issuer
	if issuer == "" {
		issuer = strings.TrimPrefix(entry.Service, "TOTP:")
	}

	titleTxt := canvas.NewText(issuer, theme.ColorTextPrimary)
	titleTxt.TextSize = 13
	titleTxt.TextStyle = fyne.TextStyle{Bold: true}

	badge := theme.KindBadge("TOTP")
	titleRow := container.NewHBox(titleTxt, badge)

	accountTxt := canvas.NewText(params.Account, theme.ColorTextSecondary)
	accountTxt.TextSize = 11
	accountTxt.TextStyle = fyne.TextStyle{Monospace: true}

	codeTxt := canvas.NewText("------", theme.ColorTextPrimary)
	codeTxt.TextSize = 13
	codeTxt.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}

	code, _, genErr := totp.GenerateCode(params)
	if genErr == nil {
		codeTxt.Text = totp.FormatCode(code)
	}

	copyBtn := theme.CreateSmallIconButton(theme.IconCopy, func() {
		c, _, err := totp.GenerateCode(params)
		if err != nil {
			widgets.ShowAppError(err, w)
			return
		}
		clipboardAutoClear(w, c)
		widgets.ShowAppInformation("Copied", "TOTP code copied (auto-clears in 30s)", w)
	})

	deleteBtn := theme.CreateSmallIconButton(theme.IconTrash, func() {
		widgets.ShowAppConfirm("Delete", fmt.Sprintf("Delete TOTP for %s?", issuer), func(ok bool) {
			if ok {
				deleteEntryByID(entry.ID, "TOTP account", w, fyneApp, appState)
			}
		}, w)
	})

	left := container.NewHBox(icon, container.NewVBox(titleRow, accountTxt))
	buttons := container.NewHBox(codeTxt, copyBtn, deleteBtn)
	row := container.NewBorder(nil, nil, left, buttons)

	return theme.CardWithHeader("", "", nil, row)
}

// onViewChange registers a cleanup function called when the navigation changes.
func (ns *NavigationState) onViewChange(cleanup func()) {
	if ns.viewCleanup != nil {
		ns.viewCleanup()
	}
	ns.viewCleanup = cleanup
}

// clipboardAutoClear copies content to the clipboard, then clears it after 30 seconds.
// If the clipboard content has changed in the meantime, it does not overwrite.
func clipboardAutoClear(w fyne.Window, content string) {
	w.Clipboard().SetContent(content)
	go func() {
		time.Sleep(30 * time.Second)
		fyne.Do(func() {
			if w.Clipboard().Content() == content {
				w.Clipboard().SetContent("")
			}
		})
	}()
}

func loadImageFromPath(path string) (image.Image, error) {
	res, err := fyne.LoadResourceFromPath(path)
	if err != nil {
		return nil, fmt.Errorf("load file: %w", err)
	}
	img, _, err := image.Decode(bytes.NewReader(res.Content()))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}
	return img, nil
}
