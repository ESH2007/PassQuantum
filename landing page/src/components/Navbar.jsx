import { useState, useEffect } from 'react'
import { Menu, X } from 'lucide-react'
import GithubIcon from './GithubIcon'
import Mark from './Mark'
import { useLang } from '../LangContext'
import t from '../translations'
import { REPO_URL } from '../config'
import { useLatestRelease } from '../hooks/useLatestRelease'

export default function Navbar() {
  const [scrolled, setScrolled] = useState(false)
  const [menuOpen, setMenuOpen] = useState(false)
  const { lang, setLang } = useLang()
  const tr = t[lang].nav
  const release    = useLatestRelease()
  const versionTag = release?.tag ?? 'v1.0-beta'

  useEffect(() => {
    const handler = () => setScrolled(window.scrollY > 24)
    window.addEventListener('scroll', handler, { passive: true })
    return () => window.removeEventListener('scroll', handler)
  }, [])

  return (
    <header
      className={`fixed top-0 inset-x-0 z-50 transition-all duration-300 ${
        scrolled
          ? 'bg-bg/95 backdrop-blur-md border-b border-wire'
          : 'bg-transparent'
      }`}
    >
      <nav className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">

        {/* Logo */}
        <a href="#" className="flex items-center gap-2 group">
          <Mark className="w-8 h-8" />
          <span className="font-bold text-white tracking-tight select-none">
            Pass<span className="text-accent">Quantum</span>
          </span>
          <span className="hidden sm:inline text-[10px] font-mono text-accent/60 border border-accent/20 rounded px-1.5 py-0.5">
            {versionTag}
          </span>
        </a>

        {/* Desktop links */}
        <div className="hidden md:flex items-center gap-8">
          {tr.links.map(l => (
            <a
              key={l.href}
              href={l.href}
              className="text-sm text-slate-400 hover:text-accent transition-colors font-mono tracking-wide"
            >
              {l.label}
            </a>
          ))}
        </div>

        {/* Desktop right side: lang toggle + GitHub */}
        <div className="hidden md:flex items-center gap-3">
          {/* ES | EN pill */}
          <div className="flex items-center rounded-lg border border-wire overflow-hidden font-mono text-xs">
            {['es', 'en'].map(l => (
              <button
                key={l}
                onClick={() => setLang(l)}
                className={`px-2.5 py-1.5 uppercase tracking-widest transition-colors ${
                  lang === l
                    ? 'bg-accent/15 text-accent'
                    : 'text-slate-500 hover:text-slate-300'
                }`}
              >
                {l}
              </button>
            ))}
          </div>

          <a
            href={REPO_URL}
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-2 px-4 py-2 rounded-lg border border-accent/30 text-accent text-sm font-mono hover:bg-accent/10 transition-all"
          >
            <GithubIcon className="w-4 h-4" />
            GitHub
          </a>
        </div>

        {/* Mobile toggle */}
        <button
          className="md:hidden text-slate-400 hover:text-white transition-colors"
          onClick={() => setMenuOpen(v => !v)}
          aria-label="Toggle menu"
        >
          {menuOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
        </button>
      </nav>

      {/* Mobile drawer */}
      {menuOpen && (
        <div className="md:hidden bg-surface/98 border-b border-wire px-4 pb-5">
          {tr.links.map(l => (
            <a
              key={l.href}
              href={l.href}
              onClick={() => setMenuOpen(false)}
              className="block py-3 text-slate-400 hover:text-accent transition-colors font-mono text-sm border-b border-wire/60"
            >
              {l.label}
            </a>
          ))}

          {/* Mobile lang toggle */}
          <div className="flex items-center gap-0 mt-4 mb-2 rounded-lg border border-wire overflow-hidden w-fit font-mono text-xs">
            {['es', 'en'].map(l => (
              <button
                key={l}
                onClick={() => setLang(l)}
                className={`px-4 py-2 uppercase tracking-widest transition-colors ${
                  lang === l
                    ? 'bg-accent/15 text-accent'
                    : 'text-slate-500 hover:text-slate-300'
                }`}
              >
                {l}
              </button>
            ))}
          </div>

          <a
            href={REPO_URL}
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center justify-center gap-2 px-4 py-3 rounded-xl border border-accent/30 text-accent text-sm font-mono hover:bg-accent/10 transition-all"
          >
            <GithubIcon className="w-4 h-4" />
            GitHub
          </a>
        </div>
      )}
    </header>
  )
}
