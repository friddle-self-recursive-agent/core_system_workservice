/* misc.js —— 薄屏集合：系统树 / 文档库 / 设计文档 / 连接层 / 提案 / 系统详情。
 * 每屏职责单一但同族薄实现（合并一文件以少一份间接——总行数远低于 800 红线）。
 * 数据面：树/文档/连接/提案=快照聚合（store）；设计文档=GET /v1/design-docs 现读
 * （HEAD 权威零缓存——缺席=诚实降级）。展开态/过滤态=视图态（内存态，非业务状态）。 */

import { esc, trunc } from '../core/format.js';
import { chip, chipGo, kv, emptyBox, screenHead, kindBadge, objRoute, docRoute, involvedRoute, syncBadge } from '../render/chips.js';
import { renderMarkdown, rawDisclosure } from '../render/md.js';
import { t as tr } from '../core/i18n.js';

/* ---------- 视图态 ---------- */
const treeOpen = new Set();      // tree 已展开 kind 集合（内存视图态）
let treeSeeded = false;          // 首次渲染播种标记：对象数 ≤3 的 kind 自动展开一次，其后纯用户手控
const docFilter = { kind: '', q: '' };
const propOpen = new Set();      // 提案卡已展开集合（按 name 索引——内存视图态）
const dd = { tree: null, sel: null, cache: new Map() }; // designdoc 工作台态（HEAD 现读——cache 仅防抖同屏重绘）

/* ---------- 系统树（#/tree） ---------- */

/** 树行落点（2026-11-26 用户裁决：点系统默认=看 goal——设计文档渲染视图，不是对象工作台；
 * 文档/对象类仍进工作台。落点词形：ServiceDesc→doc 渲染（goal 是 ServiceDesc 的 goalRef）；
 * 其余 kind 保持 objRoute。 */
function treeLandingRoute(o) {
  if (o && o.kind === 'ServiceDesc' && o.name) {
    // goalRef 词形 system://<name>/goal —— 落设计文档仓的 goal.md（渲染视图）。
    return 'designdoc/' + encodeURIComponent(repoOfScope(o.scope || o.name)) + '/' + encodeURIComponent('goal.md');
  }
  return objRoute(o.kind, o.scope, o.name);
}

/** scope → 设计文档仓名（core_system 三仓布局：scope=系统名；默认 core_design）。
 * 声明侧（goal/desc/rule/concepts）都在 core_design 仓；实现侧（instance/workservice）是
 * 运行时设计，不在本树落点。外部系统（非本仓家族）=返回 core_design（一致落点，不发明第三仓）。 */
function repoOfScope(scope) {
  const s = String(scope || '').toLowerCase();
  if (s.includes('instance')) return 'core_system_instance';
  if (s.includes('workservice') || s.includes('deploy') || s.includes('storage') || s.includes('user')) return 'core_design';
  return 'core_design';
}

function treeScreen(el, ctx) {
  const objs = ctx.store.objects();
  const kinds = Array.from(new Set(objs.map((o) => o.kind))).sort();
  const scopeMap = {};
  objs.forEach((o) => {
    const k = (scopeMap[o.kind] = scopeMap[o.kind] || {});
    (k[o.scope] = k[o.scope] || []).push(o);
  });
  if (!treeSeeded) {
    // 默认展开规则（§19.3）：对象总数 ≤3 的 kind 首渲染自动进 treeOpen——窄列+全折叠=逐个点箭头不可用
    kinds.forEach((kind) => {
      if (objs.filter((o) => o.kind === kind).length <= 3) treeOpen.add(kind);
    });
    treeSeeded = true;
  }
  const body = kinds.map((kind) => {
    const open = treeOpen.has(kind);
    const n = objs.filter((o) => o.kind === kind).length;
    const rows = open
      ? Object.keys(scopeMap[kind] || {}).sort().map((scope) => {
        const items = (scopeMap[kind][scope] || []).slice().sort((a, b) => (a.name < b.name ? -1 : 1));
        return '<div class="tree-scope"><span class="small faint mono">' + esc(scope) + '/</span>' +
          items.map((o) => '<div class="tree-row leaf" data-go="' + esc(treeLandingRoute(o)) + '">' +
            '<span style="flex:1;min-width:0">' + esc(o.name) + '</span>' +
            ' <span class="faint tiny mono">r' + esc(String(o.rv || 0)) + '</span>' +
            (o.kind === 'ServiceDesc' ? '<span class="ws-chip">' + chipGo(tr('Service 大屏'), 'service/' + encodeURIComponent(String(o.name || '')), 'acc', true, tr('深链 Service 大屏')) + '</span>' : '') +
            '</div>').join('') + '</div>';
      }).join('')
      : '';
    return '<div class="tree-node">' +
      '<div class="tree-row grp" data-tree-toggle="' + esc(kind) + '"><b>' + esc(kind) + '</b>' + chip(n + tr(' 个'), n ? 'acc' : 'mute') +
      '<span class="faint">' + (open ? '▾' : '▸') + '</span></div>' + rows + '</div>';
  }).join('');
  el.innerHTML = '<div class="view-inner">' +
    screenHead(tr('系统树'), tr('kinds → scope → 对象（快照聚合投影；对象名可点进工作台；≤3 对象的 kind 默认展开）')) +
    (kinds.length ? '<div class="tree-box" style="min-width:min(680px,100%)">' + body + '</div>' : emptyBox(tr('快照暂无对象枚举'), tr('等待首个快照聚合。'))) +
    '</div>';
  el.querySelectorAll('[data-tree-toggle]').forEach((row) => {
    row.addEventListener('click', () => {
      const k = row.dataset.treeToggle;
      if (treeOpen.has(k)) treeOpen.delete(k); else treeOpen.add(k);
      treeScreen(el, ctx);
    });
  });
}

/* ---------- 文档库（#/docs） ---------- */
const DOC_KINDS = ['Goal', 'Rule', 'Runbook', 'DesignDoc', 'DriftReport', 'DocReview', 'Agent', 'WorkService'];

/* frontmatter 元数据小解析（tags 列表 + importance 数值）——数据源=docCache 缓存正文
 * （main.js hydrateDocs 预取）；未缓存=返回 null（列显示 —，不冒充）。 */
function fmMeta(content) {
  if (typeof content !== 'string' || !content.slice(0, 4000).includes('tags:')) return null;
  const head = content.slice(0, 4000);
  const tags = [];
  let importance = null;
  const flat = head.match(/tags:\s*\[([^\]]*)\]/);
  if (flat) flat[1].split(',').forEach((x) => { const v = x.trim().replace(/^['"]|['"]$/g, ''); if (v) tags.push(v); });
  const imp = head.match(/^importance:\s*(\d{1,3})\s*$/m);
  if (imp) importance = Number(imp[1]);
  return { tags, importance };
}

function docsScreen(el, ctx) {
  let objs = ctx.store.objects();
  const presentKinds = Array.from(new Set(objs.map((o) => o.kind)));
  const docKinds = DOC_KINDS.filter((k) => presentKinds.includes(k));
  const pool = docKinds.length ? objs.filter((o) => docKinds.includes(o.kind)) : objs; // 无声明 kind=全量兜底（诚实注记）
  let rows = pool;
  if (docFilter.kind) rows = rows.filter((o) => o.kind === docFilter.kind);
  if (docFilter.q) {
    const q = docFilter.q.toLowerCase();
    rows = rows.filter((o) => (o.kind + '/' + o.scope + '/' + o.name).toLowerCase().includes(q));
  }
  if (docFilter.tag) {
    const tg = docFilter.tag.toLowerCase();
    rows = rows.filter((o) => {
      const c = ctx.store.cachedDoc(o.kind, o.scope || 'global', o.name, o.rv);
      const meta = c ? fmMeta(c.content) : null;
      return !!(meta && meta.tags.some((x) => x.toLowerCase().includes(tg)));
    });
  }
  const draftSet = new Set(ctx.store.drafts().map((d) => d.kind + '|' + (d.scope || 'global') + '|' + d.name));
  el.innerHTML = '<div class="view-inner">' +
    screenHead(tr('文档库'), tr('目录树式文档清单——tag 过滤+版本/重要度列；点行进渲染视图与 AI 修文档'),
      docKinds.length ? '' : chip(tr('未注册声明 kind——全量枚举兜底'), 'mute')) +
    '<div class="card" style="margin-bottom:10px"><div class="card-b small">' +
    esc(tr('新建文档：对话提交需求（总览或会话屏）→ agent 起草 → Apply 落库——文档创建走对话+Apply，无独立表单。')) + ' ' +
    chipGo(tr('打开对话'), 'chat/core-headless', 'acc', true, tr('在会话屏提交需求——agent 起草后 Apply 落库')) +
    '</div></div>' +
    '<div class="filter-bar">' +
    tr('<input class="inp" id="docs-q" placeholder="按 kind/scope/name 过滤…" value="') + esc(docFilter.q) + '">' +
    tr('<input class="inp" id="docs-tag" placeholder="tag 过滤（缓存正文现匹配）…" value="') + esc(docFilter.tag || '') + '" style="max-width:220px">' +
    '<span class="fchip' + (!docFilter.kind ? ' on' : '') + tr('" data-docs-kind="">全部</span>') +
    docKinds.map((k) => '<span class="fchip' + (docFilter.kind === k ? ' on' : '') + '" data-docs-kind="' + esc(k) + '">' + esc(k) + '</span>').join('') +
    '</div>' +
    (rows.length
      ? '<div class="card">' + rows.slice(0, 300).map((o) => {
        const c = ctx.store.cachedDoc(o.kind, o.scope || 'global', o.name, o.rv);
        const meta = c ? fmMeta(c.content) : null;
        const tagChips = (meta && meta.tags ? meta.tags.slice(0, 4) : []).map((x) => chip(x, 'mute')).join('');
        const imp = meta && meta.importance != null ? chip('I:' + meta.importance, meta.importance >= 80 ? 'acc' : 'mute') : chip(tr('重要度—'), 'mute');
        const sync = syncBadge(ctx.store.runs(), o.name);
        return '<div class="mini-row" data-go="' + esc(docRoute(o.kind, o.scope, o.name)) + '">' +
          '<div class="mr-main"><div><b>' + esc(o.name) + '</b>' + (draftSet.has(o.kind + '|' + o.scope + '|' + o.name) ? '<span class="draft-dot">●</span>' : '') + '</div>' +
          '<div class="mr-sub small mono">' + esc(o.kind + '/' + o.scope) + ' · rv ' + esc(String(o.rv != null ? o.rv : '—')) + '</div>' +
          '<div class="mr-sub">' + tagChips + imp + sync + '</div></div>' +
          '<span class="ws-chip">' + chipGo(tr('打开'), docRoute(o.kind, o.scope, o.name), 'acc', true, tr('渲染视图')) + '</span></div>';
      }).join('') + (rows.length > 300 ? tr('<div class="small faint" style="padding:8px 14px">…共 ') + rows.length + tr(' 行——用过滤收窄。</div>') : '') + '</div>'
      : emptyBox(tr('窗口内无文档'), tr('过滤零命中或快照为空（空态显式）。'))) +
    '<div class="small faint" style="margin-top:8px">' + esc(tr('tag/重要度来自缓存正文的 frontmatter（预取窗口 24 篇——未缓存篇显示 —，现开即读）。')) + '</div>' +
    '</div>';
  const inp = el.querySelector('#docs-q');
  if (inp) inp.addEventListener('input', () => { docFilter.q = inp.value; docsScreen(el, ctx); });
  const tagInp = el.querySelector('#docs-tag');
  if (tagInp) tagInp.addEventListener('input', () => { docFilter.tag = tagInp.value; docsScreen(el, ctx); });
  el.querySelectorAll('[data-docs-kind]').forEach((c) => {
    c.addEventListener('click', () => { docFilter.kind = c.dataset.docsKind; docsScreen(el, ctx); });
  });
  el.querySelectorAll('.mini-row[data-go]').forEach((row) => {
    row.addEventListener('click', () => ctx.go(row.dataset.go));
  });
}

/* ---------- 设计文档工作台（#/designdoc[/:repo/:path]） ---------- */
async function hydrateDesignDoc(el, ctx, repo, path) {
  const stamp = typeof location !== 'undefined' ? location.hash : '';
  const cacheKey = repo + '|' + path;
  if (!dd.cache.has(cacheKey)) {
    const r = await ctx.api.designDoc(repo, path);
    if (!el.isConnected || (typeof location !== 'undefined' ? location.hash : '') !== stamp) return;
    dd.cache.set(cacheKey, r);
  }
  const res = dd.cache.get(cacheKey);
  const box = document.querySelector('[data-dd-body]');
  if (!box) return;
  if (!res || !res.ok) {
    box.innerHTML = tr('<div class="ob-empty">正文暂不可读（') + esc(String(res && res.error || res && res.status || tr('网络'))) + tr('）——HEAD 权威零缓存，重进即现读。</div>');
    return;
  }
  const data = res.data && typeof res.data === 'object' ? res.data : {};
  const content = data.content != null ? String(data.content) : (typeof res.data === 'string' ? res.data : '');
  box.innerHTML = '<div data-md-render="1">' + renderMarkdown(content) + '</div>' + rawDisclosure(tr('原文（HEAD）'), content) +
    '<div class="small faint" style="margin-top:6px">' + (data.sha ? 'HEAD ' + esc(String(data.sha).slice(0, 8)) : 'HEAD') +
    tr('（git 权威零缓存） · 草稿/Apply 通路经编辑器屏（doc/…）。</div>');
}

function designDocScreen(el, ctx) {
  const repo = ctx.params && ctx.params.repo;
  const path = ctx.params && ctx.params.path;
  if (repo && path) { dd.sel = { repo, path }; }
  const sel = dd.sel;
  el.innerHTML = '<div class="view-inner">' +
    screenHead(tr('设计文档'), tr('白名单仓树（GET /v1/design-docs——HEAD 权威零缓存）')) +
    tr('<div class="dd-grid"><div data-dd-tree><div class="small faint">仓树现读中…</div></div>') +
    '<div class="card"><div class="card-b" data-dd-body>' + (sel ? tr('<div class="small faint">正文现读中…</div>') : emptyBox(tr('未选档'), tr('左侧选择仓文件（默认首仓首文件）。'))) + '</div></div></div></div>';

  ctx.api.designDocs().then((r) => {
    if (!el.isConnected) return;
    const treeBox = el.querySelector('[data-dd-tree]');
    if (!treeBox) return;
    if (!r.ok) {
      treeBox.innerHTML = tr('<div class="ob-empty">仓树暂不可读（') + esc(String(r.error || r.status)) + '）。</div>';
      return;
    }
    dd.tree = r.data;
    const files = Array.isArray(r.data) ? r.data : (r.data && r.data.files) || [];
    if (!files.length) { treeBox.innerHTML = tr('<div class="ob-empty">仓树为空（诚实空态）。</div>'); return; }
    treeBox.innerHTML = files.map((f) => {
      const fr = f.repo || (r.data.repo || '');
      const fp = f.path || String(f);
      const key = fr + '|' + fp;
      return '<div class="mini-row' + (sel && sel.repo === fr && sel.path === fp ? ' sel' : '') + '" data-dd-file="' + esc(key) + '">' +
        '<div class="mr-main"><div class="mono small">' + esc(trunc(fp, 40)) + '</div><div class="mr-sub tiny">' + esc(fr) + '</div></div></div>';
    }).join('');
    treeBox.querySelectorAll('[data-dd-file]').forEach((row) => {
      row.addEventListener('click', () => {
        const [fr, ...rest] = row.dataset.ddFile.split('|');
        const fp = rest.join('|');
        ctx.go('designdoc/' + encodeURIComponent(fr) + '/' + encodeURIComponent(fp));
      });
    });
    if (!sel && files.length) {
      const f0 = files[0];
      ctx.go('designdoc/' + encodeURIComponent(String(f0.repo || r.data.repo || '')) + '/' + encodeURIComponent(String(f0.path || f0)));
      return;
    }
    if (sel) hydrateDesignDoc(el, ctx, sel.repo, sel.path);
  }).catch(() => { /* 降级词形在 then 分支之外——网络异常=树盒保持现读词形 */ });
}

/* ---------- 连接层（#/connectors） ---------- */
function connectorsScreen(el, ctx) {
  const kinds = ['ConnectorOps', 'WorkService', 'Provider', 'Ref'];
  const secs = kinds.map((kind) => {
    const rows = ctx.store.refs(kind);
    return '<section class="card"><div class="card-h"><span class="h-title">' + esc(kind) + '</span>' +
      '<span class="h-right">' + chip(rows.length + tr(' 个'), rows.length ? 'acc' : 'mute') + '</span></div>' +
      '<div class="card-b">' + (rows.length
        ? rows.map((o) => '<div class="mini-row" data-go="' + esc(docRoute(kind, o.scope, o.name)) + '">' +
          '<div class="mr-main"><div><b class="mono">' + esc(o.name) + '</b></div>' +
          '<div class="mr-sub small mono">' + esc(o.scope) + ' · r' + esc(String(o.rv || 0)) + '</div></div>' +
          '<span class="ws-chip">' + chipGo(tr('工作台'), objRoute(kind, o.scope, o.name), 'mute', true) + '</span></div>').join('')
        : tr('<div class="small faint">该 kind 暂无在册对象（空态显式）。</div>')) + '</div></section>';
  }).join('');
  el.innerHTML = '<div class="view-inner">' + screenHead(tr('连接层'), tr('ConnectorOps / WorkService / Provider / Ref 四族（文档与资源同规则）')) +
    '<div class="grid cols-2">' + secs + '</div></div>';
  el.querySelectorAll('.mini-row[data-go]').forEach((row) => {
    row.addEventListener('click', () => ctx.go(row.dataset.go));
  });
}

/* ---------- 需求提案（#/proposals） ---------- */
function proposalsScreen(el, ctx) {
  const rows = ctx.store.proposals();
  const cards = rows.map((p, i) => {
    const keys = p && typeof p === 'object' ? Object.keys(p) : [];
    const key = String((p && p.name) || '') || '#' + i; // 无 name 提案退化为序号键（展开态仍可用）
    const open = propOpen.has(key);
    const known = ['title', 'summary', 'status', 'proposer', 'createdAt', 'scope', 'name'];
    const sub = known.filter((k) => p[k] != null && k !== 'title').map((k) => esc(k) + ': ' + esc(trunc(String(p[k]), 40))).join(' · ');
    // 展开区要素：known 之外的对象字段（acceptance/impact/priority/origin/landing 等）有则显示
    const extras = keys.filter((k) => !known.includes(k) && k !== 'summary' && p[k] != null && p[k] !== '')
      .map((k) => '<span class="k">' + esc(k) + '</span><span class="v">' + esc(trunc(String(p[k]), 200)) + '</span>').join('');
    return '<div class="mini-row" data-prop="' + esc(key) + '">' +
      '<div class="mr-main"><div><b>' + esc(String(p.title || p.name || tr('未命名提案'))) + '</b> ' + chip(String(p.status || tr('待定')), 'mute') + '</div>' +
      '<div class="mr-sub small">' + (sub || tr('<span class="faint">要素字段缺席——不编造</span>')) + '</div>' +
      (open
        ? '<div class="mr-sub" style="margin-top:6px;white-space:pre-wrap">' + esc(trunc(String(p.summary || ''), 2000) || tr('summary 缺席——不编造')) + '</div>' +
          (extras ? '<div class="kv" style="margin-top:6px">' + extras + '</div>' : '') +
          '<div class="small faint" style="margin-top:6px">' + esc(tr('再点一次卡片收起。')) + '</div>'
        : '') +
      '</div>' +
      '<span class="ws-chip">' + chipGo(tr('去审批'), 'approvals', 'acc', true) + '</span>' +
      '</div>';
  }).join('');
  el.innerHTML = '<div class="view-inner">' + screenHead(tr('需求提案'), tr('提案面（点卡片展开 summary 全文与要素）——批准/驳回走审批与 Apply（LLM 无 Apply 权）')) +
    (rows.length ? '<div class="card">' + cards + '</div>'
      : emptyBox(tr('暂无提案'), tr('快照 proposals 聚合为空（空态显式）。'))) +
    '</div>';
  el.querySelectorAll('[data-prop]').forEach((row) => {
    row.addEventListener('click', (ev) => {
      if (ev.target.closest('[data-go]')) return; // 「去审批」chip 走自身深链——不触发展开
      const key = row.dataset.prop;
      if (propOpen.has(key)) propOpen.delete(key); else propOpen.add(key);
      proposalsScreen(el, ctx);
    });
  });
}

/* ---------- 系统详情（#/s/:name——scope 视角的对象清单+摘要） ---------- */
function systemScreen(el, ctx) {
  const scope = String(ctx.params.name || '');
  const objs = ctx.store.objects().filter((o) => o.scope === scope);
  const evs = ctx.store.events().filter((e) => String(e.involved || '').includes(scope)).slice(0, 10);
  const byKind = {};
  objs.forEach((o) => { byKind[o.kind] = (byKind[o.kind] || 0) + 1; });
  el.innerHTML = '<div class="view-inner">' +
    screenHead(scope, 'scope=' + scope + tr(' 的对象与事件（快照聚合投影）'), chipGo(tr('系统树'), 'tree', 'mute', false)) +
    '<div class="statbar">' +
    Object.keys(byKind).sort().map((k) =>
      '<div class="stat" data-go="' + esc('obj/' + encodeURIComponent(k)) + '" role="link"><div class="st-num">' + byKind[k] + '</div>' +
      '<div class="st-label">' + esc(k) + '</div></div>').join('') +
    '</div>' +
    (objs.length
      ? tr('<h2 class="sec-title">对象</h2><div class="card">') + objs.map((o) =>
        '<div class="mini-row" data-go="' + esc(objRoute(o.kind, o.scope, o.name)) + '">' +
        '<div class="mr-main"><div><b>' + esc(o.name) + '</b> ' + kindBadge(o.kind) + '</div>' +
        '<div class="mr-sub small mono">r' + esc(String(o.rv || 0)) + '</div></div></div>').join('') + '</div>'
      : emptyBox(tr('该 scope 暂无在册对象'), tr('不装假——scope 词形可能为实例桶（树内可见）。'))) +
    (evs.length ? tr('<h2 class="sec-title">最近相关事件</h2><div class="card">') + evs.map((e) =>
      '<div class="mini-row" data-go="event/' + encodeURIComponent(String(e.id ?? '')) + '">' +
      '<div class="mr-main"><div>' + esc(String(e.type || '')) + '</div>' +
      '<div class="mr-sub small mono">' + esc(trunc(String(e.involved || '—'), 60)) + '</div></div></div>').join('') + '</div>' : '') +
    '</div>';
  el.querySelectorAll('[data-go]').forEach((x) => {
    x.addEventListener('click', () => ctx.go(x.dataset.go));
  });
}

export function render(el, ctx) {
  const r = ctx.route;
  if (r === 'tree') treeScreen(el, ctx);
  else if (r === 'docs') docsScreen(el, ctx);
  else if (r === 'designdoc') designDocScreen(el, ctx);
  else if (r === 'connectors') connectorsScreen(el, ctx);
  else if (r === 'proposals') proposalsScreen(el, ctx);
  else if (r === 's') systemScreen(el, ctx);
  else el.innerHTML = '<div class="view-inner">' + emptyBox(tr('未知薄屏'), String(r)) + '</div>';
}

/* 业务控制台占位屏（2026-11-26 用户裁决：先不加东西——显式占位词形，不装内容）。
 * 后续 ServiceAgent 登录后此级承载业务视角（Service 维度的业务大屏与操作面）。 */
export function renderBiz(el, ctx) {
  el.innerHTML = '<div class="view-inner">' +
    '<div class="biz-hero">' +
    '<div class="biz-hero-mark">B</div>' +
    '<h2 class="sec-title">' + esc(tr('业务控制台')) + '</h2>' +
    '<div class="small faint">' + esc(tr('业务视角占位——后续 ServiceAgent 登录后在此承载业务大屏与操作面（当前零内容，显式占位不装假）。')) + '</div>' +
    '</div></div>';
}
