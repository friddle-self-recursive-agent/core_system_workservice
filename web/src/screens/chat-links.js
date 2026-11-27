/* chat-links.js —— 会话屏 V-10 回链（2026-11-27 从 chat.js 拆出：R-ENG-7 ≤800 行）。
 * 快照三源反查：Task.agentSession ∪ 事件 agentSession → 可点 chips；无关联=诚实空态。
 * 词形/样式零改动（唯一真源=chat.css；测试仍从 screens/chat.js re-export 取词）。 */
import { esc, trunc } from '../core/format.js';
import { t as tr } from '../core/i18n.js';

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
