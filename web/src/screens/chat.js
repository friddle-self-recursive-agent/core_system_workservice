/* chat.js —— 会话屏（清单 + 转录 + 上下文面板）。faces 注入点保留：window.__CC.chatFaces
 * 同款接口——{sessions(ctx)?, transcript(id)?} 可由宿主注入后现读；未注入=快照 sessions
 * 清单 + 诚实降级词形（不冒充转录内容——零编造）。深链：#/chat/:agent（agent 入口）、
 * #/session/:id（转录+上下文——sessionctx.js 挂载 ContextPack 契约面 G-14）。 */

import { esc, trunc, relAge } from '../core/format.js';
import { chip, chipGo, kv, emptyBox, screenHead } from '../render/chips.js';
import { mountSessionCtx } from './sessionctx.js';
import { t as tr } from '../core/i18n.js';

function faces() {
  return (typeof window !== 'undefined' && window.__CC && window.__CC.chatFaces) || null;
}

function sessionRow(s, activeId) {
  const id = String(s && (s.id || s.sessionId || ''));
  if (!id) return '';
  const state = String((s && (s.state || s.status)) || '');
  const cls = state === 'suspected-stall' || state === 'escalated' ? 'esc' : state ? 'acc' : 'mute';
  return '<div class="mini-row' + (id === String(activeId || '') ? ' sel' : '') + '" data-sess="' + esc(id) + '">' +
    '<div class="mr-main"><div><b class="mono">' + esc(trunc(id, 34)) + '</b> ' + (state ? chip(state, cls) : '') + '</div>' +
    '<div class="mr-sub small">' + esc(String((s && (s.agent || s.route)) || '—')) +
    (s && s.startedAt ? ' · ' + esc(relAge(s.startedAt) || '') : '') + '</div></div>' +
    '</div>';
}

/** 会话清单（faces.sessions 注入优先；快照 sessions 投影次之；双缺=诚实空态）。 */
function sessionList(ctx, activeId) {
  const f = faces();
  let rows = null;
  if (f && typeof f.sessions === 'function') {
    try { rows = f.sessions(ctx) || []; } catch { rows = null; }
  }
  if (!Array.isArray(rows)) rows = ctx.store.sessions();
  const list = rows.map((s) => sessionRow(s, activeId)).filter(Boolean);
  if (!list.length) {
    return tr('<div class="card"><div class="card-b"><div class="small faint">会话清单暂不可得') +
      (f ? '' : tr('（faces 未注入且快照 sessions 缺席）')) + tr('——不冒充在册会话。</div></div></div>');
  }
  return '<div class="card">' + list.join('') + '</div>';
}

/** 转录（faces.transcript 注入优先——返回 HTML/文本；未注入=诚实降级词形）。 */
function transcript(ctx, id) {
  const f = faces();
  if (f && typeof f.transcript === 'function') {
    try {
      const html = f.transcript(id, ctx);
      if (html) return String(html); // faces 实现方担保已净化（宿主契约）
    } catch { /* 注入面失败=落降级词形 */ }
  }
  return tr('<div class="ob-empty">转录面未装配——宿主注入 window.__CC.chatFaces 后现读；') +
    tr('会话原文不入公共库（intake 留痕面为准）。</div>');
}

/** #/chat/:agent —— agent 会话入口（与 root 对话的深链词形兼容）。 */
function chatScreen(el, ctx) {
  const agent = String(ctx.params.agent || 'root-agent');
  const sessions = ctx.store.sessions().filter((s) => !s || !s.agent || String(s.agent) === agent);
  el.innerHTML = '<div class="view-inner">' +
    screenHead(tr('会话'), 'agent=' + agent + tr('（对话是一等面——同一 dsh 会话机制投影）'),
      chipGo(tr('看 root 会话'), 'chat/root-agent', 'acc', false)) +
    '<div class="chat-grid"><div>' +
    tr('<h2 class="sec-title">会话清单</h2>') + sessionList(ctx) +
    '</div><div>' +
    tr('<h2 class="sec-title">入口</h2>') +
    '<div class="card"><div class="card-b">' +
    kv([
      ['agent', esc(agent)],
      [tr('会话'), chip(sessions.length + tr(' 个相关'), sessions.length ? 'acc' : 'mute')],
      ['CLI', '<code class="mono">corectl chat ' + esc(agent) + '</code>'],
    ]) +
    tr('<div class="small faint" style="margin-top:8px">转录与终端面由宿主 faces 注入（__CC.chatFaces / __CC.termFaces）；') +
    tr('未注入=清单降级快照投影，转录面显式缺席词形。</div>') +
    '</div></div></div></div></div>';
  el.querySelectorAll('[data-sess]').forEach((row) => {
    row.addEventListener('click', () => ctx.go('session/' + encodeURIComponent(row.dataset.sess)));
  });
}

/** #/session/:id —— 转录屏 + 上下文面板（G-14：ContextPack/TagCtxShard/takeBudget 查看
 * 与编辑分叉——sessionctx.js 异步注水，hash stamp 丢弃过期响应）。 */
function sessionScreen(el, ctx) {
  const id = String(ctx.params.id || '');
  el.innerHTML = '<div class="view-inner">' +
    screenHead(tr('会话转录'), 'id=' + id, chipGo(tr('← 会话'), 'chat/root-agent', 'mute', false)) +
    tr('<div class="chat-grid"><div><h2 class="sec-title">会话清单</h2>') + sessionList(ctx, id) + '</div>' +
    tr('<div><h2 class="sec-title">转录</h2><div class="card"><div class="card-b transcript">') + transcript(ctx, id) + '</div></div>' +
    tr('<h2 class="sec-title" style="margin-top:14px">上下文</h2>') +
    tr('<div data-ctx-panel><div class="card"><div class="card-b"><div class="small faint">上下文注水中…（GET /v1/sessions/') +
    esc(encodeURIComponent(id)) + '/context）</div></div></div></div>' +
    '</div></div></div>';
  el.querySelectorAll('[data-sess]').forEach((row) => {
    row.addEventListener('click', () => ctx.go('session/' + encodeURIComponent(row.dataset.sess)));
  });
  mountSessionCtx(el, ctx, id);
}

export function render(el, ctx) {
  if (ctx.params && ctx.params.id != null && ctx.params.id !== '') sessionScreen(el, ctx);
  else chatScreen(el, ctx);
}
