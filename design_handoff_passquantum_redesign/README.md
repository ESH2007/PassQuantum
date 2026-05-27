# Handoff: PassQuantum UI Redesign

## Overview

This is a **high-fidelity design reference** for a complete visual redesign of the PassQuantum Fyne/Go desktop app.
The direction is "defense / cryptographic, quiet utilitarian" — think Signal Desktop or Tailscale's UI — serious,
readable, restrained. The design moves away from the neon cyan + magenta palette toward a single institutional blue
accent on graphite slate.

**Your task:** implement this design inside the existing Fyne codebase using a custom `fyne.Theme` and updated
widget layouts. The bundled HTML prototype is a **design reference only** — do not ship it. Recreate the look,
structure, and interactions in Fyne Go.

---

## Fidelity

**High-fidelity.** Colors, typography, spacing, component anatomy, and interactions are all final. Implement
pixel-faithfully within Fyne's constraints. Where Fyne's widget system cannot achieve an exact effect (e.g. a
blurred top bar or a CSS hex-lattice background), approximate with the closest Fyne equivalent and note any
deviation.

---

## Design Files

| File | Purpose |
|---|---|
| `PassQuantum.html` | Full interactive prototype — all screens, all states, Tweaks panel |
| `tokens.css` | All design tokens (colors, type scale, spacing, radii) |
| `components.jsx` | All shared primitives: Button, Field, Input, Card, Pill, StrengthMeter, Icons |
| `screens.jsx` | Every screen component: Vaults, AddItem, Items, Generator, Checker, Settings tabs |
| `app.jsx` | App shell: sidebar, topbar, routing, tweaks wiring |

Open `PassQuantum.html` in a browser to interact with the full prototype. Use the **Tweaks** button (bottom-right)
to see accent/density/font/layout variations.

---

## Design Tokens

### Colors

Map these to a custom `fyne.Theme` implementation. Implement `Color(name theme.ColorName, variant theme.ThemeVariant) color.Color`.

```go
// Surfaces
bg0  := color.NRGBA{R: 0x0b, G: 0x0e, B: 0x13, A: 0xff} // page / window background
bg1  := color.NRGBA{R: 0x0f, G: 0x13, B: 0x19, A: 0xff} // sidebar, inputs
bg2  := color.NRGBA{R: 0x13, G: 0x18, B: 0x22, A: 0xff} // card surface, main area
bg3  := color.NRGBA{R: 0x1a, G: 0x20, B: 0x30, A: 0xff} // input hover, selected row
bg4  := color.NRGBA{R: 0x23, G: 0x2a, B: 0x3a, A: 0xff} // pressed state

// Borders / strokes
line1 := color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x0f} // hairlines (alpha 6%)
line2 := color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x1a} // dividers, input borders (10%)
line3 := color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x29} // emphasized borders (16%)

// Text
fg0 := color.NRGBA{R: 0xe7, G: 0xea, B: 0xf0, A: 0xff} // primary
fg1 := color.NRGBA{R: 0xb3, G: 0xba, B: 0xc8, A: 0xff} // secondary
fg2 := color.NRGBA{R: 0x7a, G: 0x82, B: 0x94, A: 0xff} // tertiary / label
fg3 := color.NRGBA{R: 0x55, G: 0x5c, B: 0x6c, A: 0xff} // placeholder

// Accent — institutional blue
accent     := color.NRGBA{R: 0x3b, G: 0x82, B: 0xf6, A: 0xff} // #3b82f6
accentSoft := color.NRGBA{R: 0x3b, G: 0x82, B: 0xf6, A: 0x24} // 14% alpha — bg tint
accentLine := color.NRGBA{R: 0x3b, G: 0x82, B: 0xf6, A: 0x66} // 40% alpha — borders
accentFg   := color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff} // text on accent btn

// Semantic
ok      := color.NRGBA{R: 0x2e, G: 0xa9, B: 0x6b, A: 0xff} // #2ea96b green
warn    := color.NRGBA{R: 0xd9, G: 0x90, B: 0x30, A: 0xff} // #d99030 amber
danger  := color.NRGBA{R: 0xd0, G: 0x4a, B: 0x4a, A: 0xff} // #d04a4a red

// Strength meter (5 levels)
str1 := "#d04a4a" // very weak
str2 := "#d97a3a" // weak
str3 := "#d9b73a" // fair
str4 := "#6fb73a" // strong
str5 := "#2ea96b" // excellent
```

**Fyne theme mapping:**

| Fyne ColorName | Use this token |
|---|---|
| `theme.ColorNameBackground` | `bg0` |
| `theme.ColorNameMenuBackground` | `bg1` |
| `theme.ColorNameOverlayBackground` | `bg2` |
| `theme.ColorNameButton` | `bg3` |
| `theme.ColorNameInputBackground` | `bg1` |
| `theme.ColorNameInputBorder` | `line2` |
| `theme.ColorNameForeground` | `fg0` |
| `theme.ColorNamePlaceHolder` | `fg3` |
| `theme.ColorNameDisabled` | `fg2` |
| `theme.ColorNameShadow` | `rgba(0,0,0,0.40)` |
| `theme.ColorNamePrimary` | `accent` |
| `theme.ColorNameFocus` | `accentLine` |
| `theme.ColorNameSelection` | `accentSoft` |
| `theme.ColorNameSuccess` | `ok` |
| `theme.ColorNameWarning` | `warn` |
| `theme.ColorNameError` | `danger` |
| `theme.ColorNameSeparator` | `line1` |

### Typography

```go
// Primary font: IBM Plex Sans (weights: 400 regular, 500 medium, 600 semibold)
// Mono font:    IBM Plex Mono (weights: 400, 500)
// Download: https://fonts.google.com/specimen/IBM+Plex+Sans

// Fyne font sizes (in theme.Size calls):
theme.SizeNameText           → 13px  // body
theme.SizeNameHeadingText    → 22px  // page H1
theme.SizeNameSubHeadingText → 15px  // card titles, modal H3
theme.SizeNameCaptionText    → 11px  // mono labels, eyebrows
theme.SizeNameInputBorder    → 1px

// Custom sizes to handle in layout:
// eyebrow (uppercase mono labels): 10px, weight 600, letter-spacing 0.08em
// page sub-heading: 13px, color fg1
// mono data (paths, hashes, entropy): 11–13px, IBM Plex Mono
```

To load IBM Plex fonts in Fyne, embed the TTF files and implement `Font()` in your custom theme:

```go
//go:embed fonts/IBMPlexSans-Regular.ttf
var ibmPlexSansRegular []byte
//go:embed fonts/IBMPlexSans-Medium.ttf
var ibmPlexSansMedium []byte
//go:embed fonts/IBMPlexMono-Regular.ttf
var ibmPlexMonoRegular []byte

func (t QuantumTheme) Font(style fyne.TextStyle) fyne.Resource {
    if style.Monospace {
        return fyne.NewStaticResource("IBMPlexMono", ibmPlexMonoRegular)
    }
    if style.Bold {
        return fyne.NewStaticResource("IBMPlexSans-Medium", ibmPlexSansMedium)
    }
    return fyne.NewStaticResource("IBMPlexSans", ibmPlexSansRegular)
}
```

### Spacing

```go
// Base unit: 4px. All spacing is multiples.
space1 := 4
space2 := 8
space3 := 12
space4 := 16
space5 := 20
space6 := 24
space7 := 32
space8 := 40
space9 := 56

// Fyne theme sizes:
theme.SizeNameInnerPadding  → 8   // tight inner padding
theme.SizeNamePadding       → 12  // standard padding
theme.SizeNameScrollBar     → 10
theme.SizeNameScrollBarSmall→ 6
```

### Border Radii

Fyne does not have a global radius token — apply per custom widget:

```
radius1 → 4px  // checkboxes, small pills, meter segments
radius2 → 6px  // inputs, buttons, icon buttons
radius3 → 10px // cards
radius4 → 14px // modals
```

---

## App Shell

### Window

- **No titlebar chrome** — the design is borderless/undecorated. In Fyne: `w.SetFixedSize(false)` — but keep the system title bar (Fyne doesn't support true borderless on all platforms; use the standard window and focus on interior layout). Use `w.SetPadded(false)` so the outer padding doesn't fight the design.

### Layout

```
┌──────────────┬───────────────────────────────────────────┐
│              │ TOPBAR (48px height)                      │
│   SIDEBAR    ├───────────────────────────────────────────┤
│  224px (exp) │                                           │
│   64px (col) │   PAGE CONTENT (scrollable)               │
│              │   max-width: 960px, centered              │
│              │                                           │
└──────────────┴───────────────────────────────────────────┘
```

Use `container.NewBorder(topbar, nil, sidebar, nil, mainContent)` at the top level.

---

## Sidebar

**Expanded width:** 224px | **Collapsed width:** 64px (icons only)

```
┌──────────────────────────┐
│  [◉] PassQuantum v1.0   │  ← brand mark (28×28 rounded rect, accentSoft bg)
│        PQ-Safe          │    + brand name + meta (IBM Plex Mono, 10px, uppercase)
├──────────────────────────┤
│  — VAULT ——             │  ← section eyebrow (Mono, 10px, fg3, uppercase)
│  □ Vaults               │  ← nav item (see below)
│  □ Add item             │
│  □ Items              24 │  ← right-side count badge (Mono, 10px, fg2)
├──────────────────────────┤
│  — TOOLS ——             │
│  □ Generate             │
│  □ Analyze              │
├──────────────────────────┤
│  — SYSTEM ——            │
│  □ Settings             │
├──────────────────────────┤
│  [←] Collapse           │  ← toggle button (Mono, 11px, line2 border, radius2)
│  □ Lock vault           │
└──────────────────────────┘
```

**Nav item anatomy:**
- Height: 32px (compact) / 36px (default)
- Padding: 8px 12px
- Icon: 18×18 Lucide-style stroke icon, `fg2` color
- Label: 13px, weight 500, `fg1`
- Hover: `bg2` background, `fg0` text + `accent` icon
- Active (current page):
  - Background: `bg3`
  - Border: 1px `line2` on all sides
  - Left accent bar: 2px wide, `accent` color, inset 8px top & bottom
  - Icon: `accent`
  - Text: `fg0`

**Collapsed state:**
- Hide label, section headers, brand text
- Center icon
- Show tooltip on hover: `bg3` background, 12px text, `line2` border, 4px radius, appears to the right

**Icons used** (implement as SVG resources or find equivalent in Fyne's built-in icons — prefer custom SVG):

| Nav item | Icon description |
|---|---|
| Vaults | Rounded rectangle with inner circle + 4 tick marks (vault dial) |
| Add item | Plus sign |
| Items | Key (circle + shaft + 2 notches) |
| Generate | Magic wand with sparkles |
| Analyze | Shield with checkmark |
| Settings | Gear / cogwheel |
| Lock vault | Padlock (closed) |
| Collapse | Panel-left-close chevron |

---

## Topbar

**Height:** 48px | **Background:** `bg0` at 78% opacity with backdrop blur (use solid `bg1` in Fyne as blur isn't native)

```
[breadcrumb path: Vaults / Default / Items]  ——spacer——  [Vault · Default pill] [Watching · ON pill] [Unlocked pill]
```

**Breadcrumb:** 13px, `fg2`, IBM Plex Mono, items separated by `/` in `fg3`

**Status pills:**
```
Border: 1px rounded-full, background tinted per state, Mono 10px uppercase
dot: 6px circle, colored per state

Vault pill:       bg=accentSoft, border=accentLine, text=#c9dcfb
Watching ON:      bg=okSoft,     border=rgba(ok,0.35), text=#b6e7cb, dot glows ok
Watching OFF:     bg=bg2,        border=line2,          text=fg2
Unlocked pill:    bg=bg2,        border=line2,          text=fg2
```

---

## Screen: Vaults

**Page header:**
- Eyebrow: `PassQuantum / Vaults` (Mono 10px, fg2, uppercase)
- H1: "Your vaults" (22px, weight 600, fg0, letter-spacing -0.015em)
- Sub: "Encrypted containers…" (13px, fg1, max 56ch)
- Right: Primary button "New vault" with Plus icon

**Vault list item (row card):**
```
[Vault icon 36×36 | bg3, line2 border, radius2]  [Title + Active pill | path + "Modified X ago"]  [Switch | Export | Delete]
```
- Active vault icon bg: `accentSoft`, border: `accentLine`, icon color: `#c9dcfb`
- Title: 13px, weight 600, fg0
- Active pill: ok-toned (bg=okSoft, border=ok@35%, text=#b6e7cb)
- Count pill: muted (bg=bg2, border=line2, Mono 10px)
- Path: Mono 11px, fg2
- Card: bg2, 1px line1 border, 10px radius, 16px 20px padding, hover border→line2

**Encryption status card** (below the list):
- Key-value table: 2 columns, `dt` = Mono 10px uppercase fg2, `dd` = 13px fg0
- Small text below `dd`: 12px fg2, Mono
- Entries: Algorithm, Key encapsulation, KDF, Vault file

---

## Screen: Add Item

**Page header** same pattern. Right: ghost "← All items" button.

**Single card, full-width** (no max narrower than the content area):
- Section header inside card: none — form fields stack directly
- Footer bar at bottom of card (bg1 bg, line1 top border): left = "Encrypted on save · AES-256-GCM" (Mono 11px, fg2) | right = [Cancel ghost] [Save item primary]

**Item type selector:** full-width `<select>` equivalent (Fyne `widget.Select`) — options: Password, Card, Cyphered note.

**Password fields:**
1. Service — text input, placeholder "e.g. github.com"
2. Username / email — text input
3. Password — password input with eye-toggle + copy icon in trailing slot; label right side has "Generate →" link
4. Strength — meter + label row (see Strength section below)

**Card fields:**
1. Card type — select (Credit / Debit / Prepaid)
2. Nickname — text
3. Cardholder — text
4. Card number — text, Mono
5. Expiry + CVV — 2-column row; CVV has eye-toggle

**Note fields:**
1. Note title — text
2. Cyphered note — multiline textarea, min-height ~180px, Mono, hint below: "Plaintext is encrypted at rest… Markdown is preserved."

---

## Strength Meter Component

Display after the password field. Two rows:

**Row 1:** 5-segment horizontal bar (flex, gap 4px, height 6px, radius 999px) + label right-aligned
- Filled segments use color from `str1–str5` palette
- Empty segments: `bg3`
- Label: Mono 11px, fg1 (e.g. "Weak", "Strong", "Excellent")

**Row 2:** Mono 11px, fg2 — `Crack time · Instantly` · `Entropy · 48 bits`

Strength levels → segment count:
| Score | Level | Label | Crack time |
|---|---|---|---|
| empty | 0 | — | — |
| 1 | 1 | Very weak | Instantly |
| 2 | 2 | Weak | Minutes |
| 3 | 3 | Fair | Days |
| 4 | 4 | Strong | Centuries |
| 5 | 5 | Excellent | 10⁹⁺ centuries |

---

## Screen: Items List

**Page header:** H1 "Vault items · {n}" (the count is muted weight 400), right = "Add item" primary button.

**Search bar:** full-width input with search icon in leading slot (14px icon, fg2).

**Layout toggle:** small segmented control (Rows / Grid) — right of search bar. Active segment: bg1 bg + line2 border shadow.

**Row layout (default):**
```
[Icon 36×36]  [Title + Kind badge | username · ••••••••••••]  [Show] [Copy] [Edit icon] [Delete icon]
```
- Kind badge: Mono 10px, 1px line2 border, radius 4px, 2px 6px padding — text: "Password" / "Card" / "Note"
- Masked secret: Mono 11px, fg1
- Hover on row: border→line2

**Grid layout (alternate):**
- `grid-template-columns: repeat(auto-fill, minmax(260px, 1fr))`
- In Fyne: `container.NewGridWrap(fyne.NewSize(260, 0), ...cards)`
- Each card: icon + title/sub stacked, mono masked row, then [Show] [Copy] [Delete] buttons at bottom

**Delete:** icon button only (trash icon, 28×28, transparent bg, danger color on hover)

---

## Screen: Generator

**Page header:** right = `Pill: CSPRNG · crypto/rand` (accent-toned)

**Output card** ("Generated password"):
- Right of card header: "Regenerate" ghost button with refresh icon
- Password display box: Mono 22px, 18px 20px padding, bg1 bg, line2 border, radius 8px, wordBreak
- Strength block (same as above)
- Footer row: [Copy] [Save to vault primary]

**Options card** ("Parameters"):
- Length: label + right-side number display; range slider + number input (width 76px)
- Character classes: 2-column check grid
  - Uppercase A–Z
  - Lowercase a–z
  - Digits 0–9
  - Symbols !@#$%^&*
  - Exclude ambiguous (i I l 1 L o 0 O) — full width

**Save to Vault modal** (triggered by "Save to vault"):
```
Modal: bg2 bg, line2 border, radius4 (14px), 24px shadow
Header: eyebrow "SAVE TO VAULT" + h3 "New password entry"
Body:
  - Vault selector
  - Service name input
  - Username/email input
  - Generated password (read-only display, Mono, bg1)
Footer: [Cancel ghost] [Save entry primary]
```

---

## Screen: Checker (Strength Analyzer)

**Page header:** right = `Pill: Offline · local only` (ok-toned)

**Input card:** single password input (eye toggle). No HIBP — purely local.

**Analysis card:**
- Strength meter block (same component)
- Divider
- Key-value table:
  - Length: `{n} characters`
  - Character set: `A–Z · a–z · 0–9 · symbols` (only detected sets shown)
  - Estimated entropy: `{n} bits`
  - Brute-force estimate: `{crack time}`
- Placeholder hint if empty: Mono 12px fg2 "Start typing to analyze password strength."

---

## Screen: Settings

**Tabs row:** Security | Vaults | Appearance | About — underline-style tabs (not pill/card). Active tab has 2px `accent` bottom border, `fg0` text. Inactive: `fg2`.

### Tab: Security

**Master password card:**
- 2-col grid: description text | "Change master password" button (with refresh icon)
- Below description: Mono 11px fg2 "Last rotated · 47 days ago · Strength [Strong in ok color]"

**Auto-lock card:**
- Select: inactivity duration (1m / 5m / 15m / 1h / Never)
- 3 checkboxes (see design file for copy)

**Presence Guard card ("Face-detection auto-kill"):**
- 2-col header: description + toggle switch (right-aligned)
  - Toggle: 38×22px, radius 999px, thumb 16×16px white
  - ON state: `accentSoft` bg, `accentLine` border, thumb moves 16px right, thumb color = `accent`
  - OFF state: `bg1` bg, `line2` border, thumb at left, thumb color = `fg3`
- Warning banner:
  - bg: `rgba(217,144,48,0.10)`, border: `rgba(217,144,48,0.30)`, radius 8px
  - Left: AlertTriangle icon in `warn` color
  - Right: bold header "FORCE-KILL WARNING" (#f0d4a2) + body text (12px fg1)
- Process list: each entry is a checkbox row with extra detail:
  - Outer: bg1 bg, line1 border, radius 6px, 10px 12px padding, 12px gap
  - Left: checkbox (16×16, accent when checked)
  - Right: app name (13px fg0) + process name (Mono 11px fg2)
- "Add process to monitor" ghost button with Plus icon

### Tab: Vaults

- Active vault info card (key-value: path, size on disk, items, last modified) — Mono values
- Compact & verify card: description + "Compact vault" button
- Backup & Restore card:
  - 2-column grid: Export column | Import column — each has primary + ghost button
  - Divider
  - Auto-backup row: description + "Back up now" button

### Tab: Appearance

- Density card: refers user to Tweaks panel; lists tweak keys as Mono text
- Accent palette card: 5 swatches (96×56px each, radius2, line2 border) with name + hex below
- App icon card: 3-col grid: [64×64 rounded icon preview] | [filename + dimensions] | [Change / Reset buttons]

### Tab: About

**Product card:** 2-col: [96×96 icon block (accentSoft bg, accentLine border, radius 18px, Atom icon 48px)] | [eyebrow + H2 + mono build info + description paragraph]

**Security stack card (key-value):**
| Key | Value |
|---|---|
| Bulk encryption | AES-256-GCM · 96-bit IV · 128-bit auth tag |
| Key encapsulation | ML-KEM-768 (Kyber) · NIST FIPS 203 |
| Password KDF | Argon2id · m=64 MiB · t=3 · p=4 · 16-byte salt |
| Authentication | HMAC-SHA-512 · per-block integrity |
| Random source | crypto/rand · OS CSPRNG · getrandom(2) |
| Architecture | Zero-knowledge, offline-first · No telemetry |

**Meta card:** License, build hash, maintainer, repo link. [Documentation] [Check for updates] buttons.

---

## Buttons

Three variants:

**Primary:** bg=`accent`, border=`accent`, text=`accentFg` (#fff), hover: 88% accent lightness
**Default:** bg=`bg3`, border=`line2`, text=`fg0`, hover: bg→bg4, border→line3
**Ghost:** bg=transparent, border=transparent, text=`fg1`, hover: bg→bg2, text→fg0
**Danger:** bg=transparent, border=`rgba(danger, 0.35)`, text=`#f3a8a8`, hover: bg→dangerSoft, text=white

Sizes:
- Default: 9px 14px padding, 13px text, radius2 (6px)
- Small: 6px 10px padding, 12px text, radius2

Icon buttons: 28×28px, transparent bg, `fg2` icon, hover: bg→bg3, fg0 icon. 1px transparent border (gains line2 on hover).

---

## Inputs / Fields

```
Background:     bg1
Border:         1px line2
Border-radius:  6px
Padding:        10px 12px
Font:           13px IBM Plex Sans, fg0
Placeholder:    fg3
Focus border:   accent, + 3px accentSoft box-shadow (glow)
Hover border:   line3
```

Password inputs have trailing slot (right: 6px): eye-toggle + copy icon buttons (28×28 each).

Select (dropdown): same as input + right chevron (8×8px, fg2, 45° rotated), no native arrow.

Textarea: same as input, min-height 110px, resizable.

Field label (eyebrow above input): Mono 10px, uppercase, `fg2`, letter-spacing 0.08em. Can have right-side element (e.g. "Generate →").

Field hint (below input): 12px, fg2.

---

## Cards

```
Background:     bg2
Border:         1px line1
Border-radius:  10px
```

Card header (optional): 13px 20px padding, line1 bottom border.
- Left: eyebrow (Mono 10px uppercase fg2) + title (13px weight 600 fg0)
- Right: any action (button, pill)

Card body: 20px padding.

Divide variant: children separated by 1px line1 borders (no outer padding per child; add 16px 20px padding inside each child).

---

## Checkbox

```
Box size:       16×16px, radius 4px
Unchecked:      bg1 bg, line3 border
Checked:        accent bg, accent border, white checkmark (11×11px SVG)
Label:          13px, fg1 unchecked → fg0 checked
```

---

## Section Eyebrows

Used for card headers and settings subsections:
- IBM Plex Mono, 10px, uppercase, letter-spacing 0.08em, `fg2` color

---

## Modals / Dialogs

```
Overlay scrim:  rgba(4,6,10,0.65) + backdrop-blur(4px) — use semi-opaque solid in Fyne
Modal width:    440px (or narrower on small windows)
Background:     bg2
Border:         1px line2
Border-radius:  14px
Shadow:         0 24px 60px rgba(0,0,0,0.5)
```

Parts: header (20px padding, line1 bottom border) | body (20px padding, flex-col gap 16px) | footer (16px 20px padding, line1 top border, bg1 bg, flex-row justify-end gap 12px)

---

## Separator / Divider

1px solid `line1` (`rgba(255,255,255, 0.06)`).

---

## Background Texture (Hex Lattice)

This is purely decorative and is a CSS SVG background-image. In Fyne, approximate with:

```go
// Option A: Custom canvas object that draws a hex grid in a loop
// Option B: Embed the SVG tile as a fyne.Resource and tile it as a canvas.Image
// Option C: Use a plain bg0 background (acceptable fallback)
```

The hex tile SVG (56×64px viewBox):
```svg
<svg xmlns="http://www.w3.org/2000/svg" width="56" height="64" viewBox="0 0 56 64">
  <g fill="none" stroke="#ffffff" stroke-opacity="0.035" stroke-width="1">
    <path d="M14 1 L42 1 L56 16 L42 31 L14 31 L0 16 Z"/>
    <path d="M14 33 L42 33 L56 48 L42 63 L14 63 L0 48 Z"/>
  </g>
</svg>
```

The two faint radial blue glows (top-left and bottom-right, `rgba(59,130,246,0.05)`) can be implemented as
very-low-alpha circles drawn on a `canvas.Circle` or omitted.

---

## Key-Value Table

Used in Vaults (encryption info), Settings → About:

```
2-column grid: label col 180px | value col 1fr
dt: Mono 10px uppercase fg2, letter-spacing 0.04em
dd: 13px fg0
dd small: 12px fg2 Mono, display block, 4px margin-top
```

In Fyne: use `container.NewGridWithColumns(2, ...)` with alternating label/value widgets.

---

## Interactions & Behavior

| Trigger | Behavior |
|---|---|
| Nav item click | Route change (no animation needed; or fade 100ms) |
| Collapse toggle | Sidebar width transitions between 224px and 64px; labels hide |
| Password field input | Strength meter updates live |
| Generator: change length / options | Password regenerates immediately |
| Save item | Item prepended to Items list; route switches to Items |
| Delete item | Item removed from list immediately |
| Save to Vault (generator) | Modal closes; item added to Items list |
| Lock vault | Alert/dialog; clear decrypted state from memory |
| Presence guard toggle | Top bar pill updates between "Watching · ON" and "Presence guard · OFF" |
| Settings tab | Tab switches with underline moving to new tab |
| Items layout toggle | List re-renders in rows or grid layout |

---

## Fyne Implementation Notes

1. **Custom theme first.** Implement `QuantumTheme` (satisfies `fyne.Theme`) before touching any other widget. This gives you correct colors everywhere without per-widget overrides.

2. **Custom widgets for pill and strength meter.** Fyne's built-ins won't match — implement `widget.BaseWidget` + `WidgetRenderer` for:
   - `StatusPill` (colored dot + Mono label + border)
   - `StrengthMeter` (5 colored segments)
   - `NavItem` (left accent bar when active)
   - `SectionEyebrow` (Mono uppercase label)

3. **IBM Plex fonts.** Download the TTF files, embed them with `//go:embed`, and load via the theme's `Font()` method. Without this, every eyebrow and Mono field falls back to the system monospace font, which breaks the design character.

4. **Sidebar collapse.** Drive with a `bool` state variable + `fyne.Animation` on the container width, or simply swap the sidebar container children between expanded and collapsed versions.

5. **Background texture.** The hex lattice is CSS-only. In Fyne, the closest approximation is a `canvas.Raster` that draws the hexagon grid in Go using `image.NRGBA`. Or fall back to plain `bg0` — the design still reads clearly without it.

6. **Top bar backdrop blur.** Not available in Fyne — use solid `bg1` instead.

7. **Focus glow.** Fyne's default focus ring is a colored border. You can simulate the 3px `accentSoft` glow by overriding the input widget's renderer to draw a slightly larger rounded rect behind the input in `accentSoft`.

8. **No emoji.** The original design uses none. The redesign uses none. Don't add any.

---

## Prompt to paste into Claude Code

Paste this into a Claude Code conversation in your PassQuantum repository:

```
I have a design handoff package for a complete visual redesign of the PassQuantum Fyne app.
The folder `design_handoff_passquantum_redesign/` contains:
- README.md — full spec (colors, typography, spacing, every screen, every component, Fyne implementation notes)
- PassQuantum.html — interactive HTML prototype (open in browser to see all screens)
- tokens.css, components.jsx, screens.jsx, app.jsx — design source files

Please read the README.md carefully, then open PassQuantum.html in a browser or inspect the JSX
files to understand the intended design.

Then implement the redesign in the existing Fyne codebase:
1. Create a custom QuantumTheme that implements fyne.Theme with the token values from the README
2. Embed IBM Plex Sans + IBM Plex Mono fonts
3. Create custom widgets: StatusPill, StrengthMeter, NavItem, SectionEyebrow
4. Rebuild the sidebar (collapsible, grouped sections, left-accent active bar)
5. Rebuild the topbar (breadcrumb + status pills)
6. Update each screen to match the README spec: Vaults, AddItem (3 types), Items (rows+grid),
   Generator, Checker, Settings (4 tabs)

Work screen by screen, starting with the theme and sidebar. Reference the README for every
measurement, color, and interaction.
```
