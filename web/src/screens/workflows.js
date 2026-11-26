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
    tr('<h2 class="sec-title">运行记录</h2>') +
    (runRows ? '<div class="card">' + runRows + '</div>' : tr('<div class="small faint">该 workflow 暂无运行记录（空态显式）。</div>')) +
    '</div>';
  el.querySelectorAll('.mini-row[data-go]').forEach((row) => {
    row.addEventListener('click', () => ctx.go(row.dataset.go));
  });
}

/** 运行独立页（#/wf/:name/r/:n——运行抽屉的独立路由形态）。 */
function runScreen(el, ctx, name, no) {
  const runs = ctx.store.runs().filter((r) => String(r && (r.workflow || r.name || '')) === String(name));
  const run = runs.find((r) => String(r.no != null ? r.no : '') === String(no)) || runs[0];
  const head = screenHead(tr('运行详情'), 'workflow=' + String(name) + ' · run #' + String(no),
    chipGo('← ' + String(name), 'wf/' + encodeURIComponent(String(name)), 'mute', false));
  if (!run) {
    el.innerHTML = '<div class="view-inner">' + head + emptyBox(tr('运行记录不在册'), tr('快照 runs 聚合无此运行（可能已出窗）。')) + '</div>';
    return;
  }
  const rows = [
    [tr('状态'), runChip(run)],
    ['route', esc(trunc(String(run.route || '—'), 120))],
  ];
  if (run.startedAt || run.at || run.ts) rows.push([tr('开始于'), timeHtml(run.startedAt || run.at || run.ts)]);
  if (run.finishedAt || run.doneAt) rows.push([tr('结束于'), timeHtml(run.finishedAt || run.doneAt)]);
  if (run.error) rows.push([tr('错误'), chip(trunc(String(run.error), 160), 'esc')]);
  const steps = Array.isArray(run.steps) ? run.steps : [];
  el.innerHTML = '<div class="view-inner">' + head +
    tr('<div class="card"><div class="card-h"><span class="h-title">概览</span></div><div class="card-b">') + kv(rows) + '</div></div>' +
    (steps.length
      ? tr('<h2 class="sec-title">步骤</h2><div class="card">') + steps.map((s) =>
        '<div class="mini-row"><div class="mr-main"><div>' + esc(String(s && (s.name || s.step) || tr('步骤'))) + ' ' + runChip(s) + '</div>' +
        '<div class="mr-sub small mono">' + esc(trunc(String(s && (s.detail || s.message) || ''), 120)) + '</div></div></div>'
      ).join('') + '</div>'
      : tr('<div class="small faint">契约未携带步骤明细——不显示不编造。</div>')) +
    '</div>';
}

export function render(el, ctx) {
  const p = ctx.params || {};
  if (p.n != null && p.n !== '') runScreen(el, ctx, p.name, p.n);
  else if (p.name != null && p.name !== '') detailScreen(el, ctx, p.name);
  else listScreen(el, ctx);
}
