// Shared, non-translated constants (product facts and links).

export const REPO_URL = 'https://github.com/ESH2007/PassQuantum'
export const AUTHOR_URL = 'https://github.com/ESH2007'
export const AUTHOR_HANDLE = '@ESH2007'
export const VERSION = 'v1.0-beta'

export const DOWNLOADS = {
  linux: `${REPO_URL}/releases/latest/download/PassQuantum-linux-amd64.zip`,
  windows: `${REPO_URL}/releases/latest/download/PassQuantum-windows-amd64.zip`,
  'darwin-dmg': `${REPO_URL}/releases/latest/download/PassQuantum-macos-arm64-dmg.zip`,
  'darwin-app': `${REPO_URL}/releases/latest/download/PassQuantum-macos-arm64-app.zip`,
}

// Cryptography labels surfaced in the copy. Not translated — they are names.
export const TECH = {
  aes: 'AES-256-GCM',
  mlkem: 'ML-KEM-768 (Kyber768)',
  mldsa: 'ML-DSA (Dilithium)',
}

// Canonical Argon2id parameter string, shared by Features and Architecture
// so the formatting stays identical in both places.
export const ARGON2 = '64 MB · iter 1 · 4 threads · salt 16B'
