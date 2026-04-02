# PassQuantum Settings Screen Design

## Overview
A comprehensive tabbed settings interface organized into 5 main categories with both current and future features.

## Settings Tabs

### 1. Security Tab 🔒

**Purpose:** Master password and session security management

#### Components:

**Password Strength Requirements**
- Dropdown selector: Weak | Medium | Strong | Very Strong
- Helps users set password creation rules
- Future: Enforce minimum strength on new entries

**Master Password Management**
- "Change Master Password" button
- Dialog form with:
  - Current password verification
  - New password entry
  - Password confirmation
- Future: Password complexity validator

**Session Management**
- **Auto-lock Timeout:** 5 min | 15 min | 30 min | 1 hour | Never
  - Automatically lock vault after inactivity
  - Future: Idle time detection

- **Clipboard Clear Timeout:** 15 sec | 30 sec | 1 min | 5 min
  - Auto-clear copied passwords from clipboard
  - Prevents accidental exposure

**Advanced Features (Future)**
- Enable Two-Factor Authentication checkbox
  - Coming soon notice
  - Biometric support (Touch ID, Windows Hello)
  - TOTP integration

---

### 2. Vault Tab 📦

**Purpose:** Vault maintenance and management

#### Components:

**Vault Information**
- Current vault display: "Current Vault: [VaultName]"
- Total vaults counter
- Vault statistics (number of entries, size)

**Vault Maintenance**
- **Compact Vault Button:** "Optimize vault storage"
  - Removes deleted entries
  - Reclaims disk space
  - Background operation
  - Future: Scheduled maintenance

**Backup & Restore**
- **Export Vault (Encrypted):** Download encrypted backup
  - Still encrypted with master password
  - Portable backup
  - Future: Cloud storage integration

- **Import Vault Backup:** Restore from previous export
  - File selector dialog
  - Validation before import
  - Merge or replace option

- **Last Backup Timestamp:** Shows when vault was last backed up

---

### 3. Display Tab 🎨

**Purpose:** User interface customization

#### Components:

**Appearance Settings**
- **Theme Selector:** Dark | Light | System
  - Dark (default): Cyberpunk aesthetic
  - Light: Professional appearance
  - System: Follow OS settings
  - Real-time theme switching

- **Font Size Selector:** Small | Medium | Large
  - Affects all UI text
  - Accessibility option
  - Preserves layout

**Behavior Settings**
- **Show Password on Hover** checkbox
  - Hover over masked password reveals it
  - Accessibility feature
  - Default: OFF (security-first)

- **Confirm Before Deleting Passwords** checkbox
  - Confirmation dialog on delete
  - Prevents accidental loss
  - Default: ON (checked)

---

### 4. Backup Tab 💾

**Purpose:** Automated backup and recovery

#### Components:

**Automatic Backups**
- **Enable Automatic Backups** toggle switch
  - Checkbox to enable feature
  - Default: OFF

- **Backup Frequency Selector:** Daily | Weekly | Monthly
  - Only active if auto-backup enabled
  - Future: Runs in background scheduler

**Manual Backup**
- **Backup Now Button**
  - Creates immediate encrypted backup
  - Updates "Last Backup" timestamp
  - Future: Choose backup location

- **Last Backup Info Display**
  - Shows timestamp of last backup
  - Updates after each backup

**Recovery Options**
- **Restore from Backup Button**
  - File picker for backup selection
  - Confirmation dialog: "This will replace your current vault"
  - Merge vs replace option (future)

**Cloud Backup (Future)**
- Coming soon notice
- Cloud provider selection (future)
- Sync settings (future)

---

### 5. About Tab ℹ️

**Purpose:** Application info, features, and updates

#### Components:

**Application Branding**
- Large logo display
- App name: "PassQuantum" (bold, centered)
- Version: "Version 1.0.0"
- Tagline: "A post-quantum cryptography password manager using Kyber and AES-256-GCM"

**Feature Highlight**
- Bulleted list of key features:
  - ✓ Post-Quantum Cryptography (Kyber-768)
  - ✓ AES-256-GCM Encryption
  - ✓ Multiple Vault Support
  - ✓ Secure Key Derivation
  - ✓ Zero-Knowledge Architecture

**Credits**
- "Developed by: PassQuantum Team"
- "License: MIT"

**Resources**
- **📖 Documentation Button:** Links to GitHub repo
  - Opens browser with documentation
  - Future: In-app help system

- **🔄 Check for Updates Button**
  - Connects to update server (future)
  - Shows "You are running the latest version!"
  - Auto-update capability (future)

---

## Design Patterns

### Tab Navigation
```
┌─────────────────────────────────────────┐
│ [Security] [Vault] [Display] [Backup] [About] │
├─────────────────────────────────────────┤
│                                         │
│     Tab Content Area                    │
│     (Scrollable if needed)              │
│                                         │
├─────────────────────────────────────────┤
│  ← Back          [Settings Title]       │
└─────────────────────────────────────────┘
```

### Card Styling
- Dark background (matching login screen aesthetic)
- Cyan borders for interactive elements
- Clear spacing and hierarchy
- Consistent color scheme throughout

### Form Elements
- **Selectors (Dropdowns):** For preset options
- **Toggles:** For boolean settings
- **Buttons:** Clear action labels with icons
- **Text Input:** Password change fields
- **Displays:** Read-only information (timestamps, paths)

---

## Visual Hierarchy

1. **Tab Headers** - Large, bold titles
2. **Section Headers** - Medium weight, clear grouping
3. **Control Labels** - Regular weight, descriptive
4. **Helper Text** - Small italic for future features

---

## Accessibility Features

- High contrast for readability
- Large touch targets for buttons
- Keyboard navigation support (future)
- Screen reader friendly labels
- Clear error messages
- Visual feedback on interactions

---

## Future Enhancements

### Settings v2.0
- **Profiles:** Save/load different settings configurations
- **Keyboard Shortcuts:** Customizable hotkey bindings
- **Language Support:** Multi-language UI
- **Sync Settings:** Cloud sync across devices
- **Advanced Encryption:** Custom cipher options
- **Privacy Mode:** Disable screenshots, session recording
- **Activity Log:** View login/access history

### Security Enhancements
- Master password strength meter
- Breach database integration
- Password generation policies
- Login attempt tracking
- Session timeout warnings

### Backup Enhancements
- Incremental backups
- Scheduled backup notifications
- Backup verification/integrity check
- Multiple backup destinations
- Backup encryption algorithm selection
- Recovery codes for emergency access

---

## Implementation Notes

All settings components are in `settings_screen.go`:
- `ShowSettingsScreen()` - Main entry point
- `buildSecuritySettings()` - Security tab UI
- `buildVaultSettings()` - Vault tab UI
- `buildDisplaySettings()` - Display tab UI
- `buildBackupSettings()` - Backup tab UI
- `buildAboutSettings()` - About tab UI
- `showChangeMasterPasswordDialog()` - Master password change dialog

### Current Implementation Status
✅ Complete UI structure and layout
✅ All controls and buttons
⏳ Backend logic (future implementation)
⏳ Settings persistence (future)
⏳ Cloud integration (future)
⏳ Feature callbacks (partially implemented)
