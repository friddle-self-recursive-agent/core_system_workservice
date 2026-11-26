/* welcome.js —— 首屏：快照聚合视图。零前端自建状态——一切计数来自 store
 * （/v1/snapshot 基线 + patch 增量）；当前循环卡=GET /v1/loop/current 现读
 * （缺席/失败=诚实降级词形，不冒充无循环——L78 同款纪律）。 */

import { esc, relAge, fmtNum } from '../core/format.js';
import { chip, chipGo, emptyBox, objRoute } from '../render/chips.js';

/** 摘要统计卡（点击深链）。 */
function statCard(label, num, sub, route, cls) {
  return '<div class="stat ' + esc(cls || '') + '" data-go="' + esc(route) + '" role="link">' +
    '<div class="st-num">' + esc(fmtNum(num)) + '</div>' +
    '<div class="st-label">' + esc(label) + '<span class="small faint">' + esc(sub || '') + '</span></div>' +
    '</div>';
}

/** kind 摘要卡（计数+首个对象可点——对象枚举来自快照 refs）。 */
function kindCard(kindRow, refs) {
  const k = kindRow && (kindRow.name || kindRow.kind);
  if (!k) return '';
  const rows = refs.filter((r) => r.kind === k);
  const first = rows[0];
  const firstHtml = first
    ? chipGo(first.name, objRoute(first.kind, first.scope, first.name), 'mute', true, first.kind + '/' + first.scope + '/' + first.name)
    : chip('暂无在册对象', 'mute');
  return '<div class="w-svc" data-go="' + esc(rows.length ? objRoute(k, rows[0].scope, rows[0].name) : 'docs') + '" role="link">' +
    '<div class="ws-main"><div class="ws-title">' + esc(k) + '</div>' +
    '<div class="ws-desc">' + esc(rows.length ? '首个：' + (first ? first.scope + '/' + first.name : '—') : '该 kind 暂无对象') + '</div></div>' +
    '<span class="ws-chip">' + chip(String(rows.length) + ' 个', rows.length ? 'acc' : 'mute') + '</span>' +
    '</div>';
}

/** 当前循环卡（GET /v1/loop/current 现读——契约词形容错：字段缺席=不显示不编造）。 */
function loopCardHtml(data) {
  const d = data && typeof data === 'object' ? data : {};
  const fields = [
    ['循环', d.loop != null ? d.loop : d.id],
    ['相位', d.phase || d.state],
    ['负责', d.agent || d.route],
    ['开始于', d.startedAt ? (relAge(d.startedAt) || String(d.startedAt)) : (d.started_at ? (relAge(d.started_at) || String(d.started_at)) : null)],
    ['触发', d.trigger || d.reason],
  ].filter(([, v]) => v != null && v !== '');
  const chips = fields.map(([k, v]) => chip(k + ': ' + v, 'mute')).join('');
  return '<div class="card report-loop"><div class="card-b">' +
    '<div class="report-loop-top">' + chip('当前循环', 'acc') + chips +
    (d.eventId ? '<span class="report-loop-acts">' + chipGo('触发事件', 'event/' + encodeURIComponent(String(d.eventId)), 'mute') + '</span>' : '') +
    '</div></div></div>';
}

/** 循环卡注水（现读不缓存——失败=降级词形；tasks.js 报告屏同槽位复用）。 */
export async function hydrateLoop(el, ctx) {
  const box = el.querySelector('[data-loop-box]');
  if (!box) return;
  const r = await ctx.api.loopCurrent();
  if (!el.isConnected || box.getAttribute('data-loop-box') === 'stale') return;
  if (!r.ok) {
    box.innerHTML = '<div class="card report-loop" data-report-loop="degraded"><div class="card-b">' +
      '<div class="report-loop-top">' + chip('当前循环', 'mute') +
      '<span class="small faint">循环状态暂不可达（' + esc(String(r.error || r.status)) + '）——不冒充无循环。</span></div></div></div>';
    return;
  }
  if (!r.data || (!r.data.loop && !r.data.id && !r.data.phase)) {
    box.innerHTML = '<div class="card report-loop" data-report-loop="idle"><div class="card-b">' +
      '<div class="report-loop-top">' + chip('当前无进行中循环', 'mute') +
      '<span class="small faint">本周期暂无循环记录。点火下一轮：<code class="mono">corectl loop</code></span></div></div></div>';
    return;
  }
  box.innerHTML = loopCardHtml(r.data);
}

/** 首屏渲染（整屏重绘——store→render 单向数据流；store 变更由 main 触发重进）。 */
export function render(el, ctx) {
  const st = ctx.store;
  if (!st.isReady) {
    el.innerHTML = '<div class="view-inner welcome"><div class="w-hello"><h1>Welcome</h1></div>' +
      emptyBox('快照装载中…', '正在建立 /v1/snapshot 基线（连接建立即推全量）') + '</div>';
    return;
  }
  const objs = st.objects();
  const kinds = st.kinds();
  const kindsList = kinds.length
    ? kinds
    : Array.from(new Set(objs.map((o) => o.kind))).map((k) => ({ name: k }));
  const drafts = st.drafts();
  const events = st.events();
  const tasks = st.tasks();
  const inbox = st.inbox();
  const inboxN = inbox ? (Array.isArray(inbox.items) ? inbox.items.length : (Number(inbox.total) || 0)) : 0;
  const approvalN = drafts.length + events.filter((e) => e && e.type === 'ApprovalNeeded').length;
  const runningN = tasks.filter((t) => t && t.status && t.status.phase === 'running').length;

  const kindCards = kindsList.slice(0, 12).map((k) => kindCard(k, objs)).join('');
  const kindMore = kindsList.length > 12
    ? '<div class="small faint" style="grid-column:1/-1">…另有 ' + (kindsList.length - 12) + ' 个 kind——见对象工作台左侧清单。</div>'
    : '';

  el.innerHTML = '<div class="view-inner welcome">' +
    '<div class="w-hello"><h1>Welcome</h1><span class="w-date mono">' + esc(new Date().toISOString().slice(0, 10)) + '</span>' +
    '<span class="w-conn" data-conn-slot></span></div>' +
    '<div class="statbar">' +
    statCard('在册对象', objs.length, '全部 kind 合计', 'obj') +
    statCard('kinds', kindsList.length, 'scheme 注册', 'docs') +
    statCard('事件（窗口内）', events.length, '归档行', 'events') +
    statCard('待审', approvalN, approvalN ? '草稿∪审批事件——点击处理' : '待审区为空（空态显式）', 'approvals', approvalN ? 'alert' : '') +
    statCard('收件箱', inboxN, inboxN ? '待你处理' : '暂无待办', 'inbox', inboxN ? 'alert' : '') +
    statCard('进行中任务', runningN, '任务中心', 'tasks') +
    '</div>' +
    '<h2 class="sec-title">kinds 摘要</h2>' +
    (objs.length || kindsList.length
      ? '<div class="w-services">' + kindCards + kindMore + '</div>'
      : emptyBox('快照暂无对象枚举', '等待首个快照聚合或 patch 增量')) +
    '<h2 class="sec-title">当前循环</h2>' +
    '<div data-loop-box="pending">' + emptyBox('循环状态现读中…') + '</div>' +
    '<div class="w-foot small faint">数据面：GET /v1/snapshot 基线 + /v1/stream patch 增量（WS 断线自动降级轮询）；' +
    '对象正文按 resourceVersion 缓存（rv 未变不重读）。</div>' +
    '</div>';

  hydrateLoop(el, ctx);
}
