import { Lock, ScanFace, Database } from 'lucide-react'
import { useLang } from '../LangContext'
import t from '../translations'
import { TECH, ARGON2 } from '../config'

// Static config (icons, colours) stays in component — only text comes from translations
const FEATURE_STYLE = [
  {
    Icon:     Lock,
    specs: [
      { k: 'Payload cipher',    v: TECH.aes },
      { k: 'Key encapsulation', v: TECH.mlkem },
      { k: 'Digital signature', v: TECH.mldsa },
      { k: 'KDF',               v: `Argon2id · ${ARGON2}` },
    ],
    topBar:   'from-accent to-accent/20',
    iconRing: 'border-accent/25 bg-accent/8',
    iconClr:  'text-accent',
    hoverBg:  'hover:bg-accent/3',
  },
  {
    Icon:     ScanFace,
    specs: [
      { k: 'Framework',     v: 'MediaPipe Face Mesh' },
      { k: 'Landmarks',     v: '468 pts' },
      { k: 'Liveness gate', v: 'Blink detection' },
      { k: 'Lock timeout',  v: '5 s' },
    ],
    topBar:   'from-sky-400 to-sky-400/20',
    iconRing: 'border-sky-400/25 bg-sky-400/8',
    iconClr:  'text-sky-400',
    hoverBg:  'hover:bg-sky-400/3',
  },
  {
    Icon:     Database,
    specs: [
      { k: 'Storage format', v: '.pqdb (AES-GCM)' },
      { k: 'Integrity',      v: 'HMAC-SHA256' },
      { k: 'File perms',     v: 'POSIX 0600' },
      { k: 'Network access', v: 'Zero · Offline' },
    ],
    topBar:   'from-violet-400 to-violet-400/20',
    iconRing: 'border-violet-400/25 bg-violet-400/8',
    iconClr:  'text-violet-400',
    hoverBg:  'hover:bg-violet-400/3',
  },
]

export default function Features() {
  const { lang } = useLang()
  const tr = t[lang].features

  return (
    <section id="features" className="py-24 px-4 sm:px-6 lg:px-8">
      <div className="max-w-7xl mx-auto">

        {/* Header */}
        <div className="text-center mb-16">
          <p className="text-xs font-mono text-accent tracking-[0.2em] uppercase mb-3">
            {tr.tag}
          </p>
          <h2 className="text-3xl sm:text-4xl font-bold text-white mb-4">
            {tr.title}
          </h2>
          <p className="text-slate-400 max-w-2xl mx-auto text-sm sm:text-base leading-relaxed">
            {tr.sub}
          </p>
        </div>

        {/* Grid */}
        <div className="grid md:grid-cols-3 gap-6">
          {FEATURE_STYLE.map(({ Icon, specs, topBar, iconRing, iconClr, hoverBg }, i) => {
            const { title, sub, desc } = tr.items[i]
            return (
              <div
                key={i}
                className={`group relative rounded-2xl border border-wire bg-surface p-6 transition-all duration-300 hover:-translate-y-1.5 overflow-hidden ${hoverBg}`}
              >
                <div className={`absolute inset-x-0 top-0 h-[2px] bg-gradient-to-r ${topBar}`} />

                <div className="relative">
                  <div className={`inline-flex p-3 rounded-xl border ${iconRing} ${iconClr} mb-5`}>
                    <Icon className="w-6 h-6" />
                  </div>

                  <h3 className="text-base font-bold text-white mb-0.5">{title}</h3>
                  <p className="text-[11px] font-mono text-slate-500 mb-4">{sub}</p>
                  <p className="text-slate-400 text-sm leading-relaxed mb-6">{desc}</p>

                  <div className="space-y-2.5 border-t border-wire pt-4">
                    {specs.map(({ k, v }) => (
                      <div key={k} className="flex items-center justify-between gap-3">
                        <span className="text-[11px] font-mono text-slate-500 shrink-0">{k}</span>
                        <span className="text-[11px] font-mono text-slate-300 bg-bg px-2 py-0.5 rounded border border-wire text-right">
                          {v}
                        </span>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            )
          })}
        </div>
      </div>
    </section>
  )
}
