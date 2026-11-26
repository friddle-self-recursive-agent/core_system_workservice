/* chat.js —— 会话屏（§19.6 用户裁决版：会话屏=原生 DeepSeek 会话主面）：
 * 布局=两列——左列（~300px）会话/Agent 列表（store 快照注水+点选）；右列=宿主原生会话主面
 * 全高 iframe 直开 dsh 同域词形 /?embedded=1#/chat/<sessionId|agentName>（真实转录+真实
 * composer+工具过程+任务条+模型切换全在原生面——权威轨迹=原生面）。
 * 自绘 composer 与本地线程自会话屏退场；composerHtml/bindComposer 导出保留（总览大屏在用）。
 * 重绘纪律：store patch 触发的整屏重绘=局部换左列（iframe 不动——滚动与输入态保真）；
 * 会话点选=就地换 iframe src（不重绘、不改 hash）。ContextPack 面板=右列上方可折叠条（默认收起）。 */

import { esc, trunc, relAge } from '../core/format.js';
import { chip, screenHead } from '../render/chips.js';
import { renderMessage } from '../render/refs.js';
import { renderMarkdown } from '../render/md.js';
import { mountSessionCtx } from './sessionctx.js';
import { t as tr, getLocale } from '../core/i18n.js';

function faces() {
  return (typeof window !== 'undefined' && window.__CC && window.__CC.chatFaces) || null;
}

/* 线程态（页面级内存——诚实口径：本地线程只记本控制台发出的消息；权威轨迹=宿主原生会话面）。 */
const THREAD = { sessionId: '', msgs: [] };

/* 上次渲染标记（路由词形+locale——store patch 重绘 vs 导航/locale 切换的判据）。 */
const LAST = { route: '', paramKey: '', locale: '' };

function sessionOptions(ctx, cur) {
  const rows = (ctx.store.sessions() || []).map((s) => String(s && (s.id || s.sessionId) || '')).filter(Boolean);
  const seen = new Set(rows);
  const opts = rows.slice(0, 30).map((id) =>
    '<option value="' + esc(id) + '"' + (id === cur ? ' selected' : '') + '>' + esc(trunc(id, 30)) + '</option>').join('');
  const extra = cur && !seen.has(cur)
    ? '<option value="' + esc(cur) + '" selected>' + esc(trunc(cur, 30)) + '</option>' : '';
  return '<option value="">' + esc(tr('— 选择 session（或输入）—')) + '</option>' + extra + opts;
}

/* 对话 composer（session 可选 + 文本域 + 发送——总览大屏在用；会话屏本体已改原生主面）。 */
export function composerHtml(ctx, curSession) {
  const f = faces();
  const canSend = !!(f && typeof f.send === 'function');
  return '<div class="cc-composer">' +
    '<div class="cc-comp-row">' +
    '<select class="inp cc-sess" data-cc-sess>' + sessionOptions(ctx, curSession) + '</select>' +
    '<input class="inp cc-sess-free" data-cc-sess-free placeholder="' + esc(tr('或手填 session id')) + '" value="' + esc(curSession || '') + '">' +
    '</div>' +
    '<div class="cc-comp-row">' +
    '<textarea class="inp cc-text" data-cc-text rows="2" placeholder="' +
    esc(tr('提交需求／找文档／更新文档——例如：进行代码审核，看看实际和文档差距有多大，并进行修复')) + '"></textarea>' +
    '<button class="btn primary cc-send" data-cc-send' + (canSend ? '' : ' disabled') + '>' + esc(tr('发送')) + '</button>' +
    '</div>' +
    '<div class="small faint">' +
    (canSend
      ? esc(tr('Enter 发送 · queue 语义（送达 agent 会话；结论由 agent 落文档后经状态徽与节点图可见）'))
      : esc(tr('对话通路未注入（chatfaces.js 缺席）——送达面显式缺席，不冒充已发送'))) +
    '</div></div>';
}

/* 线程渲染（消息数组 → 气泡流；expand 内建——composerHtml/bindComposer 保留通路的书签位）。 */
function threadHtml(msgs, names) {
  if (!msgs.length) {
    return '<div class="card"><div class="ob-empty">' + esc(tr('线程为空——发第一条需求；agent 的完整轨迹在原生界面（新窗口）可见。')) + '</div></div>';
  }
  return '<div class="card thread">' + msgs.map((m) => renderMessage(m, names)).join('') + '</div>';
}

/* 完整文档模式弹层（data-doc-full → 现读对象正文 → 渲染 md + 写入草稿入口）。 */
function bindDocFull(el, ctx) {
  el.querySelectorAll('[data-doc-full]').forEach((btn) => {
    btn.addEventListener('click', async () => {
      const ref = String(btn.getAttribute('data-doc-full') || '');
      const parts = ref.split('/');
      if (parts.length < 3) return;
      const [kind, scope, name] = parts;
      const r = await ctx.api.object(decodeURIComponent(kind), decodeURIComponent(scope), decodeURIComponent(name));
      const body = r.ok && r.data && r.data.content != null ? String(r.data.content) : (tr('正文暂不可读（') + (r.ok ? tr('内容缺席') : String(r.error || r.status)) + '）');
      let ov = document.querySelector('[data-doc-ov]');
      if (!ov) {
        ov = document.createElement('div');
        ov.setAttribute('data-doc-ov', '1');
        document.body.appendChild(ov);
      }
      ov.className = 'doc-ov';
      ov.innerHTML = '<div class="doc-ov-card">' +
        '<div class="doc-ov-h"><b class="mono">' + esc(decodeURIComponent(kind) + '/' + decodeURIComponent(scope) + '/' + decodeURIComponent(name)) + '</b>' +
        '<button class="btn sm" data-doc-ov-close>' + esc(tr('关闭')) + '</button>' +
        '<button class="btn primary sm" data-doc-ov-edit>' + esc(tr('写入草稿（编辑器）')) + '</button></div>' +
        '<div class="doc-ov-b md-body">' + renderMarkdown(body) + '</div></div>';
      ov.querySelector('[data-doc-ov-close]').addEventListener('click', () => { ov.className = 'doc-ov hidden'; });
      ov.querySelector('[data-doc-ov-edit]').addEventListener('click', () => {
        ctx.go('doc/' + encodeURIComponent(decodeURIComponent(kind)) + '/' + encodeURIComponent(decodeURIComponent(scope)) + '/' + encodeURIComponent(decodeURIComponent(name)));
        ov.className = 'doc-ov hidden';
      });
    });
  });
}

/* 绑定 composer 行为（session 解析+发送+本地回执——总览大屏在用）。 */
export function bindComposer(el, ctx, rootEl) {
  const sendBtn = rootEl.querySelector('[data-cc-send]');
  const sessSel = rootEl.querySelector('[data-cc-sess]');
  const sessFree = rootEl.querySelector('[data-cc-sess-free]');
  const textTa = rootEl.querySelector('[data-cc-text]');
  if (!sendBtn || !textTa) return;
  const doSend = async () => {
    const text = textTa.value.trim();
    if (!text) return;
    const sessionId = (sessFree && sessFree.value.trim()) || (sessSel ? sessSel.value : '') || 'root';
    THREAD.sessionId = sessionId;
    const f = faces();
    if (!f || typeof f.send !== 'function') return;
    THREAD.msgs.push({ role: 'user', text, ts: new Date().toISOString().slice(11, 19) });
    textTa.value = '';
    paintThread(rootEl, ctx);
    const r = await f.send(sessionId, text);
    if (r && r.status === 200) {
      THREAD.msgs.push({
        role: 'assistant', name: 'rootAgent', seq: (r.data && r.data.requestId) || '',
        ts: new Date().toISOString().slice(11, 19),
        text: tr('已送达（queue）——agent 会话已受理。结论将由 agent 生成并自行落文档；文档状态徽与 workflow 节点图随后可见。原生轨迹：'),
        tool: 'send → sessionController.prompt · requestId=' + ((r.data && r.data.requestId) || '—'),
      });
    } else {
      THREAD.msgs.push({
        role: 'assistant', name: 'system',
        text: tr('送达失败（') + ((r && r.status) || '—') + '）' + ((r && r.data && r.data.error) ? ' · ' + r.data.error : ''),
      });
    }
    paintThread(rootEl, ctx);
  };
  sendBtn.addEventListener('click', doSend);
  textTa.addEventListener('keydown', (ev) => {
    if (ev.key === 'Enter' && !ev.shiftKey) { ev.preventDefault(); doSend(); }
  });
}

function paintThread(rootEl, ctx) {
  const box = rootEl.querySelector('[data-thread]');
  if (!box) return;
  const names = new Set((ctx.store.objects() || []).map((o) => o.name).filter(Boolean));
  box.innerHTML = threadHtml(THREAD.msgs, names);
  box.scrollTop = box.scrollHeight;
  bindDocFull(box, ctx);
}

/* ---------- 原生会话主面（§19.6） ---------- */

/* 原生面词形：同域 /?embedded=1#/chat/<target>——target=session id 或 agent 名（原生路由双词形）。 */
function nativeSrc(target) {
  let origin = '/';
  try { origin = location.origin + '/'; } catch { /* 无 origin 环境（file://）降级相对词形 */ }
  return origin + '?embedded=1#/chat/' + (target ? encodeURIComponent(String(target)) : '');
}

/* 左列：会话行（活性账 store.sessions()——state/lastBeat 词形投影；点选=就地换 src）。 */
function sessRowsHtml(ctx, cur) {
  const rows = (ctx.store.sessions() || []).map((s) => {
    const id = String((s && (s.id || s.sessionId)) || '');
    if (!id) return '';
    const state = String((s && s.state) || '');
    const beat = String((s && s.lastBeat) || '');
    const sel = id === cur ? ' style="outline:1px solid var(--accent);background:var(--panel-2);border-radius:8px"' : '';
    return '<div class="dash-row" data-sess-go="' + esc(id) + '" role="link"' + sel +
      ' title="' + esc(trunc(String((s && s.lastAction) || ''), 120)) + '">' +
      '<div class="dr-main"><b class="mono">' + esc(trunc(id, 24)) + '</b></div>' +
      '<div class="dr-side">' +
      (state ? chip(state, state === 'escalated' || state === 'suspected-stall' ? 'esc' : 'acc') : '') +
      (beat ? '<span class="small faint">' + esc(relAge(beat) || '') + '</span>' : '') +
      '</div></div>';
  }).filter(Boolean).join('');
  return rows || '<div class="small faint" style="padding:4px 6px">' +
    esc(tr('快照无活性会话行——可在右侧原生面直接对话（新会话由宿主原生建档）')) + '</div>';
}

/* 左列：Agent 行（kind=Agent 在册对象——#/chat/<agent> 原生按 agent 打开会话）。 */
function agentRowsHtml(ctx) {
  const rows = (ctx.store.objects() || []).filter((o) => o && o.kind === 'Agent' && o.name).map((o) =>
    '<div class="dash-row" data-go="' + esc('chat/' + encodeURIComponent(String(o.name))) + '" role="link">' +
    '<div class="dr-main"><b class="mono">' + esc(trunc(String(o.name), 24)) + '</b></div>' +
    '<div class="dr-side"><span class="small faint">' + esc(String(o.scope || 'global')) + '</span></div></div>').join('');
  return rows || '<div class="small faint" style="padding:4px 6px">' + esc(tr('快照暂无 Agent 对象枚举')) + '</div>';
}

/* ContextPack 折叠条（默认收起——不占主面高度；展开态由 sessionctx 自管）。 */
function ctxBarHtml(sessionId) {
  if (!sessionId) return '';
  return '<details class="ex-block" style="margin:0 0 10px;background:var(--panel-2)">' +
    '<summary class="ex-sum"><span class="ex-chev">▶</span>' + esc(tr('上下文（ContextPack 契约面）')) + '</summary>' +
    '<div class="ex-body"><div data-ctx-panel><div class="card"><div class="card-b"><div class="small faint">' +
    esc(tr('上下文注水中…（GET /v1/sessions/')) + esc(encodeURIComponent(sessionId)) + '/context）</div></div></div></div></div></details>';
}

/* 右列主面：装载占位（load 后撤除——不冒充已就绪）+ 全高 iframe（圆角卡片边框）。 */
function faceHtml(target) {
  return '<div style="position:relative">' +
    '<div data-native-ph style="position:absolute;inset:0;display:flex;align-items:center;justify-content:center;' +
    'border:1px dashed var(--border);border-radius:12px;background:var(--panel-2);z-index:2;pointer-events:none">' +
    '<div class="small faint">' + esc(tr('原生会话面装载中…')) + '</div></div>' +
    '<iframe data-native-iframe data-target="' + esc(target) + '" title="' + esc(tr('宿主原生会话面（真实转录与发送）')) +
    '" src="' + esc(nativeSrc(target)) + '" referrerpolicy="same-origin" ' +
    'style="width:100%;height:calc(100vh - 96px);border:1px solid var(--border);border-radius:12px;background:var(--panel-2);display:block"></iframe>' +
    '</div>';
}

/* iframe load → 撤装载占位（换 src 后复位——下一次 load 再撤）。 */
function bindFace(el) {
  const ifr = el.querySelector('[data-native-iframe]');
  const ph = el.querySelector('[data-native-ph]');
  if (ifr && ph) ifr.addEventListener('load', () => { ph.style.display = 'none'; });
}

/* 就地换 src（保滚动——不重绘整屏；占位词复位待 load 撤除）。 */
function swapFace(el, target) {
  const ifr = el.querySelector('[data-native-iframe]');
  if (!ifr) return;
  const src = nativeSrc(target);
  if (ifr.getAttribute('src') !== src) {
    ifr.setAttribute('src', src);
    const ph = el.querySelector('[data-native-ph]');
    if (ph) ph.style.display = '';
  }
  ifr.setAttribute('data-target', String(target));
}

/* 会话点选：THREAD.sessionId=该 id + iframe 就地换 src + 左列高亮 + ContextPack 条补挂。 */
function bindSessRows(el, ctx) {
  el.querySelectorAll('[data-sess-go]').forEach((row) => {
    row.addEventListener('click', () => {
      const id = String(row.getAttribute('data-sess-go') || '');
      if (!id) return;
      THREAD.sessionId = id;
      swapFace(el, id);
      const listBox = el.querySelector('[data-sess-list]');
      if (listBox) { listBox.innerHTML = sessRowsHtml(ctx, id); bindSessRows(el, ctx); }
      ensureCtxBar(el, ctx, id);
    });
  });
}

/* ContextPack 条补挂（agent 面→点选会话后：holder 现挂折叠条+面板注水新 id）。 */
function ensureCtxBar(el, ctx, id) {
  const holder = el.querySelector('[data-ctx-holder]');
  if (!holder) return;
  if (!holder.querySelector('[data-ctx-panel]')) holder.innerHTML = ctxBarHtml(id);
  mountSessionCtx(el, ctx, id);
}

/* 「新窗口原生界面」——同词形弹原生窗（不经 data-go 委托：目标在宿主，不在控制台路由）。 */
function bindPop(el, target) {
  const pop = el.querySelector('[data-native-pop]');
  if (pop) pop.addEventListener('click', () => {
    try { window.open(nativeSrc(target), '_blank'); } catch { /* 弹窗被拦=显式再点即可 */ }
  });
}

/** 会话屏主体（两列：左=会话/Agent 列表；右=原生会话主面全高直开）。 */
function dialogScreen(el, ctx, target) {
  el.innerHTML = '<div class="view-inner">' +
    screenHead(tr('会话'), tr('原生 DeepSeek 会话面——真实转录与发送；左侧切换会话'),
      chip(tr('新窗口原生界面'), 'mute', false, 'data-native-pop="1" style="cursor:pointer"')) +
    '<div style="display:grid;grid-template-columns:300px minmax(0,1fr);gap:14px;align-items:start">' +
    '<aside class="card" style="padding:8px;max-height:calc(100vh - 96px);overflow:auto">' +
    '<div class="small faint" style="padding:2px 6px 8px;font-weight:700">' + esc(tr('会话（活性账）')) + '</div>' +
    '<div data-sess-list>' + sessRowsHtml(ctx, target) + '</div>' +
    '<div class="small faint" style="padding:12px 6px 8px;font-weight:700">' + esc(tr('Agent')) + '</div>' +
    '<div data-agent-list>' + agentRowsHtml(ctx) + '</div>' +
    '</aside>' +
    '<section style="min-width:0">' +
    '<div data-ctx-holder>' + ctxBarHtml(THREAD.sessionId) + '</div>' +
    faceHtml(target) +
    '</section>' +
    '</div></div>';
  bindFace(el);
  bindSessRows(el, ctx);
  bindPop(el, target);
  if (THREAD.sessionId) mountSessionCtx(el, ctx, THREAD.sessionId);
}

/** #/chat/:agent 与 #/session/:id —— 会话屏（原生 DeepSeek 会话主面）。 */
export function render(el, ctx) {
  const id = ctx.params && ctx.params.id;
  const agent = ctx.params && ctx.params.agent;
  const route = String(ctx.route || '');
  const paramKey = (id != null && id !== '') ? 'session/' + String(id) : 'chat/' + String(agent || '');
  const locale = getLocale();
  const fresh = LAST.route !== route || LAST.paramKey !== paramKey || LAST.locale !== locale;
  LAST.route = route; LAST.paramKey = paramKey; LAST.locale = locale;
  if (fresh) {
    if (id != null && id !== '') {
      THREAD.sessionId = String(id); // #/session/:id —— 直开该会话
    } else if (agent != null && agent !== '') {
      /* #/chat/:agent —— 词形双关（agent 名或 session id）：活性账命中=session 面，否则=agent 面。 */
      THREAD.sessionId = sessKnown(ctx, String(agent)) ? String(agent) : '';
    }
  } else if (!THREAD.sessionId && agent != null && agent !== '' && sessKnown(ctx, String(agent))) {
    /* 深链 #/chat/<sessionId> 在快照到货后升格=session 面（补 ContextPack 条——一次性重建）。 */
    THREAD.sessionId = String(agent);
    dialogScreen(el, ctx, THREAD.sessionId);
    return;
  }
  const target = THREAD.sessionId || String(agent || id || '');
  /* store patch 重绘=局部换左列（iframe 不动——滚动/输入态保真；点选换 src 已同步 data-target）。 */
  const face = el.querySelector('[data-native-iframe]');
  if (!fresh && face && face.getAttribute('data-target') === target) {
    const sess = el.querySelector('[data-sess-list]');
    if (sess) { sess.innerHTML = sessRowsHtml(ctx, target); bindSessRows(el, ctx); }
    const ag = el.querySelector('[data-agent-list]');
    if (ag) ag.innerHTML = agentRowsHtml(ctx);
    return;
  }
  dialogScreen(el, ctx, target);
}

/* 活性账命中判据：param 是否为已知 sessionId（#/chat/:agent 词形双关的裁决面）。 */
function sessKnown(ctx, candidate) {
  return (ctx.store.sessions() || []).some((s) => String((s && (s.id || s.sessionId)) || '') === candidate);
}
