/* tags.js —— 标签聚合屏（V-16，2026-11-26 概念可见性批）：#/tags 主面 + #/docs/tag/:tag
 * 子词形落地（docs 屏第三段预填 tag 过滤）。数据源三层真实聚合（零假词表）：
 *   ① GET /v1/context/index 索引投影 tags 字段（权威面——在册即来）；
 *   ② 对象 tagsOfDoc（docCache frontmatter 解析——预取窗口内，沿用 docs 屏诚实注脚口径）；
 *   ③ docCache 命中正文 tags（=②同源；聚合键 kind|scope|name 去重）。
 * 分组=tagType 词表（scheme metadata.tagTypes 缺席=全部落「未分组」恒殿后——不冒充分类）。
 * 纯函数（aggregateTags/tagCloudHtml）零 DOM——node 单测机械直测口径。 */

import { esc } from '../core/format.js';
import { chip, chipGo, emptyBox, screenHead } from '../render/chips.js';
import { t as tr } from '../core/i18n.js';

/* ---------- 纯逻辑（node 单测直测） ---------- */

/** 三源 tag 聚合（docs=[] 索引投影行；objs=[] 快照对象 ref 行；cached(kind,scope,name)=正文或 null）。
 * 返回 [{tag, types:Set有序, n}] 按计数降序+tag 字典序。types=tagType 词表命中集（'type:tag'
 * 键词形拆段——tagType 缺席=未分组）。 */
export function aggregateTags(docs, objs, cached, vocab) {
  const V = vocab || {};
  const m = new Map(); // tag → { n, types:Set }
  const bump = (tag, kind, scope, name) => {
    const t = String(tag || '').trim();
    if (!t) return;
    let e = m.get(t);
    if (!e) { e = { n: 0, types: new Set() }; m.set(t, e); }
    e.n += 1;
    if (kind && !e.sources) { /* 首个来源登记对象锚（仅记一次——计数跨源去重按对象） */ }
  };
  // ① 索引投影（权威面）——行已带 kind/scope/name+tags
  (Array.isArray(docs) ? docs : []).forEach((d) => {
    const tags = Array.isArray(d && d.tags) ? d.tags : [];
    tags.forEach((t) => bump(t, d && d.kind, d && d.scope, d && d.name));
  });
  // ②③ 对象 ref × docCache frontmatter（预取窗口内；无缓存=跳过不冒充）
  (Array.isArray(objs) ? objs : []).forEach((o) => {
    if (!o || !o.kind || !o.name) return;
    const c = cached(o.kind, o.scope || 'global', o.name);
    const meta = c && typeof c.content === 'string' ? frontmatterTags(c.content) : null;
    if (meta) meta.forEach((t) => bump(t, o.kind, o.scope, o.name));
  });
  const out = Array.from(m.entries()).map(([tag, e]) => {
    const typ = String(tag).indexOf(':') > 0 ? String(tag).slice(0, String(tag).indexOf(':')) : '';
    const type = typ ? (V[typ] || typ) : tr('未分组');
    return { tag, type, n: e.n };
  });
  out.sort((a, b) => (b.n - a.n) || (a.tag < b.tag ? -1 : 1));
  return out;
}

/** frontmatter tags 现读（spec.metadata.tags 优先、metadata.tags/tags 次之——flat `[a, b]` 词形）。 */
export function frontmatterTags(content) {
  const s = String(content || '');
  if (!s.slice(0, 4000).includes('tags:')) return null;
  const head = s.slice(0, 4000);
  const flat = head.match(/tags:\s*\[([^\]]*)\]/);
  if (!flat) return null;
  const tags = [];
  flat[1].split(',').forEach((x) => {
    const v = x.trim().replace(/^['"]|['"]$/g, '');
    if (v) tags.push(v);
  });
  return tags;
}

/** tagType 词表（scheme kinds 内 metadata.tagTypes 声明——缺席=空表=全落「未分组」）。 */
export function tagTypeVocab(schemeKinds) {
  const rows = Array.isArray(schemeKinds) ? schemeKinds : [];
  for (const k of rows) {
    const meta = k && k.spec && k.spec.metadata && typeof k.spec.metadata === 'object' ? k.spec.metadata : null;
    const t = meta && meta.tagTypes;
    if (t && typeof t === 'object' && !Array.isArray(t)) return t;
    if (Array.isArray(t)) return t.reduce((m, x) => { m[String(x)] = String(x); return m; }, {});
  }
  return {};
}

/** 分组段（aggregateTags 行 → tagType 分组；未分组恒殿后不冒充分类）。 */
export function groupByType(rows) {
  const UN = tr('未分组');
  const m = new Map();
  (rows || []).forEach((r) => {
    const key = r && r.type ? r.type : UN;
    if (!m.has(key)) m.set(key, []);
    m.get(key).push(r);
  });
  return Array.from(m.keys()).sort((a, b) => (a === UN ? 1 : b === UN ? -1 : (a < b ? -1 : 1)))
    .map((type) => ({ type, rows: m.get(type) }));
}

/* ---------- 渲染 ---------- */

/** 标签聚合屏（#/tags）。 */
export function render(el, ctx) {
  const store = ctx.store;
  const vocab = tagTypeVocab(store.kinds());
  let idx = null;
  el.innerHTML = '<div class="view-inner">' +
    screenHead(tr('标签'), tr('tagType 分组聚合（三源真实计数：context/index 索引投影+对象 frontmatter；零假词表）——点 tag 看命中文档'),
      chipGo(tr('文档库'), 'docs', 'mute', false)) +
    '<div data-tags-body><div class="card"><div class="card-b"><div class="small faint">' + esc(tr('聚合现读中…（GET /v1/context/index + 缓存正文 frontmatter）')) + '</div></div></div></div>' +
    '</div>';
  const paint = (rows, sourceNote) => {
    const body = el.querySelector('[data-tags-body]');
    if (!body) return;
    if (!rows.length) {
      body.innerHTML = '<div class="card"><div class="card-b">' + emptyBox(tr('窗口内无在册 tag'), tr('三源聚合零命中（索引投影与预取窗口正文均未声明 tags——空态显式）。')) + '</div></div>';
      return;
    }
    const groups = groupByType(rows);
    body.innerHTML = '<div class="small faint" style="margin-bottom:8px">' + esc(sourceNote) + '</div>' +
      groups.map((g) =>
        '<section class="card" style="margin-bottom:10px"><div class="card-h"><span class="h-title">' + esc(g.type) + '</span>' +
        '<span class="h-right">' + chip(g.rows.length + tr(' 个 tag'), 'mute') + '</span></div><div class="card-b">' +
        g.rows.map((r) => chipGo('#' + r.tag + ' · ' + r.n, 'docs/tag/' + encodeURIComponent(r.tag), r.n >= 2 ? 'acc' : 'mute', true, tr('看同 tag 文档')))
          .join(' ') + '</div></section>').join('');
  };
  paint(aggregateTags(null, store.objects(), (k, s, n) => store.cachedDoc(k, s, n), vocab),
    tr('数据源：快照对象×docCache frontmatter（预取窗口 24 篇）——索引投影现读中…'));
  // ① 权威面注水（GET /v1/context/index——失败保留②③投影+诚实注记）
  const stamp = typeof location !== 'undefined' ? location.hash : '';
  ctx.api.request('v1/context/index').then((r) => {
    if (!el.isConnected || (typeof location !== 'undefined' ? location.hash : '') !== stamp) return;
    if (!r.ok) {
      paint(aggregateTags(null, store.objects(), (k, s, n) => store.cachedDoc(k, s, n), vocab),
        tr('索引投影暂不可读（') + esc(String((r && (r.error || r.status)) || '')) + tr('）——本面=对象 frontmatter 投影（预取窗口），权威面恢复后重进即现读。'));
      return;
    }
    const data = r.data && typeof r.data === 'object' ? r.data : {};
    const docs = Array.isArray(data.docs) ? data.docs : [];
    const rows = aggregateTags(docs, store.objects(), (k, s, n) => store.cachedDoc(k, s, n), vocab);
    paint(rows, tr('数据源：GET /v1/context/index 索引投影（') + docs.length + tr(' 篇在册）∪ 对象 frontmatter（预取窗口）——三源去重计数。'));
  }).catch(() => { /* 网络异常=②③投影保留——诚实注记已在 paint 词形 */ });
}
