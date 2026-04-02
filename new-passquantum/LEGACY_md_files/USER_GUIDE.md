# PassQuantum GUI - User Guide

## Getting Started

### First Run
When you launch PassQuantum for the first time:
1. The application checks for existing Kyber768 keypair files
2. If none exist, a new keypair is automatically generated and saved
3. Your **public key** is saved to `public.key`
4. Your **private key** is saved to `private.key` (keep this safe!)

### Main Window

The PassQuantum window displays:
- **Application Title**: "PassQuantum - Post-Quantum Safe Password Manager"
- **Password Input**: A masked text field to enter new passwords
- **Buttons**:
  - **Add Password**: Encrypts and saves the password
  - **View Passwords**: Displays all saved passwords
  - **Exit**: Closes the application

## Usage

### Adding a Password

1. Type a password in the "Enter a new password" field
2. Click "Add Password"
3. The application will:
   - Encrypt the password using Kyber768 + AES-256-GCM
   - Save it to `passwords.txt`
   - Display a success message
   - Clear the input field
4. Your password is now stored securely

### Viewing Passwords

1. Click "View Passwords"
2. A new window opens showing all your saved passwords
3. Each password is decrypted using your private key
4. Passwords are displayed in a numbered list
5. Close the window to return to the main screen

### Exiting

- Click "Exit" to close the application
- Your keypair and passwords remain saved on disk

## Storage Details

### Files Created

After first run, PassQuantum creates these files in the current directory:

| File | Purpose | Security |
|------|---------|----------|
| `public.key` | Public key for encryption | Can be shared |
| `private.key` | Private key for decryption | **KEEP SECRET!** |
| `passwords.txt` | Encrypted password database | Useless without private key |

### Backup Recommendations

1. **Backup `public.key` and `private.key` together**
   - These files are required to recover all passwords
   - Store in a secure location (USB drive, encrypted cloud storage)

2. **Do NOT backup only `passwords.txt`**
   - Without the private key, the file cannot be decrypted

3. **Protect `private.key` carefully**
   - Anyone with this file can decrypt all your passwords
   - Consider encrypting your home directory

## Error Messages

| Error | Cause | Solution |
|-------|-------|----------|
| "password cannot be empty" | You clicked Add with blank field | Type a password first |
| "encapsulation failed" | Crypto error (rare) | Restart the app |
| "encryption failed" | AES error (rare) | Restart the app |
| "failed to save password" | Disk write error | Check disk space & permissions |
| "failed to read passwords" | File corrupted or permission denied | Check `passwords.txt` |
| "No passwords stored yet" | Database is empty | Add a password first |

## Security Considerations

### Strengths
- ✅ Post-quantum cryptography (Kyber768)
- ✅ AES-256 symmetric encryption
- ✅ Secure random nonce generation
- ✅ No hardcoded keys

### Current Limitations
- ⚠️ No master password protection
- ⚠️ Private key stored unencrypted on disk
- ⚠️ Decrypted passwords visible in memory
- ⚠️ No authentication/authorization

### Recommendations
1. Use encrypted filesystem
2. Restrict file permissions: `chmod 600 private.key`
3. Don't share `private.key` with untrusted parties
4. Keep backups in secure locations
5. Consider adding a master password in future versions

## Technical Details

### Encryption Algorithm
- **Key Encapsulation**: Kyber768 (post-quantum safe)
  - Generates random shared secret
  - Client has no control over randomness (deterministic security)
  
- **Symmetric Encryption**: AES-256-GCM
  - 256-bit key derived from Kyber's shared secret
  - Random 96-bit nonce per password
  - Authenticated encryption (prevents tampering)

### Why Kyber768?
- Resistant to future quantum computers
- NIST-standardized (2022)
- Efficient key sizes and computation
- Proven to be secure even against quantum attacks

### File Format
Each line in `passwords.txt` contains one encrypted password:
```
[base64(kyber_ciphertext), base64(nonce), base64(aes_ciphertext)], \n
```

The base64 encoding allows storage of binary data in text files.

## Troubleshooting

### Application won't start
- Check that Fyne dependencies are installed
- Ensure you have X11 or Wayland display server
- Try: `./passquantum-app -help`

### Forgot to save files before moving directory
- Keypair files (`*.key`) and `passwords.txt` are tied to the working directory
- Copy all files together when moving

### Lost private key
- ❌ Unfortunately, encrypted passwords cannot be recovered
- This is a security feature (no backdoor)
- Always maintain backups of `private.key`

### Corrupted `passwords.txt`
- The application skips malformed entries
- Try removing `passwords.txt` and recreating it
- Or edit the file to fix the problematic line

## Advanced Usage

### Running Multiple Instances
- Each instance uses the same keypair files
- Changes from one instance won't be visible to another until restart
- Recommended: Only run one instance at a time

### Command-Line Building (Development)
```bash
cd new-passquantum
go mod tidy
go build -o passquantum-app ./ui
./passquantum-app
```

### Modifying Source Code
The modular architecture makes it easy to:
- Change encryption algorithm (modify `core/crypto/`)
- Change storage format (modify `core/storage/`)
- Add new UI features (modify `ui/main.go`)

See `ARCHITECTURE.md` for detailed package documentation.
