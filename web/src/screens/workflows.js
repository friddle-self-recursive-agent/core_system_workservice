/* workflows.js —— 工作流三区列表（§19.3：在册注册表=GET /v1/workflows 现读 ∪ 运行记录=
 * GET /v1/workflows/runs 现读+快照 runs 聚合并显去重 ∪ 模板三卡（§9b 静态卡，实例化置灰））
 * + 详情（概览/节点图）+ 运行独立页（hero 批闸按钮=ApprovalGranted/ApprovalRejected 事件发布）。
 * 快照不带 workflows 键=旧「假空态」根因——注册表/runs 一律走活体 GET 面，失败=诚实降级词形。
 * wfAnchorOfRoute：route→workflow 锚映射（'workflows/x' 剥前缀；'-workflow$' 兜底
 * 'core-workflow'——语义移植旧 console-app 同名函数）。 */

import { esc, trunc, timeHtml, relAge } from '../core/format.js';
import { chip, chipGo, kv, emptyBox, screenHead, kindBadge, involvedRoute, repoLinkChip, objRoute, docRoute } from '../render/chips.js';
import { depChips } from './tasks.js';
import { t as tr } from '../core/i18n.js';

/** route → workflow 锚（不匹配=空串——不编造锚）。 */
export function wfAnchorOfRoute(route) {
  const r = String(route || '');
  if (!r) return '';
  if (r.startsWith('workflows/')) return r.slice('workflows/'.length);
  if (/-workflow$/.test(r)) return 'core-workflow';
  return '';
}


/* 相位管线节点图（数据=run.phases/transitions/phase——声明序列原样；缺席=单节点诚实降级）。
 * F-04：节点标色按 run.history 运行轨迹（真实走过的步），不按下标盲标绿——失败/挂起 run
 * 不再把未执行节点标绿。节点点击 → 下方抽屉（该相位关联事件链+involved 文档 chips）。 */
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
    ? '<div class="card" style="margin-bottom:10px"><div class="card-h"><span class="h-title mono">bug-lifecycle</span>' +
      '<span class="h-right">' + chip('Workflow', 'acc') + ' ' + chip(bugRuns.length + tr(' 次运行'), bugRuns.length ? 'acc' : 'mute') + '</span></div>' +
      '<div class="card-b"><div class="small">' + esc(tr('缺陷闭环真实 Workflow 对象（BugReport→临时 BugDoc→验证→关闭）。')) + '</div>' +
      '<div class="dh-actions" style="margin-top:8px">' +
      chipGo(tr('工作流详情'), 'wf/bug-lifecycle', 'acc', true) + ' ' +
      chipGo(tr('文档库 BugDoc'), 'docs', 'mute', true) +
      '</div></div></div>'
    : '';
  const tplBlock = tpls.length
    ? '<div class="grid cols-2">' + tpls.map(realTplCardHtml).join('') + '</div>'
    : '<div class="card"><div class="card-b"><div class="small faint">' +
      esc(tr('暂无流程模板（Template 对象在册后此处显示）。')) + ' ' +
      chipGo(tr('设计见 workflow.md §9b'), wfDesignRoute(), 'mute', true, tr('设计文档工作台')) +
      '</div></div></div>';
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

/** 详情（概览 + 该 workflow 的运行清单）。 */
function detailScreen(el, ctx, name) {
  const wfs = ctx.store.workflows();
  const runs = ctx.store.runs().filter((r) => String(r && (r.workflow || r.name || '')) === String(name));
  const wf = wfs.find((w) => String(w && (w.name || w.workflow || '')) === String(name));
  const head = screenHead(tr('工作流详情'), 'name=' + String(name), chipGo(tr('← 工作流'), 'workflows', 'mute', false));
  if (!wf && !runs.length) {
    el.innerHTML = '<div class="view-inner">' + head + emptyBox(tr('工作流不在册'), tr('暂无此工作流（可能已下线或名称有误）。')) + '</div>';
    return;
  }
  const rows = [];
  if (wf) {
    if (wf.scope) rows.push(['scope', esc(String(wf.scope))]);
    if (wf.resourceVersion != null) rows.push(['rv', esc(String(wf.resourceVersion))]);
    if (wf.spec && wf.spec.title) rows.push([tr('标题'), esc(String(wf.spec.title))]);
    if (wf.spec && wf.spec.goalRef) rows.push([tr('目标'), chipGo(String(wf.spec.goalRef), involvedRoute(String(wf.spec.goalRef)) || 'docs', 'mute', true, tr('目标文档'))]);
    // 代码 ref 直达仓库（pkg://<org>/<repo>/<path> / git:// 等——不匹配=不显示）。
    const impl = (wf.spec && wf.spec.impl) || (wf.execution && wf.execution.impl) || '';
    const repoChip = impl ? repoLinkChip(String(impl), String(impl)) : '';
    if (repoChip) rows.push([tr('实现仓库'), repoChip]);
  }
  if (wf && wf.spec && wf.spec.dependsOn) rows.push([tr('依赖'), depChips(wf.spec.dependsOn, 'workflow')]);
  rows.push([tr('运行'), chip(runs.length + tr(' 次'), runs.length ? 'acc' : 'mute')]);
  const runRows = runs.map((r, i) => {
    const no = r.run != null ? r.run : (r.no != null ? r.no : (runs.length - i));
    return '<div class="mini-row" data-go="wf/' + encodeURIComponent(String(name)) + '/r/' + encodeURIComponent(String(no)) + '">' +
      '<div class="mr-main"><div><b class="mono">run #' + esc(String(no)) + '</b> ' + runChip(r) + '</div>' +
      '<div class="mr-sub small mono">' + esc(trunc(String(r.route || r.trigger || '—'), 80)) + '</div></div>' +
      '<span class="w-ev-time mono">' + esc(relAge(r.startedAt || r.at || r.ts) || '—') + '</span>' +
      '</div>';
  }).join('');
  /* 节点图数据=最近一条 run（快照聚合缺席=诚实单节点降级——不编造相位序列）。
   * （修复：此处原引用未定义的 run——ReferenceError 中断整屏渲染=详情页假死。） */
  const run0 = runs[0] || { workflow: String(name), phases: [], phase: '', status: '' };
  el.innerHTML = '<div class="view-inner">' + head +
    tr('<div class="card"><div class="card-h"><span class="h-title">概览</span></div><div class="card-b">') + kv(rows) + '</div></div>' +
    tr('<h2 class="sec-title">节点图</h2><div class="card"><div class="card-b" data-pg-box>') +
    (phaseGraphHtml(run0, ctx)) + '</div></div>' +
    tr('<h2 class="sec-title">运行记录</h2>') +
    (runRows ? '<div class="card">' + runRows + '</div>' : tr('<div class="small faint">该 workflow 暂无运行记录。</div>')) +
    '</div>';
  bindPhaseGraph(el, ctx, run0);
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
  const st = String(run.status || '');
  const stCls = st === 'suspended' ? 'warn' : (RUN_CLS[st] || 'mute');
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
  /* runs face 现读重注水（phases/phase/status/involved/transitions——字段缺席=保留原值不编造） */
  const refreshFace = () => ctx.api.workflowRuns().then((r) => {
    if (!r.ok || !el.isConnected) return;
    const arr = Array.isArray(r.data && r.data.runs) ? r.data.runs : (Array.isArray(r.data) ? r.data : []);
    const full = arr.find((x) => String(x && (x.run != null ? x.run : '')) === String(no));
    if (!full) return;
    run.phases = Array.isArray(full.phases) ? full.phases : run.phases;
    run.phase = full.phase || run.phase;
    run.status = full.status || run.status;
    run.involved = full.involved || run.involved;
    run.transitions = Array.isArray(full.transitions) ? full.transitions : run.transitions;
    run.history = Array.isArray(full.history) ? full.history : run.history;
    paintLine(); paintNow(); paintTl();
  });
  refreshFace();
  /* 批闸动作（hero 按钮——confirm 写明副作用；失败=显式 alert 不装成功；成功=runs 面重注水+快照重读） */
  el.querySelectorAll('[data-gate-act]').forEach((btn) => {
    btn.addEventListener('click', async () => {
      const grant = btn.getAttribute('data-gate-act') === 'grant';
      const confirmed = grant
        ? window.confirm(tr('确认批准继续？发布 ApprovalGranted（') + gateAnchor + tr('）——run 沿声明相位序列续行。'))
        : window.confirm(tr('确认驳回终止？发布 ApprovalRejected（') + gateAnchor + tr('）——run 将进入终态 Failed（不可恢复）。'));
      if (!confirmed) return;
      btn.disabled = true;
      btn.textContent = tr('发布中…');
      const runNo = Number(no) || 0;
      const r = await ctx.api.eventPub({ type: grant ? 'ApprovalGranted' : 'ApprovalRejected', involved: gateAnchor, ...(runNo > 0 ? { run: runNo } : {}) });
      if (!r.ok) {
        btn.disabled = false;
        btn.textContent = grant ? tr('批准继续') : tr('驳回终止');
        window.alert(tr('事件发布失败（') + String(r.error || r.status) + tr('）——run 仍挂原相位，未装成功。'));
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
