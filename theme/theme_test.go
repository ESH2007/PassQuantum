package theme

import (
	"image/color"
	"testing"
)

func TestPickAdaptiveTextColorForDarkBackground(t *testing.T) {
	bg := color.NRGBA{R: 11, G: 15, B: 20, A: 255}
	picked := PickAdaptiveTextColor(bg)

	expected := color.NRGBA{R: 245, G: 248, B: 252, A: 255}
	if picked != expected {
		t.Fatalf("expected light text for dark background, got %#v", picked)
	}

	contrast := wcagContrastRatio(picked, bg)
	if contrast < 4.5 {
		t.Fatalf("expected AA contrast >= 4.5, got %.2f", contrast)
	}
}

func TestPickAdaptiveTextColorForLightBackground(t *testing.T) {
	bg := color.NRGBA{R: 242, G: 245, B: 248, A: 255}
	picked := PickAdaptiveTextColor(bg)

	expected := color.NRGBA{R: 20, G: 24, B: 30, A: 255}
	if picked != expected {
		t.Fatalf("expected dark text for light background, got %#v", picked)
	}

	contrast := wcagContrastRatio(picked, bg)
	if contrast < 4.5 {
		t.Fatalf("expected AA contrast >= 4.5, got %.2f", contrast)
	}
}

func TestPickAdaptiveTextColorAlwaysPicksHigherContrastCandidate(t *testing.T) {
	backgrounds := []color.NRGBA{
		{R: 120, G: 120, B: 120, A: 255},
		{R: 80, G: 112, B: 148, A: 255},
		{R: 205, G: 182, B: 150, A: 255},
	}

	light := color.NRGBA{R: 245, G: 248, B: 252, A: 255}
	dark := color.NRGBA{R: 20, G: 24, B: 30, A: 255}

	for _, bg := range backgrounds {
		picked := PickAdaptiveTextColor(bg)
		pickedContrast := wcagContrastRatio(picked, bg)
		lightContrast := wcagContrastRatio(light, bg)
		darkContrast := wcagContrastRatio(dark, bg)

		if pickedContrast < lightContrast || pickedContrast < darkContrast {
			t.Fatalf("picked text %#v did not maximize contrast for background %#v", picked, bg)
		}
	}
}
