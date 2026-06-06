import { useEffect, useRef, useState } from 'react'
import { Download, ChevronDown, Terminal, Monitor } from 'lucide-react'
import { useLang } from '../LangContext'
import t from '../translations'
import { useLatestRelease } from '../hooks/useLatestRelease'

const BASE        = 'https://github.com/ESH2007/PassQuantum/releases/latest/download'
const ASSET_NAMES = {
  linux:        'PassQuantum-linux-amd64.zip',
  windows:      'PassQuantum-windows-amd64.zip',
  'darwin-dmg': 'PassQuantum-macos-arm64-dmg.zip',
  'darwin-app': 'PassQuantum-macos-arm64-app.zip',
}
const DOWNLOADS = Object.fromEntries(
  Object.entries(ASSET_NAMES).map(([os, file]) => [os, `${BASE}/${file}`])
)

// Returns the platform key used for DOWNLOADS / translations.options.
// macOS is detected explicitly so Mac users are never silently handed a
// Linux binary; Mac users default to the .dmg installer (the .app remains
// selectable in the dropdown). Genuinely unknown platforms fall back to
// 'linux' (our most common desktop target) — a deliberate default, not a
// misclassified Mac.
function detectOS() {
  const ua = navigator.userAgent || ''
  if (/Win/i.test(ua)) return 'windows'
  if (/Mac/i.test(ua)) return 'darwin-dmg'
  return 'linux'
}

// Per-platform icon. A module-scope component (rather than a render-time
// variable) so it isn't re-created on every render.
function OsIcon({ os, className }) {
  return os === 'linux'
    ? <Terminal className={className} />
    : <Monitor className={className} />
}

export default function DownloadButton() {
  const { lang } = useLang()
  const tr = t[lang].hero.download

  const release      = useLatestRelease()
  const versionLabel = release?.tag ?? tr.version

  const detectedOS = detectOS()
  const [selected, setSelected] = useState(detectedOS)
  const [open, setOpen]         = useState(false)
  const containerRef            = useRef(null)

  // Close dropdown when clicking outside
  useEffect(() => {
    function handleOutside(e) {
      if (containerRef.current && !containerRef.current.contains(e.target)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', handleOutside)
    return () => document.removeEventListener('mousedown', handleOutside)
  }, [])

  const href      = DOWNLOADS[selected]
  const available = Boolean(href)

  return (
    <div ref={containerRef} className="relative inline-flex">
      {/* Main area: active download link, or a disabled "coming soon" state
          when the selected platform has no build yet (macOS). */}
      {available ? (
        <a
          href={href}
          download
          className="flex items-center gap-2 px-5 py-3.5 rounded-l-xl border border-r-0 border-wire text-slate-300 text-sm hover:border-accent/40 hover:text-accent transition-all"
        >
          <Download className="w-4 h-4 shrink-0" />
          <span>
            {tr.label}{' '}
            <span className="font-semibold">{tr.options[selected]}</span>
          </span>
          <span className="text-xs font-mono text-slate-500 hidden sm:inline">{versionLabel}</span>
          <OsIcon os={selected} className="w-3.5 h-3.5 shrink-0 text-slate-500" />
        </a>
      ) : (
        <div
          aria-disabled="true"
          className="flex items-center gap-2 px-5 py-3.5 rounded-l-xl border border-r-0 border-wire text-slate-500 text-sm cursor-not-allowed select-none"
        >
          <Download className="w-4 h-4 shrink-0 opacity-50" />
          <span>{tr.unavailable}</span>
          <OsIcon os={selected} className="w-3.5 h-3.5 shrink-0 text-slate-600" />
        </div>
      )}

      {/* Chevron toggle */}
      <button
        type="button"
        onClick={() => setOpen(v => !v)}
        aria-label="Select platform"
        className="flex items-center justify-center px-2.5 rounded-r-xl border border-wire text-slate-500 hover:border-accent/40 hover:text-accent transition-all"
      >
        <ChevronDown className={`w-4 h-4 transition-transform ${open ? 'rotate-180' : ''}`} />
      </button>

      {/* Dropdown */}
      {open && (
        <div className="absolute bottom-full mb-1.5 left-0 z-50 min-w-[13rem] rounded-xl border border-wire bg-surface/95 backdrop-blur-sm shadow-xl overflow-hidden">
          {Object.entries(tr.options).map(([key, label]) => {
            return (
              <button
                key={key}
                type="button"
                onClick={() => { setSelected(key); setOpen(false) }}
                className={`w-full flex items-center gap-2.5 px-4 py-2.5 text-sm text-left transition-colors
                  ${selected === key
                    ? 'text-accent bg-accent/8'
                    : 'text-slate-400 hover:text-slate-200 hover:bg-white/4'
                  }`}
              >
                <OsIcon os={key} className="w-3.5 h-3.5" />
                {label}
                {!(key in DOWNLOADS) && (
                  <span className="ml-auto text-[10px] font-mono text-slate-500 uppercase tracking-wide">soon</span>
                )}
                {key === detectedOS && key in DOWNLOADS && (
                  <span className="ml-auto text-[10px] font-mono text-accent/60 uppercase tracking-wide">auto</span>
                )}
              </button>
            )
          })}
        </div>
      )}
    </div>
  )
}
