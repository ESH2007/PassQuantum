package main

import (
	"image/color"
	"math"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// ENHANCED UI COMPONENTS
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// CreateEnhancedCard creates a card with animated border glow
func CreateEnhancedCard(content fyne.CanvasObject, width, height float32) fyne.CanvasObject {
	// Background
	bg := canvas.NewRectangle(ColorCardBg)
	bg.SetMinSize(fyne.NewSize(width, height))
	bg.CornerRadius = BorderRadius

	// Animated border glow
	borderGlow := canvas.NewRectangle(ColorBorderCyan)
	borderGlow.SetMinSize(fyne.NewSize(width+2, height+2))
	borderGlow.CornerRadius = BorderRadius

	// Outer subtle glow
	outerGlow := canvas.NewRectangle(color.NRGBA{R: ColorBorderCyan.R, G: ColorBorderCyan.G, B: ColorBorderCyan.B, A: 20})
	outerGlow.SetMinSize(fyne.NewSize(width+6, height+6))
	outerGlow.CornerRadius = BorderRadius + 2

	// Start glow animation
	go animateCardGlow(borderGlow, outerGlow)

	return container.NewStack(
		container.NewCenter(outerGlow),
		container.NewCenter(borderGlow),
		container.NewCenter(bg),
		container.NewPadded(content),
	)
}

// animateCardGlow creates a subtle pulsing effect for card borders
func animateCardGlow(border, outer *canvas.Rectangle) {
	ticker := time.NewTicker(time.Millisecond * 80)
	defer ticker.Stop()

	phase := 0.0
	for range ticker.C {
		phase += 0.04
		if phase > 6.28 {
			phase = 0
		}

		// Calculate pulsing alpha
		pulse := math.Sin(phase)
		baseAlpha := uint8(120 + pulse*40)
		outerAlpha := uint8(15 + pulse*10)

		// Update UI on main thread
		fyne.Do(func() {
			border.FillColor = color.NRGBA{R: ColorBorderCyan.R, G: ColorBorderCyan.G, B: ColorBorderCyan.B, A: baseAlpha}
			outer.FillColor = color.NRGBA{R: ColorBorderCyan.R, G: ColorBorderCyan.G, B: ColorBorderCyan.B, A: outerAlpha}
			border.Refresh()
			outer.Refresh()
		})
	}
}

// CreateStyledInput creates an input field with enhanced styling
func CreateStyledInput(entry *widget.Entry, width, height float32) fyne.CanvasObject {
	// Background with subtle gradient effect
	bg := canvas.NewRectangle(ColorInputBg)
	bg.SetMinSize(fyne.NewSize(width, height))
	bg.CornerRadius = BorderRadius

	// Border
	border := canvas.NewRectangle(color.NRGBA{R: 34, G: 211, B: 238, A: 80})
	border.SetMinSize(fyne.NewSize(width+2, height+2))
	border.CornerRadius = BorderRadius

	return container.NewStack(
		container.NewCenter(border),
		container.NewCenter(bg),
		container.NewPadded(entry),
	)
}

// CreateStyledPasswordInput creates a password input with enhanced styling
func CreateStyledPasswordInput(entry *widget.Entry, width, height float32) fyne.CanvasObject {
	return CreateStyledInput(entry, width, height)
}

// CreateIconButton creates a small icon button with glow
func CreateIconButton(icon, label string, onClick func(), btnColor color.Color) fyne.CanvasObject {
	btn := widget.NewButton(icon+" "+label, onClick)
	btn.Importance = widget.LowImportance

	// Button glow
	glow := canvas.NewRectangle(color.NRGBA{
		R: btnColor.(color.NRGBA).R,
		G: btnColor.(color.NRGBA).G,
		B: btnColor.(color.NRGBA).B,
		A: 40,
	})
	glow.CornerRadius = 4
	glow.SetMinSize(fyne.NewSize(72, 32))

	return container.NewStack(
		container.NewCenter(glow),
		btn,
	)
}

// CreateHeaderText creates stylized header text with glow
func CreateHeaderText(text string, size float32) fyne.CanvasObject {
	// Main text
	mainText := canvas.NewText(text, ColorAccentCyan)
	mainText.TextSize = size
	mainText.TextStyle = fyne.TextStyle{Bold: true}

	// Shadow/glow layer
	shadowText := canvas.NewText(text, color.NRGBA{R: 34, G: 211, B: 238, A: 60})
	shadowText.TextSize = size
	shadowText.TextStyle = fyne.TextStyle{Bold: true}

	return container.NewStack(
		container.NewPadded(shadowText),
		mainText,
	)
}

// CreateLoadingSpinner creates an animated loading indicator
func CreateLoadingSpinner() fyne.CanvasObject {
	spinner := widget.NewProgressBarInfinite()

	spinnerBg := canvas.NewRectangle(ColorCardBg)
	spinnerBg.CornerRadius = BorderRadius
	spinnerBg.SetMinSize(fyne.NewSize(200, 40))

	return container.NewStack(
		spinnerBg,
		container.NewPadded(spinner),
	)
}

// CreateStatusIndicator creates a colored status indicator
func CreateStatusIndicator(statusText string, statusType string) fyne.CanvasObject {
	var indicatorColor color.Color
	var icon string

	switch statusType {
	case "success":
		indicatorColor = ColorSuccess
		icon = "✓"
	case "error":
		indicatorColor = ColorDanger
		icon = "✕"
	case "warning":
		indicatorColor = ColorWarning
		icon = "⚠"
	case "info":
		indicatorColor = ColorAccentCyan
		icon = "ℹ"
	default:
		indicatorColor = ColorTextSecondary
		icon = "•"
	}

	iconLabel := CreateLabel(icon, 14, indicatorColor, true)
	textLabel := CreateLabel(statusText, 11, ColorTextPrimary, false)

	// Background
	bg := canvas.NewRectangle(color.NRGBA{
		R: indicatorColor.(color.NRGBA).R,
		G: indicatorColor.(color.NRGBA).G,
		B: indicatorColor.(color.NRGBA).B,
		A: 30,
	})
	bg.CornerRadius = BorderRadius
	bg.SetMinSize(fyne.NewSize(300, 40))

	content := container.NewHBox(iconLabel, textLabel)

	return container.NewStack(
		bg,
		container.NewCenter(content),
	)
}

// CreateGlowingDivider creates an animated glowing divider
func CreateGlowingDivider() fyne.CanvasObject {
	line := canvas.NewRectangle(ColorBorderCyan)
	line.SetMinSize(fyne.NewSize(500, 2))

	// Animate glow
	go func() {
		ticker := time.NewTicker(time.Millisecond * 100)
		defer ticker.Stop()

		phase := 0.0
		for range ticker.C {
			phase += 0.05
			if phase > 6.28 {
				phase = 0
			}

			pulse := math.Sin(phase)
			alpha := uint8(100 + pulse*80)

			// Update UI on main thread
			fyne.Do(func() {
				line.FillColor = color.NRGBA{R: 34, G: 211, B: 238, A: alpha}
				line.Refresh()
			})
		}
	}()

	return line
}

// CreateToolbar creates a styled toolbar with action buttons
func CreateToolbar(buttons []fyne.CanvasObject) fyne.CanvasObject {
	bg := canvas.NewRectangle(ColorSidebarBg)
	bg.CornerRadius = BorderRadius
	bg.SetMinSize(fyne.NewSize(800, 50))

	buttonContainer := container.NewHBox(buttons...)

	return container.NewStack(
		bg,
		container.NewCenter(buttonContainer),
	)
}

// CreateMetricCard creates a card showing a metric with label and value
func CreateMetricCard(label, value string, iconColor color.Color) fyne.CanvasObject {
	valueText := CreateLabel(value, 20, ColorTextPrimary, true)
	labelText := CreateLabel(label, 10, ColorTextSecondary, false)

	content := container.NewVBox(
		container.NewCenter(valueText),
		container.NewCenter(labelText),
	)

	return CreateEnhancedCard(content, 180, 100)
}

// CreateTooltip creates a styled tooltip
func CreateTooltip(text string) fyne.CanvasObject {
	bg := canvas.NewRectangle(ColorCardBg)
	bg.CornerRadius = 6
	bg.SetMinSize(fyne.NewSize(200, 40))

	border := canvas.NewRectangle(ColorBorderCyan)
	border.CornerRadius = 6
	border.SetMinSize(fyne.NewSize(202, 42))

	label := CreateLabel(text, 9, ColorTextSecondary, false)

	return container.NewStack(
		border,
		bg,
		container.NewCenter(label),
	)
}

// CreateResponsiveContainer creates a container that adapts to screen size
func CreateResponsiveContainer(content fyne.CanvasObject, minWidth, minHeight float32) fyne.CanvasObject {
	scroll := container.NewScroll(content)
	scroll.SetMinSize(fyne.NewSize(minWidth, minHeight))
	return scroll
}
