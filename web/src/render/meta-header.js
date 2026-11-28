/* meta-header.js —— 文档元数据公用 header 组件（2026-11-28 用户裁决：所有文档都有
 * metadatainfo 侧车——tag/创建时间/提交 agent/提交人/相应 ref 等原数据要上屏，公用一个好 header）。
 * 数据源：①frontmatter 现解析（docType/scope/title/tags/importance/ownerRefs——R-DOC-6 侧车
 * 导出词形，git 设计文档与 DB 对象同源通用）；②/v1/doc-history（创建时间=最早 rev 行 at、
 * 最近提交=最新 rev 行 fieldManager——提交人·agent 词形）；git 文档无 author 词形=HEAD sha
 * 诚实不冒充。纯函数 node 直测；一切动态文本 esc——零原文直插；缺席字段整行不渲染（零编造）。 */
import { esc, trunc, relAge } from '../core/format.js';
import { chip, chipGo, kindBadge, tagRoute, objRoute } from './chips.js';
import { splitFrontmatter, loadYaml } from './md.js';
import { t as tr } from '../core/i18n.js';

/** frontmatter → 元数据（纯函数；tags/importance/ownerRefs 三词形兼容——对象 {tag,tag_type} 对、
 * 裸字符串数组、根/metadata 区双落位）。 */
export function frontMetaOf(content, kindHint, scopeHint, nameHint) {
  const { fm } = splitFrontmatter(String(content || ''));
  const tree = fm ? loadYaml(fm) : null;
  const t = (tree && typeof tree === 'object') ? tree : {};
  const meta = (t.metadata && typeof t.metadata === 'object') ? t.metadata : {};
  const spec = (t.spec && typeof t.spec === 'object') ? t.spec : {};
  const pick = (...cands) => { for (const c of cands) { if (c != null && c !== '') return c; } return null; };
  const rawTags = pick(meta.tags, t.tags, spec.tags) || [];
  const tags = (Array.isArray(rawTags) ? rawTags : []).map((x) => {
    if (x && typeof x === 'object' && x.tag) return { tag: String(x.tag), type: String(x.tag_type || x.tagType || '') };
    const s = String(x || '').trim();
    if (!s) return null;
    const i = s.indexOf(':');
    return i > 0 ? { tag: s.slice(i + 1).trim(), type: s.slice(0, i).trim() } : { tag: s, type: '' };
  }).filter(Boolean);
  const rawRefs = pick(t.ownerRefs, meta.ownerRefs, spec.ownerRefs) || [];
  const ownerRefs = (Array.isArray(rawRefs) ? rawRefs : []).map((r) => {
    if (r && typeof r === 'object' && r.kind && r.name) return { kind: String(r.kind), name: String(r.name), scope: String(r.scope || 'global') };
    const s = String(r || '').trim();
    const seg = s.split('/').filter(Boolean);
    return seg.length >= 2 ? { kind: seg[0], name: seg.slice(1).join('/'), scope: 'global' } : null;
  }).filter(Boolean);
  const impRaw = pick(meta.importance, t.importance, spec.importance);
  return {
    name: String(pick(t.name, meta.name, nameHint) || nameHint || ''),
    kind: String(pick(t.kind, kindHint) || kindHint || ''),
    docType: String(pick(t.docType, meta.docType, spec.docType) || ''),
    scope: String(pick(t.scope, meta.scope, scopeHint) || scopeHint || 'global'),
    title: String(pick(t.title, meta.title) || ''),
    tags,
    importance: impRaw != null && Number.isFinite(Number(impRaw)) ? Number(impRaw) : null,
    ownerRefs,
  };
}

/** 元数据 header HTML（纯函数）。history={createdAt,lastAt,lastBy,head,rv} 缺席成员整项不渲染。 */
export function metaHeaderHtml(meta, history) {
  const m = meta || {};
  const h = history || {};
  const parts = [];
  // 行 1：身份（名+kind/docType 徽+scope+版本锚）
  const ident = [];
  if (m.title && m.title !== m.name) ident.push('<b>' + esc(trunc(m.title, 60)) + '</b>');
  ident.push('<b class="mono">' + esc(trunc(String(m.name || ''), 60)) + '</b>');
  if (m.kind) ident.push(kindBadge(m.kind));
  if (m.docType && m.docType !== m.kind) ident.push(chip('docType: ' + m.docType, 'mute'));
  if (m.scope) ident.push(chip('scope: ' + m.scope, 'mute'));
  if (h.rv != null) ident.push(chip('r' + String(h.rv), 'mute'));
  if (h.head) ident.push(chip((h.headSha ? 'HEAD ' + String(h.headSha).slice(0, 8) : 'HEAD'), 'mute'));
  parts.push('<div class="mh-ident">' + ident.join(' ') + '</div>');
  // 行 2：账（创建时间 · 最近提交 fieldManager=提交人/agent 词形——零编造缺席不渲染）
  const ledger = [];
  if (h.createdAt) ledger.push('<span>' + esc(tr('创建 ')) + '<span class="mono" title="' + esc(String(h.createdAt)) + '">' + esc(String(h.createdAt).replace('T', ' ').slice(0, 16)) + (relAge(h.createdAt) ? ' · ' + esc(relAge(h.createdAt)) : '') + '</span></span>');
  if (h.lastAt || h.lastBy) {
    const seg = esc(tr('最近提交 ')) +
      (h.lastBy ? '<span class="mono">' + esc(trunc(String(h.lastBy), 44)) + '</span>' : '') +
      (h.lastAt ? '<span class="mono" title="' + esc(String(h.lastAt)) + '">' + (relAge(h.lastAt) ? ' · ' + esc(relAge(h.lastAt)) : '') + '</span>' : '');
    ledger.push('<span>' + seg + '</span>');
  }
  if (ledger.length) parts.push('<div class="mh-ledger small faint">' + ledger.join(' <span class="mh-dot">·</span> ') + '</div>');
  // 行 3：tags + importance（tag 可点跳聚合页）
  if ((m.tags && m.tags.length) || m.importance != null) {
    const tagHtml = (m.tags || []).map((x) =>
      chipGo('#' + x.tag, tagRoute(x.tag), 'mute', true, x.type ? 'tagType: ' + x.type : tr('看同 tag 文档'))
    ).join(' ');
    const imp = m.importance != null ? chip(tr('重要度 ') + m.importance, Number(m.importance) >= 80 ? 'warn' : 'mute') : '';
    parts.push('<div class="mh-tags">' + [imp, tagHtml].filter(Boolean).join(' ') + '</div>');
  }
  // 行 4：ownerRefs（归属引用——可点对象工作台）
  if (m.ownerRefs && m.ownerRefs.length) {
    const refs = m.ownerRefs.slice(0, 6).map((r) =>
      chipGo(r.kind + '/' + r.name, objRoute(r.kind, r.scope || 'global', r.name), 'mute', true, tr('归属对象'))
    ).join(' ');
    parts.push('<div class="mh-refs small">' + '<span class="faint">' + esc(tr('归属：')) + '</span> ' + refs + '</div>');
  }
  return '<div class="meta-header" data-meta-header>' + parts.join('') + '</div>';
}

/** paintMetaHeader DOM 助手：即paint frontmatter 元数据 → doc-history 异步补账（创建/最近提交）。
 * 槽缺席（无 DOM）安全 no-op；对象无 frontmatter（纯 YAML 对象）=只画身份+账两行。 */
export function paintMetaHeader(slot, ctx, ref, content, headInfo) {
  if (!slot) return;
  const meta = frontMetaOf(content, ref && ref.kind, ref && ref.scope, ref && ref.name);
  const base = (headInfo && headInfo.head) ? { head: true, headSha: headInfo.headSha } : { rv: ref ? ref.rv : null };
  slot.innerHTML = metaHeaderHtml(meta, base);
  slot.querySelectorAll('[data-go]').forEach((n) => {
    n.addEventListener('click', (ev) => {
      ev.stopPropagation();
      const c = (typeof window !== 'undefined' && window.__CC_CONSOLE_CTX__) || null;
      if (c && c.go) c.go(n.getAttribute('data-go'));
    });
  });
  if (!ctx || !ctx.api || typeof ctx.api.docHistory !== 'function' || !ref || !ref.kind || !ref.name) return;
  const stamp = (typeof location !== 'undefined') ? location.hash : '';
  ctx.api.docHistory(ref.kind, ref.scope, ref.name).then((r) => {
    if (!slot.isConnected || (typeof location !== 'undefined' ? location.hash : '') !== stamp) return;
    const rows = (r && r.ok && Array.isArray(r.data && (r.data.rows || r.data.history || r.data.items))) ? (r.data.rows || r.data.history || r.data.items) : (r && r.ok && Array.isArray(r.data) ? r.data : []);
    if (!rows.length) return; // 无账=保持 frontmatter 词形（新建对象首次落库前）
    const first = rows[0];
    const last = rows[rows.length - 1];
    slot.innerHTML = metaHeaderHtml(meta, Object.assign({}, base, {
      createdAt: String((first && (first.at || first.appliedAt)) || ''),
      lastAt: String((last && (last.at || last.appliedAt)) || ''),
      lastBy: String((last && last.fieldManager) || ''),
    }));
    slot.querySelectorAll('[data-go]').forEach((n) => {
      n.addEventListener('click', (ev) => {
        ev.stopPropagation();
        const c = (typeof window !== 'undefined' && window.__CC_CONSOLE_CTX__) || null;
        if (c && c.go) c.go(n.getAttribute('data-go'));
      });
    });
  }).catch(() => { /* 账缺席=保持现词形 */ });
}
