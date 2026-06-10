# theme

Package `theme` holds PassQuantum's design system: the color tokens, font
assets, icon set, and widget factory functions that give the app its look. The
palette is a refined dark scheme — a teal accent (`#2dd4bf`) on near-black
blue-grey backgrounds (`#0a0d12`) — the result of the redesign away from the
older neon style (some factory names such as `CreateNeonButton` predate that
rename and are kept for compatibility).

## Contents

- **theme.go** — color/design tokens (`ColorBg`, `ColorAccentCyan`, …), base widget factories (`CreateNeonButton`, `CreateCard`, `CreateLabel`, `CreateStyledInput`, `CreateNavigationSidebar`), and the `Particle`/`ParticleBackground` animated widget, `GlowButton`, and `NavigationItem` types
- **components.go** — the higher-level design-system components: status pills (`StatusPill`), `PageHeader`, `SectionEyebrow`, segmented strength meters, and other composed building blocks (the largest file in the package)
- **enhancements.go** — richer composite widgets: `CreateEnhancedCard`, `CreateStyledInput`/`CreateStyledPasswordInput`, `CreateIconButton`, `CreateHeaderText`, `CreateLoadingSpinner`
- **quantum_theme.go** — `QuantumTheme`, the `fyne.Theme` implementation that feeds the tokens to Fyne's built-in widgets
- **icons.go** — inline SVG icon resources (`svgIcon` builds `fyne.StaticResource`s from path data)
- **fonts.go** — embedded font assets (`//go:embed`)
- **theme_test.go** — contrast-ratio tests for adaptive text color selection

## Design tokens

All `Color*` variables are exported package-level `color.NRGBA` vars. They can be
mutated at runtime (by the color-personalisation screen in `ui/screens`) to apply
custom palettes without restarting the app.

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
