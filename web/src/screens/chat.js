/* chat.js —— 会话屏 v5（2026-11-27 D 批）：DeepSeek harness 展示类型对账（chat-ds.md）。
 * v4（C 批）通路底座全保留：chatFaces 桥/增量渲染/贴底跟随/requestId 销账/steer|queue/停止。
 * v5 新词族（映射纯函数拆 chat-nodes.js——上游 locale.ts zh 段+conversation-nodes 逐条对账）：
 *   notice（D1 轮触发行卡：后台任务状态更新/自动化任务/子任务状态更新…+D2 上下文注入行：
 *   form=notice 摘要一行）/turntail（D3 已完成，用时 X/处理失败/已停止/max-tokens+hint）/
 *   todo（D4 更新任务清单 {done}/{total}）/retry（D8）/tool 五态+专名标题+关键参数摘要+
 *   meta 计数（D5/D6/D7）+usage pill（D10）+时钟（D13）。
 *   病根修复：tool/result 结果词形=data.message.content 块（旧读 data.content——「执行中…」卡死）。
 * B 批（用户点名）：chat-refs.js——引用条（图2 词形 chip+打开）+详情弹窗；md.js 代码块头横幅
 * （语言/复制/换行，D11）；思考行对齐「思考」+首行预览（D12）。 */

import { esc, trunc, relAge } from '../core/format.js';
import { chip } from '../render/chips.js';
import { sessBacklinksHtml } from './chat-links.js';
export { sessBacklinksHtml } from './chat-links.js';
import { mountSessionCtx } from './sessionctx.js';
import { mountSessionStages, repaintSessionStages } from './session-stages.js';
import { openProviderDrawer } from './workflow-config.js';
import { t as tr, getLocale } from '../core/i18n.js';
import {
  mapTranscript, blocksOf, nodeKey, turnTailOf, todoModel, usageTotal, usageDetail, formatClock,
} from './chat-nodes.js';
import { nodeHtml } from './chat-views.js';
import { extractRefs, chunkRefs, mountRefs, classifyMentions } from './chat-refs.js';
export { blocksOf, mapTranscript, nodeKey };

function faces() {
  return (typeof window !== 'undefined' && window.__CC && window.__CC.chatFaces) || null;
}

/* ---------- 纯函数（chat-nodes.js v5 权威——顶部 import+export 已兼容旧测试锚） ---------- */

/** 增量分派（C-04）：nodeKey(seq+kind) 已渲染的不重出；返回真正的新节点尾片。
 * 全量分片词形（firstPaint=true 或 seq 回退）=整列重画（路由切换/重连兜底）。
 * nodeKey 权威在 chat-nodes.js（chat-views 依赖）——此处 re-export 兼容旧测试锚。 */
export function nextBatch(nodes, paintedKeys, firstPaint) {
  const rows = Array.isArray(nodes) ? nodes : [];
  if (firstPaint) return rows;
  const out = rows.filter((n) => !paintedKeys.has(nodeKey(n)));
  return out;
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
  lastNodes: [],       // 最近映射节点（B1 引用抽取源）
  refs: {},            // 引用条模型 {docs, files, chunks}（B2）
  refsChunks: [],      // ContextPack chunks（B4 同源，会话切换时懒拉）
  paintRefs: null,     // mountRefs 返回的条重绘 fn（B3 挂载后可用）
};

const LAST = { route: '', paramKey: '', locale: '' };

/* ---------- 渲染（nodeHtml v5 权威=chat-views.js——import 已在文件头） ---------- */

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
  const steer = sd ? ' <span class="chip llm tiny">' + esc(tr('追加 {n}').replace('{n}', String(sd))) + '</span>' : '';
  const body = tool
    ? tr('正在调用') + ' <span class="chat-live-tool">' + esc(tool) + '</span>… ' + dur
    : tr('深度求索中…');
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
    esc(tr('提交需求／找文档／更新文档……')) + '"></textarea>' +
    '<button class="chat-send" data-cc-send data-mode="' + mode + '"' + (can ? '' : ' disabled') +
    ' title="' + esc(running ? tr('停止') : tr('发送')) + '">' +
    '<span class="chat-send-ico">' + (running ? '■' : '↑') + '</span></button>' +
    '</div>' +
    '<div class="cc-comp-tools"><label class="btn btn-ghost sm cc-img-btn" title="' + esc(tr('附加图片（≤8MB）')) + '">🖼 ' + esc(tr('图片')) +
    '<input type="file" accept="image/png,image/jpeg,image/gif,image/webp,image/svg+xml" data-cc-img hidden></label>' +
    '<span class="cc-img-previews" data-cc-img-previews></span></div>' +
    '<div class="chat-foot">' +
    esc(tr('Enter 发送 · Shift+Enter 换行')) +
    '</div>';
}

function screenShell(can) {
  return '<div class="chat-screen">' +
    '<div class="chat-top">' +
    '<div class="chat-top-title"><h1 class="chat-h1">' + esc(tr('会话')) + '</h1>' +
    '<div class="dh-sub">' + esc(tr('自有对话面——真实转录与发送；会话 id 由宿主发，永不手填')) + '</div></div>' +
    '<div class="chat-top-actions">' +
    '<button class="btn btn-chip sm" data-cc-provider title="' + esc(tr('配置 OpenAIProvider（api-key 引用 / base 地址 / 模型分级）')) + '">Provider</button>' +
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
    '</div>' +
    '<aside class="chat-right" data-cc-right>' +
    '<div data-stages-panel><div class="card"><div class="card-b"><div class="small faint">' + esc(tr('三段式 Session 现读中…')) + '</div></div></div></div>' +
    '<div data-ctx-panel><div class="card"><div class="card-b"><div class="small faint">' + esc(tr('上下文现读中…（选定会话后装配）')) + '</div></div></div></div>' +
    '<div data-provider-drawer></div>' +
    '</aside></div>' +
    '<div class="chat-dock">' +
    '<div class="chat-refs-slot" data-cc-refs></div>' +
    composerHtmlInner(can, false) +
    '<div class="chat-note" data-cc-note></div>' +
    '</div></div>';
}

/* ---------- 数据面 ---------- */

function streamEl(el) { return el.querySelector('[data-cc-tr]'); }

/** 引用条重绘（B2）：模型=最近节点抽取 + ContextPack chunks 同源（B4）。 */
function paintRefs(el) {
  CHAT.refs = { ...extractRefs(CHAT.lastNodes), chunks: CHAT.refsChunks };
  if (CHAT.paintRefs) CHAT.paintRefs(CHAT.refs);
}

/** ContextPack chunks 懒拉（B4——会话切换时一次；面缺席=空组不冒充）。 */
async function refreshCtxChunks(el) {
  const cc = (typeof window !== 'undefined' && window.__CC_CONSOLE_CTX__) || null;
  if (!cc || !cc.api || typeof cc.api.sessionContext !== 'function' || !CHAT.sessionId) {
    CHAT.refsChunks = [];
    paintRefs(el);
    return;
  }
  try {
    const r = await cc.api.sessionContext(CHAT.sessionId);
    const d = (r && (r.ok || r.status === 200) && r.data && typeof r.data === 'object') ? r.data : {};
    const pack = (d.pack && typeof d.pack === 'object') ? d.pack : {};
    CHAT.refsChunks = (Array.isArray(pack.chunks) ? pack.chunks : [])
      .map((c) => ({ kind: String((c && c.kind) || ''), scope: String((c && c.scope) || 'global'), name: String((c && c.name) || ''), score: Number((c && c.score) || 0) }))
      .filter((c) => c.name);
  } catch {
    CHAT.refsChunks = []; // 面缺席=诚实空组
  }
  paintRefs(el);
}

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
  classifyMentions(col); // B3 联动：内联 code 路径词形 → 可点提及 chip
  paintRefs(el); // B2 引用条随转录刷新
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
  CHAT.lastNodes = [];
  CHAT.refsChunks = [];
  CHAT.state = null;
  paintState(el);
  refreshSessions(el, f);
  refreshTranscript(el, f);
  ensureCtxBar(el, id);
  refreshCtxChunks(el); // B4 chunks 同源懒拉
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
  // composer 按钮随态换（发送↔停止——焦点保真：只动按钮；placeholder 恒一，注入机制静默）
  const btn = el.querySelector('[data-cc-send]');
  const running = !!(st && st.running === true);
  if (btn && running !== (btn.dataset.mode === 'steer')) {
    btn.dataset.mode = running ? 'steer' : 'queue';
    btn.title = running ? tr('停止') : tr('发送');
    const ico = btn.querySelector('.chat-send-ico');
    if (ico) ico.textContent = running ? '■' : '↑';
    btn.classList.toggle('stop', running);
  }
}

async function refreshState(el, f) {
  if (!CHAT.sessionId) { CHAT.state = null; paintState(el); return; }
  const r = await f.state(CHAT.sessionId);
  if (r.status !== 200 || !r.data || typeof r.data !== 'object') { return; }
  CHAT.state = r.data;
  try {
    window.__CC_CHAT_STATE__ = CHAT.state; // session-stages 三段进度消费面
    repaintSessionStages(el);
  } catch { /* 右栏缺席不阻塞主链路 */ }
  paintState(el);
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
  CHAT.lastNodes = nodes; // B1 引用抽取源随转录更新
  // 未读 highwater：清单侧已见尾 seq（当前会话实时更新）
  const lastEvSeq = events.reduce((m, e) => Math.max(m, Number(e && e.seq) || 0), 0);
  if (lastEvSeq) CHAT.seenSeq.set(CHAT.sessionId, Math.max(CHAT.seenSeq.get(CHAT.sessionId) || 0, lastEvSeq));
  applyIncremental(el, nodes);
  refreshState(el, f);
}

/* 右栏装配（2026-11-27 用户裁决：右边可以看到完整缓存后的上下文——sessionctx 面板 +
 * 三段卡 + Provider 抽屉常驻右栏，不再折叠在 composer 上方）。 */
function ensureCtxBar(el, sessionId) {
  const rail = el.querySelector('[data-cc-right]');
  if (!rail) return;
  const ctx = (typeof window !== 'undefined' && window.__CC_CONSOLE_CTX__) || null;
  if (sessionId) {
    mountSessionCtx(el, ctx, sessionId); // [data-ctx-panel] 现于右栏（sessionctx.js 同一挂载契约）
  }
  mountSessionStages(el);
  const provBtn = el.querySelector('[data-cc-provider]');
  if (provBtn && !provBtn.dataset.wired) {
    provBtn.dataset.wired = '1';
    provBtn.addEventListener('click', () => {
      const c = (typeof window !== 'undefined' && window.__CC_CONSOLE_CTX__) || null;
      if (c) openProviderDrawer(el, c, 'openai-default');
    });
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
      refreshCtxChunks(el);
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
      CHAT.lastNodes = []; CHAT.refsChunks = [];
      refreshSessions(el, f);
      refreshTranscript(el, f);
      ensureCtxBar(el, CHAT.sessionId);
      refreshCtxChunks(el);
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
  // B 批：引用条+详情弹窗挂载（ctxc=console 全局 ctx 的 api 面；go 优先路由 ctx）
  const cc = (typeof window !== 'undefined' && window.__CC_CONSOLE_CTX__) || null;
  if (cc && cc.api) {
    CHAT.paintRefs = mountRefs(el, { api: cc.api, go: (x) => ctx.go(x) });
  }
  // D11 代码块头横幅：复制（成功闪词 1s）+自动换行切换（类名驱动——零 data:* 面）
  el.addEventListener('click', (ev) => {
    const copy = ev.target.closest('.md-code-copy');
    if (copy) {
      const block = copy.closest('.md-code-block');
      const code = block ? block.querySelector('pre code') : null;
      const text = code ? String(code.textContent || '') : '';
      if (text && navigator.clipboard && typeof navigator.clipboard.writeText === 'function') {
        navigator.clipboard.writeText(text).then(() => {
          copy.textContent = tr('复制成功');
          copy.classList.add('ok');
          setTimeout(() => { copy.textContent = tr('复制'); copy.classList.remove('ok'); }, 1000);
        }).catch(() => { /* 剪贴板缺席=静默（不冒充成功） */ });
      }
      return;
    }
    const wrap = ev.target.closest('.md-code-wrap');
    if (wrap) {
      const block = wrap.closest('.md-code-block');
      if (block) {
        const on = block.classList.toggle('wrap');
        wrap.textContent = on ? tr('自动换行') : tr('取消自动换行');
      }
    }
  });
  ensureCtxBar(el, CHAT.sessionId);
  refreshCtxChunks(el);
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
    CHAT.lastNodes = [];
    CHAT.refsChunks = [];
    CHAT.paintRefs = null;
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


