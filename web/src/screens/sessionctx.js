/* sessionctx.js —— 会话上下文面板（G-14：ContextPack/TagCtxShard/takeBudget 查看 + 编辑分叉）。
 * 数据流：GET /v1/sessions/{id}/context（apiserver routes_sessions.go 契约面）→ normalizeContext()
 * 归一 → 渲染；编辑分叉=PUT 全量覆盖（乐观锁 baseRevision，409=显式冲突词形不静默重基）+
 * POST …/revert 撤销（服务端写前快照恢复）。写面随 --allow-apply 门控——405=诚实关断词形
 * （输入禁用+注记，不冒充可编辑）。诚实纪律：本代无三段式装配器，pack.degraded=true——
 * chunks=索引 TopK 候选投影（面板明示词形），不冒充会话真实装配包。
 * 纯函数（parseTagInput/capTags/normalizeContext/buildSavePayload/renderSessionCtx）零 DOM
 * 依赖——node 单测机械直测口径；样式只用 styles/ 三件既有体系类，零新 CSS。 */

import { esc, trunc, relAge, fmtNum } from '../core/format.js';
import { chip, chipGo, kv, objRoute, docRoute, tagRoute } from '../render/chips.js';
import { t as tr } from '../core/i18n.js';

/* 面板内存态：sessionId → {model}（store 触发的整屏重绘间复用——避免闪空）。 */
const PANEL = new Map();

/* ---------- 纯逻辑（node 单测直测） ---------- */

/** tag 输入解析：逗号/空白（含中文逗号顿号）分隔 → 去空词形数组。 */
export function parseTagInput(raw) {
  return String(raw == null ? '' : raw)
    .split(/[\s,，、]+/)
    .map((t) => t.trim())
    .filter(Boolean);
}

/** ContextPolicy.packLimits 装配面收敛（保序去重；总 ≤maxTotal、每 tagType ≤maxPerType；
 * 裸 tag 归 type=''——与后端 splitTagKey 同词形）。返回 {kept, dropped}。 */
export function capTags(tags, maxTotal = 20, maxPerType = 5) {
  const kept = [];
  const dropped = [];
  const seen = new Set();
  const perType = new Map();
  for (const raw of (Array.isArray(tags) ? tags : [])) {
    const key = String(raw || '').trim();
    if (!key || seen.has(key)) continue;
    seen.add(key);
    const c = key.indexOf(':');
    const typ = c > 0 ? key.slice(0, c) : '';
    const n = perType.get(typ) || 0;
    if (kept.length >= maxTotal || n >= maxPerType) { dropped.push(key); continue; }
    perType.set(typ, n + 1);
    kept.push(key);
  }
  return { kept, dropped };
}

/** GET 载荷 → 视图模型（字段逐项归一缺省——缺席=null/空集/0，不冒充）。
 * shardsFacePresent=false=分片缓存面缺席（后端 shards:null）——渲染区显式降级词形。 */
export function normalizeContext(data) {
  const d = data && typeof data === 'object' ? data : {};
  const pack = d.pack && typeof d.pack === 'object' ? d.pack : {};
  const pol = d.policy && typeof d.policy === 'object' ? d.policy : {};
  const sess = d.session && typeof d.session === 'object' ? d.session : null;
  const ov = d.override && typeof d.override === 'object' ? d.override : null;
  const chunks = (Array.isArray(pack.chunks) ? pack.chunks : []).map((c) => {
    const row = c && typeof c === 'object' ? c : {};
    return {
      kind: String(row.kind || ''), scope: String(row.scope || ''), name: String(row.name || ''),
      score: Number(row.score) || 0, why: String(row.why || ''),
    };
  });
  const shards = (Array.isArray(d.shards) ? d.shards : []).map((s) => {
    const row = s && typeof s === 'object' ? s : {};
    return {
      tagKey: String(row.tagKey || ''), docs: Number(row.docs) || 0,
      topDocs: (Array.isArray(row.topDocs) ? row.topDocs : []).map(String),
      stale: !!row.stale,
    };
  });
  const stats = d.shardsStats && typeof d.shardsStats === 'object'
    ? {
      total: Number(d.shardsStats.total) || 0,
      stale: Number(d.shardsStats.stale) || 0,
      threshold: Number(d.shardsStats.rebuildThreshold) || 0,
    }
    : null;
  return {
    present: !!(d.sessionId || d.pack || d.policy),
    sessionId: String(d.sessionId || ''),
    session: sess ? {
      state: String(sess.state || ''), phase: String(sess.phase || ''),
      lastAction: String(sess.lastAction || ''),
      heartbeats: Number(sess.heartbeats) || 0,
      lastBeat: String(sess.lastBeat || ''),
    } : null,
    takeBudget: Number(pack.takeBudget) || 0,
    takeBudgetDefault: Number(pack.takeBudgetDefault) || 0,
    tags: (Array.isArray(pack.tagSig) ? pack.tagSig : [])
      .map((t) => String((t && typeof t === 'object' ? t.tagKey : t) || '')).filter(Boolean),
    tagSigSource: String(pack.tagSigSource || 'none'),
    chunks, shards, stats,
    shardsFacePresent: Array.isArray(d.shards),
    degraded: pack.degraded !== false,
    degradedReason: String(pack.degradedReason || ''),
    policy: {
      shardMaxTokens: Number(pol.shardMaxTokens) || 0,
      minTakeBudget: Number(pol.minTakeBudget) || 0,
      maxTakeBudget: Number(pol.maxTakeBudget) || 0,
      maxTagsPerType: Number(pol.maxTagsPerType) || 5,
      maxTagsTotal: Number(pol.maxTagsTotal) || 20,
    },
    override: ov ? {
      revision: Number(ov.revision) || 0,
      takeBudget: Number(ov.takeBudget) || 0,
      updatedAt: String(ov.updatedAt || ''),
      note: String(ov.note || ''), by: String(ov.by || ''),
    } : null,
    canRevert: !!d.canRevert,
    writeOpen: !(d.writeGate && d.writeGate.open === false),
  };
}

/** 编辑分叉 payload 构建（表单输入 → PUT 载荷；非法=ok:false 携诚实错误词形）。
 * 乐观锁基准=面板当前 override.revision（无覆盖=0）；fieldManager 词形=console:web。 */
export function buildSavePayload(model, input) {
  const pol = (model && model.policy) || {};
  const min = Number(pol.minTakeBudget) || 0;
  const max = Number(pol.maxTakeBudget) || 0;
  const budget = Number(input && input.takeBudget);
  if (!Number.isFinite(budget) || budget < min || budget > max) {
    return { ok: false, error: tr('takeBudget 须为 ') + min + '–' + max + tr(' 的数值（ContextPolicy 预算域）') };
  }
  const { kept, dropped } = capTags(
    parseTagInput(input && input.tagList),
    Number(pol.maxTagsTotal) || 20, Number(pol.maxTagsPerType) || 5,
  );
  return {
    ok: true, dropped,
    payload: {
      takeBudget: Math.round(budget),
      tagSig: kept,
      note: String((input && input.note) == null ? '' : input.note).slice(0, 512),
      by: 'console:web',
      baseRevision: model && model.override ? Number(model.override.revision) || 0 : 0,
    },
  };
}

/* ---------- 渲染（一切动态文本过 esc——UI 红线） ---------- */

function sessionRowHtml(m) {
  if (!m.session) {
    return tr('<div class="small faint">会话不在活性账（无心跳行）——契约面仍可查看/编辑本会话上下文覆盖（不冒充在册）。</div>');
  }
  const s = m.session;
  const cls = s.state === 'suspected-stall' || s.state === 'escalated' ? 'esc' : s.state ? 'acc' : 'mute';
  return kv([
    ['state', s.state ? chip(s.state, cls) : '<span class="faint">—</span>'],
    ['phase', esc(s.phase || '—')],
    [tr('心跳'), esc(fmtNum(s.heartbeats)) + tr(' 次') + (s.lastBeat ? tr(' · 最近 ') + esc(relAge(s.lastBeat) || tr('时间未知')) : '')],
    [tr('动作'), esc(trunc(s.lastAction || '—', 80))],
  ]);
}

function chunksHtml(m) {
  if (!m.chunks.length) {
    return tr('<div class="small faint">无索引候选（tagSig 未装配或索引零命中——不冒充装配包内容）。</div>');
  }
  return m.chunks.map((c) => {
    const key = [c.kind, c.scope, c.name].filter(Boolean).join('/');
    // V-06：chunks（kind/scope/name+score）→ 对象工作台深链
    const go = c.kind && c.name ? objRoute(c.kind, c.scope || 'global', c.name) : '';
    const titleHtml = go
      ? '<span class="lk mono" data-go="' + esc(go) + '">' + esc(trunc(key, 64)) + '</span>'
      : '<b class="mono">' + esc(trunc(key, 64)) + '</b>';
    return '<div class="mini-row"' + (go ? ' data-go="' + esc(go) + '"' : '') + '><div class="mr-main"><div>' + titleHtml + ' ' +
      chip(String(Math.round(c.score)), c.score >= 60 ? 'ok' : 'mute') + '</div>' +
      (c.why ? '<div class="mr-sub small">' + esc(trunc(c.why, 90)) + '</div>' : '') + '</div></div>';
  }).join('');
}

function topDocRoute(d) {
  const s = String(d || '').trim();
  if (!s) return '';
  const p = s.split('/');
  if (p.length >= 3) return objRoute(p[0], p[1], p.slice(2).join('/'));
  if (p.length === 1) return docRoute('Doc', 'global', s);
  return '';
}

function shardsHtml(m) {
  if (!m.shardsFacePresent) {
    return tr('<div class="small faint">分片缓存面缺席（contextidx 未暴露 Shards——诚实降级，不冒充空缓存）。</div>');
  }
  if (!m.shards.length) {
    return tr('<div class="small faint">无生效 tag 分片（tagSig 未装配——编辑分叉后逐 tagKey 投影 TagCtxShard）。</div>');
  }
  return m.shards.map((s) => {
    const topHtml = s.topDocs.length
      ? s.topDocs.slice(0, 4).map((d) => {
        const go = topDocRoute(d);
        return go
          ? '<span class="lk mono small" data-go="' + esc(go) + '">' + esc(trunc(String(d), 40)) + '</span>'
          : '<span class="mono small faint">' + esc(trunc(String(d), 40)) + '</span>';
      }).join(' · ')
      : '';
    return '<div class="mini-row"><div class="mr-main"><div>' +
      chipGo(s.tagKey, tagRoute(s.tagKey), 'mute', true, tr('看同 tag 文档')) + ' ' +
      (s.stale ? chip('stale', 'warn') : chip('fresh', 'ok')) + '</div>' +
      tr('<div class="mr-sub small">相关对象 ') + esc(fmtNum(s.docs)) + tr(' 个') +
      (topHtml ? ' · ' + topHtml : '') + '</div></div></div>';
  }).join('');
}

function formHtml(m) {
  if (!m.writeOpen) {
    return tr('<div class="small faint">写面关断（--allow-apply 未开启）——编辑分叉不可用（与心跳注册面同开关，不冒充可编辑）。</div>');
  }
  return '<div class="filter-bar" data-ctx-form>' +
    '<input class="inp" data-ctx-budget type="number" min="' + m.policy.minTakeBudget + '" max="' + m.policy.maxTakeBudget + '" step="500" value="' + m.takeBudget + tr('" title="提取预算（token 数）">') +
    tr('<input class="inp" data-ctx-tags placeholder="tag 集：tagType:tag 逗号/空格分隔（每类≤') + m.policy.maxTagsPerType + tr(' 总≤') + m.policy.maxTagsTotal + '）">' +
    tr('<input class="inp" data-ctx-note placeholder="分叉备注（可选）">') +
    tr('<button class="btn btn-secondary sm" data-ctx-save>保存分叉</button>') +
    '<button class="btn btn-ghost sm" data-ctx-revert' + (m.canRevert ? '' : ' disabled') + tr('>撤销上次</button>') +
    '</div>';
}

/** 面板整卡渲染（纯函数——输入模型 → HTML 词形）。 */
export function renderSessionCtx(m) {
  const headChips =
    (m.degraded ? chip(tr('索引候选投影'), 'warn') : '') +
    (m.writeOpen ? chip(tr('写面开'), 'ok') : chip(tr('写面关断'), 'esc')) +
    (m.canRevert ? chip(tr('可撤销'), 'acc') : '');
  const ovLine = m.override
    ? tr('<div class="small faint" style="margin-top:6px">现行覆盖：revision ') + esc(fmtNum(m.override.revision)) +
      ' · takeBudget ' + esc(fmtNum(m.override.takeBudget)) + ' · by ' + esc(m.override.by || '—') +
      (m.override.note ? tr(' · 备注 ') + esc(trunc(m.override.note, 60)) : '') +
      (m.override.updatedAt ? ' · ' + esc(relAge(m.override.updatedAt) || '') : '') + '</div>'
    : tr('<div class="small faint" style="margin-top:6px">无覆盖（生效值=ContextPolicy 缺省——PUT 后此处现读覆盖行）。</div>');
  return tr('<div class="card-h"><span class="h-title">上下文（ContextPack 契约面）</span><span class="h-right">') + headChips + '</span></div>' +
    '<div class="card-b">' +
    sessionRowHtml(m) +
    '<div style="margin-top:10px">' + kv([
      ['takeBudget', '<b class="mono">' + esc(fmtNum(m.takeBudget)) + tr('</b> <span class="small faint">（缺省 ') + esc(fmtNum(m.takeBudgetDefault)) + tr(' · 域 ') + m.policy.minTakeBudget + '–' + m.policy.maxTakeBudget + '）</span>'],
      ['tagSig', m.tags.length
        ? m.tags.map((t) => chipGo(t, tagRoute(t), 'mute', true, tr('看同 tag 文档'))).join(' ')
        : tr('<span class="faint small">未装配（tagSigSource=') + esc(m.tagSigSource) + '）</span>'],
      [tr('分片缓存'), m.stats
        ? tr('共 ') + esc(fmtNum(m.stats.total)) + tr(' 片 · stale ') + esc(fmtNum(m.stats.stale)) + tr(' · 重建阈值 ') + esc(fmtNum(m.stats.threshold))
        : '<span class="faint">—</span>'],
      ['shardMaxTokens', '<span class="mono">' + esc(fmtNum(m.policy.shardMaxTokens)) + '</span>'],
    ]) + '</div>' +
    ovLine + '<div data-ctx-notice></div>' +
    '<div style="margin:10px 0 4px">' + formHtml(m) + '</div>' +
    tr('<h3 class="sec-title" style="margin:10px 0 6px">chunks（索引 TopK 候选）</h3>') + chunksHtml(m) +
    tr('<h3 class="sec-title" style="margin:10px 0 6px">相关 Tag 分片（TagCtxShard）</h3>') + shardsHtml(m) +
    (m.degraded && m.degradedReason ? '<div class="small faint" style="margin-top:8px">' + esc(m.degradedReason) + '</div>' : '') +
    '</div>';
}

/* ---------- 挂载（异步注水 + 事件接线） ---------- */

function focusIn(box) {
  return !!(document.activeElement && box.contains(document.activeElement));
}

function paint(box, ctx, id, model) {
  PANEL.set(id, { model });
  box.innerHTML = '<div class="card" data-ctx-card>' + renderSessionCtx(model) + '</div>';
  wireSessionCtx(box, ctx, id);
}

async function refresh(ctx, id, box) {
  const r = await ctx.api.sessionContext(id);
  if (!r.ok || focusIn(box)) return; // 读失败保留旧词形；编辑态不重绘（焦点保真）
  paint(box, ctx, id, normalizeContext(r.data));
}

function wireSessionCtx(box, ctx, id) {
  const q = (sel) => box.querySelector(sel);
  const notice = (html) => { const slot = q('[data-ctx-notice]'); if (slot) slot.innerHTML = html; };
  const saveBtn = q('[data-ctx-save]');
  if (saveBtn) saveBtn.addEventListener('click', async () => {
    const cur = (PANEL.get(id) || {}).model;
    if (!cur) return;
    const built = buildSavePayload(cur, {
      takeBudget: q('[data-ctx-budget]').value,
      tagList: q('[data-ctx-tags]').value,
      note: q('[data-ctx-note]').value,
    });
    if (!built.ok) { notice('<div class="small faint">' + esc(built.error) + '</div>'); return; }
    saveBtn.disabled = true;
    const r = await ctx.api.saveSessionContext(id, built.payload);
    saveBtn.disabled = false;
    if (!r.ok) {
      notice(tr('<div class="small faint">保存失败（') + esc(String(r.error || r.status)) + '）' +
        (r.status === 409 ? tr('——乐观锁冲突：请重进面板对齐 revision 后重试（稿不覆盖，不静默重基）') : '') + '</div>');
      return;
    }
    notice(tr('<div class="small faint">已保存（revision ') +
      esc(String((r.data && r.data.override && r.data.override.revision) || '?')) + '）' +
      (built.dropped.length ? tr('；超限 tag 收敛 ') + built.dropped.length + tr(' 个：') + esc(trunc(built.dropped.join(', '), 80)) : '') + '。</div>');
    refresh(ctx, id, box);
  });
  const revertBtn = q('[data-ctx-revert]');
  if (revertBtn) revertBtn.addEventListener('click', async () => {
    if (!window.confirm(tr('撤销最近一次上下文分叉（恢复服务端写前快照）？'))) return;
    revertBtn.disabled = true;
    const r = await ctx.api.revertSessionContext(id);
    revertBtn.disabled = false;
    if (!r.ok) { notice(tr('<div class="small faint">撤销失败（') + esc(String(r.error || r.status)) + tr('）——栈空不可撤（不冒充撤销成功）。</div>')); return; }
    notice(tr('<div class="small faint">已撤销——写前快照已恢复。</div>'));
    refresh(ctx, id, box);
  });
}

/** mountSessionCtx 会话屏上下文面板挂载（chat.js #/session/:id 调用——hash stamp 丢弃过期注水）。 */
export async function mountSessionCtx(el, ctx, id) {
  const box = el.querySelector('[data-ctx-panel]');
  if (!box || !id) return;
  if (focusIn(box)) return; // 编辑态零重绘（输入焦点保真——editor.js 同纪律）
  const cached = PANEL.get(id);
  if (cached && cached.model) box.innerHTML = '<div class="card" data-ctx-card>' + renderSessionCtx(cached.model) + '</div>';
  const stamp = typeof location !== 'undefined' ? location.hash : '';
  const r = await ctx.api.sessionContext(id);
  if ((typeof location !== 'undefined' ? location.hash : '') !== stamp) return; // 路由已变=丢弃
  if (!r.ok) {
    if (!focusIn(box)) {
      box.innerHTML = tr('<div class="card"><div class="card-b"><div class="small faint">上下文面暂不可达（') +
        esc(String(r.error || r.status)) + (r.status === 503 ? tr('——contextidx 未装配（防御面）') : '') +
        tr('）。恢复后重进即现读。</div></div></div>');
    }
    return;
  }
  if (focusIn(box)) return;
  paint(box, ctx, id, normalizeContext(r.data));
}
