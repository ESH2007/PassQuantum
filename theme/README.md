# theme

Package `theme` provides all visual design tokens and widget factory functions for PassQuantum's cyberpunk UI.

## Contents

- **theme.go** — color constants (`ColorBg`, `ColorAccentCyan`, …), widget factories (`CreateNeonButton`, `CreateCard`, `CreateLabel`, …), `NavigationItem`, `GlowButton`
- **enhancements.go** — animated widgets: `ParticleBackground`, gradient helpers, glow effects
- **theme_test.go** — contrast-ratio tests for adaptive text color selection

## Design tokens

All `Color*` variables are exported package-level `color.NRGBA` vars. They can be mutated at runtime (by the color personalisation screen in `ui/screens`) to apply custom palettes without restarting the app.

## Widget factories

| Factory | Returns |
|---|---|
| `CreateNeonButton` | `*fyne.Container` with glow effect |
| `CreateCard` | Bordered container card |
| `CreateLabel` | Styled `*canvas.Text` |
| `CreateStyledInput` | Input with dark background overlay |
| `CreateNavigationSidebar` | Side nav from `[]NavigationItem` |

## Test

```sh
go test ./theme/...
```
