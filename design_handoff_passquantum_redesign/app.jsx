/* global React, ReactDOM, Icons, Button, IconBtn, Pill,
   ScreenVaults, ScreenAddItem, ScreenItems, ScreenGenerator, ScreenChecker, ScreenSettings,
   useTweaks, TweaksPanel, TweakSection, TweakRadio, TweakColor, TweakToggle, TweakSelect */

// ─────────────────────────────────────────────────────────────────────────────
// PassQuantum — main app shell.
// Owns nav state, vault/items state, and the tweaks bridge that repaints CSS
// custom properties on <html> based on user selections in the Tweaks panel.
// ─────────────────────────────────────────────────────────────────────────────

const TWEAK_DEFAULTS = /*EDITMODE-BEGIN*/{
  "accent": "#3b82f6",
  "density": "comfortable",
  "fontPair": "plex",
  "background": "hex",
  "sidebar": "expanded",
  "cardLayout": "rows",
  "watching": true
}/*EDITMODE-END*/;

const ACCENT_PALETTES = [
  ["#3b82f6", "rgba(59,130,246,0.14)", "rgba(59,130,246,0.40)", "#c9dcfb"],   // Blue
  ["#10b981", "rgba(16,185,129,0.14)", "rgba(16,185,129,0.40)", "#a7e6d0"],   // Emerald
  ["#c8a464", "rgba(200,164,100,0.14)", "rgba(200,164,100,0.40)", "#ecd9b0"], // Gold
  ["#e6e6e6", "rgba(230,230,230,0.10)", "rgba(230,230,230,0.32)", "#f2f2f2"], // Mono
  ["#06b6d4", "rgba(6,182,212,0.14)", "rgba(6,182,212,0.40)", "#a6e3ef"],     // Cyan-quiet
];

const FONT_PAIRS = {
  plex:   { sans: '"IBM Plex Sans", ui-sans-serif, system-ui, sans-serif',     mono: '"IBM Plex Mono", ui-monospace, monospace' },
  inter:  { sans: '"Inter", ui-sans-serif, system-ui, sans-serif',             mono: '"JetBrains Mono", ui-monospace, monospace' },
  geist:  { sans: '"Geist", ui-sans-serif, system-ui, sans-serif',             mono: '"Geist Mono", ui-monospace, monospace' },
};

function applyTweaks(t) {
  const root = document.documentElement;
  // Accent
  const idx = ACCENT_PALETTES.findIndex(p => p[0].toLowerCase() === String(t.accent).toLowerCase());
  const [a, soft, line] = ACCENT_PALETTES[idx >= 0 ? idx : 0];
  root.style.setProperty("--accent", a);
  root.style.setProperty("--accent-soft", soft);
  root.style.setProperty("--accent-line", line);
  root.style.setProperty("--accent-fg", a === "#e6e6e6" ? "#0b0e13" : "#ffffff");

  // Density
  root.style.setProperty("--d-step", t.density === "compact" ? "0.82" : t.density === "comfy" ? "1.08" : "1");

  // Fonts
  const fp = FONT_PAIRS[t.fontPair] || FONT_PAIRS.plex;
  root.style.setProperty("--font-sans", fp.sans);
  root.style.setProperty("--font-mono", fp.mono);
}

// ─────────────────────────────────────────────────────────────────────────────
// Seed data
// ─────────────────────────────────────────────────────────────────────────────
const SEED_VAULTS = [
  { id: "default", name: "Default", path: "~/.config/passquantum/default.enc", itemCount: 24, modified: "2 min ago" },
  { id: "work",    name: "Work",    path: "~/.config/passquantum/work.enc",    itemCount: 41, modified: "yesterday" },
  { id: "legacy",  name: "Legacy",  path: "~/Documents/vaults/legacy.enc",     itemCount: 9,  modified: "3 mo ago" },
];

const SEED_ITEMS = [
  { id: "1",  kind: "password", title: "github.com",          sub: "ada@personal.io",       passwordLength: 18 },
  { id: "2",  kind: "password", title: "Tailscale admin",     sub: "ada@acme.dev",          passwordLength: 22 },
  { id: "3",  kind: "card",     title: "Personal Visa",       sub: "Ada Lovelace · 04/29",  last4: "2345" },
  { id: "4",  kind: "note",     title: "Recovery codes · 2FA",sub: "Updated 4d ago",        noteSnippet: "8 unused codes · last rotated…" },
  { id: "5",  kind: "password", title: "Vanguard",            sub: "ada.l@personal.io",     passwordLength: 16 },
  { id: "6",  kind: "password", title: "AWS root",            sub: "ops+root@acme.dev",     passwordLength: 32 },
  { id: "7",  kind: "card",     title: "Acme Corporate",      sub: "Ada Lovelace · 11/27",  last4: "0017" },
  { id: "8",  kind: "note",     title: "Server SSH keys",     sub: "infra · 12 entries",    noteSnippet: "ed25519 keypair, prod & staging…" },
  { id: "9",  kind: "password", title: "Cloudflare",          sub: "ada@acme.dev",          passwordLength: 24 },
];

const SEED_MONITORED = [
  { label: "PassQuantum",       process: "passquantum",       enabled: true },
  { label: "Firefox (private)", process: "firefox --private", enabled: true },
  { label: "Signal Desktop",    process: "signal-desktop",    enabled: true },
  { label: "Slack",             process: "slack",             enabled: false },
  { label: "VS Code",           process: "code",              enabled: false },
];

// ─────────────────────────────────────────────────────────────────────────────
// App
// ─────────────────────────────────────────────────────────────────────────────
function App() {
  const [t, setTweak] = useTweaks(TWEAK_DEFAULTS);
  React.useEffect(() => { applyTweaks(t); }, [t.accent, t.density, t.fontPair]);

  const [route, setRoute] = React.useState("vaults");
  const [activeVaultId, setActiveVaultId] = React.useState("default");
  const [vaults, setVaults] = React.useState(SEED_VAULTS);
  const [items, setItems] = React.useState(SEED_ITEMS);
  const [monitoredApps, setMonitoredApps] = React.useState(SEED_MONITORED);
  const [draft, setDraft] = React.useState({
    kind: "password",
    service: "", username: "", password: "",
    cardType: "credit", cardNickname: "", cardholder: "", cardNumber: "", expiry: "", cvv: "",
    noteTitle: "", noteBody: "",
  });

  const activeVault = vaults.find(v => v.id === activeVaultId) || vaults[0];

  const sidebarCollapsed = t.sidebar === "collapsed";

  const navIfActive = (key) => route === key ? { "aria-current": "page" } : {};

  const ROUTES = [
    { key: "vaults",    label: "Vaults",    icon: <Icons.Vault size={18} /> },
    { key: "passwords", label: "Add item",  icon: <Icons.Plus size={18} /> },
    { key: "items",     label: "Items",     icon: <Icons.Key size={18} />, meta: items.length },
    { key: "generate",  label: "Generate",  icon: <Icons.Wand size={18} /> },
    { key: "check",     label: "Analyze",   icon: <Icons.ShieldCheck size={18} /> },
    { key: "settings",  label: "Settings",  icon: <Icons.Settings size={18} /> },
  ];

  const topbarCrumbs = {
    vaults:    { title: "Vaults",        crumb: ["Vaults"] },
    passwords: { title: "Add item",      crumb: ["Vaults", activeVault.name, "Add item"] },
    items:     { title: "Items",         crumb: ["Vaults", activeVault.name, "Items"] },
    generate:  { title: "Generator",     crumb: ["Tools", "Generator"] },
    check:     { title: "Analyzer",      crumb: ["Tools", "Analyzer"] },
    settings:  { title: "Settings",      crumb: ["Application", "Settings"] },
  }[route];

  // Backdrop class swap
  const backgroundCls = {
    hex:   "bg-hex",
    grid:  "bg-grid",
    dots:  "bg-dots",
    plain: "bg-plain",
  }[t.background] || "bg-hex";

  return (
    <div className={"app " + backgroundCls}
         style={{ ["--rail-w"]: sidebarCollapsed ? "var(--rail-w-collapsed)" : "var(--rail-w-expanded)" }}>

      {/* ── Sidebar ─────────────────────────────────────────────────────── */}
      <aside className={"rail" + (sidebarCollapsed ? " collapsed" : "")}>
        <div className="rail-brand">
          <div className="rail-brand-mark">
            <Icons.Atom size={18} />
          </div>
          <div className="rail-brand-text">
            <span className="rail-brand-name">PassQuantum</span>
            <span className="rail-brand-meta">v1.0 · PQ-Safe</span>
          </div>
        </div>

        <div className="rail-section">Vault</div>
        <nav className="nav">
          {ROUTES.slice(0, 3).map(r => (
            <a key={r.key} className="nav-item" data-tip={r.label}
               {...navIfActive(r.key)}
               onClick={() => setRoute(r.key)}>
              <span className="nav-item-icon">{r.icon}</span>
              <span className="nav-item-label">{r.label}</span>
              {r.meta != null && <span className="nav-item-meta">{r.meta}</span>}
            </a>
          ))}
        </nav>

        <div className="rail-section">Tools</div>
        <nav className="nav">
          {ROUTES.slice(3, 5).map(r => (
            <a key={r.key} className="nav-item" data-tip={r.label}
               {...navIfActive(r.key)}
               onClick={() => setRoute(r.key)}>
              <span className="nav-item-icon">{r.icon}</span>
              <span className="nav-item-label">{r.label}</span>
            </a>
          ))}
        </nav>

        <div className="rail-section">System</div>
        <nav className="nav">
          <a className="nav-item" data-tip="Settings"
             {...navIfActive("settings")}
             onClick={() => setRoute("settings")}>
            <span className="nav-item-icon"><Icons.Settings size={18} /></span>
            <span className="nav-item-label">Settings</span>
          </a>
        </nav>

        <div className="rail-foot">
          <button className="rail-collapse-btn"
                  onClick={() => setTweak("sidebar", sidebarCollapsed ? "expanded" : "collapsed")}>
            {sidebarCollapsed ? <Icons.PanelLeft size={14} /> : <Icons.PanelLeftClose size={14} />}
            <span className="rail-foot-text">{sidebarCollapsed ? "" : "Collapse"}</span>
          </button>
          <a className="nav-item" data-tip="Lock vault"
             onClick={() => alert("Vault locked. Cleared decrypted state from memory.")}
             style={{ color: "var(--fg-1)" }}>
            <span className="nav-item-icon"><Icons.Lock size={18} /></span>
            <span className="nav-item-label">Lock vault</span>
          </a>
        </div>
      </aside>

      {/* ── Main ────────────────────────────────────────────────────────── */}
      <main className="main">
        <header className="topbar">
          <div className="topbar-crumb">
            {topbarCrumbs.crumb.map((c, i) => (
              <React.Fragment key={i}>
                {i > 0 && <span className="topbar-crumb-sep">/</span>}
                <span>{c}</span>
              </React.Fragment>
            ))}
          </div>
          <div className="topbar-spacer" />
          <Pill tone="accent">Vault · {activeVault.name}</Pill>
          <Pill tone={t.watching ? "ok" : "mute"}>
            <Icons.Face size={11} />
            <span style={{ marginLeft: 2 }}>{t.watching ? "Watching · ON" : "Presence guard · OFF"}</span>
          </Pill>
          <Pill tone="mute" dot={false}>
            <Icons.LockOpen size={11} />
            <span style={{ marginLeft: 2 }}>Unlocked</span>
          </Pill>
        </header>

        <div className="page">
          <div className="page-inner">
            {route === "vaults" && (
              <ScreenVaults
                vaults={vaults}
                activeVaultId={activeVaultId}
                onOpen={(id) => { setActiveVaultId(id); setRoute("items"); }}
                onDelete={(id) => setVaults(vaults.filter(v => v.id !== id))}
                onCreate={() => alert("Create vault flow (placeholder)")}
              />
            )}
            {route === "passwords" && (
              <ScreenAddItem
                activeVault={activeVault.name}
                draft={draft}
                setDraft={setDraft}
                goToItems={() => setRoute("items")}
                onSave={() => {
                  // Synthesize an item from the draft and push it.
                  const id = String(Date.now());
                  if (draft.kind === "password" && draft.service) {
                    setItems([{ id, kind: "password", title: draft.service, sub: draft.username || "—", passwordLength: draft.password.length || 12 }, ...items]);
                  } else if (draft.kind === "card" && (draft.cardNickname || draft.cardNumber)) {
                    setItems([{ id, kind: "card", title: draft.cardNickname || "Card", sub: draft.cardholder || "—", last4: draft.cardNumber.slice(-4) || "0000" }, ...items]);
                  } else if (draft.kind === "note" && draft.noteTitle) {
                    setItems([{ id, kind: "note", title: draft.noteTitle, sub: "Just now", noteSnippet: (draft.noteBody || "").slice(0, 80) }, ...items]);
                  }
                  setRoute("items");
                }}
                onCancel={() => setRoute("items")}
              />
            )}
            {route === "items" && (
              <ScreenItems
                activeVault={activeVault.name}
                items={items}
                layout={t.cardLayout}
                onLayoutChange={(v) => setTweak("cardLayout", v)}
                onBack={() => setRoute("vaults")}
                onDelete={(id) => setItems(items.filter(i => i.id !== id))}
                onAdd={() => setRoute("passwords")}
              />
            )}
            {route === "generate" && (
              <ScreenGenerator
                vaults={vaults}
                activeVaultId={activeVaultId}
                onSaveToVault={(data) => {
                  const id = String(Date.now());
                  setItems([{ id, kind: "password", title: data.service, sub: data.username, passwordLength: data.password.length }, ...items]);
                  setRoute("items");
                }}
              />
            )}
            {route === "check" && <ScreenChecker />}
            {route === "settings" && (
              <ScreenSettings
                watching={t.watching}
                setWatching={(v) => setTweak("watching", v)}
                monitoredApps={monitoredApps}
                setMonitoredApps={setMonitoredApps}
              />
            )}
          </div>
        </div>
      </main>

      {/* ── Tweaks panel ─────────────────────────────────────────────────── */}
      <TweaksPanel title="Tweaks">
        <TweakSection label="Theme" />
        <TweakColor
          label="Accent"
          value={t.accent}
          options={ACCENT_PALETTES.map(p => p[0])}
          onChange={(v) => setTweak("accent", v)}
        />
        <TweakSelect
          label="Background pattern"
          value={t.background}
          options={[
            { value: "hex",   label: "Hex lattice" },
            { value: "grid",  label: "Square grid" },
            { value: "dots",  label: "Dotgrid" },
            { value: "plain", label: "Plain" },
          ]}
          onChange={(v) => setTweak("background", v)}
        />

        <TweakSection label="Typography &amp; density" />
        <TweakSelect
          label="Font pair"
          value={t.fontPair}
          options={[
            { value: "plex",  label: "IBM Plex Sans / Mono" },
            { value: "inter", label: "Inter / JetBrains Mono" },
            { value: "geist", label: "Geist Sans / Mono" },
          ]}
          onChange={(v) => setTweak("fontPair", v)}
        />
        <TweakRadio
          label="Density"
          value={t.density}
          options={[
            { value: "compact",      label: "Compact" },
            { value: "comfortable",  label: "Default" },
            { value: "comfy",        label: "Comfy" },
          ]}
          onChange={(v) => setTweak("density", v)}
        />

        <TweakSection label="Layout" />
        <TweakRadio
          label="Sidebar"
          value={t.sidebar}
          options={[
            { value: "expanded",  label: "Expanded" },
            { value: "collapsed", label: "Icons" },
          ]}
          onChange={(v) => setTweak("sidebar", v)}
        />
        <TweakRadio
          label="Items view"
          value={t.cardLayout}
          options={[
            { value: "rows", label: "Rows" },
            { value: "grid", label: "Cards" },
          ]}
          onChange={(v) => setTweak("cardLayout", v)}
        />

        <TweakSection label="Status" />
        <TweakToggle
          label="Presence guard"
          value={t.watching}
          onChange={(v) => setTweak("watching", v)}
        />
      </TweaksPanel>
    </div>
  );
}

const root = ReactDOM.createRoot(document.getElementById("root"));
root.render(<App />);
