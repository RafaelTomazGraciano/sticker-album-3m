const TOTAL_STICKERS = 28;
const POLL_INTERVAL_MS = 1500;
const STICKER_IMG_BASE_URL = "https://rgcoelho01.github.io/album/docs/images";

const els = {
  nodeIp: document.getElementById("node-ip"),
  nodeId: document.getElementById("node-id"),
  connDot: document.getElementById("conn-dot"),
  figNum: document.getElementById("fig-num"),
  btnVerAlbum: document.getElementById("btn-ver-album"),
  btnVoltar: document.getElementById("btn-voltar"),
  logsList: document.getElementById("logs-list"),
  albumGrid: document.getElementById("album-grid"),
  albumCounter: document.getElementById("album-counter"),
  popupOverlay: document.getElementById("popup-overlay"),
  popupFrom: document.getElementById("popup-from"),
  popupOffers: document.getElementById("popup-offers"),
  popupWants: document.getElementById("popup-wants"),
  btnAceitar: document.getElementById("btn-aceitar"),
  btnRejeitar: document.getElementById("btn-rejeitar"),
};

let knownLogCount = 0;

function figId(numLike) {
  const n = parseInt(numLike, 10);
  if (!n || n < 1 || n > TOTAL_STICKERS) return null;
  return "FIG-" + String(n).padStart(2, "0");
}

function setConnectionStatus(online) {
  els.connDot.classList.toggle("online", online);
  els.connDot.classList.toggle("offline", !online);
}

function addClientLog(text) {
  const line = document.createElement("div");
  line.className = "log-line client";
  line.textContent = text;
  els.logsList.appendChild(line);
  els.logsList.scrollTop = els.logsList.scrollHeight;
}

function renderServerLogs(logs) {
  const wasAtBottom =
    els.logsList.scrollTop + els.logsList.clientHeight >= els.logsList.scrollHeight - 4;

  for (let i = knownLogCount; i < logs.length; i++) {
    const line = document.createElement("div");
    line.className = "log-line";
    line.textContent = logs[i];
    els.logsList.appendChild(line);
  }
  knownLogCount = logs.length;

  if (wasAtBottom) {
    els.logsList.scrollTop = els.logsList.scrollHeight;
  }
}

async function api(path, options) {
  const res = await fetch(path, options);
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `HTTP ${res.status}`);
  }
  return res.json();
}

async function pollStatus() {
  try {
    const data = await api("/api/status");
    els.nodeIp.textContent = data.ip || "--.--.--.--";
    els.nodeId.textContent = data.id || "ALUNO-XX";
    setConnectionStatus(true);
  } catch (e) {
    setConnectionStatus(false);
  }
}

async function pollLogs() {
  try {
    const data = await api("/api/logs");
    renderServerLogs(data.logs || []);
  } catch (e) {
    // servidor indisponível, mantém o que já está na tela
  }
}

async function pollPendingOffer() {
  try {
    const data = await api("/api/pending-offer");
    if (data.pending) {
      els.popupFrom.textContent = data.from || "--";
      els.popupOffers.textContent = data.offers || "--";
      els.popupWants.textContent = data.wants || "--";
      els.popupOverlay.classList.remove("hidden");
    } else {
      els.popupOverlay.classList.add("hidden");
    }
  } catch (e) {
    // sem dados, mantém estado atual do popup
  }
}

function renderAlbum(stickers) {
  const owned = stickers.filter((s) => s.qty > 0).length;
  els.albumCounter.textContent = `${owned}/${TOTAL_STICKERS}`;

  els.albumGrid.innerHTML = "";
  for (const sticker of stickers) {
    const card = document.createElement("div");
    card.className = "sticker-card" + (sticker.qty > 0 ? "" : " missing");

    const imgWrap = document.createElement("div");
    imgWrap.className = "sticker-img-wrap";

    const img = document.createElement("img");
    img.src = `${STICKER_IMG_BASE_URL}/${sticker.id}.png`;
    img.alt = sticker.id;
    imgWrap.appendChild(img);

    if (sticker.qty > 0) {
      const badge = document.createElement("span");
      badge.className = "qty-badge";
      badge.textContent = sticker.qty;
      imgWrap.appendChild(badge);
    }

    const label = document.createElement("span");
    label.className = "sticker-label";
    label.textContent = sticker.id;

    card.appendChild(imgWrap);
    card.appendChild(label);
    els.albumGrid.appendChild(card);
  }
}

async function pollAlbum() {
  try {
    const data = await api("/api/album");
    renderAlbum(data.stickers || []);
  } catch (e) {
    if (!els.albumGrid.dataset.placeholder) {
      const placeholder = [];
      for (let i = 1; i <= TOTAL_STICKERS; i++) {
        placeholder.push({ id: "FIG-" + String(i).padStart(2, "0"), qty: 0 });
      }
      renderAlbum(placeholder);
      els.albumGrid.dataset.placeholder = "1";
    }
  }
}

function showScreen(id) {
  document.querySelectorAll(".screen").forEach((s) => s.classList.remove("active"));
  document.getElementById(id).classList.add("active");
}

els.btnVerAlbum.addEventListener("click", () => {
  showScreen("screen-album");
  pollAlbum();
});

els.btnVoltar.addEventListener("click", () => {
  showScreen("screen-ip");
});

document.querySelectorAll(".cmd-btn").forEach((btn) => {
  btn.addEventListener("click", async () => {
    const cmd = btn.dataset.cmd;

    if (cmd === "search" || cmd === "offer") {
      const sticker = figId(els.figNum.value);
      if (!sticker) {
        addClientLog("Informe um número de figurinha válido (1 a 28) em 'Num FIG'.");
        return;
      }
      try {
        await api(`/api/${cmd}`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ sticker }),
        });
        addClientLog(`Comando enviado: ${cmd} ${sticker}`);
      } catch (e) {
        addClientLog(`Erro ao executar ${cmd}: ${e.message}`);
      }
      return;
    }

    if (cmd === "list") {
      try {
        const data = await api("/api/status");
        const entries = Object.entries(data.inventory || {});
        if (entries.length === 0) {
          addClientLog("Seu inventário está vazio.");
        } else {
          addClientLog("Seu inventário: " + entries.map(([k, v]) => `${k}=${v}`).join(", "));
        }
      } catch (e) {
        addClientLog(`Erro ao listar inventário: ${e.message}`);
      }
      return;
    }

    if (cmd === "peers") {
      try {
        const data = await api("/api/peers");
        const peers = data.peers || [];
        addClientLog(peers.length ? "Seus vizinhos: " + peers.join(", ") : "Nenhum vizinho conectado.");
      } catch (e) {
        addClientLog(`Erro ao listar vizinhos: ${e.message}`);
      }
    }
  });
});

els.btnAceitar.addEventListener("click", () => decideTrade("accept"));
els.btnRejeitar.addEventListener("click", () => decideTrade("reject"));

async function decideTrade(decision) {
  try {
    await api("/api/trade/decision", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ decision }),
    });
  } catch (e) {
    addClientLog(`Erro ao registrar decisão da troca: ${e.message}`);
  } finally {
    els.popupOverlay.classList.add("hidden");
  }
}

function startPolling() {
  pollStatus();
  pollLogs();
  pollPendingOffer();
  pollAlbum();

  setInterval(pollStatus, POLL_INTERVAL_MS);
  setInterval(pollLogs, POLL_INTERVAL_MS);
  setInterval(pollPendingOffer, POLL_INTERVAL_MS);
  setInterval(() => {
    if (document.getElementById("screen-album").classList.contains("active")) {
      pollAlbum();
    }
  }, POLL_INTERVAL_MS);
}

startPolling();
