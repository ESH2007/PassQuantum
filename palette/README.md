# palette/

Image-based palette extraction and Fyne theme construction.
Used by the theme picker in the UI to derive a coherent color scheme
from a user-supplied image.

| File | Description |
|---|---|
| `extractor.go` | Pixel sampling and k-means clustering to extract a representative color palette from a `image.Image`. |
| `color_utils.go` | Utility functions for luminance, saturation, and contrast calculations used when selecting role colors from the palette. |
| `roles.go` | Maps extracted palette colors to semantic roles (background, foreground, accent, etc.) based on luminance/contrast heuristics. |
| `theme.go` | Fyne `fyne.Theme` adapter that wraps a `PaletteRoles` and satisfies the Fyne theme interface so the extracted palette can be applied to the app. |
