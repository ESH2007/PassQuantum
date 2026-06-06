// Brand mark — simplified magnifier + teal core, optimized for small sizes.
export default function Mark({ className = 'w-6 h-6' }) {
  return (
    <svg viewBox="0 0 64 64" className={className} role="img" aria-label="PassQuantum">
      <rect x="2" y="2" width="60" height="60" rx="14" fill="#0E1117" />
      <path
        d="M26 18 L26 12 A6 6 0 0 1 38 12 L38 18"
        fill="none"
        stroke="#EAEEF2"
        strokeWidth="3.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <circle cx="32" cy="34" r="16" fill="none" stroke="#EAEEF2" strokeWidth="3.5" />
      <line x1="42" y1="44" x2="50" y2="52" stroke="#EAEEF2" strokeWidth="4" strokeLinecap="round" />
      <circle cx="32" cy="34" r="4" fill="#2DD4BF" />
    </svg>
  )
}
