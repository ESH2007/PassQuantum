/* PassQuantum — Popup Script */

document.addEventListener('DOMContentLoaded', init);

async function init() {
  const statusDot = document.getElementById('status-dot');
  const statusText = document.getElementById('status-text');
  const vaultName = document.getElementById('vault-name');
  const pairingSection = document.getElementById('pairing-section');
  const connectedSection = document.getElementById('connected-section');
  const btnPair = document.getElementById('btn-pair');
  const pairingInput = document.getElementById('pairing-input');
  const pairCode = document.getElementById('pair-code');
  const btnConfirmPair = document.getElementById('btn-confirm-pair');
  const pairError = document.getElementById('pair-error');

  // Check status
  const status = await browser.runtime.sendMessage({ type: 'CHECK_STATUS' });

  if (status.error === 'unreachable') {
    statusDot.className = 'status-dot disconnected';
    statusText.textContent = 'PassQuantum not running';
    pairingSection.classList.add('hidden');
    connectedSection.classList.add('hidden');
    return;
  }

  if (!status.paired) {
    statusDot.className = 'status-dot disconnected';
    statusText.textContent = 'Not connected';
    pairingSection.classList.remove('hidden');
    connectedSection.classList.add('hidden');
  } else if (!status.app_unlocked) {
    statusDot.className = 'status-dot disconnected';
    statusText.textContent = 'PassQuantum is locked';
    vaultName.textContent = 'Enter your master password in PassQuantum';
    pairingSection.classList.add('hidden');
    connectedSection.classList.remove('hidden');
  } else if (!status.unlocked) {
    statusDot.className = 'status-dot warning';
    statusText.textContent = 'No vault open';
    vaultName.textContent = 'Select a vault in PassQuantum';
    pairingSection.classList.add('hidden');
    connectedSection.classList.remove('hidden');
  } else {
    statusDot.className = 'status-dot connected';
    statusText.textContent = 'Connected';
    if (status.vault) {
      vaultName.textContent = 'Vault: ' + status.vault;
    }
    pairingSection.classList.add('hidden');
    connectedSection.classList.remove('hidden');
    loadNeverSaveList();
  }

  // Pairing flow
  btnPair.addEventListener('click', async () => {
    btnPair.disabled = true;
    btnPair.textContent = 'Connecting...';

    try {
      await browser.runtime.sendMessage({ type: 'INITIATE_PAIRING' });
      pairingInput.classList.remove('hidden');
      btnPair.classList.add('hidden');
      pairCode.focus();
    } catch (err) {
      btnPair.disabled = false;
      btnPair.textContent = 'Connect to PassQuantum';
      showPairError('Could not reach PassQuantum');
    }
  });

  btnConfirmPair.addEventListener('click', async () => {
    const code = pairCode.value.trim();
    if (code.length !== 6) {
      showPairError('Enter the 6-digit code');
      return;
    }

    btnConfirmPair.disabled = true;
    pairError.classList.add('hidden');

    try {
      const result = await browser.runtime.sendMessage({
        type: 'COMPLETE_PAIRING',
        token: code,
      });

      if (result.error) {
        showPairError(result.error);
        btnConfirmPair.disabled = false;
        return;
      }

      if (result.secret || result.status === 'paired') {
        statusDot.className = 'status-dot connected';
        statusText.textContent = 'Connected';
        pairingSection.classList.add('hidden');
        connectedSection.classList.remove('hidden');
        loadNeverSaveList();
      }
    } catch (err) {
      showPairError('Pairing failed: ' + err.message);
      btnConfirmPair.disabled = false;
    }
  });

  pairCode.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') btnConfirmPair.click();
  });

  function showPairError(msg) {
    pairError.textContent = msg;
    pairError.classList.remove('hidden');
  }
}

async function loadNeverSaveList() {
  const listEl = document.getElementById('never-save-list');
  const emptyEl = document.getElementById('never-save-empty');

  try {
    const result = await browser.runtime.sendMessage({ type: 'GET_NEVER_SAVE_LIST' });
    const domains = result.domains || [];

    listEl.innerHTML = '';

    if (domains.length === 0) {
      emptyEl.classList.remove('hidden');
      return;
    }

    emptyEl.classList.add('hidden');

    for (const domain of domains) {
      const li = document.createElement('li');

      const span = document.createElement('span');
      span.textContent = domain;

      const btn = document.createElement('button');
      btn.className = 'btn btn-ghost btn-sm';
      btn.textContent = 'Remove';
      btn.addEventListener('click', async () => {
        await browser.runtime.sendMessage({ type: 'REMOVE_NEVER_SAVE', domain });
        loadNeverSaveList();
      });

      li.appendChild(span);
      li.appendChild(btn);
      listEl.appendChild(li);
    }
  } catch {
    emptyEl.textContent = 'Could not load list';
    emptyEl.classList.remove('hidden');
  }
}
