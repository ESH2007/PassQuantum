/* global React */
// PassQuantum — icons + shared primitives.
// All icons are stroke-only line icons matching Lucide proportions (24x24 viewBox,
// 1.5 stroke) so they sit comfortably with the humanist sans body type.

const Icon = ({ children, size = 16, className = "ico", style }) => (
  <svg
    viewBox="0 0 24 24"
    width={size}
    height={size}
    fill="none"
    stroke="currentColor"
    strokeWidth="1.6"
    strokeLinecap="round"
    strokeLinejoin="round"
    className={className}
    style={style}
    aria-hidden="true"
  >
    {children}
  </svg>
);

const Icons = {
  Vault: (p) => (
    <Icon {...p}>
      <rect x="3" y="4" width="18" height="16" rx="2" />
      <circle cx="12" cy="12" r="3.2" />
      <path d="M12 8.8V7.5M12 16.5v-1.3M15.2 12H16.5M7.5 12h1.3" />
    </Icon>
  ),
  Key: (p) => (
    <Icon {...p}>
      <circle cx="8.5" cy="14.5" r="3.5" />
      <path d="M11 12 21 2M17 6l2 2M14 9l2 2" />
    </Icon>
  ),
  Wand: (p) => (
    <Icon {...p}>
      <path d="M15 4V2M15 16v-2M8 9h2M20 9h2M17.8 11.8l1.4 1.4M12.2 6.2l1.4 1.4M17.8 6.2l-1.4 1.4M3 21l9-9" />
    </Icon>
  ),
  ShieldCheck: (p) => (
    <Icon {...p}>
      <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10Z" />
      <path d="m9 12 2 2 4-4" />
    </Icon>
  ),
  Search: (p) => (
    <Icon {...p}>
      <circle cx="11" cy="11" r="7" />
      <path d="m20 20-3.5-3.5" />
    </Icon>
  ),
  Settings: (p) => (
    <Icon {...p}>
      <circle cx="12" cy="12" r="2.6" />
      <path d="M19.4 15a1.7 1.7 0 0 0 .3 1.8l.1.1a2 2 0 1 1-2.8 2.8l-.1-.1a1.7 1.7 0 0 0-1.8-.3 1.7 1.7 0 0 0-1 1.5V21a2 2 0 1 1-4 0v-.1a1.7 1.7 0 0 0-1-1.5 1.7 1.7 0 0 0-1.8.3l-.1.1a2 2 0 1 1-2.8-2.8l.1-.1a1.7 1.7 0 0 0 .3-1.8 1.7 1.7 0 0 0-1.5-1H3a2 2 0 1 1 0-4h.1A1.7 1.7 0 0 0 4.6 9a1.7 1.7 0 0 0-.3-1.8l-.1-.1a2 2 0 1 1 2.8-2.8l.1.1a1.7 1.7 0 0 0 1.8.3H9a1.7 1.7 0 0 0 1-1.5V3a2 2 0 1 1 4 0v.1a1.7 1.7 0 0 0 1 1.5 1.7 1.7 0 0 0 1.8-.3l.1-.1a2 2 0 1 1 2.8 2.8l-.1.1a1.7 1.7 0 0 0-.3 1.8V9a1.7 1.7 0 0 0 1.5 1H21a2 2 0 1 1 0 4h-.1a1.7 1.7 0 0 0-1.5 1Z" />
    </Icon>
  ),
  Lock: (p) => (
    <Icon {...p}>
      <rect x="4" y="11" width="16" height="10" rx="2" />
      <path d="M8 11V7a4 4 0 0 1 8 0v4" />
    </Icon>
  ),
  LockOpen: (p) => (
    <Icon {...p}>
      <rect x="4" y="11" width="16" height="10" rx="2" />
      <path d="M8 11V7a4 4 0 0 1 7.5-2" />
    </Icon>
  ),
  Eye: (p) => (
    <Icon {...p}>
      <path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7S2 12 2 12Z" />
      <circle cx="12" cy="12" r="2.6" />
    </Icon>
  ),
  EyeOff: (p) => (
    <Icon {...p}>
      <path d="M9.9 5.1A10 10 0 0 1 12 5c6.5 0 10 7 10 7a14 14 0 0 1-3 3.7M6.5 7.2A14 14 0 0 0 2 12s3.5 7 10 7c1.9 0 3.5-.5 5-1.3M3 3l18 18M10 10.5a2.6 2.6 0 0 0 3.7 3.7" />
    </Icon>
  ),
  Copy: (p) => (
    <Icon {...p}>
      <rect x="9" y="9" width="11" height="11" rx="2" />
      <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
    </Icon>
  ),
  Trash: (p) => (
    <Icon {...p}>
      <path d="M3 6h18M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6M10 11v6M14 11v6" />
    </Icon>
  ),
  Plus: (p) => <Icon {...p}><path d="M12 5v14M5 12h14" /></Icon>,
  Edit: (p) => (
    <Icon {...p}>
      <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
      <path d="M18.4 2.6a2 2 0 0 1 2.8 2.8L12 14.6 8 16l1.4-4Z" />
    </Icon>
  ),
  Card: (p) => (
    <Icon {...p}>
      <rect x="2" y="5" width="20" height="14" rx="2" />
      <path d="M2 10h20M6 15h4" />
    </Icon>
  ),
  Note: (p) => (
    <Icon {...p}>
      <path d="M14 3H6a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9Z" />
      <path d="M14 3v6h6M8 13h8M8 17h5" />
    </Icon>
  ),
  Chevron: (p) => <Icon {...p}><path d="m6 9 6 6 6-6" /></Icon>,
  ChevronLeft: (p) => <Icon {...p}><path d="m15 6-6 6 6 6" /></Icon>,
  ChevronRight: (p) => <Icon {...p}><path d="m9 6 6 6-6 6" /></Icon>,
  Check: (p) => <Icon {...p}><path d="m5 12 5 5 9-11" /></Icon>,
  Cube: (p) => (
    <Icon {...p}>
      <path d="M12 2.5 21 7v10l-9 4.5L3 17V7Z" />
      <path d="M3 7 12 11.5M21 7 12 11.5M12 11.5V21.5" />
    </Icon>
  ),
  AlertTriangle: (p) => (
    <Icon {...p}>
      <path d="M10.3 3.7 2.6 17a2 2 0 0 0 1.7 3h15.4a2 2 0 0 0 1.7-3L13.7 3.7a2 2 0 0 0-3.4 0Z" />
      <path d="M12 9v4M12 17v.01" />
    </Icon>
  ),
  Info: (p) => (
    <Icon {...p}>
      <circle cx="12" cy="12" r="9" />
      <path d="M12 8v.01M11 12h1v5h1" />
    </Icon>
  ),
  Face: (p) => (
    <Icon {...p}>
      <rect x="3" y="3" width="18" height="18" rx="3" />
      <circle cx="9" cy="10" r="0.8" fill="currentColor" stroke="none" />
      <circle cx="15" cy="10" r="0.8" fill="currentColor" stroke="none" />
      <path d="M9 15.2c.9.9 2 1.3 3 1.3s2.1-.4 3-1.3" />
    </Icon>
  ),
  Refresh: (p) => (
    <Icon {...p}>
      <path d="M3 12a9 9 0 0 1 15.5-6.3L21 8M21 3v5h-5M21 12a9 9 0 0 1-15.5 6.3L3 16M3 21v-5h5" />
    </Icon>
  ),
  Download: (p) => (
    <Icon {...p}>
      <path d="M12 3v12M7 10l5 5 5-5M5 21h14" />
    </Icon>
  ),
  Upload: (p) => (
    <Icon {...p}>
      <path d="M12 17V5M7 10l5-5 5 5M5 21h14" />
    </Icon>
  ),
  ExternalLink: (p) => (
    <Icon {...p}>
      <path d="M15 3h6v6M21 3l-9 9M19 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V7a2 2 0 0 1 2-2h6" />
    </Icon>
  ),
  PanelLeft: (p) => (
    <Icon {...p}>
      <rect x="3" y="4" width="18" height="16" rx="2" />
      <path d="M9 4v16" />
    </Icon>
  ),
  PanelLeftClose: (p) => (
    <Icon {...p}>
      <rect x="3" y="4" width="18" height="16" rx="2" />
      <path d="M9 4v16M15 10l-2 2 2 2" />
    </Icon>
  ),
  Atom: (p) => (
    <Icon {...p}>
      <circle cx="12" cy="12" r="1.4" />
      <ellipse cx="12" cy="12" rx="9" ry="3.6" />
      <ellipse cx="12" cy="12" rx="9" ry="3.6" transform="rotate(60 12 12)" />
      <ellipse cx="12" cy="12" rx="9" ry="3.6" transform="rotate(120 12 12)" />
    </Icon>
  ),
};

// ── Primitives ─────────────────────────────────────────────────────────────

function Button({ kind = "default", size, block, children, leadingIcon, trailingIcon, onClick, ...rest }) {
  const cls = [
    "btn",
    kind === "primary" && "btn-primary",
    kind === "ghost" && "btn-ghost",
    kind === "danger" && "btn-danger",
    size === "sm" && "btn-sm",
    block && "btn-block",
  ].filter(Boolean).join(" ");
  return (
    <button type="button" className={cls} onClick={onClick} {...rest}>
      {leadingIcon}
      {children}
      {trailingIcon}
    </button>
  );
}

function IconBtn({ icon, onClick, title, ...rest }) {
  return (
    <button type="button" className="icon-btn" onClick={onClick} title={title} aria-label={title} {...rest}>
      {icon}
    </button>
  );
}

function Field({ label, hint, children, htmlFor, right }) {
  return (
    <div className="field">
      {label && (
        <label className="field-label" htmlFor={htmlFor}>
          <span>{label}</span>
          {right}
        </label>
      )}
      {children}
      {hint && <div className="field-help">{hint}</div>}
    </div>
  );
}

function TextInput({ value, onChange, placeholder, mono, type = "text", id }) {
  return (
    <input
      id={id}
      type={type}
      className={"input" + (mono ? " input-mono" : "")}
      value={value}
      onChange={(e) => onChange?.(e.target.value)}
      placeholder={placeholder}
    />
  );
}

function PasswordInput({ value, onChange, placeholder, id }) {
  const [show, setShow] = React.useState(false);
  return (
    <div className="input-wrap">
      <input
        id={id}
        type={show ? "text" : "password"}
        className="input input-mono"
        value={value}
        onChange={(e) => onChange?.(e.target.value)}
        placeholder={placeholder}
        autoComplete="off"
      />
      <div className="input-affix">
        <IconBtn
          icon={show ? <Icons.EyeOff /> : <Icons.Eye />}
          onClick={() => setShow((v) => !v)}
          title={show ? "Hide" : "Show"}
        />
        <IconBtn icon={<Icons.Copy />} title="Copy" />
      </div>
    </div>
  );
}

function Select({ value, onChange, children }) {
  return (
    <div className="select-wrap">
      <select className="select" value={value} onChange={(e) => onChange?.(e.target.value)}>
        {children}
      </select>
    </div>
  );
}

function Check({ checked, onChange, children }) {
  return (
    <label className={"check" + (checked ? " on" : "")} onClick={() => onChange?.(!checked)}>
      <input type="checkbox" checked={!!checked} readOnly />
      <span className="box">
        <Icons.Check />
      </span>
      <span>{children}</span>
    </label>
  );
}

function Card({ eyebrow, title, right, children, padded = true }) {
  return (
    <section className="card">
      {(title || eyebrow || right) && (
        <header className="card-hd">
          <div className="card-hd-l">
            <div>
              {eyebrow && <div className="card-hd-eyebrow">{eyebrow}</div>}
              {title && <div className="card-hd-title">{title}</div>}
            </div>
          </div>
          {right}
        </header>
      )}
      <div className={padded ? "card-body" : ""}>{children}</div>
    </section>
  );
}

function Pill({ tone = "mute", dot = true, children }) {
  return (
    <span className={`pill pill-${tone}`}>
      {dot && <span className="dot" />}
      {children}
    </span>
  );
}

function Tabs({ items, value, onChange }) {
  return (
    <div className="tabs" role="tablist">
      {items.map((it) => (
        <button
          key={it.value}
          role="tab"
          className="tab"
          aria-current={value === it.value}
          onClick={() => onChange?.(it.value)}
        >
          {it.label}
        </button>
      ))}
    </div>
  );
}

function StrengthMeter({ level }) {
  return (
    <div className="meter" data-level={level}>
      <i /><i /><i /><i /><i />
    </div>
  );
}

// Compute a strength level 0-5 from a password. Cheap heuristic — good enough
// to drive UI; the real Go app presumably uses zxcvbn or equivalent.
function scorePassword(p) {
  if (!p) return { level: 0, label: "—", crack: "—", entropy: 0 };
  let score = 0;
  if (p.length >= 8) score++;
  if (p.length >= 12) score++;
  if (p.length >= 16) score++;
  if (/[A-Z]/.test(p) && /[a-z]/.test(p)) score++;
  if (/\d/.test(p)) score++;
  if (/[^A-Za-z0-9]/.test(p)) score++;
  const level = Math.min(5, score);
  const labels = ["Empty", "Very weak", "Weak", "Fair", "Strong", "Excellent"];
  const crack = ["—", "Instantly", "Minutes", "Days", "Centuries", "10⁹⁺ centuries"];
  const charset =
    (/[a-z]/.test(p) ? 26 : 0) +
    (/[A-Z]/.test(p) ? 26 : 0) +
    (/\d/.test(p) ? 10 : 0) +
    (/[^A-Za-z0-9]/.test(p) ? 32 : 0);
  const entropy = charset > 0 ? Math.round(p.length * Math.log2(charset)) : 0;
  return { level, label: labels[level], crack: crack[level], entropy };
}

Object.assign(window, {
  Icons, Icon,
  Button, IconBtn, Field, TextInput, PasswordInput, Select, Check, Card, Pill, Tabs,
  StrengthMeter, scorePassword,
});
