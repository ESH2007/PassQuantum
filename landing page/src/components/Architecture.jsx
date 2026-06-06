import { Cpu, Lock, Key, PenTool, Fingerprint, HardDrive } from 'lucide-react'
import { useLang } from '../LangContext'
import t from '../translations'
import { ARGON2 } from '../config'

// Static config: icons, tech values, details, colours
const STACK_STYLE = [
  { layer: '01', Icon: Cpu,         tech: 'Go + Fyne',     detail: 'Cross-platform · CGO · OpenGL',        clr: 'text-sky-400',    brd: 'border-sky-500/20',    bg: 'bg-sky-500/6'    },
  { layer: '02', Icon: Lock,        tech: 'AES-256-GCM',   detail: 'HMAC-SHA256 · nonce prefix',            clr: 'text-accent',  brd: 'border-accent/20',  bg: 'bg-accent/6'  },
  { layer: '03', Icon: Key,         tech: 'ML-KEM-768',    detail: 'NIST FIPS 203 · Kyber768',              clr: 'text-amber-400',  brd: 'border-amber-500/20',  bg: 'bg-amber-500/6'  },
  { layer: '04', Icon: PenTool,     tech: 'ML-DSA',        detail: 'NIST FIPS 204 · Dilithium',             clr: 'text-rose-400',   brd: 'border-rose-500/20',   bg: 'bg-rose-500/6'   },
  { layer: '05', Icon: Fingerprint, tech: 'Argon2id',      detail: ARGON2,                                  clr: 'text-violet-400', brd: 'border-violet-500/20', bg: 'bg-violet-500/6' },
  { layer: '06', Icon: HardDrive,   tech: '.pqdb + POSIX', detail: 'perms 0600 · private.key 0600',          clr: 'text-cyan-400',   brd: 'border-cyan-500/20',   bg: 'bg-cyan-500/6'   },
]

const FLOW = [
  ['Master Password', 'password'],
  ['→ Argon2id KDF',  'arrow'],
  ['→ Vault Keys',    'arrow'],
  ['→ AES-256-GCM',   'arrow'],
  ['→ ML-KEM Encap.', 'arrow'],
  ['→ Shared Secret', 'arrow'],
  ['→ AES-256-GCM',   'arrow'],
  ['→ .pqdb',         'arrow'],
]

export default function Architecture() {
  const { lang } = useLang()
  const tr = t[lang].architecture

  return (
    <section id="architecture" className="py-24 px-4 sm:px-6 lg:px-8 bg-surface/40">
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

        {/* Desktop table */}
        <div className="hidden md:block rounded-2xl border border-wire overflow-hidden mb-8">
          <div className="grid grid-cols-[auto_160px_1fr_1fr_90px] gap-0 bg-bg px-6 py-3 border-b border-wire">
            {tr.headers.map(h => (
              <div key={h} className="text-[10px] font-mono text-slate-600 uppercase tracking-widest">{h}</div>
            ))}
          </div>

          {STACK_STYLE.map(({ layer, Icon, tech, detail, clr, brd, bg }, i) => {
            const { name, role } = tr.stackItems[i]
            return (
              <div
                key={i}
                className="grid grid-cols-[auto_160px_1fr_1fr_90px] gap-0 px-6 py-4 border-b border-wire/50 last:border-0 hover:bg-surface transition-colors"
              >
                <div className="flex items-center gap-3 pr-6">
                  <div className={`p-1.5 rounded-lg border ${brd} ${bg}`}>
                    <Icon className={`w-4 h-4 ${clr}`} />
                  </div>
                  <span className="text-[11px] font-mono text-slate-600">{layer}</span>
                </div>
                <div className="flex items-center">
                  <span className={`text-sm font-bold font-mono ${clr}`}>{tech}</span>
                </div>
                <div className="flex flex-col justify-center pr-4">
                  <span className="text-sm text-white font-medium">{name}</span>
                  <span className="text-xs text-slate-500">{role}</span>
                </div>
                <div className="flex items-center">
                  <span className="text-xs font-mono text-slate-500">{detail}</span>
                </div>
                <div className="flex items-center">
                  <span className="flex items-center gap-1.5 text-xs font-mono text-accent">
                    <span className="w-1.5 h-1.5 rounded-full bg-accent" />
                    {tr.status}
                  </span>
                </div>
              </div>
            )
          })}
        </div>

        {/* Mobile cards */}
        <div className="md:hidden grid sm:grid-cols-2 gap-4 mb-8">
          {STACK_STYLE.map(({ layer, Icon, tech, detail, clr, brd, bg }, i) => {
            const { name, role } = tr.stackItems[i]
            return (
              <div key={i} className="rounded-xl border border-wire bg-surface p-4">
                <div className="flex items-center gap-3 mb-3">
                  <div className={`p-2 rounded-lg border ${brd} ${bg}`}>
                    <Icon className={`w-4 h-4 ${clr}`} />
                  </div>
                  <div>
                    <div className={`text-sm font-bold font-mono ${clr}`}>{tech}</div>
                    <div className="text-[10px] font-mono text-slate-600">{tr.layerLabel} {layer}</div>
                  </div>
                </div>
                <div className="text-sm text-white font-medium mb-0.5">{name}</div>
                <div className="text-xs text-slate-500">{role}</div>
                <div className="text-[11px] font-mono text-slate-600 mt-2">{detail}</div>
              </div>
            )
          })}
        </div>

        {/* Encryption flow */}
        <div className="rounded-2xl border border-wire bg-surface p-6">
          <p className="text-[10px] font-mono text-slate-600 uppercase tracking-widest mb-5">
            {tr.flowTag}
          </p>
          <div className="flex flex-wrap items-center gap-x-1 gap-y-2">
            {FLOW.map(([label, type], i) => (
              <span
                key={i}
                className={`text-xs sm:text-sm font-mono ${
                  type === 'password' ? 'text-accent font-bold' : 'text-slate-500'
                }`}
              >
                {label}
              </span>
            ))}
          </div>
          <p className="text-[10px] font-mono text-slate-600 mt-5">
            {tr.flowNote}
          </p>
        </div>

      </div>
    </section>
  )
}
