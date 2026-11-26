/* chat.js —— 会话屏 v3（2026-11-26 用户裁决）：自有 DeepSeek 式对话界面——「代码拷贝过来」
 * 不是 iframe 引 dsh 壳：控制台画自己的对话面（只有对话——无宿主侧栏/壳导航）。
 * 数据面=window.__CC.chatFaces（plugin 转录桥）：sessions（会话清单）/create（宿主发 id——
 * 零手填）/transcript（durable journal 词形投影）/state（running/queue）/send（queue 落投）。
 * 转录映射（dsh-client-ui-chat conversation-nodes 同族词形）：
 *   user/message(source.user)→用户泡；assistant/message→AI 泡（text=md 渲染；reasoning=思考
 *   expand；tool-call 块=工具 expand）；tool/call+tool/result 按 callId 归组=工具过程 expand；
 *   system/developer/message=灰行；delta/live-chunk 已由桥剥除。
 * 交互纪律：轮询 3.5s 只重绘转录区（composer 焦点保真）；会话点选=就地换数据源；
 * ContextPack=右列上折叠条（cored G-14 面）。 */

import { esc, trunc, relAge } from '../core/format.js';
import { chip, screenHead } from '../render/chips.js';
import { expandBlock, linkifyRefs } from '../render/refs.js';
import { renderMarkdown } from '../render/md.js';
import { mountSessionCtx } from './sessionctx.js';
import { t as tr, getLocale } from '../core/i18n.js';

function faces() {
  return (typeof window !== 'undefined' && window.__CC && window.__CC.chatFaces) || null;
}

/* 会话屏内存态：当前会话 + 轮询定时器（导航离开即清）+ 待达回声 + 标题缓存。 */
const CHAT = { sessionId: '', timer: 0, pending: [], titles: new Map() };
const LAST = { route: '', paramKey: '', locale: '' };

/* ---------- 转录映射（journal 事件 → 渲染节点） ---------- */

/** content 块 → {text, reasoning, toolCalls[]}（user/assistant 两族共用块词形）。 */
function blocksOf(content) {
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
    if (type === 'user/message') {
      const src = (data.source && typeof data.source === 'object') ? String(data.source.kind || '') : '';
      if (src === 'compact-checkpoint') continue; // 压缩检查点=替换面事件，不渲染
      const b = blocksOf(data.content);
      nodes.push({ kind: 'user', seq, time: ev.time, text: b.text, injected: src !== 'user' });
    } else if (type === 'assistant/message') {
      // 词形（CT116 活体实证）：{turn, step, message:{role, content:[blocks]}}——块在 message.content
      const b = blocksOf(data.message && Array.isArray(data.message.content) ? data.message.content : data.content);
      for (const tc of b.toolCalls) order.push(tc.id);
      nodes.push({ kind: 'assistant', seq, time: ev.time, text: b.text, reasoning: b.reasoning });
    } else if (type === 'tool/call') {
      const id = String(data.callId || data.id || '');
      if (!tools.has(id)) tools.set(id, { seq, name: String(data.name || ''), args: typeof data.arguments === 'string' ? data.arguments : JSON.stringify(data.arguments ?? data.args ?? ''), result: '' });
      order.push(id);
    } else if (type === 'tool/result') {
      const id = String(data.callId || data.id || '');
      const t = tools.get(id);
      if (t) t.result = typeof data.content === 'string' ? data.content : (data.output != null ? String(typeof data.output === 'string' ? data.output : JSON.stringify(data.output)) : '');
    } else if (type === 'system/message' || type === 'developer/message') {
      nodes.push({ kind: 'system', seq, time: ev.time, text: String(data.text ?? data.content ?? ''), dev: type === 'developer/message' });
    }
  }
  for (const id of order) {
    const t = tools.get(id);
    if (!t) continue;
    nodes.push({ kind: 'tool', seq: t.seq, name: t.name || tr('工具'), args: t.args, result: t.result });
  }
  nodes.sort((a, b) => a.seq - b.seq);
  return nodes;
}

/* ---------- 渲染 ---------- */

function nodeHtml(n, names) {
  if (n.kind === 'user') {
    // 注入行（source≠user——system-reminder/技能目录等）默认折叠成细条（可展开全文——
    // 全文泡会淹没对话；真用户消息照常气泡）。
    if (n.injected && !n.pending) {
      return '<div class="cc-msg tool">' +
        expandBlock('<div class="mono small" style="white-space:pre-wrap">' + esc(trunc(String(n.text || ''), 1200)) + '</div>',
          tr('注入上下文（') + trunc(String(n.text || ''), 60) + tr('…）')) + '</div>';
    }
    return '<div class="cc-msg user">' +
      '<div class="cc-bubble">' + (n.text ? linkifyRefs(n.text, names) : '<span class="faint">…</span>') + '</div>' +
      (n.pending ? '<div class="cc-meta faint">' + esc(tr('送达中…')) + '</div>' : '') +
      '</div>';
  }
  if (n.kind === 'assistant') {
    return '<div class="cc-msg ai">' +
      '<div class="cc-avatar">◆</div>' +
      '<div class="cc-body">' +
      (n.reasoning ? expandBlock('<div class="msg-text faint">' + linkifyRefs(n.reasoning, names) + '</div>', tr('思考过程')) : '') +
      (n.text ? '<div class="cc-md md-body">' + renderMarkdown(n.text) + '</div>' : '<div class="cc-md faint">' + esc(tr('（空回复）')) + '</div>') +
      '</div></div>';
  }
  if (n.kind === 'tool') {
    const detail = '<div class="mono small" style="white-space:pre-wrap">' +
      esc(trunc(String(n.args || ''), 800)) +
      (n.result ? '\n────\n' + esc(trunc(String(n.result), 800)) : '') + '</div>';
    return '<div class="cc-msg tool">' + expandBlock(detail, tr('工具：') + n.name) + '</div>';
  }
  return '<div class="cc-meta faint">' + (n.dev ? esc(tr('开发者')) + ' · ' : '') + esc(trunc(String(n.text || ''), 200)) + '</div>';
}

function transcriptHtml(nodes, names) {
  if (!nodes.length) {
    return '<div class="cc-empty">' + esc(tr('转录为空——下方发第一条消息；结论由 agent 落文档后经状态徽与节点图可见。')) + '</div>';
  }
  return nodes.map((n) => nodeHtml(n, names)).join('');
}

function sessionRowHtml(s, cur, display, pick) {
  const id = String(pick != null ? pick : s.sessionId || '');
  const label = String(display || id);
  const sel = id === cur ? ' style="outline:1px solid var(--accent);background:var(--panel-2);border-radius:8px"' : '';
  const flag = s.blank ? chip(tr('空白'), 'mute') : (s.running ? chip(tr('运行中'), 'acc') : '');
  return '<div class="dash-row" data-sess-pick="' + esc(id) + '" role="link"' + sel + '>' +
    '<div class="dr-main"><b class="mono">' + esc(trunc(label, 24)) + '</b></div>' +
    '<div class="dr-side">' + flag + (s.updatedAt ? '<span class="small faint">' + esc(relAge(new Date(Number(s.updatedAt)).toISOString()) || '') + '</span>' : '') + '</div></div>';
}

function sessionsHtml(items, cur) {
  const rows = (Array.isArray(items) ? items : []).filter((s) => s && s.sessionId)
    .sort((a, b) => (Number(b.updatedAt) || 0) - (Number(a.updatedAt) || 0));
  if (!rows.length) return '<div class="small faint" style="padding:4px 6px">' + esc(tr('会话清单为空——点「新会话」由宿主建档。')) + '</div>';
  const head = rows.slice(0, 40).map((s) => {
    const title = CHAT.titles.get(String(s.sessionId));
    return sessionRowHtml(s, cur, title, String(s.sessionId));
  }).join('');
  return head + (rows.length > 40 ? '<div class="small faint" style="padding:4px 6px">' + esc(tr('…共 ') + rows.length + tr(' 会话（最近 40 条）')) + '</div>' : '');
}

/* 标题注水（惰性——最近 10 条：转录首页首条 user 消息截断为题；缓存免重读）。 */
async function hydrateTitles(el, f, items) {
  const rows = (Array.isArray(items) ? items : []).filter((s) => s && s.sessionId)
    .sort((a, b) => (Number(b.updatedAt) || 0) - (Number(a.updatedAt) || 0)).slice(0, 10);
  let dirty = false;
  await Promise.all(rows.map(async (s) => {
    const id = String(s.sessionId);
    if (CHAT.titles.has(id) || s.blank) return;
    try {
      const r = await f.transcript(id, 12);
      if (r.status !== 200 || !Array.isArray(r.data && r.data.events)) return;
      const firstUser = r.data.events.filter((e) => e && e.type === 'user/message')
        .map((e) => String((((e.data && e.data.content) || []).filter((b) => b && b.type === 'text').map((b) => b.text).join(' ')) || '')).filter(Boolean)[0];
      if (firstUser) { CHAT.titles.set(id, trunc(firstUser, 24)); dirty = true; }
    } catch { /* 标题缺席=行显 id（诚实降级） */ }
  }));
  if (dirty) refreshSessions(el, f);
}

/* ---------- 数据面 ---------- */

async function refreshSessions(el, f) {
  const box = el.querySelector('[data-cc-sessions]');
  if (!box) return;
  const r = await f.sessions();
  if (r.status === 200) {
    box.innerHTML = sessionsHtml(r.data && r.data.items, CHAT.sessionId);
    hydrateTitles(el, f, r.data && r.data.items);
  } else {
    box.innerHTML = '<div class="small faint" style="padding:4px 6px">' + esc(tr('会话清单暂不可读（') + String(r.status) + tr('）——不冒充空。')) + '</div>';
  }
  box.querySelectorAll('[data-sess-pick]').forEach((row) => {
    row.addEventListener('click', () => {
      CHAT.sessionId = String(row.getAttribute('data-sess-pick') || '');
      refreshSessions(el, f);
      refreshTranscript(el, f, true);
      ensureCtxBar(el, CHAT.sessionId);
      paintBacklinks(el);
    });
  });
}

/* V-10：会话屏回链 task/event（按 sessionId 反查快照 Task.status.agentSession ∪ 事件 agentSession/session——零编造）。 */
export function sessBacklinksHtml(store, sessionId) {
  const sid = String(sessionId || '').trim();
  if (!sid || !store) return '';
  const tasks = (store.tasks ? store.tasks() : []).filter((t) => {
    const s = (t && t.status) || {};
    return String(s.agentSession || s.sessionId || '') === sid;
  });
  const evs = (store.events ? store.events() : []).filter((e) => {
    const s = String((e && (e.agentSession || e.session)) || '').replace(/^"+|"+$/g, '');
    return s === sid;
  }).slice(0, 4);
  const taskChips = tasks.slice(0, 3).map((t) => {
    const id = String((t && t.scope) || 'global') + '/' + String((t && t.name) || '');
    const title = String((t && t.spec && t.spec.title) || (t && t.name) || id);
    const ph = String((t && t.status && t.status.phase) || '');
    return '<span class="chip acc tiny click" data-go="task/' + encodeURIComponent(id) + '" title="' + esc(id) + '">' +
      'Task: ' + esc(trunc(title, 24)) + (ph ? ' · ' + esc(ph) : '') + '</span>';
  }).join(' ');
  const evChips = evs.map((e) => {
    const eid = e && (e.id != null ? e.id : (e.eventId != null ? e.eventId : ''));
    return eid !== ''
      ? '<span class="chip mute tiny click" data-go="event/' + encodeURIComponent(String(eid)) + '" title="' + esc(tr('事件详情')) + '">' + esc(trunc(String(e.type || ''), 24)) + '</span>'
      : chip(trunc(String(e.type || ''), 24), 'mute');
  }).join(' ');
  if (!taskChips && !evChips) {
    return '<div class="small faint" data-sess-backlinks="1">' + esc(tr('关联任务/事件：快照反查无在册关联（独立会话——零编造）。')) + '</div>';
  }
  return '<div class="small" data-sess-backlinks="1" style="display:flex;gap:6px;flex-wrap:wrap;align-items:center">' +
    '<span class="faint">' + esc(tr('关联：')) + '</span>' + taskChips + evChips + '</div>';
}

function paintBacklinks(el) {
  const ctx = (typeof window !== 'undefined' && window.__CC_CONSOLE_CTX__) || null;
  const slot = el && el.querySelector ? el.querySelector('[data-sess-backlinks-slot]') : null;
  if (slot && ctx) slot.innerHTML = sessBacklinksHtml(ctx.store, CHAT.sessionId);
}

async function refreshState(el, f) {
  const slot = el.querySelector('[data-cc-state]');
  if (!slot || !CHAT.sessionId) { if (slot) slot.innerHTML = ''; return; }
  const r = await f.state(CHAT.sessionId);
  if (r.status !== 200 || !r.data || typeof r.data !== 'object') { slot.innerHTML = ''; return; }
  const d = r.data;
  slot.innerHTML = d.running ? chip(tr('运行中'), 'acc') : chip(tr('空闲'), 'mute');
}

async function refreshTranscript(el, f) {
  const box = el.querySelector('[data-cc-tr]');
  if (!box) return;
  if (!CHAT.sessionId) {
    box.innerHTML = '<div class="cc-empty">' + esc(tr('左侧选择会话，或点「新会话」由宿主建档——会话 id 由宿主发，永不手填。')) + '</div>';
    return;
  }
  const r = await f.transcript(CHAT.sessionId, 80);
  if (r.status !== 200) {
    box.innerHTML = '<div class="cc-empty">' + esc(tr('转录面暂不可读（') + String(r.status) + (r.data && r.data.error ? ' · ' + String(r.data.error) : '') + tr('）——不冒充空。')) + '</div>';
    return;
  }
  const events = (r.data && Array.isArray(r.data.events)) ? r.data.events : [];
  // 待达回声去重：同文 user/message 入账即摘
  if (CHAT.pending.length) {
    const seen = new Set(events.filter((e) => e && e.type === 'user/message').map((e) => String(((e.data && e.data.content) || []).filter((b) => b && b.type === 'text').map((b) => b.text).join('\n'))));
    CHAT.pending = CHAT.pending.filter((p) => !seen.has(p.text));
  }
  const nodes = mapTranscript(events);
  for (const p of CHAT.pending) nodes.push({ kind: 'user', seq: Number.MAX_SAFE_INTEGER, text: p.text, pending: true });
  box.innerHTML = transcriptHtml(nodes, null);
  box.scrollTop = box.scrollHeight;
  refreshState(el, f);
}

/* ContextPack 折叠条（cored G-14 面——默认收起）。 */
function ensureCtxBar(el, sessionId) {
  const holder = el.querySelector('[data-cc-ctx]');
  if (!holder) return;
  if (!sessionId) { holder.innerHTML = ''; return; }
  if (!holder.querySelector('[data-ctx-panel]')) {
    holder.innerHTML = '<details class="ex-block" style="margin:0 0 10px;background:var(--panel-2)">' +
      '<summary class="ex-sum"><span class="ex-chev">▶</span>' + esc(tr('上下文（ContextPack 契约面）')) + '</summary>' +
      '<div class="ex-body"><div data-ctx-panel><div class="card"><div class="card-b"><div class="small faint">' +
      esc(tr('上下文注水中…（GET /v1/sessions/')) + esc(encodeURIComponent(sessionId)) + '/context）</div></div></div></div></div></details>';
    const ctx = (typeof window !== 'undefined' && window.__CC_CONSOLE_CTX__) || null;
    if (ctx) mountSessionCtx(el, ctx, sessionId);
  }
}

/* ---------- 屏主体 ---------- */

/** 会话屏 v3：自有 DeepSeek 式对话面（左清单+右转录+底部 composer）。 */
function dialogScreen(el, ctx) {
  const f = faces();
  const can = !!(f && typeof f.sessions === 'function');
  el.innerHTML = '<div class="view-inner">' +
    screenHead(tr('会话'), tr('自有对话面——真实转录与发送；会话 id 由宿主发，永不手填'),
      '<button class="btn btn-chip sm" data-cc-new>' + esc(tr('新会话')) + '</button>') +
    (can ? '' : '<div class="card"><div class="card-b"><div class="small faint">' +
      esc(tr('对话通路未注入（chatfaces.js 缺席）——转录与发送面显式缺席，不冒充。')) + '</div></div></div>') +
    '<div data-sess-backlinks-slot style="margin-bottom:8px">' + sessBacklinksHtml(ctx.store, CHAT.sessionId) + '</div>' +
    '<div style="display:grid;grid-template-columns:280px minmax(0,1fr);gap:14px;align-items:start">' +
    '<aside class="card" style="padding:8px;max-height:calc(100vh - 160px);overflow:auto">' +
    '<div class="small faint" style="padding:2px 6px 8px;font-weight:700">' + esc(tr('会话')) + '</div>' +
    '<div data-cc-sessions><div class="small faint" style="padding:4px 6px">' + esc(tr('会话现读中…')) + '</div></div>' +
    '</aside>' +
    '<section style="min-width:0;display:flex;flex-direction:column;gap:10px">' +
    '<div data-cc-ctx></div>' +
    '<div data-cc-state style="min-height:22px"></div>' +
    '<div class="card cc-tr" data-cc-tr style="max-height:calc(100vh - 260px);overflow:auto;padding:14px"><div class="small faint">' + esc(tr('转录现读中…')) + '</div></div>' +
    '<div class="cc-composer">' +
    '<div class="cc-comp-row">' +
    '<textarea class="inp cc-text" data-cc-text rows="2" placeholder="' +
    esc(tr('提交需求／找文档／更新文档——例如：进行代码审核，看看实际和文档差距有多大，并进行修复')) + '"></textarea>' +
    sendBtnHtml(can) +
    '</div>' +
    '<div class="small faint">' +
    esc(tr('Enter 发送 · queue 语义——送达宿主 agent 会话（durable journal；转录区随后可见双方消息）')) +
    '</div></div>' +
    '</section></div></div>';

  if (!can) return;
  window.__CC_CONSOLE_CTX__ = ctx;
  const newBtn = el.querySelector('[data-cc-new]');
  if (newBtn) newBtn.addEventListener('click', async () => {
    newBtn.disabled = true;
    const r = await f.create();
    newBtn.disabled = false;
    if (r.status === 200 && r.data && r.data.sessionId) {
      CHAT.sessionId = String(r.data.sessionId);
      await refreshSessions(el, f);
      refreshTranscript(el, f);
      ensureCtxBar(el, CHAT.sessionId);
      paintBacklinks(el);
    } else {
      const slot = el.querySelector('[data-cc-state]');
      if (slot) slot.innerHTML = chip(tr('新会话失败（') + String(r.status) + tr('）'), 'esc');
    }
  });
  const sendBtn = el.querySelector('[data-cc-send]');
  const textTa = el.querySelector('[data-cc-text]');
  if (sendBtn && textTa) {
    const doSend = async () => {
      const text = textTa.value.trim();
      if (!text) return;
      if (!CHAT.sessionId) {
        const cr = await f.create();
        if (cr.status === 200 && cr.data && cr.data.sessionId) CHAT.sessionId = String(cr.data.sessionId);
        else return;
      }
      textTa.value = '';
      CHAT.pending.push({ text });
      refreshTranscript(el, f);
      const r = await f.send(CHAT.sessionId, text);
      if (!(r && r.status === 200)) {
        CHAT.pending = CHAT.pending.filter((p) => p.text !== text);
        const slot = el.querySelector('[data-cc-state]');
        if (slot) slot.innerHTML = chip(tr('送达失败（') + String(r && r.status) + tr('）'), 'esc');
      }
      setTimeout(() => refreshTranscript(el, f), 900);
      setTimeout(() => refreshTranscript(el, f), 3000);
    };
    sendBtn.addEventListener('click', doSend);
    textTa.addEventListener('keydown', (ev) => {
      if (ev.key === 'Enter' && !ev.shiftKey) { ev.preventDefault(); doSend(); }
    });
  }
  ensureCtxBar(el, CHAT.sessionId);
  refreshSessions(el, f);
  refreshTranscript(el, f);
  if (CHAT.timer) clearInterval(CHAT.timer);
  CHAT.timer = setInterval(() => {
    const trBox = document.querySelector('[data-cc-tr]');
    if (!trBox || !trBox.isConnected) { clearInterval(CHAT.timer); CHAT.timer = 0; return; }
    const f2 = faces();
    if (f2) refreshTranscript(document, f2);
  }, 3500);
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
        if (pick && !CHAT.sessionId) {
          CHAT.sessionId = String(pick.sessionId);
          refreshSessions(el, f);
          refreshTranscript(el, f);
          ensureCtxBar(el, CHAT.sessionId);
          paintBacklinks(el);
        }
      }, 700);
    }
    return;
  }
  paintBacklinks(el);
  const f = faces();
  if (f) refreshTranscript(el, f);
}

/* ---------- composer 导出（总览大屏在用——同款无手填词形） ---------- */

function sessionOptions() {
  /* 正常服务没有手填 session id（用户裁决）：总览 composer 恒发 root 会话。
   * 会话选择在会话屏（宿主发 id）。 */
  return '<option value="root" selected>' + esc(tr('root 会话（默认）')) + '</option>';
}

/* 发送主按钮（每屏唯一主操作——会话屏与总览大屏分别渲染，不共存）。 */
function sendBtnHtml(enabled) {
  return '<button class="btn btn-primary" data-cc-send' + (enabled ? '' : ' disabled') + '>' + esc(tr('发送')) + '</button>';
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
    sendBtnHtml(canSend) +
    '</div>' +
    '<div class="small faint">' +
    (canSend
      ? esc(tr('Enter 发送 · queue 语义（送达 root agent 会话；结论由 agent 落文档后经状态徽与节点图可见）'))
      : esc(tr('对话通路未注入（chatfaces.js 缺席）——送达面显式缺席，不冒充已发送'))) +
    '</div></div>';
}

/** 绑定 composer 行为（总览大屏——root 会话 queue 落投）。 */
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
