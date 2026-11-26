/* tasks.js —— 任务中心 + 任务详情 + 关注报告（词形层从旧 console-tasks.js 语义移植，
 * 代码全部重写为 ESM）。任务状态单一写者=调度器（Task.status）；本模块只做投影：
 *   ① 分组=确定性纯函数 groupTasks()（正在做/等人/卡住/完成/排队——人话词形固定）；
 *   ② 报告=attention（快照聚合）severity 排序 + 证据 ref 路由映射 evidenceRoute()
 *      （未知词形=不可点不冒充）+ 数据可信度注脚；
 *   ③ 数据纪律：零前端自建状态；读取失败=显式降级词形（零命中不装假）。 */

import { esc, trunc } from '../core/format.js';
import { chip, chipGo, kv, emptyBox, screenHead, involvedRoute } from '../render/chips.js';
import { hydrateLoop } from './welcome.js';

/* ---------- 相位人话词形（禁 run #N/session id 当主标题——标题取 spec.title/name） ---------- */
const PHASE_VIEW = {
  running: { label: '正在做', chip: 'acc' },
  verifying: { label: '验证中', chip: 'acc' },
  queued: { label: '排队中', chip: 'mute' },
  ready: { label: '待启动', chip: 'mute' },
  blocked: { label: '受阻', chip: 'warn' },
  failed: { label: '失败', chip: 'esc' },
  done: { label: '已完成', chip: 'ok' },
  canceled: { label: '已取消', chip: 'mute' },
};
export function statusOf(t) { return (t && t.status) || {}; }
export function specOf(t) { return (t && t.spec) || {}; }

export function phaseChip(t) {
  const phase = String(statusOf(t).phase || '');
  const v = PHASE_VIEW[phase];
  if (!v) return chip(phase || '状态未知', 'mute'); // 未知相位不冒充人话——原词如实显示
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
  { id: 'doing', label: '正在做' },
  { id: 'waiting', label: '等人' },
  { id: 'stuck', label: '卡住' },
  { id: 'done', label: '完成' },
  { id: 'queued', label: '排队' },
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
export function taskTitle(t) {
  const spec = specOf(t);
  return String(spec.title || (t && t.title) || (t && t.name) || '未命名任务');
}

function taskRow(t) {
  const id = taskIdOf(t);
  const s = statusOf(t);
  const spec = specOf(t);
  const deps = (spec.dependsOn || s.dependsOn || []).length;
  const agent = String(s.agent || spec.agent || '').trim();
  const sess = String(s.agentSession || '').trim();
  const last = String(s.lastAction || '').trim();
  return '<div class="task-row">' +
    '<div class="task-main"><div class="task-title"><b>' + esc(taskTitle(t)) + '</b>' + phaseChip(t) +
    chip(id, 'mute') +
    (deps ? chip('依赖 ' + deps, 'mute') : '') +
    (agent ? chip('负责 ' + agent, 'mute') : '') + '</div>' +
    '<div class="task-sub small">' + esc(last || '最近动作：—') +
    (String(s.lastError || '').trim() ? ' ' + chip(trunc(String(s.lastError), 60), 'esc') : '') +
    '</div></div>' +
    '<div class="task-acts">' + (sess ? chipGo('对话', 'session/' + encodeURIComponent(sess), 'mute', true, '关联会话') : '') + '</div>' +
    '</div>';
}

/** 任务中心（分组固定词形——正在做/等人/卡住/完成/排队）。 */
function tasksScreen(el, ctx) {
  const tasks = ctx.store.tasks();
  const g = groupTasks(tasks);
  const body = GROUPS.map((grp) => {
    const rows = g[grp.id] || [];
    return '<div class="card"><div class="card-h"><span class="h-title">' + esc(grp.label) + '</span>' + chip(String(rows.length) + ' 项', rows.length ? 'acc' : 'mute') + '</div>' +
      (rows.length ? '<div class="card-b">' + rows.map(taskRow).join('') + '</div>' : '<div class="card-b small faint">' + esc(grp.label) + '：暂无（空态显式）。</div>') +
      '</div>';
  }).join('');
  el.innerHTML = '<div class="view-inner">' +
    screenHead('任务中心', '分组=确定性投影（状态单一写者=调度器）；点击行进详情', chipGo('关注报告', 'report', 'acc', false)) +
    '<div class="grid cols-2">' + body + '</div></div>';

  el.querySelectorAll('.task-row').forEach((row, i) => {
    let n = 0;
    for (const grp of GROUPS) {
      const rows = g[grp.id] || [];
      if (i < n + rows.length) {
        const t = rows[i - n];
        row.style.cursor = 'pointer';
        row.addEventListener('click', () => ctx.go('task/' + encodeURIComponent(taskIdOf(t))));
        return;
      }
      n += rows.length;
    }
  });
}

/** 任务详情（store 在册投影；不在册=诚实空态——URL 手输名不冒充）。 */
function taskDetailScreen(el, ctx, id) {
  const tasks = ctx.store.tasks();
  const t = tasks.find((x) => taskIdOf(x) === String(id));
  const head = screenHead('任务详情', 'id=' + String(id), chipGo('← 任务中心', 'tasks', 'mute', false));
  if (!t) {
    el.innerHTML = '<div class="view-inner">' + head + emptyBox('任务不在册', '快照任务清单无此 id（可能已完成出窗或 id 有误）。') + '</div>';
    return;
  }
  const s = statusOf(t);
  const spec = specOf(t);
  const rows = [
    ['相位', phaseChip(t)],
    ['标题', esc(taskTitle(t))],
    ['id', esc(taskIdOf(t))],
  ];
  if (spec.agent || s.agent) rows.push(['负责', esc(String(s.agent || spec.agent))]);
  if (s.agentSession) rows.push(['会话', chipGo(trunc(String(s.agentSession), 24), 'session/' + encodeURIComponent(String(s.agentSession)), 'acc', true, '关联会话')]);
  if (s.lastAction) rows.push(['最近动作', esc(trunc(String(s.lastAction), 200))]);
  if (s.lastError) rows.push(['最近错误', chip(trunc(String(s.lastError), 120), 'esc')]);
  const deps = spec.dependsOn || s.dependsOn || [];
  if (deps.length) rows.push(['依赖', deps.map((d) => chip(String(d), 'mute')).join(' ')]);
  el.innerHTML = '<div class="view-inner">' + head + '<div class="card"><div class="card-b">' +
    kv(rows) +
    '</div></div></div>';
}

/* ---------- 关注报告（R-UX-6：attention 排序 + 证据路由 + 循环卡 + 可信度注脚） ---------- */
const SEV_VIEW = { critical: { label: '严重', chip: 'esc' }, high: { label: '重要', chip: 'warn' }, medium: { label: '一般', chip: 'mute' } };
function sevView(sev) { return SEV_VIEW[String(sev || '')] || { label: String(sev || '未分级'), chip: 'mute' }; }
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
    return route ? chipGo(ref, route, 'mute', true, '证据——点击进入') : chip(ref || (ev && ev.kind) || '证据', 'mute');
  }).filter(Boolean);
  const action = a.action || {};
  const route = action.target ? String(action.target).replace(/^#\/?/, '') : '';
  return '<div class="att-card sev-' + esc(String(a.severity || 'unknown')) + '">' +
    '<div class="att-h">' + chip(sev.label, sev.chip) +
    (a.category ? chip(a.category, 'mute') : '') +
    '<b>' + esc(String(a.title || '未命名关注项')) + '</b></div>' +
    (a.why ? '<div class="att-why small">' + esc(String(a.why)) + '</div>' : '') +
    '<div class="att-ev">' + (evs.length ? evs.join(' ') : '<span class="small faint">证据：—</span>') + '</div>' +
    (route ? '<div class="att-acts"><button class="btn sm" data-go="' + esc(route) + '">' + esc(String(action.label || '去处理')) + '</button></div>' : '') +
    '</div>';
}

/** 数据可信度注脚（generatedBy/window/missing——缺席数据显式标注）。 */
export function reportFooter(body) {
  const b = body || {};
  const dq = b.dataQuality || {};
  const missing = dq.missing;
  const missText = Array.isArray(missing) ? missing.map((m) => String(m)).join('、') : String(missing || '');
  const gb = String(b.generatedBy || '');
  const gbLabel = gb === 'llm' ? '智能归纳' : gb === 'rules' ? '基础汇总' : (gb || '未注明生成方式');
  return '<div class="report-foot small faint">数据可信度：' + esc(gbLabel) +
    ' · 检查范围 ' + esc(String(b.scope || 'global')) +
    (b.asOf ? ' · 截至 ' + esc(String(b.asOf)) : '') +
    (missText ? ' · 缺席数据：' + esc(missText) : '') +
    ' · 报告为派生视图可随时重建，不落对象账。</div>';
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
    if (!evs.length) degradedNote = '<div class="small faint" style="margin-bottom:10px">快照未携带 attention 聚合，事件归档亦无审批/升级行——本轮报告按在册事实收敛为零命中（不装假）。</div>';
    else degradedNote = '<div class="small faint" style="margin-bottom:10px">快照未携带 attention 聚合——本卡由事件归档行（ApprovalNeeded/Escalate）投影生成。</div>';
  }
  const sorted = sortAttention(att);
  el.innerHTML = '<div class="view-inner">' +
    screenHead('关注报告', 'severity 排序 · 证据可点 · 数据可信度显式', chipGo('任务中心', 'tasks', 'mute', false)) +
    degradedNote +
    '<div data-loop-box="pending"></div>' +
    (sorted.length
      ? '<div class="att-list">' + sorted.map(attentionCard).join('') + '</div>'
      : emptyBox('本窗口无关注项', 'attention 聚合零命中（空态显式——不冒充有风险）。')) +
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
