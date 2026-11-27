/* welcome.js —— 总览大屏（2026-11-26 用户裁决：总览=大屏；§19.3 就地批闸+点火接通路）：
 *   统计条 + 当前循环卡 + 三大区（正在跑的 Workflows / 最近需要的审批 / 核心 Agent）。
 *   #/service/:name 复用同布局投影 ServiceDesc 视角（后续 ServiceAgent 登录即此视图）。
 * §19.3 变更：
 *   · 审批区一等行源=挂起 runs（status=suspended ∨ phase=AwaitingApproval）——行尾
 *     「批准继续 / 驳回终止」经 POST /core-console/event-pub 发布 ApprovalGranted /
 *     ApprovalRejected（involved=Workflow/<scope>/<workflow> 三段词形）；批准=run 沿
 *     声明相位序列续行，驳回=run 终态 Failed（confirm 文案写明）。
 *   · 循环卡「触发事件」接 LoopRequested 点火通路（发布中 disabled；成功=chip 词形+3s
 *     后重跑 hydrateLoop；失败=显式错误词形）。
 *   · 核心 Agent 行下注会话活性（GET /v1/sessions，60s 模块级缓存——本文件自持）。
 *   · 收件箱统计卡在快照缺席（null/undefined）时整卡不渲染（不是显示「—」）。
 * 零自建状态——一切来自 store（/v1/snapshot 基线 + patch 增量）；循环卡=GET /v1/loop/current
 * 现读（缺席/失败=诚实降级词形，不冒充无循环）。 */

import { esc, relAge, fmtNum, trunc } from '../core/format.js';
import { chip, chipGo, emptyBox, objRoute } from '../render/chips.js';
import { heroComposerHtml, bindHeroComposer } from './home.js';
import { t as tr } from '../core/i18n.js';

/* ── F-05：任务相位读取兼容扁平词形（/v1/tasks 简表 t.phase 与详情 t.status.phase 双词形）── */
function taskPhaseOf(x) {
  const s = x && x.status;
  return String((s && (s.phase || s)) || (x && x.phase) || '');
}

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
    '<span class="report-loop-acts"><span data-ignite-note></span>' +
    '<button class="btn btn-secondary sm" data-ignite="1">' + esc(tr('触发事件')) + '</button></span>' +
    '</div></div></div>';
}

/* 点火按钮（LoopRequested → 桥接面拉起循环；§19.3 接 eventPub 通路）。 */
function bindIgnite(el, ctx) {
  const btn = el.querySelector('[data-ignite]');
  if (!btn) return;
  btn.addEventListener('click', async () => {
    if (btn.disabled) return;
    const note = el.querySelector('[data-ignite-note]');
    btn.disabled = true;
    btn.textContent = tr('点火中…');
    const r = await ctx.api.eventPub({
      type: 'LoopRequested',
      involved: 'ServiceDesc/core/core_system',
      dedupKey: 'console-ignite-' + Date.now(),
    });
    if (!el.isConnected) return;
    if (r && r.ok) {
      if (note) note.innerHTML = chip(tr('已注入 LoopRequested（queue 语义——桥接面拉起循环）'), 'ok');
      setTimeout(() => { if (el.isConnected) hydrateLoop(el, ctx); }, 3000);
    } else {
      btn.disabled = false;
      btn.textContent = tr('触发事件');
      if (note) note.innerHTML = chip(
        tr('点火失败（') + String((r && r.status) || '—') + ((r && r.error) ? ' ' + String(r.error) : '') + tr('）——可重试'),
        'danger');
    }
  });
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
  bindIgnite(el, ctx);
}

/* ── 区块标题 ── */
function zoneHead(label, route) {
  return '<div class="zone-h"><span class="zone-t">' + esc(label) + '</span>' +
    (route ? chipGo(tr('全部'), route, 'mute', false) : '') + '</div>';
}

/* ── 区一：正在跑的 Workflows（snapshot.runs 活体；空态显式；scope 收窄——F-10） ── */
function runningWorkflowsHtml(ctx, scope) {
  const runs = ctx.store.runs().filter((r) => {
    if (scope) {
      let hay = '';
      try { hay = JSON.stringify([r && r.scope, r && r.ownerScope, r && r.involved, r && r.route, r && r.workflow]); } catch { hay = ''; }
      if (!hay.includes(scope)) return false;
    }
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

/* ── 区二：最近需要的审批 ──
 * 一等行源=挂起 runs（status=suspended ∨ phase=AwaitingApproval）——行尾就地批闸按钮
 * （批准=发布 ApprovalGranted，run 沿声明相位序列续行；驳回=发布 ApprovalRejected，
 * run 终态 Failed——confirm 文案写明）。草稿 ∪ ApprovalNeeded/Escalate 归档行保留其后。 */

/* run → 审批 involved 的 owner scope（involved 词形 Workflow/<scope>/<name> 优先，缺=global）。 */
function runOwnerScope(r) {
  const inv = String((r && (r.involved || r.route || '')) || '');
  const m = /^Workflow\/([^/\s]+)\//.exec(inv);
  if (m) return m[1];
  return String((r && (r.scope || r.ownerScope)) || '') || 'global';
}

function pendingRunsOf(ctx) {
  return ctx.store.runs().filter((r) => r && (
    String(r.status || r.state || '') === 'suspended' ||
    String(r.phase || '') === 'AwaitingApproval'));
}

function pendingRunRowHtml(x) {
  const btnAttrs = (t) =>
    ' data-appr-' + t + '="1" data-appr-wf="' + esc(x.wf) + '" data-appr-scope="' + esc(x.scope) +
    '" data-appr-run="' + esc(x.run) + '" data-appr-key="' + esc(x.key) + '"';
  /* L-02：审批行「上标题+状态、下操作按钮」双行结构——窄列（1100px）不再硬塞
   * 5 个横向元素（core-workflow 断词/2h 折行/徽章挤压收口）。 */
  return '<div class="dash-row appr-row" data-appr-row>' +
    '<div class="dr-main">' + chip(tr('待批'), 'warn') +
    '<b class="mono">' + esc(trunc(x.wf, 22)) + '</b>' +
    '<span class="small faint">#' + esc(x.run) + '</span>' +
    chip(x.phase, 'warn') +
    (x.started ? '<span class="small faint">' + esc(x.started) + '</span>' : '') +
    '</div>' +
    '<div class="appr-acts">' +
    /* V-04（2026-11-26 概念可见性批）：待批行加「运行详情」→ wf/:name/r/:n（此前泛链 absent——
     * 人批准前看不到 run 全貌）。 */
    (x.wf && x.run != null && x.run !== ''
      ? '<button class="btn btn-ghost sm" data-go="wf/' + encodeURIComponent(String(x.wf)) + '/r/' + encodeURIComponent(String(x.run)) + '">' + esc(tr('运行详情')) + '</button>' : '') +
    '<button class="btn btn-chip sm"' + btnAttrs('grant') + '>' + esc(tr('批准继续')) + '</button>' +
    '<button class="btn btn-ghost btn-danger sm"' + btnAttrs('reject') + '>' + esc(tr('驳回终止')) + '</button>' +
    '</div></div>';
}

/* 就地批闸（confirm→eventPub→行消隐+快照重基；失败=alert 显式词形）。 */
function bindApprovalGates(el, ctx) {
  const act = async (btn, type) => {
    if (btn.disabled) return;
    const wf = btn.getAttribute('data-appr-wf') || '';
    const scope = btn.getAttribute('data-appr-scope') || 'global';
    const run = btn.getAttribute('data-appr-run') || '';
    const key = btn.getAttribute('data-appr-key') || run;
    const ask = type === 'ApprovalGranted'
      ? tr('确认批准继续？run 将沿 Workflow 声明相位序列续行：')
      : tr('确认驳回终止？run 将进入终态 Failed（不可恢复）：');
    if (!window.confirm(ask + wf + ' #' + run)) return;
    btn.disabled = true;
    const r = await ctx.api.eventPub({
      type,
      involved: 'Workflow/' + (scope || 'global') + '/' + wf,
      dedupKey: 'console-' + key,
    });
    if (r && r.ok) {
      const row = btn.closest('[data-appr-row]');
      // V-15：就地批闸后显示结果与已决台账指引（不静默消失）
      if (row) {
        const grant = type === 'ApprovalGranted';
        row.innerHTML = '<div class="dr-main">' +
          chip(grant ? tr('已批准（ApprovalGranted 已发布）') : tr('已驳回（ApprovalRejected 已发布）'), 'ok') +
          '<span class="small faint">' + esc(tr('判定已落审批台账——快照重基后行集刷新。')) + '</span></div>' +
          '<div class="dr-side">' + chipGo(tr('查看已决台账 →'), 'approvals', 'acc', true, tr('审批页已决台账')) + '</div>';
      }
      ctx.api.snapshot().then((s) => { if (s && s.ok) ctx.store.init(s.data); });
    } else {
      btn.disabled = false;
      const verb = type === 'ApprovalGranted' ? tr('批准失败') : tr('驳回失败');
      window.alert(verb + tr('（') + String((r && r.status) || '—') +
        ((r && r.error) ? ' ' + String(r.error) : '') + tr('）——run 仍在等审批，可重试。'));
    }
  };
  el.querySelectorAll('[data-appr-grant]').forEach((btn) =>
    btn.addEventListener('click', (ev) => { ev.stopPropagation(); act(btn, 'ApprovalGranted'); }));
  el.querySelectorAll('[data-appr-reject]').forEach((btn) =>
    btn.addEventListener('click', (ev) => { ev.stopPropagation(); act(btn, 'ApprovalRejected'); }));
}

function pendingApprovalsHtml(ctx, scope) {
  const runRows = pendingRunsOf(ctx).slice(0, 8).map((r) => {
    const wf = String(r.workflow || r.wf || '');
    if (!wf) return '';
    if (scope) {
      const sc = runOwnerScope(r);
      if (sc !== scope) return '';
    }
    const run = String(r.run != null ? r.run : '');
    const key = (r.eventId != null && r.eventId !== '') ? String(r.eventId) : run;
    return pendingRunRowHtml({
      wf, run, key,
      scope: runOwnerScope(r),
      phase: String(r.phase || r.state || 'AwaitingApproval'),
      started: r.startedAt ? (relAge(r.startedAt) || '') : '',
    });
  }).filter(Boolean);
  // L-03 修复随行：pendingRunRowHtml 原模板尾部多余 '</div></div>' 已在行内闭合——无残留标签。
  const drafts = ctx.store.drafts()
    .filter((d) => d && d.kind && d.name)
    .filter((d) => !scope || String(d.scope || 'global') === scope)
    .map((d) => ({ src: 'draft', t: d.updatedAt || d.ts || null,
      go: 'doc/' + encodeURIComponent(d.kind) + '/' + encodeURIComponent(d.scope || 'global') + '/' + encodeURIComponent(d.name),
      title: d.kind + '/' + (d.scope || 'global') + '/' + d.name, tag: tr('稿') }));
  const evs = ctx.store.events()
    .filter((e) => e && (e.type === 'ApprovalNeeded' || e.type === 'Escalate'))
    .filter((e) => !scope || String(e.involved || '').includes(scope))
    .slice(0, 8)
    .map((e) => ({ src: 'ev', t: e.ts || null, go: 'event/' + encodeURIComponent(String(e.id ?? '')),
      title: e.type + ' · ' + String(e.involved || ''), tag: e.type === 'ApprovalNeeded' ? tr('待审') : tr('升级') }));
  const rest = drafts.concat(evs)
    .sort((a, b) => String(b.t || '').localeCompare(String(a.t || '')))
    .slice(0, 8)
    .map((x) => '<div class="dash-row" data-go="' + esc(x.go) + '" role="link">' +
      '<div class="dr-main">' + chip(x.tag, x.src === 'draft' ? 'warn' : 'esc') +
      '<b class="mono">' + esc(trunc(x.title, 34)) + '</b></div>' +
      (x.t ? '<div class="dr-side"><span class="small faint">' + esc(relAge(x.t) || '') + '</span></div>' : '') +
      '</div>').join('');
  const rows = runRows.join('') + rest;
  if (!rows) return '<div class="card"><div class="ob-empty">' + esc(tr('待审区为空')) + '</div></div>';
  return '<div class="card">' + rows + '</div>';
}

/* ── 区三：核心 Agent（kind=Agent 在册对象 ∪ 会话活性——GET /v1/sessions 60s 缓存注水） ── */
let AG_SESS = { list: null, at: 0 };
const AG_SESS_TTL_MS = 60000;

/* 会话行集缓存（60s；失败不缓存失败态——返回 null，下一拍重读）。 */
async function sessionsCached(ctx) {
  if (AG_SESS.list && Date.now() - AG_SESS.at < AG_SESS_TTL_MS) return AG_SESS.list;
  const r = await ctx.api.sessions();
  if (r && r.ok && Array.isArray(r.data && r.data.sessions)) {
    AG_SESS = { list: r.data.sessions, at: Date.now() };
    return AG_SESS.list;
  }
  return null;
}

/* state → chip 色（escalated=esc；running=acc；suspected-stall=warn；其余 mute）。 */
function sessStateCls(st) {
  return st === 'escalated' ? 'esc' : (st === 'running' ? 'acc' : (st === 'suspected-stall' ? 'warn' : 'mute'));
}

/* 命中判据：sessionId 含 agent 名，或 involved/route/agent 词形含名。 */
function sessHitOf(list, name) {
  const byId = list.find((s) => String((s && s.sessionId) || '').includes(name));
  if (byId) return byId;
  return list.find((s) => {
    try { return JSON.stringify([s && s.involved, s && s.route, s && s.agent]).includes(name); } catch { return false; }
  });
}

function sessLiveHtml(hit) {
  if (!hit) return chip(tr('无会话'), 'mute');
  const st = String(hit.state || '');
  const sid = String(hit.sessionId || hit.id || '');
  const sub = [
    String(hit.phase || ''),
    trunc(String(hit.lastAction || ''), 18),
    (hit.lastBeat ? (relAge(hit.lastBeat) || '') : ''),
  ].filter(Boolean).join(' · ');
  const stChip = st
    ? (sid ? chipGo(st, 'session/' + encodeURIComponent(sid), sessStateCls(st), true, tr('会话详情——活性账现读')) : chip(st, sessStateCls(st)))
    : '';
  return stChip +
    (sub ? '<span class="small faint">' + esc(sub) + '</span>' : '');
}

/* Agent 行会话活性注水（现读缓存；失败=诚实降级词形）。 */
async function hydrateAgentLive(el, ctx) {
  const slots = el.querySelectorAll('[data-agent-live]');
  if (!slots.length) return;
  const list = await sessionsCached(ctx);
  if (!el.isConnected) return;
  if (!list) {
    slots.forEach((slot) => { slot.innerHTML = chip(tr('会话活性暂不可读'), 'mute'); });
    return;
  }
  slots.forEach((slot) => {
    const name = String(slot.getAttribute('data-agent-live') || '');
    if (!name) return;
    slot.innerHTML = sessLiveHtml(sessHitOf(list, name));
  });
}

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
    const sid = sess ? String(sess.sessionId || sess.id || '') : '';
    const liveChip = state
      ? (sid ? chipGo(state, 'session/' + encodeURIComponent(sid), state === 'escalated' || state === 'suspected-stall' ? 'esc' : 'acc', true, tr('会话详情')) : chip(state, state === 'escalated' || state === 'suspected-stall' ? 'esc' : 'acc'))
      : chip(tr('无会话'), 'mute');
    return '<div class="dash-row" data-go="' + esc('chat/' + encodeURIComponent(a.name)) + '" role="link">' +
      '<div class="dr-main"><b class="mono">' + esc(trunc(a.name, 30)) + '</b>' +
      '<span class="small faint">' + esc(a.scope) + '</span></div>' +
      '<div class="dr-side" data-agent-live="' + esc(a.name) + '">' + liveChip +
      '</div>' +
      '</div>';
  }).join('');
  if (!rows) return '<div class="card"><div class="ob-empty">' + esc(tr('暂无对象')) + '</div></div>';
  return '<div class="card">' + rows + '</div>';
}

/* ── 大屏骨架（系统视角与 Service 视角共用布局——同构换数据源） ── */
function dashLayout(ctx, opts) {
  const store = ctx.store;
  const scope = opts.scope || ''; // F-10：Service 视角 scope 收窄（空=系统视角全量）
  const inScope = (o) => !scope || String((o && (o.scope || (o.raw && o.raw.scope))) || '') === scope;
  /* 收件箱卡：快照缺席（null/undefined）=整卡不渲染（§19.3——不显示「—」像坏了）。 */
  const inbox = store.inbox();
  const inboxItems = inbox ? ((Array.isArray(inbox) ? inbox : ((inbox && inbox.items) || [])) || []) : [];
  const inboxScoped = scope ? inboxItems.filter((x) => !x || !x.scope || String(x.scope) === scope) : inboxItems;
  const inboxCard = inbox
    ? statCard(tr('收件箱'),
      inboxScoped.length,
      tr('待你处理'), 'inbox', '')
    : '';
  const stats =
    statCard(tr('在册对象'), (store.objects() || []).filter(inScope).length, tr('全部 kind 合计'), 'obj', '') +
    statCard(tr('事件'), (store.events() || []).filter((e) => !scope || String(e.involved || '').includes(scope)).length, tr('归档行'), 'events', '') +
    statCard(tr('待审'), (store.drafts() || []).filter((d) => !scope || String(d.scope || 'global') === scope).length + (store.events() || []).filter((e) => e && e.type === 'ApprovalNeeded' && (!scope || String(e.involved || '').includes(scope))).length, tr('待审明细'), 'approvals', 'warn') +
    inboxCard +
    statCard(tr('进行中任务'), (store.tasks() || []).filter((x) => ['running', 'verifying'].includes(taskPhaseOf(x))).filter(inScope).length, tr('任务中心'), 'tasks', '');
  return '<div class="view-inner">' + (opts.head || '') +
    (opts.composer ? '<div class="dash-composer">' + heroComposerHtml() + '</div>' : '') +
    '<div class="dash-stats' + (inboxCard ? '' : ' cols-4') + '">' + stats + '</div>' +
    '<div data-loop-box>' +
    '<div class="card report-loop"><div class="card-b"><div class="report-loop-top">' + chip(tr('当前循环'), 'mute') +
    tr('<span class="small faint">循环状态现读中…</span>') + '</div></div></div></div>' +
    '<div class="dash-grid">' +
    '<div class="dash-zone">' + zoneHead(tr('正在跑的 Workflows'), 'workflows') + runningWorkflowsHtml(ctx, scope) + '</div>' +
    '<div class="dash-zone">' + zoneHead(tr('最近需要的审批'), 'approvals') + pendingApprovalsHtml(ctx, scope) + '</div>' +
    '<div class="dash-zone">' + zoneHead(tr('核心 Agent'), 'tree') + agentsHtml(ctx, opts.coreAgents) + '</div>' +
    '</div>' +
    (opts.foot || '') +
    '</div>';
}

/* ── #/（总览大屏·系统视角） ── */
export function render(el, ctx) {
  el.innerHTML = dashLayout(ctx, {
    head: '<div class="dash-head"><h2 class="page-title">' + esc(tr('总览')) + '</h2>' +
      '<span class="small faint">' + esc(tr('系统视角大屏——正在跑的任务、最近需要的审批、核心 Agent 一眼可读')) + '</span></div>',
    composer: true,
  });
  bindHeroComposer(el, ctx);
  hydrateLoop(el, ctx);
  bindApprovalGates(el, ctx);
  hydrateAgentLive(el, ctx);
}

/* ── #/service/:name（Service 视角——同布局换数据源+scope 收窄；ServiceAgent 登录后即此视图） ──
 * F-10 修复：① 全部统计/区卡按 scope=<name> 过滤（dashLayout opts.scope）；② agents 解析
 * 不再读对象 refs 上不存在的 sd.content——走 docCache/api.object 现读 ServiceDesc 正文
 * （refs 行=快照聚合仅 kind/scope/name/rv，正文在 rv 缓存或现读）；读失败=诚实注记。 */
export function renderService(el, ctx) {
  const name = String(ctx.params.name || '');
  const sd = ctx.store.objects().find((o) => o.kind === 'ServiceDesc' && o.name === name);
  el.innerHTML = dashLayout(ctx, {
    scope: name,
    head: '<div class="dash-head"><h2 class="page-title">' + esc(tr('Service')) + ' · ' + esc(name) + '</h2>' +
      '<span class="small faint">' + esc(tr('Service 视角大屏（scope=' + name + ' 收窄投影——与系统大屏同构布局）')) + '</span></div>',
    coreAgents: null,
    foot: '<div class="card"><div class="card-b"><div class="small faint" data-sd-note>' +
      esc(tr('ServiceDesc 正文现读中…（声明 agents=agent://<name> 词形解析）')) +
      '</div></div></div>',
  });
  hydrateLoop(el, ctx);
  bindApprovalGates(el, ctx);
  hydrateAgentLive(el, ctx);
  // ServiceDesc 正文现读（缓存命中=零请求；rv 缓存语义与对象工作台同源）。
  (async () => {
    const note = el.querySelector('[data-sd-note]');
    if (!note) return;
    let content = '';
    const cached = sd && ctx.store.cachedDoc ? ctx.store.cachedDoc('ServiceDesc', (sd && sd.scope) || 'global', name, sd && sd.rv) : null;
    if (cached && typeof cached.content === 'string') {
      content = cached.content;
    } else if (ctx.api && typeof ctx.api.object === 'function') {
      const r = await ctx.api.object('ServiceDesc', (sd && sd.scope) || 'global', name);
      if (!el.isConnected) return;
      if (r && r.ok) {
        const data = r.data && typeof r.data === 'object' ? r.data : {};
        content = data.content != null ? String(data.content) : '';
        if (sd && ctx.store.cacheDoc) ctx.store.cacheDoc('ServiceDesc', (sd && sd.scope) || 'global', name, sd && sd.rv, content);
      }
    }
    const agents = [...content.matchAll(/agent:\/\/([A-Za-z0-9_.-]+)/g)].map((m) => m[1]);
    note.textContent = sd
      ? (content
        ? tr('ServiceDesc 在册——agents=') + (agents.join(', ') || '—')
        : tr('ServiceDesc 在册——正文暂不可读（agents 解析缺席，不冒充）'))
      : tr('ServiceDesc 对象不在册（词形投影自快照——不冒充在册）');
  })().catch(() => { /* 注记词形保持现读中——失败不装成功 */ });
}
