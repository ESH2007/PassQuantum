package screens

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"passquantum/app"
	"passquantum/theme"
)

func (ns *NavigationState) createItemsView() fyne.CanvasObject {
	addItemBtn := theme.CreatePrimaryButtonWithIcon("Add item", theme.IconPlus, func() {
		ns.switchView(NavViewAddItem)
	})

	countLabel := canvas.NewText("Loading…", theme.ColorTextSecondary)
	countLabel.TextSize = 13

	// Track all decrypted cards for filtering
	type cardEntry struct {
		service  string
		username string
		card     fyne.CanvasObject
	}
	var allCards []cardEntry

	itemsContainer := container.NewVBox()
	loadingText := canvas.NewText("Loading vault items...", theme.ColorFg2)
	loadingText.TextSize = 13
	itemsContainer.Objects = []fyne.CanvasObject{container.NewCenter(loadingText)}

	// Search bar
	searchEntry := widget.NewEntry()
	searchEntry.PlaceHolder = "Search items…"
	searchEntry.OnChanged = func(query string) {
		q := strings.ToLower(query)
		var filtered []fyne.CanvasObject
		for _, ce := range allCards {
			if q == "" || strings.Contains(strings.ToLower(ce.service), q) || strings.Contains(strings.ToLower(ce.username), q) {
				filtered = append(filtered, ce.card)
			}
		}
		if len(filtered) == 0 && query != "" {
			noMatch := canvas.NewText("No items match “"+query+"”", theme.ColorFg2)
			noMatch.TextSize = 13
			itemsContainer.Objects = []fyne.CanvasObject{container.NewCenter(noMatch)}
		} else {
			itemsContainer.Objects = filtered
		}
		itemsContainer.Refresh()
	}

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
				countLabel.Text = "Error"
				countLabel.Refresh()
			})
			return
		}

		fyne.Do(func() {
			if len(entries) == 0 {
				// Rich empty state
				vaultIco := canvas.NewImageFromResource(theme.IconVault)
				vaultIco.SetMinSize(fyne.NewSize(40, 40))
				emptyTitle := canvas.NewText("No items yet", theme.ColorTextPrimary)
				emptyTitle.TextSize = 15
				emptyTitle.TextStyle = fyne.TextStyle{Bold: true}
				emptySubtitle := canvas.NewText("Add your first credential to this vault.", theme.ColorFg2)
				emptySubtitle.TextSize = 12
				addFirstBtn := theme.CreatePrimaryButton("Add item", func() {
					ns.switchView(NavViewAddItem)
				})
				emptyState := container.NewCenter(container.NewVBox(
					container.NewCenter(vaultIco),
					container.NewCenter(emptyTitle),
					container.NewCenter(emptySubtitle),
					container.NewCenter(addFirstBtn),
				))
				itemsContainer.Objects = []fyne.CanvasObject{emptyState}
				itemsContainer.Refresh()
				countLabel.Text = "No items"
				countLabel.Refresh()
				return
			}

			allCards = nil
			var cards []fyne.CanvasObject
			for _, entry := range entries {
				ss, err := app.Decapsulate(entry.KyberCiphertext, ns.appState.PrivateKey)
				if err != nil {
					continue
				}
				plaintext, err := app.DecryptAES256GCM(entry.Nonce, entry.Ciphertext, ss)
				if err != nil {
					continue
				}
				card := createVaultItemCard(0, entry, plaintext, ns.window, ns.app, ns.appState)
				allCards = append(allCards, cardEntry{
					service:  entry.Service,
					username: entry.Username,
					card:     card,
				})
				cards = append(cards, card)
			}

			itemsContainer.Objects = cards
			itemsContainer.Refresh()

			n := len(allCards)
			if n == 1 {
				countLabel.Text = "1 item"
			} else {
				countLabel.Text = fmt.Sprintf("%d items", n)
			}
			countLabel.Refresh()
		})
	}()

	// Build header manually to include a live countLabel
	eyebrow := canvas.NewText("PASSQUANTUM / "+ns.appState.CurrentVault, theme.ColorFg2)
	eyebrow.TextSize = 10
	eyebrow.TextStyle = fyne.TextStyle{Monospace: true}
	headerTitle := canvas.NewText("Vault items", theme.ColorTextPrimary)
	headerTitle.TextSize = 22
	headerTitle.TextStyle = fyne.TextStyle{Bold: true}
	headerLeft := container.NewVBox(eyebrow, headerTitle, countLabel)
	headerRow := container.NewBorder(nil, nil, headerLeft, container.NewCenter(addItemBtn))
	headerDivider := canvas.NewRectangle(theme.ColorLine1)
	headerDivider.SetMinSize(fyne.NewSize(0, 1))
	header := container.NewVBox(
		container.New(layout.NewCustomPaddedLayout(0, theme.Space4, 0, 0), headerRow),
		headerDivider,
	)

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
