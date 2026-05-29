package screens

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"passquantum/app"
	"passquantum/core/filevault"
	"passquantum/theme"
	"passquantum/ui/widgets"
)

// Preference key for the "delete source file after import" choice.
// Values: "always" | "never" | "" (ask each time).
const PrefDeleteSourceAfterImport = "pref_delete_source_after_import"

func (ns *NavigationState) createFilesView() fyne.CanvasObject {
	storeBtn := theme.CreatePrimaryButtonWithIcon("Store file", theme.IconFileUp, func() {
		ns.pickAndStoreFile()
	})

	header := theme.PageHeader(
		"PASSQUANTUM / "+ns.appState.CurrentVault+" / FILES",
		"Secure files",
		"Encrypted file storage with per-file post-quantum key encapsulation.",
		storeBtn,
	)

	summaryLabel := canvas.NewText("Loading...", theme.ColorTextSecondary)
	summaryLabel.TextSize = 13

	itemsContainer := container.NewVBox()
	loadingText := canvas.NewText("Loading encrypted files...", theme.ColorFg2)
	loadingText.TextSize = 13
	itemsContainer.Objects = []fyne.CanvasObject{container.NewCenter(loadingText)}

	type fileCard struct {
		name string
		card fyne.CanvasObject
	}
	var allCards []fileCard

	searchEntry := widget.NewEntry()
	searchEntry.PlaceHolder = "Search files..."
	searchEntry.OnChanged = func(query string) {
		q := strings.ToLower(query)
		var filtered []fyne.CanvasObject
		for _, fc := range allCards {
			if q == "" || strings.Contains(strings.ToLower(fc.name), q) {
				filtered = append(filtered, fc.card)
			}
		}
		if len(filtered) == 0 && query != "" {
			noMatch := canvas.NewText("No files match “"+query+"”", theme.ColorFg2)
			noMatch.TextSize = 13
			itemsContainer.Objects = []fyne.CanvasObject{container.NewCenter(noMatch)}
		} else {
			itemsContainer.Objects = filtered
		}
		itemsContainer.Refresh()
	}

	go func() {
		store, err := ns.ensureFileStore()
		if err != nil {
			fyne.Do(func() {
				errText := canvas.NewText("Failed to init file store: "+err.Error(), theme.ColorDanger)
				errText.TextSize = 13
				itemsContainer.Objects = []fyne.CanvasObject{errText}
				itemsContainer.Refresh()
				summaryLabel.Text = "Error"
				summaryLabel.Refresh()
			})
			return
		}

		files := store.ListFiles()

		fyne.Do(func() {
			if len(files) == 0 {
				fileIco := canvas.NewImageFromResource(theme.IconFolder)
				fileIco.SetMinSize(fyne.NewSize(40, 40))
				emptyTitle := canvas.NewText("No files stored yet", theme.ColorTextPrimary)
				emptyTitle.TextSize = 15
				emptyTitle.TextStyle = fyne.TextStyle{Bold: true}
				emptySubtitle := canvas.NewText("Store your first encrypted file in this vault.", theme.ColorFg2)
				emptySubtitle.TextSize = 12
				addFirstBtn := theme.CreatePrimaryButton("Store file", func() {
					ns.pickAndStoreFile()
				})
				emptyState := container.NewCenter(container.NewVBox(
					container.NewCenter(fileIco),
					container.NewCenter(emptyTitle),
					container.NewCenter(emptySubtitle),
					container.NewCenter(addFirstBtn),
				))
				itemsContainer.Objects = []fyne.CanvasObject{emptyState}
				itemsContainer.Refresh()
				summaryLabel.Text = "No files"
				summaryLabel.Refresh()
				return
			}

			var cards []fyne.CanvasObject
			var totalSize int64
			for _, meta := range files {
				card := buildFileCard(meta, store, ns.window, ns.app, ns.appState, func() {
					ns.switchView(NavViewFiles)
				})
				allCards = append(allCards, fileCard{name: meta.OriginalName, card: card})
				cards = append(cards, card)
				totalSize += meta.Size
			}

			itemsContainer.Objects = cards
			itemsContainer.Refresh()

			countStr := fmt.Sprintf("%d file", len(files))
			if len(files) != 1 {
				countStr += "s"
			}
			summaryLabel.Text = fmt.Sprintf("%s • %s total", countStr, humanSize(totalSize))
			summaryLabel.Refresh()
		})
	}()

	searchBg := canvas.NewRectangle(theme.ColorSidebarBg)
	searchBg.CornerRadius = theme.RadiusInput
	searchBorder := canvas.NewRectangle(color.Transparent)
	searchBorder.CornerRadius = theme.RadiusInput
	searchBorder.StrokeWidth = 1
	searchBorder.StrokeColor = theme.ColorLine2
	searchBorder.FillColor = color.Transparent
	searchRow := container.NewStack(searchBg, searchBorder, container.NewPadded(searchEntry))

	return container.NewVBox(header, summaryLabel, searchRow, itemsContainer)
}

const maxThumbnailFileSize = 10 * 1024 * 1024 // 10 MB

func buildFileCard(meta *filevault.FileMetadata, store *filevault.Store, w fyne.Window, fyneApp fyne.App, appState *app.AppState, onRefresh func()) fyne.CanvasObject {
	icon := theme.TypeIcon(fileIconForExt(filepath.Ext(meta.OriginalName)), theme.ColorAccentCyan)

	titleTxt := canvas.NewText(meta.OriginalName, theme.ColorTextPrimary)
	titleTxt.TextSize = 13
	titleTxt.TextStyle = fyne.TextStyle{Bold: true}

	ext := strings.TrimPrefix(filepath.Ext(meta.OriginalName), ".")
	if ext == "" {
		ext = "FILE"
	}
	badge := theme.KindBadge(strings.ToUpper(ext))
	titleRow := container.NewHBox(titleTxt, badge)

	details := canvas.NewText(
		fmt.Sprintf("%s • %s", humanSize(meta.Size), meta.StoredAt.Format("2006-01-02 15:04")),
		theme.ColorTextSecondary,
	)
	details.TextSize = 11
	details.TextStyle = fyne.TextStyle{Monospace: true}

	// Image thumbnail (loaded async, only for image files under 10 MB)
	var thumbnailContainer *fyne.Container
	if isImageMime(meta.MimeType) && meta.Size <= maxThumbnailFileSize {
		placeholder := canvas.NewRectangle(theme.ColorBg3)
		placeholder.SetMinSize(fyne.NewSize(64, 64))
		placeholder.CornerRadius = 4
		thumbnailContainer = container.NewGridWrap(fyne.NewSize(64, 64), placeholder)

		go func() {
			data, err := store.DecryptToMemory(meta.UUID)
			if err != nil {
				return
			}
			img, _, err := image.Decode(bytes.NewReader(data))
			if err != nil {
				return
			}
			fyne.Do(func() {
				thumb := canvas.NewImageFromImage(img)
				thumb.FillMode = canvas.ImageFillContain
				thumb.SetMinSize(fyne.NewSize(64, 64))
				thumbnailContainer.Objects = []fyne.CanvasObject{thumb}
				thumbnailContainer.Refresh()
			})
		}()
	}

	openBtn := theme.CreateSmallIconButton(theme.IconExternalLink, func() {
		go func() {
			_, err := store.OpenFile(meta.UUID)
			if err != nil {
				fyne.Do(func() {
					widgets.ShowAppError(fmt.Errorf("open file: %w", err), w)
				})
			}
		}()
	})

	exportBtn := theme.CreateSmallIconButton(theme.IconFileUp, func() {
		widgets.PickSaveFile("Export file", meta.OriginalName, func(dstPath string) {
			go func() {
				err := store.RetrieveFile(meta.UUID, dstPath, nil)
				fyne.Do(func() {
					if err != nil {
						widgets.ShowAppError(fmt.Errorf("export: %w", err), w)
					} else {
						widgets.ShowAppInformation("Exported", "File exported to "+dstPath, w)
					}
				})
			}()
		}, func(err error) {
			widgets.ShowAppError(err, w)
		})
	})

	deleteBtn := theme.CreateSmallIconButton(theme.IconTrash, func() {
		widgets.ShowAppConfirm("Delete", fmt.Sprintf("Delete '%s'? This cannot be undone.", meta.OriginalName), func(ok bool) {
			if !ok {
				return
			}
			go func() {
				err := store.DeleteFile(meta.UUID)
				fyne.Do(func() {
					if err != nil {
						widgets.ShowAppError(fmt.Errorf("delete: %w", err), w)
					} else {
						widgets.ShowAppInformation("Deleted", "File deleted successfully", w)
						if onRefresh != nil {
							onRefresh()
						}
					}
				})
			}()
		}, w)
	})

	var left fyne.CanvasObject
	if thumbnailContainer != nil {
		left = container.NewHBox(thumbnailContainer, container.NewVBox(titleRow, details))
	} else {
		left = container.NewHBox(icon, container.NewVBox(titleRow, details))
	}
	buttons := container.NewHBox(openBtn, exportBtn, deleteBtn)
	row := container.NewBorder(nil, nil, left, buttons)

	return theme.CardWithHeader("", "", nil, row)
}

func (ns *NavigationState) pickAndStoreFile() {
	widgets.PickAnyFile("Select file to encrypt", func(srcPath string) {
		go func() {
			store, err := ns.ensureFileStore()
			if err != nil {
				fyne.Do(func() {
					widgets.ShowAppError(err, ns.window)
				})
				return
			}

			_, err = store.StoreFile(srcPath, nil)
			fyne.Do(func() {
				if err != nil {
					widgets.ShowAppError(fmt.Errorf("store file: %w", err), ns.window)
					return
				}
				widgets.ShowAppInformation("Stored", "File encrypted and stored in vault", ns.window)
				ns.switchView(NavViewFiles)
				ns.handleSourceFileCleanup(srcPath)
			})
		}()
	}, func(err error) {
		widgets.ShowAppError(err, ns.window)
	})
}

// handleSourceFileCleanup deletes (or prompts to delete) the original
// unencrypted source file after a successful encrypted store. Honors the
// "always" / "never" / "ask" preference.
func (ns *NavigationState) handleSourceFileCleanup(srcPath string) {
	prefs := ns.app.Preferences()
	choice := prefs.StringWithFallback(PrefDeleteSourceAfterImport, "")

	switch choice {
	case "always":
		ns.deleteSourceFile(srcPath, false)
		return
	case "never":
		return
	}

	// Ask the user
	fileName := filepath.Base(srcPath)
	widgets.ShowAppConfirmWithRemember(
		"Delete source file?",
		fmt.Sprintf("The encrypted copy is now in the vault, but the original '%s' is still on your disk in plain text. Delete it now?", fileName),
		"Don't ask me again",
		func(confirmed, remember bool) {
			if remember {
				if confirmed {
					prefs.SetString(PrefDeleteSourceAfterImport, "always")
				} else {
					prefs.SetString(PrefDeleteSourceAfterImport, "never")
				}
			}
			if confirmed {
				ns.deleteSourceFile(srcPath, true)
			}
		},
		ns.window,
	)
}

func (ns *NavigationState) deleteSourceFile(srcPath string, notify bool) {
	if err := os.Remove(srcPath); err != nil {
		log.Printf("[FileVault] failed to delete source file %q: %v", srcPath, err)
		widgets.ShowAppError(fmt.Errorf("could not delete source file: %w", err), ns.window)
		return
	}
	if notify {
		widgets.ShowAppInformation("Deleted", "Source file removed from disk", ns.window)
	}
}

// ensureFileStore lazily initializes the file store if needed.
func (ns *NavigationState) ensureFileStore() (*filevault.Store, error) {
	ns.appState.Mu.Lock()
	defer ns.appState.Mu.Unlock()

	if ns.appState.FileStore != nil {
		return ns.appState.FileStore, nil
	}

	if err := app.InitFileStore(ns.appState); err != nil {
		log.Printf("[FileVault] WARNING: %v", err)
		return nil, err
	}
	return ns.appState.FileStore, nil
}

func fileIconForExt(ext string) *fyne.StaticResource {
	switch strings.ToLower(ext) {
	case ".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt":
		return theme.IconFile
	default:
		return theme.IconFile
	}
}

func isImageMime(mime string) bool {
	return strings.HasPrefix(mime, "image/")
}

func humanSize(bytes int64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
	case bytes >= gb:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(gb))
	case bytes >= mb:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(mb))
	case bytes >= kb:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(kb))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
