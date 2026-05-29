/* PassQuantum — Background Service Worker */

const API_BASE = 'http://127.0.0.1:8765';
const PENDING_TTL_MS = 2 * 60 * 1000; // 2 minutes

// Pending credentials survive page navigation (multi-step login flows like GitHub 2FA).
// Keyed by normalized domain. Cleared on retrieval or after TTL.
const pendingCredentials = {};

// --- Authenticated fetch wrapper ---

async function getSecret() {
  const result = await browser.storage.local.get('pqSecret');
  return result.pqSecret || null;
}

async function apiFetch(path, options = {}) {
  const secret = await getSecret();
  if (!secret) {
    throw new Error('not_paired');
  }

  const headers = {
    'X-Secret': secret,
    'Content-Type': 'application/json',
    ...(options.headers || {}),
  };

  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
  });

  if (response.status === 401) {
    await browser.storage.local.remove('pqSecret');
    throw new Error('auth_failed');
  }
  if (response.status === 423) {
    throw new Error('vault_locked');
  }
  if (!response.ok) {
    const body = await response.json().catch(() => ({}));
    throw new Error(body.error || `HTTP ${response.status}`);
  }
  return response;
}

// --- Status check (no auth required) ---

async function checkStatus() {
  try {
    const resp = await fetch(`${API_BASE}/vault/status`, {
      headers: { 'Content-Type': 'application/json' },
    });
    return await resp.json();
  } catch {
    return { unlocked: false, error: 'unreachable' };
  }
}

// --- Pairing ---

async function initiatePairing() {
  const resp = await fetch(`${API_BASE}/vault/pair`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({}),
  });
  return await resp.json();
}

async function completePairing(token) {
  const resp = await fetch(`${API_BASE}/vault/pair`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ token }),
  });

  if (!resp.ok) {
    const body = await resp.json().catch(() => ({}));
    throw new Error(body.error || 'Pairing failed');
  }

  const data = await resp.json();
  if (data.secret) {
    await browser.storage.local.set({ pqSecret: data.secret });
  }
  return data;
}

// --- Message handler from content script / popup ---

browser.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  const handler = messageHandlers[msg.type];
  if (handler) {
    handler(msg, sender).then(sendResponse).catch(err => {
      sendResponse({ error: err.message });
    });
    return true;
  }
});

const messageHandlers = {
  CHECK_STATUS: async () => {
    const secret = await getSecret();
    const status = await checkStatus();
    return { ...status, paired: !!secret };
  },

  INITIATE_PAIRING: async () => {
    return await initiatePairing();
  },

  COMPLETE_PAIRING: async (msg) => {
    return await completePairing(msg.token);
  },

  FORM_SUBMITTED: async (msg) => {
    try {
      const secret = await getSecret();
      if (!secret) return { action: 'skip', reason: 'not_paired' };

      const domain = normalizeDomain(msg.domain);

      // Check never-save list
      const nsResp = await apiFetch('/vault/never-save');
      const { domains } = await nsResp.json();
      if (domains.includes(domain)) {
        return { action: 'skip', reason: 'never_save' };
      }

      // Check if credentials exist
      const existsResp = await apiFetch(`/vault/exists?domain=${encodeURIComponent(domain)}`);
      const { found, credentials } = await existsResp.json();

      let result;
      if (found) {
        const match = credentials.find(c => c.username === msg.username);
        if (match) {
          result = { action: 'prompt_update', entryId: match.id, username: match.username, domain };
        } else {
          result = { action: 'prompt_save', domain, username: msg.username };
        }
      } else {
        result = { action: 'prompt_save', domain, username: msg.username };
      }

      // Store pending credentials so the banner survives page navigation
      // (e.g., GitHub login → 2FA page)
      if (result.action === 'prompt_save' || result.action === 'prompt_update') {
        pendingCredentials[domain] = {
          domain,
          username: msg.username,
          password: msg.password,
          action: result.action,
          entryId: result.entryId,
          timestamp: Date.now(),
        };
      }

      return result;
    } catch (err) {
      return { action: 'skip', reason: err.message };
    }
  },

  CHECK_PENDING: async (msg) => {
    const domain = normalizeDomain(msg.domain);
    const entry = pendingCredentials[domain];
    if (!entry) return { hasPending: false };

    if (Date.now() - entry.timestamp > PENDING_TTL_MS) {
      delete pendingCredentials[domain];
      return { hasPending: false };
    }

    // Return but don't delete yet — content script will clear via DISMISS_PENDING
    // after the user interacts with the banner
    return {
      hasPending: true,
      action: entry.action,
      domain: entry.domain,
      username: entry.username,
      password: entry.password,
      entryId: entry.entryId,
    };
  },

  DISMISS_PENDING: async (msg) => {
    const domain = normalizeDomain(msg.domain);
    delete pendingCredentials[domain];
    return { cleared: true };
  },

  SAVE_CREDENTIAL: async (msg) => {
    const resp = await apiFetch('/vault/save', {
      method: 'POST',
      body: JSON.stringify({
        domain: msg.domain,
        username: msg.username,
        password: msg.password,
      }),
    });
    return await resp.json();
  },

  UPDATE_CREDENTIAL: async (msg) => {
    const resp = await apiFetch(`/vault/update/${msg.entryId}`, {
      method: 'PUT',
      body: JSON.stringify({ password: msg.password }),
    });
    return await resp.json();
  },

  NEVER_SAVE: async (msg) => {
    const resp = await apiFetch('/vault/never-save', {
      method: 'POST',
      body: JSON.stringify({ domain: msg.domain }),
    });
    return await resp.json();
  },

  REMOVE_NEVER_SAVE: async (msg) => {
    const resp = await apiFetch(`/vault/never-save?domain=${encodeURIComponent(msg.domain)}`, {
      method: 'DELETE',
    });
    return await resp.json();
  },

  GET_NEVER_SAVE_LIST: async () => {
    const resp = await apiFetch('/vault/never-save');
    return await resp.json();
  },
};

// --- Domain normalization (mirrors Go logic) ---

function normalizeDomain(raw) {
  if (!raw) return '';
  try {
    if (!raw.includes('://')) raw = 'https://' + raw;
    const url = new URL(raw);
    let host = url.hostname;
    if (!host) return raw;
    if (host === 'localhost') return host;
    if (/^\d+\.\d+\.\d+\.\d+$/.test(host)) return host;

    const parts = host.split('.');
    if (parts.length <= 2) return host;

    const knownSLDs = ['co', 'com', 'org', 'net', 'ac', 'gov', 'edu', 'mil'];
    const last = parts[parts.length - 1];
    const secondLast = parts[parts.length - 2];

    if (last.length === 2 && knownSLDs.includes(secondLast)) {
      return parts.slice(-3).join('.');
    }
    return parts.slice(-2).join('.');
  } catch {
    return raw;
  }
}
