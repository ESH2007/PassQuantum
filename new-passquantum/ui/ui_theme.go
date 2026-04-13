package main

import (
	"image/color"
	"math"
	"math/rand"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// EXACT COLOR PALETTE FROM REFERENCE IMAGES
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
var (
	// Background colors - Dark theme
	ColorBg        = color.NRGBA{R: 11, G: 15, B: 20, A: 255} // #0b0f14 - Main background
	ColorSidebarBg = color.NRGBA{R: 20, G: 25, B: 32, A: 255} // #141920 - Sidebar background
	ColorCardBg    = color.NRGBA{R: 26, G: 31, B: 40, A: 255} // #1a1f28 - Card background
	ColorInputBg   = color.NRGBA{R: 30, G: 40, B: 50, A: 255} // #1e2832 - Input fields

	// Accent colors - Cyan and Magenta
	ColorAccentCyan = color.NRGBA{R: 34, G: 211, B: 238, A: 255} // #22d3ee - Primary cyan
	ColorAccentPink = color.NRGBA{R: 236, G: 72, B: 153, A: 255} // #ec4899 - Magenta/Pink
	ColorPurple     = color.NRGBA{R: 168, G: 85, B: 247, A: 200} // #a855f7 - Purple accent

	// Action roles
	ColorPrimaryButton   = ColorAccentCyan
	ColorSecondaryButton = ColorAccentPink

	// Text colors
	ColorTextPrimary   = color.NRGBA{R: 255, G: 255, B: 255, A: 255} // White
	ColorTextSecondary = color.NRGBA{R: 148, G: 163, B: 184, A: 255} // #94a3b8 - Gray

	// Border and glow
	ColorBorderCyan = color.NRGBA{R: 34, G: 211, B: 238, A: 180} // Cyan with alpha
	ColorGlowCyan   = color.NRGBA{R: 34, G: 211, B: 238, A: 80}  // Subtle cyan glow

	// Status colors
	ColorDanger  = color.NRGBA{R: 239, G: 68, B: 68, A: 220}  // #ef4444 - Red
	ColorWarning = color.NRGBA{R: 250, G: 204, B: 21, A: 220} // #facc15 - Yellow
	ColorSuccess = color.NRGBA{R: 34, G: 197, B: 94, A: 255}  // #22c55e - Green

	// Backward compatibility aliases
	ColorAccentCyn = ColorAccentCyan
	ColorTextSec   = ColorTextSecondary
	ColorTextPrim  = ColorTextPrimary
	ColorBorder    = ColorBorderCyan
	ColorMagenta   = ColorAccentPink
)

// Sizing constants
const (
	BorderWidth  = 1
	BorderRadius = 8
	SidebarWidth = 175
	PaddingLarge = 24
	PaddingMed   = 16
	PaddingSmall = 8
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// PARTICLE BACKGROUND ANIMATION
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

type Particle struct {
	X, Y      float32
	VX, VY    float32
	Size      float32
	Opacity   uint8
	Phase     float32 // For pulsing effect
	PulseRate float32 // Individual pulse speed
}

type ParticleBackground struct {
	widget.BaseWidget
	particles      []Particle
	canvas         fyne.Canvas
	width, height  float32
	animationPhase float64
}

func NewParticleBackground(particleCount int) *ParticleBackground {
	pb := &ParticleBackground{
		particles: make([]Particle, particleCount),
		width:     1920,
		height:    1080,
	}

	// Initialize particles with varied properties
	for i := range pb.particles {
		// Random position
		x := rand.Float32() * pb.width
		y := rand.Float32() * pb.height

		// Varied speeds (slower, more organic movement)
		speedFactor := 0.1 + rand.Float32()*0.3
		vx := (rand.Float32() - 0.5) * speedFactor
		vy := (rand.Float32() - 0.5) * speedFactor

		// Varied sizes (1-3 pixels)
		size := 0.8 + rand.Float32()*2.2

		// Varied opacity (15-80 alpha)
		opacity := uint8(15 + rand.Intn(65))

		// Random phase for pulsing
		phase := rand.Float32() * 6.28 // 0 to 2π
		pulseRate := 0.02 + rand.Float32()*0.05

		pb.particles[i] = Particle{
			X:         x,
			Y:         y,
			VX:        vx,
			VY:        vy,
			Size:      size,
			Opacity:   opacity,
			Phase:     phase,
			PulseRate: pulseRate,
		}
	}

	// Start animation
	go pb.animate()

	pb.ExtendBaseWidget(pb)
	return pb
}

func (pb *ParticleBackground) animate() {
	ticker := time.NewTicker(time.Second / 60) // 60 FPS for smoother animation
	defer ticker.Stop()

	for range ticker.C {
		pb.animationPhase += 0.01

		for i := range pb.particles {
			p := &pb.particles[i]

			// Update position
			p.X += p.VX
			p.Y += p.VY

			// Update pulse phase
			p.Phase += p.PulseRate
			if p.Phase > 6.28 {
				p.Phase -= 6.28
			}

			// Wrap around screen with smooth transitions
			if p.X < -10 {
				p.X = pb.width + 10
			}
			if p.X > pb.width+10 {
				p.X = -10
			}
			if p.Y < -10 {
				p.Y = pb.height + 10
			}
			if p.Y > pb.height+10 {
				p.Y = -10
			}
		}
		fyne.Do(func() {
			pb.Refresh()
		})
	}
}

func (pb *ParticleBackground) CreateRenderer() fyne.WidgetRenderer {
	return &particleRenderer{particles: pb.particles}
}

type particleRenderer struct {
	particles []Particle
}

func (r *particleRenderer) Layout(size fyne.Size) {}
func (r *particleRenderer) MinSize() fyne.Size    { return fyne.NewSize(800, 600) }
func (r *particleRenderer) Refresh()              {}
func (r *particleRenderer) Destroy()              {}
func (r *particleRenderer) Objects() []fyne.CanvasObject {
	objects := make([]fyne.CanvasObject, len(r.particles))
	for i, p := range r.particles {
		// Calculate pulsing opacity
		pulse := float32(math.Sin(float64(p.Phase)))
		opacityDelta := uint8(pulse * 15) // ±15 opacity variation
		currentOpacity := p.Opacity
		if pulse > 0 {
			if int(p.Opacity)+int(opacityDelta) <= 255 {
				currentOpacity = p.Opacity + opacityDelta
			}
		} else {
			if int(p.Opacity)-int(opacityDelta) >= 10 {
				currentOpacity = p.Opacity - uint8(-opacityDelta)
			}
		}

		// Alternate colors for variety
		var particleColor color.NRGBA
		if i%3 == 0 {
			// Cyan particles
			particleColor = color.NRGBA{R: 34, G: 211, B: 238, A: currentOpacity}
		} else if i%3 == 1 {
			// Purple particles
			particleColor = color.NRGBA{R: 168, G: 85, B: 247, A: currentOpacity}
		} else {
			// Pink particles
			particleColor = color.NRGBA{R: 236, G: 72, B: 153, A: currentOpacity}
		}

		circle := canvas.NewCircle(particleColor)
		circle.Resize(fyne.NewSize(p.Size, p.Size))
		circle.Move(fyne.NewPos(p.X, p.Y))
		objects[i] = circle
	}
	return objects
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// BUTTON WITH GLOW EFFECT
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

type GlowButton struct {
	widget.Button
	glowIntensity float32
	glowColor     color.Color
}

type clickOverlay struct {
	widget.BaseWidget
	onTap func()
}

func newClickOverlay(onTap func()) *clickOverlay {
	o := &clickOverlay{onTap: onTap}
	o.ExtendBaseWidget(o)
	return o
}

func (o *clickOverlay) Tapped(_ *fyne.PointEvent) {
	if o.onTap != nil {
		o.onTap()
	}
}

func (o *clickOverlay) TappedSecondary(_ *fyne.PointEvent) {}

func (o *clickOverlay) CreateRenderer() fyne.WidgetRenderer {
	hitBox := canvas.NewRectangle(color.Transparent)
	return widget.NewSimpleRenderer(hitBox)
}

func NewGlowButton(label string, tapped func(), accentColor color.Color) *GlowButton {
	btn := &GlowButton{
		glowIntensity: 0.3,
		glowColor:     accentColor,
	}
	btn.Text = label
	btn.OnTapped = tapped
	btn.Importance = widget.HighImportance
	btn.ExtendBaseWidget(btn)
	return btn
}

// CreateNeonButton creates a cyberpunk-styled button with glow
func CreateNeonButton(label string, onClick func(), width, height float32) fyne.CanvasObject {
	return createColorRoleButton(label, onClick, width, height, ColorPrimaryButton)
}

// CreateSecondaryButton creates a secondary action button styled with the secondary role color.
func CreateSecondaryButton(label string, onClick func(), width, height float32) fyne.CanvasObject {
	return createColorRoleButton(label, onClick, width, height, ColorSecondaryButton)
}

func createColorRoleButton(label string, onClick func(), width, height float32, roleColor color.NRGBA) fyne.CanvasObject {
	btn := newClickOverlay(onClick)
	btn.Resize(fyne.NewSize(width, height))

	bg := canvas.NewRectangle(roleColor)
	bg.CornerRadius = 6
	bg.SetMinSize(fyne.NewSize(width, height))

	border := canvas.NewRectangle(color.NRGBA{R: roleColor.R, G: roleColor.G, B: roleColor.B, A: 210})
	border.CornerRadius = 6
	border.SetMinSize(fyne.NewSize(width+2, height+2))

	textColor := readableTextColor(roleColor)
	text := CreateLabel(label, 10, textColor, true)

	return container.NewStack(
		container.NewCenter(border),
		container.NewCenter(bg),
		btn,
		container.NewCenter(text),
	)
}

func readableTextColor(bg color.NRGBA) color.Color {
	return pickAdaptiveTextColor(bg)
}

func pickAdaptiveTextColor(bg color.NRGBA) color.NRGBA {
	light := color.NRGBA{R: 245, G: 248, B: 252, A: 255}
	dark := color.NRGBA{R: 20, G: 24, B: 30, A: 255}

	lightContrast := wcagContrastRatio(light, bg)
	darkContrast := wcagContrastRatio(dark, bg)

	if lightContrast >= darkContrast {
		return light
	}
	return dark
}

func wcagContrastRatio(fg color.NRGBA, bg color.NRGBA) float64 {
	fgL := wcagRelativeLuminance(fg)
	bgL := wcagRelativeLuminance(bg)

	lighter := math.Max(fgL, bgL)
	darker := math.Min(fgL, bgL)

	return (lighter + 0.05) / (darker + 0.05)
}

func wcagRelativeLuminance(c color.NRGBA) float64 {
	r := wcagLinearizedChannel(c.R)
	g := wcagLinearizedChannel(c.G)
	b := wcagLinearizedChannel(c.B)
	return 0.2126*r + 0.7152*g + 0.0722*b
}

func wcagLinearizedChannel(v uint8) float64 {
	srgb := float64(v) / 255.0
	if srgb <= 0.03928 {
		return srgb / 12.92
	}
	return math.Pow((srgb+0.055)/1.055, 2.4)
}

// CreateOutlinedButton creates a small outlined button (VER, EDIT, COPY, DEL style)
func CreateOutlinedButton(label string, onClick func(), btnColor color.Color) fyne.CanvasObject {
	btn := widget.NewButton(label, onClick)
	btn.Importance = widget.LowImportance

	// Create border
	border := canvas.NewRectangle(btnColor)
	border.StrokeWidth = 1
	border.StrokeColor = btnColor
	border.FillColor = color.Transparent
	border.CornerRadius = 4
	border.SetMinSize(fyne.NewSize(50, 28))

	return container.NewStack(border, btn)
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// UI COMPONENTS FROM REFERENCE IMAGES
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// CreateCard creates a card matching the reference design
func CreateCard(content fyne.CanvasObject, width, height float32, hasBorder bool) fyne.CanvasObject {
	bg := canvas.NewRectangle(ColorCardBg)
	bg.SetMinSize(fyne.NewSize(width, height))
	bg.CornerRadius = BorderRadius

	if hasBorder {
		border := canvas.NewRectangle(ColorBorderCyan)
		border.SetMinSize(fyne.NewSize(width, height))
		border.CornerRadius = BorderRadius
		return container.NewMax(border, bg, container.NewPadded(content))
	}

	return container.NewMax(bg, container.NewPadded(content))
}

// CreateCardWithBorderColor creates a card with specific border color
func CreateCardWithBorderColor(content fyne.CanvasObject, width, height float32, borderColor color.Color) fyne.CanvasObject {
	bg := canvas.NewRectangle(ColorCardBg)
	bg.SetMinSize(fyne.NewSize(width, height))
	bg.CornerRadius = BorderRadius

	border := canvas.NewRectangle(borderColor)
	border.SetMinSize(fyne.NewSize(width, height))
	border.CornerRadius = BorderRadius

	return container.NewMax(border, bg, container.NewPadded(content))
}

// CreateLabel creates a styled text label
func CreateLabel(text string, size float32, textColor color.Color, bold bool) fyne.CanvasObject {
	txt := canvas.NewText(text, textColor)
	txt.TextSize = size
	txt.TextStyle = fyne.TextStyle{Bold: bold}
	return txt
}

// CreateDivider creates a horizontal divider line
func CreateDivider() fyne.CanvasObject {
	line := canvas.NewRectangle(ColorBorderCyan)
	line.SetMinSize(fyne.NewSize(0, 1))
	return line
}

// CreateBackgroundContainer wraps content with animated background
func CreateBackgroundContainer(content fyne.CanvasObject) fyne.CanvasObject {
	bg := canvas.NewRectangle(ColorBg)
	bg.SetMinSize(fyne.NewSize(1200, 800))

	// Add particle background
	particles := NewParticleBackground(50)

	return container.NewStack(bg, particles, content)
}

// CreateSearchBar creates a search input matching reference design
func CreateSearchBar(placeholder string) *widget.Entry {
	search := widget.NewEntry()
	search.SetPlaceHolder(placeholder)
	return search
}

// CreateSidebarButton creates a sidebar menu button
func CreateSidebarButton(label string, onClick func(), isActive bool) fyne.CanvasObject {
	var textColor color.Color
	if isActive {
		textColor = ColorAccentCyan
	} else {
		textColor = ColorTextSecondary
	}

	text := CreateLabel(label, 11, textColor, false)
	btn := widget.NewButton("", onClick)
	btn.Importance = widget.LowImportance

	// Add subtle background for active state
	var bg fyne.CanvasObject
	if isActive {
		bgRect := canvas.NewRectangle(color.NRGBA{R: 34, G: 211, B: 238, A: 20})
		bgRect.CornerRadius = 4
		bg = bgRect
	} else {
		bg = canvas.NewRectangle(color.Transparent)
	}

	content := container.NewPadded(text)
	return container.NewStack(bg, btn, content)
}

// CreateNavButton creates a navigation button for the sidebar with icon
func CreateNavButton(icon, label string, onClick func(), isActive bool) fyne.CanvasObject {
	var textColor color.Color
	var bgColor color.Color
	var iconColor color.Color

	if isActive {
		textColor = ColorTextPrimary
		iconColor = ColorAccentCyan
		bgColor = color.NRGBA{R: 34, G: 211, B: 238, A: 30}
	} else {
		textColor = ColorTextSecondary
		iconColor = ColorTextSecondary
		bgColor = color.Transparent
	}

	iconLabel := CreateLabel(icon, 14, iconColor, false)
	textLabel := CreateLabel(label, 11, textColor, false)

	content := container.NewHBox(
		iconLabel,
		widget.NewLabel("  "), // Spacing
		textLabel,
	)

	btn := widget.NewButton("", onClick)
	btn.Importance = widget.LowImportance

	bg := canvas.NewRectangle(bgColor)
	bg.CornerRadius = 6

	padded := container.NewPadded(content)

	return container.NewStack(bg, btn, padded)
}

// NavigationItem represents a navigation menu item
type NavigationItem struct {
	Icon     string
	Label    string
	OnClick  func()
	IsActive bool
}

// CreateNavigationSidebar creates the left sidebar with navigation
func CreateNavigationSidebar(items []NavigationItem, lockItem NavigationItem, width float32) fyne.CanvasObject {
	var navButtons []fyne.CanvasObject

	// Add regular navigation items
	for _, item := range items {
		btn := CreateNavButton(item.Icon, item.Label, item.OnClick, item.IsActive)
		navButtons = append(navButtons, btn)
		navButtons = append(navButtons, widget.NewLabel("")) // Spacing
	}

	navContainer := container.NewVBox(navButtons...)

	// Create lock button at the bottom
	lockBtn := CreateNavButton(lockItem.Icon, lockItem.Label, lockItem.OnClick, false)

	// App title at the top
	titleLabel := CreateLabel("PassQuantum", 14, ColorAccentCyan, true)
	subtitleLabel := CreateLabel("Post-Quantum Safe", 9, ColorTextSecondary, false)
	header := container.NewVBox(
		titleLabel,
		subtitleLabel,
		widget.NewLabel(""),
		CreateDivider(),
		widget.NewLabel(""),
	)

	// Build sidebar with header, navigation, and lock at bottom
	bg := canvas.NewRectangle(ColorSidebarBg)

	mainContent := container.NewBorder(
		header,
		container.NewVBox(
			widget.NewLabel(""),
			CreateDivider(),
			widget.NewLabel(""),
			lockBtn,
		),
		nil,
		nil,
		container.NewVBox(navContainer),
	)

	sizedBg := container.NewStack(bg)
	sizedBg.Resize(fyne.NewSize(width, 0))

	paddedContent := container.NewPadded(mainContent)
	sidebar := container.NewStack(sizedBg, paddedContent)

	// Set fixed width
	sidebar.Resize(fyne.NewSize(width, 0))

	return sidebar
}

// CreatePasswordItem creates a password list item matching reference
func CreatePasswordItem(service, username, password string, onView, onEdit, onCopy, onDelete func()) fyne.CanvasObject {
	// Service name with icon
	serviceIcon := canvas.NewCircle(ColorAccentCyan)
	serviceIcon.Resize(fyne.NewSize(8, 8))

	serviceLabel := CreateLabel(service, 12, ColorTextPrimary, true)
	usernameLabel := CreateLabel(username, 10, ColorTextSecondary, false)
	passwordLabel := CreateLabel("••••••••••", 10, ColorTextSecondary, false)

	// Action buttons - small outlined style
	verBtn := CreateOutlinedButton("VIEW", onView, ColorAccentCyan)
	editBtn := CreateOutlinedButton("EDIT", onEdit, ColorAccentCyan)
	copyBtn := CreateOutlinedButton("COPY", onCopy, ColorAccentCyan)
	delBtn := CreateOutlinedButton("DELETE", onDelete, ColorDanger)

	buttons := container.NewHBox(verBtn, editBtn, copyBtn, delBtn)

	leftSide := container.NewVBox(
		container.NewHBox(serviceIcon, serviceLabel),
		usernameLabel,
		passwordLabel,
	)

	content := container.NewBorder(nil, nil, leftSide, buttons)

	// Wrap in card with border
	bg := canvas.NewRectangle(ColorCardBg)
	bg.CornerRadius = BorderRadius
	border := canvas.NewRectangle(ColorBorderCyan)
	border.CornerRadius = BorderRadius
	border.StrokeWidth = 1

	return container.NewStack(bg, container.NewPadded(content))
}

// CreateCheckItem creates a checkbox item for validation results
func CreateCheckItem(label string, checked bool) fyne.CanvasObject {
	var icon string
	var iconColor color.Color

	if checked {
		icon = "✓"
		iconColor = ColorSuccess
	} else {
		icon = "✕"
		iconColor = ColorDanger
	}

	iconLabel := CreateLabel(icon, 12, iconColor, true)
	textLabel := CreateLabel(label, 10, ColorTextSecondary, false)

	return container.NewHBox(iconLabel, textLabel)
}

// CreateSlider creates a custom styled slider
func CreateSlider(min, max, value float64) *widget.Slider {
	slider := widget.NewSlider(min, max)
	slider.Value = value
	return slider
}

// CreateCheckbox creates a styled checkbox
func CreateCheckbox(label string, checked bool, onChange func(bool)) *widget.Check {
	check := widget.NewCheck(label, onChange)
	check.SetChecked(checked)
	return check
}

// CreateTabButton creates a tab button for settings
func CreateTabButton(label string, isActive bool, onClick func()) fyne.CanvasObject {
	var bgColor color.NRGBA
	var textColor color.Color
	var borderColor color.Color

	if isActive {
		bgColor = ColorCardBg
		textColor = readableTextColor(bgColor)
		borderColor = ColorPrimaryButton
	} else {
		bgColor = color.NRGBA{R: ColorBg.R, G: ColorBg.G, B: ColorBg.B, A: 50}
		textColor = readableTextColor(bgColor)
		borderColor = color.Transparent
	}

	text := CreateLabel(label, 11, textColor, false)
	btn := newClickOverlay(onClick)

	bg := canvas.NewRectangle(bgColor)
	bg.CornerRadius = 4

	border := canvas.NewRectangle(borderColor)
	border.CornerRadius = 4

	content := container.NewCenter(text)
	return container.NewStack(border, bg, btn, content)
}

// Helper function for creating glow effect (shadow simulation)
func createGlowLayer(baseColor color.Color, radius float32) fyne.CanvasObject {
	glowColor := color.NRGBA{
		R: baseColor.(color.NRGBA).R,
		G: baseColor.(color.NRGBA).G,
		B: baseColor.(color.NRGBA).B,
		A: 40, // Low opacity for glow
	}

	glow := canvas.NewRectangle(glowColor)
	glow.CornerRadius = radius + 2
	return glow
}

// Animate opacity for glow effect
func animateGlow(obj fyne.CanvasObject, duration time.Duration) {
	ticker := time.NewTicker(duration / 60)
	defer ticker.Stop()

	phase := float64(0)
	for range ticker.C {
		phase += 0.1
		opacity := uint8(40 + 20*math.Sin(phase))

		if rect, ok := obj.(*canvas.Rectangle); ok {
			fyne.Do(func() {
				rect.FillColor = color.NRGBA{
					R: ColorAccentCyan.R,
					G: ColorAccentCyan.G,
					B: ColorAccentCyan.B,
					A: opacity,
				}
				rect.Refresh()
			})
		}
	}
}
