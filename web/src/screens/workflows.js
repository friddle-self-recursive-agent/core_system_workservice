/* workflows.js —— 工作流三区列表 + 详情（全节点可视化配置工作台/节点图/任务/会话） + 运行独立页 */
import { esc, trunc, timeHtml, relAge } from '../core/format.js';
import { chip, chipGo, kv, emptyBox, screenHead, kindBadge, involvedRoute, repoLinkChip, objRoute, docRoute } from '../render/chips.js';
import { depChips } from './tasks.js';
import { confirmModal, toast } from './modal.js';
import { t as tr } from '../core/i18n.js';
import { yamlTextView } from '../render/md.js';
import { htmlModal } from './modal.js';
import { openDocChat } from './doc-chat.js';
import { workflowStudioHtml, bindWorkflowStudio } from './workflow-config.js';

export function wfAnchorOfRoute(route) {
  const r = String(route || '');
  if (!r) return '';
  if (r.startsWith('workflows/')) return r.slice('workflows/'.length);
  if (/-workflow$/.test(r)) return 'core-workflow';
  return '';
}

/* 相位管线节点图（节点标色按 run.history 运行轨迹） */
function phaseGraphHtml(run, ctx) {
  const phases = Array.isArray(run.phases) ? run.phases : [];
  const cur = String(run.phase || '');
  const st = String(run.status || '');
  // 已走过节点集=history 轨迹 to 词集（含收束行——Reporting 计入）；静态表不再当轨迹源。
  const visited = new Set((Array.isArray(run.history) ? run.history : [])
    .map((h) => String((h && h.to) || '')).filter(Boolean));
  const isTerminal = st === 'done' || st === 'converged';
  const nodeHtml = (p, i) => {
    const active = p === cur && !isTerminal;
    const done = visited.has(p) && (p !== cur || isTerminal);
    const cls = 'pg-node' + (active ? ' active' : '') + (done ? ' done' : '') + (p === 'Failed' ? ' fail' : '');
    return '<div class="' + cls + '" data-pg-node="' + esc(p) + '"><div class="pg-dot"></div>' +
      '<div class="pg-name">' + esc(p) + '</div>' + (active ? chip(st || tr('进行中'), 'acc') : '') + '</div>';
  };
  const body = phases.length
    ? phases.map((p, i) => nodeHtml(p, i)).join('<span class="pg-edge">→</span>')
    : '<div class="pg-node active"><div class="pg-dot"></div><div class="pg-name">' + esc(cur || tr('单节点（声明未投影相位）')) + '</div>' + runChip(run) + '</div>';
  return '<div class="pg-wrap">' + body + '</div>' +
    '<div class="small faint">' + esc(tr('点击节点查看该阶段的事件与涉及文档。')) + '</div>' +
    '<div data-pg-drawer style="margin-top:10px"></div>';
}

function bindPhaseGraph(el, ctx, run) {
  el.querySelectorAll('[data-pg-node]').forEach((n) => {
    n.addEventListener('click', () => {
      const ph = String(n.getAttribute('data-pg-node') || '');
      const drawer = el.querySelector('[data-pg-drawer]');
      if (!drawer) return;
      const evs = ctx.store.events().filter((e) => {
        const hay = JSON.stringify([e.dedupKey, e.involved, e.type]);
        return hay.includes(String(run.workflow || '')) || hay.includes(ph);
      }).slice(0, 10);
      const involved = String(run.involved || '');
      const evRows = evs.length
        ? evs.map((e) => {
          const eid = e && (e.id != null ? e.id : (e.eventId != null ? e.eventId : ''));
          const goAttr = eid !== '' ? ' data-go="event/' + encodeURIComponent(String(eid)) + '" role="link" style="cursor:pointer"' : '';
          return '<div class="mini-row"' + goAttr + '><div class="mr-main"><div><b>' + esc(String(e.type || '')) + '</b></div>' +
            '<div class="mr-sub small mono">' + esc(trunc(String(e.involved || '—'), 60)) + '</div></div>' +
            '<span class="small faint">' + esc(relAge(e.ts) || '') + '</span></div>';
        }).join('')
        : '<div class="small faint">' + esc(tr('窗口内无该相位事件。')) + '</div>';
      drawer.innerHTML = '<div class="card"><div class="card-h"><span class="h-title">' + esc(ph) + '</span></div><div class="card-b">' +
        (involved
          ? '<div class="mr-sub">' + esc(tr('涉及')) + '：' + involved.split(',').map((x) => {
            const parts = x.trim().split('/');
            return parts.length >= 3
              ? '<span class="ref-chip" data-go="obj/' + esc(parts[0]) + '/' + esc(parts[1]) + '/' + esc(parts[2]) + '">' + esc(x.trim()) + '</span>'
              : '<span class="small faint mono">' + esc(x.trim()) + '</span>';
          }).join(' ') + '</div>'
          : '') +
        '<div class="small faint">' + esc(tr('事件链（归档窗口）：')) + '</div>' + evRows +
        '</div></div>';
    });
  });
}

const RUN_CLS = { running: 'acc', done: 'ok', ok: 'ok', success: 'ok', converged: 'ok', failed: 'esc', error: 'esc', canceled: 'mute', queued: 'mute', pending: 'mute' };
/** 单次 run 终态判定（F-04 闭合集：converged|failed|canceled|done=终态——终态不写「正在发生」）。 */
function isTerminalRun(st) { return ['done', 'converged', 'failed', 'canceled', 'ok', 'success'].includes(String(st || '')); }
const REG_CLS = { running: 'acc', active: 'acc', suspended: 'warn', awaitingapproval: 'warn', failed: 'esc', error: 'esc', done: 'ok', ok: 'ok', idle: 'mute', registered: 'mute' };
const PH_CLS = { awaitingapproval: 'warn', failed: 'esc', done: 'ok', executing: 'acc', running: 'acc' };

function runChip(r) {
  const st = String((r && (r.status || r.state || r.phase)) || tr('未知'));
  return chip(st, RUN_CLS[st] || 'mute');
}

function phChip(ph) {
  const p = String(ph || '');
  return chip(p || '—', PH_CLS[p.toLowerCase()] || 'mute');
}

function wfDesignRoute() {
  return 'designdoc/' + encodeURIComponent('core_design') + '/' + encodeURIComponent('concepts/workflow.md');
}

/* V-20/V-19：退役静态假卡——模板区只展示 snapshot 中 Template kind 真实对象；
 * 零在册=诚实空态（含设计稿深链）；若在册 Workflow 含 bug-lifecycle 同步附真运行入口。 */
function realTplCardHtml(o) {
  const scope = o.scope || 'global';
  return '<div class="card"><div class="card-h"><span class="h-title mono">' + esc(o.name) + '</span>' +
    '<span class="h-right">' + chip('Template', 'acc') + ' ' + chip('r' + String(o.rv || 0), 'mute') + '</span></div>' +
    '<div class="card-b"><div class="small mono faint">scope: ' + esc(scope) + '</div>' +
    '<div class="dh-actions" style="margin-top:8px">' +
    chipGo(tr('对象工作台'), objRoute('Template', scope, o.name), 'acc', true, tr('查看 Template 声明')) + ' ' +
    chipGo(tr('编辑器'), docRoute('Template', scope, o.name), 'mute', true, tr('查看正文')) +
    '</div></div></div>';
}

export function templatesSectionHtml(ctx) {
  const tpls = ctx.store.refs('Template');
  const bugWf = ctx.store.refs('Workflow').find((w) => w && w.name === 'bug-lifecycle');
  const bugRuns = ctx.store.runs().filter((r) => r && String(r.workflow || r.name || '') === 'bug-lifecycle');
  const bugBlock = (bugWf || bugRuns.length)
    ? '<div class="card" style="margin-bottom:10px"><div class="card-h"><span class="h-title mono">bug-lifecycle</span><span class="h-right">' + chip('Workflow', 'acc') + ' ' + chip(bugRuns.length + tr(' 次运行'), bugRuns.length ? 'acc' : 'mute') + '</span></div><div class="card-b"><div class="dh-actions">' + chipGo(tr('工作流详情'), 'wf/bug-lifecycle', 'acc', true) + ' ' + chipGo(tr('文档库 BugDoc'), 'docs', 'mute', true) + '</div></div></div>'
    : '';
  const tplBlock = tpls.length
    ? '<div class="grid cols-2">' + tpls.map(realTplCardHtml).join('') + '</div>'
    : '<div class="card"><div class="card-b"><div class="small faint">' + esc(tr('暂无流程模板（Template 对象在册后此处显示）。')) + ' ' + chipGo(tr('设计见 workflow.md §9b'), wfDesignRoute(), 'mute', true, tr('设计文档工作台')) + '</div></div></div>';
  return bugBlock + tplBlock;
}

/** 列表三区（§19.3）：在册注册表（GET /v1/workflows 现读）∪ 运行记录（GET /v1/workflows/runs
 * 现读+快照 runs 聚合并显去重）∪ 模板三卡。快照不带 workflows 键=旧假空态根因——此处全部走活体面。 */
function listScreen(el, ctx) {
  const state = { rows: [], filter: '', err: '' }; // rows=runs face 全窗口行；filter=按 workflow 过滤（注册表行「运行记录」点击）；err=runs 面读失败词形
  const snapRuns = ctx.store.runs().slice(0, 8); // 快照聚合 recent（并显去重——缺键=空数组不装假）

  const runRowHtml = (r) => {
    const wf = String((r && (r.workflow || r.name)) || '');
    const no = r && (r.run != null ? r.run : (r.no != null ? r.no : ''));
    if (!wf && (no == null || no === '')) return '';
    const started = r && (r.startedAt || r.at || r.ts);
    const invChips = String((r && r.involved) || '').split(',').map((x) => x.trim()).filter(Boolean).slice(0, 3).map((x) => {
      const rt = involvedRoute(x);
      return rt ? chipGo(trunc(x, 36), rt, 'mute', true, tr('去对象工作台')) : '<span class="chip mute tiny">' + esc(trunc(x, 36)) + '</span>';
    }).join(' ');
    const go = wf && no != null && no !== '' ? 'wf/' + encodeURIComponent(wf) + '/r/' + encodeURIComponent(String(no)) : '';
    return '<tr' + (go ? ' data-go="' + esc(go) + '" style="cursor:pointer"' : '') + '>' +
      '<td class="mono small">#' + esc(String(no != null && no !== '' ? no : '—')) + '</td>' +
      '<td class="mono small">' + esc(wf || '—') + '</td>' +
      '<td>' + phChip(r && r.phase) + '</td>' +
      '<td>' + runChip(r) + '</td>' +
      '<td class="small mono">' + esc(relAge(started) || (started ? String(started).slice(0, 16) : '—')) + '</td>' +
      '<td class="small">' + (invChips || '<span class="faint">—</span>') + '</td></tr>';
  };

  const paintRuns = () => {
    const slot = el.querySelector('[data-wf-runs]');
    if (!slot) return;
    const seen = new Set();
    const merged = [];
    for (const r of state.rows.concat(snapRuns)) {
      const k = String((r && (r.workflow || r.name)) || '') + '#' + String(r && (r.run != null ? r.run : (r.no != null ? r.no : '')));
      if (seen.has(k)) continue;
      seen.add(k);
      merged.push(r);
    }
    const shown = (state.filter ? merged.filter((r) => String((r && (r.workflow || r.name)) || '') === state.filter) : merged).slice(0, 50);
    slot.innerHTML = '<div class="card"><div class="card-b" style="overflow-x:auto">' +
      '<div style="margin-bottom:6px">' +
      chip(merged.length + tr(' 条运行记录'), 'mute') + ' ' +
      (state.filter ? chipGo(tr('清除过滤：') + state.filter, 'workflows', 'warn', true, tr('清除过滤并回全表')) : '') +
      '</div>' +
      (state.err ? '<div class="small faint" style="margin-bottom:6px">' + esc(tr('运行记录暂不可读（')) + esc(state.err) + esc(tr('）——下表仅概览行。')) + '</div>' : '') +
      (shown.length
        ? '<table class="tbl"><thead><tr><th>#</th><th>' + esc(tr('工作流')) + '</th><th>' + esc(tr('相位')) + '</th><th>' + esc(tr('状态')) + '</th><th>' + esc(tr('开始')) + '</th><th>' + esc(tr('涉及')) + '</th></tr></thead><tbody>' +
          shown.map(runRowHtml).join('') + '</tbody></table>' +
          '<div class="small faint" style="margin-top:6px">' + esc(tr('行点击进运行详情；involved chip=对象工作台深链。')) + '</div>'
        : '<div class="small faint">' + esc(state.filter ? tr('该 workflow 窗口内暂无运行行。') : tr('暂无运行记录。'))) +
      '</div></div>';
  };

  el.innerHTML = '<div class="view-inner">' +
    screenHead(tr('工作流'), tr('三区=在册注册表（GET /v1/workflows 现读）∪ 运行记录（GET /v1/workflows/runs+快照聚合并显）∪ 模板（§9b 设计卡）。'),
      chipGo(tr('设计文档'), wfDesignRoute(), 'mute', true, tr('workflow.md §9b'))) +
    '<h2 class="sec-title">' + esc(tr('在册工作流')) + '</h2>' +
    '<div data-wf-reg><div class="card"><div class="card-b"><div class="small faint">' + esc(tr('注册表现读中…')) + '</div></div></div></div>' +
    '<h2 class="sec-title" style="margin-top:18px">' + esc(tr('运行记录')) + '</h2>' +
    '<div data-wf-runs><div class="card"><div class="card-b"><div class="small faint">' + esc(tr('运行记录现读中…')) + '</div></div></div></div>' +
    '<h2 class="sec-title" style="margin-top:18px">' + esc(tr('模板')) + '</h2>' +
    templatesSectionHtml(ctx) +
    '</div>';

  paintRuns();

  /* 注册表现读（活体面——不再读快照 workflows 键） */
  (async () => {
    const stamp = typeof location !== 'undefined' ? location.hash : '';
    const slot = el.querySelector('[data-wf-reg]');
    const r = await ctx.api.workflows();
    if (!slot || !slot.isConnected || (typeof location !== 'undefined' ? location.hash : '') !== stamp) return;
    if (!r.ok) {
      slot.innerHTML = '<div class="card"><div class="card-b"><div class="small faint">' + esc(tr('工作流清单暂不可读（')) + esc(String(r.error || r.status)) + esc(tr('），请刷新重试。')) + '</div></div></div>';
      return;
    }
    const data = r.data && typeof r.data === 'object' ? r.data : {};
    const rows = Array.isArray(data.workflows) ? data.workflows : [];
    if (!rows.length) {
      slot.innerHTML = '<div class="card"><div class="card-b"><div class="small faint">' + esc(tr('暂无在册工作流。')) + '</div></div></div>';
      return;
    }
    slot.innerHTML = '<div class="card"><div class="card-b" style="overflow-x:auto">' +
      chip((data.count != null ? data.count : rows.length) + tr(' 条在册'), 'acc') + ' ' +
      (data.scanCapped ? chip(tr('扫描截断（窗口内 ') + String(data.scanned != null ? data.scanned : '?') + tr(' 条）'), 'warn') : '') +
      '<table class="tbl" style="margin-top:8px"><thead><tr><th>' + esc(tr('名称')) + '</th><th>' + esc(tr('状态')) + '</th><th>' + esc(tr('连接')) + '</th><th>' + esc(tr('最近')) + '</th><th></th></tr></thead><tbody>' +
      rows.map((w) => {
        const name = String((w && w.name) || '');
        const ex = (w && w.execution) || {};
        const exState = String(ex.state || '');
        const connTxt = String(ex.connector || '') + (ex.connectorScope ? '/' + String(ex.connectorScope) : '');
        const connCell = connTxt
          ? chipGo(connTxt, 'connectors', 'mute', true, tr('连接层——查看绑定 connector'))
          : '<span class="faint">—</span>';
        const latest = w && w.latest;
        const latestTxt = typeof latest === 'string' ? latest : String((latest && (latest.summary || latest.note || latest.status || latest.state)) || '');
        return '<tr>' +
          '<td class="mono small">' + (name ? '<span class="lk" data-go="wf/' + encodeURIComponent(name) + '">' + esc(name) + '</span>' : '—') + '</td>' +
          '<td>' + chip(exState || '—', REG_CLS[exState.toLowerCase()] || 'mute') + '</td>' +
          '<td class="mono small">' + connCell + '</td>' +
          '<td class="small">' + esc(trunc(latestTxt || '—', 60)) + '</td>' +
          '<td style="text-align:right">' + (name ? '<span class="chip mute tiny click" data-wf-runfilter="' + esc(name) + '" title="' + esc(tr('过滤本页运行记录')) + '">' + esc(tr('运行记录')) + '</span>' : '') + '</td>' +
          '</tr>';
      }).join('') + '</tbody></table></div></div>';
    slot.querySelectorAll('[data-wf-runfilter]').forEach((c) => {
      c.addEventListener('click', () => {
        state.filter = String(c.getAttribute('data-wf-runfilter') || '');
        paintRuns();
        const rs = el.querySelector('[data-wf-runs]');
        if (rs && rs.scrollIntoView) rs.scrollIntoView({ behavior: 'smooth', block: 'start' });
      });
    });
  })();

  /* 运行记录现读（GET /v1/workflows/runs 全窗口——失败=快照聚合行仍并显，不装假） */
  (async () => {
    const stamp = typeof location !== 'undefined' ? location.hash : '';
    const r = await ctx.api.workflowRuns();
    if (!el.isConnected || (typeof location !== 'undefined' ? location.hash : '') !== stamp) return;
    if (!r.ok) {
      state.err = String(r.error || r.status);
    } else {
      const data = r.data && typeof r.data === 'object' ? r.data : {};
      state.rows = Array.isArray(data.runs) ? data.runs : [];
    }
    paintRuns();
  })();
}

/* ── 工作流详情 v2（定义/以人为本可视化配置/节点/任务/节点会话） ── */
const WF_ALIASES = {
  'bug-workflow': 'core-system-developer-loop',
  'bug-lifecycle': 'core-system-developer-loop',
  'core-workflow': 'core-system-developer-loop',
  'dev-delivery': 'core-system-developer-loop',
  'apply-workflow': 'default-task',
  'simple-shot': 'default-task',
  'watch-only': 'default-task',
};

function wfApiRows(res) { return (res && res.ok && res.data && Array.isArray(res.data.workflows)) ? res.data.workflows : []; }
function wfApiRuns(res) { return (res && res.ok && res.data && Array.isArray(res.data.runs)) ? res.data.runs : []; }
function wfApiTasks(res) { return (res && res.ok && res.data && Array.isArray(res.data.tasks)) ? res.data.tasks : []; }
function wfTaskIndex(tasks) {
  const m = new Map();
  for (const t of tasks) if (t && t.name) m.set(String(t.scope || 'global') + '/' + String(t.name), t);
  return m;
}
export function wfTaskKey(involved) {
  const m = /^Task\/([^/]+)\/([^/]+)/.exec(String(involved || ''));
  return m ? m[1] + '/' + m[2] : '';
}

const SESSION_MATCH_WINDOW_MS = 10 * 60 * 1000;

/** 真实会话锚+来源：history 节点会话 → run 运行会话 → Task journal 任务会话。
 * 旧 task.agentSession=taskId 不当 session；时刻最近匹配避免多次重跑全挂最新会话。 */
export function wfSessionAnchor(run, task, sessionRows, hist = null) {
  const hs = String((hist && hist.sessionId) || '');
  if (/^session[-_]/i.test(hs)) return { sid: hs, label: tr('节点会话') };
  const direct = String((run && run.sessionId) || '');
  if (/^session[-_]/i.test(direct)) return { sid: direct, label: tr('运行会话') };
  const key = wfTaskKey(run && run.involved);
  const candidates = (Array.isArray(sessionRows) ? sessionRows : [])
    .filter((x) => x && x.taskId === key && /^session[-_]/i.test(String(x.sessionId || '')));
  if (!candidates.length) return { sid: '', label: tr('任务会话') };
  const started = Date.parse(String((run && (run.startedAt || run.at || run.ts)) || ''));
  if (!Number.isFinite(started)) {
    // 无 run 时刻：仅当 Task 只有一个真实会话时可无歧义回退；多候选=不猜。
    return candidates.length === 1 ? { sid: String(candidates[0].sessionId), label: tr('任务会话') } : { sid: '', label: tr('任务会话') };
  }
  const ranked = candidates.map((x) => ({
    x,
    d: Number.isFinite(Number(x.startedAt)) ? Math.abs(Number(x.startedAt) - started) : Number.POSITIVE_INFINITY,
  })).sort((a, b) => a.d - b.d);
  // 任务卡与 run 起始时刻超过 10m=不是该次运行的真实会话（宁缺勿滥，避免多次重跑全挂最新会话）。
  if (!ranked.length || ranked[0].d > SESSION_MATCH_WINDOW_MS) return { sid: '', label: tr('任务会话') };
  return { sid: String(ranked[0].x.sessionId), label: tr('任务会话') };
}

export function wfRealSession(run, task, sessionRows) {
  return wfSessionAnchor(run, task, sessionRows).sid;
}

/** 会话锚 chip：只认真实 session-*（dsh 会话 id）→ 会话屏；taskId/agent 引用不当会话。 */
function wfSessionChip(sid, label) {
  const s = String(sid || '');
  if (!s) return '<span class="small faint">' + esc(tr('会话未记录')) + '</span>';
  const isSession = /^session[-_]/i.test(s);
  if (!isSession) return '<span class="small faint">' + esc((label ? tr(label) + ' ' : '') + tr('任务锚（非 sessionId）') + ' ' + trunc(s, 24)) + '</span>';
  return chipGo((label ? tr(label) + ' ' : '') + trunc(s, 24), 'session/' + encodeURIComponent(s), 'acc', true, s);
}

/** 节点聚合（当前相位/到过次数/失败数）。 */
function wfNodeStats(runs, phases) {
  const stats = new Map();
  for (const p of phases) stats.set(p, { cur: 0, visited: 0, failed: 0 });
  for (const r of runs) {
    const cur = String(r.phase || '');
    if (stats.has(cur)) stats.get(cur).cur++;
    const seen = new Set((Array.isArray(r.history) ? r.history : []).map((h) => String((h && h.to) || '')));
    for (const p of seen) if (stats.has(p)) stats.get(p).visited++;
    if (String(r.status || '') === 'failed' && stats.has('Failed')) stats.get('Failed').failed++;
  }
  return stats;
}

/** 节点图（定义相位链 + 失败侧轨；点击看到达节点的人/任务/会话）。 */
export function wfNodeGraphHtml(wf, runs) {
  const phases = (Array.isArray(wf && wf.phases) ? wf.phases : []).map(String);
  const failed = runs.filter((r) => String(r.status || '') === 'failed').length;
  const list = phases.concat(failed && !phases.includes('Failed') ? ['Failed'] : []);
  if (!list.length) return '<div class="small faint">' + esc(tr('该 workflow 未投影相位链。')) + '</div>';
  const stats = wfNodeStats(runs, list);
  const body = list.map((p) => {
    const s = stats.get(p) || { cur: 0, visited: 0, failed: 0 };
    const cls = 'pg-node' + (s.cur ? ' active' : '') + ((s.visited || s.failed) ? ' done' : '') + (p === 'Failed' ? ' fail' : '');
    return '<div class="' + cls + '" data-wf-node="' + esc(p) + '" role="button">' +
      '<div class="pg-dot"></div><div class="pg-name">' + esc(p) + '</div>' +
      (s.cur ? '<span class="chip acc tiny">' + esc(tr('当前') + ' ' + s.cur) + '</span>' : '') +
      (s.visited ? '<span class="chip mute tiny">' + esc(tr('到过') + ' ' + s.visited) + '</span>' : '') +
      '</div>';
  }).join('<span class="pg-edge">→</span>');
  return '<div class="pg-wrap">' + body + '</div>' +
    '<div class="small faint">' + esc(tr('点击节点：到达该节点的运行/任务与真实会话锚。')) + '</div>' +
    '<div data-wf-node-drawer class="wf-node-drawer"></div>';
}

function bindWfNodeGraph(el, runs, taskIdx, wfName, sessionRows = []) {
  el.querySelectorAll('[data-wf-node]').forEach((n) => {
    n.addEventListener('click', () => {
      const ph = String(n.getAttribute('data-wf-node') || '');
      const drawer = el.querySelector('[data-wf-node-drawer]');
      if (!drawer) return;
      const hits = runs.filter((r) => String(r.phase || '') === ph ||
        (Array.isArray(r.history) ? r.history : []).some((h) => String((h && h.to) || '') === ph || String((h && h.from) || '') === ph));
      const rows = hits.length ? hits.map((r) => {
        const no = r.run != null ? r.run : (r.no != null ? r.no : '');
        const task = taskIdx.get(wfTaskKey(r.involved));
        const hist = (Array.isArray(r.history) ? r.history : []).find((h) => String((h && h.to) || '') === ph || String((h && h.from) || '') === ph);
        const anchor = wfSessionAnchor(r, task, sessionRows, hist);
        return '<div class="mini-row"><div class="mr-main"><div><b class="mono">run #' + esc(String(no)) + '</b> ' + runChip(r) + '</div>' +
          '<div class="mr-sub small">' + esc(task ? String(task.title || task.name || '') : String(r.involved || '—')) + '</div></div>' +
          '<div class="mr-side">' + wfSessionChip(anchor.sid, anchor.label) + ' ' +
          chipGo(tr('运行'), 'wf/' + encodeURIComponent(wfName) + '/r/' + encodeURIComponent(String(no)), 'mute', true) + '</div></div>';
      }).join('') : '<div class="small faint">' + esc(tr('暂无运行到达该节点。')) + '</div>';
      drawer.innerHTML = '<div class="card"><div class="card-h"><span class="h-title">' + esc(ph) + '</span>' +
        '<span class="h-right">' + esc(tr('到达') + ' ' + hits.length) + '</span></div><div class="card-b">' + rows + '</div></div>';
    });
  });
}

/** ConnectorOps 骨架节点（真实 executionFlow；缺席=执行面未绑定，不伪造）。 */
export function wfFlowNodesHtml(wf) {
  const nodes = wf && Array.isArray(wf.executionFlow) ? wf.executionFlow : [];
  if (!nodes.length) {
    return '<div class="small faint">' + esc(tr('执行面未绑定（ConnectorOps 骨架缺席）——workflow apply 后的落地 agent 待接线；当前节点=相位链，节点级会话未单独记录。')) + '</div>';
  }
  return '<div class="wf-flow">' + nodes.map((n) =>
    '<div class="wf-flow-node"><div class="wf-flow-name mono">' + esc(String(n.name || '—')) + '</div>' +
    '<div class="small">' + esc(String(n.title || '')) + '</div>' +
    '<div class="small faint">' + esc([n.unit, n.op, n.connector, n.provider].filter(Boolean).join(' · ')) + '</div></div>').join('') + '</div>';
}

/** 单 run 的逐节点会话表（history→run→task 三级锚；空=未记录）。 */
export function wfRunNodesHtml(run, task, sessionRows = []) {
  const phases = (Array.isArray(run.phases) && run.phases.length ? run.phases : [run.phase || tr('单节点')]).map(String);
  const byTo = new Map();
  for (const h of (Array.isArray(run.history) ? run.history : [])) {
    if (h && h.to) byTo.set(String(h.to), h);
  }
  const cur = String(run.phase || '');
  const rows = phases.map((p) => {
    const h = byTo.get(p);
    const state = p === cur ? tr('当前') : (h ? tr('已走过') : tr('未到达'));
    const anchor = wfSessionAnchor(run, task, sessionRows, h);
    // 未到达节点没有会话；到了的节点优先 history/run，缺省才显示 Task 会话（节点未单独记录）。
    const sessionHtml = (p === cur || h)
      ? wfSessionChip(anchor.sid, anchor.label)
      : '<span class="small faint">' + esc(tr('无会话（未到达）')) + '</span>';
    return '<div class="mini-row wf-node-row"><div class="mr-main"><div><b class="mono">' + esc(p) + '</b> ' +
      chip(state, p === cur ? 'acc' : (h ? 'ok' : 'mute')) + '</div>' +
      '<div class="mr-sub small mono">' + esc(h ? String(h.on || '') : '') + '</div></div>' +
      '<div class="mr-side">' + sessionHtml + (h && h.at ? '<span class="w-ev-time mono">' + esc(relAge(h.at) || '') + '</span>' : '') + '</div></div>';
  }).join('');
  return '<div class="small faint">' + esc(tr('节点会话锚=history.sessionId → run.sessionId → Task journal 反查（空=未记录，不编造；旧 task.agentSession=taskId 不当 session）。')) + '</div>' + rows;
}

/** 任务列表（运行 involved=Task/* 的真实关联；标题/phase/会话取 Task 对象）。 */
function wfTasksHtml(runs, taskIdx, sessionRows = []) {
  const seen = new Map();
  for (const r of runs) {
    const key = wfTaskKey(r.involved);
    if (!key) continue;
    if (!seen.has(key)) seen.set(key, { key, task: taskIdx.get(key), runs: [] });
    seen.get(key).runs.push(r);
  }
  if (!seen.size) return '';
  const rows = [...seen.values()].map((x) => {
    const t = x.task;
    const title = t ? String(t.title || t.name || x.key) : x.key;
    const ph = t ? String(t.phase || '') : '';
    const taskSessions = (Array.isArray(sessionRows) ? sessionRows : [])
      .filter((r) => r && r.taskId === x.key && /^session[-_]/i.test(String(r.sessionId || '')))
      .sort((a, b) => (Number(b.updatedAt) || 0) - (Number(a.updatedAt) || 0));
    const latest = taskSessions[0] || null;
    const direct = /^session[-_]/i.test(String((t && t.agentSession) || '')) ? String(t.agentSession) : '';
    const sid = latest ? String(latest.sessionId) : direct;
    const sessionHtml = wfSessionChip(sid, '任务会话') +
      (taskSessions.length > 1 ? '<span class="small faint">' + esc(tr('共 {n} 个真实会话').replace('{n}', String(taskSessions.length))) + '</span>' : '');
    return '<div class="mini-row"><div class="mr-main"><div><b>' + esc(trunc(title, 64)) + '</b> ' +
      (ph ? chip(ph, PH_CLS[ph.toLowerCase()] || 'mute') : '') + '</div>' +
      '<div class="mr-sub small mono">Task/' + esc(x.key) + ' · ' + esc(tr('runs') + ' ' + x.runs.length) + '</div></div>' +
      '<div class="mr-side">' + sessionHtml + ' ' +
      chipGo(tr('任务'), 'task/' + encodeURIComponent(x.key), 'mute', true) + '</div></div>';
  }).join('');
  return '<h2 class="sec-title">' + esc(tr('任务')) + '</h2><div class="card">' + rows + '</div>';
}

/** 运行/任务行（details 展开=逐节点会话）。 */
function wfRunsHtml(runs, taskIdx, sessionRows = []) {
  if (!runs.length) return '';
  return '<h2 class="sec-title">' + esc(tr('运行记录')) + '</h2><div class="card">' + runs.map((r, i) => {
    const no = r.run != null ? r.run : (r.no != null ? r.no : i + 1);
    const task = taskIdx.get(wfTaskKey(r.involved));
    const anchor = wfSessionAnchor(r, task, sessionRows);
    return '<details class="ex-block wf-run"><summary class="ex-sum wf-run-sum">' +
      '<span class="ex-chev">▶</span><b class="mono">run #' + esc(String(no)) + '</b> ' + runChip(r) +
      '<span class="small wf-run-title">' + esc(task ? String(task.title || task.name || '') : String(r.involved || '—')) + '</span>' +
      '<span class="wf-run-side">' + wfSessionChip(anchor.sid, anchor.label) +
      '<span class="w-ev-time mono">' + esc(relAge(r.startedAt || r.at || r.ts) || '—') + '</span>' +
      chipGo(tr('运行详情'), 'wf/' + encodeURIComponent(String(r.workflow || '')) + '/r/' + encodeURIComponent(String(no)), 'mute', true) +
      '</span></summary><div class="ex-body">' + wfRunNodesHtml(r, task, sessionRows) + '</div></details>';
  }).join('') + '</div>';
}

/** 编辑入口：object 源=编辑器；builtin 源=设计文档/scheme（高敏，不伪造表单）。 */
function wfEditChips(wf, name) {
  // R-UX-10 二次收紧（2026-11-28）：修改默认=对话（LLM 改文档→Apply）；手改=配置详情弹窗（studio 表单/YAML）。
  const chatFix = '<button type="button" class="btn btn-chip sm" data-wf-chatfix="' + esc(String(name)) + '" title="' +
    esc(tr('发起对话修改（LLM 改文档→Apply——主通路）')) + '">' + esc(tr('对话修改')) + '</button> ';
  if (wf && String(wf.specSource || '') === 'object') {
    const scope = String(wf.scope || 'global');
    return chatFix +
      '<button type="button" class="btn btn-chip sm" data-wf-config-btn title="' + esc(tr('配置详情（表单/YAML 手改——默认不展示）')) + '">' + esc(tr('配置详情')) + '</button> ' +
      chipGo(tr('对象工作台'), objRoute('Workflow', scope, name), 'mute', true);
  }
  return chatFix +
    chipGo(tr('编辑定义（workflow.md）'), wfDesignRoute(), 'mute', true, tr('设计文档工作台')) + ' ' +
    chipGo(tr('相位表（scheme）'), 'designdoc/' + encodeURIComponent('core_design') + '/' + encodeURIComponent('core/scheme.yaml'), 'mute', true);
}

/** 详情：真实三源现读 → 定义卡 + 节点图 + 骨架节点 + 任务 + 运行/节点会话。 */
let wfDetailSeq = 0;
function detailScreen(el, ctx, name) {
  const stamp = '#' + (++wfDetailSeq);
  el.setAttribute('data-wf-detail-stamp', stamp);
  el.innerHTML = '<div class="view-inner">' +
    screenHead(tr('工作流'), String(name), chipGo(tr('← 工作流'), 'workflows', 'mute', false)) +
    '<div data-wf-detail><div class="small faint">' + esc(tr('工作流现读中…（定义/运行/任务）')) + '</div></div></div>';
  (async () => {
    const [wfsRes, runsRes, tasksRes] = await Promise.all([
      ctx.api.workflows(), ctx.api.workflowRuns(), ctx.api.tasks(),
    ]);
    if (!el.isConnected || el.getAttribute('data-wf-detail-stamp') !== stamp) return;
    const rows = wfApiRows(wfsRes);
    const runsAll = wfApiRuns(runsRes);
    const taskIdx = wfTaskIndex(wfApiTasks(tasksRes));
    const serverAliases = (wfsRes && wfsRes.ok && wfsRes.data && wfsRes.data.aliases) || {};
    const alias = serverAliases[String(name)] || WF_ALIASES[String(name)];
    let wf = rows.find((x) => String(x && x.name) === String(name));
    if (!wf && alias) wf = rows.find((x) => String(x && x.name) === String(alias));
    const usedName = wf ? String(wf.name) : String(alias || name);
    const runs = runsAll.filter((r) => String(r && r.workflow) === usedName || String(r && r.workflow) === String(name))
      .sort((a, b) => (Number(b.run) || 0) - (Number(a.run) || 0));
    let sessionRows = [];
    const taskIds = [...new Set(runs.map((r) => wfTaskKey(r.involved)).filter(Boolean))];
    if (taskIds.length && ctx.api.taskSessions) {
      try {
        const sr = await ctx.api.taskSessions(taskIds);
        if (sr && sr.ok && Array.isArray(sr.data && sr.data.items)) sessionRows = sr.data.items;
      } catch { /* 映射缺席=会话未记录，不编造 */ }
    }
    if (!el.isConnected || el.getAttribute('data-wf-detail-stamp') !== stamp) return;
    const detail = el.querySelector('[data-wf-detail]');
    if (!detail) return;
    if (!rows.length && !runs.length) {
      detail.innerHTML = emptyBox(tr('工作流面暂不可读'),
        tr('GET /v1/workflows 与 /v1/workflows/runs 均未返回数据——') + ' ' +
        chipGo(tr('重试'), 'wf/' + encodeURIComponent(String(name)), 'acc', true));
      return;
    }
    if (!wf) {
      const similar = rows.filter((x) => String(x.name || '').toLowerCase().includes(String(name).toLowerCase().split('-')[0]))
        .slice(0, 8);
      detail.innerHTML = emptyBox(tr('工作流不在册'), tr('暂无此工作流（可能已下线或名称有误）。')) +
        (similar.length ? '<div class="card"><div class="card-h"><span class="h-title">' + esc(tr('相近工作流')) + '</span></div><div class="card-b">' +
          similar.map((x) => chipGo(String(x.name), 'wf/' + encodeURIComponent(String(x.name)), 'mute', true)).join(' ') + '</div></div>' : '');
      return;
    }
    const latest = wf.latest || null;
    const exec = wf.execution || {};
    const aliasNote = String(name) !== usedName
      ? chip(tr('词形兼容') + '：' + String(name) + ' → ' + usedName, 'mute') : '';
    // 概览 kv（复用既有 kv 渲染）
    const kvRows = [
      [tr('来源'), chip(String(wf.specSource || '—'), wf.specSource === 'builtin' ? 'acc' : (wf.specSource === 'object' ? 'ok' : 'mute'))],
      [tr('执行面'), exec.connector ? chip(String(exec.state || '—'), exec.state === 'bound' ? 'ok' : 'warn') + ' ' +
        chipGo(String(exec.connector), 'connectors', 'mute', true, tr('ConnectorOps')) : '—'],
      [tr('骨架'), exec.skeleton ? chip(tr('在册'), 'ok') : chip(tr('缺席'), 'mute')],
      [tr('运行'), chip(tr('环内 ') + runs.length + ' / 累计 ' + String(wf.runsTotal != null ? wf.runsTotal : '—'), runs.length ? 'acc' : 'mute') +
        (wf.runs24h != null ? ' ' + chip('24h ' + String(wf.runs24h), 'mute') : '')],
      [tr('最新'), latest ? runChip(latest) + ' ' + chip('run #' + String(latest.run), 'mono') + ' ' + esc(relAge(latest.startedAt) || '') : '—'],
      [tr('生效'), wf.effective ? chip(String(wf.effective.scope || '') + ' · ' + (wf.effective.enabled ? tr('启用') : tr('停用')), wf.effective.enabled ? 'ok' : 'mute') : '—'],
    ];
    const runsHtml = wfRunsHtml(runs, taskIdx, sessionRows);
    const tasksHtml = wfTasksHtml(runs, taskIdx, sessionRows);
    detail.innerHTML =
      '<div class="run-hero"><div class="rh-top">' +
      (aliasNote ? aliasNote : '') + '</div>' +
      '<div class="rh-title">' + esc(usedName) + ' ' + chip(String(wf.specSource || '—'), 'acc') + ' ' +
      chip(String(exec.state || '—'), exec.state === 'bound' ? 'ok' : 'warn') + '</div>' +
      '<div class="rh-meta"><span>' + esc(tr('定义/运行/任务均为现读真实数据')) + '</span>' +
      (latest ? '<span>' + esc(tr('最近运行') + ' ') + runChip(latest) + '</span>' : '') + '</div>' +
      '<div class="dh-actions" style="margin-top:10px">' + wfEditChips(wf, usedName) + '</div></div>' +
      '<div class="card"><div class="card-h"><span class="h-title">' + esc(tr('定义')) + '</span><span class="h-right">' +
      // R-UX-10 2026-11-28：YAML 声明=弹窗按需（人不主动看 scheme——fixed 卡退役为按钮）。
      (String(wf.rawYaml || '') ? '<button type="button" class="btn btn-chip sm" data-wf-yaml-btn>' + esc(tr('YAML 声明')) + '</button>' : '') +
      '</span></div><div class="card-b">' + kv(kvRows) + '</div></div>' +
      '<h2 class="sec-title">' + esc(tr('节点（相位链）')) + '</h2><div class="card"><div class="card-b">' + wfNodeGraphHtml(wf, runs) + '</div></div>' +
      '<h2 class="sec-title">' + esc(tr('执行节点（ConnectorOps 骨架）')) + '</h2><div class="card"><div class="card-b">' + wfFlowNodesHtml(wf) + '</div></div>' +
      // R-UX-10 二次收紧（2026-11-28）：配置详情表单（studio A–F 节+YAML 手改）默认不展示——
      // DOM 整体挂隐藏位，「配置详情」按钮点击移入弹窗（绑定随节点保留，关闭归位）。
      '<div data-wf-studio-holder style="display:none">' + workflowStudioHtml(wf, usedName) + '</div>' +
      (tasksHtml || (runs.length ? '<div class="small faint">' + esc(tr('运行未关联 Task 对象（触发源=Proposal/Template/其他——见运行 involved）。')) + '</div>' : '')) +
      (runsHtml || '<div class="card"><div class="card-b"><div class="small faint">' +
        esc(tr('该 workflow 暂无运行记录——')) + (exec.state === 'unbound' ? esc(tr('执行面未绑定（ConnectorOps 骨架缺席），任务不会拉起 run。')) : '') +
        '</div></div></div>');
    bindWfNodeGraph(el, runs, taskIdx, usedName, sessionRows);
    bindWorkflowStudio(el, ctx, wf, usedName, () => detailScreen(el, ctx, usedName));
    const yamlBtn = el.querySelector('[data-wf-yaml-btn]');
    if (yamlBtn) yamlBtn.addEventListener('click', () => {
      htmlModal({ title: tr('YAML 声明 · ') + usedName, wide: true, html: yamlTextView(String(wf.rawYaml || '')) });
    });
    // 配置详情=studio 移动式弹窗（手改最后手段——表单/YAML 手填都在弹窗内，默认不占页面）。
    const cfgBtn = el.querySelector('[data-wf-config-btn]');
    const studioHolder = el.querySelector('[data-wf-studio-holder]');
    if (cfgBtn && studioHolder) cfgBtn.addEventListener('click', () => {
      const studio = studioHolder.querySelector('[data-wf-studio]');
      if (!studio) { toast(tr('配置详情缺席（该 workflow 无可配置 studio）'), 'error'); return; }
      htmlModal({ title: tr('配置详情 · ') + usedName, wide: true, node: studio, holder: studioHolder });
    });
    // 对话修改=主通路（R-UX-10 二次裁决：不跳转——原地嵌入面板，rawYaml 为基准，Apply 走对象 rv 锁）。
    const chatBtn = el.querySelector('[data-wf-chatfix]');
    if (chatBtn) chatBtn.addEventListener('click', () => {
      openDocChat({
        title: tr('对话修改 · Workflow/') + usedName,
        refKey: 'obj|Workflow|' + String(wf.scope || 'global') + '|' + usedName,
        ref: { kind: 'Workflow', scope: String(wf.scope || 'global'), name: usedName, rv: Number(wf.resourceVersion) || 0 },
        current: String(wf.rawYaml || ''),
        ctx,
        onApplied: () => detailScreen(el, ctx, usedName),
      });
    });
  })().catch((e) => {
    if (!el.isConnected || el.getAttribute('data-wf-detail-stamp') !== stamp) return;
    const detail = el.querySelector('[data-wf-detail]');
    if (detail) detail.innerHTML = '<div class="card"><div class="card-b"><div class="small faint">' +
      esc(tr('工作流现读失败：') + String((e && e.message) || e)) + '</div></div></div>';
  });
}

/** 运行独立页（#/wf/:name/r/:n）——人视角布局：
 *  hero（名+run+状态+时间+审批动作锚）→ 真连线 stepper（视觉中心）→ 双栏（当前相位卡 | 事件时间线）。
 *  空态全部内联弱化（不裸句挂底）。 */
/** 区内 ref-chip 的 data-go 点击（runScreen 局部——不走全仓委托）。 */
function bindDocLinks(scopeEl, ctx) {
  scopeEl.querySelectorAll('[data-go]').forEach((n) => {
    n.addEventListener('click', () => ctx.go(n.getAttribute('data-go')));
  });
}

function runScreen(el, ctx, name, no) {
  const runs = ctx.store.runs().filter((r) => String(r && (r.workflow || r.name || '')) === String(name));
  const run = runs.find((r) => String(r.run != null ? r.run : (r.no != null ? r.no : '')) === String(no)) || runs[0];
  if (!run) {
    el.innerHTML = '<div class="view-inner">' +
      '<div class="run-hero"><div class="rh-top">' + chipGo(tr('← ' + String(name)), 'wf/' + encodeURIComponent(String(name)), 'mute', false) + '</div>' +
      '<div class="rh-title">' + esc(String(name)) + ' <span class="rh-run">#' + esc(String(no)) + '</span></div></div>' +
      emptyBox(tr('运行记录不在册'), tr('暂无此运行记录（可能已归档）。')) + '</div>';
    return;
  }
  let st = String(run.status || '');
  let stCls = st === 'suspended' ? 'warn' : (RUN_CLS[st] || 'mute');
  const startedAt = run.startedAt || run.at || run.ts;
  const gateWord = st === 'suspended' ? tr('等审批')
    : (st === 'running' ? tr('运行中')
      : (st === 'converged' ? tr('已收束') : (st === 'failed' ? tr('已驳回') : st)));
  /* 批闸（§19 R3——run 恢复唯一人工通路=对 Workflow 锚发布 ApprovalGranted/ApprovalRejected；
   * scope=run.owner||'global'；无闸锚词形不显示按钮=诚实）。 */
  const canGate = String(run.phase || '') === 'AwaitingApproval' || st === 'suspended';
  const gateAnchor = 'Workflow/' + String(run.owner || 'global') + '/' + String(run.workflow || name);
  el.innerHTML = '<div class="view-inner">' +
    /* hero */
    '<div class="run-hero">' +
    '<div class="rh-top">' + chipGo(tr('← ' + String(name)), 'wf/' + encodeURIComponent(String(name)), 'mute', false) + '</div>' +
    '<div class="rh-title">' + esc(String(name)) + ' <span class="rh-run">#' + esc(String(run.run != null ? run.run : no)) + '</span>' +
    chip(st || tr('未知'), stCls) + '</div>' +
    '<div class="rh-meta">' +
    (startedAt ? '<span>' + esc(tr('开始于')) + ' ' + esc(relAge(startedAt) || String(startedAt)) + '</span>' : '') +
    (run.error ? '<span class="rh-err">' + esc(trunc(String(run.error), 120)) + '</span>' : '') +
    (st === 'suspended' ? '<span class="ref-chip" data-go="approvals">' + esc(tr('去审批 →')) + '</span>' : '') +
    '</div>' +
    (canGate
      ? '<div class="dh-actions" style="margin-top:10px">' +
        '<button class="btn btn-primary sm" data-gate-act="grant">' + esc(tr('批准继续')) + '</button> ' +
        '<button class="btn btn-ghost btn-danger sm" data-gate-act="reject">' + esc(tr('驳回终止')) + '</button> ' +
        '<span class="small faint mono">' + esc(tr('批闸锚：')) + esc(gateAnchor) + '</span></div>'
      : '') +
    '</div>' +
    /* stepper（视觉中心） */
    '<div class="step-line" data-step-line></div>' +
    '<div class="card" style="margin-bottom:16px"><div class="card-h"><span class="h-title">' + esc(tr('会话')) + '</span></div>' +
    '<div class="card-b" data-run-session>' + wfSessionChip(run.sessionId) + '</div></div>' +
    '<div class="card" style="margin-bottom:16px"><div class="card-h"><span class="h-title">' + esc(tr('节点会话')) + '</span></div>' +
    '<div class="card-b" data-run-node-sessions><div class="small faint">' + esc(tr('节点会话现读中…')) + '</div></div></div>' +
    /* 双栏：当前相位 | 事件时间线 */
    '<div class="run-cols">' +
    '<div class="run-col"><div class="panel-h">' + esc(tr('当前相位')) + '</div><div data-run-now></div></div>' +
    '<div class="run-col"><div class="panel-h">' + esc(tr('事件时间线')) + '</div><div data-run-tl></div></div>' +
    '</div>' +
    '<div class="foot-note">' + esc(tr('流程按声明阶段推进；' + gateWord + '。')) + '</div>' +
    '</div>';

  /* stepper 渲染+选中态（phases 每次现读 run——注水后重绘全序列） */
  const sel = { idx: 0, phases: [] };
  const lineEl = el.querySelector('[data-step-line]');
  const paintLine = () => {
    const phases = Array.isArray(run.phases) && run.phases.length ? run.phases : [run.phase || tr('单节点')];
    sel.phases = phases;
    const curIdx = phases.indexOf(run.phase);
    if (curIdx >= 0) sel.idx = curIdx;
    // F-04：标绿判据=history 运行轨迹（真实走过的步）——按下标盲标绿的缺陷（失败 run
    // 全绿）收口；轨迹缺席=只亮当前相位（不冒充走过）。
    const visited = new Set((Array.isArray(run.history) ? run.history : [])
      .map((h) => String((h && h.to) || '')).filter(Boolean));
    const isTerminal = st === 'done' || st === 'converged';
    lineEl.innerHTML = phases.map((p, i) => {
      const done = visited.has(p) && (p !== run.phase || isTerminal);
      const active = i === curIdx && !isTerminal;
      const dot = done ? '✓' : (active ? '' : String(i + 1));
      const conn = i < phases.length - 1
        ? '<div class="step-conn' + (done || (curIdx >= 0 && i < curIdx) ? ' fill' : '') + '"></div>' : '';
      return '<div class="step' + (done ? ' done' : '') + (active ? ' active' : '') + (p === 'Failed' ? ' failstep' : '') +
        '" data-step="' + i + '" role="button">' +
        '<div class="step-dot">' + dot + '</div><div class="step-name">' + esc(p) + '</div></div>' + conn;
    }).join('');
    el.querySelectorAll('[data-step]').forEach((n) => {
      n.addEventListener('click', () => { sel.idx = Number(n.getAttribute('data-step')); paintNow(); paintSel(); });
    });
    paintSel();
  };
  const paintSel = () => {
    el.querySelectorAll('[data-step]').forEach((n, i) => n.classList.toggle('sel', i === sel.idx));
  };

  /* 当前相位卡+事件时间线 */
  const nowEl = el.querySelector('[data-run-now]');
  const tlEl = el.querySelector('[data-run-tl]');
  const paintNow = () => {
    const phases = sel.phases.length ? sel.phases : [run.phase || tr('单节点')];
    const ph = phases[sel.idx] || run.phase || tr('单节点');
    const isCur = ph === run.phase;
    const involved = String(run.involved || '');
    const refs = involved.split(',').map((x) => x.trim()).filter((x) => x.split('/').length >= 3)
      .map((x) => {
        const p2 = x.split('/');
        return '<span class="ref-chip" data-go="obj/' + esc(p2[0]) + '/' + esc(p2[1]) + '/' + esc(p2[2]) + '">' + esc(x) + '</span>';
      }).join(' ');
    nowEl.innerHTML =
      '<div class="now-ph">' + esc(ph) + (isCur ? chip(gateWord || st || tr('进行中'), stCls) : '') + '</div>' +
      '<div class="small faint">' + esc(isCur
        ? (isTerminalRun(st) ? tr('该相位已收束——单次运行结束') : tr('进行中——状态实时刷新'))
        : tr('历史相位——按运行轨迹回放')) + '</div>' +
      (refs ? '<div class="now-refs">' + esc(tr('涉及')) + '：' + refs + '</div>' : '') +
      (isCur && st === 'suspended' ? '<div class="now-act"><span class="ref-chip" data-go="approvals">' + esc(tr('去处理这条审批')) + '</span></div>' : '');
    bindDocLinks(nowEl, ctx);
  };
  const paintTl = () => {
    /* F-04：时间线=history 运行轨迹（真实迁移行含时刻）；静态 transitions 不再当事件时间线展示
     *（轨迹缺席=不画不编造）。下方仍为归档窗口事件链。 */
    const trans = Array.isArray(run.history) ? run.history.filter(Boolean) : [];
    const transHtml = trans.length
      ? '<div class="small faint" style="margin-bottom:6px">' + esc(tr('阶段推进记录：')) + '</div><div class="tl" style="margin-bottom:10px">' +
        trans.slice(0, 20).map((x) => {
          const at = String((x && (x.at || x.ts || x.time)) || '');
          const from = String((x && (x.from || x.fromPhase)) || '—');
          const to = String((x && (x.to || x.toPhase || x.phase)) || '—');
          return '<div class="tl-row"><div class="tl-dot"></div><div class="tl-main"><b>' + esc(from) + ' → ' + esc(to) + '</b></div>' +
            '<div class="tl-time mono">' + esc(relAge(at) || (at ? at.slice(0, 16) : '')) + '</div></div>';
        }).join('') + '</div>'
      : '';
    const evs = ctx.store.events().filter((e) => {
      const hay = JSON.stringify([e.dedupKey, e.involved, e.type]);
      return hay.includes(String(run.workflow || '')) || hay.includes(String(run.phase || ''));
    }).slice(0, 12);
    tlEl.innerHTML = transHtml + (evs.length
      ? '<div class="tl">' + evs.map((e) => {
        const eid = e && (e.id != null ? e.id : (e.eventId != null ? e.eventId : ''));
        const goAttr = eid !== '' ? ' data-go="event/' + encodeURIComponent(String(eid)) + '" role="link" style="cursor:pointer"' : '';
        return '<div class="tl-row"' + goAttr + '><div class="tl-dot"></div><div class="tl-main"><b>' + esc(String(e.type || '')) + '</b>' +
          '<div class="mr-sub small mono">' + esc(trunc(String(e.involved || '—'), 56)) + '</div></div>' +
          '<div class="tl-time">' + esc(relAge(e.ts) || '') + '</div></div>';
      }).join('') + '</div>'
      : '<div class="tl-empty">' + esc(tr('暂无相关事件——流程推进时此处实时更新。')) + '</div>');
  };
  paintLine(); paintNow(); paintTl();
  /* 节点会话：Task 列表 + journal taskId 反查（真实 session-*）；读失败=会话未记录，不编造。 */
  let nodeSessionRows = [];
  let nodeTask = null;
  const hydrateNodeSessions = async () => {
    const box = el.querySelector('[data-run-node-sessions]');
    if (!box || !el.isConnected) return;
    const key = wfTaskKey(run.involved);
    if (!key) { box.innerHTML = '<div class="small faint">' + esc(tr('该运行未关联 Task 对象——节点会话未记录。')) + '</div>'; return; }
    try {
      const [tasksRes, sessionsRes] = await Promise.all([
        ctx.api.tasks(),
        ctx.api.taskSessions ? ctx.api.taskSessions([key]) : Promise.resolve({ ok: false }),
      ]);
      if (!el.isConnected) return;
      const tasks = wfApiTasks(tasksRes);
      nodeTask = tasks.find((t) => String(t.scope || 'global') + '/' + String(t.name || '') === key) || null;
      nodeSessionRows = sessionsRes && sessionsRes.ok && Array.isArray(sessionsRes.data && sessionsRes.data.items) ? sessionsRes.data.items : [];
      box.innerHTML = wfRunNodesHtml(run, nodeTask, nodeSessionRows);
      const topAnchor = wfSessionAnchor(run, nodeTask, nodeSessionRows);
      const sessionBox = el.querySelector('[data-run-session]');
      if (sessionBox) sessionBox.innerHTML = wfSessionChip(topAnchor.sid, topAnchor.label);
    } catch {
      if (el.isConnected) box.innerHTML = '<div class="small faint">' + esc(tr('节点会话读取失败——不冒充已记录。')) + '</div>';
    }
  };
  hydrateNodeSessions();
  /* runs face 现读重注水（phases/phase/status/involved/transitions——字段缺席=保留原值不编造） */
  const refreshFace = () => ctx.api.workflowRuns().then((r) => {
    if (!r.ok || !el.isConnected) return;
    const arr = Array.isArray(r.data && r.data.runs) ? r.data.runs : (Array.isArray(r.data) ? r.data : []);
    const full = arr.find((x) => String(x && (x.run != null ? x.run : '')) === String(no));
    if (!full) return;
    run.phases = Array.isArray(full.phases) ? full.phases : run.phases;
    run.phase = full.phase || run.phase;
    run.status = full.status || run.status;
    st = String(run.status || '');
    stCls = st === 'suspended' ? 'warn' : (RUN_CLS[st] || 'mute');
    run.sessionId = full.sessionId || run.sessionId;
    run.involved = full.involved || run.involved;
    run.transitions = Array.isArray(full.transitions) ? full.transitions : run.transitions;
    run.history = Array.isArray(full.history) ? full.history : run.history;
    const sessionBox = el.querySelector('[data-run-session]');
    if (sessionBox) sessionBox.innerHTML = wfSessionChip(run.sessionId);
    const nodeBox = el.querySelector('[data-run-node-sessions]');
    if (nodeBox && wfTaskKey(run.involved)) nodeBox.innerHTML = wfRunNodesHtml(run, nodeTask, nodeSessionRows);
    paintLine(); paintNow(); paintTl();
  });
  refreshFace();
  /* 批闸动作（hero 按钮——confirmModal 写明副作用；失败=显式 toast 不装成功；成功=runs 面重注水+快照重读。
   * 2026-11-28 弹窗裁决——style.md §3.3：原生 confirm/alert 禁用） */
  el.querySelectorAll('[data-gate-act]').forEach((btn) => {
    btn.addEventListener('click', async () => {
      const grant = btn.getAttribute('data-gate-act') === 'grant';
      const confirmed = await confirmModal({
        title: grant ? tr('批准继续') : tr('驳回终止'),
        body: esc(grant
          ? tr('确认批准继续？发布 ApprovalGranted（') + gateAnchor + tr('）——run 沿声明相位序列续行。')
          : tr('确认驳回终止？发布 ApprovalRejected（') + gateAnchor + tr('）——run 将进入终态 Failed（不可恢复）。')),
        confirmText: grant ? tr('批准') : tr('驳回'),
        danger: !grant,
      });
      if (!confirmed) return;
      btn.disabled = true;
      btn.textContent = tr('发布中…');
      const runNo = Number(no) || 0;
      const r = await ctx.api.eventPub({ type: grant ? 'ApprovalGranted' : 'ApprovalRejected', involved: gateAnchor, ...(runNo > 0 ? { run: runNo } : {}) });
      if (!r.ok) {
        btn.disabled = false;
        btn.textContent = grant ? tr('批准继续') : tr('驳回终止');
        toast(esc(tr('事件发布失败（') + String(r.error || r.status) + tr('）——run 仍挂原相位，未装成功。')), 'error');
        return;
      }
      refreshFace();
      ctx.api.snapshot().then((s) => { if (s.ok) ctx.store.init(s.data); });
    });
  });
}

export function render(el, ctx) {
  const p = ctx.params || {};
  if (p.n != null && p.n !== '') runScreen(el, ctx, p.name, p.n);
  else if (p.name != null && p.name !== '') detailScreen(el, ctx, p.name);
  else listScreen(el, ctx);
}
