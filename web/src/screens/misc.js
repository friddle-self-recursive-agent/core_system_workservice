/* misc.js —— 薄屏集合：系统树 / 文档库 / 设计文档 / 连接层 / 提案 / 系统详情。
 * 每屏职责单一但同族薄实现（合并一文件以少一份间接——总行数远低于 800 红线）。
 * 数据面：树/文档/连接/提案=快照聚合（store）；设计文档=GET /v1/design-docs 现读
 * （HEAD 权威零缓存——缺席=诚实降级）。展开态/过滤态=视图态（内存态，非业务状态）。 */

import { esc, trunc, relAge } from '../core/format.js';
import { chip, chipGo, kv, emptyBox, screenHead, kindBadge, objRoute, docRoute, involvedRoute, syncBadge, tagRoute } from '../render/chips.js';
import { renderMarkdown, rawDisclosure, splitFrontmatter } from '../render/md.js';
import { paintMetaHeader } from '../render/meta-header.js';
import { frontmatterTags, tagsOfRef, loadContextIndexDocs, cachedIndexTagsOf, hasCachedIndex } from './tags.js';
import { t as tr } from '../core/i18n.js';

/* ---------- 视图态 ---------- */
const treeOpen = new Set();      // tree 已展开 kind 集合（内存视图态）
let treeSeeded = false;          // 首次渲染播种标记：对象数 ≤3 的 kind 自动展开一次，其后纯用户手控
const docFilter = { kind: '', q: '', tag: '', routeTag: '' };
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
    (kinds.length ? '<div class="tree-box" style="min-width:min(680px,100%)">' + body + '</div>' : emptyBox(tr('暂无对象'), tr('等待首个快照聚合。'))) +
    '</div>';
  el.querySelectorAll('[data-tree-toggle]').forEach((row) => {
    row.addEventListener('click', () => {
      const k = row.dataset.treeToggle;
      if (treeOpen.has(k)) treeOpen.delete(k); else treeOpen.add(k);
      treeScreen(el, ctx);
    });
  });
}

/* ---------- 文档库（#/docs 与 #/docs/tag/:tag） ---------- */
// V-19/F-06：文档族 kind 词表保留导出（tags 屏与机读判据消费）；文档库列表已改为在册
// 全量对象枚举（不再按本表白名单过滤——F-06 18 对象失明缺口收口）。
export const DOC_KINDS = ['Goal', 'Rule', 'Runbook', 'DesignDoc', 'DriftReport', 'DocReview', 'BugDoc', 'Agent', 'WorkService'];

/* frontmatter 元数据小解析（tags 列表 + importance 数值）——数据源=docCache 缓存正文
 * （main.js hydrateDocs 预取）；未缓存=返回 null（列显示 —，不冒充）。 */
function fmMeta(content) {
  if (typeof content !== 'string') return null;
  const tags = frontmatterTags(content);
  const head = content.slice(0, 4000);
  const imp = head.match(/^importance:\s*(\d{1,3})\s*$/m);
  const importance = imp ? Number(imp[1]) : null;
  if (!tags.length && importance == null) return null;
  return { tags, importance };
}

/** 某对象三源去重 tags（context/index ∪ ref tags ∪ docCache frontmatter）。 */
function objectTagsOf(ctx, o) {
  const set = new Set();
  cachedIndexTagsOf(o.kind, o.scope || 'global', o.name).forEach((x) => set.add(x));
  tagsOfRef(o).forEach((x) => set.add(x));
  const c = ctx.store.cachedDoc(o.kind, o.scope || 'global', o.name, o.rv);
  if (c) frontmatterTags(c.content).forEach((x) => set.add(x));
  return Array.from(set);
}

let docsIdxLoading = false;

function docsScreen(el, ctx) {
  // V-01：#/docs/tag/:tag 第三段=tag 过滤预填；切回裸 #/docs 时清除旧路由预填态
  const routeTag = String((ctx.params && ctx.params.tag) || '');
  if (routeTag) {
    if (docFilter.routeTag !== routeTag) { docFilter.tag = routeTag; docFilter.routeTag = routeTag; }
  } else if (docFilter.routeTag) {
    docFilter.tag = '';
    docFilter.routeTag = '';
  }

  let objs = ctx.store.objects();
  const presentKinds = Array.from(new Set(objs.map((o) => o.kind)));
  // F-06：文档库不再按 DOC_KINDS 白名单失明过滤——在册全量对象默认可见（kind chips=
  // 在册 kind 动态枚举，随快照现势；文档族 kind 置前排序，工程对象族殿后）。
  const DOC_FIRST = ['Goal', 'Rule', 'Runbook', 'DesignDoc', 'DriftReport', 'DocReview', 'BugDoc'];
  const kindChips = presentKinds.slice().sort((a, b) => {
    const ia = DOC_FIRST.indexOf(a), ib = DOC_FIRST.indexOf(b);
    return (ia < 0 ? DOC_FIRST.length + 1 : ia) - (ib < 0 ? DOC_FIRST.length + 1 : ib) || (a < b ? -1 : 1);
  });
  // 按 tag 过滤且未指定 kind 时跨全量对象命中（确保与 #/tags 三源聚合计数一致）
  const pool = objs;
  let rows = pool;
  if (docFilter.kind) rows = rows.filter((o) => o.kind === docFilter.kind);
  if (docFilter.q) {
    const q = docFilter.q.toLowerCase();
    rows = rows.filter((o) => (o.kind + '/' + o.scope + '/' + o.name).toLowerCase().includes(q));
  }
  if (docFilter.tag) {
    const tg = docFilter.tag.toLowerCase();
    rows = rows.filter((o) => objectTagsOf(ctx, o).some((x) => x.toLowerCase().includes(tg)));
    if (!hasCachedIndex() && !docsIdxLoading && ctx.api) {
      docsIdxLoading = true;
      const stamp = typeof location !== 'undefined' ? location.hash : '';
      loadContextIndexDocs(ctx.api).then((res) => {
        docsIdxLoading = false;
        if (res && res.ok && el.isConnected && (typeof location !== 'undefined' ? location.hash : '') === stamp) {
          docsScreen(el, ctx);
        }
      }).catch(() => { docsIdxLoading = false; });
    }
  }
  const draftSet = new Set(ctx.store.drafts().map((d) => d.kind + '|' + (d.scope || 'global') + '|' + d.name));
  el.innerHTML = '<div class="view-inner">' +
    screenHead(tr('文档库'), tr('全部文档与对象——按标签/类型过滤；点行进渲染视图'),
      chipGo(tr('标签聚合'), 'tags', 'mute', false)) +
    '<div class="card" style="margin-bottom:10px"><div class="card-b small">' +
    esc(tr('新建文档：对话提交需求（总览或会话屏）→ agent 起草 → Apply 落库——文档创建走对话+Apply，无独立表单。')) + ' ' +
    chipGo(tr('打开对话'), 'chat/core-headless', 'acc', true, tr('在会话屏提交需求——agent 起草后 Apply 落库')) +
    (docFilter.tag ? ' <span class="chip warn tiny click" data-clear-tag="1" title="' + esc(tr('清除 tag 过滤')) + '">' + esc(tr('清除 tag：')) + esc(docFilter.tag) + ' ×</span>' : '') +
    (routeTag ? '<div class="small faint" style="margin-top:6px">' + esc(tr('当前 URL 直链 tag 过滤（#/docs/tag/')) + esc(routeTag) + esc(tr('）——三源真实命中，可刷新恢复。')) + '</div>' : '') +
    '</div></div>' +
    '<div class="filter-bar">' +
    tr('<input class="inp" id="docs-q" placeholder="按 kind/scope/name 过滤…" value="') + esc(docFilter.q) + '">' +
    tr('<input class="inp" id="docs-tag" placeholder="tag 过滤（三源聚合现匹配）…" value="') + esc(docFilter.tag || '') + '" style="max-width:220px">' +
    '<span class="fchip' + (!docFilter.kind ? ' on' : '') + tr('" data-docs-kind="">全部</span>') +
    kindChips.map((k) => '<span class="fchip' + (docFilter.kind === k ? ' on' : '') + '" data-docs-kind="' + esc(k) + '">' + esc(k) + '</span>').join('') +
    '</div>' +
    (rows.length
      ? '<div class="card">' + rows.slice(0, 300).map((o) => {
        const c = ctx.store.cachedDoc(o.kind, o.scope || 'global', o.name, o.rv);
        const meta = c ? fmMeta(c.content) : null;
        const allTags = objectTagsOf(ctx, o);
        const tagChips = allTags.slice(0, 4).map((x) => chipGo('#' + x, tagRoute(x), 'mute', true, tr('看同 tag 文档'))).join(' ');
        const imp = meta && meta.importance != null ? chip('I:' + meta.importance, meta.importance >= 80 ? 'acc' : 'mute') : chip(tr('重要度—'), 'mute');
        const sync = syncBadge(ctx.store.runs(), o.name);
        const scope = o.scope || 'global';
        const hasDraft = draftSet.has(o.kind + '|' + scope + '|' + o.name);
        // V-13：draft-dot 可点→approvals 对应 diff 卡深链
        const draftDot = hasDraft
          ? ' <span class="draft-dot click" data-go="approvals/' + encodeURIComponent(o.kind + '/' + scope + '/' + o.name) +
            '" title="' + esc(tr('有未落库草稿——点进审批 diff 卡')) + '">● ' + esc(tr('草稿待审')) + '</span>'
          : '';
        return '<div class="mini-row" data-doc-row="' + esc(docRoute(o.kind, scope, o.name)) + '">' +
          '<div class="mr-main"><div><b>' + esc(o.name) + '</b>' + draftDot + '</div>' +
          '<div class="mr-sub small mono">' + esc(o.kind + '/' + scope) + ' · rv ' + esc(String(o.rv != null ? o.rv : '—')) + '</div>' +
          '<div class="mr-sub" style="display:flex;gap:4px;flex-wrap:wrap;align-items:center">' + tagChips + ' ' + imp + ' ' + sync + '</div></div>' +
          '<span class="ws-chip">' + chipGo(tr('打开'), docRoute(o.kind, scope, o.name), 'acc', true, tr('渲染视图')) + '</span></div>';
      }).join('') + (rows.length > 300 ? tr('<div class="small faint" style="padding:8px 14px">…共 ') + rows.length + tr(' 行——用过滤收窄。</div>') : '') + '</div>'
      : emptyBox(tr('暂无文档'), tr('过滤零命中或快照为空。'))) +
    '<div class="small faint" style="margin-top:8px">' + esc(tr('tag 来自三源真实聚合（/v1/context/index ∪ 对象 tags ∪ 预取 24 篇缓存 frontmatter）；重要度来自缓存正文。')) + '</div>' +
    '</div>';
  const inp = el.querySelector('#docs-q');
  if (inp) inp.addEventListener('input', () => { docFilter.q = inp.value; docsScreen(el, ctx); });
  const tagInp = el.querySelector('#docs-tag');
  if (tagInp) tagInp.addEventListener('input', () => { docFilter.tag = tagInp.value; docFilter.routeTag = ''; docsScreen(el, ctx); });
  const clr = el.querySelector('[data-clear-tag]');
  if (clr) clr.addEventListener('click', () => { docFilter.tag = ''; docFilter.routeTag = ''; ctx.go('docs'); });
  el.querySelectorAll('[data-docs-kind]').forEach((c) => {
    c.addEventListener('click', () => { docFilter.kind = c.dataset.docsKind; docsScreen(el, ctx); });
  });
  el.querySelectorAll('.mini-row[data-doc-row]').forEach((row) => {
    row.style.cursor = 'pointer';
    row.addEventListener('click', (ev) => {
      if (ev.target && ev.target.closest('[data-go]')) return; // tag / Running / draft-dot 走自身深链
      ctx.go(row.dataset.docRow);
    });
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
  const { body } = splitFrontmatter(content);
  // R-UX-10：正文先行——frontmatter 不再占首屏（原文折叠可查）。
  // 元数据 header（2026-11-28 用户裁决）：metadatainfo 侧车上屏公用组件（HEAD 文档账=sha 诚实）。
  box.innerHTML = '<div data-meta-header-slot style="margin-bottom:8px"></div><div data-dd-bodywrap></div>';
  paintMetaHeader(box.querySelector('[data-meta-header-slot]'), ctx,
    { kind: 'DesignDoc', scope: repo, name: path }, content, { head: true, headSha: String(data.sha || '') });
  box.querySelector('[data-dd-bodywrap]').innerHTML =
    '<div data-md-render="1">' + renderMarkdown(body) + '</div>' + rawDisclosure(tr('原文（HEAD）'), content) +
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
    // F-07：/v1/design-docs 载荷键=docs（数组行 {repo,path}；旧词形 files/裸数组兼容读——
    // 不再读不存在的 r.data.files 致 107 份真实文档全显「文档为空」）。
    const files = Array.isArray(r.data) ? r.data : (Array.isArray(r.data && r.data.docs) ? r.data.docs : ((r.data && r.data.files) || []));
    if (!files.length) { treeBox.innerHTML = tr('<div class="ob-empty">仓树为空。</div>'); return; }
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

/** V-21：通知送达投影片渲染（纯函数——node 单测直测；数据=GET /v1/notifications/deliveries）。 */
export function deliveriesPanelHtml(data) {
  if (!data || typeof data !== 'object') {
    return '<div class="small faint">' + esc(tr('送达投影面未就绪（GET /v1/notifications/deliveries 缺席——不冒充空）。')) + '</div>';
  }
  const rows = Array.isArray(data.deliveries) ? data.deliveries : [];
  const summary = '<div style="display:flex;gap:6px;flex-wrap:wrap;margin-bottom:8px">' +
    chip(tr('总 ') + String(data.count != null ? data.count : rows.length), 'mute') +
    (data.accepted != null ? chip('accepted ' + data.accepted, 'warn') : '') +
    (data.sent != null ? chip('sent ' + data.sent, 'ok') : '') +
    (data.failed != null ? chip('failed ' + data.failed, Number(data.failed) > 0 ? 'esc' : 'mute') : '') +
    '</div>';
  if (!rows.length) {
    return summary + '<div class="small faint">' + esc(tr('送达投影暂无记录（accepted→sent/failed 出站时此处实时落行）。')) + '</div>';
  }
  const stCls = (st) => st === 'sent' ? 'ok' : (st === 'failed' ? 'esc' : (st === 'accepted' ? 'warn' : 'mute'));
  return summary + rows.slice(0, 30).map((d) => {
    const st = String((d && d.status) || 'unknown');
    const att = d && d.attempts != null ? Number(d.attempts) : 0;
    const eid = d && d.eventId ? String(d.eventId) : '';
    const goAttr = eid ? ' data-go="event/' + encodeURIComponent(eid) + '" role="link" style="cursor:pointer"' : '';
    return '<div class="mini-row"' + goAttr + '><div class="mr-main"><div>' +
      chip(st, stCls(st)) + ' <b>' + esc(String((d && d.type) || '—')) + '</b> ' +
      chip('channel: ' + String((d && d.channel) || '*'), 'mute') + ' ' +
      chip('attempts: ' + att, att > 0 ? 'acc' : 'mute') + '</div>' +
      '<div class="mr-sub small mono">' + esc(trunc(String((d && (d.dedupKey || d.route)) || '—'), 64)) +
      (d && d.lastError ? ' · ' + esc(trunc(String(d.lastError), 60)) : '') + '</div></div>' +
      '<span class="small faint">' + esc(relAge(d && (d.completedAt || d.at)) || '') + '</span></div>';
  }).join('');
}

/* ---------- 连接层（#/connectors——V-12 互链 workflows + V-21 Notification 与送达投影） ---------- */
function connectorsScreen(el, ctx) {
  const kinds = ['ConnectorOps', 'WorkService', 'Provider', 'Ref', 'Notification'];
  const secs = kinds.map((kind) => {
    const rows = ctx.store.refs(kind);
    return '<section class="card"><div class="card-h"><span class="h-title">' + esc(kind) + '</span>' +
      '<span class="h-right">' + chip(rows.length + tr(' 个'), rows.length ? 'acc' : 'mute') + '</span></div>' +
      '<div class="card-b">' + (rows.length
        ? rows.map((o) => '<div class="mini-row" data-conn-row="' + esc(docRoute(kind, o.scope, o.name)) + '">' +
          '<div class="mr-main"><div><b class="mono">' + esc(o.name) + '</b></div>' +
          '<div class="mr-sub small mono">' + esc(o.scope) + ' · r' + esc(String(o.rv || 0)) + '</div></div>' +
          '<span class="ws-chip">' + chipGo(tr('工作台'), objRoute(kind, o.scope, o.name), 'mute', true) + '</span></div>').join('')
        : tr('<div class="small faint">该 kind 暂无在册对象。</div>')) + '</div></section>';
  }).join('');
  const snapDel = ctx.store.deliveries ? ctx.store.deliveries() : null;
  el.innerHTML = '<div class="view-inner">' +
    screenHead(tr('连接层'), tr('ConnectorOps / WorkService / Provider / Ref / Notification 五族 + 通知送达投影 + 绑定工作流互链'),
      chipGo(tr('工作流'), 'workflows', 'mute', false)) +
    '<div class="grid cols-2">' + secs + '</div>' +
    '<h2 class="sec-title" style="margin-top:16px">' + esc(tr('通知送达投影（/v1/notifications/deliveries）')) + '</h2>' +
    '<div class="card"><div class="card-b" data-conn-deliveries>' +
    (snapDel ? deliveriesPanelHtml(snapDel) : '<div class="small faint">' + esc(tr('送达投影现读中…')) + '</div>') +
    '</div></div>' +
    '<h2 class="sec-title" style="margin-top:16px">' + esc(tr('绑定工作流（workflows ↔ connectors 互链）')) + '</h2>' +
    '<div class="card"><div class="card-b" data-conn-wfs><div class="small faint">' + esc(tr('工作流绑定现读中…')) + '</div></div></div>' +
    '</div>';

  el.querySelectorAll('.mini-row[data-conn-row]').forEach((row) => {
    row.style.cursor = 'pointer';
    row.addEventListener('click', (ev) => {
      if (ev.target && ev.target.closest('[data-go]')) return;
      ctx.go(row.dataset.connRow);
    });
  });

  const stamp = typeof location !== 'undefined' ? location.hash : '';
  // V-21：实时现读 /v1/notifications/deliveries
  if (ctx.api && typeof ctx.api.deliveries === 'function') {
    ctx.api.deliveries().then((r) => {
      if (!el.isConnected || (typeof location !== 'undefined' ? location.hash : '') !== stamp) return;
      const box = el.querySelector('[data-conn-deliveries]');
      if (!box) return;
      if (r && r.ok) box.innerHTML = deliveriesPanelHtml(r.data);
      else if (!snapDel) box.innerHTML = '<div class="small faint">' + esc(tr('送达投影暂不可读（')) + esc(String((r && (r.error || r.status)) || '')) + '）</div>';
    }).catch(() => {});
  }
  // V-12：反查在册工作流绑定的 connector，回链 wf/:name
  if (ctx.api && typeof ctx.api.workflows === 'function') {
    ctx.api.workflows().then((r) => {
      if (!el.isConnected || (typeof location !== 'undefined' ? location.hash : '') !== stamp) return;
      const box = el.querySelector('[data-conn-wfs]');
      if (!box) return;
      const wfs = (r && r.ok && r.data && Array.isArray(r.data.workflows)) ? r.data.workflows : [];
      if (!wfs.length) { box.innerHTML = '<div class="small faint">' + esc(tr('暂无在册工作流绑定。')) + '</div>'; return; }
      box.innerHTML = wfs.map((w) => {
        const nm = String((w && w.name) || '');
        const ex = (w && w.execution) || {};
        const conn = String(ex.connector || '') + (ex.connectorScope ? '/' + String(ex.connectorScope) : '');
        return '<div class="mini-row" data-go="wf/' + encodeURIComponent(nm) + '" style="cursor:pointer">' +
          '<div class="mr-main"><div><b class="mono">' + esc(nm) + '</b> ' + chip(conn || tr('未绑定'), conn ? 'acc' : 'mute') + '</div>' +
          '<div class="mr-sub small">' + esc(String(ex.state || '—')) + '</div></div>' +
          '<span class="ws-chip">' + chipGo(tr('工作流详情'), 'wf/' + encodeURIComponent(nm), 'mute', true) + '</span></div>';
      }).join('');
    }).catch(() => {});
  }
}

/* ---------- 需求提案（#/proposals 与 #/prop/:name） ---------- */
function proposalsScreen(el, ctx) {
  const rows = ctx.store.proposals();
  const focusName = String((ctx.params && ctx.params.name) || '').trim();
  if (focusName) propOpen.add(focusName);
  const cards = rows.map((p, i) => {
    const keys = p && typeof p === 'object' ? Object.keys(p) : [];
    const key = String((p && p.name) || '') || '#' + i;
    const open = propOpen.has(key);
    const isFocus = !!focusName && key === focusName;
    const known = ['title', 'summary', 'status', 'proposer', 'createdAt', 'scope', 'name'];
    const sub = known.filter((k) => p[k] != null && k !== 'title').map((k) => esc(k) + ': ' + esc(trunc(String(p[k]), 40))).join(' · ');
    const extras = keys.filter((k) => !known.includes(k) && k !== 'summary' && p[k] != null && p[k] !== '')
      .map((k) => '<span class="k">' + esc(k) + '</span><span class="v">' + esc(trunc(String(p[k]), 200)) + '</span>').join('');
    // V-14：「去审批」携带提案标识 approvals/:focus；并附对象工作台与直达链接 prop/:name
    const approveRoute = p && p.name ? 'approvals/' + encodeURIComponent(String(p.name)) : 'approvals';
    const propLink = p && p.name ? 'prop/' + encodeURIComponent(String(p.name)) : '';
    return '<div class="mini-row" data-prop="' + esc(key) + '"' + (isFocus ? ' data-prop-focused="1" style="outline:2px solid var(--accent)"' : '') + '>' +
      '<div class="mr-main"><div><b>' + esc(String(p.title || p.name || tr('未命名提案'))) + '</b> ' + chip(String(p.status || tr('待定')), 'mute') + '</div>' +
      '<div class="mr-sub small">' + (sub || tr('<span class="faint">要素字段缺席</span>')) + '</div>' +
      (open
        ? '<div class="mr-sub" style="margin-top:6px;white-space:pre-wrap">' + esc(trunc(String(p.summary || ''), 2000) || tr('summary 缺席——详情见对象工作台')) + '</div>' +
          (extras ? '<div class="kv" style="margin-top:6px">' + extras + '</div>' : '') +
          '<div style="margin-top:6px;display:flex;gap:6px;flex-wrap:wrap">' +
          (p && p.name ? chipGo(tr('对象工作台'), objRoute('Proposal', p.scope || 'global', p.name), 'mute', true) : '') +
          (propLink ? chipGo(tr('直达链接'), propLink, 'mute', true) : '') +
          '</div>' +
          '<div class="small faint" style="margin-top:4px">' + esc(tr('再点一次卡片收起。')) + '</div>'
        : '') +
      '</div>' +
      '<span class="ws-chip">' + chipGo(tr('去审批'), approveRoute, 'acc', true, tr('定位至审批页该提案')) + '</span>' +
      '</div>';
  }).join('');
  el.innerHTML = '<div class="view-inner">' +
    screenHead(tr('需求提案'), tr('提案面（点卡片展开 summary 全文与要素；支持 #/prop/:name 直达）——批准/驳回走审批与 Apply（LLM 无 Apply 权）'),
      focusName ? chipGo(tr('全部提案'), 'proposals', 'mute', false) : '') +
    (rows.length ? '<div class="card">' + cards + '</div>'
      : emptyBox(tr('暂无提案'), tr('快照 proposals 聚合为空。'))) +
    '</div>';
  el.querySelectorAll('[data-prop]').forEach((row) => {
    row.addEventListener('click', (ev) => {
      if (ev.target.closest('[data-go]')) return;
      const key = row.dataset.prop;
      if (propOpen.has(key)) propOpen.delete(key); else propOpen.add(key);
      proposalsScreen(el, ctx);
    });
  });
  if (focusName) {
    const hit = el.querySelector('[data-prop-focused="1"]');
    if (hit && hit.scrollIntoView) hit.scrollIntoView({ behavior: 'smooth', block: 'center' });
  }
}

/* ---------- V-11：系统详情补 tasks / drafts / runs 三区块 ---------- */
function systemExtraBlocks(ctx, scope) {
  const tasks = ctx.store.tasks().filter((t) => t && (String(t.scope || '') === scope || String((t.spec && t.spec.scopeRef && t.spec.scopeRef.name) || '') === scope));
  const drafts = ctx.store.drafts().filter((d) => d && String(d.scope || 'global') === scope);
  const runs = ctx.store.runs().filter((r) => {
    if (!r) return false;
    let hay = '';
    try { hay = JSON.stringify([r.scope, r.ownerScope, r.involved, r.route, r.workflow]); } catch { hay = ''; }
    return hay.includes(scope);
  });
  const sec = (title, bodyHtml, emptyText) =>
    '<h2 class="sec-title">' + esc(title) + '</h2><div class="card">' +
    (bodyHtml || '<div class="card-b"><div class="small faint">' + esc(emptyText) + '</div></div>') + '</div>';
  const taskHtml = tasks.slice(0, 8).map((t) => {
    const id = (t.scope || 'global') + '/' + (t.name || '');
    const ph = String((t.status && t.status.phase) || '');
    return '<div class="mini-row" data-go="task/' + encodeURIComponent(id) + '" style="cursor:pointer">' +
      '<div class="mr-main"><div><b>' + esc(String((t.spec && t.spec.title) || t.name || id)) + '</b> ' +
      (ph ? chip(ph, ph === 'running' || ph === 'verifying' ? 'acc' : 'mute') : '') + '</div>' +
      '<div class="mr-sub small mono">' + esc(id) + '</div></div></div>';
  }).join('');
  const draftHtml = drafts.slice(0, 8).map((d) => {
    const sc = d.scope || 'global';
    return '<div class="mini-row" data-go="' + esc(docRoute(d.kind, sc, d.name)) + '" style="cursor:pointer">' +
      '<div class="mr-main"><div><b>' + esc(String(d.name || '')) + '</b> ' + kindBadge(String(d.kind || '')) + ' ' + chip(tr('草稿'), 'warn') + '</div>' +
      '<div class="mr-sub small mono">' + esc(String(d.fieldManager || '—')) + '</div></div>' +
      '<span class="ws-chip">' + chipGo(tr('去审批'), 'approvals/' + encodeURIComponent(d.kind + '/' + sc + '/' + d.name), 'acc', true) + '</span></div>';
  }).join('');
  const runHtml = runs.slice(0, 8).map((r) => {
    const wf = String(r.workflow || r.name || '');
    const no = r.run != null ? r.run : (r.no != null ? r.no : '');
    const go = wf && no !== '' && no != null ? 'wf/' + encodeURIComponent(wf) + '/r/' + encodeURIComponent(String(no)) : (wf ? 'wf/' + encodeURIComponent(wf) : '');
    return '<div class="mini-row"' + (go ? ' data-go="' + esc(go) + '" style="cursor:pointer"' : '') + '>' +
      '<div class="mr-main"><div><b class="mono">' + esc(wf || '—') + '</b> ' + (no !== '' ? '<span class="small faint">#' + esc(String(no)) + '</span>' : '') + '</div>' +
      '<div class="mr-sub small mono">' + esc(trunc(String(r.involved || r.route || '—'), 56)) + '</div></div></div>';
  }).join('');
  return sec(tr('相关任务'), taskHtml, tr('暂无相关任务。')) +
    sec(tr('草稿'), draftHtml, tr('暂无草稿。')) +
    sec(tr('运行记录'), runHtml, tr('暂无运行记录。'));
}

/* ---------- 系统详情（#/s/:name——scope 视角的对象清单+摘要） ---------- */
function systemScreen(el, ctx) {
  const scope = String(ctx.params.name || '');
  const objs = ctx.store.objects().filter((o) => o.scope === scope);
  const evs = ctx.store.events().filter((e) => String(e.involved || '').includes(scope)).slice(0, 10);
  const byKind = {};
  objs.forEach((o) => { byKind[o.kind] = (byKind[o.kind] || 0) + 1; });
  el.innerHTML = '<div class="view-inner">' +
    screenHead(scope, 'scope=' + scope + tr(' 的对象、任务、草稿、运行与事件（快照聚合投影）'), chipGo(tr('系统树'), 'tree', 'mute', false)) +
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
      : emptyBox(tr('该 scope 暂无在册对象'), tr('该 scope 下暂无对象。'))) +
    (evs.length ? tr('<h2 class="sec-title">最近相关事件</h2><div class="card">') + evs.map((e) =>
      '<div class="mini-row" data-go="event/' + encodeURIComponent(String(e.id ?? '')) + '">' +
      '<div class="mr-main"><div>' + esc(String(e.type || '')) + '</div>' +
      '<div class="mr-sub small mono">' + esc(trunc(String(e.involved || '—'), 60)) + '</div></div></div>').join('') + '</div>' : '') +
    systemExtraBlocks(ctx, scope) +
    '</div>';
}

export function render(el, ctx) {
  const r = ctx.route;
  if (r === 'tree') treeScreen(el, ctx);
  else if (r === 'docs') docsScreen(el, ctx);
  else if (r === 'designdoc') designDocScreen(el, ctx);
  else if (r === 'connectors') connectorsScreen(el, ctx);
  else if (r === 'proposals' || r === 'prop') proposalsScreen(el, ctx);
  else if (r === 's') systemScreen(el, ctx);
  else el.innerHTML = '<div class="view-inner">' + emptyBox(tr('未知薄屏'), String(r)) + '</div>';
}

/* 业务控制台占位屏（2026-11-26 用户裁决：先不加东西——显式占位词形，不装内容）。
 * V-22：词形显式声明「业务层未部署（设计稿在册）」+ 设计文档深链。 */
export function renderBiz(el, ctx) {
  void ctx;
  el.innerHTML = '<div class="view-inner">' +
    '<div class="biz-hero">' +
    '<div class="biz-hero-mark">B</div>' +
    '<h2 class="sec-title">' + esc(tr('业务控制台')) + '</h2>' +
    '<div class="small faint">' + esc(tr('业务层未接入——后续业务服务登录后在此承载业务大屏与操作面。')) + '</div>' +
    '<div style="margin-top:10px">' + chipGo(tr('查看业务设计稿'), 'designdoc/' + encodeURIComponent('core_design') + '/' + encodeURIComponent('concepts/business-system.md'), 'mute', true) + '</div>' +
    '</div></div>';
}
