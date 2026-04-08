# PassQuantum - User Experience Guide

> Complete user-facing documentation including installation, usage, troubleshooting, and UI navigation.

## 📋 Table of Contents

- [Getting Started](#getting-started)
- [Installation](#installation)
- [First Time Setup](#first-time-setup)
- [Using PassQuantum](#using-passquantum)
- [UI Navigation](#ui-navigation)
- [Managing Passwords](#managing-passwords)
- [Managing Vaults](#managing-vaults)
- [Settings](#settings)
- [Backup and Recovery](#backup-and-recovery)
- [Troubleshooting](#troubleshooting)
- [Security Best Practices](#security-best-practices)
- [FAQ](#faq)

---

## Getting Started

### System Requirements

**Minimum Requirements**:
- **Operating System**: Linux, macOS 10.12+, or Windows 10+
- **Memory**: 4 GB RAM
- **Disk Space**: 100 MB free space
- **Display**: 800x600 resolution or higher
- **Graphics**: OpenGL 2.0 compatible graphics card (or software renderer)

**Recommended**:
- **Memory**: 8 GB RAM or more
- **Display**: 1920x1080 resolution
- **Graphics**: Modern GPU with updated drivers

### Platform-Specific Notes

**Linux**:
- Requires X11 or Wayland display server
- Install dependencies: `sudo apt install libgl1-mesa-dev libxcursor-dev libxinerama-dev libxrandr-dev`

**macOS**:
- Xcode command-line tools required: `xcode-select --install`
- May need to allow app in Security & Privacy settings

**Windows**:
- May require graphics driver updates for OpenGL support
- See "Windows OpenGL Issues" section if encountering display problems

---

## Installation

### Option 1: Download Pre-built Binary (Recommended)

1. Visit the [Releases page](https://github.com/yourusername/passquantum/releases)
2. Download the appropriate version for your platform:
   - **Linux**: `PassQuantum-linux-amd64.tar.gz`
   - **Windows**: `PassQuantum-windows-amd64.zip`
   - **macOS**: `PassQuantum-macos-amd64.zip`
3. Extract the archive
4. Run the executable:
   - Linux/macOS: `./PassQuantum`
   - Windows: Double-click `PassQuantum.exe`

### Option 2: Build from Source

```bash
# Prerequisites: Go 1.22 or later

# Clone repository
git clone https://github.com/yourusername/passquantum.git
cd passquantum/new-passquantum

# Install dependencies
go mod tidy

# Build
go build -o passquantum ./ui

# Run
./passquantum
```

---

## First Time Setup

### Launch Application

When you first launch PassQuantum:

```
┌─────────────────────────────────────┐
│     PassQuantum                     │
│   Quantum-Proof Encryption          │
│                                     │
│  Enter Master Password:             │
│  [____________________________]     │
│                                     │
│       [  DESBLOQUEAR  ]            │
└─────────────────────────────────────┘
```

### What Happens on First Run

1. **Keypair Generation**: PassQuantum automatically generates:
   - `public.key` (Kyber768 public key, 1184 bytes)
   - `private.key` (Kyber768 private key, 2400 bytes)
   
   **Important**: These files are essential for encryption/decryption. Back them up securely!

2. **Master Password Entry**: You'll be prompted to enter a master password

3. **Vault Creation**: Your first vault is created (default name: "Default_Vault")

4. **Main Screen**: You're ready to start storing passwords

### Choosing a Strong Master Password

Your master password is the key to all your stored passwords. Choose wisely:

✅ **Good Practices**:
- At least 16 characters
- Mix of uppercase, lowercase, numbers, and symbols
- Use a passphrase (e.g., "Correct-Horse-Battery-Staple-2024!")
- Unique password not used elsewhere

❌ **Avoid**:
- Common words or phrases
- Personal information (birthdays, names)
- Short passwords (< 12 characters)
- Reused passwords from other services

---

## Using PassQuantum

### Main Application Flow

```
Login Screen
    ↓
Vault Selection
    ↓
Main Password Manager
    ↓
View/Manage Passwords
```

### Login Screen

When you launch PassQuantum:

1. Enter your master password
2. Click "DESBLOQUEAR" (Unlock)
3. If correct: proceed to vault selection
4. If incorrect: error message displayed

**Features**:
- Password field is masked for security
- Animated particle background
- Neon-styled cyberpunk aesthetic

### Vault Selection Screen

After login, you'll see all your available vaults:

```
┌─────────────────────────────────────────┐
│         Your Vaults                     │
│  ───────────────────────────────────── │
│                                         │
│  ┌───────────────────────────────────┐ │
│  │ Personal                          │ │
│  │ Location: vaults/Personal.pqdb    │ │
│  │ [Open] [Delete]                   │ │
│  └───────────────────────────────────┘ │
│                                         │
│  ┌───────────────────────────────────┐ │
│  │ Work                              │ │
│  │ Location: vaults/Work.pqdb        │ │
│  │ [Open] [Delete]                   │ │
│  └───────────────────────────────────┘ │
│                                         │
│  [+ Create New Vault] [⚙ Settings]     │
│  [🔒 Lock & Exit]                      │
└─────────────────────────────────────────┘
```

**Actions**:
- **Open**: Access a vault to view/manage passwords
- **Delete**: Remove vault (with confirmation dialog)
- **Create New Vault**: Create additional vault with custom name
- **Settings**: Access application settings
- **Lock & Exit**: Lock all vaults and close application

---

## UI Navigation

### Screen Flow Diagram

```
┌──────────────┐
│ Login Screen │
└──────┬───────┘
       │
       ▼
┌──────────────────┐      ┌──────────────┐
│ Vault Selection  │◄────►│  Settings    │
└──────┬───────────┘      └──────────────┘
       │
       ▼
┌────────────────────┐
│ Main Screen        │
│ (Password Manager) │
└──────┬─────────────┘
       │
       ▼
┌──────────────────┐
│ Passwords View   │
└──────────────────┘
```

### Navigation Controls

| Action | How To | What Happens |
|--------|--------|-------------|
| **Go Back** | Click "← Back" button | Return to previous screen |
| **Lock & Exit** | Click "🔒 Lock & Exit" | Close app, clear memory |
| **Switch Vault** | Click "← Back to Vaults" | Return to vault selection |
| **Open Settings** | Click "⚙ Settings" | Access settings panel |

---

## Managing Passwords

### Adding a Password

1. From the **Main Screen**, you'll see the password entry form:

```
┌─────────────────────────────────────────┐
│  ADD NEW PASSWORD                       │
│  ─────────────────────────────────────  │
│                                         │
│  Service Name:                          │
│  [Gmail_____________________________]   │
│                                         │
│  Username / Email:                      │
│  [user@gmail.com____________________]   │
│                                         │
│  Password:                              │
│  [••••••••••••••••••••••••••••••••]   │
│                                         │
│       [➕ SAVE PASSWORD]                │
└─────────────────────────────────────────┘
```

2. Fill in the fields:
   - **Service Name**: Website or app name (e.g., "Gmail", "GitHub", "Netflix")
   - **Username/Email**: Your username or email for that service
   - **Password**: The password to encrypt and store

3. Click "➕ SAVE PASSWORD"

4. Password is encrypted and added to your vault

5. Success message is displayed

### Viewing Passwords

1. From the **Main Screen**, click "📋 VIEW ALL"

2. All passwords are decrypted and displayed in cards:

```
┌─────────────────────────────────────────┐
│  YOUR PASSWORDS                          │
│  Total: 3 passwords                      │
│  ─────────────────────────────────────  │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ #1 - Gmail                      │   │
│  │ 👤 user@gmail.com               │   │
│  │ 🔐 ••••••••••                   │   │
│  │ [👁 Show] [📋 Copy] [🗑 Delete] │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ #2 - GitHub                     │   │
│  │ 👤 myusername                   │   │
│  │ 🔐 ••••••••••                   │   │
│  │ [👁 Show] [📋 Copy] [🗑 Delete] │   │
│  └─────────────────────────────────┘   │
│                                         │
│  [← Back]                               │
└─────────────────────────────────────────┘
```

### Password Card Actions

Each password card provides several actions:

**Show/Hide Password**:
- Click "👁 Show" to reveal the password
- Click "👁 Hide" to mask it again

**Copy to Clipboard**:
- Click "📋 Copy" to copy password to clipboard
- Password remains on clipboard for the configured timeout (default: 30 seconds)
- "Password copied!" confirmation appears

**Delete Password**:
- Click "🗑 Delete" to remove password
- Confirmation dialog appears if enabled in settings
- Password is permanently removed from vault

---

## Managing Vaults

### Creating a New Vault

1. From **Vault Selection**, click "[+ Create New Vault]"

2. Enter vault details:
   ```
   ┌─────────────────────────────────┐
   │  Create New Vault               │
   │  ───────────────────────────── │
   │                                 │
   │  Vault Name:                    │
   │  [Work_Passwords____________]   │
   │                                 │
   │  [Create] [Cancel]              │
   └─────────────────────────────────┘
   ```

3. Click "Create"

4. New vault is created with:
   - Independent master password (same as your current session)
   - Separate `.pqdb` file in `vaults/` directory
   - Empty password list

### Opening a Vault

1. From **Vault Selection**, find your desired vault

2. Click "Open" on the vault card

3. Vault is loaded and you see the **Main Screen** for that vault

4. Vault name is displayed at the top

### Switching Between Vaults

1. From **Main Screen**, click "← Back to Vaults"

2. You return to **Vault Selection**

3. Click "Open" on a different vault

4. Current vault is closed, new vault is opened

**Note**: Only one vault can be open at a time

### Deleting a Vault

1. From **Vault Selection**, click "Delete" on a vault card

2. Confirmation dialog appears:
   ```
   Are you sure you want to delete vault "Work"?
   This action cannot be undone.
   
   [Delete] [Cancel]
   ```

3. Click "Delete" to confirm

4. Vault file is permanently removed

**Warning**: Deleted vaults cannot be recovered. Make sure you have backups!

---

## Settings

Access settings by clicking "⚙ Settings" from **Vault Selection**.

### Security Tab 🔒

**Password Strength Requirements**:
- Configure minimum password complexity
- Options: Weak | Medium | Strong | Very Strong

**Master Password Management**:
- Click "Change Master Password" to update your master password
- Requires current password verification

**Session Management**:
- **Auto-lock Timeout**: Automatically lock after inactivity
  - Options: 5 min | 15 min | 30 min | 1 hour | Never
- **Clipboard Clear Timeout**: Auto-clear copied passwords
  - Options: 15 sec | 30 sec | 1 min | 5 min

**Two-Factor Authentication** (Coming Soon):
- Enable additional security layer
- Will support TOTP and advanced-auth authentication

### Vault Tab 📦

**Vault Information**:
- Current vault name
- Total number of vaults
- Vault statistics (passwords count, file size)

**Vault Maintenance**:
- **Compact Vault**: Optimize vault storage, remove deleted entries
- **Export Vault**: Create encrypted backup file
- **Import Vault**: Restore from backup file

**Last Backup**: Shows timestamp of most recent backup

### Display Tab 🎨

**Appearance**:
- **Theme**: Dark | Light | System
  - Dark: Cyberpunk neon aesthetic (default)
  - Light: Clean professional look
  - System: Follow OS theme
- **Font Size**: Small | Medium | Large
  - Adjust text size for readability

**Behavior**:
- **Show password on hover**: Reveal password when hovering (default: OFF)
- **Confirm before deleting**: Show confirmation dialog (default: ON)

### Backup Tab 💾

**Automatic Backups**:
- **Enable Automatic Backups**: Toggle on/off
- **Backup Frequency**: Daily | Weekly | Monthly

**Manual Backup**:
- **Backup Now**: Create immediate encrypted backup
- Updates "Last Backup" timestamp

**Recovery**:
- **Restore from Backup**: Select backup file to restore

**Cloud Backup** (Coming Soon):
- Encrypted cloud sync capability
- Major cloud provider support

### About Tab ℹ️

**Application Information**:
- PassQuantum logo and branding
- Version number
- Feature highlights checklist

**Resources**:
- **📖 Documentation**: Link to full documentation
- **🔄 Check for Updates**: Check for newer versions

---

## Backup and Recovery

### Why Backup?

**Critical Files to Backup**:
1. **`public.key`** & **`private.key`**: Required to decrypt passwords
2. **`vaults/*.pqdb`**: Your encrypted password vaults

**Without these files, your passwords are irrecoverable!**

### Creating Backups

#### Method 1: Manual File Backup

```bash
# Create backup directory
mkdir passquantum-backup

# Copy key files
cp public.key passquantum-backup/
cp private.key passquantum-backup/

# Copy all vaults
cp -r vaults/ passquantum-backup/

# Create archive
tar -czf passquantum-backup-$(date +%Y%m%d).tar.gz passquantum-backup/
```

#### Method 2: Using Built-in Backup Feature

1. Go to **Settings** → **Vault Tab**
2. Click "Export Vault (Encrypted)"
3. Choose destination
4. Backup file is created with timestamp

### Restoring from Backup

#### Restore Key Files

```bash
# Extract backup
tar -xzf passquantum-backup-20260313.tar.gz

# Copy keys to PassQuantum directory
cp passquantum-backup/public.key .
cp passquantum-backup/private.key .
cp -r passquantum-backup/vaults/ .
```

#### Using Built-in Restore

1. Go to **Settings** → **Vault Tab**  
2. Click "Import Vault Backup"
3. Select backup file
4. Choose "Merge" or "Replace" option
5. Vault is restored

### Backup Best Practices

✅ **Do**:
- Backup regularly (weekly or after adding important passwords)
- Store backups in multiple locations (USB drive, encrypted cloud storage)
- Test restores periodically to verify backups work
- Encrypt backup archives with additional password

❌ **Don't**:
- Store backups on the same device only
- Share backups over unsecured channels
- Backup without testing restore process
- Leave backups unencrypted on shared storage

---

## Troubleshooting

### Common Issues and Solutions

#### Application Won't Start

**Symptoms**: App crashes immediately or window doesn't open

**Solutions**:
1. **Linux**: Check display server is running
   ```bash
   echo $DISPLAY  # Should show ":0" or similar
   export DISPLAY=:0
   ./passquantum
   ```

2. **Check dependencies** (Linux):
   ```bash
   sudo apt install libgl1-mesa-dev libxcursor-dev libxinerama-dev libxrandr-dev
   ```

3. **Windows OpenGL Error**: See "Windows Graphics Issues" section below

---

#### Windows Graphics Issues

**Error Message**:
```
Fyne error: window creation error
Cause: APIUnavailable: WGL: The driver does not appear to support OpenGL
```

**Solution 1 - Update Graphics Drivers** (Recommended):
1. Identify your graphics card (Start → Device Manager → Display adapters)
2. Download latest drivers:
   - **NVIDIA**: https://www.nvidia.com/Download/index.aspx
   - **AMD**: https://www.amd.com/en/support
   - **Intel**: https://www.intel.com/content/www/us/en/download-center/home.html
3. Install drivers and restart computer

**Solution 2 - Software Rendering**:
1. Download Mesa3D from: https://fdossena.com/?p=mesa/index.frag
2. Extract `opengl32.dll` from the archive
3. Place `opengl32.dll` in the same folder as `PassQuantum.exe`
4. Run `PassQuantum.exe`

**Note**: Software rendering is slower but will work on any system.

---

#### Cannot Decrypt Passwords

**Symptoms**: "Failed to decrypt password" or "Invalid authentication tag" errors

**Possible Causes**:
1. **Wrong master password**: Re-enter correct password
2. **Missing private.key**: Restore from backup
3. **Corrupted vault file**: Restore from backup
4. **Vault file from different keypair**: Cannot decrypt without original keys

**Solutions**:
- Verify you're using the correct master password
- Check that `private.key` exists and has correct permissions
- Restore keypair and vault from backup if corrupted

---

#### Forgotten Master Password

**Bad News**: There is **no way to recover** your master password. This is by design for security.

**Your Options**:
1. **If you have a backup**: Restore keypair and vaults from before password change
2. **If no backup**: Your passwords are permanently inaccessible

**Prevention**:
- Write down your master password and store it securely
- Use a master password you can remember
- Consider using a passphrase (easier to remember than random characters)

---

#### Vault File Corrupted

**Symptoms**: "Failed to read vault" or "HMAC verification failed"

**Causes**:
- Disk error or unexpected shutdown during write
- Manual editing of `.pqdb` file
- File system corruption

**Solutions**:
1. Restore vault from backup
2. If no backup and file is partially readable, contact support
3. For future: Enable automatic backups in settings

---

#### App Freezes or Crashes

**Symptoms**: Application becomes unresponsive or closes unexpectedly

**Solutions**:
1. **Check system resources**: Ensure sufficient RAM available
2. **Update to latest version**: Bug fixes in newer releases
3. **Check for large vaults**: 1000+ passwords may cause slowness
4. **View logs** (Linux/macOS):
   ```bash
   ./passquantum 2>&1 | tee passquantum.log
   ```
5. **Report issue**: Provide crash logs and steps to reproduce

---

#### Cannot Add Password

**Symptoms**: "Failed to save password" error

**Possible Causes**:
1. Disk full
2. No write permission to `vaults/` directory
3. Vault file locked by another process

**Solutions**:
1. Check disk space: `df -h` (Linux/macOS) or File Explorer (Windows)
2. Verify permissions:
   ```bash
   ls -la vaults/
   chmod 600 vaults/*.pqdb  # Fix if needed
   ```
3. Ensure no other PassQuantum instances are running
4. Try restarting application

---

## Security Best Practices

### Password Management

✅ **Recommended Practices**:
1. **Use unique passwords** for each service
2. **Generate random passwords** (15-20 characters)
3. **Rotate passwords periodically** (every 90-180 days for critical accounts)
4. **Don't reuse passwords** across different vaults
5. **Enable 2FA** on services where available (in addition to storing passwords)

### Master Password Security

✅ **Do**:
- Choose a strong, memorable passphrase
- Use 16+ characters
- Mix character types
- Write it down and store physically in a secure location
- Never share it

❌ **Don't**:
- Use common passwords or dictionary words
- Share your master password via email or messaging
- Store it in a cloud note or text file
- Use the same master password for multiple vaults

### File Security

✅ **Protect Your Files**:
1. **Set proper permissions**:
   ```bash
   chmod 600 private.key
   chmod 600 vaults/*.pqdb
   ```

2. **Backup regularly**:
   - Weekly backups to external drive
   - Monthly backups to encrypted cloud storage

3. **Encrypt your system**:
   - Use full-disk encryption (FileVault, BitLocker, LUKS)
   - Store backups on encrypted media

4. **Secure your device**:
   - Keep OS and software updated
   - Use antivirus/anti-malware
   - Enable firewall
   - Lock screen when away

### Physical Security

✅ **Protect Physical Access**:
- Lock your computer when leaving
- Don't leave PassQuantum open and unattended
- Store backup USB drives in secure location (safe, bank deposit box)
- Shred or securely erase old backup media

### Network Security

✅ **Stay Safe Online**:
- PassQuantum works offline (no network communication)
- Be cautious of phishing attempts asking for your master password
- Verify official PassQuantum releases before downloading
- Don't install PassQuantum from untrusted sources

---

## FAQ

### General Questions

**Q: Is PassQuantum really quantum-safe?**  
A: Yes. PassQuantum uses Kyber768, a post-quantum key encapsulation mechanism standardized by NIST. It's designed to resist attacks from both classical and quantum computers.

**Q: Can I use PassQuantum on multiple computers?**  
A: Yes, but you must copy your `public.key`, `private.key`, and `vaults/` directory to each computer. Cloud sync is planned for a future release.

**Q: Does PassQuantum connect to the internet?**  
A: No. PassQuantum is completely offline and stores all data locally. No telemetry, no cloud sync (yet), no network connections.

**Q: Is my data safe if someone steals my computer?**  
A: If your disk is encrypted and your computer is locked, yes. Without your master password, your vaults cannot be decrypted. However, enable full-disk encryption for maximum protection.

**Q: What happens if I forget my master password?**  
A: Unfortunately, there is no recovery mechanism. Your passwords are permanently inaccessible. This is a security feature, not a bug.

---

### Technical Questions

**Q: What encryption algorithms does PassQuantum use?**  
A: 
- **Key Derivation**: Argon2id (64MB memory, GPU-resistant)
- **Vault Encryption**: AES-256-GCM (authenticated encryption)
- **Integrity**: HMAC-SHA256
- **Post-Quantum**: Kyber768 (NIST PQC standard)

**Q: Where are my passwords stored?**  
A: In encrypted `.pqdb` files in the `vaults/` directory. Each vault is a separate file.

**Q: Can I move PassQuantum to a different folder?**  
A: Yes. Copy the entire directory including `public.key`, `private.key`, and `vaults/` folder. All paths are relative.

**Q: How do I uninstall PassQuantum?**  
A: Delete the application folder. To completely remove all data, also delete `vaults/`, `public.key`, and `private.key`. **Warning**: This is irreversible!

**Q: Is PassQuantum open source?**  
A: Info about licensing should be in the LICENSE file and README. The code is available for audit and contribution.

---

### Features Questions

**Q: Can I import passwords from another password manager?**  
A: Not yet, but this feature is planned. For now, you must manually add passwords.

**Q: Does PassQuantum have a browser extension?**  
A: Not yet. Browser integration is on the roadmap.

**Q: Can I share passwords with other users?**  
A: Not yet. Secure sharing is planned for a future release.

**Q: Does PassQuantum have a mobile app?**  
A: Not yet. iOS and Android apps are in the roadmap.

**Q: Can I generate random passwords in PassQuantum?**  
A: Not yet, but a password generator is planned.

---

### Troubleshooting Questions

**Q: Why is vault unlock so slow (~2 seconds)?**  
A: This is intentional. Argon2id key derivation is designed to be slow to prevent brute-force attacks. The 64MB memory requirement makes password guessing expensive.

**Q: Why does my password list take a while to display?**  
A: Each password must be decrypted individually. With 100+ passwords, this can take 1-2 seconds. This is expected behavior.

**Q: Can I speed up vault unlocking?**  
A: You can reduce Argon2id parameters in the code, but this significantly weakens security. Not recommended.

**Q: My vault file is growing large. How do I compact it?**  
A: Go to **Settings** → **Vault Tab** → **Compact Vault**. This removes deleted entries and optimizes storage.

---

## Getting Help

### Documentation

- **README.md**: Project overview and quick start
- **ARCHITECTURE.md**: Technical architecture and API reference  
- **USER_EXPERIENCE.md**: This document - complete user guide

### Community Support

- **GitHub Issues**: Report bugs and request features
- **GitHub Discussions**: Ask questions and share tips
- **Email Support**: security@passquantum.example.com (for security issues only)

### Reporting Bugs

When reporting a bug, please include:
1. **Operating System** and version
2. **PassQuantum version**
3. **Steps to reproduce** the issue
4. **Expected behavior** vs **actual behavior**
5. **Error messages** (if any)
6. **Screenshots** (if applicable)

**Do NOT include**:
- Your master password
- Your private key
- Your vault files
- Decrypted passwords

---

**PassQuantum** - Your passwords, quantum-safe and secure. 🔐✨

