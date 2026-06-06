import { TriangleAlert, CheckCheck } from 'lucide-react'
import Mark from './Mark'
import RichText from './RichText'
import { useLang } from '../LangContext'
import t from '../translations'
import { AUTHOR_URL, AUTHOR_HANDLE } from '../config'
import { useLatestRelease } from '../hooks/useLatestRelease'

const codeChip = text => (
  <code className="font-mono text-xs bg-bg px-1.5 py-0.5 rounded border border-wire">{text}</code>
)

export default function Footer() {
  const { lang } = useLang()
  const tr = t[lang].footer
  const release    = useLatestRelease()
  const versionTag = release?.tag ?? 'v1.0-beta'

  return (
    <footer id="transparency" className="border-t border-wire">

      {/* Transparency banner */}
      <div className="bg-surface/60 border-b border-wire">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">

          <div className="flex items-start gap-3 mb-8">
            <TriangleAlert className="w-5 h-5 text-amber-400 mt-0.5 shrink-0" />
            <div>
              <p className="text-[11px] font-mono text-amber-400 uppercase tracking-widest mb-3">
                {tr.transparencyTag}
              </p>
              <p className="text-slate-300 text-sm sm:text-base leading-relaxed max-w-3xl">
                <RichText
                  text={tr.transparency}
                  parts={{
                    indie: <span className="text-white font-semibold">{tr.transparencyIndie}</span>,
                    nativeCrypto: <span className="text-accent font-semibold">{tr.transparencyNativeCrypto}</span>,
                    lib1: codeChip('golang.org/x/crypto'),
                    lib2: codeChip('filippo.io/mlkem768'),
                    notAudited: <span className="text-amber-300 font-medium">{tr.transparencyNotAudited}</span>,
                  }}
                />
              </p>
            </div>
          </div>

          {/* Truth cards */}
          <div className="grid sm:grid-cols-3 gap-4">
            {tr.cards.map(({ label, detail }, i) => (
              <div key={i} className="rounded-xl border border-wire bg-bg p-4">
                <div className="flex items-center gap-2 mb-2">
                  <CheckCheck className="w-4 h-4 text-accent shrink-0" />
                  <span className="text-accent font-mono text-sm font-bold">{label}</span>
                </div>
                <p className="text-slate-500 text-xs leading-relaxed">{detail}</p>
              </div>
            ))}
          </div>

        </div>
      </div>

      {/* Footer bar */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="flex flex-col sm:flex-row items-center justify-between gap-6">

          {/* Logo */}
          <div className="flex items-center gap-2">
            <Mark className="w-5 h-5" />
            <span className="font-bold font-mono text-sm text-slate-300">
              Pass<span className="text-accent">Quantum</span>
            </span>
            <span className="text-slate-600 text-[10px] font-mono border border-wire rounded px-1.5 py-0.5">
              {versionTag}
            </span>
          </div>

          {/* Links */}
          <nav className="flex items-center gap-5">
            {tr.navLinks.map(({ href, label, external }) => (
              <a
                key={label}
                href={href}
                {...(external ? { target: '_blank', rel: 'noopener noreferrer' } : {})}
                className="text-xs font-mono text-slate-500 hover:text-accent transition-colors"
              >
                {label}
              </a>
            ))}
          </nav>

          {/* Made by */}
          <div className="flex items-center gap-1 text-xs text-slate-600 font-mono">
            {tr.madeWith}{' '}{tr.by}{' '}
            <a
              href={AUTHOR_URL}
              target="_blank"
              rel="noopener noreferrer"
              className="text-slate-400 hover:text-accent transition-colors ml-0.5"
            >
              {AUTHOR_HANDLE}
            </a>
          </div>

        </div>

        <p className="text-center text-slate-700 text-[10px] font-mono mt-6">
          {tr.disclaimer}
        </p>
      </div>

    </footer>
  )
}
