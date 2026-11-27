/* chat.js —— 会话屏 v4（2026-11-26 C 批）：DeepSeek 对话版型 × 通路接满。
 * 数据面=window.__CC.chatFaces（plugin 转录桥）：sessions（会话清单）/create（宿主发 id——
 * 零手填）/transcript（durable journal 词形投影）/state（running/queue/queueDepth/currentTool/
 * currentToolSince/lastSeq——C-03 活体）/send({mode})（C-01 steer|queue）/stop（C-02 cancel）。
 * 交互纪律（C-04）：转录增量渲染——只 append seq>lastSeq 新节点；贴底才跟随，上翻时右下浮标
 * 「↓ N 条新消息」点击回底；轮询 3.5s 只动转录区与会话行（composer 焦点保真）。
 * 销账纪律（C-05）：pending={text, requestId, at}——按 requestId 主匹配销账、文本兜底；
 * 失败行内「重试」钮。受理词形如实区分：运行中=steer（立即注入）、空闲=queue（排队落投）。
 * 视觉（C-B）：三段式全高布局/转录列居中 800px/用户泡 panel-2/助手全宽无泡/思考与工具折叠/
 * 空态示例 chips 真实可发/零 inline style——chat 域样式唯一真源=styles/chat.css。
 * 转录映射（dsh-client-ui-chat conversation-nodes 同族词形）：user/message(source.user)→用户泡；
 * assistant/message→AI 消息（text=md 渲染；reasoning=已深度思考折叠）；tool/call+tool/result 按
 * callId 归组=已调用 X 折叠行；system/developer/message=灰行；delta/live-chunk 已由桥剥除。 */

import { esc, trunc, relAge } from '../core/format.js';
import { chip } from '../render/chips.js';
import { linkifyRefs } from '../render/refs.js';
import { renderMarkdown } from '../render/md.js';
import { mountSessionCtx } from './sessionctx.js';
import { t as tr, getLocale } from '../core/i18n.js';

function faces() {
  return (typeof window !== 'undefined' && window.__CC && window.__CC.chatFaces) || null;
}

/* ---------- 纯函数（node 单测机械直测） ---------- */

/** content 块 → {text, reasoning, toolCalls[]}（user/assistant 两族共用块词形）。 */
export function blocksOf(content) {
  const out = { text: '', reasoning: '', toolCalls: [] };
  if (!Array.isArray(content)) return out;
  for (const b of content) {
    if (!b || typeof b !== 'object') continue;
    if (b.type === 'text' && typeof b.text === 'string') out.text += (out.text ? '\n' : '') + b.text;
    else if (b.type === 'reasoning' && typeof b.text === 'string') out.reasoning += (out.reasoning ? '\n' : '') + b.text;
    else if (b.type === 'tool-call') out.toolCalls.push({ id: String(b.id || b.callId || ''), name: String(b.name || ''), argsRaw: typeof b.arguments === 'string' ? b.arguments : JSON.stringify(b.arguments ?? b.args ?? '') });
  }
  return out;
}

/** journal 事件 → 渲染节点族（seq 升序；工具组按首次 call 顺序排在其后）。 */
export function mapTranscript(events) {
  const nodes = [];
  const tools = new Map(); // callId → {seq, name, args, result}
  const order = [];
  const rows = Array.isArray(events) ? [...events] : [];
  rows.sort((a, b) => (Number(a && a.seq) || 0) - (Number(b && b.seq) || 0));
  for (const ev of rows) {
    if (!ev || typeof ev !== 'object') continue;
    const type = String(ev.type || '');
    const data = (ev.data && typeof ev.data === 'object') ? ev.data : {};
    const seq = Number(ev.seq) || 0;
    const time = Number(ev.time) || 0;
    if (type === 'user/message') {
      const src = (data.source && typeof data.source === 'object') ? String(data.source.kind || '') : '';
      if (src === 'compact-checkpoint') continue; // 压缩检查点=替换面事件，不渲染
      const b = blocksOf(data.content);
      nodes.push({ kind: 'user', seq, time, text: b.text, injected: src !== 'user' });
    } else if (type === 'assistant/message') {
      const b = blocksOf(data.content);
      for (const tc of b.toolCalls) order.push(tc.id);
      // 工具轮次（只有 tool-call、无 text/reasoning）不生成助手节点——
      // 工具本身会作为 kind:'tool' 节点渲染（「已调用 bash」），
      // 否则每条工具调用都多一个「（空回复）」气泡，转录被空壳刷屏（2026-11-27 实测 bug）。
      if (!b.text && !b.reasoning && b.toolCalls.length) continue;
      nodes.push({ kind: 'assistant', seq, time, text: b.text, reasoning: b.reasoning });
    } else if (type === 'tool/call') {
      const id = String(data.callId || data.id || '');
      if (!tools.has(id)) tools.set(id, { seq, time, name: String(data.name || ''), args: typeof data.arguments === 'string' ? data.arguments : JSON.stringify(data.arguments ?? data.args ?? ''), result: '' });
      order.push(id);
    } else if (type === 'tool/result') {
      const id = String(data.callId || data.id || '');
      const t = tools.get(id);
      if (t) t.result = typeof data.content === 'string' ? data.content : (data.output != null ? String(typeof data.output === 'string' ? data.output : JSON.stringify(data.output)) : '');
    } else if (type === 'system/message' || type === 'developer/message') {
      nodes.push({ kind: 'system', seq, time, text: String(data.text ?? data.content ?? ''), dev: type === 'developer/message' });
    }
  }
  // callId 去重：assistant 消息里的 tool-call block 与独立 tool/call 事件可能同时携带
  // 同一 callId，不去重会渲染两个「已调用 X」节点（2026-11-27 实测）。
  const seenToolIds = new Set();
  for (const id of order) {
    if (seenToolIds.has(id)) continue;
    seenToolIds.add(id);
    const t = tools.get(id);
    if (!t) continue;
    nodes.push({ kind: 'tool', seq: t.seq, time: t.time, name: t.name || tr('工具'), args: t.args, result: t.result });
  }
  nodes.sort((a, b) => a.seq - b.seq);
  return nodes;
}

/** 增量分派（C-04）：nodeKey(seq+kind) 已渲染的不重出；返回真正的新节点尾片。
 * 全量分片词形（firstPaint=true 或 seq 回退）=整列重画（路由切换/重连兜底）。 */
export function nextBatch(nodes, paintedKeys, firstPaint) {
  const rows = Array.isArray(nodes) ? nodes : [];
  if (firstPaint) return rows;
  const out = rows.filter((n) => !paintedKeys.has(nodeKey(n)));
  return out;
}

/** 节点渲染去重键（seq+kind——同 seq 双族不互撞）。 */
export function nodeKey(n) {
  return String(Number(n && n.seq) || 0) + ':' + String((n && n.kind) || '');
}

/** 发送模式判别（C-01）：state.running=true → steer（立即注入），否则 queue（排队落投）。
 * 纯函数——判别词形单点，测试锚 = 两分支词形。 */
export function modeForState(st) {
  return !!(st && st.running === true) ? 'steer' : 'queue';
}

/** 会话行分组键（C-06）：today / earlier 两词形（宿主 updatedAt 毫秒）。 */
export function dayGroupOf(updatedAt, nowMs) {
  const t = Number(updatedAt) || 0;
  if (!t) return 'earlier';
  const now = Number(nowMs) || Date.now();
  const startOf = new Date(now); startOf.setHours(0, 0, 0, 0);
  return (now - t) < 24 * 3600 * 1000 && startOf.getTime() <= t ? 'today' : 'earlier';
}

/* ---------- 屏内存态 ---------- */

const CHAT = {
  sessionId: '',       // 当前会话（宿主发 id——零手填）
  timer: 0,            // 轮询定时器（导航离开即清）
  pending: [],         // 待达回声 {text, requestId, at}（C-05 按 requestId 销账）
  lastSeq: 0,          // 已渲染转录尾（增量分派锚）
  painted: new Set(),  // 已渲染节点键（nodeKey 族）
  seenSeq: new Map(),  // 会话 id → 已见尾 seq（未读点 highwater）
  state: null,         // 最近一次 chat-state 词形
  el: null,            // 当前屏根（路由切换换引用）
};

const LAST = { route: '', paramKey: '', locale: '' };

/* ---------- 渲染 ---------- */

function nodeHtml(n, names) {
  if (n.kind === 'user') {
    const pending = n.pending
      ? '<div class="chat-pending">' + esc(tr('送达中…')) + '</div>'
      : (n.injected ? '<div class="chat-pending">' + esc(tr('注入上下文')) + '</div>' : '');
    const err = n.error
      ? '<div class="chat-err"><span>' + esc(tr('送达失败（') + String(n.error) + tr('）')) + '</span>' +
        '<button class="btn btn-ghost sm" data-cc-retry="' + esc(String(n.requestId || '')) + '">' + esc(tr('重试')) + '</button></div>'
      : '';
    return '<div class="chat-msg-user" data-node="' + esc(nodeKey(n)) + '">' +
      '<div class="chat-bubble">' + (n.text ? renderMarkdown(n.text) : '<span class="faint">…</span>') + '</div>' +
      (pending || err ? '<div>' + pending + err + '</div>' : '') +
      '</div>';
  }
  if (n.kind === 'assistant') {
    // 无正文助手节点不渲染（工具轮次已在 mapTranscript 过滤，这里是纯空流防御）。
    if (!n.text && !n.reasoning) return '';
    const reasoning = n.reasoning
      ? '<details class="chat-fold"><summary><span class="ex-chev">▸</span>' + esc(tr('已深度思考')) + '</summary>' +
        '<div class="ex-body">' + linkifyRefs(n.reasoning, names) + '</div></details>'
      : '';
    const body = n.text ? '<div class="chat-md">' + renderMarkdown(n.text) + '</div>' : '';
    return '<div class="chat-msg-ai" data-node="' + esc(nodeKey(n)) + '">' + reasoning + body + '</div>';
  }
  if (n.kind === 'tool') {
    const detail = esc(trunc(String(n.args || ''), 800)) +
      (n.result ? '\n' + tr('──── 结果 ────') + '\n' + esc(trunc(String(n.result), 800)) : '');
    const done = n.result
      ? '<span class="chat-tool-ok">' + esc(tr('完成')) + '</span>'
      : '<span class="chat-tool-none">' + esc(tr('执行中…')) + '</span>';
    return '<details class="chat-fold chat-tool" data-node="' + esc(nodeKey(n)) + '">' +
      '<summary><span class="ex-chev">▸</span>' + esc(tr('已调用')) + ' <span class="chat-tool-name">' + esc(n.name) + '</span> ' + done + '</summary>' +
      '<div class="ex-body">' + detail + '</div></details>';
  }
  const label = (n.dev ? tr('开发者') + ' · ' : '') + trunc(String(n.text || ''), 200);
  return '<div class="chat-sysrow" data-node="' + esc(nodeKey(n)) + '">' + esc(label) + '</div>';
}

/** 待达回声 → 渲染节点（无 seq——气泡恒显示，入账即摘）。 */
function pendingNodes() {
  return CHAT.pending.map((p) => ({ kind: 'user', seq: Number.MAX_SAFE_INTEGER, text: p.text, pending: true, requestId: p.requestId }));
}

/** 活体行（C-03）：running + currentTool → 「正在调用 X…（已 N 秒）」；queueDepth chip。 */
export function liveRowHtml(st) {
  if (!st || st.running !== true) return '';
  const tool = String(st.currentTool || '');
  let dur = '';
  if (tool && Number(st.currentToolSince) > 0) {
    const s = Math.max(0, Math.round((Date.now() - Number(st.currentToolSince)) / 1000));
    dur = '（' + esc(tr('已 {n} 秒').replace('{n}', String(s))) + '）';
  }
  const qd = Number(st.queueDepth) || 0;
  const sd = Number(st.steerDepth) || 0;
  const queue = qd ? ' <span class="chip acc tiny">' + esc(tr('排队 {n}').replace('{n}', String(qd))) + '</span>' : '';
  const steer = sd ? ' <span class="chip llm tiny">' + esc(tr('插话 {n}').replace('{n}', String(sd))) + '</span>' : '';
  const body = tool
    ? tr('正在调用') + ' <span class="chat-live-tool">' + esc(tool) + '</span>… ' + dur
    : tr('agent 运行中…');
  return '<div class="chat-live"><span class="lamp pulse-ok"></span><span>' + body + '</span>' + queue + steer + '</div>';
}

/** 空态（C-10）：居中问候 + 4 枚示例问题 chips（data-cc-sample——点击填 composer 真实可发）。 */
export function emptyStateHtml() {
  const samples = [
    tr('进行代码审核，看看实际和文档差距有多大，并进行修复'),
    tr('找出所有 drift 的文档并生成对账清单'),
    tr('把当前循环的状态写成一份 SessionLog 摘要'),
    tr('检查 workflow 引擎的运行记录，标出失败原因'),
  ];
  return '<div class="chat-empty">' +
    '<div class="chat-empty-hi">' + esc(tr('有什么可以帮忙的？')) + '</div>' +
    '<div class="chat-empty-sub">' + esc(tr('找文档／更新文档／提交需求——会话由宿主建档，转录即真实 journal。')) + '</div>' +
    '<div class="chat-samples">' + samples.map((s) =>
      '<button class="chat-sample" data-cc-sample>' + esc(s) + '</button>').join('') + '</div>' +
    '</div>';
}

function stateChipsHtml(st) {
  if (!st) return '';
  const running = st.running === true;
  const chips = [running ? chip(tr('运行中'), 'acc') : chip(tr('空闲'), 'mute')];
  const qd = Number(st.queueDepth) || 0;
  if (qd) chips.push(chip(tr('排队 {n}').replace('{n}', String(qd)), 'llm'));
  return chips.join(' ');
}

/** 会话行（C-06：未读点 + 相对时间 + today/earlier 分组——title 注水保留）。 */
export function sessionRowHtml(s, cur, highwater) {
  const id = String(s.sessionId || '');
  const seen = Number(highwater instanceof Map ? highwater.get(id) : 0) || 0;
  const unread = !s.blank && Number(s.updatedAt) > seen && id !== cur;
  const live = s.running === true;
  const dot = live ? ' live' : (unread ? ' unread' : '');
  const title = String(s.title || '');
  return '<div class="chat-sess' + (id === cur ? ' active' : '') + '" data-sess-pick="' + esc(id) + '" role="link">' +
    '<span class="chat-dot' + dot + '"></span>' +
    '<div class="chat-sess-main">' +
    '<div class="chat-sess-id">' + esc(trunc(title || id, 26)) + '</div>' +
    (title ? '<div class="chat-sess-sub mono">' + esc(trunc(id, 26)) + '</div>' : '') +
    '</div>' +
    '<div class="chat-sess-side">' +
    (live ? chip(tr('运行中'), 'ok') : (s.blank ? chip(tr('空白'), 'mute') : '')) +
    (s.updatedAt ? '<span class="chat-sess-time">' + esc(relAge(Number(s.updatedAt)) || '') + '</span>' : '') +
    '</div></div>';
}

/** 会话清单（today/earlier 分组；组头词形随 locale）。 */
export function sessionsHtml(items, cur, seenSeq) {
  const rows = (Array.isArray(items) ? items : []).filter((s) => s && s.sessionId);
  if (!rows.length) return '<div class="chat-sess-day">' + esc(tr('会话清单为空——点「新会话」由宿主建档。')) + '</div>';
  const today = rows.filter((s) => dayGroupOf(s.updatedAt) === 'today');
  const earlier = rows.filter((s) => dayGroupOf(s.updatedAt) !== 'today');
  const out = [];
  if (today.length) {
    out.push('<div class="chat-sess-day">' + esc(tr('今天')) + '</div>');
    out.push(today.map((s) => sessionRowHtml(s, cur, seenSeq)).join(''));
  }
  if (earlier.length) {
    out.push('<div class="chat-sess-day">' + esc(tr('更早')) + '</div>');
    out.push(earlier.map((s) => sessionRowHtml(s, cur, seenSeq)).join(''));
  }
  return out.join('');
}

/* ---------- 屏骨架 ---------- */

function composerHtmlInner(can, running) {
  const mode = running ? 'steer' : 'queue';
  return '<div class="chat-comp">' +
    '<textarea class="chat-text" data-cc-text rows="1" placeholder="' +
    esc(running ? tr('插话（立即注入运行中会话——steer）') : tr('提交需求／找文档／更新文档……')) + '"></textarea>' +
    '<button class="chat-send" data-cc-send data-mode="' + mode + '"' + (can ? '' : ' disabled') +
    ' title="' + esc(running ? tr('停止（interrupt 运行中回合）') : tr('发送')) + '">' +
    '<span class="chat-send-ico">' + (running ? '■' : '↑') + '</span></button>' +
    '</div>' +
    '<div class="cc-comp-tools"><label class="btn btn-ghost sm cc-img-btn" title="' + esc(tr('附加图片（≤8MB）')) + '">🖼 ' + esc(tr('图片')) +
    '<input type="file" accept="image/png,image/jpeg,image/gif,image/webp,image/svg+xml" data-cc-img hidden></label>' +
    '<span class="cc-img-previews" data-cc-img-previews></span></div>' +
    '<div class="chat-foot">' +
    esc(running
      ? tr('Enter 插话 · 「■」停止运行中回合 · 空闲即回 queue 落投')
      : tr('Enter 发送 · Shift+Enter 换行 · 运行中变插话与停止')) +
    '</div>';
}

function screenShell(can) {
  return '<div class="chat-screen">' +
    '<div class="chat-top">' +
    '<div class="chat-top-title"><h1 class="chat-h1">' + esc(tr('会话')) + '</h1>' +
    '<div class="dh-sub">' + esc(tr('自有对话面——真实转录与发送；会话 id 由宿主发，永不手填')) + '</div></div>' +
    '<div class="chat-top-actions">' +
    '<span data-cc-state></span>' +
    '<button class="btn btn-chip sm" data-cc-new>' + esc(tr('新会话')) + '</button>' +
    '</div></div>' +
    (can ? '' : '<div class="card"><div class="card-b"><div class="small faint">' +
      esc(tr('对话通路未注入（chatfaces.js 缺席）——转录与发送面显式缺席，不冒充。')) + '</div></div></div>') +
    '<div data-sess-backlinks-slot class="chat-backlinks"></div>' +
    '<div class="chat-main">' +
    '<aside class="chat-side">' +
    '<div class="chat-side-head">' + esc(tr('会话')) +
    '<span class="chat-side-new"></span></div>' +
    '<div class="chat-sess-list" data-cc-sessions><div class="chat-sess-day">' + esc(tr('会话现读中…')) + '</div></div>' +
    '</aside>' +
    '<div class="chat-stream-wrap">' +
    '<div class="chat-stream" data-cc-tr><div class="chat-sess-day">' + esc(tr('转录现读中…')) + '</div></div>' +
    '<button class="chat-jump" data-cc-jump hidden>↓ <span data-cc-jump-n></span></button>' +
    '</div></div>' +
    '<div class="chat-dock">' +
    '<div data-cc-ctx></div>' +
    composerHtmlInner(can, false) +
    '<div class="chat-note" data-cc-note></div>' +
    '</div></div>';
}

/* ---------- 数据面 ---------- */

function streamEl(el) { return el.querySelector('[data-cc-tr]'); }

/** 贴底判别（C-04 滚动锚）：距底 <48px=贴底。 */
function nearBottom(box) {
  return box.scrollHeight - box.scrollTop - box.clientHeight < 48;
}

function followBottom(box) { box.scrollTop = box.scrollHeight; }

/** 增量渲染（C-04）：只 append seq>lastSeq 新节点 + 滚动锚（贴底才跟随）。
 * 返回 {appended, jumpVisible}——测试锚与浮标词形共用。 */
export function applyIncremental(el, nodes) {
  const box = streamEl(el);
  if (!box) return { appended: 0, jumpVisible: false };
  // 列容器（居中 800px 列——活体行/浮标同列）
  let col = box.querySelector('[data-cc-col]');
  if (!col) {
    box.innerHTML = '<div class="chat-col" data-cc-col></div>';
    col = box.querySelector('[data-cc-col]');
  }
  const wasBottom = nearBottom(box);
  const fresh = nextBatch(nodes, CHAT.painted, CHAT.lastSeq === 0);
  if (CHAT.lastSeq === 0 && !CHAT.painted.size) {
    col.innerHTML = ''; // 首绘清骨架
  }
  for (const n of fresh) {
    CHAT.painted.add(nodeKey(n));
    if (Number(n.seq) && Number(n.seq) !== Number.MAX_SAFE_INTEGER && Number(n.seq) > CHAT.lastSeq) CHAT.lastSeq = Number(n.seq);
    col.insertAdjacentHTML('beforeend', nodeHtml(n, null));
  }
  // 待达回声区（无 seq——按 requestId 键渲染在尾）
  const pendBox = box.querySelector('[data-cc-pend]');
  const pendHtml = pendingNodes().map((n) => nodeHtml(n, null)).join('');
  if (pendHtml) {
    if (!pendBox) {
      col.insertAdjacentHTML('beforeend', '<div data-cc-pend>' + pendHtml + '</div>');
    } else {
      pendBox.innerHTML = pendHtml;
    }
  } else if (pendBox) {
    pendBox.remove();
  }
  // 空态（零节点且零回声）
  const empty = box.querySelector('[data-cc-empty]');
  const isEmpty = !nodes.length && !CHAT.pending.length;
  if (isEmpty && !empty) col.insertAdjacentHTML('afterbegin', emptyStateHtml());
  else if (!isEmpty && empty) empty.remove();
  // 滚动锚：贴底跟随；上翻=浮标计数
  const jump = el.querySelector('[data-cc-jump]');
  if (wasBottom) {
    followBottom(box);
    if (jump) { jump.hidden = true; }
    return { appended: fresh.length, jumpVisible: false };
  }
  if (jump && fresh.length) {
    jump.hidden = false;
    const n = jump.querySelector('[data-cc-jump-n]');
    const prev = Number(n && n.textContent) || 0;
    if (n) n.textContent = String(prev + fresh.length);
  }
  return { appended: fresh.length, jumpVisible: !!(jump && !jump.hidden) };
}

async function refreshSessions(el, f) {
  const box = el.querySelector('[data-cc-sessions]');
  if (!box) return;
  const r = await f.sessions();
  if (r.status === 200) {
    box.innerHTML = sessionsHtml(r.data && r.data.items, CHAT.sessionId, CHAT.seenSeq);
  } else {
    box.innerHTML = '<div class="chat-sess-day">' + esc(tr('会话清单暂不可读（') + String(r.status) + tr('）——不冒充空。')) + '</div>';
  }
  box.querySelectorAll('[data-sess-pick]').forEach((row) => {
    row.addEventListener('click', () => {
      pickSession(el, f, String(row.getAttribute('data-sess-pick') || ''));
    });
  });
}

function pickSession(el, f, id) {
  if (!id || id === CHAT.sessionId) return;
  CHAT.sessionId = id;
  CHAT.lastSeq = 0;
  CHAT.painted = new Set();
  CHAT.state = null;
  paintState(el);
  refreshSessions(el, f);
  refreshTranscript(el, f);
  ensureCtxBar(el, id);
}

/** 状态面（顶条 chips + 活体行 + composer 词形随态换 C-01/C-03）。 */
function paintState(el) {
  const slot = el.querySelector('[data-cc-state]');
  if (slot) slot.innerHTML = stateChipsHtml(CHAT.state);
  const live = el.querySelector('[data-cc-live]');
  const st = CHAT.state;
  if (st && st.running === true) {
    const html = liveRowHtml(st);
    if (live) { live.innerHTML = html; }
    else {
      const box = streamEl(el);
      if (box) {
        const col = box.querySelector('[data-cc-col]');
        (col || box).insertAdjacentHTML('afterbegin', '<div data-cc-live>' + html + '</div>');
      }
    }
  } else if (live) {
    live.remove();
  }
  // composer 词形随态换（发送↔插话/停止——焦点保真：只动按钮与 placeholder）
  const btn = el.querySelector('[data-cc-send]');
  const ta = el.querySelector('[data-cc-text]');
  const running = !!(st && st.running === true);
  if (btn && running !== (btn.dataset.mode === 'steer')) {
    btn.dataset.mode = running ? 'steer' : 'queue';
    btn.title = running ? tr('停止（interrupt 运行中回合）') : tr('发送');
    const ico = btn.querySelector('.chat-send-ico');
    if (ico) ico.textContent = running ? '■' : '↑';
    btn.classList.toggle('stop', running);
  }
  if (ta) {
    const ph = running ? tr('插话（立即注入运行中会话——steer）') : tr('提交需求／找文档／更新文档……');
    if (ta.placeholder !== ph) ta.placeholder = ph;
  }
}

async function refreshState(el, f) {
  if (!CHAT.sessionId) { CHAT.state = null; paintState(el); return; }
  const r = await f.state(CHAT.sessionId);
  if (r.status !== 200 || !r.data || typeof r.data !== 'object') { return; }
  CHAT.state = r.data;
  paintState(el);
  paintSendModeNote(el);
}

/** 受理词形注记（C-01 纪律：运行中=steer/空闲=queue 两种受理词形如实区分）。 */
function paintSendModeNote(el) {
  const note = el.querySelector('[data-cc-note]');
  if (!note) return;
  const st = CHAT.state;
  const running = !!(st && st.running === true);
  note.textContent = running
    ? tr('受理词形=steer（立即注入运行中会话）')
    : tr('受理词形=queue（排队落投——agent 收尾即读）');
}

async function refreshTranscript(el, f) {
  const box = streamEl(el);
  if (!box) return;
  if (!CHAT.sessionId) {
    box.innerHTML = '<div class="chat-col" data-cc-col>' + emptyStateHtml() + '</div>';
    return;
  }
  const r = await f.transcript(CHAT.sessionId, 80);
  if (r.status !== 200) {
    box.innerHTML = '<div class="chat-col" data-cc-col><div class="chat-empty"><div class="chat-empty-sub">' +
      esc(tr('转录面暂不可读（') + String(r.status) + (r.data && r.data.error ? ' · ' + String(r.data.error) : '') + tr('）——不冒充空。')) +
      '</div></div></div>';
    CHAT.lastSeq = 0; CHAT.painted = new Set();
    return;
  }
  const events = (r.data && Array.isArray(r.data.events)) ? r.data.events : [];
  // C-05 销账：requestId 主匹配 + 同文兜底（重发同文不错销）
  if (CHAT.pending.length) {
    const byReq = new Set(events.map((e) => String((e && e.data && e.data.requestId) || '')));
    const texts = new Set(events.filter((e) => e && e.type === 'user/message').map((e) => String(((e.data && e.data.content) || []).filter((b) => b && b.type === 'text').map((b) => b.text).join('\n'))));
    CHAT.pending = CHAT.pending.filter((p) => !(p.requestId && byReq.has(p.requestId)) && !texts.has(p.text));
  }
  const nodes = mapTranscript(events);
  // 未读 highwater：清单侧已见尾 seq（当前会话实时更新）
  const lastEvSeq = events.reduce((m, e) => Math.max(m, Number(e && e.seq) || 0), 0);
  if (lastEvSeq) CHAT.seenSeq.set(CHAT.sessionId, Math.max(CHAT.seenSeq.get(CHAT.sessionId) || 0, lastEvSeq));
  applyIncremental(el, nodes);
  refreshState(el, f);
}

/* ContextPack 折叠条（cored G-14 面——默认收起）。 */
function ensureCtxBar(el, sessionId) {
  const holder = el.querySelector('[data-cc-ctx]');
  if (!holder) return;
  if (!sessionId) { holder.innerHTML = ''; return; }
  if (!holder.querySelector('[data-ctx-panel]')) {
    holder.innerHTML = '<details class="ex-block chat-ctx">' +
      '<summary class="ex-sum"><span class="ex-chev">▶</span>' + esc(tr('上下文（ContextPack 契约面）')) + '</summary>' +
      '<div class="ex-body"><div data-ctx-panel><div class="small faint">' +
      esc(tr('上下文加载中…（GET /v1/sessions/')) + esc(encodeURIComponent(sessionId)) + '/context）</div></div></div></details>';
    const ctx = (typeof window !== 'undefined' && window.__CC_CONSOLE_CTX__) || null;
    if (ctx) mountSessionCtx(el, ctx, sessionId);
  }
}

/* ---------- 发送与停止（C-01/C-02/C-05） ---------- */

async function doSend(el, f, textOverride) {
  const ta = el.querySelector('[data-cc-text]');
  const text = String(textOverride != null ? textOverride : (ta ? ta.value : '')).trim();
  if (!text) return;
  if (!CHAT.sessionId) {
    const cr = await f.create();
    if (cr.status === 200 && cr.data && cr.data.sessionId) {
      CHAT.sessionId = String(cr.data.sessionId);
      refreshSessions(el, f);
      ensureCtxBar(el, CHAT.sessionId);
    } else {
      const note = el.querySelector('[data-cc-note]');
      if (note) note.textContent = tr('新会话失败（') + String(cr.status) + tr('）');
      return;
    }
  }
  if (ta && textOverride == null) ta.value = '';
  autoGrow(ta);
  const mode = modeForState(CHAT.state); // C-01：running→steer，否则 queue
  const requestId = 'cc-' + Date.now().toString(36) + '-' + Math.random().toString(36).slice(2, 8);
  CHAT.pending.push({ text, requestId, at: Date.now() });
  applyIncremental(el, []);
  let r;
  try {
    r = await (mode === 'steer' && typeof f.send === 'function'
      ? f.send(CHAT.sessionId, text, 'steer')
      : f.send(CHAT.sessionId, text));
  } catch (e) {
    r = { status: 0, data: { error: String((e && e.message) || e) } };
  }
  if (!(r && r.status === 200)) {
    failPending(requestId, String(r && r.status) + (r && r.data && r.data.error ? ' · ' + String(r.data.error) : ''));
    applyIncremental(el, []);
  }
  setTimeout(() => refreshTranscript(el, f), 900);
  setTimeout(() => refreshTranscript(el, f), 3000);
}

/** 失败标注（C-05 行内重试）：按 requestId 销账→标记 error。 */
function failPending(requestId, errText) {
  const p = CHAT.pending.find((x) => x.requestId === requestId);
  if (p) CHAT.pending = CHAT.pending.filter((x) => x !== p);
  CHAT.failed = CHAT.failed || [];
  CHAT.failed.push({ text: p ? p.text : '', requestId, error: errText || '0' });
}

/** 重试绑定（事件委托——失败行 data-cc-retry）。 */
function bindRetry(el, f) {
  el.addEventListener('click', (ev) => {
    const btn = ev.target.closest('[data-cc-retry]');
    if (!btn) return;
    const requestId = String(btn.getAttribute('data-cc-retry') || '');
    const p = (CHAT.failed || []).find((x) => x.requestId === requestId);
    if (!p) return;
    CHAT.failed = (CHAT.failed || []).filter((x) => x !== p);
    doSend(el, f, p.text);
  });
}

/** 图片附加（2026-11-27）：选中/粘贴 → f.uploadImage → 把 ![image](url) 追加到 composer。
 * 不拦 doSend —— 图片作为 markdown 文本随原发送流程走（agent prompt 契约不变）。 */
function bindImageUpload(el, f) {
  if (!f || typeof f.uploadImage !== 'function') return;
  const input = el.querySelector('[data-cc-img]');
  const previews = el.querySelector('[data-cc-img-previews]');
  const ta = el.querySelector('[data-cc-text]');
  if (!input || !ta) return;
  const note = (msg) => { const n = el.querySelector('[data-cc-note]'); if (n) n.textContent = msg; };
  const uploadOne = async (file) => {
    if (!file || !String(file.type || '').startsWith('image/')) return;
    if (file.size > (8 << 20)) { note(tr('图片超过 8MB 上限')); return; }
    note(tr('图片上传中…'));
    try {
      const url = await f.uploadImage(file);
      const tag = '![image](' + url + ')';
      ta.value = ta.value.trimEnd() + (ta.value.trimEnd() ? '\n' : '') + tag;
      autoGrow(ta);
      note(tr('图片已附加（随消息发送）'));
      if (previews) {
        const img = document.createElement('img');
        img.src = url; img.alt = 'preview';
        img.style.cssText = 'width:44px;height:44px;object-fit:cover;border-radius:6px;margin-left:6px';
        previews.appendChild(img);
      }
    } catch (e) { note(tr('图片上传失败（') + String((e && e.message) || e) + tr('）')); }
  };
  input.addEventListener('change', () => {
    Array.from(input.files || []).forEach(uploadOne);
    input.value = '';
  });
  ta.addEventListener('paste', (ev) => {
    const items = Array.from((ev.clipboardData && ev.clipboardData.items) || []);
    const hit = items.find((it) => String(it.type || '').startsWith('image/'));
    if (!hit) return;
    const file = hit.getAsFile();
    if (file) { ev.preventDefault(); uploadOne(file); }
  });
  if (previews) previews.addEventListener('click', (ev) => {
    const img = ev.target.closest('img');
    if (!img) return;
    const src = img.getAttribute('src') || '';
    ta.value = ta.value.split('\n').filter((ln) => ln.trim() !== '![image](' + src + ')').join('\n');
    autoGrow(ta);
    img.remove();
  });
}

async function doStop(el, f) {
  const note = el.querySelector('[data-cc-note]');
  if (!CHAT.sessionId) return;
  if (typeof f.stop !== 'function') {
    if (note) note.textContent = tr('停止通路未注入（chatfaces.js.stop 缺席）——显式缺席，不冒充。');
    return;
  }
  let r;
  try {
    r = await f.stop(CHAT.sessionId);
  } catch (e) {
    r = { status: 0, data: { error: String((e && e.message) || e) } };
  }
  if (r && r.status === 200) {
    if (note) note.textContent = tr('已请求停止——会话回到空闲后转录可见中断词形。');
  } else {
    if (note) note.textContent = tr('停止失败（') + String(r && r.status) + (r && r.data && r.data.error ? ' · ' + String(r.data.error) : '') + tr('）');
  }
  setTimeout(() => refreshState(el, f), 800);
}

/** textarea 自增长（1-8 行——C-08）。 */
function autoGrow(ta) {
  if (!ta) return;
  ta.style.height = 'auto';
  const lh = 24; // 与 chat.css line-height 1.55×15px≈23.3 取整
  const max = lh * 8 + 8;
  ta.style.height = Math.min(ta.scrollHeight, max) + 'px';
}

/* ---------- 屏主体 ---------- */

/** 会话屏 v4：三段式全高对话面（左玻璃抽屉清单+右转录列+底部 composer）。 */
function dialogScreen(el, ctx) {
  const f = faces();
  const can = !!(f && typeof f.send === 'function');
  CHAT.el = el;
  el.innerHTML = screenShell(can);
  el.classList.add('route-chat');
  // V-10：回链块随会话切换重绘（快照在 WS 到账后也比初渲染丰富——订阅重绘兜底）
  const paintLinks = () => {
    const slot = el.querySelector('[data-sess-backlinks-slot]');
    if (slot) slot.innerHTML = sessBacklinksHtml(ctx.store, CHAT.sessionId);
    slot && slot.querySelectorAll('[data-go]').forEach((n) => {
      n.addEventListener('click', () => ctx.go(n.getAttribute('data-go')));
    });
  };
  paintLinks();
  const unsubLinks = ctx.store.subscribe(() => { paintLinks(); refreshSessionsSoon(el, f); });
  el.addEventListener('disconnected', () => {
    try { unsubLinks(); } catch { /* no-op */ }
    CHAT.el = null;
  }, { once: true });

  const newBtn = el.querySelector('[data-cc-new]');
  if (newBtn) newBtn.addEventListener('click', async () => {
    newBtn.disabled = true;
    const r = await f.create();
    newBtn.disabled = false;
    if (r.status === 200 && r.data && r.data.sessionId) {
      CHAT.sessionId = String(r.data.sessionId);
      CHAT.lastSeq = 0; CHAT.painted = new Set();
      refreshSessions(el, f);
      refreshTranscript(el, f);
      ensureCtxBar(el, CHAT.sessionId);
    } else {
      const note = el.querySelector('[data-cc-note]');
      if (note) note.textContent = tr('新会话失败（') + String(r.status) + tr('）');
    }
  });

  const sendBtn = el.querySelector('[data-cc-send]');
  const textTa = el.querySelector('[data-cc-text]');
  if (sendBtn && textTa) {
    sendBtn.addEventListener('click', () => {
      if (sendBtn.dataset.mode === 'steer' && CHAT.state && CHAT.state.running === true) {
        // 运行中=同位「停止」（C-02；插话走 Enter——同位双词形不互斥）
        doStop(el, f);
      } else {
        doSend(el, f);
      }
    });
    textTa.addEventListener('keydown', (ev) => {
      if (ev.key === 'Enter' && !ev.shiftKey) { ev.preventDefault(); doSend(el, f); }
    });
    textTa.addEventListener('input', () => autoGrow(textTa));
  }
  // 示例 chips（C-10：点击填 composer 真实可发——不放假对话）
  el.addEventListener('click', (ev) => {
    const s = ev.target.closest('[data-cc-sample]');
    if (!s) return;
    const ta = el.querySelector('[data-cc-text]');
    if (ta) { ta.value = s.textContent.trim(); autoGrow(ta); ta.focus(); }
  });
  // 浮标（C-04）：点击回底并清计数
  const jump = el.querySelector('[data-cc-jump]');
  if (jump) jump.addEventListener('click', () => {
    const box = streamEl(el);
    if (box) followBottom(box);
    jump.hidden = true;
    const n = jump.querySelector('[data-cc-jump-n]');
    if (n) n.textContent = '';
  });

  bindRetry(el, f);
  bindImageUpload(el, f);
  ensureCtxBar(el, CHAT.sessionId);
  refreshSessions(el, f);
  refreshTranscript(el, f);
  if (CHAT.timer) clearInterval(CHAT.timer);
  CHAT.timer = setInterval(() => {
    const root = CHAT.el && CHAT.el.isConnected ? CHAT.el : document.querySelector('.route-chat');
    if (!root || !root.isConnected) { clearInterval(CHAT.timer); CHAT.timer = 0; return; }
    const f2 = faces();
    if (!f2) return;
    refreshTranscript(root, f2);
    refreshSessionsSoon(root, f2);
  }, 3500);
}

/* 会话清单节流刷新（轮询侧 1/4 拍刷一次——行侧不每拍全量重绘）。 */
let sessTick = 0;
function refreshSessionsSoon(el, f) {
  sessTick++;
  if (sessTick % 4 !== 1) return;
  refreshSessions(el, f);
}

/** #/chat/:agent 与 #/session/:id —— 自有对话面。 */
export function render(el, ctx) {
  const id = ctx.params && ctx.params.id;
  const agent = ctx.params && ctx.params.agent;
  const paramKey = (id != null && id !== '') ? 'session/' + String(id) : 'chat/' + String(agent || '');
  const fresh = LAST.route !== String(ctx.route || '') || LAST.paramKey !== paramKey || LAST.locale !== getLocale();
  LAST.route = String(ctx.route || ''); LAST.paramKey = paramKey; LAST.locale = getLocale();
  if (fresh) {
    if (id != null && id !== '') CHAT.sessionId = String(id); // #/session/:id 直开
    else CHAT.sessionId = '';
    CHAT.pending = [];
    CHAT.failed = [];
    CHAT.lastSeq = 0;
    CHAT.painted = new Set();
    dialogScreen(el, ctx);
    if (id == null && agent) {
      // #/chat/<agent>：清单到账后选最新一条（宿主档案；blank 优先复用）——零手填
      const f = faces();
      if (f) setTimeout(async () => {
        if (CHAT.sessionId) return;
        const r = await f.sessions();
        const items = (r.status === 200 && r.data && Array.isArray(r.data.items)) ? r.data.items : [];
        const pick = items.filter((s) => s && s.sessionId)
          .sort((a, b) => (Number(b.updatedAt) || 0) - (Number(a.updatedAt) || 0))[0];
        if (pick && !CHAT.sessionId && CHAT.el === el) {
          pickSession(el, f, String(pick.sessionId));
        }
      }, 700);
    }
    return;
  }
  const f = faces();
  if (f) refreshTranscript(el, f);
}

/* ---------- V-10 回链（快照三源反查——词形不变，样式类迁 chat.css） ---------- */

export function sessBacklinksHtml(store, sessionId) {
  const sid = String(sessionId || '');
  if (!sid) return '';
  const tasks = (store.tasks() || []).filter((t) => String((t && t.status && (t.status.agentSession || t.status.sessionId)) || '') === sid);
  const evs = (store.events() || []).filter((e) => {
    const s = String((e && (e.agentSession || e.session)) || '');
    return s === sid;
  }).slice(0, 4);
  const taskChips = tasks.slice(0, 3).map((t) => {
    const id = String((t && t.scope) || 'global') + '/' + String((t && t.name) || '');
    const title = String((t && t.spec && t.spec.title) || (t && t.name) || id);
    const ph = String((t && t.status && (t.status.phase || t.status)) || '');
    return '<span class="chip acc tiny click" data-go="task/' + encodeURIComponent(id) + '" title="' + esc(id) + tr('——任务详情') + '">Task: ' +
      esc(trunc(title, 22)) + (ph ? ' · ' + esc(ph) : '') + '</span>';
  }).join(' ');
  const evChips = evs.map((e) =>
    '<span class="chip mute tiny click" data-go="event/' + encodeURIComponent(String(e.id ?? '')) + '" title="' + esc(tr('事件详情')) + '">' +
    esc(trunc(String(e.type || ''), 26)) + '</span>').join(' ');
  if (!taskChips && !evChips) {
    return '<div class="small faint" data-sess-backlinks="1">' + esc(tr('快照三源（Task.agentSession ∪ 事件 agentSession）反查无在册关联——会话型独立会话（零编造）。')) + '</div>';
  }
  return '<div class="small" data-sess-backlinks="1">' +
    '<span class="faint">' + esc(tr('关联：')) + '</span> ' + taskChips + ' ' + evChips + '</div>';
}

/* ---------- composer 导出（总览大屏在用——同版型词形） ---------- */

function sessionOptions() {
  /* 正常服务没有手填 session id（用户裁决）：总览 composer 恒发 root 会话。
   * 会话选择在会话屏（宿主发 id）。 */
  return '<option value="root" selected>' + esc(tr('root 会话（默认）')) + '</option>';
}

/** 对话 composer（总览大屏在用）。 */
export function composerHtml(ctx, curSession) {
  const f = faces();
  const canSend = !!(f && typeof f.send === 'function');
  void ctx; void curSession;
  return '<div class="cc-composer">' +
    '<div class="cc-comp-row">' +
    '<select class="inp cc-sess" data-cc-sess>' + sessionOptions() + '</select>' +
    '</div>' +
    '<div class="cc-comp-row">' +
    '<textarea class="inp cc-text" data-cc-text rows="2" placeholder="' +
    esc(tr('提交需求／找文档／更新文档——例如：进行代码审核，看看实际和文档差距有多大，并进行修复')) + '"></textarea>' +
    '<button class="btn btn-primary" data-cc-send' + (canSend ? '' : ' disabled') + '>' + esc(tr('发送')) + '</button>' +
    '</div>' +
    '<div class="small faint">' +
    (canSend
      ? esc(tr('Enter 发送 · 消息送达系统会话；处理结论以文档与流程状态呈现'))
      : esc(tr('对话通路未注入（chatfaces.js 缺席）——送达面显式缺席，不冒充已发送'))) +
    '</div></div>';
}

/** 绑定 composer 行为（总览大屏——root 会话 queue 落投；回执 requestId 销账同 C-05 词形）。 */
export function bindComposer(el, ctx, rootEl) {
  const sendBtn = rootEl.querySelector('[data-cc-send]');
  const textTa = rootEl.querySelector('[data-cc-text]');
  if (!sendBtn || !textTa) return;
  const doSend = async () => {
    const text = textTa.value.trim();
    if (!text) return;
    const f = faces();
    if (!f || typeof f.send !== 'function') return;
    textTa.value = '';
    const r = await f.send('root', text);
    const note = document.createElement('div');
    note.className = 'small faint';
    note.textContent = r && r.status === 200
      ? tr('已送达（queue）——root agent 会话已受理；结论经状态徽与节点图可见。')
      : tr('送达失败（') + String(r && r.status) + tr('）');
    sendBtn.parentElement.appendChild(note);
    setTimeout(() => note.remove(), 6000);
  };
  sendBtn.addEventListener('click', doSend);
  textTa.addEventListener('keydown', (ev) => {
    if (ev.key === 'Enter' && !ev.shiftKey) { ev.preventDefault(); doSend(); }
  });
}
