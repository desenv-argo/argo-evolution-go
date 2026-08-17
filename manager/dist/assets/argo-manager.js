const styles = String.raw`
  :host {
    --argo-bg: hsl(var(--background, 222 47% 4%));
    --argo-panel: rgba(10, 19, 36, .82);
    --argo-panel-strong: rgba(9, 18, 34, .96);
    --argo-border: rgba(148, 163, 184, .14);
    --argo-text: hsl(var(--foreground, 210 40% 98%));
    --argo-muted: #8fa1b8;
    --argo-blue: #1677ff;
    --argo-cyan: #18b8ff;
    --argo-green: #24c78e;
    --argo-red: #ff5d73;
    --argo-amber: #ffb648;
    display: block;
    min-height: 100%;
    color: var(--argo-text);
    font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  }
  * { box-sizing: border-box; }
  button, input, select { font: inherit; }
  button { color: inherit; }
  .page {
    position: relative;
    isolation: isolate;
    min-height: calc(100vh - 64px);
    padding: 28px;
    overflow: hidden;
  }
  .page::before {
    content: "";
    position: fixed;
    z-index: -1;
    right: clamp(-120px, 3vw, 50px);
    bottom: clamp(-100px, 2vw, 30px);
    width: min(47vw, 650px);
    aspect-ratio: 1;
    background: url("/assets/argo-watermark.png") center / contain no-repeat;
    opacity: .045;
    pointer-events: none;
    filter: saturate(1.15);
  }
  .header, .toolbar, .toolbar-group, .metric-head, .conversation-meta,
  .message-meta, .empty, .brand { display: flex; align-items: center; }
  .header { justify-content: space-between; gap: 24px; margin-bottom: 24px; }
  .brand { gap: 13px; }
  .brand img { width: 44px; height: 44px; border-radius: 13px; box-shadow: 0 10px 30px rgba(22,119,255,.22); }
  .eyebrow { margin: 0 0 3px; color: var(--argo-cyan); font-size: 11px; font-weight: 800; letter-spacing: .16em; text-transform: uppercase; }
  h1 { margin: 0; font-size: clamp(25px, 3vw, 34px); line-height: 1.15; letter-spacing: -.035em; }
  .subtitle { margin: 7px 0 0; color: var(--argo-muted); font-size: 14px; }
  .toolbar { flex-wrap: wrap; justify-content: flex-end; gap: 10px; }
  .toolbar-group { gap: 8px; }
  .field { display: grid; gap: 6px; }
  .field label { color: var(--argo-muted); font-size: 11px; font-weight: 700; letter-spacing: .04em; text-transform: uppercase; }
  input, select {
    min-height: 40px;
    border: 1px solid var(--argo-border);
    border-radius: 10px;
    outline: none;
    background: rgba(15, 28, 49, .9);
    color: var(--argo-text);
    padding: 0 12px;
    transition: border-color .18s, box-shadow .18s;
  }
  input:focus, select:focus { border-color: rgba(24,184,255,.65); box-shadow: 0 0 0 3px rgba(24,184,255,.1); }
  input::placeholder { color: #66778c; }
  .button {
    min-height: 40px;
    border: 1px solid var(--argo-border);
    border-radius: 10px;
    background: rgba(15, 28, 49, .9);
    padding: 0 14px;
    cursor: pointer;
  }
  .button:hover { border-color: rgba(24,184,255,.45); background: rgba(20, 39, 67, .95); }
  .button.primary { border-color: transparent; background: linear-gradient(135deg, #176efb, #0aa9ed); font-weight: 750; }
  .button.capture-on { border-color: rgba(36,199,142,.35); background: rgba(36,199,142,.1); color: #68e6ba; font-weight: 750; }
  .metrics { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 14px; margin-bottom: 18px; }
  .card, .panel {
    border: 1px solid var(--argo-border);
    background: linear-gradient(145deg, rgba(15,29,51,.92), rgba(7,16,31,.84));
    box-shadow: 0 18px 50px rgba(0,0,0,.16);
    backdrop-filter: blur(18px);
  }
  .card { min-height: 128px; border-radius: 16px; padding: 18px; }
  .metric-head { justify-content: space-between; gap: 12px; }
  .metric-label { color: var(--argo-muted); font-size: 12px; font-weight: 650; }
  .metric-icon { display: grid; place-items: center; width: 34px; height: 34px; border-radius: 10px; background: rgba(22,119,255,.12); color: var(--argo-cyan); font-size: 17px; }
  .metric-value { margin-top: 16px; font-size: 30px; font-weight: 820; letter-spacing: -.04em; }
  .metric-note { margin-top: 5px; color: var(--argo-muted); font-size: 11px; }
  .good { color: var(--argo-green); } .bad { color: var(--argo-red); } .warn { color: var(--argo-amber); }
  .dashboard-grid { display: grid; grid-template-columns: minmax(0, 1.35fr) minmax(330px, .65fr); gap: 18px; }
  .panel { border-radius: 17px; overflow: hidden; }
  .panel-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; padding: 18px 20px; border-bottom: 1px solid var(--argo-border); }
  .panel-title { margin: 0; font-size: 15px; font-weight: 750; }
  .panel-subtitle { margin: 4px 0 0; color: var(--argo-muted); font-size: 12px; }
  .panel-body { padding: 18px 20px; }
  .chart { display: flex; align-items: flex-end; gap: 10px; min-height: 235px; padding: 20px 4px 2px; }
  .bar-column { display: flex; flex: 1 1 0; min-width: 18px; height: 205px; flex-direction: column; justify-content: flex-end; align-items: stretch; gap: 8px; }
  .bar-stack { display: flex; flex-direction: column-reverse; justify-content: flex-start; overflow: hidden; min-height: 3px; border-radius: 8px 8px 3px 3px; background: rgba(148,163,184,.08); }
  .bar-in { background: linear-gradient(180deg, #20d39c, #119f7a); min-height: 2px; }
  .bar-out { background: linear-gradient(180deg, #22b7ff, #176efb); min-height: 2px; }
  .bar-label { overflow: hidden; color: var(--argo-muted); font-size: 10px; text-align: center; text-overflow: ellipsis; white-space: nowrap; }
  .legend { display: flex; gap: 16px; color: var(--argo-muted); font-size: 11px; }
  .legend span::before { content: ""; display: inline-block; width: 8px; height: 8px; margin-right: 6px; border-radius: 999px; background: var(--legend); }
  .instance-list { display: grid; gap: 9px; }
  .instance-row { display: grid; grid-template-columns: 9px minmax(0,1fr) auto; align-items: center; gap: 11px; padding: 11px 12px; border: 1px solid rgba(148,163,184,.1); border-radius: 12px; background: rgba(9,20,38,.6); }
  .status-dot { width: 8px; height: 8px; border-radius: 999px; background: var(--argo-red); box-shadow: 0 0 12px currentColor; }
  .status-dot.connected { background: var(--argo-green); }
  .instance-name { overflow: hidden; font-size: 13px; font-weight: 700; text-overflow: ellipsis; white-space: nowrap; }
  .instance-owner, .instance-state { color: var(--argo-muted); font-size: 11px; }
  .empty { min-height: 160px; justify-content: center; flex-direction: column; gap: 8px; color: var(--argo-muted); text-align: center; }
  .empty strong { color: var(--argo-text); }
  .skeleton { position: relative; overflow: hidden; background: rgba(148,163,184,.09); color: transparent; border-radius: 8px; }
  .skeleton::after { content: ""; position: absolute; inset: 0; transform: translateX(-100%); background: linear-gradient(90deg, transparent, rgba(255,255,255,.06), transparent); animation: shimmer 1.4s infinite; }
  @keyframes shimmer { to { transform: translateX(100%); } }
  .error { margin-bottom: 16px; border: 1px solid rgba(255,93,115,.3); border-radius: 12px; background: rgba(255,93,115,.08); color: #ffb4bf; padding: 12px 14px; font-size: 13px; }

  .conversation-page { padding-bottom: 18px; }
  .conversation-layout { display: grid; grid-template-columns: minmax(285px, 360px) minmax(0, 1fr); min-height: 650px; max-height: calc(100vh - 230px); }
  .conversation-sidebar { display: flex; min-width: 0; flex-direction: column; border-right: 1px solid var(--argo-border); }
  .conversation-filters { display: grid; grid-template-columns: 1fr 1fr; gap: 9px; padding: 14px; border-bottom: 1px solid var(--argo-border); }
  .conversation-filters .wide { grid-column: 1 / -1; }
  .conversation-list { overflow: auto; padding: 8px; }
  .conversation-item { width: 100%; border: 1px solid transparent; border-radius: 13px; background: transparent; padding: 12px; cursor: pointer; text-align: left; }
  .conversation-item:hover { background: rgba(22,119,255,.07); }
  .conversation-item.selected { border-color: rgba(24,184,255,.28); background: rgba(22,119,255,.12); }
  .conversation-meta { justify-content: space-between; gap: 10px; }
  .conversation-name { overflow: hidden; font-size: 13px; font-weight: 760; text-overflow: ellipsis; white-space: nowrap; }
  .conversation-time { flex: 0 0 auto; color: var(--argo-muted); font-size: 10px; }
  .conversation-preview { display: -webkit-box; overflow: hidden; margin: 7px 0 8px; color: var(--argo-muted); font-size: 12px; line-height: 1.45; -webkit-box-orient: vertical; -webkit-line-clamp: 2; }
  .chips { display: flex; flex-wrap: wrap; gap: 6px; }
  .chip { border: 1px solid rgba(148,163,184,.12); border-radius: 999px; background: rgba(148,163,184,.07); color: var(--argo-muted); padding: 3px 7px; font-size: 9px; font-weight: 700; }
  .chip.alert { border-color: rgba(255,182,72,.24); background: rgba(255,182,72,.1); color: #ffc875; }
  .timeline { display: flex; min-width: 0; flex-direction: column; background: rgba(3,10,22,.32); }
  .timeline-header { min-height: 68px; padding: 13px 18px; border-bottom: 1px solid var(--argo-border); }
  .timeline-name { margin: 0; font-size: 14px; font-weight: 760; }
  .timeline-id { margin: 4px 0 0; color: var(--argo-muted); font-size: 11px; }
  .messages { display: flex; overflow: auto; flex: 1; flex-direction: column; gap: 9px; padding: 20px; }
  .message { align-self: flex-start; width: fit-content; max-width: min(76%, 720px); border: 1px solid var(--argo-border); border-radius: 5px 15px 15px 15px; background: rgba(20,34,55,.92); padding: 10px 12px; box-shadow: 0 8px 24px rgba(0,0,0,.12); }
  .message.outbound { align-self: flex-end; border-color: rgba(22,119,255,.24); border-radius: 15px 5px 15px 15px; background: linear-gradient(145deg, rgba(19,78,151,.86), rgba(15,55,111,.92)); }
  .sender { margin-bottom: 5px; color: #7ed9ff; font-size: 10px; font-weight: 750; }
  .message-text { overflow-wrap: anywhere; color: #f2f7ff; font-size: 13px; line-height: 1.48; white-space: pre-wrap; }
  .message-media { display: inline-flex; margin-top: 8px; color: #95dcff; font-size: 11px; font-weight: 700; text-decoration: none; }
  .message-meta { justify-content: flex-end; gap: 7px; margin-top: 7px; color: rgba(220,231,245,.65); font-size: 9px; }
  .load-more { align-self: center; margin-bottom: 2px; }
  @media (max-width: 1120px) {
    .metrics { grid-template-columns: repeat(2, minmax(0, 1fr)); }
    .dashboard-grid { grid-template-columns: 1fr; }
  }
  @media (max-width: 820px) {
    .page { padding: 18px; }
    .header { align-items: flex-start; flex-direction: column; }
    .toolbar { justify-content: flex-start; }
    .metrics { grid-template-columns: 1fr; }
    .conversation-layout { grid-template-columns: 1fr; max-height: none; }
    .conversation-sidebar { max-height: 480px; border-right: 0; border-bottom: 1px solid var(--argo-border); }
    .timeline { min-height: 600px; }
  }
`;

const numberFormatter = new Intl.NumberFormat("pt-BR");
const percentFormatter = new Intl.NumberFormat("pt-BR", { maximumFractionDigits: 1 });
const shortDateFormatter = new Intl.DateTimeFormat("pt-BR", { day: "2-digit", month: "short" });
const dateTimeFormatter = new Intl.DateTimeFormat("pt-BR", {
  day: "2-digit",
  month: "2-digit",
  year: "2-digit",
  hour: "2-digit",
  minute: "2-digit",
});

function authConfig() {
  const fallback = { apiUrl: window.location.origin, apiKey: "" };
  try {
    const persisted = JSON.parse(localStorage.getItem("evolution-auth") || "{}");
    const state = persisted.state || persisted;
    return {
      apiUrl: String(state.apiUrl || fallback.apiUrl).replace(/\/$/, ""),
      apiKey: state.apiKey || "",
    };
  } catch {
    return fallback;
  }
}

async function request(path, params = {}, signal, options = {}) {
  const auth = authConfig();
  const url = new URL(`${auth.apiUrl}${path}`);
  for (const [key, value] of Object.entries(params)) {
    if (value !== "" && value !== undefined && value !== null) url.searchParams.set(key, value);
  }
  const response = await fetch(url, {
    signal,
    method: options.method || "GET",
    headers: { Accept: "application/json", "Content-Type": "application/json", apikey: auth.apiKey },
    body: options.body ? JSON.stringify(options.body) : undefined,
  });
  const payload = await response.json().catch(() => ({}));
  if (response.status === 401) {
    localStorage.removeItem("evolution-auth");
    window.location.assign("/manager/login");
    throw new Error("Sessao expirada");
  }
  if (!response.ok) throw new Error(payload.error || payload.message || "Falha ao consultar a API");
  return payload.data;
}

function element(tag, className, text) {
  const node = document.createElement(tag);
  if (className) node.className = className;
  if (text !== undefined) node.textContent = text;
  return node;
}

function option(value, text) {
  const node = element("option", "", text);
  node.value = value;
  return node;
}

function periodRange(days) {
  const to = new Date();
  const from = new Date(to);
  from.setDate(from.getDate() - Number(days));
  return { from: from.toISOString(), to: to.toISOString() };
}

function displayName(conversation) {
  return conversation.push_name || conversation.contact || conversation.chat_jid || "Conversa";
}

function messageBody(message) {
  return message.text || message.caption || `[${message.message_type || "mensagem"}]`;
}

function safeMediaURL(value) {
  try {
    const url = new URL(value);
    return url.protocol === "https:" || url.protocol === "http:" ? url.href : "";
  } catch {
    return "";
  }
}

class ArgoBaseElement extends HTMLElement {
  constructor() {
    super();
    this.attachShadow({ mode: "open" });
    this.controller = null;
  }

  disconnectController() {
    this.controller?.abort();
    this.controller = new AbortController();
    return this.controller.signal;
  }

  disconnectedCallback() {
    this.controller?.abort();
  }

  setError(message) {
    const target = this.shadowRoot.querySelector("[data-error]");
    if (!target) return;
    target.hidden = !message;
    target.textContent = message || "";
  }
}

class ArgoDashboard extends ArgoBaseElement {
  connectedCallback() {
    this.instances = [];
    this.summary = null;
    this.render();
    this.load();
  }

  render() {
    this.shadowRoot.innerHTML = `
      <style>${styles}</style>
      <main class="page">
        <header class="header">
          <div class="brand">
            <img src="/assets/argo-brand.png" alt="Argo" />
            <div><p class="eyebrow">Argo Evolution</p><h1>Visão operacional</h1><p class="subtitle">Saúde das instâncias e atividade das conversas em um único lugar.</p></div>
          </div>
          <div class="toolbar">
            <div class="field"><label for="period">Período</label><select id="period"><option value="7">Últimos 7 dias</option><option value="30">Últimos 30 dias</option><option value="90">Últimos 90 dias</option></select></div>
            <div class="field"><label for="instance">Instância</label><select id="instance"><option value="">Todas as instâncias</option></select></div>
            <button class="button" data-capture type="button">Captura</button>
            <button class="button" data-refresh type="button">Atualizar</button>
          </div>
        </header>
        <div class="error" data-error hidden></div>
        <section class="metrics" data-metrics></section>
        <section class="dashboard-grid">
          <article class="panel">
            <div class="panel-header"><div><h2 class="panel-title">Volume de mensagens</h2><p class="panel-subtitle">Recebidas e enviadas por dia</p></div><div class="legend"><span style="--legend:#20d39c">Recebidas</span><span style="--legend:#1677ff">Enviadas</span></div></div>
            <div class="panel-body"><div class="chart" data-chart></div></div>
          </article>
          <article class="panel">
            <div class="panel-header"><div><h2 class="panel-title">Saúde das instâncias</h2><p class="panel-subtitle">Conexões que precisam de atenção</p></div></div>
            <div class="panel-body"><div class="instance-list" data-instances></div></div>
          </article>
        </section>
      </main>`;
    this.shadowRoot.querySelector("[data-refresh]").addEventListener("click", () => this.load());
    this.shadowRoot.querySelector("[data-capture]").addEventListener("click", () => this.toggleCapture());
    this.shadowRoot.querySelector("#period").addEventListener("change", () => this.loadSummary());
    this.shadowRoot.querySelector("#instance").addEventListener("change", () => this.loadSummary());
    this.renderLoading();
  }

  async load() {
    const signal = this.disconnectController();
    this.setError("");
    try {
      const instancesPayload = await request("/instance/all", {}, signal);
      this.instances = Array.isArray(instancesPayload) ? instancesPayload : [];
      this.renderInstanceOptions();
      this.renderInstances();
      await this.loadSummary(signal);
    } catch (error) {
      if (error.name !== "AbortError") this.setError(error.message);
    }
  }

  async loadSummary(existingSignal) {
    const signal = existingSignal || this.disconnectController();
    this.setError("");
    const days = this.shadowRoot.querySelector("#period").value;
    const instanceId = this.shadowRoot.querySelector("#instance").value;
    try {
      this.summary = await request("/analytics/dashboard", { ...periodRange(days), instanceId }, signal);
      this.renderMetrics();
      this.renderChart();
      this.renderCaptureState();
    } catch (error) {
      if (error.name !== "AbortError") this.setError(error.message);
    }
  }

  renderLoading() {
    const target = this.shadowRoot.querySelector("[data-metrics]");
    target.replaceChildren(...Array.from({ length: 8 }, () => element("div", "card skeleton", "Carregando")));
  }

  renderInstanceOptions() {
    const select = this.shadowRoot.querySelector("#instance");
    const current = select.value;
    select.replaceChildren(option("", "Todas as instâncias"));
    for (const instance of this.instances) select.append(option(instance.id, instance.name));
    if ([...select.options].some((item) => item.value === current)) select.value = current;
  }

  renderMetrics() {
    const value = this.summary || {};
    const cards = [
      ["Instâncias", value.instances_total, "Total configurado", "◫", ""],
      ["Conectadas", value.instances_connected, `${percentFormatter.format(value.connection_rate || 0)}% disponíveis`, "●", "good"],
      ["Desconectadas", value.instances_offline, value.instances_offline ? "Requer acompanhamento" : "Tudo conectado", "!", value.instances_offline ? "bad" : "good"],
      ["Mensagens", value.messages_total, `${numberFormatter.format(value.inbound_messages || 0)} recebidas`, "↕", ""],
      ["Conversas ativas", value.active_conversations, "No periodo selecionado", "◎", ""],
      ["Contatos únicos", value.unique_contacts, "Conversas individuais", "♙", ""],
      ["Taxa de entrega", `${percentFormatter.format(value.delivery_rate || 0)}%`, `${numberFormatter.format(value.delivered_messages || 0)} entregues`, "✓", "good"],
      ["Taxa de leitura", `${percentFormatter.format(value.read_rate || 0)}%`, `${numberFormatter.format(value.read_messages || 0)} lidas`, "✓✓", "good"],
    ];
    const target = this.shadowRoot.querySelector("[data-metrics]");
    target.replaceChildren(...cards.map(([label, metric, note, icon, tone]) => {
      const card = element("article", "card");
      const head = element("div", "metric-head");
      head.append(element("span", "metric-label", label), element("span", `metric-icon ${tone}`, icon));
      card.append(head, element("div", `metric-value ${tone}`, typeof metric === "number" ? numberFormatter.format(metric) : metric), element("div", "metric-note", note));
      return card;
    }));
  }

  renderCaptureState() {
    const button = this.shadowRoot.querySelector("[data-capture]");
    const enabled = Boolean(this.summary?.message_capture_enabled);
    button.classList.toggle("capture-on", enabled);
    button.textContent = enabled ? "● Captura ativa" : "○ Ativar captura";
    button.title = enabled ? "Novas mensagens estao sendo armazenadas para consultas e insights" : "Ative sem alterar variaveis no Azure";
  }

  async toggleCapture() {
    const enabled = Boolean(this.summary?.message_capture_enabled);
    if (enabled && !window.confirm("Pausar a captura de novas mensagens? O historico existente sera preservado.")) return;
    const signal = this.disconnectController();
    try {
      await request("/analytics/settings", {}, signal, { method: "PUT", body: { message_capture_enabled: !enabled } });
      await this.loadSummary(signal);
    } catch (error) {
      if (error.name !== "AbortError") this.setError(error.message);
    }
  }

  renderChart() {
    const target = this.shadowRoot.querySelector("[data-chart]");
    const volume = this.summary?.volume || [];
    if (!volume.length) {
      const empty = element("div", "empty");
      empty.append(element("strong", "", "Ainda não há mensagens"), element("span", "", "Os dados passam a aparecer conforme novas conversas forem capturadas."));
      target.replaceChildren(empty);
      return;
    }
    const maximum = Math.max(...volume.map((point) => point.total), 1);
    target.replaceChildren(...volume.map((point) => {
      const column = element("div", "bar-column");
      const stack = element("div", "bar-stack");
      stack.title = `${shortDateFormatter.format(new Date(point.bucket))}: ${point.total} mensagens`;
      stack.style.height = `${Math.max((point.total / maximum) * 100, 3)}%`;
      const inbound = element("div", "bar-in");
      const outbound = element("div", "bar-out");
      inbound.style.height = `${point.total ? (point.inbound / point.total) * 100 : 0}%`;
      outbound.style.height = `${point.total ? (point.outbound / point.total) * 100 : 0}%`;
      stack.append(inbound, outbound);
      column.append(stack, element("span", "bar-label", shortDateFormatter.format(new Date(point.bucket))));
      return column;
    }));
  }

  renderInstances() {
    const target = this.shadowRoot.querySelector("[data-instances]");
    if (!this.instances.length) {
      target.replaceChildren(element("div", "empty", "Nenhuma instância cadastrada."));
      return;
    }
    const sorted = [...this.instances].sort((a, b) => Number(a.connected) - Number(b.connected) || a.name.localeCompare(b.name));
    target.replaceChildren(...sorted.slice(0, 12).map((instance) => {
      const row = element("div", "instance-row");
      row.append(element("span", `status-dot ${instance.connected ? "connected" : ""}`));
      const info = element("div");
      info.append(element("div", "instance-name", instance.name), element("div", "instance-owner", instance.jid ? instance.jid.split("@")[0] : "Sem número vinculado"));
      row.append(info, element("span", `instance-state ${instance.connected ? "good" : "bad"}`, instance.connected ? "Conectada" : "Offline"));
      return row;
    }));
  }
}

class ArgoConversations extends ArgoBaseElement {
  connectedCallback() {
    this.instances = [];
    this.conversations = [];
    this.selected = null;
    this.messages = [];
    this.searchTimer = null;
    this.render();
    this.loadInitial();
  }

  disconnectedCallback() {
    super.disconnectedCallback();
    clearTimeout(this.searchTimer);
  }

  render() {
    this.shadowRoot.innerHTML = `
      <style>${styles}</style>
      <main class="page conversation-page">
        <header class="header">
          <div class="brand"><img src="/assets/argo-brand.png" alt="Argo" /><div><p class="eyebrow">Inteligência conversacional</p><h1>Conversas</h1><p class="subtitle">Consulte o histórico estruturado por canal, contato e período.</p></div></div>
          <div class="toolbar"><button class="button" data-capture type="button">Captura</button><button class="button" data-refresh type="button">Atualizar conversas</button></div>
        </header>
        <div class="error" data-error hidden></div>
        <section class="panel conversation-layout">
          <aside class="conversation-sidebar">
            <div class="conversation-filters">
              <input class="wide" id="search" type="search" placeholder="Buscar nome, numero ou mensagem..." />
              <select id="instance"><option value="">Todas as instâncias</option></select>
              <select id="period"><option value="7">7 dias</option><option value="30" selected>30 dias</option><option value="90">90 dias</option><option value="365">1 ano</option></select>
              <select id="direction"><option value="">Entrada e saida</option><option value="inbound">Recebidas</option><option value="outbound">Enviadas</option></select>
              <select id="messageType"><option value="">Todos os formatos</option><option value="text">Texto</option><option value="image">Imagem</option><option value="audio">Áudio</option><option value="video">Vídeo</option><option value="document">Documento</option><option value="interaction">Interação</option></select>
            </div>
            <div class="conversation-list" data-conversations><div class="empty">Carregando conversas...</div></div>
          </aside>
          <section class="timeline">
            <div class="timeline-header" data-timeline-header><p class="timeline-name">Selecione uma conversa</p><p class="timeline-id">As mensagens serão exibidas aqui.</p></div>
            <div class="messages" data-messages><div class="empty"><strong>Nenhuma conversa selecionada</strong><span>Use os filtros ao lado para encontrar um contato ou canal.</span></div></div>
          </section>
        </section>
      </main>`;
    this.shadowRoot.querySelector("[data-refresh]").addEventListener("click", () => this.loadConversations());
    this.shadowRoot.querySelector("[data-capture]").addEventListener("click", () => this.toggleCapture());
    for (const id of ["instance", "period", "direction", "messageType"]) {
      this.shadowRoot.querySelector(`#${id}`).addEventListener("change", () => this.loadConversations());
    }
    this.shadowRoot.querySelector("#search").addEventListener("input", () => {
      clearTimeout(this.searchTimer);
      this.searchTimer = setTimeout(() => this.loadConversations(), 350);
    });
  }

  filters() {
    const days = this.shadowRoot.querySelector("#period").value;
    return {
      ...periodRange(days),
      instanceId: this.shadowRoot.querySelector("#instance").value,
      direction: this.shadowRoot.querySelector("#direction").value,
      messageType: this.shadowRoot.querySelector("#messageType").value,
      search: this.shadowRoot.querySelector("#search").value.trim(),
      limit: 60,
    };
  }

  async loadInitial() {
    const signal = this.disconnectController();
    this.setError("");
    try {
      const instancesPayload = await request("/instance/all", {}, signal);
      this.instances = Array.isArray(instancesPayload) ? instancesPayload : [];
      const select = this.shadowRoot.querySelector("#instance");
      for (const instance of this.instances) select.append(option(instance.id, instance.name));
      await this.loadCaptureState(signal);
      await this.loadConversations(signal);
    } catch (error) {
      if (error.name !== "AbortError") this.setError(error.message);
    }
  }

  async loadCaptureState(signal) {
    const settings = await request("/analytics/settings", {}, signal);
    this.captureEnabled = Boolean(settings.message_capture_enabled);
    const button = this.shadowRoot.querySelector("[data-capture]");
    button.classList.toggle("capture-on", this.captureEnabled);
    button.textContent = this.captureEnabled ? "● Captura ativa" : "○ Ativar captura";
  }

  async toggleCapture() {
    if (this.captureEnabled && !window.confirm("Pausar a captura de novas mensagens? O historico existente sera preservado.")) return;
    const signal = this.disconnectController();
    try {
      const settings = await request("/analytics/settings", {}, signal, { method: "PUT", body: { message_capture_enabled: !this.captureEnabled } });
      this.captureEnabled = Boolean(settings.message_capture_enabled);
      await this.loadCaptureState(signal);
      if (this.captureEnabled) await this.loadConversations(signal);
    } catch (error) {
      if (error.name !== "AbortError") this.setError(error.message);
    }
  }

  async loadConversations(existingSignal) {
    const signal = existingSignal || this.disconnectController();
    this.setError("");
    const target = this.shadowRoot.querySelector("[data-conversations]");
    target.replaceChildren(element("div", "empty", "Atualizando conversas..."));
    try {
      const page = await request("/analytics/conversations", this.filters(), signal);
      this.conversations = page.items || [];
      this.renderConversations();
      if (this.selected) {
        const stillVisible = this.conversations.find((item) => this.key(item) === this.key(this.selected));
        if (stillVisible) this.selectConversation(stillVisible);
      }
    } catch (error) {
      if (error.name !== "AbortError") this.setError(error.message);
    }
  }

  key(conversation) { return `${conversation.instance_id}:${conversation.chat_jid}`; }

  renderConversations() {
    const target = this.shadowRoot.querySelector("[data-conversations]");
    if (!this.conversations.length) {
      const empty = element("div", "empty");
      empty.append(element("strong", "", "Nenhuma conversa encontrada"), element("span", "", "A captura estruturada comeca a partir desta versao. Ajuste o periodo ou os filtros."));
      target.replaceChildren(empty);
      return;
    }
    target.replaceChildren(...this.conversations.map((conversation) => {
      const button = element("button", `conversation-item ${this.selected && this.key(this.selected) === this.key(conversation) ? "selected" : ""}`);
      button.type = "button";
      const meta = element("div", "conversation-meta");
      meta.append(element("span", "conversation-name", displayName(conversation)), element("time", "conversation-time", dateTimeFormatter.format(new Date(conversation.last_message_at))));
      const chips = element("div", "chips");
      chips.append(element("span", "chip", conversation.instance_name || "Instancia"), element("span", "chip", `${numberFormatter.format(conversation.message_count)} msgs`));
      if (conversation.is_group) chips.append(element("span", "chip", "Grupo"));
      if (conversation.unanswered_inbound) chips.append(element("span", "chip alert", "Sem resposta"));
      button.append(meta, element("p", "conversation-preview", conversation.last_message_text), chips);
      button.addEventListener("click", () => this.selectConversation(conversation));
      return button;
    }));
  }

  async selectConversation(conversation) {
    this.selected = conversation;
    this.messagePage = null;
    this.renderConversations();
    const header = this.shadowRoot.querySelector("[data-timeline-header]");
    header.replaceChildren(element("p", "timeline-name", displayName(conversation)), element("p", "timeline-id", `${conversation.chat_jid} · ${conversation.instance_name}`));
    const target = this.shadowRoot.querySelector("[data-messages]");
    target.replaceChildren(element("div", "empty", "Carregando mensagens..."));
    const signal = this.disconnectController();
    try {
      const page = await request("/analytics/messages", {
        ...this.filters(),
        instanceId: conversation.instance_id,
        chatJid: conversation.chat_jid,
        limit: 100,
      }, signal);
      this.messages = (page.items || []).reverse();
      this.messagePage = page;
      this.renderMessages(page);
    } catch (error) {
      if (error.name !== "AbortError") this.setError(error.message);
    }
  }

  async loadOlderMessages() {
    if (!this.selected || !this.messagePage?.next_cursor) return;
    const target = this.shadowRoot.querySelector("[data-messages]");
    const previousHeight = target.scrollHeight;
    const signal = this.disconnectController();
    try {
      const page = await request("/analytics/messages", {
        ...this.filters(),
        instanceId: this.selected.instance_id,
        chatJid: this.selected.chat_jid,
        before: this.messagePage.next_cursor,
        limit: 100,
      }, signal);
      this.messages = [...(page.items || []).reverse(), ...this.messages];
      this.messagePage = page;
      this.renderMessages(page, false);
      target.scrollTop = target.scrollHeight - previousHeight;
    } catch (error) {
      if (error.name !== "AbortError") this.setError(error.message);
    }
  }

  renderMessages(page, scrollToBottom = true) {
    const target = this.shadowRoot.querySelector("[data-messages]");
    if (!this.messages.length) {
      target.replaceChildren(element("div", "empty", "Nenhuma mensagem encontrada neste periodo."));
      return;
    }
    const nodes = this.messages.map((message) => {
      const bubble = element("article", `message ${message.direction === "outbound" ? "outbound" : ""}`);
      if (message.direction !== "outbound" && (message.push_name || message.participant_jid)) bubble.append(element("div", "sender", message.push_name || message.participant_jid));
      bubble.append(element("div", "message-text", messageBody(message)));
      const mediaURL = safeMediaURL(message.media_url);
      if (mediaURL) {
        const link = element("a", "message-media", `Abrir ${message.message_type || "mídia"} ↗`);
        link.href = mediaURL;
        link.target = "_blank";
        link.rel = "noopener noreferrer";
        bubble.append(link);
      }
      const meta = element("div", "message-meta");
      meta.append(element("span", "", message.message_type || "mensagem"), element("time", "", dateTimeFormatter.format(new Date(message.sent_at))));
      if (message.direction === "outbound") meta.append(element("span", "", message.status || "Sent"));
      bubble.append(meta);
      return bubble;
    });
    if (page.has_more) {
      const loadMore = element("button", "button load-more", "Carregar mensagens anteriores");
      loadMore.type = "button";
      loadMore.addEventListener("click", () => this.loadOlderMessages());
      nodes.unshift(loadMore);
    }
    target.replaceChildren(...nodes);
    if (scrollToBottom) target.scrollTop = target.scrollHeight;
  }
}

if (!customElements.get("argo-dashboard")) customElements.define("argo-dashboard", ArgoDashboard);
if (!customElements.get("argo-conversations")) customElements.define("argo-conversations", ArgoConversations);
