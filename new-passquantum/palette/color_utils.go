package palette

import (
	"image/color"
	"math"
)

func Luminance(c color.RGBA) float64 {
	lin := func(v uint8) float64 {
		srgb := float64(v) / 255.0
		if srgb <= 0.04045 {
			return srgb / 12.92
		}
		return math.Pow((srgb+0.055)/1.055, 2.4)
	}

	r := lin(c.R)
	g := lin(c.G)
	b := lin(c.B)
	return 0.2126*r + 0.7152*g + 0.0722*b
}

func Saturation(c color.RGBA) float64 {
	r := float64(c.R) / 255.0
	g := float64(c.G) / 255.0
	b := float64(c.B) / 255.0

	maxV := math.Max(r, math.Max(g, b))
	minV := math.Min(r, math.Min(g, b))
	if maxV == minV {
		return 0
	}

	delta := maxV - minV
	l := (maxV + minV) / 2
	if l > 0.5 {
		return delta / (2 - maxV - minV)
	}
	return delta / (maxV + minV)
}

func Lighten(c color.RGBA, factor float64) color.RGBA {
	f := clamp01(factor)
	return color.RGBA{
		R: uint8(float64(c.R) + (255.0-float64(c.R))*f),
		G: uint8(float64(c.G) + (255.0-float64(c.G))*f),
		B: uint8(float64(c.B) + (255.0-float64(c.B))*f),
		A: c.A,
	}
}

func Darken(c color.RGBA, factor float64) color.RGBA {
	f := 1 - clamp01(factor)
	return color.RGBA{
		R: uint8(float64(c.R) * f),
		G: uint8(float64(c.G) * f),
		B: uint8(float64(c.B) * f),
		A: c.A,
	}
}

func WithAlpha(c color.RGBA, a uint8) color.RGBA {
	c.A = a
	return c
}

func ContrastColor(bg color.RGBA) color.RGBA {
	if Luminance(bg) > 0.179 {
		return color.RGBA{R: 0, G: 0, B: 0, A: 255}
	}
	return color.RGBA{R: 255, G: 255, B: 255, A: 255}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
