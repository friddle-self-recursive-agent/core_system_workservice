/* tasks.js —— 任务中心 + 任务详情 + 关注报告（词形层从旧 console-tasks.js 语义移植，
 * 代码全部重写为 ESM）。任务状态单一写者=调度器（Task.status）；本模块只做投影：
 *   ① 分组=确定性纯函数 groupTasks()（正在做/等人/卡住/完成/排队——人话词形固定）；
 *   ② 报告=attention（快照聚合）severity 排序 + 证据 ref 路由映射 evidenceRoute()
 *      （未知词形=不可点不冒充）+ 数据可信度注脚；
 *   ③ 数据纪律：零前端自建状态；读取失败=显式降级词形（零命中不装假）。 */

import { esc, relAge, trunc } from '../core/format.js';
import { confirmModal, toast } from './modal.js';
import { chip, chipGo, kv, emptyBox, screenHead, involvedRoute, objRoute, sessionTargetOf } from '../render/chips.js';
import { hydrateLoop } from './welcome.js';
import { t as tr } from '../core/i18n.js';

/* ---------- 相位人话词形（禁 run #N/session id 当主标题——标题取 spec.title/name） ---------- */
const PHASE_VIEW = {
  running: { label: tr('正在做'), chip: 'acc' },
  verifying: { label: tr('验证中'), chip: 'acc' },
  queued: { label: tr('排队中'), chip: 'mute' },
  ready: { label: tr('待启动'), chip: 'mute' },
  pending: { label: tr('等待拉起'), chip: 'mute' },
  blocked: { label: tr('受阻'), chip: 'warn' },
  failed: { label: tr('失败'), chip: 'esc' },
  done: { label: tr('已完成'), chip: 'ok' },
  canceled: { label: tr('已取消'), chip: 'mute' },
};
export function statusOf(t) {
  const s = (t && t.status) || {};
  // 扁平顶层字段兜底（F-05——/v1/tasks 简表与快照 tasks 行是扁平 phase/attempts/…投影，
  // 无 status 包裹）：t.status 缺席键时逐键回退顶层字段——两词形同屏兼容，不编造缺省值。
  const flat = (k) => (s[k] != null ? s[k] : (t ? t[k] : undefined));
  return {
    phase: flat('phase'), attempts: flat('attempts'),
    agentSession: flat('agentSession'), sessionId: flat('sessionId'),
    agent: flat('agent'), lastAction: flat('lastAction'), lastError: flat('lastError'),
    dependsOn: s.dependsOn != null ? s.dependsOn : (t && t.dependsOn),
  };
}
export function specOf(t) { return (t && t.spec) || {}; }

export function phaseChip(t) {
  const phase = String(statusOf(t).phase || '');
  const v = PHASE_VIEW[phase];
  if (!v) return chip(phase || tr('状态未知'), 'mute'); // 未知相位不冒充人话——原词如实显示
  return chip(v.label, v.chip);
}

/* ---------- 分组（确定性纯函数——机械直测口径） ----------
 * 正在做=running/verifying；排队=queued/ready 与未知相位；完成=done（canceled 同组徽如实）；
 * blocked/failed 再分：原因命中审批词形=等人，否则=卡住。 */
export function humanWait(t) {
  const s = statusOf(t);
  return /审批|批准|待审|等待人|approval|approve/i.test(String(s.lastError || '') + ' ' + String(s.lastAction || ''));
}
export function groupOf(t) {
  const phase = String(statusOf(t).phase || '');
  if (phase === 'running' || phase === 'verifying') return 'doing';
  if (phase === 'done' || phase === 'canceled') return 'done';
  if (phase === 'blocked' || phase === 'failed') return humanWait(t) ? 'waiting' : 'stuck';
  return 'queued';
}
export const GROUPS = [
  { id: 'doing', label: tr('正在做') },
  { id: 'waiting', label: tr('等人') },
  { id: 'stuck', label: tr('卡住') },
  { id: 'done', label: tr('完成') },
  { id: 'queued', label: tr('排队') },
];
export function groupTasks(tasks) {
  const out = { doing: [], waiting: [], stuck: [], done: [], queued: [] };
  (tasks || []).forEach((t) => { out[groupOf(t)].push(t); });
  return out;
}

/* ---------- 任务标识与深链（taskId='<scope>/<name>'——整段单次编码） ---------- */
export function taskIdOf(t) {
  if (!t) return '';
  if (t.taskId) return String(t.taskId);
  return (t.scope || 'global') + '/' + String(t.name || '');
}
/** 任务详情深链（与 ctx.go('task/'+encodeURIComponent(id)) 同词形；router 侧补 '#/' 前缀）。 */
export function taskHref(t) { return 'task/' + encodeURIComponent(taskIdOf(t)); }
export function taskTitle(t) {
  const spec = specOf(t);
  return String(spec.title || (t && t.title) || (t && t.name) || tr('未命名任务'));
}

/* ---------- 取消通路（§19.3）：发布 TaskCancelRequested——调度器归档面消费 ---------- */
const cancelReq = new Set(); // 本页会话内已发布取消请求的任务 id（内存反馈态——权威投影仍以快照为准）

/** 可取消判定：status 缺席（状态未知=调度器未写 status 或简表投影——诚实词形）或 phase ∈ queued/pending。 */
function cancellable(t) {
  const phase = String(statusOf(t).phase || '');
  return !phase || phase === 'queued' || phase === 'pending';
}

/** 取消按钮词形（行尾/详情通用）。已发布过=chip 反馈，按钮收起。 */
function cancelBtnHtml(t) {
  const id = taskIdOf(t);
  if (cancelReq.has(id)) return chip(tr('取消请求已发布'), 'warn');
  if (!cancellable(t)) return '';
  return '<button class="btn btn-ghost btn-danger sm" data-task-cancel="' + esc(id) + '">' + esc(tr('取消')) + '</button>';
}

/** 发布 TaskCancelRequested → 成功=行尾 chip + 快照重读（store 权威投影跟进）；失败=toast 显式词形
 * （2026-11-28 弹窗裁决——style.md §3.3：原生 confirm/alert 禁用）。 */
async function requestCancel(el, ctx, t) {
  const name = String((t && t.name) || '');
  const scope = String((t && t.scope) || 'global') || 'global';
  if (!name) { toast(esc(tr('任务缺 name——无法构造 Task/<scope>/<name> 词形，取消未发布')), 'error'); return; }
  const ok = await confirmModal({
    title: tr('取消任务'),
    body: esc(tr('取消任务 ') + name + tr('？发布 TaskCancelRequested（调度器归档面消费——不保证已在跑的进程立刻终止）')),
    confirmText: tr('取消任务'),
    danger: true,
  });
  if (!ok) return;
  try {
    const r = await ctx.api.eventPub({
      type: 'TaskCancelRequested',
      involved: 'Task/' + scope + '/' + name,
      dedupKey: 'console-cancel-' + name + '-' + Date.now(),
    });
    if (r && r.ok) {
      cancelReq.add(taskIdOf(t));
      ctx.api.snapshot().then((s) => { if (s && s.ok) ctx.store.init(s.data); }).catch(() => { /* 快照重读失败=保留现投影与 chip 反馈 */ });
      if (el.isConnected) render(el, ctx); // 本屏立即重绘——行尾出现「取消请求已发布」
    } else {
      toast(esc(tr('取消请求发布失败：') + String((r && (r.error || r.status)) || tr('网络异常'))), 'error');
    }
  } catch (err) {
    toast(esc(tr('取消请求发布失败：') + String((err && err.message) || err || tr('网络异常'))), 'error');
  }
}

/* ---------- 依赖深链（V-03）：Task/task/<scope>/<name>→task/:id；Workflow/workflow/<scope>/<name>→wf/:name；
 * 两段 <scope>/<name>→task/:id（Task dependsOn 简写）；裸 workflow 名若由 workflow 侧传入走 wf/:name。 */
export function depRoute(d, defaultKind = '') {
  const s = String(d || '').trim();
  if (!s) return '';
  const m = /^(Task|task|Workflow|workflow)\/([^/\s]+)\/([^/\s]+)$/.exec(s);
  if (m) {
    return m[1].toLowerCase() === 'task'
      ? 'task/' + encodeURIComponent(m[2] + '/' + m[3])
      : 'wf/' + encodeURIComponent(m[3]);
  }
  const m2 = /^([A-Za-z0-9_.-]+)\/([A-Za-z0-9_.-]+)$/.exec(s);
  if (m2) {
    if (defaultKind === 'workflow') return 'wf/' + encodeURIComponent(m2[2]);
    return 'task/' + encodeURIComponent(m2[1] + '/' + m2[2]);
  }
  if (defaultKind === 'workflow' && /^[A-Za-z0-9_.-]+$/.test(s)) return 'wf/' + encodeURIComponent(s);
  return '';
}

export function depChips(deps, defaultKind = '') {
  return (deps || []).map((d) => {
    const r = depRoute(d, defaultKind);
    return r ? chipGo(String(d), r, 'mute', true, tr('依赖——点击进详情')) : chip(String(d), 'mute');
  }).join(' ');
}

/** 行主区=真链接（R-UX 链接语义：中键新标签/复制深链/键盘 Tab+Enter/读屏器可达）——
 *  href '#/task/<encoded id>' 与详情路由同词形；动作区（对话 chip/取消按钮）置于 <a> 外，
 *  避免交互元素嵌套。取消按钮仍由 data-task-cancel 局部监听接管。 */
export function taskRow(t) {
  const id = taskIdOf(t);
  const s = statusOf(t);
  const spec = specOf(t);
  const depList = spec.dependsOn || s.dependsOn || [];
  const agent = String(s.agent || spec.agent || '').trim();
  const sess = String(s.agentSession || s.sessionId || '').trim();
  const last = String(s.lastAction || '').trim();
  const unknown = !String(s.phase || '');
  return '<div class="task-row">' +
    '<a class="task-main lk" href="#/' + taskHref(t) + '" title="' + esc(tr('进任务详情')) + '">' +
    '<div class="task-title"><b>' + esc(taskTitle(t)) + '</b>' + phaseChip(t) +
    chip(id, 'mute') +
    (depList.length ? depChips(depList.slice(0, 3), 'task') + (depList.length > 3 ? chip('+' + (depList.length - 3), 'mute') : '') : '') +
    (agent ? chip(tr('负责 ') + agent, 'mute') : '') + '</div>' +
    '<div class="task-sub small">' + esc(trunc(last, 100) || tr('最近动作：—')) +
    (String(s.lastError || '').trim() ? ' ' + chip(trunc(String(s.lastError), 60), 'esc') : '') +
    '</div>' +
    (unknown ? '<div class="task-sub small faint">' + esc(tr('状态未知（可取消或进详情核查）')) + '</div>' : '') +
    '</a>' +
    '<div class="task-acts">' + (sess ? chipGo(tr('对话'), 'session/' + encodeURIComponent(sess), 'mute', true, tr('关联会话')) : '') +
    cancelBtnHtml(t) + '</div>' +
    '</div>';
}

/** 任务中心（分组固定词形——正在做/等人/卡住/完成/排队）。 */
function tasksScreen(el, ctx) {
  const tasks = ctx.store.tasks();
  const g = groupTasks(tasks);
  const body = GROUPS.map((grp) => {
    const rows = g[grp.id] || [];
    return '<div class="card"><div class="card-h"><span class="h-title">' + esc(grp.label) + '</span>' + chip(String(rows.length) + tr(' 项'), rows.length ? 'acc' : 'mute') + '</div>' +
      (rows.length ? '<div class="card-b">' + rows.map(taskRow).join('') + '</div>' : '<div class="card-b small faint">' + esc(grp.label) + tr('：暂无。</div>')) +
      '</div>';
  }).join('');
  el.innerHTML = '<div class="view-inner">' +
    screenHead(tr('任务中心'), tr('分组=确定性投影（状态单一写者=调度器）；点击行进详情'), chipGo(tr('关注报告'), 'report', 'acc', false)) +
    '<div class="grid cols-2">' + body + '</div></div>';

  // 行进详情=行主区真 <a href="#/task/...">（中键/复制深链/键盘/读屏原生可达）——
  // 无需 JS 行点击绑定；行内 data-go chip 由全仓委托接管，取消按钮由下方局部监听接管。

  // 取消按钮：stopPropagation——不触发行进详情；发布 TaskCancelRequested（§19.3 取消通路）
  const byId = new Map(tasks.map((x) => [taskIdOf(x), x]));
  el.querySelectorAll('[data-task-cancel]').forEach((btn) => {
    btn.addEventListener('click', (ev) => {
      ev.stopPropagation();
      const t = byId.get(btn.dataset.taskCancel);
      if (t) requestCancel(el, ctx, t);
    });
  });
}

/* ---------- 最近动作结构化卡（2026-11-28 用户裁定「最近动作没有结构化。暂时为空」）----------
 * 生命周期时间行：类型+相位词（TaskStateChanged dedupKey=task:<id>:<phase>——taskId 不含
 * ':'，lastIndexOf 取尾段即相位词；其余类型无相位词不冒充）+级别点+时刻。 */
function taskPhaseOf(e) {
  const dk = String((e && e.dedupKey) || '');
  if (!/^task:/.test(dk)) return '';
  return dk.slice(dk.lastIndexOf(':') + 1);
}
export { taskPhaseOf };

function taskEvRow(e) {
  const phase = taskPhaseOf(e);
  const pv = PHASE_VIEW[phase];
  const ts = e && (e.ts || e.at || e.time);
  return '<div class="mini-row">' +
    '<span class="w-ev-dot" data-level="' + esc(String(e && e.level || 'info')) + '"></span>' +
    '<div class="mr-main"><div>' + esc(String(e && e.type || tr('未命名事件'))) +
      (pv ? ' ' + chip(pv.label, pv.chip) : (phase ? ' ' + chip(phase, 'mute') : '')) + '</div>' +
      '<div class="mr-sub small mono">' + esc(trunc(String((e && e.dedupKey) || (e && e.eventId) || ''), 64)) + '</div></div>' +
    '<span class="w-ev-time mono" title="' + esc(String(ts || '')) + '">' + esc(relAge(ts) || '—') + '</span>' +
    '</div>';
}

/** 生命周期时间线现读（GET /v1/events?involved=Task/<id>——归档权威直查，非快照 50 行窗；
 * 失败/空=显式词形不装内容）。 */
async function hydrateTaskTimeline(box, ctx, id) {
  const r = await ctx.api.events(60, { involved: 'Task/' + id });
  if (!box.isConnected) return; // 已切屏=弃投（不覆写他屏 DOM）
  if (!r || !r.ok) {
    box.innerHTML = '<div class="small faint">' + esc(tr('时间线拉取失败：')) +
      esc(String((r && (r.error || r.status)) || tr('网络异常'))) + '</div>';
    return;
  }
  const d = (r && r.data) || {};
  const rows = Array.isArray(d) ? d : (Array.isArray(d.events) ? d.events : []);
  box.innerHTML = rows.length
    ? rows.map(taskEvRow).join('')
    : '<div class="small faint">' + esc(tr('归档窗内暂无本任务事件。')) + '</div>';
}

/** 任务详情（store 在册投影；不在册=诚实空态——URL 手输名不冒充）。 */
function taskDetailScreen(el, ctx, id) {
  const tasks = ctx.store.tasks();
  const t = tasks.find((x) => taskIdOf(x) === String(id));
  const head = screenHead(tr('任务详情'), 'id=' + String(id), chipGo(tr('← 任务中心'), 'tasks', 'mute', false));
  if (!t) {
    el.innerHTML = '<div class="view-inner">' + head + emptyBox(tr('任务不在册'), tr('暂无此任务（可能已完成归档或 id 有误）。')) + '</div>';
    return;
  }
  const s = statusOf(t);
  const spec = specOf(t);
  const rows = [
    [tr('相位'), phaseChip(t)],
    [tr('标题'), esc(taskTitle(t))],
    ['id', esc(taskIdOf(t))],
  ];
  // #16 双向可达：负责 agent → Agent 对象页（对象页 spec.agent 反查任务详情的另一半）。
  if (spec.agent || s.agent) {
    const ag = String(s.agent || spec.agent).trim();
    const seg = ag.split('/');
    const scope = seg.length > 1 ? seg[0] : 'global';
    const name = seg.length > 1 ? seg.slice(1).join('/') : ag;
    rows.push([tr('负责'), chipGo(ag, objRoute('Agent', scope, name), 'mute', true, tr('打开 Agent 对象页'))]);
  }
  const sessId = String(s.agentSession || s.sessionId || '').trim();
  if (sessId) {
    const sessTarget = sessionTargetOf(sessId); // #16 任务键 hint 不冒充 session 深链
    rows.push([tr('会话'), chipGo(trunc(sessId, 24), sessTarget, 'acc', true, tr('关联会话'))]);
  }
  if (!String(s.phase || '')) rows.push([tr('注记'), esc(tr('状态未知（可取消或核查调度日志）'))]);
  const deps = spec.dependsOn || s.dependsOn || [];
  if (deps.length) rows.push([tr('依赖'), depChips(deps, 'task')]);
  if (cancelReq.has(taskIdOf(t))) rows.push([tr('取消'), chip(tr('取消请求已发布'), 'warn')]);
  // 最近动作结构化卡三源：①结果注记（lastAction 全文 wrap + lastError）②关联会话现势
  // （快照心跳行——不在册=诚实注记）③生命周期时间线（事件归档 involved 直查——异步注水）。
  const lastA = String(s.lastAction || '').trim();
  const lastE = String(s.lastError || '').trim();
  const sessLive = sessId ? (ctx.store.sessions() || []).find((x) => String((x && x.sessionId) || '') === sessId) : null;
  const sessLiveHtml = sessId ?
    '<div class="mini-row">' +
      '<span class="w-ev-dot" data-level="' + esc(/escalated|suspected-stall/.test(String(sessLive && sessLive.state)) ? 'error' : 'info') + '"></span>' +
      '<div class="mr-main"><div>' + esc(tr('关联会话现势')) + ' ' +
        (sessLive
          ? chip(String(sessLive.state || tr('在册')), /escalated|suspected-stall/.test(String(sessLive.state)) ? 'esc' : 'acc')
          : chip(tr('不在心跳在册'), 'mute')) +
        (sessLive && sessLive.phase ? chip(String(sessLive.phase), 'mute') : '') +
        (sessLive && sessLive.lastAction ? '<span class="small"> ' + esc(trunc(String(sessLive.lastAction), 80)) + '</span>' : '') +
      '</div>' +
      '<div class="mr-sub small mono">' + esc(trunc(sessId, 48)) +
        (sessLive && sessLive.lastBeat ? ' · ' + esc(tr('最近心跳 ')) + esc(relAge(sessLive.lastBeat)) : '') + '</div></div>' +
      '<span class="w-ev-time">' + chipGo(tr('对话'), 'session/' + encodeURIComponent(sessId), 'mute', true, tr('关联会话')) + '</span>' +
    '</div>' : '';
  const resultHtml = (lastA || lastE) ?
    '<div style="margin:8px 0">' +
      (lastA ? '<div class="small" style="white-space:pre-wrap;word-break:break-word">' +
        '<b>' + esc(tr('结果注记')) + '</b> <span title="' + esc(lastA) + '">' + esc(trunc(lastA, 4000)) + '</span></div>' : '') +
      (lastE ? '<div style="margin-top:6px">' + chip(trunc(lastE, 160), 'esc') + '</div>' : '') +
    '</div>' : '';
  el.innerHTML = '<div class="view-inner">' + head + '<div class="card"><div class="card-b">' +
    kv(rows) +
    (cancellable(t) && !cancelReq.has(taskIdOf(t))
      ? '<div style="margin-top:10px"><button class="btn btn-ghost btn-danger sm" data-task-cancel="' + esc(taskIdOf(t)) + '">' + esc(tr('取消')) + '</button>' +
        '<span class="small faint" style="margin-left:8px">' + esc(tr('发布 TaskCancelRequested——调度器归档面消费')) + '</span></div>'
      : '') +
    '</div></div>' +
    '<div class="card" style="margin-top:12px"><div class="card-h"><span class="h-title">' + esc(tr('最近动作')) + '</span>' +
      chip(tr('结果 · 会话现势 · 生命周期时间线'), 'mute') + '</div><div class="card-b">' +
      resultHtml + sessLiveHtml +
      '<div class="small faint" style="margin-top:6px">' + esc(tr('生命周期（事件归档现读）')) + '</div>' +
      '<div data-task-tl="' + esc(taskIdOf(t)) + '"><div class="small faint">' + esc(tr('现读中…')) + '</div></div>' +
    '</div></div></div>';
  el.querySelectorAll('[data-task-cancel]').forEach((btn) => {
    btn.addEventListener('click', () => requestCancel(el, ctx, t));
  });
  const tlBox = el.querySelector('[data-task-tl]');
  if (tlBox) hydrateTaskTimeline(tlBox, ctx, taskIdOf(t));
}

/* ---------- 关注报告（R-UX-6：attention 排序 + 证据路由 + 循环卡 + 可信度注脚） ---------- */
const SEV_VIEW = { critical: { label: tr('严重'), chip: 'esc' }, high: { label: tr('重要'), chip: 'warn' }, medium: { label: tr('一般'), chip: 'mute' } };
function sevView(sev) { return SEV_VIEW[String(sev || '')] || { label: String(sev || tr('未分级')), chip: 'mute' }; }
function sevRank(sev) { const s = String(sev || ''); return s === 'critical' ? 0 : s === 'high' ? 1 : s === 'medium' ? 2 : 3; }
export function sortAttention(list) { return (list || []).slice().sort((a, b) => sevRank(a && a.severity) - sevRank(b && b.severity)); }

/** 证据 ref → 路由（已知词形可点；未知=纯文本不冒充）。 */
export function evidenceRoute(ref) {
  const r = String(ref || '').trim();
  if (!r) return '';
  if (r.startsWith('#/')) {
    const segs = r.slice(2).split('/').filter(Boolean);
    if (!segs.length) return '';
    if (segs[0] === 'inbox') return 'inbox';
    if (segs[0] === 'approvals') return 'approvals';
    if (segs[0] === 'tasks') return 'tasks';
    if (['s', 'wf'].includes(segs[0]) && segs.length >= 2) return segs.map((x) => encodeURIComponent(x)).join('/');
    if (['event', 'task', 'session'].includes(segs[0]) && segs.length >= 2) return segs[0] + '/' + encodeURIComponent(segs.slice(1).join('/'));
    if (segs[0] === 'obj' && segs.length >= 4) return segs.map((x) => encodeURIComponent(x)).join('/');
    return '';
  }
  return involvedRoute(r);
}

function attentionCard(a) {
  if (!a) return '';
  const sev = sevView(a.severity);
  const evs = (a.evidence || []).map((ev) => {
    const ref = String(ev && ev.ref || '');
    const route = evidenceRoute(ref);
    return route ? chipGo(ref, route, 'mute', true, tr('证据——点击进入')) : chip(ref || (ev && ev.kind) || tr('证据'), 'mute');
  }).filter(Boolean);
  const action = a.action || {};
  const route = action.target ? String(action.target).replace(/^#\/?/, '') : '';
  return '<div class="att-card sev-' + esc(String(a.severity || 'unknown')) + '">' +
    '<div class="att-h">' + chip(sev.label, sev.chip) +
    (a.category ? chip(a.category, 'mute') : '') +
    '<b>' + esc(String(a.title || tr('未命名关注项'))) + '</b></div>' +
    (a.why ? '<div class="att-why small">' + esc(String(a.why)) + '</div>' : '') +
    '<div class="att-ev">' + (evs.length ? evs.join(' ') : tr('<span class="small faint">证据：—</span>')) + '</div>' +
    (route ? '<div class="att-acts"><button class="btn btn-secondary sm" data-go="' + esc(route) + '">' + esc(String(action.label || tr('去处理'))) + '</button></div>' : '') +
    '</div>';
}

/** 数据可信度注脚（generatedBy/window/missing——缺席数据显式标注）。 */
export function reportFooter(body) {
  const b = body || {};
  const dq = b.dataQuality || {};
  const missing = dq.missing;
  const missText = Array.isArray(missing) ? missing.map((m) => String(m)).join('、') : String(missing || '');
  const gb = String(b.generatedBy || '');
  const gbLabel = gb === 'llm' ? tr('智能归纳') : gb === 'rules' ? tr('基础汇总') : (gb || tr('生成方式未注明'));
  return tr('<div class="report-foot small faint">报告说明：') + esc(gbLabel) +
    tr(' · 检查范围 ') + esc(String(b.scope || 'global')) +
    (b.asOf ? tr(' · 截至 ') + esc(String(b.asOf)) : '') +
    (missText ? tr(' · 缺席数据：') + esc(missText) : '') +
    tr(' · 报告为派生视图可随时重建，不落对象账。</div>');
}

/** 报告屏（attention=快照聚合；缺席=ApprovalNeeded/Escalate 事件兜底 + 诚实注记）。 */
function reportScreen(el, ctx) {
  const st = ctx.store;
  let att = st.attention();
  let degradedNote = '';
  if (!att.length) {
    const evs = st.events().filter((e) => e && (e.type === 'ApprovalNeeded' || e.type === 'Escalate'));
    att = evs.map((e) => ({
      severity: e.type === 'Escalate' ? 'high' : 'medium',
      category: e.type,
      title: String(e.summary || e.type),
      evidence: e.involved ? [{ ref: String(e.involved) }] : [],
    }));
    if (!evs.length) degradedNote = tr('<div class="small faint" style="margin-bottom:10px">当前没有需要你关注的事项。</div>');
    else degradedNote = tr('<div class="small faint" style="margin-bottom:10px">以下事项来自最近事件。</div>');
  }
  const sorted = sortAttention(att);
  el.innerHTML = '<div class="view-inner">' +
    screenHead(tr('关注报告'), tr('severity 排序 · 证据可点 · 数据可信度显式'), chipGo(tr('任务中心'), 'tasks', 'mute', false)) +
    degradedNote +
    '<div data-loop-box="pending"></div>' +
    (sorted.length
      ? '<div class="att-list">' + sorted.map(attentionCard).join('') + '</div>'
      : emptyBox(tr('本窗口无关注项'), tr('当前没有需要关注的风险项。'))) +
    reportFooter(st.get('attention') && st.get('attention').dataQuality) +
    '</div>';
  // 循环卡注水（复用 welcome 的现读通路——data-loop-box 同槽位约定）
  hydrateLoop(el, ctx);
}

export function render(el, ctx) {
  if (ctx.params && ctx.params.id != null && ctx.params.id !== '') taskDetailScreen(el, ctx, String(ctx.params.id));
  else tasksScreen(el, ctx);
}
export function renderReport(el, ctx) { reportScreen(el, ctx); }
