/* PassQuantum — Content Script: Form Detection + Save Banner */

(function () {
  'use strict';

  const BANNER_TIMEOUT_MS = 15000;
  const FEEDBACK_DURATION_MS = 2000;
  const ERROR_DURATION_MS = 5000;

  const ERROR_MESSAGES = {
    vault_locked: 'PassQuantum is locked — open a vault first',
    auth_failed: 'Re-pair the extension with PassQuantum',
    not_paired: 'Extension is not paired',
  };
  let activeBanner = null;
  let pendingCredentials = null;

  // --- Form Detection ---

  const USERNAME_PATTERNS = /user|email|login|account|identifier|handle/i;
  const REGISTRATION_PATTERNS = /register|signup|sign.up|create|join|enroll/i;
  const SUBMIT_TEXT_PATTERNS = /sign.?in|log.?in|iniciar|entrar|submit|continue|next/i;

  function classifyForm(form) {
    const passwordFields = form.querySelectorAll('input[type="password"]');
    if (passwordFields.length === 0) return null;

    const visiblePasswords = Array.from(passwordFields).filter(isVisible);
    if (visiblePasswords.length === 0) return null;

    // Registration: 2+ password fields or new-password autocomplete
    if (visiblePasswords.length >= 2) return null;

    const hasNewPassword = Array.from(visiblePasswords).some(
      f => f.autocomplete === 'new-password'
    );
    if (hasNewPassword) return null;

    // Check form attributes for registration hints
    const formAttrs = (form.id + ' ' + form.className + ' ' + (form.action || '')).toLowerCase();
    if (REGISTRATION_PATTERNS.test(formAttrs)) return null;

    return 'login';
  }

  function findUsernameField(form, passwordField) {
    // Priority 1: autocomplete hints
    const autoUsername = form.querySelector(
      'input[autocomplete="username"], input[autocomplete="email"]'
    );
    if (autoUsername && isVisible(autoUsername)) return autoUsername;

    // Priority 2: type=email
    const emailInput = form.querySelector('input[type="email"]');
    if (emailInput && isVisible(emailInput)) return emailInput;

    // Priority 3: name/id pattern match
    const textInputs = Array.from(
      form.querySelectorAll('input[type="text"], input[type="email"], input:not([type])')
    ).filter(isVisible);

    for (const input of textInputs) {
      const attrs = (input.name + ' ' + input.id + ' ' + (input.placeholder || '')).toLowerCase();
      if (USERNAME_PATTERNS.test(attrs)) return input;
    }

    // Priority 4: nearest visible text input before password field
    const allInputs = Array.from(form.querySelectorAll('input')).filter(isVisible);
    const pwIndex = allInputs.indexOf(passwordField);
    for (let i = pwIndex - 1; i >= 0; i--) {
      const inp = allInputs[i];
      if (inp.type === 'text' || inp.type === 'email' || !inp.type) return inp;
    }

    return null;
  }

  function isVisible(el) {
    if (!el) return false;
    const style = getComputedStyle(el);
    return (
      style.display !== 'none' &&
      style.visibility !== 'hidden' &&
      el.offsetParent !== null &&
      el.type !== 'hidden'
    );
  }

  // --- Form Interception ---

  function instrumentForm(form) {
    if (form.dataset.pqInstrumented) return;
    form.dataset.pqInstrumented = 'true';

    const formType = classifyForm(form);
    if (formType !== 'login') return;

    form.addEventListener(
      'submit',
      (e) => captureCredentials(form),
      true
    );

    // Also catch button clicks for SPAs that don't use form.submit()
    const buttons = form.querySelectorAll(
      'button[type="submit"], button:not([type]), input[type="submit"]'
    );
    buttons.forEach((btn) => {
      btn.addEventListener(
        'click',
        () => setTimeout(() => captureCredentials(form), 0),
        true
      );
    });
  }

  function captureCredentials(form) {
    const passwordFields = Array.from(
      form.querySelectorAll('input[type="password"]')
    ).filter(isVisible);

    const passwordField = passwordFields.find((f) => f.value.length > 0);
    if (!passwordField) return;

    const usernameField = findUsernameField(form, passwordField);
    const username = usernameField ? usernameField.value.trim() : '';
    const password = passwordField.value;

    if (!password) return;

    const creds = {
      domain: location.hostname,
      username,
      password,
    };
    pendingCredentials = creds;

    browser.runtime.sendMessage({
      type: 'FORM_SUBMITTED',
      domain: location.hostname,
      username,
      password,
    }).then((response) => {
      if (!response || response.action === 'skip') return;
      if (response.action === 'prompt_save') {
        showBanner('save', response.domain, response.username, null, creds);
      } else if (response.action === 'prompt_update') {
        showBanner('update', response.domain, response.username, response.entryId, creds);
      }
    }).catch(() => {
      // PassQuantum not reachable — fail silently
    });
  }

  // --- Scan & Observe ---

  function scanForms() {
    document.querySelectorAll('form').forEach(instrumentForm);
  }

  const observer = new MutationObserver((mutations) => {
    for (const mutation of mutations) {
      for (const node of mutation.addedNodes) {
        if (node.nodeType !== Node.ELEMENT_NODE) continue;
        if (node.tagName === 'FORM') {
          instrumentForm(node);
        } else if (node.querySelectorAll) {
          node.querySelectorAll('form').forEach(instrumentForm);
        }
      }
    }
  });

  scanForms();
  observer.observe(document.body, { childList: true, subtree: true });

  // Check if there are pending credentials from a previous page (multi-step login)
  checkPendingCredentials();

  function checkPendingCredentials() {
    browser.runtime.sendMessage({
      type: 'CHECK_PENDING',
      domain: location.hostname,
    }).then((response) => {
      if (!response || !response.hasPending) return;

      const creds = {
        domain: response.domain,
        username: response.username,
        password: response.password,
      };
      pendingCredentials = creds;

      showBanner(
        response.action === 'prompt_update' ? 'update' : 'save',
        response.domain,
        response.username,
        response.entryId,
        creds
      );
    }).catch(() => {});
  }

  // --- Banner (Shadow DOM) ---

  function showBanner(action, domain, username, entryId, credentials) {
    removeBanner();
    // Capture credentials in closure so they survive any later removeBanner() calls
    const bannerCredentials = credentials || pendingCredentials;

    const host = document.createElement('div');
    host.id = 'passquantum-banner-host';
    host.style.cssText =
      'position:fixed!important;top:16px!important;right:16px!important;z-index:2147483647!important;' +
      'font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif!important;';

    const shadow = host.attachShadow({ mode: 'closed' });

    const truncatedUser = username.length > 30 ? username.slice(0, 27) + '...' : username;
    const title = action === 'save' ? 'Save password?' : 'Update password?';
    const btnLabel = action === 'save' ? 'Save' : 'Update';

    shadow.innerHTML = `
      <style>
        :host { all: initial; }
        .banner {
          background: #1a1a2e;
          color: #e0e0e0;
          border: 1px solid #3a86ff;
          border-radius: 12px;
          padding: 16px 20px;
          min-width: 320px;
          max-width: 400px;
          box-shadow: 0 8px 32px rgba(0,0,0,0.4);
          font-size: 14px;
          line-height: 1.4;
          cursor: move;
          user-select: none;
          transition: background 0.3s, border-color 0.3s;
        }
        .banner.success {
          background: #0d3320;
          border-color: #22c55e;
        }
        .banner.error {
          background: #2d0f0f;
          border-color: #ef4444;
        }
        .title {
          font-weight: 600;
          font-size: 15px;
          margin-bottom: 4px;
          color: #ffffff;
          display: flex;
          align-items: center;
          gap: 8px;
        }
        .title svg { flex-shrink: 0; }
        .detail {
          color: #a0a0b0;
          font-size: 13px;
          margin-bottom: 12px;
        }
        .detail strong { color: #c0c0d0; }
        .buttons {
          display: flex;
          gap: 8px;
        }
        .btn {
          padding: 6px 16px;
          border-radius: 6px;
          border: none;
          cursor: pointer;
          font-size: 13px;
          font-weight: 500;
          transition: opacity 0.15s;
        }
        .btn:hover { opacity: 0.85; }
        .btn-primary {
          background: #3a86ff;
          color: #fff;
        }
        .btn-secondary {
          background: #2a2a3e;
          color: #a0a0b0;
          border: 1px solid #3a3a4e;
        }
        .btn-danger {
          background: transparent;
          color: #888;
          border: 1px solid #3a3a4e;
        }
        .feedback {
          display: none;
          align-items: center;
          gap: 8px;
          color: #22c55e;
          font-weight: 500;
        }
        .error-msg {
          display: none;
          align-items: center;
          gap: 8px;
          color: #ef4444;
          font-weight: 500;
          font-size: 13px;
        }
        .error-msg svg { flex-shrink: 0; }
        .banner.success .buttons { display: none; }
        .banner.success .feedback { display: flex; }
        .banner.error .buttons { display: none; }
        .banner.error .error-msg { display: flex; }
        .checkmark {
          width: 20px;
          height: 20px;
          animation: pop 0.3s ease-out;
        }
        @keyframes pop {
          0% { transform: scale(0); }
          80% { transform: scale(1.2); }
          100% { transform: scale(1); }
        }
      </style>
      <div class="banner" id="banner">
        <div class="title">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#3a86ff" stroke-width="2">
            <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/>
            <path d="M7 11V7a5 5 0 0 1 10 0v4"/>
          </svg>
          ${title}
        </div>
        <div class="detail">
          <strong>${escapeHtml(truncatedUser)}</strong> on <strong>${escapeHtml(domain)}</strong>
        </div>
        <div class="buttons">
          <button type="button" class="btn btn-primary" id="btn-confirm">${btnLabel}</button>
          <button type="button" class="btn btn-secondary" id="btn-dismiss">Not now</button>
          <button type="button" class="btn btn-danger" id="btn-never">Never for this site</button>
        </div>
        <div class="feedback">
          <svg class="checkmark" viewBox="0 0 24 24" fill="none" stroke="#22c55e" stroke-width="3">
            <polyline points="20 6 9 17 4 12"/>
          </svg>
          Saved to PassQuantum
        </div>
        <div class="error-msg" id="error-msg">
          <svg viewBox="0 0 24 24" fill="none" stroke="#ef4444" stroke-width="2" width="20" height="20">
            <circle cx="12" cy="12" r="10"/>
            <line x1="12" y1="8" x2="12" y2="12"/>
            <line x1="12" y1="16" x2="12.01" y2="16"/>
          </svg>
          <span id="error-text">Could not save</span>
        </div>
      </div>
    `;

    const banner = shadow.getElementById('banner');
    const btnConfirm = shadow.getElementById('btn-confirm');
    const btnDismiss = shadow.getElementById('btn-dismiss');
    const btnNever = shadow.getElementById('btn-never');

    // Auto-close after timeout
    let autoCloseTimer = setTimeout(removeBanner, BANNER_TIMEOUT_MS);
    const resetTimer = () => {
      clearTimeout(autoCloseTimer);
      autoCloseTimer = setTimeout(removeBanner, BANNER_TIMEOUT_MS);
    };

    // Confirm: save or update
    btnConfirm.addEventListener('click', (e) => {
      e.preventDefault();
      e.stopPropagation();
      clearTimeout(autoCloseTimer);
      if (!bannerCredentials) {
        console.warn('[PassQuantum] Save clicked but no credentials in banner closure');
        showError('No credentials to save');
        return;
      }

      btnConfirm.disabled = true;

      const msgType = action === 'save' ? 'SAVE_CREDENTIAL' : 'UPDATE_CREDENTIAL';
      const payload =
        action === 'save'
          ? {
              type: msgType,
              domain: bannerCredentials.domain,
              username: bannerCredentials.username,
              password: bannerCredentials.password,
            }
          : {
              type: msgType,
              entryId,
              password: bannerCredentials.password,
            };

      browser.runtime.sendMessage(payload).then((response) => {
        if (response && response.error) {
          console.error('[PassQuantum] Save failed:', response.error);
          btnConfirm.disabled = false;
          showError(ERROR_MESSAGES[response.error] || response.error);
          return;
        }

        if (!response || (action === 'save' && !response.saved) || (action === 'update' && !response.updated)) {
          console.error('[PassQuantum] Save returned unexpected response:', response);
          btnConfirm.disabled = false;
          showError('Unexpected response from PassQuantum');
          return;
        }

        browser.runtime.sendMessage({ type: 'DISMISS_PENDING', domain }).catch(() => {});
        banner.classList.add('success');
        setTimeout(removeBanner, FEEDBACK_DURATION_MS);
      }).catch((err) => {
        console.error('[PassQuantum] Save threw:', err);
        btnConfirm.disabled = false;
        showError('Could not reach PassQuantum');
      });
    });

    function showError(message) {
      const errorText = shadow.getElementById('error-text');
      if (errorText) errorText.textContent = message;
      banner.classList.add('error');
      setTimeout(removeBanner, ERROR_DURATION_MS);
    }

    // Dismiss — also clear pending so banner won't reappear on next navigation
    btnDismiss.addEventListener('click', (e) => {
      e.stopPropagation();
      browser.runtime.sendMessage({ type: 'DISMISS_PENDING', domain }).catch(() => {});
      removeBanner();
    });

    // Never save
    btnNever.addEventListener('click', (e) => {
      e.stopPropagation();
      browser.runtime.sendMessage({ type: 'NEVER_SAVE', domain });
      browser.runtime.sendMessage({ type: 'DISMISS_PENDING', domain }).catch(() => {});
      removeBanner();
    });

    // Draggable
    let isDragging = false;
    let dragOffsetX = 0;
    let dragOffsetY = 0;

    banner.addEventListener('mousedown', (e) => {
      if (e.target.tagName === 'BUTTON') return;
      isDragging = true;
      const rect = host.getBoundingClientRect();
      dragOffsetX = e.clientX - rect.left;
      dragOffsetY = e.clientY - rect.top;
      resetTimer();
    });

    document.addEventListener('mousemove', (e) => {
      if (!isDragging) return;
      host.style.left = (e.clientX - dragOffsetX) + 'px';
      host.style.top = (e.clientY - dragOffsetY) + 'px';
      host.style.right = 'auto';
    });

    document.addEventListener('mouseup', () => {
      isDragging = false;
    });

    document.body.appendChild(host);
    activeBanner = host;
  }

  function removeBanner() {
    if (activeBanner && activeBanner.parentNode) {
      activeBanner.parentNode.removeChild(activeBanner);
    }
    activeBanner = null;
    pendingCredentials = null;
  }

  function escapeHtml(str) {
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
  }
})();
