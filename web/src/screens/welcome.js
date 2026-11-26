/* welcome.js —— 总览大屏（2026-11-26 用户裁决：总览=大屏）：
 *   统计条 + 当前循环卡 + 三大区（正在跑的 Workflows / 最近需要的审批 / 核心 Agent）。
 *   #/service/:name 复用同布局投影 ServiceDesc 视角（后续 ServiceAgent 登录即此视图）。
 * 零自建状态——一切来自 store（/v1/snapshot 基线 + patch 增量）；循环卡=GET /v1/loop/current
 * 现读（缺席/失败=诚实降级词形，不冒充无循环）。 */

import { esc, relAge, fmtNum, trunc } from '../core/format.js';
import { chip, chipGo, emptyBox, objRoute } from '../render/chips.js';
import { t as tr } from '../core/i18n.js';

/* ── 摘要统计卡（点击深链） ── */
function statCard(label, num, sub, route, cls) {
  return '<div class="stat ' + esc(cls || '') + '" data-go="' + esc(route) + '" role="link">' +
    '<div class="st-num">' + esc(fmtNum(num)) + '</div>' +
    '<div class="st-label">' + esc(label) + '<span class="small faint">' + esc(sub || '') + '</span></div>' +
    '</div>';
}

/* ── 当前循环卡（GET /v1/loop/current 现读——字段缺席=不显示不编造；tasks.js 同槽复用） ── */
function loopCardHtml(data) {
  const d = data && typeof data === 'object' ? data : {};
  const fields = [
    [tr('循环'), d.loop != null ? d.loop : d.id],
    [tr('相位'), d.phase || d.state],
    [tr('负责'), d.agent || d.route],
    [tr('开始于'), d.startedAt ? (relAge(d.startedAt) || String(d.startedAt)) : (d.started_at ? (relAge(d.started_at) || String(d.started_at)) : null)],
    [tr('触发'), d.trigger || d.reason],
  ].filter(([, v]) => v != null && v !== '');
  const chips = fields.map(([k, v]) => chip(k + ': ' + v, 'mute')).join('');
  return '<div class="card report-loop"><div class="card-b">' +
    '<div class="report-loop-top">' + chip(tr('当前循环'), 'acc') + chips +
    (d.eventId ? '<span class="report-loop-acts">' + chipGo(tr('触发事件'), 'event/' + encodeURIComponent(String(d.eventId)), 'mute') + '</span>' : '') +
    '</div></div></div>';
}

/* 循环卡注水（现读不缓存——失败=降级词形；tasks.js 报告屏同槽位复用）。 */
export async function hydrateLoop(el, ctx) {
  const box = el.querySelector('[data-loop-box]');
  if (!box) return;
  const r = await ctx.api.loopCurrent();
  if (!el.isConnected || box.getAttribute('data-loop-box') === 'stale') return;
  if (!r.ok) {
    box.innerHTML = '<div class="card report-loop" data-report-loop="degraded"><div class="card-b">' +
      '<div class="report-loop-top">' + chip(tr('当前循环'), 'mute') +
      tr('<span class="small faint">循环状态暂不可达（') + esc(String(r.error || r.status)) + tr('）——不冒充无循环。</span></div></div></div>');
    return;
  }
  if (!r.data || (!r.data.loop && !r.data.id && !r.data.phase)) {
    box.innerHTML = '<div class="card report-loop" data-report-loop="idle"><div class="card-b">' +
      '<div class="report-loop-top">' + chip(tr('当前无进行中循环'), 'mute') +
      tr('<span class="small faint">本周期暂无循环记录。点火下一轮：<code class="mono">corectl loop</code></span></div></div></div>');
    return;
  }
  box.innerHTML = loopCardHtml(r.data);
}

/* ── 区块标题 ── */
function zoneHead(label, route) {
  return '<div class="zone-h"><span class="zone-t">' + esc(label) + '</span>' +
    (route ? chipGo(tr('全部'), route, 'mute', false) : '') + '</div>';
}

/* ── 区一：正在跑的 Workflows（snapshot.runs 活体；空态显式） ── */
function runningWorkflowsHtml(ctx) {
  const runs = ctx.store.runs().filter((r) => {
    const st = String((r && (r.status || r.state)) || '');
    return st && st !== 'done' && st !== 'failed' && st !== 'canceled';
  }).slice(0, 8);
  const rows = runs.map((r) => {
    const wf = String(r.workflow || r.wf || '—');
    const phase = String(r.phase || r.state || '—');
    const st = String(r.status || '');
    const started = r.startedAt ? (relAge(r.startedAt) || '') : '';
    const route = r.workflow ? 'wf/' + encodeURIComponent(wf) + '/r/' + encodeURIComponent(String(r.run != null ? r.run : '')) : 'workflows';
    return '<div class="dash-row" data-go="' + esc(route) + '" role="link">' +
      '<div class="dr-main"><b class="mono">' + esc(trunc(wf, 26)) + '</b>' +
      '<span class="small faint">#' + esc(String(r.run != null ? r.run : '—')) + '</span></div>' +
      '<div class="dr-side">' + chip(phase, st === 'running' ? 'acc' : 'mute') +
      (started ? '<span class="small faint">' + esc(started) + '</span>' : '') + '</div>' +
      '</div>';
  }).join('');
  return '<div class="card">' + (rows || '<div class="ob-empty">' + esc(tr('当前无进行中循环')) + '</div>') + '</div>';
}

/* ── 区二：最近需要的审批（草稿缓存 ∪ ApprovalNeeded/Escalate 归档行；最近 8 条） ── */
function pendingApprovalsHtml(ctx) {
  const drafts = ctx.store.drafts()
    .filter((d) => d && d.kind && d.name)
    .map((d) => ({ src: 'draft', t: d.updatedAt || d.ts || null,
      go: 'doc/' + encodeURIComponent(d.kind) + '/' + encodeURIComponent(d.scope || 'global') + '/' + encodeURIComponent(d.name),
      title: d.kind + '/' + (d.scope || 'global') + '/' + d.name, tag: tr('稿') }));
  const evs = ctx.store.events()
    .filter((e) => e && (e.type === 'ApprovalNeeded' || e.type === 'Escalate'))
    .slice(0, 8)
    .map((e) => ({ src: 'ev', t: e.ts || null, go: 'event/' + encodeURIComponent(String(e.id ?? '')),
      title: e.type + ' · ' + String(e.involved || ''), tag: e.type === 'ApprovalNeeded' ? tr('待审') : tr('升级') }));
  const rows = drafts.concat(evs)
    .sort((a, b) => String(b.t || '').localeCompare(String(a.t || '')))
    .slice(0, 8)
    .map((x) => '<div class="dash-row" data-go="' + esc(x.go) + '" role="link">' +
      '<div class="dr-main">' + chip(x.tag, x.src === 'draft' ? 'warn' : 'esc') +
      '<b class="mono">' + esc(trunc(x.title, 34)) + '</b></div>' +
      (x.t ? '<div class="dr-side"><span class="small faint">' + esc(relAge(x.t) || '') + '</span></div>' : '') +
      '</div>').join('');
  if (!rows) return '<div class="card"><div class="ob-empty">' + esc(tr('待审区为空（空态显式）')) + '</div></div>';
  return '<div class="card">' + rows + '</div>';
}

/* ── 区三：核心 Agent（kind=Agent 在册对象 ∪ 会话活性状态） ── */
function agentsHtml(ctx, filterAgents) {
  const objs = ctx.store.objects();
  let agents = objs.filter((o) => o.kind === 'Agent')
    .map((o) => ({ name: o.name, scope: o.scope || 'global', route: objRoute('Agent', o.scope || 'global', o.name) }));
  if (filterAgents && filterAgents.length) agents = agents.filter((a) => filterAgents.includes(a.name));
  const sessions = ctx.store.sessions();
  const rows = agents.map((a) => {
    const sess = sessions.find((s) => String((s && (s.agent || s.route)) || '').includes(a.name)) ||
      sessions.find((s) => String((s && s.id) || '') === a.name);
    const state = sess ? String(sess.state || sess.status || '') : '';
    return '<div class="dash-row" data-go="' + esc('chat/' + encodeURIComponent(a.name)) + '" role="link">' +
      '<div class="dr-main"><b class="mono">' + esc(trunc(a.name, 30)) + '</b>' +
      '<span class="small faint">' + esc(a.scope) + '</span></div>' +
      '<div class="dr-side">' + (state ? chip(state, state === 'escalated' || state === 'suspected-stall' ? 'esc' : 'acc') : chip(tr('无会话'), 'mute')) + '</div>' +
      '</div>';
  }).join('');
  if (!rows) return '<div class="card"><div class="ob-empty">' + esc(tr('快照暂无对象枚举')) + '</div></div>';
  return '<div class="card">' + rows + '</div>';
}

/* ── 大屏骨架（系统视角与 Service 视角共用布局——同构换数据源） ── */
function dashLayout(ctx, opts) {
  const store = ctx.store;
  const stats =
    statCard(tr('在册对象'), (store.objects() || []).length, tr('全部 kind 合计'), 'obj', '') +
    statCard(tr('事件'), (store.events() || []).length, tr('归档行'), 'events', '') +
    statCard(tr('待审'), (store.drafts() || []).length + (store.events() || []).filter((e) => e && e.type === 'ApprovalNeeded').length, tr('稿∪待审事件'), 'approvals', 'warn') +
    statCard(tr('收件箱'), (store.inbox() || []).length, tr('待你处理'), 'inbox', '') +
    statCard(tr('进行中任务'), (store.tasks() || []).filter((x) => ['running', 'verifying'].includes(String((x && x.status && (x.status.phase || x.status)) || ''))).length, tr('任务中心'), 'tasks', '');
  return '<div class="view-inner">' + (opts.head || '') +
    '<div class="dash-stats">' + stats + '</div>' +
    '<div data-loop-box>' +
    '<div class="card report-loop"><div class="card-b"><div class="report-loop-top">' + chip(tr('当前循环'), 'mute') +
    tr('<span class="small faint">循环状态现读中…</span>') + '</div></div></div></div>' +
    '<div class="dash-grid">' +
    '<div class="dash-zone">' + zoneHead(tr('正在跑的 Workflows'), 'workflows') + runningWorkflowsHtml(ctx) + '</div>' +
    '<div class="dash-zone">' + zoneHead(tr('最近需要的审批'), 'approvals') + pendingApprovalsHtml(ctx) + '</div>' +
    '<div class="dash-zone">' + zoneHead(tr('核心 Agent'), 'tree') + agentsHtml(ctx, opts.coreAgents) + '</div>' +
    '</div>' +
    (opts.foot || '') +
    '<div class="w-foot small faint">' +
    tr('<div class="w-foot small faint">数据面：GET /v1/snapshot 基线 + /v1/stream patch 增量（WS 断线自动降级轮询）；') +
    tr('（rv 未变不重读正文——降级轮询路径同样生效）</span>') +
    '</div>';
}

/* ── #/（总览大屏·系统视角） ── */
export function render(el, ctx) {
  el.innerHTML = dashLayout(ctx, {
    head: '<div class="dash-head"><h2 class="page-title">' + esc(tr('总览')) + '</h2>' +
      '<span class="small faint">' + esc(tr('系统视角大屏——正在跑的任务、最近需要的审批、核心 Agent 一眼可读')) + '</span></div>',
  });
  hydrateLoop(el, ctx);
}

/* ── #/service/:name（Service 视角——同布局换数据源；ServiceAgent 登录后即此视图） ── */
export function renderService(el, ctx) {
  const name = String(ctx.params.name || '');
  const sd = ctx.store.objects().find((o) => o.kind === 'ServiceDesc' && o.name === name);
  let agents = [];
  if (sd && typeof sd.content === 'string') {
    agents = [...String(sd.content).matchAll(/agent:\/\/([A-Za-z0-9_.-]+)/g)].map((m) => m[1]);
  }
  el.innerHTML = dashLayout(ctx, {
    head: '<div class="dash-head"><h2 class="page-title">' + esc(tr('Service')) + ' · ' + esc(name) + '</h2>' +
      '<span class="small faint">' + esc(tr('Service 视角大屏（与系统大屏同构布局——后续 ServiceAgent 登录即落此视图）')) + '</span></div>',
    coreAgents: agents.length ? agents : null,
    foot: '<div class="card"><div class="card-b"><div class="small faint">' +
      (sd
        ? esc(tr('ServiceDesc 在册——agents=')) + esc(agents.join(', ') || '—')
        : esc(tr('ServiceDesc 对象不在册（词形投影自快照——不冒充在册）'))) +
      '</div></div></div>',
  });
  hydrateLoop(el, ctx);
}
