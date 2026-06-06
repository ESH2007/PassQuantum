# PassQuantum — Web

Marketing landing page for **[PassQuantum](https://github.com/ESH2007/PassQuantum)**, a
local-first, post-quantum desktop password manager.

The site presents the project's encryption stack (AES-256-GCM + ML-KEM/Kyber768 +
ML-DSA/Dilithium, Argon2id KDF), its passive biometric lock (MediaPipe Face Mesh),
and a transparency section about the project's scope and limitations. It is
bilingual (Spanish / English) and ships as a single static page.

## Stack

- React 19
- Vite 8
- Tailwind CSS v4 (`@theme` design tokens)
- lucide-react icons
- Deployed to GitHub Pages via `gh-pages`

## Development

```bash
npm install
npm run dev      # start the dev server
npm run build    # production build to dist/
npm run preview  # preview the production build
npm run lint     # eslint
```

## Deploy

```bash
npm run deploy   # builds, then publishes dist/ to the gh-pages branch
```

Published at <https://ESH2007.github.io/PassQuantum-web>.

## Project layout

```
src/
  components/      Navbar, Hero, Features, Architecture, Footer, FaceMesh, …
  translations.js  All UI copy, keyed by language
  LangContext.jsx  Language provider + useLang() hook
  config.js        Shared constants (repo URL, version, download links)
  index.css        Tailwind import, design tokens, animations
```

## Note

This is the marketing site only. The application source lives in the
[PassQuantum](https://github.com/ESH2007/PassQuantum) repository.
