# extension/

The PassQuantum browser extension (Manifest V3, Chrome/Edge/Brave + Firefox via
`browser_specific_settings`). It auto-saves and autofills credentials by talking
to the desktop app's localhost server in [`internal/browser`](../internal/browser/README.md)
at `http://127.0.0.1:8765`.

The extension never holds the master password or any encryption keys: it only
asks the unlocked desktop app for domain-matched credentials and sends new
ones to be saved. All cryptography stays in the Go app.

## Files

| File | Description |
|---|---|
| `manifest.json` | MV3 manifest. Notable: `host_permissions` is limited to `http://127.0.0.1:8765/*`, content scripts run on `<all_urls>` at `document_idle`. |
| `background.js` | Service worker: holds the paired secret, talks to the local server, and brokers messages between the content script and popup. |
| `content.js` | Injected into pages: detects login forms, fills credentials, and offers to save on submit. |
| `popup.html` / `popup.css` / `popup.js` | Toolbar popup: pairing UI (enter the token shown by the desktop app), status, and per-site actions. |
| `browser-polyfill.min.js` | Mozilla `webextension-polyfill` so the same code runs on Chromium and Firefox. |
| `icons/` | Extension icons (16/48/128 px). |

## Pairing & usage

1. In the desktop app, open the extension/pairing dialog to display a one-time token.
2. Load this folder as an unpacked extension (`chrome://extensions` → *Load
   unpacked*, or `about:debugging` in Firefox) and enter the token in the popup.
3. Once paired and with a vault unlocked, the extension autofills matching logins
   and offers to save new ones. Per-site "never save" choices are honored by the
   server.

## Packaging

The repo root contains prebuilt artifacts (`extension.zip`, `extension.crx`) and
the signing key (`extension.pem`); the source of truth is this directory.
