package theme

import (
	"image"
	"image/color"
	"math"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// PillTone selects the color set for a StatusPill.
type PillTone int

const (
	PillAccent PillTone = iota
	PillOk
	PillWarn
	PillMute
	PillDanger
)

func pillColors(tone PillTone) (bg, border, fg, dot color.NRGBA) {
	switch tone {
	case PillAccent:
		return ColorAccentSoft, ColorAccentLine, ColorAccentPillFg, ColorAccentCyan
	case PillOk:
		return ColorOkSoft, ColorOkLine, ColorOkPillFg, ColorSuccess
	case PillWarn:
		return ColorWarnSoft, ColorWarnLine, ColorWarnPillFg, ColorWarning
	case PillDanger:
		return ColorDangerSoft, ColorDangerLine, ColorDangerPillFg, ColorDanger
	default: // PillMute
		return ColorCardBg, ColorLine2, ColorFg2, ColorFg3
	}
}

// StatusPill renders a rounded-full pill with a colored dot and mono label.
func StatusPill(label string, tone PillTone) fyne.CanvasObject {
	bgColor, borderColor, fgColor, dotColor := pillColors(tone)

	dot := canvas.NewCircle(dotColor)
	dotWrap := container.NewGridWrap(fyne.NewSize(6, 6), dot)

	txt := canvas.NewText(strings.ToUpper(label), fgColor)
	txt.TextSize = 10
	txt.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}

	content := container.NewHBox(
		container.NewCenter(dotWrap),
		txt,
	)

	bg := canvas.NewRectangle(bgColor)
	bg.CornerRadius = RadiusPill

	border := canvas.NewRectangle(borderColor)
	border.CornerRadius = RadiusPill
	border.StrokeWidth = 1
	border.StrokeColor = borderColor
	border.FillColor = color.Transparent

	padded := container.New(layout.NewCustomPaddedLayout(Space1, Space1, Space2, Space3), content)

	return container.NewStack(bg, border, padded)
}

// SectionEyebrow renders a mono 10px uppercase label in fg2.
func SectionEyebrow(label string) fyne.CanvasObject {
	txt := canvas.NewText(strings.ToUpper(label), ColorFg2)
	txt.TextSize = 10
	txt.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}
	return txt
}

// PageHeader creates the standard screen header pattern.
func PageHeader(eyebrow, title, subtitle string, rightAction fyne.CanvasObject) fyne.CanvasObject {
	var items []fyne.CanvasObject

	if eyebrow != "" {
		items = append(items, SectionEyebrow(eyebrow))
	}

	h1 := canvas.NewText(title, ColorTextPrimary)
	h1.TextSize = 22
	h1.TextStyle = fyne.TextStyle{Bold: true}
	items = append(items, h1)

	if subtitle != "" {
		sub := canvas.NewText(subtitle, ColorTextSecondary)
		sub.TextSize = 13
		items = append(items, sub)
	}

	left := container.NewVBox(items...)

	var header fyne.CanvasObject
	if rightAction != nil {
		header = container.NewBorder(nil, nil, left, rightAction)
	} else {
		header = left
	}

	divider := canvas.NewRectangle(ColorLine1)
	divider.SetMinSize(fyne.NewSize(0, 1))

	return container.NewVBox(
		container.New(layout.NewCustomPaddedLayout(0, Space4, 0, 0), header),
		divider,
	)
}

// SegmentedStrengthMeter renders a 5-segment horizontal bar.
// level 0 = empty, 1-5 = filled segments.
func SegmentedStrengthMeter(level int) fyne.CanvasObject {
	strengthColors := []color.NRGBA{ColorStr1, ColorStr2, ColorStr3, ColorStr4, ColorStr5}

	segments := make([]fyne.CanvasObject, 0, 9) // 5 segments + 4 spacers
	for i := 0; i < 5; i++ {
		seg := canvas.NewRectangle(ColorBg3)
		seg.CornerRadius = RadiusPill
		seg.SetMinSize(fyne.NewSize(0, 6))
		if i < level && level > 0 {
			idx := level - 1
			if idx >= len(strengthColors) {
				idx = len(strengthColors) - 1
			}
			seg.FillColor = strengthColors[idx]
		}
		if i > 0 {
			spacer := canvas.NewRectangle(color.Transparent)
			spacer.SetMinSize(fyne.NewSize(Space1, 6))
			segments = append(segments, spacer)
		}
		segments = append(segments, seg)
	}

	return container.NewHBox(segments...)
}

// SegmentedStrengthMeterFlex is like SegmentedStrengthMeter but stretches to fill width.
func SegmentedStrengthMeterFlex(level int) fyne.CanvasObject {
	strengthColors := []color.NRGBA{ColorStr1, ColorStr2, ColorStr3, ColorStr4, ColorStr5}

	segments := make([]fyne.CanvasObject, 5)
	for i := 0; i < 5; i++ {
		seg := canvas.NewRectangle(ColorBg3)
		seg.CornerRadius = RadiusPill
		seg.SetMinSize(fyne.NewSize(40, 6))
		if i < level && level > 0 {
			idx := level - 1
			if idx >= len(strengthColors) {
				idx = len(strengthColors) - 1
			}
			seg.FillColor = strengthColors[idx]
		}
		segments[i] = seg
	}

	return container.NewGridWithColumns(5, segments...)
}

// StrengthLabels returns the text label for a given strength level.
func StrengthLabel(level int) string {
	labels := []string{"", "Very weak", "Weak", "Fair", "Strong", "Excellent"}
	if level < 0 || level >= len(labels) {
		return ""
	}
	return labels[level]
}

// StrengthBlock renders the full strength display: meter + label + crack time + entropy.
func StrengthBlock(level int, crackTime string, entropyBits int) fyne.CanvasObject {
	meter := SegmentedStrengthMeterFlex(level)

	label := canvas.NewText(StrengthLabel(level), ColorTextSecondary)
	label.TextSize = 11
	label.TextStyle = fyne.TextStyle{Monospace: true}

	meterRow := container.NewBorder(nil, nil, nil, label, meter)

	var detail string
	if crackTime != "" {
		detail = "Crack time: " + crackTime
		if entropyBits > 0 {
			detail += "  |  Entropy: " + strings.TrimRight(strings.TrimRight(
				strings.Replace(string(rune(entropyBits+'0')), "", "", 0), "0"), ".")
		}
	}

	var items []fyne.CanvasObject
	items = append(items, meterRow)

	if detail != "" {
		detailTxt := canvas.NewText(detail, ColorFg2)
		detailTxt.TextSize = 11
		detailTxt.TextStyle = fyne.TextStyle{Monospace: true}
		items = append(items, detailTxt)
	}

	return container.NewVBox(items...)
}

// UnderlineTabs creates a horizontal underline-style tab row.
func UnderlineTabs(labels []string, activeIdx int, onSelect func(int)) fyne.CanvasObject {
	tabs := make([]fyne.CanvasObject, len(labels))
	for i, lbl := range labels {
		idx := i
		var txtColor color.Color
		if idx == activeIdx {
			txtColor = ColorTextPrimary
		} else {
			txtColor = ColorFg2
		}

		txt := canvas.NewText(lbl, txtColor)
		txt.TextSize = 13
		if idx == activeIdx {
			txt.TextStyle = fyne.TextStyle{Bold: true}
		}

		underline := canvas.NewRectangle(color.Transparent)
		underline.SetMinSize(fyne.NewSize(0, 2))
		if idx == activeIdx {
			underline.FillColor = ColorAccentCyan
		}

		overlay := NewClickOverlay(func() {
			if onSelect != nil {
				onSelect(idx)
			}
		})

		tab := container.NewStack(
			container.NewVBox(
				container.New(layout.NewCustomPaddedLayout(Space2, Space2, Space3, Space3), txt),
				underline,
			),
			overlay,
		)
		tabs[i] = tab
	}

	baseline := canvas.NewRectangle(ColorLine1)
	baseline.SetMinSize(fyne.NewSize(0, 1))

	return container.NewVBox(
		container.NewHBox(tabs...),
		baseline,
	)
}

// KVItem represents a key-value pair for the table.
type KVItem struct {
	Key    string
	Value  string
	Detail string
}

// KeyValueTable renders a 2-column key-value table.
func KeyValueTable(items []KVItem) fyne.CanvasObject {
	rows := make([]fyne.CanvasObject, 0, len(items)*2)
	for _, item := range items {
		key := canvas.NewText(strings.ToUpper(item.Key), ColorFg2)
		key.TextSize = 10
		key.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}

		val := canvas.NewText(item.Value, ColorTextPrimary)
		val.TextSize = 13

		valCol := container.NewVBox(val)
		if item.Detail != "" {
			det := canvas.NewText(item.Detail, ColorFg2)
			det.TextSize = 12
			det.TextStyle = fyne.TextStyle{Monospace: true}
			valCol.Add(det)
		}

		keyWrap := container.NewGridWrap(fyne.NewSize(180, 0), key)
		row := container.NewBorder(nil, nil, keyWrap, nil, valCol)
		rows = append(rows, container.New(layout.NewCustomPaddedLayout(Space2, Space2, 0, 0), row))
	}

	return container.NewVBox(rows...)
}

// ToggleSwitch is a custom toggle widget (38x22 track, 16x16 thumb).
type ToggleSwitch struct {
	widget.BaseWidget
	On       bool
	OnChange func(bool)
}

func NewToggleSwitch(on bool, onChange func(bool)) *ToggleSwitch {
	t := &ToggleSwitch{On: on, OnChange: onChange}
	t.ExtendBaseWidget(t)
	return t
}

func (t *ToggleSwitch) Tapped(_ *fyne.PointEvent) {
	t.On = !t.On
	if t.OnChange != nil {
		t.OnChange(t.On)
	}
	t.Refresh()
}

func (t *ToggleSwitch) MinSize() fyne.Size {
	return fyne.NewSize(38, 22)
}

func (t *ToggleSwitch) CreateRenderer() fyne.WidgetRenderer {
	track := canvas.NewRectangle(ColorSidebarBg)
	track.CornerRadius = RadiusPill

	trackBorder := canvas.NewRectangle(ColorLine2)
	trackBorder.CornerRadius = RadiusPill
	trackBorder.StrokeWidth = 1
	trackBorder.StrokeColor = ColorLine2
	trackBorder.FillColor = color.Transparent

	thumb := canvas.NewCircle(ColorFg3)

	return &toggleRenderer{
		toggle:      t,
		track:       track,
		trackBorder: trackBorder,
		thumb:       thumb,
		objects:     []fyne.CanvasObject{track, trackBorder, thumb},
	}
}

type toggleRenderer struct {
	toggle      *ToggleSwitch
	track       *canvas.Rectangle
	trackBorder *canvas.Rectangle
	thumb       *canvas.Circle
	objects     []fyne.CanvasObject
}

func (r *toggleRenderer) Layout(size fyne.Size) {
	r.track.Resize(size)
	r.track.Move(fyne.NewPos(0, 0))
	r.trackBorder.Resize(size)
	r.trackBorder.Move(fyne.NewPos(0, 0))

	thumbSize := fyne.NewSize(16, 16)
	r.thumb.Resize(thumbSize)
	yOff := (size.Height - 16) / 2

	if r.toggle.On {
		r.thumb.Move(fyne.NewPos(size.Width-16-3, yOff))
	} else {
		r.thumb.Move(fyne.NewPos(3, yOff))
	}
}

func (r *toggleRenderer) MinSize() fyne.Size {
	return fyne.NewSize(38, 22)
}

func (r *toggleRenderer) Refresh() {
	if r.toggle.On {
		r.track.FillColor = ColorAccentSoft
		r.trackBorder.StrokeColor = ColorAccentLine
		r.thumb.FillColor = ColorAccentCyan
		r.thumb.Move(fyne.NewPos(38-16-3, (22-16)/2))
	} else {
		r.track.FillColor = ColorSidebarBg
		r.trackBorder.StrokeColor = ColorLine2
		r.thumb.FillColor = ColorFg3
		r.thumb.Move(fyne.NewPos(3, (22-16)/2))
	}
	r.track.Refresh()
	r.trackBorder.Refresh()
	r.thumb.Refresh()
}

func (r *toggleRenderer) Destroy() {}

func (r *toggleRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

// Topbar creates the 48px app top bar.
func Topbar(breadcrumbs []string, pills []fyne.CanvasObject) fyne.CanvasObject {
	bg := canvas.NewRectangle(ColorSidebarBg)
	bg.SetMinSize(fyne.NewSize(0, TopbarHeight))

	bottomBorder := canvas.NewRectangle(ColorLine1)
	bottomBorder.SetMinSize(fyne.NewSize(0, 1))

	var crumbItems []fyne.CanvasObject
	for i, crumb := range breadcrumbs {
		if i > 0 {
			sep := canvas.NewText(" / ", ColorFg3)
			sep.TextSize = 11
			sep.TextStyle = fyne.TextStyle{Monospace: true}
			crumbItems = append(crumbItems, sep)
		}
		txt := canvas.NewText(crumb, ColorFg2)
		txt.TextSize = 11
		txt.TextStyle = fyne.TextStyle{Monospace: true}
		crumbItems = append(crumbItems, txt)
	}
	crumbBox := container.NewHBox(crumbItems...)

	pillBox := container.NewHBox(pills...)

	content := container.NewBorder(nil, nil,
		container.NewCenter(crumbBox),
		container.NewCenter(pillBox),
	)

	padded := container.New(layout.NewCustomPaddedLayout(0, 0, Space4, Space4), content)

	// Use NewStack without NewCenter so padded fills full width;
	// NewCenter would collapse it to its natural width, breaking left/right alignment.
	bar := container.NewStack(bg, padded)

	return container.NewVBox(
		container.NewGridWrap(fyne.NewSize(0, TopbarHeight), bar),
		bottomBorder,
	)
}

// BackgroundPattern selects the decorative background pattern.
type BackgroundPattern int

const (
	PatternPlain BackgroundPattern = iota
	PatternDots
	PatternGrid
	PatternHex
)

// BackgroundTexture renders a tiled decorative background.
func BackgroundTexture(pattern BackgroundPattern) fyne.CanvasObject {
	bg := canvas.NewRectangle(ColorBg)
	if pattern == PatternPlain {
		return bg
	}

	raster := canvas.NewRasterWithPixels(func(x, y, w, h int) color.Color {
		switch pattern {
		case PatternDots:
			return dotPattern(x, y)
		case PatternGrid:
			return gridPattern(x, y)
		case PatternHex:
			return hexPattern(x, y)
		default:
			return color.Transparent
		}
	})

	return container.NewStack(bg, raster)
}

func dotPattern(x, y int) color.Color {
	spacing := 22
	cx := x % spacing
	cy := y % spacing
	if cx*cx+cy*cy <= 1 {
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x0a}
	}
	return color.Transparent
}

func gridPattern(x, y int) color.Color {
	spacing := 32
	if x%spacing == 0 || y%spacing == 0 {
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x06}
	}
	return color.Transparent
}

func hexPattern(x, y int) color.Color {
	w := 56.0
	h := 64.0
	fx := math.Mod(float64(x), w)
	fy := math.Mod(float64(y), h)

	if isOnHexEdge(fx, fy, w, h) {
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x09}
	}
	return color.Transparent
}

func isOnHexEdge(fx, fy, w, h float64) bool {
	halfH := h / 2.0
	qW := w / 4.0

	for _, yOff := range []float64{0, halfH} {
		ly := fy - yOff
		if ly < 0 || ly > halfH {
			continue
		}
		if math.Abs(ly) < 1 && fx > qW && fx < w-qW {
			return true
		}
		if math.Abs(ly-halfH) < 1 && fx > qW && fx < w-qW {
			return true
		}
		if fx < qW {
			expected := halfH/2.0 + (fx/qW)*(halfH/2.0)
			if math.Abs(ly-expected) < 1.2 {
				return true
			}
			expected2 := halfH/2.0 - (fx/qW)*(halfH/2.0)
			if math.Abs(ly-expected2) < 1.2 {
				return true
			}
		}
		if fx > w-qW {
			rem := fx - (w - qW)
			expected := halfH/2.0 + (rem/qW)*(halfH/2.0)
			if math.Abs(ly-(halfH-expected+halfH/2.0)) < 1.2 {
				return true
			}
		}
	}
	return false
}

// CreatePrimaryButton creates a blue accent button with white text.
func CreatePrimaryButton(label string, onClick func()) fyne.CanvasObject {
	btn := NewClickOverlay(onClick)

	bg := canvas.NewRectangle(ColorAccentCyan)
	bg.CornerRadius = RadiusInput
	bg.SetMinSize(fyne.NewSize(160, 36))

	txt := canvas.NewText(label, color.White)
	txt.TextSize = 13
	txt.TextStyle = fyne.TextStyle{Bold: true}

	return container.NewStack(bg, btn,
		container.New(layout.NewCustomPaddedLayout(9, 9, 14, 14), container.NewCenter(txt)))
}

// CreatePrimaryButtonWithIcon creates a primary button with a leading icon.
func CreatePrimaryButtonWithIcon(label string, icon *fyne.StaticResource, onClick func()) fyne.CanvasObject {
	btn := NewClickOverlay(onClick)

	bg := canvas.NewRectangle(ColorAccentCyan)
	bg.CornerRadius = RadiusInput
	bg.SetMinSize(fyne.NewSize(0, 16))

	ico := canvas.NewImageFromResource(icon)
	ico.SetMinSize(fyne.NewSize(16, 16))

	txt := canvas.NewText(label, color.White)
	txt.TextSize = 13
	txt.TextStyle = fyne.TextStyle{Bold: true}

	content := container.NewHBox(ico, txt)

	return container.NewStack(bg, btn,
		container.New(layout.NewCustomPaddedLayout(9, 9, 14, 14), container.NewCenter(content)))
}

// CreateDefaultButton creates a bg3 button with line2 border.
func CreateDefaultButton(label string, onClick func()) fyne.CanvasObject {
	btn := NewClickOverlay(onClick)

	bg := canvas.NewRectangle(ColorBg3)
	bg.CornerRadius = RadiusInput
	bg.SetMinSize(fyne.NewSize(0, 36))

	border := canvas.NewRectangle(color.Transparent)
	border.CornerRadius = RadiusInput
	border.StrokeWidth = 1
	border.StrokeColor = ColorLine2
	border.FillColor = color.Transparent
	border.SetMinSize(fyne.NewSize(0, 36))

	txt := canvas.NewText(label, ColorTextPrimary)
	txt.TextSize = 13

	return container.NewStack(bg, border, btn,
		container.New(layout.NewCustomPaddedLayout(9, 9, 14, 14), container.NewCenter(txt)))
}

// CreateGhostButton creates a transparent button with fg1 text.
func CreateGhostButton(label string, onClick func()) fyne.CanvasObject {
	btn := NewClickOverlay(onClick)

	bg := canvas.NewRectangle(color.Transparent)
	bg.CornerRadius = RadiusInput
	bg.SetMinSize(fyne.NewSize(0, 36))

	txt := canvas.NewText(label, ColorTextSecondary)
	txt.TextSize = 13

	return container.NewStack(bg, btn,
		container.New(layout.NewCustomPaddedLayout(9, 9, 14, 14), container.NewCenter(txt)))
}

// CreateDangerButton creates a transparent button with danger border.
func CreateDangerButton(label string, onClick func()) fyne.CanvasObject {
	btn := NewClickOverlay(onClick)

	bg := canvas.NewRectangle(color.Transparent)
	bg.CornerRadius = RadiusInput
	bg.SetMinSize(fyne.NewSize(0, 36))

	border := canvas.NewRectangle(color.Transparent)
	border.CornerRadius = RadiusInput
	border.StrokeWidth = 1
	border.StrokeColor = ColorDangerLine
	border.FillColor = color.Transparent
	border.SetMinSize(fyne.NewSize(0, 36))

	txt := canvas.NewText(label, ColorDangerPillFg)
	txt.TextSize = 13

	return container.NewStack(bg, border, btn,
		container.New(layout.NewCustomPaddedLayout(9, 9, 14, 14), container.NewCenter(txt)))
}

// CreateSmallIconButton creates a 28x28 icon button (transparent bg, fg2 icon).
func CreateSmallIconButton(icon *fyne.StaticResource, onClick func()) fyne.CanvasObject {
	btn := NewClickOverlay(onClick)

	ico := canvas.NewImageFromResource(icon)
	ico.SetMinSize(fyne.NewSize(16, 16))

	wrapper := container.NewGridWrap(fyne.NewSize(28, 28), container.NewCenter(ico))

	return container.NewStack(wrapper, btn)
}

// NavItem represents a sidebar navigation item.
type NavItemWidget struct {
	widget.BaseWidget
	Icon     *fyne.StaticResource
	Label    string
	IsActive bool
	OnTap    func()
}

func NewNavItem(icon *fyne.StaticResource, label string, isActive bool, onTap func()) *NavItemWidget {
	n := &NavItemWidget{
		Icon:     icon,
		Label:    label,
		IsActive: isActive,
		OnTap:    onTap,
	}
	n.ExtendBaseWidget(n)
	return n
}

func (n *NavItemWidget) Tapped(_ *fyne.PointEvent) {
	if n.OnTap != nil {
		n.OnTap()
	}
}

func (n *NavItemWidget) MinSize() fyne.Size {
	return fyne.NewSize(SidebarWidth-Space4*2, 36)
}

func (n *NavItemWidget) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(color.Transparent)
	bg.CornerRadius = RadiusInput

	accentBar := canvas.NewRectangle(color.Transparent)

	ico := canvas.NewImageFromResource(n.Icon)
	ico.SetMinSize(fyne.NewSize(18, 18))

	label := canvas.NewText(n.Label, ColorTextSecondary)
	label.TextSize = 13

	border := canvas.NewRectangle(color.Transparent)
	border.CornerRadius = RadiusInput
	border.StrokeWidth = 1
	border.FillColor = color.Transparent

	return &navItemRenderer{
		item:      n,
		bg:        bg,
		accentBar: accentBar,
		ico:       ico,
		label:     label,
		border:    border,
		objects:   []fyne.CanvasObject{bg, border, accentBar, ico, label},
	}
}

type navItemRenderer struct {
	item      *NavItemWidget
	bg        *canvas.Rectangle
	accentBar *canvas.Rectangle
	ico       *canvas.Image
	label     *canvas.Text
	border    *canvas.Rectangle
	objects   []fyne.CanvasObject
}

func (r *navItemRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)
	r.border.Resize(size)

	r.accentBar.Resize(fyne.NewSize(2, size.Height-16))
	r.accentBar.Move(fyne.NewPos(0, 8))

	r.ico.Resize(fyne.NewSize(18, 18))
	r.ico.Move(fyne.NewPos(12, (size.Height-18)/2))

	r.label.Move(fyne.NewPos(38, (size.Height-r.label.MinSize().Height)/2))
}

func (r *navItemRenderer) MinSize() fyne.Size {
	return r.item.MinSize()
}

func (r *navItemRenderer) Refresh() {
	if r.item.IsActive {
		r.bg.FillColor = ColorBg3
		r.border.StrokeColor = ColorLine2
		r.accentBar.FillColor = ColorAccentCyan
		r.label.Color = ColorTextPrimary
	} else {
		r.bg.FillColor = color.Transparent
		r.border.StrokeColor = color.Transparent
		r.accentBar.FillColor = color.Transparent
		r.label.Color = ColorTextSecondary
	}
	r.bg.Refresh()
	r.border.Refresh()
	r.accentBar.Refresh()
	r.label.Refresh()
	r.ico.Refresh()
}

func (r *navItemRenderer) Destroy() {}

func (r *navItemRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

// WarningBanner renders an amber warning box with icon and text.
func WarningBanner(title, body string) fyne.CanvasObject {
	bg := canvas.NewRectangle(ColorWarnSoft)
	bg.CornerRadius = Space2

	border := canvas.NewRectangle(color.Transparent)
	border.CornerRadius = Space2
	border.StrokeWidth = 1
	border.StrokeColor = ColorWarnLine
	border.FillColor = color.Transparent

	ico := canvas.NewImageFromResource(IconAlertTriangle)
	ico.SetMinSize(fyne.NewSize(18, 18))

	titleTxt := canvas.NewText(title, ColorWarnPillFg)
	titleTxt.TextSize = 12
	titleTxt.TextStyle = fyne.TextStyle{Bold: true}

	bodyTxt := canvas.NewText(body, ColorTextSecondary)
	bodyTxt.TextSize = 12

	textBlock := container.NewVBox(titleTxt, bodyTxt)

	content := container.NewBorder(nil, nil, container.NewCenter(ico), nil, textBlock)

	padded := container.New(layout.NewCustomPaddedLayout(Space3, Space3, Space4, Space4), content)

	return container.NewStack(bg, border, padded)
}

// CardWithHeader creates a card with optional eyebrow/title header and right action.
func CardWithHeader(eyebrow, title string, rightAction, body fyne.CanvasObject) fyne.CanvasObject {
	bg := canvas.NewRectangle(ColorCardBg)
	bg.CornerRadius = RadiusCard

	border := canvas.NewRectangle(color.Transparent)
	border.CornerRadius = RadiusCard
	border.StrokeWidth = 1
	border.StrokeColor = ColorLine1
	border.FillColor = color.Transparent

	var parts []fyne.CanvasObject

	if eyebrow != "" || title != "" || rightAction != nil {
		var headerLeft []fyne.CanvasObject
		if eyebrow != "" {
			headerLeft = append(headerLeft, SectionEyebrow(eyebrow))
		}
		if title != "" {
			t := canvas.NewText(title, ColorTextPrimary)
			t.TextSize = 13
			t.TextStyle = fyne.TextStyle{Bold: true}
			headerLeft = append(headerLeft, t)
		}

		left := container.NewVBox(headerLeft...)
		var hdr fyne.CanvasObject
		if rightAction != nil {
			hdr = container.NewBorder(nil, nil, left, rightAction)
		} else {
			hdr = left
		}

		headerDivider := canvas.NewRectangle(ColorLine1)
		headerDivider.SetMinSize(fyne.NewSize(0, 1))

		parts = append(parts,
			container.New(layout.NewCustomPaddedLayout(Space4, Space3, Space5, Space5), hdr),
			headerDivider,
		)
	}

	if body != nil {
		parts = append(parts,
			container.New(layout.NewCustomPaddedLayout(Space5, Space5, Space5, Space5), body),
		)
	}

	content := container.NewVBox(parts...)

	return container.NewStack(bg, border, content)
}

// CollapsibleCardWithHeader is like CardWithHeader but the body can be toggled open/closed.
// Starts collapsed by default.
func CollapsibleCardWithHeader(eyebrow, title string, rightAction, body fyne.CanvasObject) fyne.CanvasObject {
	collapsed := true

	bg := canvas.NewRectangle(ColorCardBg)
	bg.CornerRadius = RadiusCard

	border := canvas.NewRectangle(color.Transparent)
	border.CornerRadius = RadiusCard
	border.StrokeWidth = 1
	border.StrokeColor = ColorLine1
	border.FillColor = color.Transparent

	chevron := canvas.NewText("▶", ColorFg2)
	chevron.TextSize = 10
	chevron.TextStyle = fyne.TextStyle{Monospace: true}

	var headerLeft []fyne.CanvasObject
	if eyebrow != "" {
		headerLeft = append(headerLeft, SectionEyebrow(eyebrow))
	}
	if title != "" {
		t := canvas.NewText(title, ColorTextPrimary)
		t.TextSize = 13
		t.TextStyle = fyne.TextStyle{Bold: true}
		headerLeft = append(headerLeft, t)
	}

	left := container.NewHBox(chevron, container.NewVBox(headerLeft...))

	var hdrRight fyne.CanvasObject
	if rightAction != nil {
		hdrRight = rightAction
	}

	var hdr fyne.CanvasObject
	if hdrRight != nil {
		hdr = container.NewBorder(nil, nil, left, hdrRight)
	} else {
		hdr = left
	}

	headerDivider := canvas.NewRectangle(ColorLine1)
	headerDivider.SetMinSize(fyne.NewSize(0, 1))

	bodyPadded := container.New(layout.NewCustomPaddedLayout(Space5, Space5, Space5, Space5), body)
	bodyPadded.Hide()

	headerPadded := container.New(layout.NewCustomPaddedLayout(Space4, Space3, Space5, Space5), hdr)

	content := container.NewVBox(headerPadded, headerDivider, bodyPadded)
	card := container.NewStack(bg, border, content)

	headerPadded.Objects = []fyne.CanvasObject{
		widget.NewButton("", func() {
			collapsed = !collapsed
			if collapsed {
				chevron.Text = "▶"
				bodyPadded.Hide()
			} else {
				chevron.Text = "▼"
				bodyPadded.Show()
			}
			chevron.Refresh()
			content.Refresh()
		}),
	}

	// Rebuild: transparent tap target over the full header row
	tapTarget := widget.NewButton("", func() {
		collapsed = !collapsed
		if collapsed {
			chevron.Text = "▶"
			bodyPadded.Hide()
		} else {
			chevron.Text = "▼"
			bodyPadded.Show()
		}
		chevron.Refresh()
		content.Refresh()
	})
	tapTarget.Importance = widget.LowImportance

	headerRow := container.NewStack(tapTarget, headerPadded)
	content.Objects = []fyne.CanvasObject{headerRow, headerDivider, bodyPadded}
	content.Refresh()

	return card
}

// VaultAvatar renders a small rounded square with the first letter of name centered.
func VaultAvatar(name string) fyne.CanvasObject {
	bg := canvas.NewRectangle(ColorAccentSoft)
	bg.CornerRadius = RadiusInput
	bg.SetMinSize(fyne.NewSize(32, 32))

	letter := ""
	if len(name) > 0 {
		letter = strings.ToUpper(string([]rune(name)[:1]))
	}

	lbl := canvas.NewText(letter, ColorTextPrimary)
	lbl.TextSize = 14
	lbl.TextStyle = fyne.TextStyle{Bold: true}

	return container.NewGridWrap(fyne.NewSize(32, 32), container.NewStack(bg, container.NewCenter(lbl)))
}

// FormFooter creates the footer bar for form cards.
func FormFooter(leftText string, buttons ...fyne.CanvasObject) fyne.CanvasObject {
	bg := canvas.NewRectangle(ColorSidebarBg)
	bg.SetMinSize(fyne.NewSize(0, 48))

	topBorder := canvas.NewRectangle(ColorLine1)
	topBorder.SetMinSize(fyne.NewSize(0, 1))

	leftLabel := canvas.NewText(leftText, ColorFg2)
	leftLabel.TextSize = 11
	leftLabel.TextStyle = fyne.TextStyle{Monospace: true}

	rightButtons := container.NewHBox(buttons...)

	content := container.NewBorder(nil, nil, container.NewCenter(leftLabel), rightButtons)
	padded := container.New(layout.NewCustomPaddedLayout(Space3, Space3, Space5, Space5), content)

	return container.NewVBox(topBorder, container.NewStack(bg, padded))
}

// FieldLabel creates an eyebrow-style label for form fields, with optional right element.
func FieldLabel(label string, right fyne.CanvasObject) fyne.CanvasObject {
	eyebrow := SectionEyebrow(label)
	if right != nil {
		return container.NewBorder(nil, nil, eyebrow, right)
	}
	return eyebrow
}

// KindBadge renders a small bordered badge (e.g. "Password", "Card", "Note").
func KindBadge(label string) fyne.CanvasObject {
	txt := canvas.NewText(strings.ToUpper(label), ColorFg2)
	txt.TextSize = 10
	txt.TextStyle = fyne.TextStyle{Monospace: true}

	border := canvas.NewRectangle(color.Transparent)
	border.CornerRadius = RadiusSmall
	border.StrokeWidth = 1
	border.StrokeColor = ColorLine2
	border.FillColor = color.Transparent

	return container.NewStack(border, container.New(layout.NewCustomPaddedLayout(2, 2, 6, 6), txt))
}

// TypeIcon creates a 36x36 icon with a tinted background for item type.
func TypeIcon(icon *fyne.StaticResource, tintColor color.NRGBA) fyne.CanvasObject {
	softBg := color.NRGBA{R: tintColor.R, G: tintColor.G, B: tintColor.B, A: 0x24}

	bg := canvas.NewRectangle(softBg)
	bg.CornerRadius = RadiusInput
	bg.SetMinSize(fyne.NewSize(36, 36))

	border := canvas.NewRectangle(color.Transparent)
	border.CornerRadius = RadiusInput
	border.StrokeWidth = 1
	border.StrokeColor = color.NRGBA{R: tintColor.R, G: tintColor.G, B: tintColor.B, A: 0x66}
	border.FillColor = color.Transparent
	border.SetMinSize(fyne.NewSize(36, 36))

	ico := canvas.NewImageFromResource(icon)
	ico.SetMinSize(fyne.NewSize(18, 18))

	return container.NewStack(bg, border, container.NewCenter(ico))
}

// MonoText creates a monospace text at given size and color.
func MonoText(text string, size float32, c color.Color) fyne.CanvasObject {
	txt := canvas.NewText(text, c)
	txt.TextSize = size
	txt.TextStyle = fyne.TextStyle{Monospace: true}
	return txt
}

// ensure image package is available for raster functions
var _ = image.Pt
