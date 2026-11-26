/* chat.js —— 会话屏（2026-11-26 用户裁决：直接复用宿主 DeepSeek 风格 session 界面）：
 *   #/chat/:agent —— 左栏会话清单 + 右主区 iframe 嵌宿主原生会话界面（同源 cookie 共享，
 *   原生界面自带消息流/tool 折叠/composer——零重复建设）；「新窗口打开」弹全屏原生。
 *   #/session/:id —— 同布局 + 下方 ContextPack 契约面（sessionctx.js，G-14）。
 * faces 契约保留：window.__CC.chatFaces.{sessions,transcript} 注入后清单走宿主现读；
 * 未注入=快照 sessions 投影（诚实降级，不冒充转录）。 */

import { esc, trunc, relAge } from '../core/format.js';
import { chip, chipGo, kv, screenHead } from '../render/chips.js';
import { mountSessionCtx } from './sessionctx.js';
import { t as tr } from '../core/i18n.js';

function faces() {
  return (typeof window !== 'undefined' && window.__CC && window.__CC.chatFaces) || null;
}

/* 宿主原生界面地址（同源根——cookie 已随门牌落地；embedded=1 供宿主侧未来做减法布局）。 */
function hostUiUrl() {
  try { return location.origin + '/?embedded=1'; } catch { return '/'; }
}

function sessionRow(s, activeId) {
  const id = String(s && (s.id || s.sessionId || ''));
  if (!id) return '';
  const state = String((s && (s.state || s.status)) || '');
  const cls = state === 'suspected-stall' || state === 'escalated' ? 'esc' : state ? 'acc' : 'mute';
  return '<div class="mini-row' + (id === String(activeId || '') ? ' sel' : '') + '" data-sess="' + esc(id) + '">' +
    '<div class="mr-main"><div><b class="mono">' + esc(trunc(id, 30)) + '</b> ' + (state ? chip(state, cls) : '') + '</div>' +
    '<div class="mr-sub small">' + esc(String((s && (s.agent || s.route)) || '—')) +
    (s && s.startedAt ? ' · ' + esc(relAge(s.startedAt) || '') : '') + '</div></div>' +
    '</div>';
}

/* 会话清单（faces.sessions 注入优先；快照 sessions 投影次之；双缺=诚实空态）。 */
function sessionList(ctx, activeId, agent) {
  const f = faces();
  let rows = null;
  if (f && typeof f.sessions === 'function') {
    try { rows = f.sessions(ctx) || []; } catch { rows = null; }
  }
  if (!Array.isArray(rows)) rows = ctx.store.sessions();
  if (agent) rows = rows.filter((s) => !s || !s.agent || String(s.agent) === agent);
  const list = rows.map((s) => sessionRow(s, activeId)).filter(Boolean);
  if (!list.length) {
    return '<div class="card"><div class="card-b"><div class="small faint">' + esc(tr('会话清单暂不可得')) +
      (f ? '' : esc(tr('（faces 未注入且快照 sessions 缺席）'))) +
      esc(tr('——不冒充在册会话。')) + '</div></div></div>';
  }
  return '<div class="card sess-list">' + list.join('') + '</div>';
}

/* 右主区：宿主原生界面 iframe（复用 DeepSeek 风格消息流/tool 折叠/composer）。 */
function hostEmbedHtml(agent) {
  return '<div class="host-embed-wrap">' +
    '<div class="host-embed-bar">' +
    chip(tr('宿主原生界面'), 'acc') +
    '<span class="small faint">' + esc(tr('消息流/tool 折叠/composer 全套复用宿主 DeepSeek 风格界面（同源嵌入门牌已共享）')) + '</span>' +
    '<button class="btn sm" data-host-open>' + esc(tr('新窗口打开')) + '</button>' +
    '</div>' +
    '<iframe class="host-embed" src="' + esc(hostUiUrl()) + '" data-host-frame ' +
    'referrerpolicy="same-origin" allow="clipboard-write"></iframe>' +
    (agent ? '<div class="small faint" style="margin-top:6px">' + esc(tr('agent=') ) + '<code class="mono">' + esc(agent) + '</code> · ' +
      esc(tr('在原生界面侧栏选择该 agent 的会话；CLI 通路：')) + '<code class="mono">corectl chat ' + esc(agent) + '</code></div>' : '') +
    '</div>';
}

function bindHostOpen(el) {
  const btn = el.querySelector('[data-host-open]');
  if (btn) btn.addEventListener('click', () => { try { window.open(hostUiUrl(), '_blank'); } catch { /* 弹窗拦截=静默 */ } });
}

/** #/chat/:agent —— agent 会话入口。 */
function chatScreen(el, ctx) {
  const agent = String(ctx.params.agent || 'root-agent');
  el.innerHTML = '<div class="view-inner">' +
    screenHead(tr('会话'), 'agent=' + agent + tr('（对话是一等面——同一 dsh 会话机制投影）'),
      chipGo(tr('看 root 会话'), 'chat/root-agent', 'acc', false)) +
    '<div class="chat-grid chat-grid-new">' +
    '<div class="chat-side"><h2 class="sec-title">' + esc(tr('会话清单')) + '</h2>' + sessionList(ctx, '', agent) + '</div>' +
    '<div class="chat-main">' + hostEmbedHtml(agent) + '</div>' +
    '</div></div>';
  el.querySelectorAll('[data-sess]').forEach((row) => {
    row.addEventListener('click', () => ctx.go('session/' + encodeURIComponent(row.dataset.sess)));
  });
  bindHostOpen(el);
}

/** #/session/:id —— 原生转录复用 + 上下文面板（G-14：ContextPack 查看与编辑分叉）。 */
function sessionScreen(el, ctx) {
  const id = String(ctx.params.id || '');
  el.innerHTML = '<div class="view-inner">' +
    screenHead(tr('会话转录'), 'id=' + id, chipGo(tr('← 会话'), 'chat/root-agent', 'mute', false)) +
    '<div class="chat-grid chat-grid-new">' +
    '<div class="chat-side"><h2 class="sec-title">' + esc(tr('会话清单')) + '</h2>' + sessionList(ctx, id) + '</div>' +
    '<div class="chat-main">' + hostEmbedHtml('') +
    '<h2 class="sec-title" style="margin-top:14px">' + esc(tr('上下文')) + '</h2>' +
    '<div data-ctx-panel><div class="card"><div class="card-b"><div class="small faint">' +
    esc(tr('上下文注水中…（GET /v1/sessions/')) + esc(encodeURIComponent(id)) + '/context）</div></div></div></div>' +
    '</div></div></div>';
  el.querySelectorAll('[data-sess]').forEach((row) => {
    row.addEventListener('click', () => ctx.go('session/' + encodeURIComponent(row.dataset.sess)));
  });
  bindHostOpen(el);
  mountSessionCtx(el, ctx, id);
}

export function render(el, ctx) {
  if (ctx.params && ctx.params.id != null && ctx.params.id !== '') sessionScreen(el, ctx);
  else chatScreen(el, ctx);
}
