# PassQuantum Go - Compilation and Execution Guide

## Prerequisites
You need Go installed on your system. Download from [golang.org](https://golang.org/dl)

To verify Go is installed:
```bash
go version
```

## Project Structure
```
new-passquantum/
├── main.go              # Main application (already configured)
├── passwords.txt        # Password storage (created on first run)
└── go.mod              # Module file (optional, needed for dependencies)
```

## Compilation Methods

### Method 1: Direct Compilation (Recommended for Distribution)
Compile the Go program into an executable binary:

```bash
cd /home/lenovo/dev/PassQuantum/new-passquantum
go build -o passquantum main.go
```

This creates an executable named `passquantum` in the current directory.

**Run the compiled binary:**
```bash
./passquantum
```

### Method 2: Run Directly (Development Mode)
Run without creating a binary:
```bash
cd /home/lenovo/dev/PassQuantum/new-passquantum
go run main.go
```

### Method 3: Install Globally
Install the binary to your system PATH:
```bash
go install ./passquantum
```

This allows you to run `passquantum` from anywhere in your terminal.

## Compilation for Different Operating Systems

### Compile for Windows (from Linux/Mac):
```bash
cd /home/lenovo/dev/PassQuantum/new-passquantum
GOOS=windows GOARCH=amd64 go build -o passquantum.exe main.go
```

### Compile for macOS (from Linux):
```bash
GOOS=darwin GOARCH=amd64 go build -o passquantum main.go
```

### Compile for Linux 32-bit:
```bash
GOARCH=386 go build -o passquantum main.go
```

## Usage

After compilation, run the program:
```bash
./passquantum
```

### Menu Options:
- **Press Enter** - Add a new password
- **Type "2", "see passwords", "See Passwords", or "seePass"** - View all passwords (decrypted)
- **Type "e", "exit", or "3"** - Exit the program

## Key Features

### Encryption
- Uses **AES-256-GCM** symmetric encryption
- Generates a random 64-byte secret key for each password
- Uses a random nonce for authenticated encryption
- Stores encrypted passwords in `passwords.txt` using base64 encoding

### File Format
Passwords are stored in `passwords.txt` with the format:
```
[base64_encrypted_password, base64_nonce, base64_secret_key], 
```

## Troubleshooting

### Issue: "go: command not found"
- Install Go from [golang.org](https://golang.org/dl)
- Verify installation: `go version`

### Issue: "Cannot read passwords.txt"
- This is normal on first run. The file is created when you add your first password.

### Issue: "Failed to decrypt password"
- The passwords.txt file may be corrupted. Try backing it up and starting fresh.

## Comparison: Python vs Go

| Feature | Python | Go |
|---------|--------|-----|
| Encryption Library | quantcrypt.cipher (Krypton) | crypto/aes (AES-256-GCM) |
| Secret Key Size | 64 bytes | 64 bytes (32 used for AES) |
| Verification | Custom verif_dp | GCM nonce (12 bytes) |
| Runtime | Needs Python interpreter | Single binary executable |
| Performance | Slower | Much faster (compiled) |
| File Size | All source files | Single ~7-8 MB binary |
| Dependencies | External quantcrypt library | Go standard library (built-in) |

## Performance Benefits of Go

1. **Single Binary** - No interpreter or runtime needed
2. **Cross-platform** - Easy compilation for any OS
3. **Faster Execution** - Compiled code runs significantly faster
4. **No Dependencies** - Uses Go's standard library (all built-in)
5. **Smaller Footprint** - Binary is self-contained

## Build Optimization

For smaller binary size:
```bash
go build -ldflags="-s -w" -o passquantum main.go
```

This strips debugging symbols (reduces size by ~40-50%).

For maximum optimization:
```bash
CGO_ENABLED=0 go build -ldflags="-s -w" -o passquantum main.go
```

This disables cgo and strips symbols for the smallest binary.

## Development Tips

### Enable module support (optional):
```bash
go mod init passquantum
```

### Format your code:
```bash
go fmt main.go
```

### Check for errors:
```bash
go vet main.go
```

### Run tests (if added):
```bash
go test
```
