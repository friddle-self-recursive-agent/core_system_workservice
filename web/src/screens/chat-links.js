/* chat-links.js —— 会话屏关联任务背景卡（R-UX-10 2026-11-28 用户裁决：session 能看到相应的
 * 任务——workflow 节点的描述背景；R-UX-9 双向链接同口径）。快照反查 Task.agentSession ∪
 * 事件 agentSession → 任务背景卡（title/phase/summary/来源 workflow）+事件 chips；
 * 无关联=诚实空态。纯函数零 DOM（node 单测直测）；一切动态文本过 esc——零原文直插。 */
import { esc, trunc } from '../core/format.js';
import { t as tr } from '../core/i18n.js';

/** 会话 → 关联任务行（纯函数）。 */
export function sessionTasksOf(store, sessionId) {
  const sid = String(sessionId || '');
  if (!sid) return [];
  return (store.tasks() || []).filter((t) =>
    String((t && t.status && (t.status.agentSession || t.status.sessionId)) || '') === sid
  ).slice(0, 3).map((t) => ({
    id: String((t && t.scope) || 'global') + '/' + String((t && t.name) || ''),
    title: String((t && t.spec && t.spec.title) || (t && t.name) || ''),
    phase: String((t && t.status && (t.status.phase || t.status)) || ''),
    summary: String((t && t.spec && (t.spec.summary || t.spec.title)) || ''),
    source: String((t && t.spec && (t.spec.source || t.spec.sourceRef)) || ''),
  }));
}

export function sessBacklinksHtml(store, sessionId) {
  const sid = String(sessionId || '');
  if (!sid) return '';
  const tasks = sessionTasksOf(store, sid);
  const evs = (store.events() || []).filter((e) => {
    const s = String((e && (e.agentSession || e.session)) || '');
    return s === sid;
  }).slice(0, 4);
  // 任务背景卡（R-UX-10：任务 title/相位/summary/来源 workflow 节点描述——不是裸 id 词形）
  const taskCards = tasks.map((t) =>
    '<div class="mini-row"><div class="mr-main"><div>' +
    (t.phase ? '<span class="chip acc tiny">' + esc(t.phase) + '</span> ' : '') +
    '<b>' + esc(trunc(t.title || t.id, 44)) + '</b></div>' +
    '<div class="mr-sub small">' +
    (t.summary && t.summary !== t.title ? esc(trunc(t.summary, 120)) + '<br>' : '') +
    (t.source ? esc(tr('来源：')) + '<span class="mono">' + esc(trunc(t.source, 64)) + '</span> · ' : '') +
    esc(tr('id ')) + '<span class="mono">' + esc(trunc(t.id, 48)) + '</span></div></div>' +
    '<span class="w-ev-time">' +
    '<span class="chip acc tiny click" data-go="task/' + encodeURIComponent(t.id) + '" title="' + esc(t.id) + tr('——任务详情') + '">' + esc(tr('任务详情')) + '</span></span></div>'
  ).join('');
  const evChips = evs.map((e) =>
    '<span class="chip mute tiny click" data-go="event/' + encodeURIComponent(String(e.id ?? '')) + '" title="' + esc(tr('事件详情')) + '">' +
    esc(trunc(String(e.type || ''), 26)) + '</span>').join(' ');
  if (!taskCards && !evChips) {
    return '<div class="small faint" data-sess-backlinks="1">' + esc(tr('快照反查（Task.agentSession ∪ 事件 agentSession）无在册关联——会话型独立会话（零编造）。')) + '</div>';
  }
  return '<div data-sess-backlinks="1">' +
    (taskCards ? '<div class="small faint" style="margin-bottom:4px">' + esc(tr('执行的任务（本会话由任务拉起——背景如下）')) + '</div>' + taskCards : '') +
    (evChips ? '<div class="small" style="margin-top:' + (taskCards ? '8' : '0') + 'px"><span class="faint">' + esc(tr('关联事件：')) + '</span> ' + evChips + '</div>' : '') +
    '</div>';
}
