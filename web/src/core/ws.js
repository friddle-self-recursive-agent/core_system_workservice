/* ws.js —— GET /v1/stream WebSocket 客户端（契约：连接即推 {type:"snapshot"}，之后
 * {type:"patch", patch:[RFC6902], seq:N}；服务端 30s ping）。
 * 状态机：connecting → connected ⇄ reconnecting →（连续 maxFail 次失败）degraded。
 *   - 指数退避重连：1s 起步 ×2，封顶 30s；
 *   - degraded=降级轮询兜底：fetch /v1/snapshot 每 pollMs（默认 30s）一次；
 *   - 页面隐藏（document.hidden）=暂停轮询与重连退避（visibilitychange 监听），可见即恢复；
 *     恢复时若处于 degraded 仍先试一次 WS（onopen 成功即撤轮询——WS 优先不放弃）。
 * 纯客户端传输层——不持有业务状态；onSnapshot/onPatch/onState 全部回调出去。 */

/** ws(s)://host/v1/stream（由页面 location 推导；base/loc 可注入便于测试）。 */
export function wsUrlOf(loc, path = 'v1/stream') {
  const l = loc || (typeof location !== 'undefined' ? location : null);
  if (!l || !l.host) return '/' + String(path).replace(/^\/+/, '');
  const proto = l.protocol === 'https:' ? 'wss://' : 'ws://';
  return proto + l.host + '/' + String(path).replace(/^\/+/, '');
}

function hidden() {
  return typeof document !== 'undefined' && !!document.hidden;
}

export function connectStream(opts = {}) {
  const {
    url = wsUrlOf(),
    base = '',
    onSnapshot,
    onPatch,
    onState,
    pollMs = 30000,
    maxFail = 5,
  } = opts;
  const B = String(base || '').replace(/\/+$/, '');
  let ws = null;
  let state = 'connecting';
  let fails = 0;
  let backoff = 1000;
  let timer = null;          // 重连退避定时器
  let pollTimer = null;      // 降级轮询定时器
  let pendingReconnect = false; // 隐藏期间挂起的重连
  let closed = false;

  function setState(s) {
    if (state === s) return;
    state = s;
    try { if (onState) onState(s); } catch { /* 回调异常不毁传输层 */ }
  }
  function clearTimers() {
    if (timer) { clearTimeout(timer); timer = null; }
    if (pollTimer) { clearInterval(pollTimer); pollTimer = null; }
  }

  /* ---------- 降级轮询（WS 不可用的兜底数据面） ---------- */
  function pollOnce() {
    if (closed || hidden()) return; // 页面隐藏=暂停轮询
    fetch(B + '/v1/snapshot', { credentials: 'same-origin' })
      .then((r) => (r.ok ? r.json() : null))
      .then((j) => {
        if (!j || closed) return;
        fails = 0; // 轮询成功不自动升回 WS（下一拍可见性恢复时再试）——保守不抖动
        try { if (onSnapshot) onSnapshot(j); } catch { /* 同上 */ }
      })
      .catch(() => { /* 轮询失败=保持 degraded，下一拍再试（不缓存失败态） */ });
  }
  function startPolling() {
    setState('degraded');
    if (pollTimer) return;
    pollOnce();
    pollTimer = setInterval(pollOnce, pollMs);
  }

  /* ---------- 重连（指数退避；隐藏=挂起） ---------- */
  function scheduleReconnect() {
    if (closed) return;
    if (hidden()) { pendingReconnect = true; return; }
    if (timer) return;
    timer = setTimeout(() => { timer = null; open(); }, backoff);
    backoff = Math.min(backoff * 2, 30000);
  }

  function open() {
    if (closed) return;
    setState(fails > 0 ? 'reconnecting' : 'connecting');
    let sock;
    try { sock = new WebSocket(url); } catch { onFailure(); return; }
    ws = sock;
    sock.onopen = () => {
      fails = 0;
      backoff = 1000;
      if (pollTimer) { clearInterval(pollTimer); pollTimer = null; } // WS 优先——轮询撤防
      setState('connected');
    };
    sock.onmessage = (ev) => {
      let msg = null;
      try { msg = JSON.parse(String(ev.data)); } catch { return; }
      if (!msg || !msg.type) return;
      if (msg.type === 'snapshot') {
        try { if (onSnapshot) onSnapshot(msg.snapshot !== undefined ? msg.snapshot : msg.data !== undefined ? msg.data : msg); } catch { /* 回调异常不毁传输层 */ }
      } else if (msg.type === 'patch') {
        try { if (onPatch) onPatch(Array.isArray(msg.patch) ? msg.patch : [], msg); } catch { /* 同上 */ }
      } else if (msg.type === 'ping') {
        try { if (sock.readyState === 1) sock.send(JSON.stringify({ type: 'pong' })); } catch { /* 保活失败下拍再来 */ }
      }
      // pong/未知类型=静默（30s ping 保活语义）
    };
    sock.onclose = () => { if (ws === sock) { ws = null; onFailure(); } };
    sock.onerror = () => { /* 错误必伴随 close——统一走 onFailure */ };
  }

  function onFailure() {
    if (closed) return;
    fails += 1;
    if (fails >= maxFail) startPolling(); // 连续失败=降级（轮询期间不再空转重连）
    else scheduleReconnect();
  }

  /* ---------- 可见性恢复（暂停面统一入口） ---------- */
  function resume() {
    if (closed) return;
    if (pendingReconnect) { pendingReconnect = false; scheduleReconnect(); return; }
    if (ws) { if (state === 'degraded') pollOnce(); return; }
    if (state === 'degraded') { backoff = 1000; open(); return; } // 降级面恢复时先试 WS
    if (state !== 'connected') { open(); }
  }
  function onVisibility() {
    if (hidden()) {
      if (timer) { clearTimeout(timer); timer = null; pendingReconnect = true; }
    } else {
      resume();
    }
  }
  if (typeof document !== 'undefined') {
    document.addEventListener('visibilitychange', onVisibility);
  }

  open();
  return {
    close() {
      closed = true;
      clearTimers();
      if (typeof document !== 'undefined') document.removeEventListener('visibilitychange', onVisibility);
      if (ws) { try { ws.close(); } catch { /* 已断 */ } ws = null; }
    },
    get state() { return state; },
    send(obj) {
      if (ws && ws.readyState === 1) {
        try { ws.send(JSON.stringify(obj)); return true; } catch { return false; }
      }
      return false;
    },
  };
}
