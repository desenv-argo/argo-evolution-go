const styles = String.raw`
  :host {
    --argo-bg: #090b0a;
    --argo-panel: #101311;
    --argo-panel-strong: #0d100e;
    --argo-border: rgba(255, 255, 255, .09);
    --argo-border-strong: rgba(255, 255, 255, .14);
    --argo-text: #f3f6f4;
    --argo-muted: #8c9891;
    --argo-green: #00e59b;
    --argo-green-soft: rgba(0, 229, 155, .1);
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
    max-width: 1760px;
    margin: 0 auto;
    padding: 24px 28px 28px;
    overflow: hidden;
  }
  .page::before {
    content: "";
    position: fixed;
    z-index: -1;
    right: clamp(-140px, 1vw, 20px);
    top: 90px;
    width: min(42vw, 590px);
    aspect-ratio: 1;
    background: url("/assets/argo-watermark.png") center / contain no-repeat;
    opacity: .025;
    pointer-events: none;
    filter: grayscale(.35) saturate(.7);
  }
  .header, .toolbar, .toolbar-group, .metric-head, .conversation-meta,
  .message-meta, .empty, .brand { display: flex; align-items: center; }
  .header { justify-content: space-between; gap: 24px; margin-bottom: 20px; }
  .brand { gap: 11px; }
  .brand img { width: 38px; height: 38px; border-radius: 10px; box-shadow: 0 0 0 1px rgba(0,229,155,.14); }
  .eyebrow { margin: 0 0 2px; color: var(--argo-green); font-size: 9px; font-weight: 800; letter-spacing: .14em; text-transform: uppercase; }
  h1 { margin: 0; font-size: clamp(23px, 2.2vw, 28px); line-height: 1.15; letter-spacing: -.025em; }
  .subtitle { margin: 5px 0 0; color: var(--argo-muted); font-size: 12px; }
  .toolbar { flex-wrap: wrap; justify-content: flex-end; align-items: flex-end; gap: 8px; }
  .toolbar-group { gap: 8px; }
  .field { display: grid; gap: 5px; }
  .field label { color: var(--argo-muted); font-size: 9px; font-weight: 750; letter-spacing: .055em; text-transform: uppercase; }
  input, select {
    min-height: 34px;
    border: 1px solid var(--argo-border);
    border-radius: 7px;
    outline: none;
    background: #121613;
    color: var(--argo-text);
    padding: 0 10px;
    font-size: 12px;
    transition: border-color .18s, box-shadow .18s;
  }
  input:focus, select:focus { border-color: rgba(0,229,155,.6); box-shadow: 0 0 0 3px rgba(0,229,155,.08); }
  input::placeholder { color: #667069; }
  .button {
    min-height: 34px;
    border: 1px solid var(--argo-border);
    border-radius: 7px;
    background: #141815;
    padding: 0 11px;
    font-size: 12px;
    font-weight: 650;
    cursor: pointer;
  }
  .button:hover { border-color: rgba(0,229,155,.38); background: #171d19; }
  .button:focus-visible { outline: 2px solid rgba(0,229,155,.5); outline-offset: 2px; }
  .button:disabled { cursor: wait; opacity: .55; }
  .button.primary { border-color: #00c987; background: #00c987; color: #04100b; font-weight: 750; }
  .button.capture-on { border-color: rgba(0,229,155,.35); background: var(--argo-green-soft); color: #63e9bb; font-weight: 750; }
  .metrics { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 10px; margin-bottom: 14px; }
  .card, .panel {
    border: 1px solid var(--argo-border);
    background: linear-gradient(145deg, #121613, #0e110f);
    box-shadow: 0 10px 28px rgba(0,0,0,.14);
  }
  .card { min-height: 102px; border-radius: 10px; padding: 14px 15px; }
  .metric-head { justify-content: space-between; gap: 12px; }
  .metric-label { color: var(--argo-muted); font-size: 10px; font-weight: 700; letter-spacing: .015em; }
  .metric-icon { width: 7px; height: 7px; overflow: hidden; border-radius: 999px; background: #5e6a63; color: transparent; }
  .metric-icon.good { background: var(--argo-green); } .metric-icon.bad { background: var(--argo-red); }
  .metric-value { margin-top: 11px; font-size: 25px; font-weight: 760; letter-spacing: -.035em; }
  .metric-note { margin-top: 3px; color: var(--argo-muted); font-size: 10px; }
  .good { color: var(--argo-green); } .bad { color: var(--argo-red); } .warn { color: var(--argo-amber); }
  .dashboard-grid { display: grid; grid-template-columns: minmax(0, 1.5fr) minmax(360px, .5fr); gap: 14px; }
  .panel { border-radius: 11px; overflow: hidden; }
  .panel-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; padding: 14px 16px; border-bottom: 1px solid var(--argo-border); }
  .panel-title { margin: 0; font-size: 13px; font-weight: 720; }
  .panel-subtitle { margin: 3px 0 0; color: var(--argo-muted); font-size: 10px; }
  .panel-body { padding: 14px 16px; }
  .chart { display: flex; align-items: flex-end; gap: 8px; min-height: 220px; padding: 16px 4px 2px; }
  .bar-column { display: flex; flex: 1 1 0; min-width: 18px; height: 190px; flex-direction: column; justify-content: flex-end; align-items: stretch; gap: 7px; }
  .bar-stack { display: flex; flex-direction: column-reverse; justify-content: flex-start; overflow: hidden; min-height: 3px; border-radius: 8px 8px 3px 3px; background: rgba(148,163,184,.08); }
  .bar-in { background: #00d993; min-height: 2px; }
  .bar-out { background: #66736b; min-height: 2px; }
  .bar-label { overflow: hidden; color: var(--argo-muted); font-size: 10px; text-align: center; text-overflow: ellipsis; white-space: nowrap; }
  .legend { display: flex; gap: 16px; color: var(--argo-muted); font-size: 11px; }
  .legend span::before { content: ""; display: inline-block; width: 8px; height: 8px; margin-right: 6px; border-radius: 999px; background: var(--legend); }
  .instance-list { display: grid; gap: 9px; }
  .instance-row { display: grid; grid-template-columns: 8px minmax(0,1fr) auto; align-items: center; gap: 10px; padding: 9px 10px; border: 1px solid var(--argo-border); border-radius: 8px; background: #0d110e; }
  .status-dot { width: 8px; height: 8px; border-radius: 999px; background: var(--argo-red); box-shadow: 0 0 12px currentColor; }
  .status-dot.connected { background: var(--argo-green); }
  .instance-name { overflow: hidden; font-size: 13px; font-weight: 700; text-overflow: ellipsis; white-space: nowrap; }
  .instance-owner, .instance-state, .instance-detail { color: var(--argo-muted); font-size: 11px; }
  .instance-detail { margin-top: 3px; max-width: 260px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .instance-actions { display: flex; align-items: center; justify-content: flex-end; gap: 7px; }
  .instance-action { min-height: 27px; padding: 0 8px; border-radius: 6px; font-size: 10px; }
  .empty { min-height: 160px; justify-content: center; flex-direction: column; gap: 8px; color: var(--argo-muted); text-align: center; }
  .empty strong { color: var(--argo-text); }
  .skeleton { position: relative; overflow: hidden; background: rgba(148,163,184,.09); color: transparent; border-radius: 8px; }
  .skeleton::after { content: ""; position: absolute; inset: 0; transform: translateX(-100%); background: linear-gradient(90deg, transparent, rgba(255,255,255,.06), transparent); animation: shimmer 1.4s infinite; }
  @keyframes shimmer { to { transform: translateX(100%); } }
  .error { margin-bottom: 14px; border: 1px solid rgba(255,93,115,.3); border-radius: 8px; background: rgba(255,93,115,.08); color: #ffb4bf; padding: 10px 12px; font-size: 12px; }

  .integration-page { padding-bottom: 20px; }
  .integration-summary { grid-template-columns: repeat(4, minmax(0, 1fr)); }
  .integration-grid { display: grid; grid-template-columns: minmax(330px, .72fr) minmax(0, 1.7fr); gap: 14px; align-items: start; }
  .application-list { display: grid; gap: 8px; }
  .application-row { display: grid; gap: 9px; border: 1px solid var(--argo-border); border-radius: 9px; background: #0d110e; padding: 11px 12px; }
  .application-row:hover { border-color: rgba(0,229,155,.2); }
  .application-top, .application-actions, .status-line, .operations-toolbar, .pagination, .credential-actions { display: flex; align-items: center; }
  .application-top { justify-content: space-between; gap: 12px; }
  .application-title { min-width: 0; }
  .application-name { overflow: hidden; font-size: 13px; font-weight: 760; text-overflow: ellipsis; white-space: nowrap; }
  .application-slug { margin-top: 2px; color: var(--argo-muted); font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 10px; }
  .application-meta { display: grid; grid-template-columns: 1fr 1fr; gap: 7px 12px; }
  .meta-label { color: #69746d; font-size: 8px; font-weight: 750; letter-spacing: .06em; text-transform: uppercase; }
  .meta-value { display: block; overflow: hidden; margin-top: 2px; color: #cbd4ce; font-size: 10px; text-overflow: ellipsis; white-space: nowrap; }
  .application-actions { justify-content: flex-end; gap: 6px; border-top: 1px solid rgba(255,255,255,.055); padding-top: 8px; }
  .status-line { gap: 6px; color: var(--argo-muted); font-size: 10px; }
  .status-pill { display: inline-flex; align-items: center; gap: 5px; border: 1px solid var(--argo-border); border-radius: 999px; background: rgba(148,163,184,.055); padding: 3px 7px; color: var(--argo-muted); font-size: 9px; font-weight: 740; }
  .status-pill::before { content: ""; width: 5px; height: 5px; border-radius: 999px; background: currentColor; }
  .status-pill.active { border-color: rgba(0,229,155,.2); background: rgba(0,229,155,.07); color: #57ddb0; }
  .status-pill.inactive { border-color: rgba(255,93,115,.2); background: rgba(255,93,115,.07); color: #ff8293; }
  .status-pill.healthy { border-color: rgba(0,229,155,.2); background: rgba(0,229,155,.07); color: #57ddb0; }
  .status-pill.degraded { border-color: rgba(245,181,74,.25); background: rgba(245,181,74,.07); color: #e6bd72; }
  .status-pill.offline { border-color: rgba(255,93,115,.2); background: rgba(255,93,115,.07); color: #ff8293; }
  .status-pill.unknown, .status-pill.disabled { color: #8d9891; }
  .operations-toolbar { flex-wrap: wrap; justify-content: space-between; gap: 8px; padding: 11px 14px; border-bottom: 1px solid var(--argo-border); }
  .operation-filters { display: flex; flex-wrap: wrap; gap: 7px; }
  .operation-filters select { min-width: 120px; }
  .table-wrap { overflow: auto; }
  .operations-table { width: 100%; border-collapse: collapse; font-size: 10px; }
  .operations-table th { position: sticky; z-index: 1; top: 0; background: #101411; color: #77827b; padding: 8px 10px; font-size: 8px; font-weight: 800; letter-spacing: .055em; text-align: left; text-transform: uppercase; white-space: nowrap; }
  .operations-table td { max-width: 230px; border-top: 1px solid rgba(255,255,255,.055); padding: 9px 10px; color: #c7d0ca; vertical-align: middle; }
  .operations-table tbody tr:hover { background: rgba(0,229,155,.025); }
  .mono { overflow: hidden; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; text-overflow: ellipsis; white-space: nowrap; }
  .operation-error { color: #ff8293; font-weight: 700; }
  .operation-ok { color: #57ddb0; font-weight: 700; }
  .operation-muted { color: var(--argo-muted); }
  .error-breakdown { display: flex; flex-wrap: wrap; gap: 6px; padding: 0 14px 12px; }
  .error-token { border: 1px solid rgba(255,93,115,.15); border-radius: 6px; background: rgba(255,93,115,.055); color: #dca6ae; padding: 5px 7px; font-size: 9px; }
  .pagination { justify-content: space-between; gap: 12px; border-top: 1px solid var(--argo-border); padding: 9px 14px; color: var(--argo-muted); font-size: 9px; }
  .gateway-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 14px; margin-bottom: 14px; }
  .usage-list { display: grid; gap: 7px; }
  .usage-row { display: grid; grid-template-columns: minmax(0, 1fr) auto; gap: 8px 12px; border-bottom: 1px solid rgba(255,255,255,.055); padding: 7px 0; }
  .usage-row:last-child { border-bottom: 0; }
  .usage-key { overflow: hidden; font-size: 11px; font-weight: 720; text-overflow: ellipsis; white-space: nowrap; }
  .usage-meta { color: var(--argo-muted); font-size: 9px; }
  .usage-total { color: #dce5df; font-size: 12px; font-weight: 760; text-align: right; }
  .upstream-change { display: block; overflow: hidden; color: #b9c5be; font-size: 9px; text-overflow: ellipsis; white-space: nowrap; text-decoration: none; }
  .upstream-change:hover { color: var(--argo-green); }
  .dialog-backdrop { position: fixed; z-index: 20; inset: 0; display: grid; place-items: center; background: rgba(0,0,0,.68); padding: 20px; backdrop-filter: blur(3px); }
  .dialog-backdrop[hidden] { display: none; }
  .dialog { width: min(560px, 100%); max-height: min(760px, calc(100vh - 40px)); overflow: auto; border: 1px solid var(--argo-border-strong); border-radius: 11px; background: #101411; box-shadow: 0 24px 80px rgba(0,0,0,.48); }
  .dialog-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; border-bottom: 1px solid var(--argo-border); padding: 14px 16px; }
  .dialog-body { padding: 16px; }
  .dialog-form { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
  .dialog-form .wide { grid-column: 1 / -1; }
  .dialog-form input, .dialog-form select { width: 100%; }
  .dialog-footer { display: flex; justify-content: flex-end; gap: 7px; border-top: 1px solid var(--argo-border); padding: 11px 16px; }
  .checkbox-field { display: flex; align-items: center; gap: 8px; color: #cbd4ce; font-size: 11px; }
  .checkbox-field input { width: 14px; min-height: auto; accent-color: var(--argo-green); }
  .credential-box { border: 1px solid rgba(0,229,155,.2); border-radius: 8px; background: rgba(0,229,155,.06); padding: 12px; }
  .credential-value { display: block; overflow-wrap: anywhere; color: #aaf3d8; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 11px; line-height: 1.5; }
  .credential-warning { margin: 0 0 12px; color: #d4c098; font-size: 11px; line-height: 1.5; }
  .credential-actions { justify-content: flex-end; gap: 7px; margin-top: 12px; }

  .lifecycle-page { padding-bottom: 22px; }
  .lifecycle-funnel { display: grid; grid-template-columns: repeat(5, minmax(120px, 1fr)); gap: 8px; margin-bottom: 14px; }
  .funnel-stage { position: relative; min-height: 90px; border: 1px solid var(--argo-border); border-radius: 9px; background: linear-gradient(145deg, #121713, #0c100d); padding: 12px; }
  .funnel-stage:not(:last-child)::after { content: "›"; position: absolute; z-index: 2; right: -8px; top: 31px; color: #526058; font-size: 20px; }
  .funnel-stage strong { display: block; margin-top: 8px; font-size: 23px; letter-spacing: -.03em; }
  .funnel-stage small { color: var(--argo-muted); font-size: 9px; }
  .lifecycle-alerts { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; margin-bottom: 14px; }
  .alert-card { display: flex; align-items: center; justify-content: space-between; gap: 12px; border: 1px solid var(--argo-border); border-radius: 9px; background: #101411; padding: 10px 13px; }
  .alert-card strong { font-size: 18px; }
  .lifecycle-layout { display: grid; grid-template-columns: minmax(0, 1.6fr) minmax(310px, .55fr); gap: 14px; align-items: start; }
  .latency-list, .failure-list, .lifecycle-timeline { display: grid; gap: 8px; }
  .latency-row { display: grid; grid-template-columns: minmax(80px, 1fr) repeat(4, minmax(45px, auto)); gap: 8px; align-items: center; border-bottom: 1px solid rgba(255,255,255,.055); padding: 8px 0; font-size: 10px; }
  .latency-row:last-child { border-bottom: 0; }
  .latency-row span:not(:first-child) { color: #cbd4ce; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; text-align: right; }
  .timeline-event { display: grid; grid-template-columns: 12px minmax(0, 1fr) auto; gap: 9px; align-items: start; }
  .timeline-event::before { content: ""; width: 8px; height: 8px; margin-top: 4px; border-radius: 999px; background: var(--argo-green); box-shadow: 0 0 0 3px rgba(0,229,155,.08); }
  .timeline-event.failed::before, .timeline-event.pending_aged::before { background: var(--argo-red); box-shadow: 0 0 0 3px rgba(255,93,115,.08); }
  .timeline-event time { color: var(--argo-muted); font-size: 9px; white-space: nowrap; }

  .conversation-page { padding-bottom: 18px; }
  .conversation-layout { display: grid; grid-template-columns: minmax(285px, 360px) minmax(0, 1fr); min-height: 650px; max-height: calc(100vh - 230px); }
  .conversation-sidebar { display: flex; min-width: 0; flex-direction: column; border-right: 1px solid var(--argo-border); }
  .conversation-filters { display: grid; grid-template-columns: 1fr 1fr; gap: 9px; padding: 14px; border-bottom: 1px solid var(--argo-border); }
  .conversation-filters .wide { grid-column: 1 / -1; }
  .conversation-list { overflow: auto; padding: 8px; }
  .conversation-item { width: 100%; border: 1px solid transparent; border-radius: 8px; background: transparent; padding: 10px; cursor: pointer; text-align: left; }
  .conversation-item:hover { background: rgba(0,229,155,.05); }
  .conversation-item.selected { border-color: rgba(0,229,155,.24); background: rgba(0,229,155,.08); }
  .conversation-meta { justify-content: space-between; gap: 10px; }
  .conversation-name { overflow: hidden; font-size: 13px; font-weight: 760; text-overflow: ellipsis; white-space: nowrap; }
  .conversation-time { flex: 0 0 auto; color: var(--argo-muted); font-size: 10px; }
  .conversation-preview { display: -webkit-box; overflow: hidden; margin: 7px 0 8px; color: var(--argo-muted); font-size: 12px; line-height: 1.45; -webkit-box-orient: vertical; -webkit-line-clamp: 2; }
  .chips { display: flex; flex-wrap: wrap; gap: 6px; }
  .chip { border: 1px solid rgba(148,163,184,.12); border-radius: 999px; background: rgba(148,163,184,.07); color: var(--argo-muted); padding: 3px 7px; font-size: 9px; font-weight: 700; }
  .chip.alert { border-color: rgba(255,182,72,.24); background: rgba(255,182,72,.1); color: #ffc875; }
  .timeline { display: flex; min-width: 0; flex-direction: column; background: #0b0e0c; }
  .timeline-header { min-height: 68px; padding: 13px 18px; border-bottom: 1px solid var(--argo-border); }
  .timeline-name { margin: 0; font-size: 14px; font-weight: 760; }
  .timeline-id { margin: 4px 0 0; color: var(--argo-muted); font-size: 11px; }
  .messages { display: flex; overflow: auto; flex: 1; flex-direction: column; gap: 9px; padding: 20px; }
  .message { align-self: flex-start; width: fit-content; max-width: min(76%, 720px); border: 1px solid var(--argo-border); border-radius: 5px 12px 12px 12px; background: #171b18; padding: 9px 11px; box-shadow: 0 7px 20px rgba(0,0,0,.12); }
  .message.outbound { align-self: flex-end; border-color: rgba(0,229,155,.2); border-radius: 12px 5px 12px 12px; background: #103326; }
  .sender { margin-bottom: 5px; color: #62dfb4; font-size: 10px; font-weight: 750; }
  .message-text { overflow-wrap: anywhere; color: #f2f7ff; font-size: 13px; line-height: 1.48; white-space: pre-wrap; }
  .message-media { display: inline-flex; margin-top: 8px; color: #62dfb4; font-size: 11px; font-weight: 700; text-decoration: none; }
  .document-card { display: grid; grid-template-columns: auto minmax(0, 1fr); gap: 10px; margin-top: 9px; border: 1px solid rgba(98,223,180,.18); border-radius: 8px; background: rgba(0,0,0,.16); padding: 10px; }
  .document-icon { display: grid; width: 34px; height: 40px; place-items: center; border-radius: 6px; background: rgba(0,229,155,.1); color: #62dfb4; font-size: 11px; font-weight: 800; }
  .document-info { min-width: 0; }
  .document-name { overflow: hidden; color: #f2f7ff; font-size: 11px; font-weight: 750; text-overflow: ellipsis; white-space: nowrap; }
  .document-size { margin-top: 2px; color: var(--argo-muted); font-size: 9px; }
  .document-actions { display: flex; flex-wrap: wrap; gap: 7px; margin-top: 8px; }
  .document-action { border: 0; background: transparent; color: #62dfb4; padding: 0; cursor: pointer; font: inherit; font-size: 10px; font-weight: 750; }
  .document-action:hover { text-decoration: underline; }
  .media-preview { grid-column: 1 / -1; overflow: hidden; max-width: 520px; border-radius: 7px; background: #080b09; }
  .media-preview img, .media-preview video { display: block; width: 100%; max-height: 360px; object-fit: contain; }
  .media-preview audio { display: block; width: min(420px, 100%); height: 36px; }
  .media-status { grid-column: 1 / -1; color: var(--argo-muted); font-size: 10px; }
  .media-status.unavailable { color: #ffc875; }
  .message-meta { justify-content: flex-end; gap: 7px; margin-top: 7px; color: rgba(220,231,245,.65); font-size: 9px; }
  .load-more { align-self: center; margin-bottom: 2px; }
  @media (max-width: 1120px) {
    .metrics { grid-template-columns: repeat(2, minmax(0, 1fr)); }
    .dashboard-grid { grid-template-columns: 1fr; }
    .integration-grid { grid-template-columns: 1fr; }
	.gateway-grid { grid-template-columns: 1fr; }
	.lifecycle-layout { grid-template-columns: 1fr; }
  }
  @media (max-width: 820px) {
    .page { padding: 18px; }
    .header { align-items: flex-start; flex-direction: column; }
    .toolbar { justify-content: flex-start; }
    .metrics { grid-template-columns: 1fr; }
    .integration-summary { grid-template-columns: repeat(2, minmax(0, 1fr)); }
	.lifecycle-funnel { grid-template-columns: 1fr; }
	.funnel-stage:not(:last-child)::after { display: none; }
    .dialog-form { grid-template-columns: 1fr; }
    .dialog-form .wide { grid-column: auto; }
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

const relativeFormatter = new Intl.RelativeTimeFormat("pt-BR", { numeric: "auto" });

function relativeTime(value) {
  if (!value) return "Nunca";
  const timestamp = new Date(value).getTime();
  if (!Number.isFinite(timestamp)) return "Nunca";
  const seconds = Math.round((timestamp - Date.now()) / 1000);
  const ranges = [[86400, "day"], [3600, "hour"], [60, "minute"]];
  for (const [size, unit] of ranges) {
    if (Math.abs(seconds) >= size) return relativeFormatter.format(Math.round(seconds / size), unit);
  }
  return relativeFormatter.format(seconds, "second");
}

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

function conversationChatJID(conversation) {
  return conversation.chat_jid || conversation.chat_j_id || conversation.chatJid || "";
}

function displayName(conversation) {
  return conversation.push_name || conversation.contact || conversationChatJID(conversation) || "Conversa";
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

function formatFileSize(value) {
  const bytes = Number(value || 0);
  if (!bytes) return "Tamanho não informado";
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

async function fetchCapturedMedia(message) {
  const auth = authConfig();
  const url = new URL(`${auth.apiUrl}/argo/v1/messages/${encodeURIComponent(message.message_id)}/media`);
  url.searchParams.set("instanceId", message.instance_id);
  const response = await fetch(url, { headers: { apikey: auth.apiKey } });
  if (!response.ok) {
    const fallbackURL = safeMediaURL(message.media_url);
    if (fallbackURL) return { fallbackURL };
    const payload = await response.json().catch(() => ({}));
    throw new Error(payload.error || "O arquivo ainda não está disponível para esta mensagem");
  }
  const blob = await response.blob();
  return { blobURL: URL.createObjectURL(blob), contentType: blob.type || message.mime_type || "application/octet-stream" };
}

async function openCapturedMedia(message, download = false) {
  const media = await fetchCapturedMedia(message);
  const targetURL = media.blobURL || media.fallbackURL;
  if (download) {
    const link = document.createElement("a");
    link.href = targetURL;
    link.download = message.file_name || `midia-${message.message_id}`;
    link.click();
  } else {
    window.open(targetURL, "_blank", "noopener,noreferrer");
  }
  if (media.blobURL) setTimeout(() => URL.revokeObjectURL(media.blobURL), 60000);
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
    this.health = new Map();
    this.summary = null;
    this.healthSummary = null;
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
            <div class="panel-header"><div><h2 class="panel-title">Volume de mensagens</h2><p class="panel-subtitle">Recebidas e enviadas por dia</p></div><div class="legend"><span style="--legend:#00d993">Recebidas</span><span style="--legend:#66736b">Enviadas</span></div></div>
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
      await this.loadHealth(signal);
      this.renderInstanceOptions();
      this.renderInstances();
      await this.loadSummary(signal);
    } catch (error) {
      if (error.name !== "AbortError") this.setError(error.message);
    }
  }

  async loadHealth(signal) {
    const results = await Promise.allSettled(this.instances.map((instance) =>
      request(`/instance/health/${encodeURIComponent(instance.id)}`, {}, signal)
    ));
    results.forEach((result, index) => {
      if (result.status === "fulfilled") this.health.set(this.instances[index].id, result.value);
    });
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
    const liveConnected = this.instances.filter((instance) => {
      const health = this.health.get(instance.id);
      return health ? health.transport_connected && health.logged_in : instance.connected;
    }).length;
    const instanceTotal = this.instances.length || value.instances_total || 0;
    const liveOffline = Math.max(instanceTotal - liveConnected, 0);
    const liveRate = instanceTotal ? (liveConnected / instanceTotal) * 100 : 0;
    const cards = [
      ["Instâncias", instanceTotal, "Total configurado", "◫", ""],
      ["Conectadas", liveConnected, `${percentFormatter.format(liveRate)}% disponíveis`, "●", "good"],
      ["Desconectadas", liveOffline, liveOffline ? "Requer acompanhamento" : "Tudo conectado", "!", liveOffline ? "bad" : "good"],
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
    const isLive = (instance) => {
      const health = this.health.get(instance.id);
      return health ? health.transport_connected && health.logged_in : instance.connected;
    };
    const labels = { operational: "Operacional", connected: "Conectada", degraded: "Degradada", reconnecting: "Reconectando", linked_offline: "Sessão offline", needs_pairing: "Requer QR" };
    const sorted = [...this.instances].sort((a, b) => Number(isLive(a)) - Number(isLive(b)) || a.name.localeCompare(b.name));
    target.replaceChildren(...sorted.slice(0, 12).map((instance) => {
      const health = this.health.get(instance.id);
      const live = isLive(instance);
      const state = health?.state || (live ? "connected" : "linked_offline");
      const tone = state === "degraded" || state === "reconnecting" ? "warn" : live ? "good" : "bad";
      const row = element("div", "instance-row");
      const dot = element("span", `status-dot ${live ? "connected" : ""}`);
      if (tone === "warn") dot.style.background = "var(--argo-amber)";
      row.append(dot);
      const info = element("div");
      info.append(element("div", "instance-name", instance.name), element("div", "instance-owner", instance.jid ? instance.jid.split("@")[0] : "Sem número vinculado"));
      if (health?.detail) info.append(element("div", "instance-detail", health.probe_performed && health.probe_ok ? `${health.detail} (${health.latency_ms} ms)` : health.detail));
      const actions = element("div", "instance-actions");
      actions.append(element("span", `instance-state ${tone}`, labels[state] || "Offline"));
      if (instance.jid) {
        const test = element("button", "button instance-action", "Testar");
        test.type = "button";
        test.addEventListener("click", () => this.testInstance(instance, test));
        actions.append(test);
      }
      if (!live && instance.jid) {
        const resume = element("button", "button instance-action", "Reconectar");
        resume.type = "button";
        resume.addEventListener("click", () => this.resumeInstance(instance, resume));
        actions.append(resume);
      }
      row.append(info, actions);
      return row;
    }));
  }

  async testInstance(instance, button) {
    button.disabled = true;
    button.textContent = "Testando...";
    this.setError("");
    try {
      const health = await request(`/instance/health/${encodeURIComponent(instance.id)}`, { probe: true });
      this.health.set(instance.id, health);
      this.renderInstances();
      this.renderMetrics();
    } catch (error) {
      this.setError(error.message);
      button.disabled = false;
      button.textContent = "Testar";
    }
  }

  async resumeInstance(instance, button) {
    button.disabled = true;
    button.textContent = "Reconectando...";
    this.setError("");
    try {
      const health = await request(`/instance/resume/${encodeURIComponent(instance.id)}`, {}, undefined, { method: "POST" });
      this.health.set(instance.id, health);
      this.renderInstances();
      this.renderMetrics();
    } catch (error) {
      this.setError(error.message);
      button.disabled = false;
      button.textContent = "Reconectar";
    }
  }
}

class ArgoIntegrations extends ArgoBaseElement {
  connectedCallback() {
    this.applications = [];
    this.attempts = [];
    this.summary = null;
    this.editing = null;
    this.render();
    this.load();
  }

  render() {
    this.shadowRoot.innerHTML = `
      <style>${styles}</style>
      <main class="page integration-page">
        <header class="header">
          <div class="brand"><img src="/assets/argo-brand.png" alt="Argo" /><div><p class="eyebrow">Ecossistema Argo</p><h1>Integrações</h1><p class="subtitle">Identidade, atividade e rastreabilidade das aplicações que consomem o canal.</p></div></div>
          <div class="toolbar"><button class="button" data-refresh type="button">Atualizar</button><button class="button primary" data-create type="button">Nova integração</button></div>
        </header>
        <div class="error" data-error hidden></div>
        <section class="metrics integration-summary" data-summary></section>
        <section class="gateway-grid">
          <article class="panel"><div class="panel-header"><div><h2 class="panel-title">Uso por aplicação</h2><p class="panel-subtitle">Quem está consumindo o gateway no período</p></div></div><div class="panel-body usage-list" data-application-usage></div></article>
          <article class="panel"><div class="panel-header"><div><h2 class="panel-title">Uso por instância</h2><p class="panel-subtitle">Distribuição e falhas por canal WhatsApp</p></div></div><div class="panel-body usage-list" data-instance-usage></div></article>
          <article class="panel"><div class="panel-header"><div><h2 class="panel-title">Sinais operacionais</h2><p class="panel-subtitle">Entrega, leitura e categorias de falha</p></div></div><div class="panel-body usage-list" data-gateway-signals></div></article>
          <article class="panel"><div class="panel-header"><div><h2 class="panel-title">Radar do upstream</h2><p class="panel-subtitle">Evolution GO original e defasagem do fork</p></div><span class="status-pill" data-upstream-status>Verificando</span></div><div class="panel-body usage-list" data-upstream></div></article>
        </section>
        <section class="integration-grid">
          <article class="panel">
            <div class="panel-header"><div><h2 class="panel-title">Aplicações</h2><p class="panel-subtitle">Sistemas autorizados a identificar suas operações</p></div><span class="status-pill" data-app-count>0</span></div>
            <div class="panel-body"><div class="application-list" data-applications></div></div>
          </article>
          <article class="panel">
            <div class="panel-header"><div><h2 class="panel-title">Tentativas de envio</h2><p class="panel-subtitle">Aceitação da API, latência e falhas antes da confirmação de entrega</p></div></div>
            <div class="operations-toolbar">
              <div class="operation-filters">
                <select data-period aria-label="Período"><option value="1">24 horas</option><option value="7" selected>7 dias</option><option value="30">30 dias</option><option value="90">90 dias</option></select>
                <select data-application aria-label="Aplicação"><option value="">Todas as aplicações</option></select>
                <select data-outcome aria-label="Resultado"><option value="">Todos os resultados</option><option value="succeeded">Aceitas</option><option value="failed">Falhas</option></select>
                <select data-error-code aria-label="Erro"><option value="">Todos os erros</option></select>
              </div>
              <button class="button" data-operations-refresh type="button">Recarregar</button>
            </div>
            <div class="error-breakdown" data-errors></div>
            <div class="table-wrap">
              <table class="operations-table">
                <thead><tr><th>Horário</th><th>Aplicação</th><th>Endpoint</th><th>Resultado</th><th>Latência</th><th>Erro</th><th>Correlação</th></tr></thead>
                <tbody data-attempts></tbody>
              </table>
            </div>
            <div class="pagination"><span data-attempt-count>0 operações</span><span>Últimas 100 no período</span></div>
          </article>
        </section>
      </main>
      <div class="dialog-backdrop" data-form-dialog hidden>
        <section class="dialog" role="dialog" aria-modal="true" aria-labelledby="integration-dialog-title">
          <div class="dialog-header"><div><h2 class="panel-title" id="integration-dialog-title" data-form-title>Nova integração</h2><p class="panel-subtitle">A credencial será exibida somente uma vez.</p></div><button class="button instance-action" data-close-form type="button">Fechar</button></div>
          <form data-form>
            <div class="dialog-body dialog-form">
              <div class="field"><label for="app-name">Nome</label><input id="app-name" name="name" maxlength="160" required /></div>
              <div class="field"><label for="app-slug">Identificador</label><input id="app-slug" name="slug" maxlength="100" pattern="[a-z0-9][a-z0-9-]{1,98}[a-z0-9]" placeholder="argo-erp" required /></div>
              <div class="field"><label for="app-environment">Ambiente</label><select id="app-environment" name="environment"><option value="production">Produção</option><option value="staging">Homologação</option><option value="development">Desenvolvimento</option></select></div>
              <div class="field"><label for="app-owner">Responsável</label><input id="app-owner" name="owner" maxlength="160" placeholder="Time ou pessoa" /></div>
              <div class="field wide"><label for="app-base-url">URL da aplicação</label><input id="app-base-url" name="base_url" type="url" placeholder="https://erp.argo.app.br" /></div>
              <div class="field wide"><label for="app-health-url">Endpoint de saúde</label><input id="app-health-url" name="health_url" type="url" placeholder="https://erp.argo.app.br/health" /></div>
              <div class="field"><label for="app-heartbeat">Janela esperada (segundos)</label><input id="app-heartbeat" name="expected_heartbeat_seconds" type="number" min="30" max="86400" value="300" /></div>
              <label class="checkbox-field"><input name="active" type="checkbox" checked /> Integração habilitada</label>
            </div>
            <div class="dialog-footer"><button class="button" data-cancel-form type="button">Cancelar</button><button class="button primary" data-save type="submit">Salvar integração</button></div>
          </form>
        </section>
      </div>
      <div class="dialog-backdrop" data-credential-dialog hidden>
        <section class="dialog" role="dialog" aria-modal="true" aria-labelledby="credential-title">
          <div class="dialog-header"><div><h2 class="panel-title" id="credential-title">Credencial da integração</h2><p class="panel-subtitle" data-credential-app></p></div><button class="button instance-action" data-close-credential type="button">Fechar</button></div>
          <div class="dialog-body"><p class="credential-warning">Copie e armazene esta credencial agora. Por segurança, ela não poderá ser consultada novamente.</p><div class="credential-box"><code class="credential-value" data-credential></code></div><div class="credential-actions"><button class="button primary" data-copy-credential type="button">Copiar credencial</button></div></div>
        </section>
      </div>`;

    this.shadowRoot.querySelector("[data-refresh]").addEventListener("click", () => this.load());
    this.shadowRoot.querySelector("[data-create]").addEventListener("click", () => this.openForm());
    this.shadowRoot.querySelector("[data-operations-refresh]").addEventListener("click", () => this.loadOperations());
    for (const selector of ["[data-period]", "[data-application]", "[data-outcome]", "[data-error-code]"]) {
      this.shadowRoot.querySelector(selector).addEventListener("change", () => this.loadOperations());
    }
    this.shadowRoot.querySelector("[data-form]").addEventListener("submit", (event) => this.saveApplication(event));
    for (const selector of ["[data-close-form]", "[data-cancel-form]"]) this.shadowRoot.querySelector(selector).addEventListener("click", () => this.closeForm());
    this.shadowRoot.querySelector("[data-close-credential]").addEventListener("click", () => this.closeCredential());
    this.shadowRoot.querySelector("[data-copy-credential]").addEventListener("click", () => this.copyCredential());
    this.renderLoading();
  }

  renderLoading() {
    this.shadowRoot.querySelector("[data-summary]").replaceChildren(...Array.from({ length: 4 }, () => element("div", "card skeleton", "Carregando")));
    this.shadowRoot.querySelector("[data-applications]").replaceChildren(element("div", "empty", "Carregando aplicações..."));
  }

  operationFilters() {
    const period = this.shadowRoot.querySelector("[data-period]").value;
    return {
      ...periodRange(period),
      application: this.shadowRoot.querySelector("[data-application]").value,
      outcome: this.shadowRoot.querySelector("[data-outcome]").value,
      errorCode: this.shadowRoot.querySelector("[data-error-code]").value,
      limit: 100,
    };
  }

  async load() {
    const signal = this.disconnectController();
    this.setError("");
    try {
      const applications = await request("/argo/v1/applications", {}, signal);
      this.applications = Array.isArray(applications) ? applications : [];
      this.renderApplicationFilter();
      this.renderApplications();
      await this.loadOperations(signal);
    } catch (error) {
      if (error.name !== "AbortError") this.setError(error.message);
    }
  }

  async loadOperations(existingSignal) {
    const signal = existingSignal || this.disconnectController();
    this.setError("");
    try {
      const filters = this.operationFilters();
      const [overview, attempts, upstream] = await Promise.all([
        request("/argo/v1/operations/overview", filters, signal),
        request("/argo/v1/operations/attempts", filters, signal),
        request("/argo/v1/upstream/status", {}, signal),
      ]);
      this.overview = overview || {};
      this.summary = this.overview.attempts || {};
      this.attempts = Array.isArray(attempts) ? attempts : [];
      this.upstream = upstream || {};
      this.renderSummary();
      this.renderGatewayUsage();
      this.renderUpstream();
      this.renderErrors();
      this.renderAttempts();
    } catch (error) {
      if (error.name !== "AbortError") this.setError(error.message);
    }
  }

  hasRecentActivity(application) {
    if (!application.active || !application.last_seen_at) return false;
    const expected = Math.max(Number(application.expected_heartbeat_seconds) || 300, 30) * 2000;
    return Date.now() - new Date(application.last_seen_at).getTime() <= expected;
  }

  renderSummary() {
    const total = Number(this.summary?.total || 0);
    const failed = Number(this.summary?.failed || 0);
    const failureRate = total ? (failed / total) * 100 : 0;
    const gatewayState = this.overview?.state || "unknown";
    const stateLabels = { healthy: "Saudável", degraded: "Degradado", unhealthy: "Crítico", unknown: "Desconhecido" };
    const uptimeHours = Math.floor(Number(this.overview?.runtime?.uptime_seconds || 0) / 3600);
    const cards = [
      ["Gateway", stateLabels[gatewayState], `v${this.overview?.runtime?.version || "0.0.0"} · uptime ${uptimeHours}h`, gatewayState === "healthy" ? "good" : gatewayState === "degraded" ? "warn" : "bad"],
      ["Tentativas", total, `${numberFormatter.format(this.summary?.succeeded || 0)} aceitas pela API`, ""],
      ["Taxa de falha", `${percentFormatter.format(failureRate)}%`, `${numberFormatter.format(failed)} falhas`, failed ? "bad" : "good"],
      ["Legacy / unknown", `${percentFormatter.format(this.overview?.legacy_percentage || 0)}%`, `${numberFormatter.format(this.overview?.legacy_unknown || 0)} sem identidade Argo`, this.overview?.legacy_unknown ? "warn" : "good"],
    ];
    this.shadowRoot.querySelector("[data-summary]").replaceChildren(...cards.map(([label, value, note, tone]) => {
      const card = element("article", "card");
      const head = element("div", "metric-head");
      head.append(element("span", "metric-label", label), element("span", `metric-icon ${tone}`, "•"));
      card.append(head, element("div", `metric-value ${tone}`, typeof value === "number" ? numberFormatter.format(value) : value), element("div", "metric-note", note));
      return card;
    }));
  }

  renderGatewayUsage() {
    const renderUsage = (target, items) => {
      if (!items?.length) { target.replaceChildren(element("div", "operation-muted", "Sem tráfego no período")); return; }
      target.replaceChildren(...items.slice(0, 8).map((item) => {
        const row = element("div", "usage-row");
        const info = element("div");
        const failureRate = item.total ? (Number(item.failed || 0) / Number(item.total)) * 100 : 0;
        info.append(element("div", "usage-key", item.key), element("div", "usage-meta", `${percentFormatter.format(failureRate)}% falhas · Pm ${Math.round(item.average_duration_ms || 0)} ms`));
        row.append(info, element("div", "usage-total", numberFormatter.format(item.total || 0)));
        return row;
      }));
    };
    renderUsage(this.shadowRoot.querySelector("[data-application-usage]"), this.overview?.applications);
    renderUsage(this.shadowRoot.querySelector("[data-instance-usage]"), this.overview?.instances);
    const lifecycle = this.overview?.lifecycle || {};
    const signals = [
      ["Taxa de entrega", `${percentFormatter.format(lifecycle.delivery_rate || 0)}%`],
      ["Taxa de leitura", `${percentFormatter.format(lifecycle.read_rate || 0)}%`],
      ["Pendentes envelhecidas", numberFormatter.format(lifecycle.pending_aged || 0)],
      ...((this.overview?.error_categories || []).slice(0, 5).map((item) => [`${item.category || "internal"} · ${item.code || "UNKNOWN"}`, numberFormatter.format(item.count || 0)])),
    ];
    this.shadowRoot.querySelector("[data-gateway-signals]").replaceChildren(...signals.map(([label, value]) => {
      const row = element("div", "usage-row"); row.append(element("div", "usage-key", label), element("div", "usage-total", value)); return row;
    }));
  }

  renderUpstream() {
    const status = this.upstream?.status || "unknown";
    const labels = { up_to_date: "Atualizado", update_available: "Atualização disponível", diverged: "Divergente", unavailable: "Indisponível", unknown: "Aguardando" };
    const pill = this.shadowRoot.querySelector("[data-upstream-status]");
    pill.textContent = labels[status] || status;
    pill.className = `status-pill ${status === "up_to_date" ? "healthy" : status === "update_available" ? "degraded" : status === "unknown" ? "unknown" : "offline"}`;
    const target = this.shadowRoot.querySelector("[data-upstream]");
    target.replaceChildren();
    const rows = [];
    const version = this.upstream?.latest_version || this.upstream?.latest_sha?.slice(0, 7) || "Não verificada";
    rows.push(["Fork baseado em", this.upstream?.baseline_version || this.upstream?.baseline_sha?.slice(0, 7) || "—"]);
    rows.push(["Upstream atual", version]);
    rows.push(["Commits pendentes", numberFormatter.format(this.upstream?.behind_by || 0)]);
    for (const [label, value] of rows) { const row = element("div", "usage-row"); row.append(element("div", "usage-key", label), element("div", "usage-total", value)); target.append(row); }
    if (this.upstream?.error) target.append(element("div", "media-status unavailable", this.upstream.error));
    for (const change of (this.upstream?.changes || []).slice(0, 5)) {
      const link = element("a", "upstream-change", `${change.category || "other"} · ${change.title}`);
      link.href = safeMediaURL(change.url) || "#"; link.target = "_blank"; link.rel = "noopener noreferrer"; target.append(link);
    }
    if (this.upstream?.compare_url) {
      const compare = element("a", "document-action", "Comparar no GitHub ↗"); compare.href = safeMediaURL(this.upstream.compare_url); compare.target = "_blank"; compare.rel = "noopener noreferrer"; target.append(compare);
    }
  }

  renderApplicationFilter() {
    const select = this.shadowRoot.querySelector("[data-application]");
    const current = select.value;
    select.replaceChildren(option("", "Todas as aplicações"));
    for (const application of this.applications) select.append(option(application.slug, application.name));
    if ([...select.options].some((item) => item.value === current)) select.value = current;
  }

  renderApplications() {
    const target = this.shadowRoot.querySelector("[data-applications]");
    this.shadowRoot.querySelector("[data-app-count]").textContent = `${this.applications.length} cadastradas`;
    if (!this.applications.length) {
      const empty = element("div", "empty");
      empty.append(element("strong", "", "Nenhuma integração cadastrada"), element("span", "", "Cadastre o ERP, Athlas ou outro consumidor para identificar os envios."));
      target.replaceChildren(empty);
      return;
    }
    target.replaceChildren(...this.applications.map((application) => {
      const row = element("section", "application-row");
      const top = element("div", "application-top");
      const title = element("div", "application-title");
      title.append(element("div", "application-name", application.name), element("div", "application-slug", application.slug));
      const states = { healthy: "Saudável", degraded: "Degradada", offline: "Sem heartbeat", unknown: "Aguardando heartbeat", disabled: "Desabilitada" };
      const state = application.health_state || (application.active ? "unknown" : "disabled");
      const health = element("span", `status-pill ${state}`, states[state] || state);
      top.append(title, health);
      const meta = element("div", "application-meta");
      const values = [
        ["Ambiente", application.environment || "production"],
        ["Responsável", application.owner || "Não informado"],
        ["Aplicação", application.base_url || "URL não informada"],
        ["Health check", application.health_url || "Não configurado"],
        ["Último heartbeat", relativeTime(application.last_heartbeat_at)],
        ["Sinal reportado", application.last_heartbeat_at ? `${application.last_heartbeat_status || "healthy"} · ${numberFormatter.format(application.last_heartbeat_latency_ms || 0)} ms` : "Ainda não recebido"],
        ["Versão", application.last_heartbeat_version || "Não informada"],
        ["Última chamada API", relativeTime(application.last_seen_at)],
      ];
      for (const [label, value] of values) {
        const item = element("div");
        item.append(element("span", "meta-label", label), element("span", "meta-value", value));
        meta.append(item);
      }
      const actions = element("div", "application-actions");
      const edit = element("button", "button instance-action", "Editar");
      edit.type = "button";
      edit.addEventListener("click", () => this.openForm(application));
      const rotate = element("button", "button instance-action", "Girar credencial");
      rotate.type = "button";
      rotate.addEventListener("click", () => this.rotateCredential(application, rotate));
      actions.append(edit, rotate);
      row.append(top, meta, actions);
      return row;
    }));
  }

  renderErrors() {
    const target = this.shadowRoot.querySelector("[data-errors]");
    const errors = this.summary?.errors || [];
    const select = this.shadowRoot.querySelector("[data-error-code]");
    const current = select.value;
    select.replaceChildren(option("", "Todos os erros"));
    for (const item of errors) select.append(option(item.error_code, `${item.error_code} (${item.count})`));
    if ([...select.options].some((item) => item.value === current)) select.value = current;
    target.replaceChildren(...errors.slice(0, 8).map((item) => {
      const token = element("button", "error-token", `${item.error_code} · ${numberFormatter.format(item.count)}`);
      token.type = "button";
      token.addEventListener("click", () => { select.value = item.error_code; this.loadOperations(); });
      return token;
    }));
  }

  renderAttempts() {
    const target = this.shadowRoot.querySelector("[data-attempts]");
    this.shadowRoot.querySelector("[data-attempt-count]").textContent = `${numberFormatter.format(this.attempts.length)} operações`;
    if (!this.attempts.length) {
      const row = element("tr");
      const cell = element("td", "operation-muted", "Nenhuma tentativa encontrada para os filtros selecionados.");
      cell.colSpan = 7;
      row.append(cell);
      target.replaceChildren(row);
      return;
    }
    target.replaceChildren(...this.attempts.map((attempt) => {
      const row = element("tr");
      const time = element("td", "", dateTimeFormatter.format(new Date(attempt.started_at)));
      const application = element("td", "mono", attempt.application_slug || "legacy/unknown");
      application.title = attempt.identity_verified ? "Identidade verificada" : "Identidade não verificada";
      const endpoint = element("td", "mono", attempt.endpoint || "-");
      const result = element("td", attempt.outcome === "succeeded" ? "operation-ok" : "operation-error", attempt.outcome === "succeeded" ? `${attempt.http_status} Aceita` : `${attempt.http_status} Falha`);
      const duration = element("td", "", `${numberFormatter.format(attempt.duration_ms || 0)} ms`);
      const error = element("td", attempt.error_code ? "operation-error" : "operation-muted", attempt.error_code || "-");
      error.title = attempt.error_detail || "";
      const correlation = element("td", "mono", attempt.correlation_id || "-");
      correlation.title = attempt.correlation_id || "";
      row.append(time, application, endpoint, result, duration, error, correlation);
      return row;
    }));
  }

  openForm(application = null) {
    this.editing = application;
    const form = this.shadowRoot.querySelector("[data-form]");
    form.reset();
    form.elements.active.checked = application ? Boolean(application.active) : true;
    form.elements.expected_heartbeat_seconds.value = application?.expected_heartbeat_seconds || 300;
    for (const key of ["name", "slug", "environment", "owner", "base_url", "health_url"]) {
      if (application && form.elements[key]) form.elements[key].value = application[key] || "";
    }
    form.elements.slug.disabled = Boolean(application);
    this.shadowRoot.querySelector("[data-form-title]").textContent = application ? `Editar ${application.name}` : "Nova integração";
    this.shadowRoot.querySelector("[data-form-dialog]").hidden = false;
    form.elements.name.focus();
  }

  closeForm() {
    this.shadowRoot.querySelector("[data-form-dialog]").hidden = true;
    this.editing = null;
  }

  async saveApplication(event) {
    event.preventDefault();
    const form = event.currentTarget;
    const button = this.shadowRoot.querySelector("[data-save]");
    const body = Object.fromEntries(new FormData(form).entries());
    body.active = form.elements.active.checked;
    body.expected_heartbeat_seconds = Number(body.expected_heartbeat_seconds || 300);
    if (this.editing) body.slug = this.editing.slug;
    button.disabled = true;
    button.textContent = "Salvando...";
    this.setError("");
    try {
      const path = this.editing ? `/argo/v1/applications/${encodeURIComponent(this.editing.id)}` : "/argo/v1/applications";
      const result = await request(path, {}, undefined, { method: this.editing ? "PUT" : "POST", body });
      this.closeForm();
      if (result?.credential) this.showCredential(result.application, result.credential);
      await this.load();
    } catch (error) {
      this.setError(error.message);
    } finally {
      button.disabled = false;
      button.textContent = "Salvar integração";
    }
  }

  async rotateCredential(application, button) {
    if (!window.confirm(`Girar a credencial de ${application.name}? A credencial atual deixará de identificar novas chamadas.`)) return;
    button.disabled = true;
    button.textContent = "Gerando...";
    this.setError("");
    try {
      const result = await request(`/argo/v1/applications/${encodeURIComponent(application.id)}/rotate-credential`, {}, undefined, { method: "POST" });
      this.showCredential(result.application, result.credential);
    } catch (error) {
      this.setError(error.message);
    } finally {
      button.disabled = false;
      button.textContent = "Girar credencial";
    }
  }

  showCredential(application, credential) {
    this.currentCredential = credential;
    this.shadowRoot.querySelector("[data-credential-app]").textContent = application?.name || "Integração";
    this.shadowRoot.querySelector("[data-credential]").textContent = credential;
    this.shadowRoot.querySelector("[data-copy-credential]").textContent = "Copiar credencial";
    this.shadowRoot.querySelector("[data-credential-dialog]").hidden = false;
  }

  closeCredential() {
    this.shadowRoot.querySelector("[data-credential-dialog]").hidden = true;
    this.currentCredential = "";
    this.shadowRoot.querySelector("[data-credential]").textContent = "";
  }

  async copyCredential() {
    if (!this.currentCredential) return;
    try {
      await navigator.clipboard.writeText(this.currentCredential);
      this.shadowRoot.querySelector("[data-copy-credential]").textContent = "Copiada";
    } catch {
      this.setError("Não foi possível copiar automaticamente. Selecione a credencial e copie manualmente.");
    }
  }
}

class ArgoMessageLifecycle extends ArgoBaseElement {
  connectedCallback() {
    this.applications = [];
    this.events = [];
    this.summary = {};
    this.render();
    this.loadInitial();
  }

  render() {
    this.shadowRoot.innerHTML = `
      <style>${styles}</style>
      <main class="page lifecycle-page">
        <header class="header">
          <div class="brand"><img src="/assets/argo-brand.png" alt="Argo" /><div><p class="eyebrow">Operação de mensageria</p><h1>Ciclo das mensagens</h1><p class="subtitle">Do recebimento da API à leitura no WhatsApp, com correlação e evidências imutáveis.</p></div></div>
          <div class="toolbar"><button class="button" data-refresh type="button">Atualizar</button></div>
        </header>
        <div class="error" data-error hidden></div>
        <div class="operations-toolbar panel">
          <div class="operation-filters">
            <select data-period><option value="1">24 horas</option><option value="7" selected>7 dias</option><option value="30">30 dias</option><option value="90">90 dias</option></select>
            <select data-application><option value="">Todas as aplicações</option></select>
            <input data-instance placeholder="ID da instância" />
            <select data-type><option value="">Todos os tipos</option><option value="text">Texto</option><option value="link">Link</option><option value="media">Mídia</option><option value="poll">Enquete</option><option value="sticker">Sticker</option><option value="location">Localização</option><option value="contact">Contato</option><option value="button">Botão</option></select>
            <select data-state><option value="">Todos os estados</option><option value="received">Recebida</option><option value="accepted">Aceita</option><option value="sent">Enviada</option><option value="delivered">Entregue</option><option value="read">Lida</option><option value="failed">Falha</option><option value="pending_aged">Pendente envelhecida</option></select>
          </div>
        </div>
        <section class="lifecycle-funnel" data-funnel></section>
        <section class="lifecycle-alerts" data-alerts></section>
        <section class="lifecycle-layout">
          <article class="panel">
            <div class="panel-header"><div><h2 class="panel-title">Transições recentes</h2><p class="panel-subtitle">Clique em uma transição para abrir a linha do tempo correlacionada</p></div><span class="status-pill" data-event-count>0</span></div>
            <div class="table-wrap"><table class="operations-table"><thead><tr><th>Horário</th><th>Estado</th><th>Aplicação</th><th>Instância</th><th>Tipo</th><th>Message ID</th><th>Correlação</th></tr></thead><tbody data-events></tbody></table></div>
          </article>
          <aside class="panel">
            <div class="panel-header"><div><h2 class="panel-title">Latências</h2><p class="panel-subtitle">P50, P90, P95 e P99 em milissegundos</p></div></div>
            <div class="panel-body latency-list" data-latencies></div>
            <div class="panel-header"><div><h2 class="panel-title">Falhas conhecidas</h2><p class="panel-subtitle">Categorias operacionais no período</p></div></div>
            <div class="panel-body failure-list" data-failures></div>
          </aside>
        </section>
      </main>
      <div class="dialog-backdrop" data-timeline-dialog hidden>
        <section class="dialog" role="dialog" aria-modal="true" aria-labelledby="lifecycle-timeline-title">
          <div class="dialog-header"><div><h2 class="panel-title" id="lifecycle-timeline-title">Linha do tempo</h2><p class="panel-subtitle mono" data-timeline-id></p></div><button class="button instance-action" data-close-timeline type="button">Fechar</button></div>
          <div class="dialog-body lifecycle-timeline" data-timeline></div>
        </section>
      </div>`;
    this.shadowRoot.querySelector("[data-refresh]").addEventListener("click", () => this.load());
    for (const selector of ["[data-period]", "[data-application]", "[data-type]", "[data-state]"]) {
      this.shadowRoot.querySelector(selector).addEventListener("change", () => this.load());
    }
    let timer;
    this.shadowRoot.querySelector("[data-instance]").addEventListener("input", () => {
      clearTimeout(timer);
      timer = setTimeout(() => this.load(), 350);
    });
    this.shadowRoot.querySelector("[data-close-timeline]").addEventListener("click", () => { this.shadowRoot.querySelector("[data-timeline-dialog]").hidden = true; });
  }

  filters() {
    return {
      ...periodRange(this.shadowRoot.querySelector("[data-period]").value),
      application: this.shadowRoot.querySelector("[data-application]").value,
      instanceId: this.shadowRoot.querySelector("[data-instance]").value.trim(),
      type: this.shadowRoot.querySelector("[data-type]").value,
      state: this.shadowRoot.querySelector("[data-state]").value,
      limit: 200,
    };
  }

  async loadInitial() {
    const signal = this.disconnectController();
    try {
      const applications = await request("/argo/v1/applications", {}, signal);
      this.applications = Array.isArray(applications) ? applications : [];
      const select = this.shadowRoot.querySelector("[data-application]");
      for (const application of this.applications) select.append(option(application.slug, application.name));
      await this.load(signal);
    } catch (error) {
      if (error.name !== "AbortError") this.setError(error.message);
    }
  }

  async load(existingSignal) {
    const signal = existingSignal || this.disconnectController();
    this.setError("");
    try {
      const filters = this.filters();
      const summaryFilters = { ...filters };
      delete summaryFilters.state;
      const [summary, events] = await Promise.all([
        request("/argo/v1/messages/lifecycle/summary", summaryFilters, signal),
        request("/argo/v1/messages/lifecycle", filters, signal),
      ]);
      this.summary = summary || {};
      this.events = Array.isArray(events) ? events : [];
      this.renderSummary();
      this.renderEvents();
      this.renderLatencies();
      this.renderFailures();
    } catch (error) {
      if (error.name !== "AbortError") this.setError(error.message);
    }
  }

  renderSummary() {
    const stages = [
      ["Recebidas", "received", "100% da entrada"],
      ["Aceitas", "accepted", `${percentFormatter.format(this.summary.acceptance_rate || 0)}% de aceitação`],
      ["Enviadas", "sent", `${percentFormatter.format(this.summary.send_rate || 0)}% das aceitas`],
      ["Entregues", "delivered", `${percentFormatter.format(this.summary.delivery_rate || 0)}% das enviadas`],
      ["Lidas", "read", `${percentFormatter.format(this.summary.read_rate || 0)}% das entregues`],
    ];
    this.shadowRoot.querySelector("[data-funnel]").replaceChildren(...stages.map(([label, key, note]) => {
      const stage = element("article", "funnel-stage");
      stage.append(element("small", "", label), element("strong", "", numberFormatter.format(this.summary[key] || 0)), element("small", "", note));
      return stage;
    }));
    const alerts = [
      ["Falhas", this.summary.failed || 0, "Estados terminais conhecidos", "bad"],
      ["Pendentes envelhecidas", this.summary.pending_aged || 0, `Sem receipt após ${this.summary.pending_age_minutes || 15} min`, "warn"],
    ];
    this.shadowRoot.querySelector("[data-alerts]").replaceChildren(...alerts.map(([label, value, note, tone]) => {
      const card = element("article", "alert-card");
      const copy = element("div");
      copy.append(element("div", "metric-label", label), element("div", "metric-note", note));
      card.append(copy, element("strong", tone, numberFormatter.format(value)));
      return card;
    }));
  }

  renderEvents() {
    const target = this.shadowRoot.querySelector("[data-events]");
    this.shadowRoot.querySelector("[data-event-count]").textContent = `${numberFormatter.format(this.events.length)} eventos`;
    if (!this.events.length) {
      const row = element("tr");
      const cell = element("td", "operation-muted", "Nenhuma transição encontrada para os filtros selecionados.");
      cell.colSpan = 7;
      row.append(cell);
      target.replaceChildren(row);
      return;
    }
    target.replaceChildren(...this.events.map((event) => {
      const row = element("tr");
      row.tabIndex = 0;
      row.title = "Abrir linha do tempo";
      row.addEventListener("click", () => this.openTimeline(event));
      row.addEventListener("keydown", (keyboardEvent) => { if (keyboardEvent.key === "Enter") this.openTimeline(event); });
      const stateClass = ["failed", "pending_aged"].includes(event.state) ? "operation-error" : "operation-ok";
      row.append(
        element("td", "", dateTimeFormatter.format(new Date(event.occurred_at))),
        element("td", stateClass, event.state),
        element("td", "mono", event.application_slug || "legacy/unknown"),
        element("td", "mono", event.instance_id || "-"),
        element("td", "", event.message_type || "-"),
        element("td", "mono", event.provider_message_id || "-"),
        element("td", "mono", event.correlation_id || "-"),
      );
      return row;
    }));
  }

  renderLatencies() {
    const rows = [["API → envio", this.summary.send_latency], ["Envio → entrega", this.summary.delivery_latency], ["Entrega → leitura", this.summary.read_latency]];
    this.shadowRoot.querySelector("[data-latencies]").replaceChildren(...rows.map(([label, values = {}]) => {
      const row = element("div", "latency-row");
      row.append(element("strong", "", label));
      for (const key of ["p50_ms", "p90_ms", "p95_ms", "p99_ms"]) row.append(element("span", "", `${key.slice(0, 3).toUpperCase()} ${numberFormatter.format(Math.round(values?.[key] || 0))}`));
      return row;
    }));
  }

  renderFailures() {
    const failures = this.summary.failures || [];
    const target = this.shadowRoot.querySelector("[data-failures]");
    if (!failures.length) {
      target.replaceChildren(element("div", "operation-muted", "Nenhuma falha no período."));
      return;
    }
    target.replaceChildren(...failures.slice(0, 10).map((failure) => {
      const row = element("div", "alert-card");
      const copy = element("div");
      copy.append(element("div", "metric-label", failure.category || "internal"), element("div", "metric-note mono", failure.code || "INTERNAL_ERROR"));
      row.append(copy, element("strong", "bad", numberFormatter.format(failure.count || 0)));
      return row;
    }));
  }

  async openTimeline(event) {
    const dialog = this.shadowRoot.querySelector("[data-timeline-dialog]");
    const target = this.shadowRoot.querySelector("[data-timeline]");
    dialog.hidden = false;
    this.shadowRoot.querySelector("[data-timeline-id]").textContent = event.provider_message_id || event.correlation_id || event.attempt_id || "Sem identificador";
    target.replaceChildren(element("div", "empty", "Carregando linha do tempo..."));
    try {
      const filters = { ...periodRange(366), limit: 100 };
      if (event.provider_message_id) filters.providerMessageId = event.provider_message_id;
      else if (event.correlation_id) filters.correlationId = event.correlation_id;
      const timeline = await request("/argo/v1/messages/lifecycle", filters);
      const ordered = (Array.isArray(timeline) ? timeline : []).sort((a, b) => new Date(a.occurred_at) - new Date(b.occurred_at));
      target.replaceChildren(...ordered.map((item) => {
        const row = element("div", `timeline-event ${item.state}`);
        const copy = element("div");
        copy.append(element("strong", "", item.state), element("div", "metric-note", item.failure_code ? `${item.failure_code} · ${item.failure_detail || ""}` : `${item.application_slug || "legacy/unknown"} · ${item.message_type || "mensagem"}`));
        row.append(copy, element("time", "", dateTimeFormatter.format(new Date(item.occurred_at))));
        return row;
      }));
    } catch (error) {
      target.replaceChildren(element("div", "error", error.message));
    }
  }
}

class ArgoConversations extends ArgoBaseElement {
  connectedCallback() {
    this.instances = [];
    this.conversations = [];
    this.selected = null;
    this.messages = [];
    this.searchTimer = null;
    this.mediaObjectURLs = new Set();
    this.render();
    this.loadInitial();
  }

  disconnectedCallback() {
    super.disconnectedCallback();
    clearTimeout(this.searchTimer);
    this.releaseMediaObjectURLs();
  }

  releaseMediaObjectURLs() {
    for (const url of this.mediaObjectURLs || []) URL.revokeObjectURL(url);
    this.mediaObjectURLs?.clear();
  }

  async loadMediaPreview(message, card, button) {
    button.disabled = true;
    button.textContent = "Carregando...";
    const status = card.querySelector(".media-status");
    try {
      const media = await fetchCapturedMedia(message);
      const source = media.blobURL || media.fallbackURL;
      if (media.blobURL) this.mediaObjectURLs.add(media.blobURL);
      const type = String(message.message_type || "").toLowerCase();
      let preview;
      if (type === "image" || type === "sticker") {
        preview = document.createElement("img");
        preview.alt = message.caption || message.file_name || "Imagem recebida";
      } else if (type === "audio") {
        preview = document.createElement("audio");
        preview.controls = true;
      } else {
        preview = document.createElement("video");
        preview.controls = true;
        preview.playsInline = true;
      }
      preview.src = source;
      const container = element("div", "media-preview");
      container.append(preview);
      card.append(container);
      status.textContent = "Prévia carregada";
      button.remove();
    } catch {
      status.textContent = "Mídia indisponível para prévia ou download";
      status.classList.add("unavailable");
      button.textContent = "Tentar novamente";
      button.disabled = false;
    }
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

  key(conversation) { return `${conversation.instance_id}:${conversationChatJID(conversation)}`; }

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
    const chatJid = conversationChatJID(conversation);
    if (!chatJid) {
      this.setError("A conversa nao possui um identificador WhatsApp valido. Atualize a lista e tente novamente.");
      return;
    }
    this.selected = conversation;
    this.messagePage = null;
    this.renderConversations();
    const header = this.shadowRoot.querySelector("[data-timeline-header]");
    header.replaceChildren(element("p", "timeline-name", displayName(conversation)), element("p", "timeline-id", `${chatJid} · ${conversation.instance_name}`));
    const target = this.shadowRoot.querySelector("[data-messages]");
    target.replaceChildren(element("div", "empty", "Carregando mensagens..."));
    const signal = this.disconnectController();
    try {
      const page = await request("/analytics/messages", {
        ...this.filters(),
        instanceId: conversation.instance_id,
        chatJid,
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
        chatJid: conversationChatJID(this.selected),
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
    this.releaseMediaObjectURLs();
    if (!this.messages.length) {
      target.replaceChildren(element("div", "empty", "Nenhuma mensagem encontrada neste periodo."));
      return;
    }
    const nodes = this.messages.map((message) => {
      const bubble = element("article", `message ${message.direction === "outbound" ? "outbound" : ""}`);
      if (message.direction !== "outbound" && (message.push_name || message.participant_jid)) bubble.append(element("div", "sender", message.push_name || message.participant_jid));
      bubble.append(element("div", "message-text", messageBody(message)));
      const mediaType = String(message.message_type || "").toLowerCase();
      if (["document", "image", "audio", "video", "sticker"].includes(mediaType)) {
        const card = element("div", "document-card");
        const labels = { document: message.mime_type === "application/pdf" ? "PDF" : "DOC", image: "IMG", audio: "AUD", video: "VID", sticker: "STK" };
        card.append(element("div", "document-icon", labels[mediaType]));
        const info = element("div", "document-info");
        const names = { document: "Documento", image: "Imagem", audio: "Áudio", video: "Vídeo", sticker: "Figurinha" };
        info.append(element("div", "document-name", message.file_name || names[mediaType]), element("div", "document-size", `${message.mime_type || names[mediaType]} · ${formatFileSize(message.file_size)}`));
        const actions = element("div", "document-actions");
        const view = element("button", "document-action", mediaType === "document" ? "Visualizar" : "Carregar prévia");
        view.type = "button";
        if (mediaType === "document") view.addEventListener("click", () => openCapturedMedia(message).catch((error) => this.setError(error.message)));
        else view.addEventListener("click", () => this.loadMediaPreview(message, card, view));
        const download = element("button", "document-action", "Baixar");
        download.type = "button";
        download.addEventListener("click", () => openCapturedMedia(message, true).catch(() => {
          const status = card.querySelector(".media-status");
          status.textContent = "Mídia indisponível para prévia ou download";
          status.classList.add("unavailable");
        }));
        actions.append(view, download);
        info.append(actions);
        card.append(info);
        card.append(element("div", "media-status", "Disponível sob demanda"));
        bubble.append(card);
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
if (!customElements.get("argo-integrations")) customElements.define("argo-integrations", ArgoIntegrations);
if (!customElements.get("argo-message-lifecycle")) customElements.define("argo-message-lifecycle", ArgoMessageLifecycle);
if (!customElements.get("argo-conversations")) customElements.define("argo-conversations", ArgoConversations);
