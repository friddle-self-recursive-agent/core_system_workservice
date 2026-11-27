/* approvals.js —— 审批 + 收件箱。待审区=草稿缓存（store.drafts——DocReview 类同源）
 * ∪ ApprovalNeeded 归档事件行（已批的不出现）。批准=两步（dry-run 预检 → 确认 Apply，
 * fieldManager=console:approve）；驳回=本地忽略（弃稿通路待后端词形，显式注记不装假）。
 * diff 注水=稿 content（GET /v1/drafts/{kind}/{name}）vs 面现值（GET /v1/objects）现读。
 * 已决记录=GET /v1/approvals 现读不缓存（缺席=诚实降级注记，零 4xx 假装成功）。
 * 收件箱=快照 inbox 投影（四类汇总+计数+深链）。 */

import { esc, trunc, relAge } from '../core/format.js';
import { chip, chipGo, kv, emptyBox, screenHead, kindBadge, docRoute, involvedRoute, objRoute } from '../render/chips.js';
import { diffLines, renderDiff, diffStats } from '../render/diff.js';
import { t as tr } from '../core/i18n.js';

const ignored = new Set(); // 本地忽略（内存态——刷新复位；非服务端驳回）

/** 待审区三源（F-09 ①②——2026-11-26 审批中心口径）：挂起 Workflow run（suspended ∨
 * AwaitingApproval——总览有审批页也要有）∪ 待批 Proposal（status=submitted，批闸锚=
 * Proposal/<scope>/<name>——批准/驳回=Apply 判定串通路）∪ 草稿缓存（DocReview 同源）∪
 * ApprovalNeeded 归档事件行。四源合并同构卡片，均可操作。 */
function pendingItems(ctx) {
  const drafts = ctx.store.drafts()
    .filter((d) => d && d.kind && d.name)
    .map((d) => ({ kind: 'draft', key: 'draft|' + d.kind + '|' + (d.scope || 'global') + '|' + d.name, draft: d }));
  const runs = ctx.store.runs()
    .filter((r) => r && (String(r.status || r.state || '') === 'suspended' || String(r.phase || '') === 'AwaitingApproval'))
    .map((r) => ({ kind: 'run', key: 'run|' + String(r.workflow || '') + '#' + String(r.run != null ? r.run : ''), run: r }));
  const props = ctx.store.proposals()
    .filter((p) => p && p.name && String(p.status || 'submitted') === 'submitted')
    .map((p) => ({ kind: 'proposal', key: 'proposal|' + String(p.scope || 'global') + '|' + String(p.name), proposal: p }));
  const evs = ctx.store.events()
    .filter((e) => e && e.type === 'ApprovalNeeded' && !(ignored.has('ev|' + String(e.id ?? ''))))
    .map((e) => ({ kind: 'event', key: 'ev|' + String(e.id ?? ''), ev: e }));
  return drafts.filter((x) => !ignored.has(x.key)).concat(runs, props, evs);
}

function apKeyOf(d) {
  return (d.kind + '/' + (d.scope || 'global') + '/' + d.name).replace(/[^A-Za-z0-9]/g, '_');
}

/* ── 审批信息面（2026-11-27 三次裁决「批准啥？approve 至少要展示更新了哪些文档」）──
 * 待批行/卡不得只有裸按钮：①本 run 窗口内已更新文档（DocApplied 事件归集）②待 apply
 * 文档稿（草稿缓存——批准将落哪些文档）③闸前判断摘要（ApprovalNeeded 事件 summary=
 * agent 判定词形——LLM 整理的信息随事件词形携带；缺席=诚实注记不编造）。 */

export function runApprovalInfoOf(store, wf, scope, run, startedAt) {
  const anchor = 'Workflow/' + String(scope || 'global') + '/' + String(wf || '');
  const ev = (store.events() || []).find((e) => e && e.type === 'ApprovalNeeded' &&
    String(e.involved || '').indexOf(anchor) >= 0);   // 锚唯一=最新闸前判断（LLM 判定词形随事件 summary 携带）
  const docs = (store.events() || [])
    .filter((e) => e && e.type === 'DocApplied')
    .filter((e) => !startedAt || String(e.ts || e.at || '') >= String(startedAt))
    .slice(0, 5)
    .map((e) => String(e.involved || '').split(',')[0].trim())
    .filter(Boolean);
  const drs = (store.drafts() || [])
    .filter((d) => d && d.kind && d.name)
    .slice(0, 5)
    .map((d) => d.kind + '/' + (d.scope || 'global') + '/' + d.name);
  return { judge: ev ? String(ev.summary || '') : '', docs, drafts: drs };
}

export function runApprovalInfoHtml(info) {
  const parts = [];
  if (info.docs.length) {
    parts.push('<span class="chip ok tiny">已更新 ' + info.docs.length + ' 篇</span> <span class="mono small">' +
      esc(info.docs.map((d) => trunc(d, 30)).join('、')) + '</span>');
  }
  if (info.drafts.length) {
    parts.push('<span class="chip warn tiny">待批稿 ' + info.drafts.length + ' 份</span> <span class="mono small">' +
      esc(info.drafts.map((d) => trunc(d, 30)).join('、')) + '</span>');
  }
  if (info.judge) {
    parts.push('<span class="small">' + esc(tr('判断：')) + esc(trunc(info.judge, 110)) + '</span>');
  }
  if (!parts.length) return '<span class="small faint">' + esc(tr('窗口内无文档动态（DocApplied/待批稿缺席）——批准前详情请看「运行详情」。')) + '</span>';
  return parts.join(' <span class="faint">·</span> ');
}

/** 事件行→Workflow 锚（§19 R3 批闸语义：run 恢复唯一人工通路=对 Workflow 锚发
 * ApprovalGranted/ApprovalRejected——involved 首段非 Workflow 词形=无锚，按钮置灰诚实注记）。 */
function wfAnchorOf(ev) {
  const first = String((ev && ev.involved) || '').split(',')[0].trim();
  return /^Workflow\/[^/\s]+\/[^/\s]+$/.test(first) ? first : '';
}

/** V-13/V-14/F-09 定位匹配：draft 按对象键/name/apKey；run 按 workflow/run 号；proposal 按
 * name/scope；ApprovalNeeded 事件按 involved/summary/dedupKey。 */
export function matchesApprovalFocus(it, focus) {
  const f = String(focus || '').trim();
  if (!f || !it) return false;
  if (it.kind === 'draft' && it.draft) {
    const d = it.draft;
    const key = d.kind + '/' + (d.scope || 'global') + '/' + d.name;
    return apKeyOf(d) === f || key === f || String(d.name) === f;
  }
  if (it.kind === 'run' && it.run) {
    const r = it.run;
    return String(r.workflow || '') === f || ('Workflow/' + String(r.owner || 'global') + '/' + String(r.workflow || '')) === f ||
      String(r.run) === f;
  }
  if (it.kind === 'proposal' && it.proposal) {
    const p = it.proposal;
    return String(p.name) === f || ('Proposal/' + String(p.scope || 'global') + '/' + String(p.name)) === f;
  }
  if (it.kind === 'event' && it.ev) {
    const e = it.ev;
    return String(e.involved || '').includes(f) || String(e.summary || '').includes(f) || String(e.dedupKey || '').includes(f);
  }
  return false;
}

/** 待审行（对象键 + 稿来源 + diff 容器 + 批准/忽略；run/proposal/事件行同构=摘要+锚 chips+按钮组）。 */
function itemHtml(ctx, it, idx, focused = false) {
  const approveCls = idx === 0 ? 'btn btn-primary sm' : 'btn btn-secondary sm';
  const hl = focused ? ' data-ap-focused="1" style="outline:2px solid var(--accent)"' : '';
  if (it.kind === 'run') {
    // F-09 ①：挂起 Workflow run 行（总览审批区同源——批准/驳回=对 Workflow 锚发布
    // ApprovalGranted/ApprovalRejected；锚词形与 §19 R3 批闸语义同源）。
    // 审批信息面（三次裁决）：批准前展示已更新文档/待批稿/闸前判断摘要——不再裸按钮。
    const r = it.run;
    const anchor = 'Workflow/' + String(r.owner || 'global') + '/' + String(r.workflow || '');
    const no = String(r.run != null ? r.run : '');
    const info = runApprovalInfoHtml(runApprovalInfoOf(ctx.store, String(r.workflow || ''), String(r.owner || 'global'), no, String(r.startedAt || '')));
    return '<div class="card ap-item"' + hl + '><div class="card-h"><span class="h-title mono">' +
      esc(String(r.workflow || '—')) + (no ? ' #' + esc(no) : '') + '</span>' +
      '<span class="h-right">' + chip(String(r.phase || 'AwaitingApproval'), 'warn') + '</span></div>' +
      '<div class="card-b"><div class="small">' + esc(trunc(String(r.involved || r.route || tr('工作流闸锚挂起——等待审批信号')), 160)) + '</div>' +
      '<div class="small" style="margin-top:8px;border-top:1px dashed var(--border);padding-top:8px">' + info + '</div>' +
      '<div style="margin-top:8px;display:flex;gap:6px;flex-wrap:wrap">' +
      chipGo(tr('锚 ') + anchor, 'wf/' + encodeURIComponent(String(r.workflow || '')) + (no ? '/r/' + encodeURIComponent(no) : ''), 'acc', true, tr('运行详情——批闸锚')) +
      '</div>' +
      '<div class="dh-actions" style="margin-top:10px">' +
      '<button class="' + approveCls + '" data-ap-grant="' + esc(it.key) + '" data-ap-anchor="' + esc(anchor) + '" data-ap-run="' + esc(no) + '">' + esc(tr('批准继续')) + '</button> ' +
      '<button class="btn btn-ghost btn-danger sm" data-ap-reject="' + esc(it.key) + '" data-ap-anchor="' + esc(anchor) + '" data-ap-run="' + esc(no) + '">' + esc(tr('驳回终止')) + '</button>' +
      '</div></div></div>';
  }
  if (it.kind === 'proposal') {
    // F-09 ②：待批 Proposal 行（批闸锚=Proposal/<scope>/<name>——批准/驳回=Apply
    // console:approve|console:reject 判定串通路（apiserver 落 approval_log+随行发布
    // ApprovalGranted/Rejected——ApprovalNeeded 按钮置灰死路收口）。
    const p = it.proposal;
    const anchor = 'Proposal/' + String(p.scope || 'global') + '/' + String(p.name);
    return '<div class="card ap-item"' + hl + '><div class="card-h"><span class="h-title">' +
      esc(String(p.title || p.name || tr('未命名提案'))) + '</span>' +
      '<span class="h-right">' + chip(String(p.status || 'submitted'), 'warn') + '</span></div>' +
      '<div class="card-b"><div class="small">' + esc(trunc(String(p.summary || ''), 200) || tr('（summary 缺席——详情见对象工作台）')) + '</div>' +
      '<div style="margin-top:8px;display:flex;gap:6px;flex-wrap:wrap">' +
      chipGo(tr('锚 ') + anchor, 'obj/Proposal/' + encodeURIComponent(String(p.scope || 'global')) + '/' + encodeURIComponent(String(p.name)), 'mute', true, tr('对象工作台')) +
      '</div>' +
      '<div class="dh-actions" style="margin-top:10px">' +
      '<button class="' + approveCls + '" data-ap-prop="' + esc(it.key) + '" data-ap-prop-anchor="' + esc(anchor) + '" data-ap-prop-decision="approve">' + esc(tr('批准（= Apply）')) + '</button> ' +
      '<button class="btn btn-ghost btn-danger sm" data-ap-prop="' + esc(it.key) + '" data-ap-prop-anchor="' + esc(anchor) + '" data-ap-prop-decision="reject">' + esc(tr('驳回')) + '</button>' +
      '</div></div></div>';
  }
  if (it.kind === 'event') {
    const e = it.ev;
    const involved = String(e.involved || '');
    const anchor = wfAnchorOf(e);
    const wfName = anchor ? String(anchor.split('/')[2] || '') : '';
    const route = involvedRoute(involved.split(',')[0] || '');
    const gateBtns = anchor
      ? '<button class="' + approveCls + '" data-ap-grant="' + esc(it.key) + '" data-ap-ev="' + esc(String(e.id == null ? '' : e.id)) + '" data-ap-anchor="' + esc(anchor) + '">' + esc(tr('批准继续')) + '</button> ' +
        '<button class="btn btn-ghost btn-danger sm" data-ap-reject="' + esc(it.key) + '" data-ap-ev="' + esc(String(e.id == null ? '' : e.id)) + '" data-ap-anchor="' + esc(anchor) + '">' + esc(tr('驳回终止')) + '</button> '
      : '<button class="btn btn-secondary sm" disabled title="' + esc(tr('事件未携带 Workflow 锚')) + '">' + esc(tr('批准继续')) + '</button> ' +
        '<button class="btn btn-ghost btn-danger sm" disabled title="' + esc(tr('事件未携带 Workflow 锚')) + '">' + esc(tr('驳回终止')) + '</button> ';
    return '<div class="card ap-item"' + hl + '><div class="card-h"><span class="h-title">' + esc(String(e.type)) + '</span>' +
      '<span class="h-right">' + chip(String(e.level || 'warn'), 'warn') + '</span></div>' +
      '<div class="card-b"><div class="small">' + esc(trunc(String(e.summary || e.involved || ''), 160)) + '</div>' +
      '<div style="margin-top:8px;display:flex;gap:6px;flex-wrap:wrap">' +
      (route ? chipGo(involved.split(',')[0], route, 'mute', true, tr('去对象工作台')) : chip(involved || '—', 'mute')) +
      (wfName ? chipGo(tr('锚 ') + anchor, 'wf/' + encodeURIComponent(wfName), 'acc', true, tr('Workflow 详情——批闸锚')) : '') +
      '</div>' +
      '<div class="dh-actions" style="margin-top:10px">' + gateBtns +
      '<button class="btn btn-ghost sm" data-ap-ignore="' + esc(it.key) + '">' + esc(tr('忽略')) + '</button></div></div></div>';
  }
  const d = it.draft;
  const scope = d.scope || 'global';
  return '<div class="card ap-item" data-ap-item="' + esc(apKeyOf(d)) + '"' + hl + '>' +
    '<div class="card-h"><span class="h-title"><span class="lk mono" data-go="' + esc(docRoute(d.kind, scope, d.name)) + '">' +
    esc(scope + '/' + d.name) + '</span></span>' +
    '<span class="h-right">' + kindBadge(d.kind) + ' ' + chip(tr('稿: ') + (d.fieldManager || '—'), 'mute') + '</span></div>' +
    '<div class="card-b"><div data-apdiff="' + esc(apKeyOf(d)) + tr('"><div class="small faint">diff 现读中…</div></div>') +
    '<div class="dh-actions" style="margin-top:10px">' +
    '<button class="' + approveCls + '" data-ap-approve="' + esc(apKeyOf(d)) + tr('">批准（= Apply）</button> ') +
    '<button class="btn btn-ghost sm" data-ap-ignore="' + esc(it.key) + tr('">忽略</button> ') +
    '<button class="btn btn-secondary sm" data-go="' + esc(docRoute(d.kind, scope, d.name)) + tr('">编辑器</button>') +
    '</div></div></div>';
}

/** diff 注水：稿 content + 面现值现读（不缓存——判定窗口最小化）。 */
async function hydrateDiff(ctx, d) {
  const stamp = typeof location !== 'undefined' ? location.hash : '';
  const [draftR, faceR] = await Promise.all([
    ctx.api.draft(d.kind, d.name, d.scope || 'global'),
    ctx.api.object(d.kind, d.scope || 'global', d.name),
  ]);
  if ((typeof location !== 'undefined' ? location.hash : '') !== stamp) return;
  const box = document.querySelector('[data-apdiff="' + apKeyOf(d) + '"]');
  if (!box) return;
  if (!draftR.ok) {
    box.innerHTML = tr('<div class="small faint">正文暂不可读（') + esc(String(draftR.error || draftR.status)) + tr('）——重进即恢复。</div>');
    hydrateVerdict(ctx, d); // 判决文档独立注水（稿消费后 diff 缺席——判决恰是此时最有用）
    return;
  }
  const dr = draftR.data && typeof draftR.data === 'object' ? draftR.data : {};
  const working = String(dr.content || '');
  let base = tr('（对象尚未落库——批准即创建）');
  let baseNote = tr('基准=r?');
  if (faceR.ok) {
    const data = faceR.data && typeof faceR.data === 'object' ? faceR.data : {};
    base = data.content != null ? String(data.content) : '';
    baseNote = tr('基准=r') + (dr.baseResourceVersion != null ? dr.baseResourceVersion : (data.resourceVersion != null ? data.resourceVersion : '?'));
  } else if (faceR.status !== 404) {
    base = tr('（面现值暂不可读：') + String(faceR.error || faceR.status) + '）';
  }
  const stats = diffStats(diffLines(base, working));
  box.innerHTML = renderDiff(diffLines(base, working), 26) +
    '<div class="small faint" style="margin-top:4px">' + esc(baseNote) + tr(' · 稿 ') + working.length + tr('B vs 落库值 ') + base.length + tr('B（现读，不缓存） · ') +
    '<span class="chip ok tiny">+' + stats.add + '</span> <span class="chip esc tiny">-' + stats.del + '</span></div>' +
    '<div data-apverdict="' + esc(apKeyOf(d)) + '"><div class="small faint">' + esc(tr('判决文档（ops_core_inner_agent）现读中…')) + '</div></div>';
  hydrateVerdict(ctx, d);
}

/** 判决文档注水（2026-11-28 用户裁定 ops_core_inner_agent：审批裁定走小内置 agent 整理
 * DocumentDiff 生成 Diff 文档——引擎人审挂起即生成，同名幂等 docdiff-<kind>-<name>；
 * 在册=渲染判决正文；缺席=诚实词形（生成中/未生成，不编造））。 */
async function hydrateVerdict(ctx, d) {
  try {
    const ddName = 'docdiff-' + String(d.kind + '-' + d.name).toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '');
    const ddR = await ctx.api.object('DocumentDiff', 'global', ddName);
    const vbox = document.querySelector('[data-apverdict="' + CSS.escape(apKeyOf(d)) + '"]');
    if (!vbox) return;
    if (ddR.ok && ddR.data && ddR.data.content) {
      const raw = String(ddR.data.content);
      const vm = raw.match(/verdict:\s*\|\n([\s\S]*?)(?:\n  [a-zA-Z]|\n\w|$)/);
      const cm = raw.match(/changed:\s*"?([^"\n]+)"?/);
      const verdict = vm ? vm[1].trim() : raw;
      vbox.innerHTML = '<div class="card" style="border:1px solid var(--ds-accent);margin-top:8px"><div class="card-h"><span class="h-title">' +
        esc(tr('Diff 判决 · ops_core_inner_agent')) + '</span><span class="h-right">' + esc(cm ? cm[1] : '') + '</span></div>' +
        '<div class="card-b"><div class="small" style="white-space:pre-wrap">' + esc(verdict) + '</div>' +
        '<div class="small faint" style="margin-top:6px">' + esc(tr('对象 DocumentDiff/global/')) + esc(ddName) + '</div></div></div>';
    } else {
      vbox.innerHTML = '<div class="small faint">' + esc(tr('判决文档尚未生成（DocumentDiff/global/') + esc(ddName) + tr(' 缺席——引擎人审挂起即自动生成；缺席=未到生成窗口，不编造')) + '</div></div>';
    }
  } catch (e) { /* 判决注水失败不挡 diff 主面 */ }
}

/** 批准（两步：dry-run 预检 → 人确认 → fromDraft 落库，fieldManager=console:approve）。 */
async function approve(ctx, d, btn) {
  btn.disabled = true;
  btn.textContent = tr('预检中…');
  const dry = await ctx.api.apply({
    kind: d.kind, scope: d.scope || 'global', name: d.name, fromDraft: true, dryRun: true,
    fieldManager: 'console:approve', reasonRef: 'console:approve-dry',
  });
  btn.disabled = false;
  btn.textContent = tr('批准（= Apply）');
  if (!dry.ok) {
    if (typeof window !== 'undefined') window.alert(tr('dry-run 预检未过（') + String(dry.error || dry.status) + tr('）——零副作用，稿已保留。'));
    return;
  }
  const cur = dry.data && dry.data.rv != null ? Number(dry.data.rv) : null;
  const expect = d.baseResourceVersion != null ? Number(d.baseResourceVersion) : null;
  const lockNote = cur != null && expect != null ? (cur === expect ? tr('版本对齐 r') + cur : tr('版本已变化（现 r') + cur + tr(' vs 编辑基线 r') + expect + tr('——落库前请先刷新）')) : '';
  if (!window.confirm(tr('确认批准（Apply）') + d.kind + '/' + d.name + tr('？dry-run 已过（rv=') + String(cur) + '）' + (lockNote ? '；' + lockNote : '') + '。')) return;
  btn.disabled = true;
  btn.textContent = tr('落库中…');
  const r = await ctx.api.apply({
    kind: d.kind, scope: d.scope || 'global', name: d.name, fromDraft: true,
    fieldManager: 'console:approve', reasonRef: 'console:approve',
  });
  btn.disabled = false;
  btn.textContent = tr('批准（= Apply）');
  if (!r.ok) {
    window.alert(tr('Apply 失败（') + String(r.error || r.status) + tr('）——稿保留语义，未落库。'));
    return;
  }
  ignored.add('draft|' + d.kind + '|' + (d.scope || 'global') + '|' + d.name);
  const cardEl = document.querySelector('[data-ap-item="' + apKeyOf(d) + '"]');
  if (cardEl) cardEl.remove();
  ctx.api.snapshot().then((s) => { if (s.ok) ctx.store.init(s.data); });
}

/** 批闸事件发布（§19 R3——run 在 AwaitingApproval 闸锚挂起后恢复的唯一人工通路：
 * 发布 ApprovalGranted（run 沿声明相位序列续行）/ ApprovalRejected（run 终态 Failed）。
 * involved=事件携带的 Workflow 锚；dedupKey=console-<eventId>；失败=显式 alert 不装成功）。 */
async function pubGateDecision(ctx, btn, grant) {
  const anchor = btn.getAttribute('data-ap-anchor') || '';
  const key = btn.getAttribute(grant ? 'data-ap-grant' : 'data-ap-reject') || '';
  const evId = btn.getAttribute('data-ap-ev') || '';
  if (!anchor || !key) return;
  const confirmed = grant
    ? window.confirm(tr('确认批准继续？') + anchor + tr('：发布 ApprovalGranted 后 run 沿声明相位序列续行。'))
    : window.confirm(tr('确认驳回终止？') + anchor + tr('：发布 ApprovalRejected 后 run 将进入终态 Failed（不可恢复）。'));
  if (!confirmed) return;
  btn.disabled = true;
  btn.textContent = tr('发布中…');
  const runNo = Number(btn.getAttribute('data-ap-run') || 0) || 0;
  const payload = { type: grant ? 'ApprovalGranted' : 'ApprovalRejected', involved: anchor, ...(runNo > 0 ? { run: runNo } : {}) };
  if (evId) payload.dedupKey = 'console-' + evId;
  else payload.dedupKey = 'console-appr-' + key + '-' + Date.now(); // run/proposal 行无事件锚——会话内唯一 dedup 键（不冒充幂等）
  const r = await ctx.api.eventPub(payload);
  if (!r.ok) {
    btn.disabled = false;
    btn.textContent = grant ? tr('批准继续') : tr('驳回终止');
    window.alert(tr('事件发布失败（') + String(r.error || r.status) + tr('）——run 仍挂原相位，未装成功。'));
    return;
  }
  ignored.add(key);
  const card = btn.closest('.ap-item');
  if (card) card.remove();
  ctx.api.snapshot().then((s) => { if (s.ok) ctx.store.init(s.data); });
}

/** Proposal 判定动作（F-09 ②批闸闭环——批准/驳回=Apply fromDraft 判定通路：
 * fieldManager=console:approve|console:reject（apiserver 落 approval_log + 随行发布
 * ApprovalGranted/Rejected——与草稿批准同一条 Apply 链零旁路；失败=alert 显式不装成功）。 */
async function proposalDecision(ctx, p, decision, btn) {
  const isApprove = decision === 'approve';
  const fm = isApprove ? 'console:approve' : 'console:reject';
  if (!window.confirm((isApprove ? tr('确认批准提案（Apply console:approve）') : tr('确认驳回提案（Apply console:reject）')) +
    ' ' + String(p.name) + '？')) return;
  btn.disabled = true;
  btn.textContent = tr('判定中…');
  const r = await ctx.api.apply({
    kind: 'Proposal', scope: p.scope || 'global', name: p.name, fromDraft: true,
    fieldManager: fm, reasonRef: 'console:' + decision + ':' + String(p.name || ''),
  });
  btn.disabled = false;
  btn.textContent = isApprove ? tr('批准（= Apply）') : tr('驳回');
  if (!r || !r.ok) {
    window.alert(tr('提案判定失败（') + String((r && (r.error || r.status)) || tr('网络')) + tr('）——提案状态未翻转，可重试。'));
    return;
  }
  ignored.add('proposal|' + (p.scope || 'global') + '|' + String(p.name));
  const card = btn.closest('.ap-item');
  if (card) card.remove();
  ctx.api.snapshot().then((s) => { if (s.ok) ctx.store.init(s.data); });
}

/** 已决记录（现读不缓存；缺席=诚实降级）。 */
async function hydrateLedger(ctx) {
  const stamp = typeof location !== 'undefined' ? location.hash : '';
  const slot = document.querySelector('[data-ap-ledger]');
  if (!slot) return;
  const r = await ctx.api.approvals(20);
  if (!el__isConnected(slot) || (typeof location !== 'undefined' ? location.hash : '') !== stamp) return;
  if (!r.ok) {
    slot.innerHTML = tr('<div class="small faint">审批记录暂不可读（') + esc(String(r.error || r.status)) + tr('）。</div>');
    return;
  }
  const rows = (r.data && Array.isArray(r.data.approvals)) ? r.data.approvals : [];
  if (!rows.length) {
    slot.innerHTML = tr('<div class="small faint">暂无已决记录（批准/驳回后此处显示）。</div>');
    return;
  }
  slot.innerHTML = tr('<h2 class="sec-title">已决记录</h2><div style="overflow-x:auto"><table class="tbl"><thead><tr>') +
    tr('<th>时刻</th><th>判定</th><th>对象</th><th>判定主体</th><th>rv</th></tr></thead><tbody>') +
    rows.map((x) => '<tr><td class="mono small">' + esc(String(x.ts || '').replace('T', ' ').slice(0, 19)) + '</td>' +
      '<td>' + chip(String(x.decision || ''), x.decision === 'granted' ? 'ok' : 'esc') + '</td>' +
      '<td class="mono small">' + (involvedRoute(x.obj) ? '<span class="lk" data-go="' + esc(involvedRoute(x.obj)) + '">' + esc(String(x.obj || '')) + '</span>' : esc(String(x.obj || ''))) + '</td>' +
      '<td class="mono small">' + esc(String(x.judgedBy || 'anonymous')) + '</td>' +
      '<td class="mono small">' + esc(x.rv ? 'r' + x.rv : '—') + '</td></tr>').join('') +
    '</tbody></table></div>';
}

function el__isConnected(x) { return !!(x && x.isConnected); }

/** 收件箱（快照 inbox 投影——四类汇总+计数+深链；缺席=诚实空态）。
 * F-09 ④：行读取=投影真实词形 x.digest/at/link（此前读 x.title/summary/note——字段不存在
 * 全显无字空白行；digest=服务端拼装摘要、at=时刻、link=#/ 深链路由）。 */
function inboxHtml(ctx) {
  const inbox = ctx.store.inbox();
  if (!inbox) return tr('<div class="card"><div class="card-b"><div class="small faint">收件箱暂不可用（投影缺席）。</div></div></div>');
  const items = Array.isArray(inbox.items) ? inbox.items : [];
  if (!items.length && !Number(inbox.total)) {
    return tr('<div class="card"><div class="card-b"><div class="small faint">收件箱为空（零待办）。</div></div></div>');
  }
  return '<div class="card">' + items.slice(0, 50).map((x) => {
    const linkRaw = String((x && (x.link || x.ref || x.obj)) || '');
    const link = linkRaw.replace(/^#\//, '');
    const route = link ? (involvedRoute(link) || link) : '';
    const main = String((x && (x.digest || x.title || x.summary || x.kind)) || '—');
    const cls = String((x && x.class) || (x && x.kind) || '');
    const at = relAge(String((x && x.at) || '')) || '';
    return '<div class="mini-row"><div class="mr-main"><div>' + esc(trunc(main, 80)) + ' ' + chip(cls, 'mute') + '</div>' +
      '<div class="mr-sub small faint">' + esc(at) + '</div></div>' +
      (route ? '<span class="ws-chip">' + chipGo(tr('去处理'), route, 'acc', true) + '</span>' : '') + '</div>';
  }).join('') + '</div>';
}

/** 审批屏本体。支持 approvals/:focus（V-13 草稿 diff 定位 / V-14 提案定位）。 */
function approvalsScreen(el, ctx) {
  const focus = String((ctx.params && ctx.params.focus) || '').trim();
  const rawItems = pendingItems(ctx);
  const matched = focus ? rawItems.filter((it) => matchesApprovalFocus(it, focus)) : [];
  const rest = focus ? rawItems.filter((it) => !matchesApprovalFocus(it, focus)) : rawItems;
  const items = matched.concat(rest);
  const focusBanner = focus
    ? '<div class="card" style="margin-bottom:10px"><div class="card-b small" style="display:flex;gap:8px;align-items:center;flex-wrap:wrap">' +
      chip(tr('定位：') + focus, matched.length ? 'acc' : 'warn') +
      '<span>' + esc(matched.length
        ? tr('已置顶高亮对应待审项（') + matched.length + tr(' 项）。')
        : tr('待审区暂无匹配项——可能已决或尚未落稿，请核查下方已决记录。')) + '</span>' +
      chipGo(tr('清除定位'), 'approvals', 'mute', true) + '</div></div>'
    : '';
  el.innerHTML = '<div class="view-inner">' +
    screenHead(tr('审批'), tr('待审区=草稿缓存（DocReview 类同源）∪ ApprovalNeeded 归档行。事件行批准/驳回=对 Workflow 锚发布 ApprovalGranted/ApprovalRejected（run 续行/终态 Failed）；草稿批准=Apply；均=人点确认。'),
      chip(tr('批量批准通路待后端词形——逐条批准（诚实注记）'), 'mute')) +
    focusBanner +
    (items.length
      ? items.map((it, idx) => itemHtml(ctx, it, idx, !!focus && matchesApprovalFocus(it, focus))).join('')
      : emptyBox(tr('暂无待审事项'), '', chipGo(tr('回 Welcome'), 'welcome', 'acc', false))) +
    tr('<div data-ap-ledger style="margin-top:20px"><div class="small faint">已决记录现读中…</div></div>') +
    tr('<h2 class="sec-title">收件箱（待你处理）</h2>') + inboxHtml(ctx) +
    '</div>';

  items.filter((x) => x.kind === 'draft').forEach((x) => hydrateDiff(ctx, x.draft));
  hydrateLedger(ctx);
  if (focus && matched.length) {
    const hit = el.querySelector('[data-ap-focused="1"]');
    if (hit && hit.scrollIntoView) hit.scrollIntoView({ behavior: 'smooth', block: 'center' });
  }

  el.querySelectorAll('[data-ap-approve]').forEach((btn) => {
    btn.addEventListener('click', () => {
      const d = ctx.store.drafts().find((x) => apKeyOf(x) === btn.dataset.apApprove);
      if (d) approve(ctx, d, btn);
    });
  });
  el.querySelectorAll('[data-ap-grant]').forEach((btn) => {
    btn.addEventListener('click', () => pubGateDecision(ctx, btn, true));
  });
  el.querySelectorAll('[data-ap-reject]').forEach((btn) => {
    btn.addEventListener('click', () => pubGateDecision(ctx, btn, false));
  });
  // F-09 ②：Proposal 判定按钮（data-ap-prop + decision 词形——Apply console:approve|reject 通路）。
  const propById = new Map((ctx.store.proposals() || [])
    .filter((p) => p && p.name)
    .map((p) => ['proposal|' + String(p.scope || 'global') + '|' + String(p.name), p]));
  el.querySelectorAll('[data-ap-prop]').forEach((btn) => {
    btn.addEventListener('click', () => {
      const p = propById.get(btn.getAttribute('data-ap-prop') || '');
      if (p) proposalDecision(ctx, p, String(btn.getAttribute('data-ap-prop-decision') || 'approve'), btn);
    });
  });
  el.querySelectorAll('[data-ap-ignore]').forEach((btn) => {
    btn.addEventListener('click', () => {
      ignored.add(btn.dataset.apIgnore);
      const card = btn.closest('.ap-item');
      if (card) card.remove();
    });
  });
  el.querySelectorAll('[data-go]').forEach((x) => {
    x.addEventListener('click', () => ctx.go(x.dataset.go));
  });
}

export function render(el, ctx) {
  approvalsScreen(el, ctx);
}
