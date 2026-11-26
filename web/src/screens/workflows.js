/* workflows.js —— 工作流列表 + 详情（概览/运行）+ 运行独立页。数据=快照 workflows/runs
 * 聚合（store 投影——patch 增量实时推进）；契约未携带的字段不显示不编造。
 * wfAnchorOfRoute：route→workflow 锚映射（'workflows/x' 剥前缀；'-workflow$' 兜底
 * 'core-workflow'——语义移植旧 console-app 同名函数）。 */

import { esc, trunc, timeHtml, relAge } from '../core/format.js';
import { chip, chipGo, kv, emptyBox, screenHead, kindBadge } from '../render/chips.js';
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
 * 节点点击 → 下方抽屉（该相位关联事件链+involved 文档 chips——日志=R归档窗口内事实）。 */
function phaseGraphHtml(run, ctx) {
  const phases = Array.isArray(run.phases) ? run.phases : [];
  const cur = String(run.phase || '');
  const st = String(run.status || '');
  const nodeHtml = (p, i) => {
    const active = p === cur;
    const done = st === 'done' || (cur && phases.indexOf(cur) > i);
    const cls = 'pg-node' + (active ? ' active' : '') + (done ? ' done' : '') + (p === 'Failed' ? ' fail' : '');
    return '<div class="' + cls + '" data-pg-node="' + esc(p) + '"><div class="pg-dot"></div>' +
      '<div class="pg-name">' + esc(p) + '</div>' + (active ? chip(st || tr('进行中'), 'acc') : '') + '</div>';
  };
  const body = phases.length
    ? phases.map((p, i) => nodeHtml(p, i)).join('<span class="pg-edge">→</span>')
    : '<div class="pg-node active"><div class="pg-dot"></div><div class="pg-name">' + esc(cur || tr('单节点（声明未投影相位）')) + '</div>' + runChip(run) + '</div>';
  return '<div class="pg-wrap">' + body + '</div>' +
    '<div class="small faint">' + esc(tr('点击节点看该相位的事件链与涉及文档；推进由 patch 增量实时刷新。')) + '</div>' +
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
        ? evs.map((e) => '<div class="mini-row"><div class="mr-main"><div><b>' + esc(String(e.type || '')) + '</b></div>' +
          '<div class="mr-sub small mono">' + esc(trunc(String(e.involved || '—'), 60)) + '</div></div>' +
          '<span class="small faint">' + esc(relAge(e.ts) || '') + '</span></div>').join('')
        : '<div class="small faint">' + esc(tr('窗口内无该相位事件（诚实空态）。')) + '</div>';
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

const RUN_CLS = { running: 'acc', done: 'ok', ok: 'ok', success: 'ok', failed: 'esc', error: 'esc', canceled: 'mute', queued: 'mute', pending: 'mute' };

function runChip(r) {
  const st = String((r && (r.status || r.state || r.phase)) || tr('未知'));
  return chip(st, RUN_CLS[st] || 'mute');
}

/** 列表（workflows 聚合 + runs 摘要）。 */
function listScreen(el, ctx) {
  const wfs = ctx.store.workflows();
  const runs = ctx.store.runs();
  const rows = wfs.map((w) => {
    const name = String(w && (w.name || w.workflow || ''));
    if (!name) return '';
    const myRuns = runs.filter((r) => String(r && (r.workflow || r.name || '')) === name);
    const last = myRuns[0];
    return '<div class="mini-row" data-go="wf/' + encodeURIComponent(name) + '">' +
      '<div class="mr-main"><div><b class="mono">' + esc(name) + '</b> ' + (w.spec ? kindBadge('Workflow') : '') + '</div>' +
      '<div class="mr-sub small">' + esc(trunc(String(w.summary || w.desc || (w.spec && w.spec.title) || '—'), 90)) + '</div></div>' +
      '<span class="ws-chip">' + chip(myRuns.length + tr(' 次运行'), myRuns.length ? 'acc' : 'mute') + (last ? runChip(last) : '') + '</span>' +
      '</div>';
  }).filter(Boolean);
  el.innerHTML = '<div class="view-inner">' + screenHead(tr('工作流'), tr('聚合面（快照 workflows/runs）；运行独立页见详情')) +
    (rows.length
      ? '<div class="card">' + rows.join('') + '</div>'
      : emptyBox(tr('暂无在册工作流'), tr('快照 workflows 聚合为空（不装假）。'))) +
    '</div>';
  el.querySelectorAll('.mini-row[data-go]').forEach((row) => {
    row.addEventListener('click', () => ctx.go(row.dataset.go));
  });
}

/** 详情（概览 + 该 workflow 的运行清单）。 */
function detailScreen(el, ctx, name) {
  const wfs = ctx.store.workflows();
  const runs = ctx.store.runs().filter((r) => String(r && (r.workflow || r.name || '')) === String(name));
  const wf = wfs.find((w) => String(w && (w.name || w.workflow || '')) === String(name));
  const head = screenHead(tr('工作流详情'), 'name=' + String(name), chipGo(tr('← 工作流'), 'workflows', 'mute', false));
  if (!wf && !runs.length) {
    el.innerHTML = '<div class="view-inner">' + head + emptyBox(tr('工作流不在册'), tr('快照聚合无此 workflow（可能已下线或名有误）。')) + '</div>';
    return;
  }
  const rows = [];
  if (wf) {
    if (wf.scope) rows.push(['scope', esc(String(wf.scope))]);
    if (wf.resourceVersion != null) rows.push(['rv', esc(String(wf.resourceVersion))]);
    if (wf.spec && wf.spec.title) rows.push([tr('标题'), esc(String(wf.spec.title))]);
    if (wf.spec && wf.spec.goalRef) rows.push([tr('目标'), chipGo(String(wf.spec.goalRef), involvedRoute(String(wf.spec.goalRef)) || 'docs', 'mute', true, tr('目标文档'))]);
  }
  if (wf && wf.spec && wf.spec.dependsOn) rows.push([tr('依赖'), (wf.spec.dependsOn || []).map((d) => chip(String(d), 'mute')).join(' ')]);
  rows.push([tr('运行'), chip(runs.length + tr(' 次'), runs.length ? 'acc' : 'mute')]);
  const runRows = runs.map((r, i) => {
    const no = r.no != null ? r.no : (runs.length - i);
    return '<div class="mini-row" data-go="wf/' + encodeURIComponent(String(name)) + '/r/' + encodeURIComponent(String(no)) + '">' +
      '<div class="mr-main"><div><b class="mono">run #' + esc(String(no)) + '</b> ' + runChip(r) + '</div>' +
      '<div class="mr-sub small mono">' + esc(trunc(String(r.route || r.trigger || '—'), 80)) + '</div></div>' +
      '<span class="w-ev-time mono">' + esc(relAge(r.startedAt || r.at || r.ts) || '—') + '</span>' +
      '</div>';
  }).join('');
  el.innerHTML = '<div class="view-inner">' + head +
    tr('<div class="card"><div class="card-h"><span class="h-title">概览</span></div><div class="card-b">') + kv(rows) + '</div></div>' +
    tr('<h2 class="sec-title">节点图</h2><div class="card"><div class="card-b" data-pg-box>') +
    (phaseGraphHtml(run, ctx)) + '</div></div>' +
    tr('<h2 class="sec-title">运行记录</h2>') +
    (runRows ? '<div class="card">' + runRows + '</div>' : tr('<div class="small faint">该 workflow 暂无运行记录（空态显式）。</div>')) +
    '</div>';
  el.querySelectorAll('.mini-row[data-go]').forEach((row) => {
    row.addEventListener('click', () => ctx.go(row.dataset.go));
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
      emptyBox(tr('运行记录不在册'), tr('快照 runs 聚合无此运行（可能已出窗）。')) + '</div>';
    return;
  }
  const st = String(run.status || '');
  const stCls = st === 'suspended' ? 'warn' : (RUN_CLS[st] || 'mute');
  const startedAt = run.startedAt || run.at || run.ts;
  const gateWord = st === 'suspended' ? tr('等审批') : (st === 'running' ? tr('运行中') : st);
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
    '</div></div>' +
    /* stepper（视觉中心） */
    '<div class="step-line" data-step-line></div>' +
    /* 双栏：当前相位 | 事件时间线 */
    '<div class="run-cols">' +
    '<div class="run-col"><div class="panel-h">' + esc(tr('当前相位')) + '</div><div data-run-now></div></div>' +
    '<div class="run-col"><div class="panel-h">' + esc(tr('事件时间线')) + '</div><div data-run-tl></div></div>' +
    '</div>' +
    '<div class="foot-note">' + esc(tr('相位序列=Workflow 声明原样；推进由 patch 增量实时刷新；' + gateWord + '。')) + '</div>' +
    '</div>';

  /* stepper 渲染+选中态 */
  const phases = Array.isArray(run.phases) && run.phases.length ? run.phases : [run.phase || tr('单节点')];
  const curIdx = phases.indexOf(run.phase);
  const sel = { idx: curIdx >= 0 ? curIdx : 0 };
  const lineEl = el.querySelector('[data-step-line]');
  const paintLine = () => {
    lineEl.innerHTML = phases.map((p, i) => {
      const done = st === 'done' || (curIdx >= 0 ? i < curIdx : false);
      const active = i === curIdx && st !== 'done';
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
    const ph = phases[sel.idx];
    const isCur = ph === run.phase;
    const involved = String(run.involved || '');
    const refs = involved.split(',').map((x) => x.trim()).filter((x) => x.split('/').length >= 3)
      .map((x) => {
        const p2 = x.split('/');
        return '<span class="ref-chip" data-go="obj/' + esc(p2[0]) + '/' + esc(p2[1]) + '/' + esc(p2[2]) + '">' + esc(x) + '</span>';
      }).join(' ');
    nowEl.innerHTML =
      '<div class="now-ph">' + esc(ph) + (isCur ? chip(gateWord || st || tr('进行中'), stCls) : '') + '</div>' +
      '<div class="small faint">' + esc(isCur ? tr('正在发生——推进由 patch 实时刷新') : tr('历史相位——事件按归档窗口回放')) + '</div>' +
      (refs ? '<div class="now-refs">' + esc(tr('涉及')) + '：' + refs + '</div>' : '') +
      (isCur && st === 'suspended' ? '<div class="now-act"><span class="ref-chip" data-go="approvals">' + esc(tr('去处理这条审批')) + '</span></div>' : '');
    bindDocLinks(nowEl, ctx);
  };
  const paintTl = () => {
    const evs = ctx.store.events().filter((e) => {
      const hay = JSON.stringify([e.dedupKey, e.involved, e.type]);
      return hay.includes(String(run.workflow || '')) || hay.includes(String(run.phase || ''));
    }).slice(0, 12);
    tlEl.innerHTML = evs.length
      ? '<div class="tl">' + evs.map((e) =>
        '<div class="tl-row"><div class="tl-dot"></div><div class="tl-main"><b>' + esc(String(e.type || '')) + '</b>' +
        '<div class="mr-sub small mono">' + esc(trunc(String(e.involved || '—'), 56)) + '</div></div>' +
        '<div class="tl-time">' + esc(relAge(e.ts) || '') + '</div></div>').join('') + '</div>'
      : '<div class="tl-empty">' + esc(tr('归档窗口内暂无事件——workflow 推进时此处实时落行。')) + '</div>';
  };
  paintLine(); paintNow(); paintTl();
  ctx.api.workflowRuns().then((r) => {
    if (!r.ok || !el.isConnected) return;
    const arr = Array.isArray(r.data && r.data.runs) ? r.data.runs : (Array.isArray(r.data) ? r.data : []);
    const full = arr.find((x) => String(x && (x.run != null ? x.run : '')) === String(no));
    if (!full || !Array.isArray(full.phases) || !full.phases.length) return;
    const lineEl2 = el.querySelector('[data-step-line]');
    if (lineEl2) { run.phases = full.phases; run.phase = full.phase; run.status = full.status || run.status; run.involved = full.involved || run.involved; }
  });
}

export function render(el, ctx) {
  const p = ctx.params || {};
  if (p.n != null && p.n !== '') runScreen(el, ctx, p.name, p.n);
  else if (p.name != null && p.name !== '') detailScreen(el, ctx, p.name);
  else listScreen(el, ctx);
}
