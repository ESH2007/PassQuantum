# PassQuantum UI - Visual Navigation Guide

## Application Flow Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                   START APPLICATION                         │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
        ╔═══════════════════════════╗
        ║   LOGIN SCREEN            ║
        ║                           ║
        ║  ┌─────────────────────┐  ║
        ║  │   PassQuantum Logo  │  ║
        ║  └─────────────────────┘  ║
        ║  Quantum-Proof Encryption ║
        ║                           ║
        ║  [  Master Password  ]    ║
        ║  [  DESBLOQUEAR    ]      ║
        ║  (Neon button style)      ║
        ╚═════────────┬─────────────╝
                      │
        ┌─────────────┴──────────────┐
        │ Validate Master Password   │
        └─────────────┬──────────────┘
                      │
                      ▼
        ╔═════════════════════════════════════════╗
        ║   VAULT SELECTION SCREEN                ║
        ║                                         ║
        ║  Your Vaults                            ║
        ║  ───────────                            ║
        ║                                         ║
        ║  ┌─────────────────────────────────┐  ║
        ║  │ ┌─────────────────────────────┐ │  ║
        ║  │ │ Personal                    │ │  ║
        ║  │ │ Location: vaults/Personal.. │ │  ║
        ║  │ │ [Open] [Delete]             │ │  ║
        ║  │ └─────────────────────────────┘ │  ║
        ║  │ ┌─────────────────────────────┐ │  ║
        ║  │ │ Work                        │ │  ║
        ║  │ │ Location: vaults/Work...    │ │  ║
        ║  │ │ [Open] [Delete]             │ │  ║
        ║  │ └─────────────────────────────┘ │  ║
        ║  │ ┌─────────────────────────────┐ │  ║
        ║  │ │ Finance                     │ │  ║
        ║  │ │ Location: vaults/Finance... │ │  ║
        ║  │ │ [Open] [Delete]             │ │  ║
        ║  │ └─────────────────────────────┘ │  ║
        ║  └─────────────────────────────────┘  ║
        ║                                         ║
        ║  [+ Create New Vault] [⚙ Settings]     ║
        ║  [🔒 Lock & Exit]                      ║
        ╚═════────────┬──────────────────────────╝
                      │
        ┌─────────────┼──────────────┬──────────────┐
        │             │              │              │
        │ [Open]      │ [Settings]   │ [Create New] │
        ▼             ▼              ▼              ▼
    ╔════╗    ╔════════════╗   ╔═══════════╗   ╔═════╗
    ║Main║    ║ Settings   ║   ║ Create    ║   │Done │
    ║Scrn║    ║ Screen     ║   ║ Vault     ║   │(→)  │
    ╚════╝    ╚════╤═══════╝   ╚═══════════╝   └─────┘
        │           │
        │           └──────────────────────────────┐
        │                                          │
        ▼                                          │
╔═════════════════════════════════════════╗       │
║   MAIN SCREEN (Password Manager)         ║       │
║                                         ║       │
║  PassQuantum - Personal                 ║       │
║  (Vault is encrypted and secured)       ║       │
║                                         ║       │
║  [ Add New Password Entry ]              ║       │
║  ─────────────────────────────────────  ║       │
║  Service Name: [Gmail          ]         ║       │
║  Username/Email: [user@gmail.  ]         ║       │
║  Password: [**** ****]                   ║       │
║                                         ║       │
║  [Add Password] [View All]               ║       │
║  [← Back to Vaults] [🔒 Lock & Exit]     ║       │
╚═════────────┬──────────────────────────╝       │
              │                                   │
              │ [View All]                        │
              │                                   ▼
              │                          ╔═══════════════╗
              ▼                          ║ 5 Settings Tabs
    ╔═════════════════════════╗           ║ With options  ║
    ║ PASSWORDS VIEW SCREEN    ║          ╚═══════════════╝
    ║                         ║
    ║ Your Passwords          ║
    ║ Total: 12 passwords     ║
    ║ ─────────────────────  ║
    ║                         ║
    ║ ┌─────────────────────┐ ║
    ║ │ #1 - Gmail          │ ║
    ║ │ 👤 user@gmail.com   │ ║
    ║ │ 🔐 ••••••••••       │ ║
    ║ │ [👁 Show] [📋 Copy] │ ║
    ║ │ [🗑 Delete]         │ ║
    ║ └─────────────────────┘ ║
    ║                         ║
    ║ ┌─────────────────────┐ ║
    ║ │ #2 - GitHub         │ ║
    ║ │ 👤 myusername       │ ║
    ║ │ 🔐 ••••••••••       │ ║
    ║ │ [👁 Show] [📋 Copy] │ ║
    ║ │ [🗑 Delete]         │ ║
    ║ └─────────────────────┘ ║
    ║                         ║
    ║ ... (more cards) ...     ║
    ║                         ║
    ║ [← Back]                 ║
    ╚═════────────┬──────────┘
                  │
                  └──→ Back to Main Screen
```

---

## Settings Screen - Tab Navigation

```
╔═══════════════════════════════════════════════════════════════╗
║ Settings                                                      ║
╠═══════════════════════════════════════════════════════════════╣
║ [Security] [Vault] [Display] [Backup] [About]                 ║
╠═══════════════════════════════════════════════════════════════╣
║                                                               ║
║                   Tab Content Area                            ║
║                (Scrollable if needed)                         ║
║                                                               ║
║                                                               ║
╠═══════════════════════════════════════════════════════════════╣
║ ← Back                                                        ║
╚═══════════════════════════════════════════════════════════════╝

SECURITY TAB:
┌─────────────────────────────────────────┐
│ Password Strength: [Strong ▼]            │
│                                         │
│ [Change Master Password]                │
│                                         │
│ ☐ Enable Two-Factor Authentication      │
│   (Coming Soon)                         │
│                                         │
│ Auto-lock timeout: [15 minutes ▼]       │
│ Clipboard timeout: [30 seconds ▼]       │
└─────────────────────────────────────────┘

VAULT TAB:
┌─────────────────────────────────────────┐
│ Current Vault: Personal                 │
│ Total Vaults: 3                         │
│                                         │
│ [Compact Vault]                         │
│                                         │
│ [Export Vault (Encrypted)]              │
│ [Import Vault Backup]                   │
│                                         │
│ Last Backup: 2 hours ago                │
└─────────────────────────────────────────┘

DISPLAY TAB:
┌─────────────────────────────────────────┐
│ Theme: [Dark ▼]                         │
│ Font Size: [Medium ▼]                   │
│                                         │
│ ☐ Show password on hover                │
│ ☑ Confirm before deleting               │
└─────────────────────────────────────────┘

BACKUP TAB:
┌─────────────────────────────────────────┐
│ ☐ Enable automatic backups              │
│                                         │
│ Backup frequency: [Weekly ▼]            │
│                                         │
│ [Backup Now]                            │
│ [Restore from Backup]                   │
│                                         │
│ ☁️  Cloud backup (Coming soon)          │
└─────────────────────────────────────────┘

ABOUT TAB:
┌─────────────────────────────────────────┐
│        ┌──────────────────────┐          │
│        │ [PassQuantum Logo]  │          │
│        └──────────────────────┘          │
│                                         │
│     PassQuantum v1.0.0                  │
│  Quantum-Proof Password Manager         │
│                                         │
│ ✓ Post-Quantum Cryptography             │
│ ✓ AES-256-GCM Encryption                │
│ ✓ Multiple Vault Support                │
│ ✓ Secure Key Derivation                 │
│                                         │
│ [📖 Documentation] [🔄 Check Updates]   │
└─────────────────────────────────────────┘
```

---

## Color Scheme & Visual Style

```
PRIMARY COLORS:
├─ Cyan/Turquoise: #00FFDC (A120)  - Neon borders, highlights
├─ Dark Background: #121214 (FF)   - Main surface color
└─ Text: #FFFFFF (FF)               - Primary text

ACCENT COLORS:
├─ Success Green: #00FF00           - Positive actions
├─ Error Red: #FF0000               - Errors & warnings
├─ Gray: #808080                    - Disabled elements
└─ Light Gray: #CCCCCC              - Secondary text

COMPONENT STYLING:
Cards:
  ├─ Background: #282830
  ├─ Border: Cyan with glow effect
  └─ Padding: 16px

Buttons:
  ├─ Primary: High Importance (Cyan bg)
  ├─ Secondary: Normal (Gray bg)
  └─ Danger: Red (for delete)

Text:
  ├─ Bold: Headers, titles
  ├─ Regular: Body text
  └─ Italic: Helper/secondary text
```

---

## User Interaction Patterns

### Password Entry Card - Full Interaction

```
Initial State:
┌─────────────────────────────────────┐
│ #1 - Gmail                          │
│ 👤 user@gmail.com                   │
│ 🔐 ••••••••••                       │
│ [👁 Show] [📋 Copy] [🗑 Delete]    │
└─────────────────────────────────────┘
         ▲
         │ Click "Show"
         │
Revealed State:
┌─────────────────────────────────────┐
│ #1 - Gmail                          │
│ 👤 user@gmail.com                   │
│ 🔐 Correct-Horse-Battery-Staple    │
│ [👁 Hide] [📋 Copy] [🗑 Delete]    │
└─────────────────────────────────────┘
         ▲
         │ Click "Copy"
         │
Feedback:
┌─────────────────┐
│ Password copied │
│ to clipboard!   │
│     [OK]        │
└─────────────────┘
```

### Create Vault Dialog Flow

```
Step 1: Trigger
┌─────────────────────────────────────┐
│ [+ Create New Vault]  (clicked)     │
└─────────────────────────────────────┘
              │
              ▼
Step 2: Form Dialog
┌─────────────────────────────────────┐
│ Create New Vault                    │
├─────────────────────────────────────┤
│ Vault Name:                         │
│ [_________________________]          │
│                                     │
│ Master Password:                    │
│ [_________________________]          │
│                                     │
│ Confirm Password:                   │
│ [_________________________]          │
│                                     │
│ [Create] [Cancel]                   │
└─────────────────────────────────────┘
              │
       ┌──────┴──────┐
       │             │
   [Create]     [Cancel]
       │             │
       ▼             ▼
 Vault Created    Form Closed
```

---

## Navigation Breadcrumb Examples

```
Scenario 1: Browse and View
Login → Vault Selection → Main Screen → View Passwords
           ↑                                    │
           └────────────────────────────────────┘
                    (Back button)

Scenario 2: Settings Access
Login → Vault Selection → Settings (5 tabs)
           ↑
           └──── (Back button returns here)

Scenario 3: Create Multiple Vaults
Login → Create Vault "Work" → Main Screen
           ↑                        │
           └─── [Back to Vaults] ───┘
                     │
            Vault Selection
                     │
            [Create New Vault] 
                     │
                  "Finance" 
                     │
                Back to Selection
```

---

## Responsive Behavior

### Window Sizes
```
Login Screen:        500x400 (fixed)
Vault Selection:     600x500
Main Screen:         600x500 
Passwords View:      700x550
Settings:            600x700+ (scrollable)
```

### Mobile Adaptation (Future)
- Touch-friendly button sizes
- Vertical scrolling for long lists
- Simplified dialogs for small screens
- Landscape/portrait support

---

## Error & Confirmation Dialogs

```
Error Dialog:
┌─────────────────────────────────────┐
│ ❌ Error                             │
├─────────────────────────────────────┤
│                                     │
│ Invalid master password or vault    │
│ corrupted. Please try again.        │
│                                     │
│                  [OK]               │
└─────────────────────────────────────┘

Success Dialog:
┌─────────────────────────────────────┐
│ ✓ Success                           │
├─────────────────────────────────────┤
│                                     │
│ Password saved successfully!        │
│                                     │
│                  [OK]               │
└─────────────────────────────────────┘

Confirmation Dialog:
┌─────────────────────────────────────┐
│ ⚠️  Delete Vault                    │
├─────────────────────────────────────┤
│                                     │
│ Are you sure you want to delete     │
│ 'Work'? This cannot be undone.      │
│                                     │
│           [Delete] [Cancel]         │
└─────────────────────────────────────┘
```

---

## Accessibility Features

### Keyboard Navigation
```
Tab     → Cycle through controls
Enter   → Activate button / Submit form
Escape  → Close dialog / Cancel
Arrow ↑ → Previous item in list
Arrow ↓ → Next item in list
```

### Screen Reader Support
- All buttons have descriptive labels
- Form inputs have clear names
- Dialogs have titles
- Icons paired with text
- Error messages are descriptive

### High Contrast Elements
- Cyan borders on dark background
- Good text/background contrast ratio
- Icons + text for actions
- Color-blind friendly palette

---

## Performance Indicators

```
Loading State:
│ ⏳ Loading passwords...
│
After 1-2 seconds:
│ Password list appears

Long Operation:
│ 
│ 🔄 Encrypting vault...
│ (Does not block UI)
│
After encryption:
│ ✓ Vault saved
```

---

This visual guide complements the API documentation and can be used for:
- User training
- Feature demonstrations
- UI/UX reviews
- Development reference
- Future enhancements planning
