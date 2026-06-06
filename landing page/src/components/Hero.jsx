import { ArrowRight, ChevronDown } from 'lucide-react'
import GithubIcon from './GithubIcon'
import FaceMesh from './FaceMesh'
import DownloadButton from './DownloadButton'
import RichText from './RichText'
import { useLang } from '../LangContext'
import t from '../translations'
import { REPO_URL, TECH } from '../config'

// Crypto labels come from TECH so naming stays canonical across the site.
const BADGES = [TECH.aes, TECH.mlkem, TECH.mldsa, 'Argon2id', 'MediaPipe', 'Local-first']

const mono = text => <span className="font-mono text-slate-200">{text}</span>

export default function Hero() {
  const { lang } = useLang()
  const tr = t[lang].hero

  return (
    <section className="relative min-h-screen flex flex-col justify-center px-4 sm:px-6 lg:px-8 pt-16 overflow-hidden">

      {/* Background layers */}
      <div className="absolute inset-0 grid-bg" />
      <div className="absolute inset-0 hero-glow pointer-events-none" />
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[700px] h-[500px] rounded-full bg-accent/3 blur-[120px] pointer-events-none" />

      <div className="relative z-10 max-w-7xl mx-auto w-full py-16 lg:py-24">
        <div className="grid lg:grid-cols-2 gap-12 lg:gap-20 items-center">

          {/* Left: copy */}
          <div className="text-center lg:text-left">
            {/* Live badge */}
            <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full border border-accent/25 bg-accent/5 mb-8 animate-fade-in-up">
              <span className="w-2 h-2 rounded-full bg-accent animate-pulse" />
              <span className="text-[11px] font-mono text-accent tracking-widest uppercase">
                {tr.badge}
              </span>
            </div>

            <h1 className="text-4xl sm:text-5xl lg:text-[3.4rem] font-extrabold text-white leading-[1.1] tracking-tight mb-6 animate-fade-in-up delay-100">
              <RichText
                text={tr.headline}
                parts={{
                  accent: <span className="text-accent">{tr.headlineAccent}</span>,
                  underline: (
                    <span className="relative">
                      {tr.headlineUnderline}
                      <span className="absolute -bottom-0.5 left-0 right-0 h-[2px] bg-gradient-to-r from-accent to-transparent rounded" />
                    </span>
                  ),
                }}
              />
            </h1>

            <p className="text-slate-400 text-base sm:text-lg leading-relaxed mb-8 max-w-xl mx-auto lg:mx-0 animate-fade-in-up delay-200">
              <RichText
                text={tr.sub}
                parts={{ aes: mono(TECH.aes), mlkem: mono(TECH.mlkem), mldsa: mono(TECH.mldsa) }}
              />
            </p>

            {/* CTAs */}
            <div className="flex flex-col sm:flex-row gap-3 justify-center lg:justify-start mb-10 animate-fade-in-up delay-300">
              <a
                href={REPO_URL}
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center justify-center gap-2 px-6 py-3.5 rounded-xl bg-accent text-bg font-bold text-sm hover:bg-teal-300 transition-all animate-glow-pulse group"
              >
                <GithubIcon className="w-4 h-4" />
                {tr.cta1}
                <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
              </a>
              <a
                href="#features"
                className="flex items-center justify-center gap-2 px-6 py-3.5 rounded-xl border border-wire text-slate-300 text-sm hover:border-accent/40 hover:text-accent transition-all"
              >
                {tr.cta2}
              </a>
              <DownloadButton />
            </div>

            {/* Tech badges */}
            <div className="flex flex-wrap gap-2 justify-center lg:justify-start animate-fade-in-up delay-500">
              {BADGES.map(b => (
                <span
                  key={b}
                  className="px-3 py-1 rounded-full text-xs font-mono border border-wire text-slate-500 bg-surface/70 hover:border-accent/30 hover:text-accent/80 transition-colors cursor-default"
                >
                  {b}
                </span>
              ))}
            </div>
          </div>

          {/* Right: face mesh */}
          <div className="flex justify-center lg:justify-end animate-float">
            <FaceMesh />
          </div>

        </div>
      </div>

      {/* Scroll hint */}
      <a
        href="#features"
        className="absolute bottom-8 left-1/2 -translate-x-1/2 text-slate-600 hover:text-accent transition-colors"
        aria-label="Scroll down"
      >
        <ChevronDown className="w-6 h-6 animate-bounce" />
      </a>
    </section>
  )
}
