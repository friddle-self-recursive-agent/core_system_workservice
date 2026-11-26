/* events.js —— 事件流 + 事件详情。列表=快照 events 聚合（patch 增量实时推进）；
 * 详情=GET /v1/events/{id} 全文现读 + involved 可点链路 + 关联 run/session/loop 链路条
 * （可得的字段才显示，契约未携带=不显示不编造）。过滤态=视图态（非业务状态）。 */

import { esc, relAge, timeHtml, trunc } from '../core/format.js';
import { chip, chipGo, kv, emptyBox, screenHead, involvedRoute } from '../render/chips.js';
import { wfAnchorOfRoute } from './workflows.js';

/* 视图过滤态（内存态——刷新回默认；同旧 console S.evFilter 词形）。 */
const filter = { level: '', q: '' };

const LEVEL_CLS = { critical: 'esc', error: 'esc', warn: 'warn', warning: 'warn', info: 'mute', debug: 'mute' };

function levelChip(e) {
  const lv = String((e && e.level) || 'info');
  return chip(lv, LEVEL_CLS[lv] || 'mute');
}

/** involved → 可点 chip 串（Kind/scope/name 与 '#/' 词形可点，其余纯文本）。 */
function involvedHtml(e) {
  const s = String((e && e.involved) || '');
  if (!s) return chip('involved：—', 'mute');
  return s.split(/[,;]\s*/).filter(Boolean).map((part) => {
    const route = involvedRoute(part.trim());
    return route ? chipGo(part.trim(), route, 'mute', true, '查看 ' + part.trim()) : chip(part.trim(), 'mute');
  }).join(' ');
}

/** 链路条：事件字段尽力关联 run/session/loop（可得才显示）。 */
function linksHtml(e) {
  const out = [];
  const route = String(e && e.route || '');
  const wf = wfAnchorOfRoute(route);
  if (wf) out.push(chipGo('workflow 锚 ' + wf, 'wf/' + encodeURIComponent(wf), 'acc', true, 'route ' + route + ' → 工作流'));
  if (e && e.run) out.push(chipGo('run #' + e.run, route && wf ? 'wf/' + encodeURIComponent(wf) + '/r/' + encodeURIComponent(String(e.run)) : 'workflows', 'mute', true, '关联运行'));
  const sess = String((e && (e.agentSession || e.session)) || '');
  if (sess) out.push(chipGo('会话 ' + trunc(sess, 18), 'session/' + encodeURIComponent(sess), 'mute', true, '会话详情'));
  const wake = e && (e.wake != null ? e.wake : (e.loopId != null ? e.loopId : null));
  if (wake != null) out.push(chip('循环唤醒 #' + wake, 'mute'));
  return out.join(' ');
}

function eventRow(e) {
  const id = e && (e.id != null ? e.id : (e.eventId != null ? e.eventId : ''));
  return '<div class="mini-row" data-go="event/' + encodeURIComponent(String(id)) + '">' +
    '<span class="w-ev-dot" data-level="' + esc(String(e && e.level || 'info')) + '"></span>' +
    '<div class="mr-main"><div>' + esc(String(e && e.type || '未命名事件')) + ' ' + levelChip(e) + '</div>' +
    '<div class="mr-sub mono">' + esc(trunc(String(e && e.involved || '—'), 60)) + '</div></div>' +
    '<span class="w-ev-time mono">' + esc(relAge(e && (e.ts || e.at || e.time)) || '—') + '</span>' +
    '</div>';
}

/** 事件流列表。 */
function listScreen(el, ctx) {
  let rows = ctx.store.events().slice();
  if (filter.level) rows = rows.filter((e) => String(e && e.level || '') === filter.level);
  if (filter.q) {
    const q = filter.q.toLowerCase();
    rows = rows.filter((e) => (
      String(e && (e.type || '') + ' ' + (e.involved || '') + ' ' + (e.summary || e.message || ''))
    ).toLowerCase().includes(q));
  }
  const fbar = '<div class="filter-bar">' +
    '<input class="inp" id="ev-q" placeholder="按类型/对象/摘要过滤…" value="' + esc(filter.q) + '">' +
    ['', 'error', 'warn', 'info'].map((lv) =>
      '<span class="fchip' + (filter.level === lv ? ' on' : '') + '" data-act="ev-level" data-lv="' + esc(lv) + '">' + esc(lv || '全部级别') + '</span>'
    ).join('') + '</div>';

  el.innerHTML = '<div class="view-inner">' + screenHead('事件流', '归档事件（快照聚合 + patch 增量；点击看全文与链路）') +
    fbar +
    (rows.length
      ? '<div class="card">' + rows.slice(0, 300).map(eventRow).join('') + (rows.length > 300 ? '<div class="small faint" style="padding:8px 14px">…共 ' + rows.length + ' 行——用过滤收窄。</div>' : '') + '</div>'
      : emptyBox('窗口内无事件', filter.q || filter.level ? '当前过滤零命中——放宽条件再试。' : '快照 events 聚合为空（不装假）。')) +
    '</div>';

  const inp = el.querySelector('#ev-q');
  if (inp) {
    inp.addEventListener('input', () => { filter.q = inp.value; listScreen(el, ctx); });
    inp.addEventListener('keydown', (ev) => { if (ev.key === 'Enter') ev.preventDefault(); });
  }
  el.querySelectorAll('[data-act="ev-level"]').forEach((c) => {
    c.addEventListener('click', () => { filter.level = c.dataset.lv || ''; listScreen(el, ctx); });
  });
}

/** 事件详情（全文现读——GET /v1/events/{id}）。 */
async function detailScreen(el, ctx, id) {
  const stamp = typeof location !== 'undefined' ? location.hash : '';
  el.innerHTML = '<div class="view-inner">' + screenHead('事件详情', 'id=' + id) +
    '<div class="card"><div class="card-b">' + kv([['状态', chip('现读中…', 'mute')]]) + '</div></div>' +
    '<div class="dh-actions" style="margin-top:10px">' + chipGo('← 回事件流', 'events', 'mute', false) + '</div></div>';

  const r = await ctx.api.event(id);
  if (!el.isConnected || (typeof location !== 'undefined' ? location.hash : '') !== stamp) return; // 路由已变=丢弃
  const box = el.querySelector('.card .card-b');
  if (!box) return;
  if (!r.ok) {
    box.innerHTML = '<div class="ob-empty">事件全文暂不可达（' + esc(String(r.error || r.status)) + '）——恢复后重进即现读。</div>';
    return;
  }
  const e = r.data && typeof r.data === 'object' ? r.data : { raw: r.data };
  const known = ['type', 'level', 'ts', 'at', 'time', 'involved', 'dedupKey', 'routedTo', 'summary', 'message', 'route', 'run', 'agentSession', 'session', 'wake'];
  const rows = known.filter((k) => e[k] != null && e[k] !== '').map((k) => [k, k === 'ts' || k === 'at' || k === 'time' ? timeHtml(e[k]) : esc(trunc(String(e[k]), 200))]);
  box.innerHTML =
    '<div class="ev-head">' + esc(String(e.type || '未命名事件')) + ' ' + levelChip(e) + ' ' + linksHtml(e) + '</div>' +
    (rows.length ? kv(rows) : '') +
    (e.involved ? '<div class="ev-involved">' + involvedHtml(e) + '</div>' : '') +
    '<details class="md-raw"><summary>原始载荷（JSON）</summary><pre class="mono md-raw-pre">' + esc(JSON.stringify(e, null, 2)) + '</pre></details>';
}

export function render(el, ctx) {
  const id = ctx.params && ctx.params.id;
  if (id != null && id !== '') detailScreen(el, ctx, String(id));
  else listScreen(el, ctx);
}
