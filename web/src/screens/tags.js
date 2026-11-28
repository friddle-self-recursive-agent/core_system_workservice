/* tags.js —— 标签聚合屏（V-16，2026-11-26 概念可见性批）：#/tags 主面 + #/docs/tag/:tag
 * 子词形联动。三层真实数据源按对象键 kind|scope|name 严格去重（零假词表、零重复计数）：
 *   ① GET /v1/context/index 索引投影 tags 字段（权威面）；
 *   ② 对象 ref 携带的 tags（spec.metadata.tags / metadata.tags / tags）；
 *   ③ docCache 命中正文 frontmatter 解析（预取窗口，沿用诚实注脚口径）。
 * 分组=tagType 词表（scheme metadata.tagTypes 缺席=仅按 type:value 前缀或全落「未分组」恒殿后）。
 * 纯函数零 DOM——node 单测机械直测口径。 */

import { esc } from '../core/format.js';
import { chip, chipGo, emptyBox, screenHead, tagRoute } from '../render/chips.js';
import { tagVocabFromKinds } from './objects.js';
import { t as tr } from '../core/i18n.js';

/* ---------- 纯逻辑（node 单测直测） ---------- */

/** 正文 frontmatter tags 解析：兼容 inline `tags: [a, b]` 与 YAML 列表 `tags:\n  - a`。 */
export function frontmatterTags(content) {
  const s = String(content || '');
  const head = s.slice(0, 4000);
  if (!head.includes('tags:')) return [];
  const out = [];
  const flat = head.match(/tags:\s*\[([^\]]*)\]/);
  if (flat) {
    flat[1].split(',').forEach((x) => {
      const v = x.trim().replace(/^['"]|['"]$/g, '');
      if (v) out.push(v);
    });
    return out;
  }
  const block = head.match(/tags:\s*\n((?:[ \t]+-[ \t]+[^\n]+\n?)+)/);
  if (block) {
    block[1].split('\n').forEach((line) => {
      const m = /^[ \t]+-[ \t]+(.+?)\s*$/.exec(line);
      if (m) {
        const v = m[1].trim().replace(/^['"]|['"]$/g, '');
        if (v) out.push(v);
      }
    });
  }
  return out;
}

/** 对象 ref 携带的 tags 提取（兼容 raw.spec.metadata.tags / raw.metadata.tags / tags）。 */
export function tagsOfRef(o) {
  if (!o || typeof o !== 'object') return [];
  const raw = (o.raw && typeof o.raw === 'object') ? o.raw : o;
  const spec = (raw.spec && typeof raw.spec === 'object') ? raw.spec : {};
  const meta = (spec.metadata && typeof spec.metadata === 'object') ? spec.metadata
    : ((raw.metadata && typeof raw.metadata === 'object') ? raw.metadata : {});
  const list = meta.tags || spec.tags || raw.tags || o.tags;
  return Array.isArray(list) ? list.map((x) => String(x || '').trim()).filter(Boolean) : [];
}

function docKeyOf(kind, scope, name) {
  return String(kind || '') + '|' + String(scope || 'global') + '|' + String(name || '');
}

/** 三源真实 tag 聚合（严格按 kind|scope|name 对象键去重——同对象多源命中只计 1 次）。
 * 返回 [{tag, type, n, keys}] 按计数降序 + tag 字典序。 */
export function aggregateTags(indexDocs, objs, cached, vocab) {
  const V = vocab || {};
  const UN = tr('未分组');
  const m = new Map(); // tag -> Set<docKey>
  const add = (tag, key) => {
    const t = String(tag || '').trim();
    if (!t || !key) return;
    let set = m.get(t);
    if (!set) { set = new Set(); m.set(t, set); }
    set.add(key);
  };
  // ① 索引投影（/v1/context/index——权威面）
  (Array.isArray(indexDocs) ? indexDocs : []).forEach((d) => {
    if (!d || !d.name) return;
    const key = docKeyOf(d.kind, d.scope, d.name);
    const tags = Array.isArray(d.tags) ? d.tags : [];
    tags.forEach((t) => add(t, key));
  });
  // ②③ 对象 ref tags ∪ docCache frontmatter
  (Array.isArray(objs) ? objs : []).forEach((o) => {
    if (!o || !o.kind || !o.name) return;
    const key = docKeyOf(o.kind, o.scope, o.name);
    tagsOfRef(o).forEach((t) => add(t, key));
    const c = typeof cached === 'function' ? cached(o.kind, o.scope || 'global', o.name, o.rv) : null;
    if (c && typeof c.content === 'string') {
      frontmatterTags(c.content).forEach((t) => add(t, key));
    }
  });
  const out = Array.from(m.entries()).map(([tag, set]) => {
    const colon = tag.indexOf(':');
    const prefix = colon > 0 ? tag.slice(0, colon) : '';
    const type = V[tag] || (prefix ? (V[prefix] || prefix) : '') || UN;
    return { tag, type, n: set.size, keys: Array.from(set).sort() };
  });
  out.sort((a, b) => (b.n - a.n) || (a.tag < b.tag ? -1 : 1));
  return out;
}

/** 分组段（aggregateTags 行 → tagType 分组；「未分组」恒殿后不冒充分类）。 */
export function groupByType(rows) {
  const UN = tr('未分组');
  const m = new Map();
  (rows || []).forEach((r) => {
    const key = (r && r.type) ? r.type : UN;
    if (!m.has(key)) m.set(key, []);
    m.get(key).push(r);
  });
  return Array.from(m.keys())
    .sort((a, b) => (a === UN ? 1 : b === UN ? -1 : (a < b ? -1 : 1)))
    .map((type) => ({ type, rows: m.get(type) }));
}

/* 索引投影轻量缓存（跨 #/tags 与 #/docs/tag/:tag 复用，避免重复拉取；失败不缓存）。 */
let INDEX_CACHE = { docs: null, at: 0 };
const INDEX_TTL_MS = 30000;

export async function loadContextIndexDocs(api) {
  if (INDEX_CACHE.docs && Date.now() - INDEX_CACHE.at < INDEX_TTL_MS) {
    return { ok: true, docs: INDEX_CACHE.docs };
  }
  const fn = api && (typeof api.contextIndex === 'function'
    ? () => api.contextIndex()
    : () => api.request('v1/context/index'));
  if (!fn) return { ok: false, docs: [], error: 'api 缺席' };
  const r = await fn();
  if (!r || !r.ok) return { ok: false, docs: [], error: String((r && (r.error || r.status)) || '不可读') };
  const docs = (r.data && Array.isArray(r.data.docs)) ? r.data.docs : [];
  INDEX_CACHE = { docs, at: Date.now() };
  return { ok: true, docs };
}

export function hasCachedIndex() {
  return !!(INDEX_CACHE.docs && Date.now() - INDEX_CACHE.at < INDEX_TTL_MS);
}

/** 取某对象在缓存的索引投影中的 tags（docs/tag/:tag 过滤用）。 */
export function cachedIndexTagsOf(kind, scope, name) {
  if (!INDEX_CACHE.docs) return [];
  const k = docKeyOf(kind, scope, name);
  const hit = INDEX_CACHE.docs.find((d) => d && docKeyOf(d.kind, d.scope, d.name) === k);
  return hit && Array.isArray(hit.tags) ? hit.tags : [];
}

/* ---------- 渲染 ---------- */

/** 标签聚合屏（#/tags）。 */
export function render(el, ctx) {
  const store = ctx.store;
  let vocab = tagVocabFromKinds(store.kinds());
  el.innerHTML = '<div class="view-inner">' +
    screenHead(tr('标签'), tr('tagType 分组聚合（三源严格去重：context/index 索引投影 ∪ 对象 tags ∪ 缓存 frontmatter；零假词表）——点 tag 看命中文档'),
      chipGo(tr('文档库'), 'docs', 'mute', false)) +
    '<div data-tags-body><div class="card"><div class="card-b"><div class="small faint">' + esc(tr('聚合现读中…（GET /v1/context/index + 缓存正文 frontmatter）')) + '</div></div></div></div>' +
    '</div>';

  const paint = (rows, sourceNote) => {
    const body = el.querySelector('[data-tags-body]');
    if (!body) return;
    if (!rows.length) {
      body.innerHTML = '<div class="card"><div class="card-b">' +
        emptyBox(tr('暂无标签'), tr('三源聚合零命中（索引投影与预取窗口正文均未声明 tags）。')) +
        '</div></div>';
      return;
    }
    const groups = groupByType(rows);
    body.innerHTML = '<div class="small faint" style="margin-bottom:8px">' + esc(sourceNote) + '</div>' +
      groups.map((g) =>
        '<section class="card" style="margin-bottom:10px"><div class="card-h"><span class="h-title">' + esc(g.type) + '</span>' +
        '<span class="h-right">' + chip(g.rows.length + tr(' 个 tag'), 'mute') + '</span></div><div class="card-b" style="display:flex;gap:6px;flex-wrap:wrap">' +
        g.rows.map((r) => chipGo('#' + r.tag + ' · ' + r.n, tagRoute(r.tag), r.n >= 2 ? 'acc' : 'mute', true, tr('看同 tag 文档'))).join('') +
        '</div></section>').join('');
  };

  const cachedFn = (k, s, n, rv) => store.cachedDoc(k, s, n, rv);
  paint(
    aggregateTags(INDEX_CACHE.docs, store.objects(), cachedFn, vocab),
    INDEX_CACHE.docs
      ? tr('数据源：GET /v1/context/index（') + INDEX_CACHE.docs.length + tr(' 篇）∪ 对象 tags ∪ 缓存 frontmatter——去重计数。')
      : tr('数据源：对象 tags ∪ 缓存 frontmatter（预取窗口 24 篇）——索引投影现读中…'),
  );

  const stamp = typeof location !== 'undefined' ? location.hash : '';
  const stale = store.staleDocs ? store.staleDocs(24) : [];
  const hydrateP = (stale.length && ctx.api && typeof ctx.api.object === 'function')
    ? (async () => {
      let i = 0;
      const w = async () => {
        while (i < stale.length) {
          const r = stale[i++];
          const res = await ctx.api.object(r.kind, r.scope || 'global', r.name);
          if (res && res.ok) {
            const d = res.data && typeof res.data === 'object' ? res.data : {};
            const c = d.content != null ? String(d.content) : '';
            if (c) store.cacheDoc(r.kind, r.scope || 'global', r.name, r.rv, c);
          }
        }
      };
      await Promise.all(Array.from({ length: Math.min(4, stale.length) }, w));
    })()
    : Promise.resolve();

  /* 词表现读注水（2026-11-28 tags 批：快照 kinds 只存名单字符串——spec.metadata.tagTypes 词表
   * 声明随 kind 行，须 /v1/scheme/kinds 全行现读；objects 屏 scheme 弹窗同款兜底）。
   * 词表缺席=前缀域兜底分组（type:value 词形不落未分组），存量裸词由词表精确映射兜。 */
  const kindsP = (ctx.api && typeof ctx.api.kinds === 'function')
    ? ctx.api.kinds().catch(() => null)
    : Promise.resolve(null);

  Promise.all([loadContextIndexDocs(ctx.api), hydrateP, kindsP]).then(([res, , kr]) => {
    if (!el.isConnected || (typeof location !== 'undefined' ? location.hash : '') !== stamp) return;
    if (kr && kr.ok && kr.data && Array.isArray(kr.data.kinds)) {
      vocab = tagVocabFromKinds(kr.data.kinds);
    }
    if (!res.ok) {
      paint(
        aggregateTags(null, store.objects(), cachedFn, vocab),
        tr('索引投影暂不可读（') + res.error + tr('）——当前仅展示对象 tags ∪ 预取窗口 frontmatter（诚实降级）。'),
      );
      return;
    }
    paint(
      aggregateTags(res.docs, store.objects(), cachedFn, vocab),
      tr('数据源：GET /v1/context/index 索引投影（') + res.docs.length + tr(' 篇在册）∪ 对象 tags ∪ 缓存 frontmatter（预取窗口）——三源对象键去重。'),
    );
  }).catch(() => { /* 网络异常保留②③投影 */ });
}
