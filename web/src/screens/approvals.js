/* approvals.js —— 审批 + 收件箱。待审区=草稿缓存（store.drafts——DocReview 类同源）
 * ∪ ApprovalNeeded 归档事件行（已批的不出现）。批准=两步（dry-run 预检 → 确认 Apply，
 * fieldManager=console:approve）；驳回=本地忽略（弃稿通路待后端词形，显式注记不装假）。
 * diff 注水=稿 content（GET /v1/drafts/{kind}/{name}）vs 面现值（GET /v1/objects）现读。
 * 已决记录=GET /v1/approvals 现读不缓存（缺席=诚实降级注记，零 4xx 假装成功）。
 * 收件箱=快照 inbox 投影（四类汇总+计数+深链）。 */

import { esc, trunc } from '../core/format.js';
import { chip, chipGo, kv, emptyBox, screenHead, kindBadge, docRoute, involvedRoute, objRoute } from '../render/chips.js';
import { diffLines, renderDiff, diffStats } from '../render/diff.js';

const ignored = new Set(); // 本地忽略（内存态——刷新复位；非服务端驳回）

/** 待审区两子源（drafts ∪ ApprovalNeeded 事件）。 */
function pendingItems(ctx) {
  const drafts = ctx.store.drafts()
    .filter((d) => d && d.kind && d.name)
    .map((d) => ({ kind: 'draft', key: 'draft|' + d.kind + '|' + (d.scope || 'global') + '|' + d.name, draft: d }));
  const evs = ctx.store.events()
    .filter((e) => e && e.type === 'ApprovalNeeded' && !(ignored.has('ev|' + String(e.id ?? ''))))
    .map((e) => ({ kind: 'event', key: 'ev|' + String(e.id ?? ''), ev: e }));
  return drafts.filter((x) => !ignored.has(x.key)).concat(evs);
}

function apKeyOf(d) {
  return (d.kind + '/' + (d.scope || 'global') + '/' + d.name).replace(/[^A-Za-z0-9]/g, '_');
}

/** 待审行（对象键 + 稿来源 + diff 容器 + 批准/忽略）。 */
function itemHtml(it) {
  if (it.kind === 'event') {
    const e = it.ev;
    const involved = String(e.involved || '');
    const route = involvedRoute(involved.split(',')[0] || '');
    return '<div class="card ap-item"><div class="card-h"><span class="h-title">' + esc(String(e.type)) + '</span>' +
      '<span class="h-right">' + chip(String(e.level || 'warn'), 'warn') + '</span></div>' +
      '<div class="card-b"><div class="small">' + esc(trunc(String(e.summary || e.involved || ''), 160)) + '</div>' +
      '<div style="margin-top:8px">' + (route ? chipGo(involved.split(',')[0], route, 'mute', true, '去对象工作台') : chip(involved || '—', 'mute')) +
      ' <button class="btn sm" data-ap-ignore="' + esc(it.key) + '">忽略</button></div></div></div>';
  }
  const d = it.draft;
  const scope = d.scope || 'global';
  return '<div class="card ap-item" data-ap-item="' + esc(apKeyOf(d)) + '">' +
    '<div class="card-h"><span class="h-title"><span class="lk mono" data-go="' + esc(docRoute(d.kind, scope, d.name)) + '">' +
    esc(scope + '/' + d.name) + '</span></span>' +
    '<span class="h-right">' + kindBadge(d.kind) + ' ' + chip('稿: ' + (d.fieldManager || '—'), 'mute') + '</span></div>' +
    '<div class="card-b"><div data-apdiff="' + esc(apKeyOf(d)) + '"><div class="small faint">diff 现读中…</div></div>' +
    '<div class="dh-actions" style="margin-top:10px">' +
    '<button class="btn sm primary" data-ap-approve="' + esc(apKeyOf(d)) + '">批准（= Apply）</button> ' +
    '<button class="btn sm" data-ap-ignore="' + esc(it.key) + '">忽略</button> ' +
    '<button class="btn sm" data-go="' + esc(docRoute(d.kind, scope, d.name)) + '">编辑器</button>' +
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
    box.innerHTML = '<div class="small faint">正文暂不可读（' + esc(String(draftR.error || draftR.status)) + '）——重进即恢复。</div>';
    return;
  }
  const dr = draftR.data && typeof draftR.data === 'object' ? draftR.data : {};
  const working = String(dr.content || '');
  let base = '（对象不在册——apply 将创建通路）';
  let baseNote = '基准=r?';
  if (faceR.ok) {
    const data = faceR.data && typeof faceR.data === 'object' ? faceR.data : {};
    base = data.content != null ? String(data.content) : '';
    baseNote = '基准=r' + (dr.baseResourceVersion != null ? dr.baseResourceVersion : (data.resourceVersion != null ? data.resourceVersion : '?'));
  } else if (faceR.status !== 404) {
    base = '（面现值暂不可读：' + String(faceR.error || faceR.status) + '）';
  }
  const stats = diffStats(diffLines(base, working));
  box.innerHTML = renderDiff(diffLines(base, working), 26) +
    '<div class="small faint" style="margin-top:4px">' + esc(baseNote) + ' · 稿 ' + working.length + 'B vs 落库值 ' + base.length + 'B（现读，不缓存） · ' +
    '<span class="chip ok tiny">+' + stats.add + '</span> <span class="chip esc tiny">-' + stats.del + '</span></div>';
}

/** 批准（两步：dry-run 预检 → 人确认 → fromDraft 落库，fieldManager=console:approve）。 */
async function approve(ctx, d, btn) {
  btn.disabled = true;
  btn.textContent = '预检中…';
  const dry = await ctx.api.apply({
    kind: d.kind, scope: d.scope || 'global', name: d.name, fromDraft: true, dryRun: true,
    fieldManager: 'console:approve', reasonRef: 'console:approve-dry',
  });
  btn.disabled = false;
  btn.textContent = '批准（= Apply）';
  if (!dry.ok) {
    if (typeof window !== 'undefined') window.alert('dry-run 预检未过（' + String(dry.error || dry.status) + '）——零副作用，稿已保留。');
    return;
  }
  const cur = dry.data && dry.data.rv != null ? Number(dry.data.rv) : null;
  const expect = d.baseResourceVersion != null ? Number(d.baseResourceVersion) : null;
  const lockNote = cur != null && expect != null ? (cur === expect ? '乐观锁对齐 r' + cur : '乐观锁漂移（现 r' + cur + ' vs 稿基 r' + expect + '——落库将被拒）') : '';
  if (!window.confirm('确认批准（Apply）' + d.kind + '/' + d.name + '？dry-run 已过（rv=' + String(cur) + '）' + (lockNote ? '；' + lockNote : '') + '。')) return;
  btn.disabled = true;
  btn.textContent = '落库中…';
  const r = await ctx.api.apply({
    kind: d.kind, scope: d.scope || 'global', name: d.name, fromDraft: true,
    fieldManager: 'console:approve', reasonRef: 'console:approve',
  });
  btn.disabled = false;
  btn.textContent = '批准（= Apply）';
  if (!r.ok) {
    window.alert('Apply 失败（' + String(r.error || r.status) + '）——稿保留语义，未落库。');
    return;
  }
  ignored.add('draft|' + d.kind + '|' + (d.scope || 'global') + '|' + d.name);
  const cardEl = document.querySelector('[data-ap-item="' + apKeyOf(d) + '"]');
  if (cardEl) cardEl.remove();
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
    slot.innerHTML = '<div class="small faint">审批记录暂不可读（' + esc(String(r.error || r.status)) + '）——判定留痕暂只=apply_log。</div>';
    return;
  }
  const rows = (r.data && Array.isArray(r.data.approvals)) ? r.data.approvals : [];
  if (!rows.length) {
    slot.innerHTML = '<div class="small faint">已决记录暂无判定行（批准/驳回即决即见）。</div>';
    return;
  }
  slot.innerHTML = '<h2 class="sec-title">已决记录</h2><div style="overflow-x:auto"><table class="tbl"><thead><tr>' +
    '<th>时刻</th><th>判定</th><th>对象</th><th>判定主体</th><th>rv</th></tr></thead><tbody>' +
    rows.map((x) => '<tr><td class="mono small">' + esc(String(x.ts || '').replace('T', ' ').slice(0, 19)) + '</td>' +
      '<td>' + chip(String(x.decision || ''), x.decision === 'granted' ? 'ok' : 'esc') + '</td>' +
      '<td class="mono small">' + (involvedRoute(x.obj) ? '<span class="lk" data-go="' + esc(involvedRoute(x.obj)) + '">' + esc(String(x.obj || '')) + '</span>' : esc(String(x.obj || ''))) + '</td>' +
      '<td class="mono small">' + esc(String(x.judgedBy || 'anonymous')) + '</td>' +
      '<td class="mono small">' + esc(x.rv ? 'r' + x.rv : '—') + '</td></tr>').join('') +
    '</tbody></table></div>';
}

function el__isConnected(x) { return !!(x && x.isConnected); }

/** 收件箱（快照 inbox 投影——四类汇总+计数+深链；缺席=诚实空态）。 */
function inboxHtml(ctx) {
  const inbox = ctx.store.inbox();
  if (!inbox) return '<div class="card"><div class="card-b"><div class="small faint">收件箱投影缺席（快照未携带 /v1/inbox 聚合）——待你处理面显式缺席，不装假。</div></div></div>';
  const items = Array.isArray(inbox.items) ? inbox.items : [];
  if (!items.length && !Number(inbox.total)) {
    return '<div class="card"><div class="card-b"><div class="small faint">收件箱为空（零待办——空态显式）。</div></div></div>';
  }
  return '<div class="card">' + items.slice(0, 50).map((x) => {
    const route = involvedRoute(x && (x.link || x.ref || x.obj));
    const main = String((x && (x.title || x.summary || x.kind)) || '');
    return '<div class="mini-row"><div class="mr-main"><div>' + esc(trunc(main, 80)) + ' ' + chip(String((x && x.class) || x && x.kind || ''), 'mute') + '</div>' +
      '<div class="mr-sub small">' + esc(trunc(String((x && (x.note || x.detail)) || ''), 90)) + '</div></div>' +
      (route ? '<span class="ws-chip">' + chipGo('去处理', route, 'acc', true) + '</span>' : '') + '</div>';
  }).join('') + '</div>';
}

/** 审批屏本体。 */
function approvalsScreen(el, ctx) {
  const items = pendingItems(ctx);
  const draftN = items.filter((x) => x.kind === 'draft').length;
  el.innerHTML = '<div class="view-inner">' +
    screenHead('审批', '待审区=草稿缓存（DocReview 类同源）∪ ApprovalNeeded 归档行（已批的不出现）。批准/驳回均=人点；diff=稿 vs 文档现值现读。',
      chip('批量批准通路待后端词形——逐条批准（诚实注记）', 'mute')) +
    (items.length
      ? items.map(itemHtml).join('')
      : emptyBox('待审区为空（草稿缓存与 ApprovalNeeded 归档均空）', '', chipGo('回 Welcome', 'welcome', 'primary', false))) +
    '<div data-ap-ledger style="margin-top:20px"><div class="small faint">已决记录现读中…</div></div>' +
    '<h2 class="sec-title">收件箱（待你处理）</h2>' + inboxHtml(ctx) +
    '</div>';

  items.filter((x) => x.kind === 'draft').forEach((x) => hydrateDiff(ctx, x.draft));
  hydrateLedger(ctx);

  el.querySelectorAll('[data-ap-approve]').forEach((btn) => {
    btn.addEventListener('click', () => {
      const d = ctx.store.drafts().find((x) => apKeyOf(x) === btn.dataset.apApprove);
      if (d) approve(ctx, d, btn);
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
