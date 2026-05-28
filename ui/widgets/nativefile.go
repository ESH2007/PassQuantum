package widgets

import (
	"errors"

	"fyne.io/fyne/v2"
	"github.com/ncruces/zenity"
)

// imageFileFilters are the OS-native filter patterns accepted by the two image
// pickers (icon and theme palette).
var imageFileFilters = []zenity.FileFilter{
	{Name: "Images (PNG / JPEG)", Patterns: []string{"*.png", "*.jpg", "*.jpeg"}, CaseFold: true},
}

// PickImageFile opens the OS-native file dialog in a background goroutine and
// calls onPicked(path) on the Fyne goroutine when the user confirms a choice.
// If the user cancels, onPicked is not called.  Any non-cancel error is passed
// to onErr (also on the Fyne goroutine).
func PickImageFile(title string, onPicked func(path string), onErr func(err error)) {
	go func() {
		path, err := zenity.SelectFile(
			zenity.Title(title),
			zenity.FileFilters(imageFileFilters),
		)
		if err != nil {
			if errors.Is(err, zenity.ErrCanceled) {
				return
			}
			fyne.Do(func() { onErr(err) })
			return
		}
		fyne.Do(func() { onPicked(path) })
	}()
}

// PickAnyFile opens an OS-native file dialog with no format restrictions.
func PickAnyFile(title string, onPicked func(path string), onErr func(err error)) {
	go func() {
		path, err := zenity.SelectFile(
			zenity.Title(title),
		)
		if err != nil {
			if errors.Is(err, zenity.ErrCanceled) {
				return
			}
			fyne.Do(func() { onErr(err) })
			return
		}
		fyne.Do(func() { onPicked(path) })
	}()
}

// PickSaveFile opens an OS-native save dialog with a default filename.
func PickSaveFile(title, defaultName string, onPicked func(path string), onErr func(err error)) {
	go func() {
		path, err := zenity.SelectFileSave(
			zenity.Title(title),
			zenity.Filename(defaultName),
			zenity.ConfirmOverwrite(),
		)
		if err != nil {
			if errors.Is(err, zenity.ErrCanceled) {
				return
			}
			fyne.Do(func() { onErr(err) })
			return
		}
		fyne.Do(func() { onPicked(path) })
	}()
}
