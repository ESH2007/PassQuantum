package screens

import (
	"fmt"
	"image/color"
	"math"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"passquantum/app"
	"passquantum/theme"
	"passquantum/ui/widgets"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  HSV ↔ NRGBA
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func hsvToNRGBA(h, s, v float64, a uint8) color.NRGBA {
	if s == 0 {
		c := uint8(v * 255)
		return color.NRGBA{R: c, G: c, B: c, A: a}
	}
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}
	seg := h / 60
	i := int(seg)
	f := seg - float64(i)
	p := v * (1 - s)
	q := v * (1 - s*f)
	t := v * (1 - s*(1-f))
	var r, g, b float64
	switch i {
	case 0:
		r, g, b = v, t, p
	case 1:
		r, g, b = q, v, p
	case 2:
		r, g, b = p, v, t
	case 3:
		r, g, b = p, q, v
	case 4:
		r, g, b = t, p, v
	default:
		r, g, b = v, p, q
	}
	return color.NRGBA{R: uint8(r * 255), G: uint8(g * 255), B: uint8(b * 255), A: a}
}

func nrgbaToHSV(c color.NRGBA) (h, s, v float64) {
	r := float64(c.R) / 255
	g := float64(c.G) / 255
	b := float64(c.B) / 255
	mx := math.Max(r, math.Max(g, b))
	mn := math.Min(r, math.Min(g, b))
	d := mx - mn
	v = mx
	if mx == 0 {
		s = 0
	} else {
		s = d / mx
	}
	if d == 0 {
		h = 0
		return
	}
	switch mx {
	case r:
		h = 60 * math.Mod((g-b)/d, 6)
	case g:
		h = 60 * ((b-r)/d + 2)
	default:
		h = 60 * ((r-g)/d + 4)
	}
	if h < 0 {
		h += 360
	}
	return
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  SVPicker – 2-D saturation / value selector square
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

type SVPicker struct {
	widget.BaseWidget
	Hue       float64
	S, V      float64
	OnChanged func(s, v float64)
}

func newSVPicker(hue, s, v float64) *SVPicker {
	p := &SVPicker{Hue: hue, S: s, V: v}
	p.ExtendBaseWidget(p)
	return p
}

func (p *SVPicker) applyPos(pos fyne.Position) {
	sz := p.Size()
	if sz.Width == 0 || sz.Height == 0 {
		return
	}
	p.S = math.Max(0, math.Min(1, float64(pos.X)/float64(sz.Width)))
	p.V = math.Max(0, math.Min(1, 1.0-float64(pos.Y)/float64(sz.Height)))
	p.Refresh()
	if p.OnChanged != nil {
		p.OnChanged(p.S, p.V)
	}
}

func (p *SVPicker) Tapped(e *fyne.PointEvent) { p.applyPos(e.Position) }
func (p *SVPicker) Dragged(e *fyne.DragEvent) { p.applyPos(e.Position) }
func (p *SVPicker) DragEnd()                  {}

func (p *SVPicker) CreateRenderer() fyne.WidgetRenderer {
	raster := canvas.NewRasterWithPixels(func(x, y, w, h int) color.Color {
		if w <= 1 || h <= 1 {
			return color.NRGBA{A: 255}
		}
		s := float64(x) / float64(w-1)
		v := 1.0 - float64(y)/float64(h-1)
		return hsvToNRGBA(p.Hue, s, v, 255)
	})

	cursor := canvas.NewCircle(color.Transparent)
	cursor.StrokeColor = color.White
	cursor.StrokeWidth = 2.5

	cursorShadow := canvas.NewCircle(color.Transparent)
	cursorShadow.StrokeColor = color.NRGBA{A: 180}
	cursorShadow.StrokeWidth = 1.5

	border := canvas.NewRectangle(color.Transparent)
	border.StrokeColor = theme.ColorBorderCyan
	border.StrokeWidth = 1
	border.CornerRadius = 4

	return &svPickerRenderer{
		p: p, raster: raster, cursor: cursor,
		cursorShadow: cursorShadow, border: border,
	}
}

const svCursorR = float32(10)

type svPickerRenderer struct {
	p            *SVPicker
	raster       *canvas.Raster
	cursor       *canvas.Circle
	cursorShadow *canvas.Circle
	border       *canvas.Rectangle
}

func (r *svPickerRenderer) MinSize() fyne.Size { return fyne.NewSize(220, 220) }

func (r *svPickerRenderer) Layout(size fyne.Size) {
	r.raster.Resize(size)
	r.raster.Move(fyne.NewPos(0, 0))
	r.border.Resize(size)
	r.border.Move(fyne.NewPos(0, 0))

	cx := float32(r.p.S) * size.Width
	cy := float32(1-r.p.V) * size.Height

	r.cursorShadow.Move(fyne.NewPos(cx-svCursorR-1.5, cy-svCursorR-1.5))
	r.cursorShadow.Resize(fyne.NewSize(svCursorR*2+3, svCursorR*2+3))

	r.cursor.Move(fyne.NewPos(cx-svCursorR, cy-svCursorR))
	r.cursor.Resize(fyne.NewSize(svCursorR*2, svCursorR*2))
}

func (r *svPickerRenderer) Refresh() {
	r.Layout(r.p.Size())
	r.raster.Refresh()
	r.cursor.Refresh()
	r.cursorShadow.Refresh()
	r.border.Refresh()
}

func (r *svPickerRenderer) Destroy()                     {}
func (r *svPickerRenderer) BackgroundColor() color.Color { return color.Transparent }
func (r *svPickerRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.raster, r.cursorShadow, r.cursor, r.border}
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  GradientSlider – horizontal gradient bar with draggable thumb
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

type GradientSlider struct {
	widget.BaseWidget
	Value     float64
	PixelFunc func(t float64) color.NRGBA
	OnChanged func(v float64)
}

func newGradientSlider(pf func(float64) color.NRGBA, initial float64, onChange func(float64)) *GradientSlider {
	s := &GradientSlider{Value: initial, PixelFunc: pf, OnChanged: onChange}
	s.ExtendBaseWidget(s)
	return s
}

func (s *GradientSlider) applyPos(pos fyne.Position) {
	sz := s.Size()
	if sz.Width == 0 {
		return
	}
	v := math.Max(0, math.Min(1, float64(pos.X)/float64(sz.Width)))
	s.Value = v
	s.Refresh()
	if s.OnChanged != nil {
		s.OnChanged(v)
	}
}

func (s *GradientSlider) Tapped(e *fyne.PointEvent) { s.applyPos(e.Position) }
func (s *GradientSlider) Dragged(e *fyne.DragEvent) { s.applyPos(e.Position) }
func (s *GradientSlider) DragEnd()                  {}

func (s *GradientSlider) CreateRenderer() fyne.WidgetRenderer {
	raster := canvas.NewRasterWithPixels(func(x, y, w, h int) color.Color {
		if w <= 1 {
			return color.NRGBA{A: 255}
		}
		t := float64(x) / float64(w-1)
		return s.PixelFunc(t)
	})

	track := canvas.NewRectangle(color.Transparent)
	track.CornerRadius = 8
	track.StrokeColor = color.NRGBA{R: 100, G: 110, B: 130, A: 100}
	track.StrokeWidth = 1

	thumb := canvas.NewCircle(color.White)
	thumb.StrokeColor = color.NRGBA{R: 160, G: 170, B: 185, A: 255}
	thumb.StrokeWidth = 1.5

	return &gradientSliderRenderer{s: s, raster: raster, track: track, thumb: thumb}
}

const (
	gsBarH   = float32(20)
	gsThumbR = float32(12)
)

type gradientSliderRenderer struct {
	s      *GradientSlider
	raster *canvas.Raster
	track  *canvas.Rectangle
	thumb  *canvas.Circle
}

func (r *gradientSliderRenderer) MinSize() fyne.Size { return fyne.NewSize(200, gsThumbR*2) }

func (r *gradientSliderRenderer) Layout(size fyne.Size) {
	barY := (size.Height - gsBarH) / 2
	r.raster.Resize(fyne.NewSize(size.Width, gsBarH))
	r.raster.Move(fyne.NewPos(0, barY))
	r.track.Resize(fyne.NewSize(size.Width, gsBarH))
	r.track.Move(fyne.NewPos(0, barY))

	cx := float32(r.s.Value) * size.Width
	cy := size.Height / 2
	r.thumb.Move(fyne.NewPos(cx-gsThumbR, cy-gsThumbR))
	r.thumb.Resize(fyne.NewSize(gsThumbR*2, gsThumbR*2))
}

func (r *gradientSliderRenderer) Refresh() {
	r.Layout(r.s.Size())
	r.raster.Refresh()
	r.track.Refresh()
	r.thumb.Refresh()
}

func (r *gradientSliderRenderer) Destroy()                     {}
func (r *gradientSliderRenderer) BackgroundColor() color.Color { return color.Transparent }
func (r *gradientSliderRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.raster, r.track, r.thumb}
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  colorPickerState
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

type colorPickerState struct {
	h, s, v float64
	a       uint8

	svPicker *SVPicker
	hueSl    *GradientSlider
	alphaSl  *GradientSlider

	hexEntry *widget.Entry
	rEntry   *widget.Entry
	gEntry   *widget.Entry
	bEntry   *widget.Entry
	aEntry   *widget.Entry

	colorPreview *canvas.Rectangle

	updating bool
}

func newColorPickerState(initial color.NRGBA) *colorPickerState {
	st := &colorPickerState{}
	st.h, st.s, st.v = nrgbaToHSV(initial)
	st.a = initial.A

	st.svPicker = newSVPicker(st.h, st.s, st.v)
	st.svPicker.OnChanged = func(ns, nv float64) {
		if st.updating {
			return
		}
		st.s, st.v = ns, nv
		st.syncAll(false, true)
	}

	st.hueSl = newGradientSlider(
		func(t float64) color.NRGBA { return hsvToNRGBA(t*360, 1, 1, 255) },
		st.h/360,
		func(val float64) {
			if st.updating {
				return
			}
			st.h = val * 360
			st.svPicker.Hue = st.h
			st.syncAll(false, true)
		},
	)

	st.alphaSl = newGradientSlider(
		func(t float64) color.NRGBA {
			base := hsvToNRGBA(st.h, st.s, st.v, 255)
			return color.NRGBA{R: base.R, G: base.G, B: base.B, A: uint8(t * 255)}
		},
		float64(initial.A)/255,
		func(val float64) {
			if st.updating {
				return
			}
			st.a = uint8(val * 255)
			st.syncAll(false, false)
		},
	)

	st.colorPreview = canvas.NewRectangle(initial)
	st.colorPreview.CornerRadius = 6
	st.colorPreview.SetMinSize(fyne.NewSize(38, 38))

	rgb := hsvToNRGBA(st.h, st.s, st.v, st.a)
	st.hexEntry = widget.NewEntry()
	st.hexEntry.SetText(fmt.Sprintf("#%02X%02X%02X", rgb.R, rgb.G, rgb.B))
	st.rEntry = widget.NewEntry()
	st.rEntry.SetText(strconv.Itoa(int(rgb.R)))
	st.gEntry = widget.NewEntry()
	st.gEntry.SetText(strconv.Itoa(int(rgb.G)))
	st.bEntry = widget.NewEntry()
	st.bEntry.SetText(strconv.Itoa(int(rgb.B)))
	st.aEntry = widget.NewEntry()
	st.aEntry.SetText(strconv.Itoa(int(st.a)))

	st.hexEntry.OnChanged = func(text string) {
		if st.updating {
			return
		}
		c, err := parseHexColor(strings.TrimSpace(text))
		if err != nil {
			return
		}
		st.h, st.s, st.v = nrgbaToHSV(c)
		st.syncAll(true, false)
	}

	wireRGB := func() {
		if st.updating {
			return
		}
		r64, e1 := strconv.ParseUint(strings.TrimSpace(st.rEntry.Text), 10, 8)
		g64, e2 := strconv.ParseUint(strings.TrimSpace(st.gEntry.Text), 10, 8)
		b64, e3 := strconv.ParseUint(strings.TrimSpace(st.bEntry.Text), 10, 8)
		if e1 != nil || e2 != nil || e3 != nil {
			return
		}
		st.h, st.s, st.v = nrgbaToHSV(color.NRGBA{R: uint8(r64), G: uint8(g64), B: uint8(b64), A: 255})
		st.syncAll(false, false)
	}
	st.rEntry.OnChanged = func(_ string) { wireRGB() }
	st.gEntry.OnChanged = func(_ string) { wireRGB() }
	st.bEntry.OnChanged = func(_ string) { wireRGB() }

	st.aEntry.OnChanged = func(text string) {
		if st.updating {
			return
		}
		val, err := strconv.ParseUint(strings.TrimSpace(text), 10, 8)
		if err != nil {
			return
		}
		st.a = uint8(val)
		st.syncAll(false, false)
	}

	return st
}

func (st *colorPickerState) syncAll(skipHex, skipSVPos bool) {
	if st.updating {
		return
	}
	st.updating = true
	defer func() { st.updating = false }()

	rgb := hsvToNRGBA(st.h, st.s, st.v, st.a)

	st.colorPreview.FillColor = rgb
	st.colorPreview.Refresh()

	st.svPicker.Hue = st.h
	if !skipSVPos {
		st.svPicker.S = st.s
		st.svPicker.V = st.v
	}
	st.svPicker.Refresh()

	st.hueSl.Value = st.h / 360
	st.hueSl.Refresh()

	st.alphaSl.Value = float64(st.a) / 255
	st.alphaSl.Refresh()

	if !skipHex {
		st.hexEntry.SetText(fmt.Sprintf("#%02X%02X%02X", rgb.R, rgb.G, rgb.B))
	}
	st.rEntry.SetText(strconv.Itoa(int(rgb.R)))
	st.gEntry.SetText(strconv.Itoa(int(rgb.G)))
	st.bEntry.SetText(strconv.Itoa(int(rgb.B)))
	st.aEntry.SetText(strconv.Itoa(int(st.a)))
}

func (st *colorPickerState) current() color.NRGBA {
	return hsvToNRGBA(st.h, st.s, st.v, 255)
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  buildColorPickerPanel – layout matching the reference image
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func buildColorPickerPanel(_ string, st *colorPickerState) fyne.CanvasObject {
	selectorLabel := theme.CreateLabel("Selector", 9, theme.ColorTextSec, false)
	leftCol := container.NewVBox(selectorLabel, st.svPicker)

	hexLabel := theme.CreateLabel("HEX", 9, theme.ColorTextSec, false)
	eyeBtn := widget.NewButton("✎", func() {})
	hexWrap := container.NewGridWrap(fyne.NewSize(156, 36), st.hexEntry)
	eyeWrap := container.NewGridWrap(fyne.NewSize(36, 36), eyeBtn)

	previewBg := canvas.NewRectangle(color.NRGBA{R: 60, G: 65, B: 78, A: 255})
	previewBg.CornerRadius = 6
	previewBg.SetMinSize(fyne.NewSize(38, 38))
	previewStack := container.NewStack(previewBg, st.colorPreview)

	topRight := container.NewHBox(
		container.NewVBox(
			hexLabel,
			container.NewHBox(hexWrap, eyeWrap),
		),
		canvas.NewRectangle(color.Transparent),
		container.NewVBox(
			canvas.NewText(" ", color.Transparent),
			previewStack,
		),
	)

	rLbl := canvas.NewText("R", color.NRGBA{R: 239, G: 68, B: 68, A: 255})
	rLbl.TextSize = 9
	gLbl := canvas.NewText("G", color.NRGBA{R: 34, G: 197, B: 94, A: 255})
	gLbl.TextSize = 9
	bLbl := canvas.NewText("B", color.NRGBA{R: 99, G: 179, B: 237, A: 255})
	bLbl.TextSize = 9
	aLbl := canvas.NewText("A", color.NRGBA{R: 190, G: 195, B: 210, A: 255})
	aLbl.TextSize = 9

	ew := fyne.NewSize(55, 34)
	labelsRow := container.NewHBox(
		container.NewGridWrap(ew, container.NewCenter(rLbl)),
		container.NewGridWrap(ew, container.NewCenter(gLbl)),
		container.NewGridWrap(ew, container.NewCenter(bLbl)),
		container.NewGridWrap(ew, container.NewCenter(aLbl)),
	)
	entriesRow := container.NewHBox(
		container.NewGridWrap(ew, st.rEntry),
		container.NewGridWrap(ew, st.gEntry),
		container.NewGridWrap(ew, st.bEntry),
		container.NewGridWrap(ew, st.aEntry),
	)

	hueWrap := container.NewGridWrap(fyne.NewSize(240, gsThumbR*2), st.hueSl)
	alphaWrap := container.NewGridWrap(fyne.NewSize(240, gsThumbR*2), st.alphaSl)

	rightCol := container.NewVBox(
		topRight,
		widget.NewLabel(""),
		labelsRow,
		entriesRow,
		widget.NewLabel(""),
		hueWrap,
		widget.NewLabel(""),
		alphaWrap,
	)

	gap := canvas.NewRectangle(color.Transparent)
	gap.SetMinSize(fyne.NewSize(14, 1))

	return container.NewHBox(leftCol, gap, rightCol)
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ShowColorPersonalizationDialog
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func ShowColorPersonalizationDialog(w fyne.Window, fyneApp fyne.App, appState *app.AppState) {
	bgSt := newColorPickerState(theme.ColorBg)
	primarySt := newColorPickerState(theme.ColorPrimaryButton)
	secondSt := newColorPickerState(theme.ColorSecondaryButton)

	tabNames := []string{"Background", "Main Buttons", "Secondary Buttons"}
	states := []*colorPickerState{bgSt, primarySt, secondSt}
	selected := 0

	tabStrip := container.NewGridWithColumns(3)
	content := container.NewMax()

	var rebuildTabs func()
	rebuildTabs = func() {
		tabStrip.Objects = []fyne.CanvasObject{
			theme.CreateTabButton(tabNames[0], selected == 0, func() { selected = 0; rebuildTabs() }),
			theme.CreateTabButton(tabNames[1], selected == 1, func() { selected = 1; rebuildTabs() }),
			theme.CreateTabButton(tabNames[2], selected == 2, func() { selected = 2; rebuildTabs() }),
		}
		tabStrip.Refresh()
		content.Objects = []fyne.CanvasObject{
			container.NewPadded(buildColorPickerPanel(tabNames[selected], states[selected])),
		}
		content.Refresh()
	}
	rebuildTabs()

	var d dialog.Dialog

	applyBtn := theme.CreatePrimaryButton("Apply colors", func() {
		palette := []color.NRGBA{bgSt.current(), primarySt.current(), secondSt.current()}
		fyneApp.Settings().SetTheme(fynetheme.DefaultTheme())
		applyExtractedPalette(palette)
		persistManualPalettePreferences(fyneApp, palette)
		if d != nil {
			d.Hide()
		}
		ShowSettingsScreen(w, fyneApp, appState)
		widgets.ShowAppInformation(
			"Colors Applied",
			fmt.Sprintf("Background: %s\nMain Buttons: %s\nSecondary: %s",
				toHex(palette[0]), toHex(palette[1]), toHex(palette[2])),
			w,
		)
	})

	cancelBtn := theme.CreateGhostButton("Cancel", func() {
		if d != nil {
			d.Hide()
		}
	})

	titleLabel := theme.SectionEyebrow("MANUAL COLOR PERSONALIZATION")
	divider := canvas.NewRectangle(theme.ColorLine1)
	divider.SetMinSize(fyne.NewSize(560, 1))

	body := container.NewVBox(
		container.NewCenter(titleLabel),
		divider,
		widget.NewLabel(""),
		tabStrip,
		widget.NewLabel(""),
		content,
		widget.NewLabel(""),
		container.NewCenter(container.NewHBox(cancelBtn, applyBtn)),
	)

	card := theme.CreateCard(container.NewPadded(body), 620, 0, true)
	d = dialog.NewCustomWithoutButtons("", container.NewPadded(card), w)
	d.Show()
}
