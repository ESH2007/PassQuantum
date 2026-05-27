/* global React, Icons, Button, IconBtn, Field, TextInput, PasswordInput, Select, Check, Card, Pill, Tabs, StrengthMeter, scorePassword */
// PassQuantum — screens. Each screen renders inside the App shell's <main>
// area. State is owned by App; screens are presentational + emit intents.

// ─────────────────────────────────────────────────────────────────────────────
// VAULTS
// ─────────────────────────────────────────────────────────────────────────────
function ScreenVaults({ vaults, activeVaultId, onOpen, onDelete, onCreate }) {
  return (
    <>
      <div className="page-head">
        <div>
          <div className="page-eyebrow">PassQuantum / Vaults</div>
          <h1>Your vaults</h1>
          <div className="page-sub">Encrypted containers stored locally. Each vault has its own master key derived from your passphrase.</div>
        </div>
        <Button kind="primary" leadingIcon={<Icons.Plus size={14} />} onClick={onCreate}>
          New vault
        </Button>
      </div>

      <div className="items">
        {vaults.map((v) => {
          const active = v.id === activeVaultId;
          return (
            <article className="item" key={v.id}>
              <div className={"item-mark " + (active ? "tone-blue" : "")}>
                <Icons.Cube size={18} />
              </div>
              <div className="item-body">
                <div className="item-title-row">
                  <span className="item-title">{v.name}</span>
                  {active && <Pill tone="ok">Active</Pill>}
                  <Pill tone="mute" dot={false}>{v.itemCount} items</Pill>
                </div>
                <div className="item-sub">
                  <span>{v.path}</span>
                  <span>·</span>
                  <span>Modified {v.modified}</span>
                </div>
              </div>
              <div className="item-actions">
                <Button size="sm" onClick={() => onOpen?.(v.id)}>
                  {active ? "Open" : "Switch"}
                </Button>
                <IconBtn icon={<Icons.Edit />} title="Rename" />
                <IconBtn icon={<Icons.Download />} title="Export" />
                <IconBtn icon={<Icons.Trash />} title="Delete" onClick={() => onDelete?.(v.id)} />
              </div>
            </article>
          );
        })}
      </div>

      <Card eyebrow="VAULT INTEGRITY" title="Encryption status">
        <dl className="kv">
          <dt>Algorithm</dt>
          <dd>AES-256-GCM<small>Authenticated symmetric encryption</small></dd>
          <dt>Key encapsulation</dt>
          <dd>ML-KEM-768 (Kyber)<small>NIST FIPS 203 · post-quantum safe</small></dd>
          <dt>KDF</dt>
          <dd>Argon2id<small>m=64 MiB · t=3 · p=4</small></dd>
          <dt>Vault file</dt>
          <dd className="mono">~/.config/passquantum/{activeVaultId || "default"}.enc</dd>
        </dl>
      </Card>
    </>
  );
}

// ─────────────────────────────────────────────────────────────────────────────
// ADD / EDIT ITEM
// ─────────────────────────────────────────────────────────────────────────────
function ScreenAddItem({ activeVault, draft, setDraft, onSave, onCancel, goToItems }) {
  const score = scorePassword(draft.kind === "password" ? draft.password : "");
  return (
    <>
      <div className="page-head">
        <div>
          <div className="page-eyebrow">PassQuantum / {activeVault} / Add item</div>
          <h1>Add vault item</h1>
          <div className="page-sub">Items are encrypted with the vault key before they touch disk.</div>
        </div>
        <Button kind="ghost" leadingIcon={<Icons.ChevronLeft size={14} />} onClick={goToItems}>
          All items
        </Button>
      </div>

      <Card padded={false}>
        <div className="card-body" style={{ display: "flex", flexDirection: "column", gap: "var(--space-5)" }}>
          <Field label="Item type">
            <Select value={draft.kind} onChange={(v) => setDraft({ ...draft, kind: v })}>
              <option value="password">Password</option>
              <option value="card">Card</option>
              <option value="note">Cyphered note</option>
            </Select>
          </Field>

          {draft.kind === "password" && (
            <>
              <Field label="Service">
                <TextInput
                  value={draft.service}
                  onChange={(v) => setDraft({ ...draft, service: v })}
                  placeholder="e.g. github.com, Vanguard, Tailscale"
                />
              </Field>
              <Field label="Username / email">
                <TextInput
                  value={draft.username}
                  onChange={(v) => setDraft({ ...draft, username: v })}
                  placeholder="you@example.com"
                />
              </Field>
              <Field
                label="Password"
                right={
                  <a className="muted mono" style={{ textDecoration: "none", fontSize: 10 }} href="#">
                    Generate →
                  </a>
                }
              >
                <PasswordInput
                  value={draft.password}
                  onChange={(v) => setDraft({ ...draft, password: v })}
                  placeholder="Enter or generate a password"
                />
              </Field>
              <Field label="Strength">
                <StrengthBlock score={score} />
              </Field>
            </>
          )}

          {draft.kind === "card" && (
            <>
              <Field label="Card type">
                <Select
                  value={draft.cardType || "credit"}
                  onChange={(v) => setDraft({ ...draft, cardType: v })}
                >
                  <option value="credit">Credit</option>
                  <option value="debit">Debit</option>
                  <option value="prepaid">Prepaid</option>
                </Select>
              </Field>
              <Field label="Nickname">
                <TextInput
                  value={draft.cardNickname}
                  onChange={(v) => setDraft({ ...draft, cardNickname: v })}
                  placeholder="e.g. Personal Visa"
                />
              </Field>
              <Field label="Cardholder">
                <TextInput
                  value={draft.cardholder}
                  onChange={(v) => setDraft({ ...draft, cardholder: v })}
                  placeholder="Name on card"
                />
              </Field>
              <Field label="Card number">
                <TextInput
                  value={draft.cardNumber}
                  onChange={(v) => setDraft({ ...draft, cardNumber: v })}
                  placeholder="•••• •••• •••• ••••"
                  mono
                />
              </Field>
              <div className="field-row">
                <Field label="Expiry">
                  <TextInput
                    value={draft.expiry}
                    onChange={(v) => setDraft({ ...draft, expiry: v })}
                    placeholder="MM / YY"
                    mono
                  />
                </Field>
                <Field label="CVV">
                  <PasswordInput
                    value={draft.cvv}
                    onChange={(v) => setDraft({ ...draft, cvv: v })}
                    placeholder="•••"
                  />
                </Field>
              </div>
            </>
          )}

          {draft.kind === "note" && (
            <>
              <Field label="Note title">
                <TextInput
                  value={draft.noteTitle}
                  onChange={(v) => setDraft({ ...draft, noteTitle: v })}
                  placeholder="Recovery codes · 2025"
                />
              </Field>
              <Field
                label="Cyphered note"
                hint="Plaintext is encrypted at rest with the vault key. Markdown is preserved."
              >
                <textarea
                  className="textarea input-mono"
                  value={draft.noteBody}
                  onChange={(e) => setDraft({ ...draft, noteBody: e.target.value })}
                  placeholder="Write your note here…"
                  rows={8}
                />
              </Field>
            </>
          )}
        </div>

        <footer
          className="card-hd"
          style={{ borderBottom: 0, borderTop: "1px solid var(--line-1)", background: "var(--bg-1)", justifyContent: "space-between" }}
        >
          <span className="mono" style={{ fontSize: 11, color: "var(--fg-2)" }}>
            Encrypted on save · AES-256-GCM
          </span>
          <div style={{ display: "flex", gap: 8 }}>
            <Button kind="ghost" onClick={onCancel}>Cancel</Button>
            <Button kind="primary" leadingIcon={<Icons.Plus size={14} />} onClick={onSave}>
              Save item
            </Button>
          </div>
        </footer>
      </Card>
    </>
  );
}

function StrengthBlock({ score }) {
  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
      <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", gap: 12 }}>
        <div style={{ flex: 1 }}><StrengthMeter level={score.level} /></div>
        <span className="mono" style={{ fontSize: 11, color: "var(--fg-1)", minWidth: 72, textAlign: "right" }}>
          {score.label}
        </span>
      </div>
      <div className="mono" style={{ fontSize: 11, color: "var(--fg-2)", display: "flex", gap: 16, flexWrap: "wrap" }}>
        <span>Crack time · {score.crack}</span>
        <span>Entropy · {score.entropy} bits</span>
      </div>
    </div>
  );
}

// ─────────────────────────────────────────────────────────────────────────────
// ITEMS LIST
// ─────────────────────────────────────────────────────────────────────────────
function ScreenItems({ activeVault, items, layout, onLayoutChange, onBack, onDelete, onAdd }) {
  const [query, setQuery] = React.useState("");
  const filtered = items.filter((i) => {
    const q = query.toLowerCase();
    return !q ||
      (i.title || "").toLowerCase().includes(q) ||
      (i.sub || "").toLowerCase().includes(q) ||
      (i.kind || "").toLowerCase().includes(q);
  });

  return (
    <>
      <div className="page-head">
        <div>
          <div className="page-eyebrow">PassQuantum / {activeVault} / Items</div>
          <h1>Vault items <span className="muted" style={{ fontWeight: 400, fontSize: 18 }}>· {items.length}</span></h1>
          <div className="page-sub">All entries decrypted in-memory only. Lock the vault to purge.</div>
        </div>
        <Button kind="primary" leadingIcon={<Icons.Plus size={14} />} onClick={onAdd}>Add item</Button>
      </div>

      <div style={{ display: "flex", gap: 10, alignItems: "center" }}>
        <div style={{ flex: 1, position: "relative" }}>
          <span style={{ position: "absolute", left: 10, top: "50%", transform: "translateY(-50%)", color: "var(--fg-2)" }}>
            <Icons.Search size={14} />
          </span>
          <input
            className="input"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search by service, username, kind…"
            style={{ paddingLeft: 32 }}
          />
        </div>
        <div className="btn-group" role="tablist" aria-label="Layout">
          <button
            className="btn"
            aria-pressed={layout === "rows"}
            onClick={() => onLayoutChange?.("rows")}
          >Rows</button>
          <button
            className="btn"
            aria-pressed={layout === "grid"}
            onClick={() => onLayoutChange?.("grid")}
          >Grid</button>
        </div>
      </div>

      {filtered.length === 0 && (
        <div className="empty">
          <Icons.Search size={20} />
          <div style={{ fontSize: 13, color: "var(--fg-1)" }}>No items match “{query}”</div>
          <div style={{ fontSize: 12 }}>Try a different keyword or clear the search.</div>
        </div>
      )}

      {layout === "rows" ? (
        <div className="items">{filtered.map((it) => <RowItem key={it.id} item={it} onDelete={onDelete} />)}</div>
      ) : (
        <div className="items-grid">{filtered.map((it) => <CardItem key={it.id} item={it} onDelete={onDelete} />)}</div>
      )}
    </>
  );
}

function kindMeta(kind) {
  switch (kind) {
    case "password": return { label: "Password", tone: "tone-blue", icon: <Icons.Key size={16} /> };
    case "card":     return { label: "Card",     tone: "tone-warn", icon: <Icons.Card size={16} /> };
    case "note":     return { label: "Note",     tone: "tone-green", icon: <Icons.Note size={16} /> };
    default:         return { label: "Item",     tone: "",           icon: <Icons.Vault size={16} /> };
  }
}

function maskedSecret(item) {
  if (item.kind === "password") return "•".repeat(Math.min(14, item.passwordLength || 12));
  if (item.kind === "card") return "•••• •••• •••• " + (item.last4 || "0000");
  if (item.kind === "note") return item.noteSnippet || "Encrypted note";
  return "";
}

function RowItem({ item, onDelete }) {
  const m = kindMeta(item.kind);
  return (
    <article className="item">
      <div className={"item-mark " + m.tone}>{m.icon}</div>
      <div className="item-body">
        <div className="item-title-row">
          <span className="item-title">{item.title}</span>
          <span className="item-kind">{m.label}</span>
        </div>
        <div className="item-sub">
          <span>{item.sub}</span>
          {item.kind !== "note" && <><span>·</span><span className="mono">{maskedSecret(item)}</span></>}
        </div>
      </div>
      <div className="item-actions">
        <Button size="sm" kind="ghost">Show</Button>
        <Button size="sm">Copy</Button>
        <IconBtn icon={<Icons.Edit />} title="Edit" />
        <IconBtn icon={<Icons.Trash />} title="Delete" onClick={() => onDelete?.(item.id)} />
      </div>
    </article>
  );
}

function CardItem({ item, onDelete }) {
  const m = kindMeta(item.kind);
  return (
    <div className="item-card">
      <div style={{ display: "flex", alignItems: "center", gap: 10 }}>
        <div className={"item-mark " + m.tone}>{m.icon}</div>
        <div style={{ minWidth: 0, flex: 1 }}>
          <div className="item-title-row">
            <span className="item-title">{item.title}</span>
          </div>
          <div className="item-sub"><span>{m.label}</span><span>·</span><span>{item.sub}</span></div>
        </div>
      </div>
      <div className="mono" style={{ fontSize: 12, color: "var(--fg-1)", background: "var(--bg-1)", padding: "8px 10px", borderRadius: 6, border: "1px solid var(--line-1)" }}>
        {maskedSecret(item)}
      </div>
      <div className="item-actions">
        <Button size="sm" kind="ghost" block>Show</Button>
        <Button size="sm" block>Copy</Button>
        <IconBtn icon={<Icons.Trash />} title="Delete" onClick={() => onDelete?.(item.id)} />
      </div>
    </div>
  );
}

// ─────────────────────────────────────────────────────────────────────────────
// GENERATOR
// ─────────────────────────────────────────────────────────────────────────────
function ScreenGenerator({ onSaveToVault, vaults, activeVaultId }) {
  const [length, setLength] = React.useState(20);
  const [opts, setOpts] = React.useState({ upper: true, lower: true, digits: true, symbols: true, excludeAmbiguous: true });
  const [pwd, setPwd] = React.useState("");
  const [savingOpen, setSavingOpen] = React.useState(false);

  const regen = React.useCallback(() => {
    const sets = [];
    let amb = "";
    if (opts.excludeAmbiguous) amb = "il1Lo0O";
    if (opts.upper) sets.push("ABCDEFGHIJKLMNOPQRSTUVWXYZ".split("").filter(c => !amb.includes(c)).join(""));
    if (opts.lower) sets.push("abcdefghijklmnopqrstuvwxyz".split("").filter(c => !amb.includes(c)).join(""));
    if (opts.digits) sets.push("0123456789".split("").filter(c => !amb.includes(c)).join(""));
    if (opts.symbols) sets.push("!@#$%^&*()-_=+[]{};:,.?/");
    const all = sets.join("");
    if (!all) { setPwd(""); return; }
    let out = "";
    const buf = new Uint32Array(length);
    crypto.getRandomValues(buf);
    for (let i = 0; i < length; i++) out += all[buf[i] % all.length];
    setPwd(out);
  }, [length, opts]);

  React.useEffect(() => { regen(); }, [regen]);
  const score = scorePassword(pwd);

  return (
    <>
      <div className="page-head">
        <div>
          <div className="page-eyebrow">PassQuantum / Generator</div>
          <h1>Password generator</h1>
          <div className="page-sub">Cryptographically random. Entropy is computed from your chosen character set.</div>
        </div>
        <Pill tone="accent" dot>CSPRNG · crypto.getRandomValues</Pill>
      </div>

      <Card eyebrow="OUTPUT" title="Generated password" right={
        <div style={{ display: "flex", gap: 8 }}>
          <Button size="sm" kind="ghost" leadingIcon={<Icons.Refresh size={14} />} onClick={regen}>Regenerate</Button>
        </div>
      }>
        <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
          <div style={{
            fontFamily: "var(--font-mono)",
            fontSize: 22,
            letterSpacing: "0.02em",
            padding: "18px 20px",
            background: "var(--bg-1)",
            border: "1px solid var(--line-2)",
            borderRadius: 8,
            wordBreak: "break-all",
            color: pwd ? "var(--fg-0)" : "var(--fg-3)",
            position: "relative",
          }}>
            {pwd || "—"}
          </div>
          <StrengthBlock score={score} />
          <div style={{ display: "flex", gap: 8, justifyContent: "flex-end" }}>
            <Button leadingIcon={<Icons.Copy size={14} />}>Copy</Button>
            <Button kind="primary" leadingIcon={<Icons.Vault size={14} />} onClick={() => setSavingOpen(true)}>
              Save to vault
            </Button>
          </div>
        </div>
      </Card>

      <Card eyebrow="PARAMETERS" title="Options">
        <div style={{ display: "flex", flexDirection: "column", gap: 20 }}>
          <Field label="Length" right={<span className="mono" style={{ fontSize: 11, color: "var(--fg-1)" }}>{length}</span>}>
            <div style={{ display: "flex", alignItems: "center", gap: 14 }}>
              <input
                type="range"
                min={8} max={64}
                value={length}
                onChange={(e) => setLength(Number(e.target.value))}
                style={{ flex: 1, accentColor: "var(--accent)" }}
              />
              <input
                type="number"
                value={length}
                onChange={(e) => setLength(Math.max(4, Math.min(128, Number(e.target.value) || 0)))}
                className="input input-mono"
                style={{ width: 76, textAlign: "right" }}
              />
            </div>
          </Field>
          <Field label="Character classes">
            <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 10 }}>
              <Check checked={opts.upper}    onChange={(v) => setOpts({ ...opts, upper: v })}>Uppercase  A–Z</Check>
              <Check checked={opts.lower}    onChange={(v) => setOpts({ ...opts, lower: v })}>Lowercase  a–z</Check>
              <Check checked={opts.digits}   onChange={(v) => setOpts({ ...opts, digits: v })}>Digits  0–9</Check>
              <Check checked={opts.symbols}  onChange={(v) => setOpts({ ...opts, symbols: v })}>Symbols  !@#$%^&*</Check>
              <Check checked={opts.excludeAmbiguous} onChange={(v) => setOpts({ ...opts, excludeAmbiguous: v })}>Exclude ambiguous (iIl1Lo0O)</Check>
            </div>
          </Field>
        </div>
      </Card>

      {savingOpen && (
        <SaveToVaultModal
          vaults={vaults}
          activeVaultId={activeVaultId}
          password={pwd}
          onClose={() => setSavingOpen(false)}
          onSave={(data) => { onSaveToVault?.(data); setSavingOpen(false); }}
        />
      )}
    </>
  );
}

function SaveToVaultModal({ vaults, activeVaultId, password, onClose, onSave }) {
  const [vault, setVault] = React.useState(activeVaultId);
  const [service, setService] = React.useState("");
  const [username, setUsername] = React.useState("");
  return (
    <div className="scrim" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <header className="modal-hd">
          <div className="modal-eyebrow">SAVE TO VAULT</div>
          <h3>New password entry</h3>
        </header>
        <div className="modal-body">
          <Field label="Vault">
            <Select value={vault} onChange={setVault}>
              {vaults.map(v => <option key={v.id} value={v.id}>{v.name}</option>)}
            </Select>
          </Field>
          <Field label="Service">
            <TextInput value={service} onChange={setService} placeholder="github.com" />
          </Field>
          <Field label="Username / email">
            <TextInput value={username} onChange={setUsername} placeholder="you@example.com" />
          </Field>
          <Field label="Generated password">
            <div className="input input-mono" style={{ background: "var(--bg-1)", color: "var(--fg-1)", fontSize: 13, userSelect: "all" }}>{password || "—"}</div>
          </Field>
        </div>
        <footer className="modal-ft">
          <Button kind="ghost" onClick={onClose}>Cancel</Button>
          <Button kind="primary" leadingIcon={<Icons.Check size={14} />} onClick={() => onSave({ vault, service, username, password })}>
            Save entry
          </Button>
        </footer>
      </div>
    </div>
  );
}

// ─────────────────────────────────────────────────────────────────────────────
// CHECKER
// ─────────────────────────────────────────────────────────────────────────────
function ScreenChecker() {
  const [pwd, setPwd] = React.useState("");
  const [show, setShow] = React.useState(false);
  const score = scorePassword(pwd);
  return (
    <>
      <div className="page-head">
        <div>
          <div className="page-eyebrow">PassQuantum / Check password</div>
          <h1>Strength analyzer</h1>
          <div className="page-sub">Audit a password offline. Nothing leaves this device — no HIBP lookup, no telemetry.</div>
        </div>
        <Pill tone="ok">Offline · local only</Pill>
      </div>

      <Card eyebrow="INPUT" title="Enter password to analyze">
        <div className="input-wrap">
          <input
            type={show ? "text" : "password"}
            className="input input-mono"
            value={pwd}
            onChange={(e) => setPwd(e.target.value)}
            placeholder="Enter password…"
            autoComplete="off"
          />
          <div className="input-affix">
            <IconBtn
              icon={show ? <Icons.EyeOff /> : <Icons.Eye />}
              onClick={() => setShow(!show)}
              title={show ? "Hide" : "Show"}
            />
          </div>
        </div>
      </Card>

      <Card eyebrow="ANALYSIS" title="Strength result">
        <div style={{ display: "flex", flexDirection: "column", gap: 18 }}>
          <StrengthBlock score={score} />
          <div className="divider" />
          <dl className="kv">
            <dt>Length</dt>
            <dd className="mono">{pwd.length} characters</dd>
            <dt>Character set</dt>
            <dd className="mono">
              {[
                /[A-Z]/.test(pwd) && "A–Z",
                /[a-z]/.test(pwd) && "a–z",
                /\d/.test(pwd) && "0–9",
                /[^A-Za-z0-9]/.test(pwd) && "symbols",
              ].filter(Boolean).join(" · ") || "—"}
            </dd>
            <dt>Estimated entropy</dt>
            <dd className="mono">{score.entropy} bits</dd>
            <dt>Brute-force estimate</dt>
            <dd className="mono">{score.crack}</dd>
          </dl>
          {!pwd && (
            <div style={{ fontSize: 12, color: "var(--fg-2)", fontFamily: "var(--font-mono)" }}>
              Start typing to analyze password strength.
            </div>
          )}
        </div>
      </Card>
    </>
  );
}

// ─────────────────────────────────────────────────────────────────────────────
// SETTINGS
// ─────────────────────────────────────────────────────────────────────────────
function ScreenSettings({ watching, setWatching, monitoredApps, setMonitoredApps }) {
  const [tab, setTab] = React.useState("security");
  return (
    <>
      <div className="page-head">
        <div>
          <div className="page-eyebrow">PassQuantum / Settings</div>
          <h1>Settings</h1>
          <div className="page-sub">Security, vault management, appearance, and application info.</div>
        </div>
      </div>

      <Tabs
        items={[
          { value: "security", label: "Security" },
          { value: "vaults",   label: "Vaults" },
          { value: "visuals",  label: "Appearance" },
          { value: "about",    label: "About" },
        ]}
        value={tab}
        onChange={setTab}
      />

      {tab === "security" && <SettingsSecurity watching={watching} setWatching={setWatching} monitoredApps={monitoredApps} setMonitoredApps={setMonitoredApps} />}
      {tab === "vaults" && <SettingsVaults />}
      {tab === "visuals" && <SettingsVisuals />}
      {tab === "about" && <SettingsAbout />}
    </>
  );
}

function SettingsSecurity({ watching, setWatching, monitoredApps, setMonitoredApps }) {
  return (
    <div style={{ display: "flex", flexDirection: "column", gap: "var(--space-5)" }}>
      <Card eyebrow="MASTER PASSWORD" title="Vault access">
        <div style={{ display: "grid", gridTemplateColumns: "1fr auto", gap: "var(--space-5)", alignItems: "center" }}>
          <div>
            <div style={{ fontSize: 13, color: "var(--fg-1)" }}>App-level verifier active and bound to the current Kyber-768 keypair.</div>
            <div className="mono" style={{ fontSize: 11, color: "var(--fg-2)", marginTop: 6 }}>
              Last rotated · 47 days ago · Strength <span style={{ color: "var(--ok)" }}>Strong</span>
            </div>
          </div>
          <Button leadingIcon={<Icons.Refresh size={14} />}>Change master password</Button>
        </div>
      </Card>

      <Card eyebrow="AUTO-LOCK" title="Session controls">
        <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
          <Field label="Lock vault after inactivity">
            <Select value="5">
              <option value="1">1 minute</option>
              <option value="5">5 minutes</option>
              <option value="15">15 minutes</option>
              <option value="60">1 hour</option>
              <option value="never">Never (not recommended)</option>
            </Select>
          </Field>
          <Check checked onChange={() => {}}>Lock on system sleep / lid close</Check>
          <Check checked onChange={() => {}}>Clear clipboard 30 seconds after copy</Check>
          <Check checked={false} onChange={() => {}}>Require password to view (not just copy) secrets</Check>
        </div>
      </Card>

      <Card eyebrow="PRESENCE GUARD" title="Face-detection auto-kill">
        <div style={{ display: "grid", gridTemplateColumns: "1fr auto", gap: "var(--space-4)", alignItems: "flex-start", marginBottom: 16 }}>
          <div>
            <div style={{ fontSize: 13, color: "var(--fg-1)" }}>
              Watch the webcam for your face. When you walk away, force-close the apps listed below.
            </div>
            <div className="mono" style={{ fontSize: 11, color: "var(--fg-2)", marginTop: 6 }}>
              Detection model · MediaPipe Face Mesh · Local inference
            </div>
          </div>
          <label className="check" style={{ flexDirection: "row-reverse" }}>
            <span style={{ color: watching ? "var(--fg-0)" : "var(--fg-2)" }}>{watching ? "Watching" : "Idle"}</span>
            <ToggleSwitch on={watching} onChange={setWatching} />
          </label>
        </div>

        <div style={{
          background: "rgba(217,144,48,0.10)",
          border: "1px solid rgba(217,144,48,0.30)",
          borderRadius: 8,
          padding: "12px 14px",
          display: "grid",
          gridTemplateColumns: "auto 1fr",
          gap: 10,
          marginBottom: 16,
        }}>
          <div style={{ color: "var(--warn)" }}><Icons.AlertTriangle size={16} /></div>
          <div style={{ fontSize: 12, color: "var(--fg-1)" }}>
            <strong style={{ color: "#f0d4a2", letterSpacing: "0.04em" }}>FORCE-KILL WARNING</strong>
            <div style={{ marginTop: 4 }}>
              Processes below will be killed with SIGKILL — no save prompts — as soon as your face is not detected for 5 seconds. Make sure unsaved work is acceptable to lose.
            </div>
          </div>
        </div>

        <div style={{ display: "flex", flexDirection: "column", gap: 6 }}>
          {monitoredApps.map((a, i) => (
            <label key={a.process} className={"check" + (a.enabled ? " on" : "")}
                   onClick={() => setMonitoredApps(monitoredApps.map((m, j) => j === i ? { ...m, enabled: !m.enabled } : m))}
                   style={{ padding: "10px 12px", background: "var(--bg-1)", border: "1px solid var(--line-1)", borderRadius: 6, gap: 12 }}>
              <input type="checkbox" checked={a.enabled} readOnly />
              <span className="box"><Icons.Check /></span>
              <span style={{ flex: 1 }}>
                <span style={{ display: "block", fontSize: 13, color: "var(--fg-0)" }}>{a.label}</span>
                <span className="mono" style={{ fontSize: 11, color: "var(--fg-2)" }}>{a.process}</span>
              </span>
            </label>
          ))}
          <button className="btn btn-ghost" style={{ alignSelf: "flex-start", marginTop: 6 }}>
            <Icons.Plus size={14} /> Add process to monitor
          </button>
        </div>
      </Card>
    </div>
  );
}

function ToggleSwitch({ on, onChange }) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={!!on}
      onClick={() => onChange(!on)}
      style={{
        width: 38, height: 22, borderRadius: 999, position: "relative",
        border: "1px solid " + (on ? "var(--accent-line)" : "var(--line-2)"),
        background: on ? "var(--accent-soft)" : "var(--bg-1)",
        padding: 0, cursor: "default", transition: "all .15s",
      }}
    >
      <span style={{
        position: "absolute", top: 2, left: on ? 18 : 2,
        width: 16, height: 16, borderRadius: 999,
        background: on ? "var(--accent)" : "var(--fg-3)",
        transition: "left .15s, background .15s",
      }}/>
    </button>
  );
}

function SettingsVaults() {
  return (
    <div style={{ display: "flex", flexDirection: "column", gap: "var(--space-5)" }}>
      <Card eyebrow="ACTIVE VAULT" title="Default">
        <dl className="kv">
          <dt>Vault file</dt>
          <dd className="mono">~/.config/passquantum/default.enc</dd>
          <dt>Size on disk</dt>
          <dd className="mono">14.2 KB</dd>
          <dt>Items</dt>
          <dd className="mono">3 entries</dd>
          <dt>Last modified</dt>
          <dd className="mono">2 minutes ago</dd>
        </dl>
      </Card>

      <Card eyebrow="MAINTENANCE" title="Compact &amp; verify">
        <div style={{ display: "grid", gridTemplateColumns: "1fr auto", gap: "var(--space-4)", alignItems: "center" }}>
          <div style={{ fontSize: 13, color: "var(--fg-1)" }}>
            Rewrites the vault to reclaim deleted-item space and re-MACs every block. Run after bulk deletes.
          </div>
          <Button>Compact vault</Button>
        </div>
      </Card>

      <Card eyebrow="BACKUP &amp; RESTORE" title="Encrypted exports">
        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "var(--space-4)" }}>
          <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
            <div className="eyebrow">Export</div>
            <Button leadingIcon={<Icons.Download size={14} />}>Export vault (.pqv)</Button>
            <Button kind="ghost" leadingIcon={<Icons.Download size={14} />}>Plain JSON (decrypted)</Button>
          </div>
          <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
            <div className="eyebrow">Import</div>
            <Button leadingIcon={<Icons.Upload size={14} />}>Import vault (.pqv)</Button>
            <Button kind="ghost" leadingIcon={<Icons.Upload size={14} />}>Migrate from 1Password / KeePass</Button>
          </div>
        </div>
        <div className="divider" />
        <div style={{ display: "grid", gridTemplateColumns: "1fr auto", gap: 12, alignItems: "center" }}>
          <div>
            <div style={{ fontSize: 13, color: "var(--fg-1)" }}>Automatic backup</div>
            <div className="mono" style={{ fontSize: 11, color: "var(--fg-2)", marginTop: 4 }}>Last backup · 2 minutes ago · Retains last 7</div>
          </div>
          <Button leadingIcon={<Icons.Refresh size={14} />}>Back up now</Button>
        </div>
      </Card>
    </div>
  );
}

function SettingsVisuals() {
  return (
    <div style={{ display: "flex", flexDirection: "column", gap: "var(--space-5)" }}>
      <Card eyebrow="DENSITY" title="Layout density">
        <div style={{ fontSize: 13, color: "var(--fg-1)", marginBottom: 12 }}>
          Use the Tweaks panel (lower-right) to switch density, accent, typography, and background pattern live.
        </div>
        <div className="mono" style={{ fontSize: 11, color: "var(--fg-2)" }}>
          tweaks · density · accent · font · sidebar mode · bg-texture · card layout
        </div>
      </Card>

      <Card eyebrow="ACCENTS" title="Accent palette">
        <div style={{ display: "flex", gap: 12, flexWrap: "wrap" }}>
          {[
            { name: "Blue", color: "#3b82f6" },
            { name: "Emerald", color: "#10b981" },
            { name: "Gold", color: "#c8a464" },
            { name: "Mono", color: "#e6e6e6" },
            { name: "Cyan", color: "#06b6d4" },
          ].map((a) => (
            <div key={a.name}>
              <div className="swatch" style={{ background: a.color }} />
              <div className="swatch-label">{a.name}</div>
              <div className="mono" style={{ fontSize: 11, color: "var(--fg-1)" }}>{a.color}</div>
            </div>
          ))}
        </div>
      </Card>

      <Card eyebrow="ICON" title="App icon">
        <div style={{ display: "grid", gridTemplateColumns: "auto 1fr auto", gap: 16, alignItems: "center" }}>
          <div style={{
            width: 64, height: 64,
            borderRadius: 14,
            background: "var(--accent-soft)",
            border: "1px solid var(--accent-line)",
            display: "grid", placeItems: "center",
            color: "var(--accent)",
          }}>
            <Icons.ShieldCheck size={28} />
          </div>
          <div>
            <div style={{ fontSize: 13, color: "var(--fg-0)" }}>passquantum.png</div>
            <div className="mono" style={{ fontSize: 11, color: "var(--fg-2)", marginTop: 4 }}>512 × 512 · PNG · 12 KB</div>
          </div>
          <div style={{ display: "flex", gap: 8 }}>
            <Button>Change</Button>
            <Button kind="ghost">Reset</Button>
          </div>
        </div>
      </Card>
    </div>
  );
}

function SettingsAbout() {
  return (
    <div style={{ display: "flex", flexDirection: "column", gap: "var(--space-5)" }}>
      <Card padded={false}>
        <div style={{ padding: "var(--space-6)", display: "grid", gridTemplateColumns: "auto 1fr", gap: 24, alignItems: "center" }}>
          <div style={{
            width: 96, height: 96, borderRadius: 18,
            background: "linear-gradient(135deg, var(--accent-soft), transparent)",
            border: "1px solid var(--accent-line)",
            display: "grid", placeItems: "center",
            color: "var(--accent)",
          }}>
            <Icons.Atom size={48} />
          </div>
          <div>
            <div className="eyebrow">PRODUCT</div>
            <h2 style={{ margin: "4px 0 6px", fontSize: 22, fontWeight: 600, letterSpacing: "-0.015em" }}>PassQuantum</h2>
            <div className="mono" style={{ color: "var(--fg-1)", fontSize: 12 }}>v1.0.0 · linux-amd64 · go1.22</div>
            <div style={{ marginTop: 10, color: "var(--fg-1)", fontSize: 13, maxWidth: 540 }}>
              A post-quantum credential vault. Symmetric encryption with AES-256-GCM, key encapsulation with ML-KEM-768 (Kyber), and an Argon2id-derived master key.
            </div>
          </div>
        </div>
      </Card>

      <Card eyebrow="CRYPTOGRAPHY" title="Security stack">
        <dl className="kv">
          <dt>Bulk encryption</dt>
          <dd>AES-256-GCM<small>96-bit IV · 128-bit auth tag · per-record nonce</small></dd>
          <dt>Key encapsulation</dt>
          <dd>ML-KEM-768 (Kyber)<small>NIST FIPS 203 · IND-CCA2 secure against quantum adversaries</small></dd>
          <dt>Password KDF</dt>
          <dd>Argon2id<small>m=64 MiB · t=3 · p=4 · 16-byte salt</small></dd>
          <dt>Authentication</dt>
          <dd>HMAC-SHA-512<small>Per-block integrity</small></dd>
          <dt>Random source</dt>
          <dd>crypto/rand (Go)<small>OS CSPRNG · getrandom(2) on Linux</small></dd>
          <dt>Architecture</dt>
          <dd>Zero-knowledge, offline-first<small>No telemetry. No cloud sync. No network calls.</small></dd>
        </dl>
      </Card>

      <Card eyebrow="META" title="About this build">
        <dl className="kv">
          <dt>License</dt>
          <dd>MIT</dd>
          <dt>Build hash</dt>
          <dd className="mono">2f9c4a1 · 2025-05-24</dd>
          <dt>Maintainer</dt>
          <dd>PassQuantum Team</dd>
          <dt>Repository</dt>
          <dd className="mono" style={{ display: "inline-flex", alignItems: "center", gap: 6 }}>
            github.com/passquantum/passquantum
            <Icons.ExternalLink size={12} />
          </dd>
        </dl>
        <div className="divider" />
        <div style={{ display: "flex", gap: 8 }}>
          <Button leadingIcon={<Icons.ExternalLink size={14} />}>Documentation</Button>
          <Button kind="ghost" leadingIcon={<Icons.Refresh size={14} />}>Check for updates</Button>
        </div>
      </Card>
    </div>
  );
}

Object.assign(window, {
  ScreenVaults, ScreenAddItem, ScreenItems, ScreenGenerator, ScreenChecker, ScreenSettings,
});
