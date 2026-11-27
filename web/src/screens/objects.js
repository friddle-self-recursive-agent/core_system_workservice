/* objects.js —— 通用对象工作台（六片固定：标签 / 适用规则 / 上下文 / 操作 / 内容 / scheme）。
 * 语义移植旧 console-obj.js（代码全部重写为 ESM+库）。缺片显式「无」不装假（诚实空态词形）。
 * 单一真相：对象枚举/kinds=快照（store）；对象正文=docCache(rv 缓存)+api.object 现读。
 * 规则命中=确定性纯函数 rulesFor()（tag/kind/scope/正文引用四路——命中来源显式标注；
 * 零命中=「无适用规则」诚实空态；规则索引缺席=降级词形不冒充）。
 * 安全红线：一切动态文本 esc/或经 md 渲染管线；零原文直插。 */

import { esc, trunc } from '../core/format.js';
import { chip, chipGo, kv, emptyBox, screenHead, kindBadge, objRoute, docRoute, repoLinkChip, tagRoute } from '../render/chips.js';
import { loadYaml, yamlToCards, renderMarkdown, rawDisclosure } from '../render/md.js';
import { t as tr } from '../core/i18n.js';

/** tagType 词表提取（V-16：读 scheme kinds 内 metadata.tagTypes 声明；缺席=空表全落未分组）。
 * F-08 ③：词表读 /v1/scheme/kinds 行（store.kinds 快照聚合缺席=空表降级——不再「暂不可达」）。 */
export function tagVocabFromKinds(kinds) {
  const rows = Array.isArray(kinds) ? kinds : [];
  for (const k of rows) {
    const meta = k && k.spec && k.spec.metadata && typeof k.spec.metadata === 'object' ? k.spec.metadata : null;
    const t = meta && meta.tagTypes;
    if (t && typeof t === 'object' && !Array.isArray(t)) return t;
    if (Array.isArray(t)) return t.reduce((acc, x) => { acc[String(x)] = String(x); return acc; }, {});
  }
  return {};
}

/* ---------- tagType 分组（词表缺席=「未分组」；未分组恒殿后，tags 字典序） ----------
 * 支持两式词表：vocab[tag] 精确映射 ∪ type:value 前缀命中 vocab[type] 或显式前缀。 */
export function tagGroups(tags, vocab) {
  const V = vocab || {};
  const UN = tr('未分组');
  const m = {};
  (tags || []).forEach((t) => {
    const s = String(t || '');
    const colon = s.indexOf(':');
    const prefix = colon > 0 ? s.slice(0, colon) : '';
    const type = V[s] || (prefix ? (V[prefix] || prefix) : '') || UN;
    (m[type] = m[type] || []).push(s);
  });
  return Object.keys(m).sort((a, b) => {
    if (a === UN) return 1;
    if (b === UN) return -1;
    return a < b ? -1 : 1;
  }).map((type) => ({ type, tags: m[type].sort() }));
}

/* ---------- 规则命中（确定性纯函数——四路显式来源；≥1 路命中即列；上限 8） ---------- */
export function rulesFor(obj, rules) {
  if (!rules) return { hits: [], degraded: true, reason: tr('规则文档族索引暂不可达——命中判定暂缺（不冒充无适用）。'), scanned: 0 };
  const o = obj || {};
  const text = String(o.content || '');
  const low = text.toLowerCase();
  const nameLow = String(o.name || '').toLowerCase();
  const scopeLow = String(o.scope || '').toLowerCase();
  const kindLow = String(o.kind || '').toLowerCase();
  const objTags = o.tags || [];
  const out = [];
  (rules || []).forEach((r) => {
    if (!r || !r.path) return;
    const base = String(r.path).replace(/^.*\//, '').replace(/\.md$/, '');
    const rlow = String(r.content || '').toLowerCase();
    const rtags = r.tags || [];
    const why = [];
    const tagHit = objTags.filter((t) => rtags.includes(t));
    if (tagHit.length) why.push(tr('标签「') + tagHit.join('、') + tr('」同组'));
    if (kindLow && rlow.includes(kindLow)) why.push(tr('规则原文提及 ') + o.kind);
    if (scopeLow && scopeLow !== 'global' && scopeLow !== 'builtin' && rlow.includes(scopeLow)) why.push(tr('规则原文提及 ') + o.scope);
    const nameStem = nameLow.replace(/[-_].*$/, '');
    if (nameStem && nameStem.length >= 3 && rlow.includes(nameStem)) why.push(tr('规则原文提及 ') + nameStem);
    if (base && low && low.includes(base.toLowerCase())) why.push(tr('对象正文引用 ') + base);
    if (!why.length) return;
    out.push({ ref: (r.repo || 'core_design') + '/' + r.path + '#' + base, name: base, why });
  });
  out.sort((a, b) => b.why.length - a.why.length || (a.name < b.name ? -1 : 1));
  return { hits: out.slice(0, 8), degraded: false, reason: '', scanned: (rules || []).length };
}

/** 必填段完成度（requiredSections 与正文标题对照——诚实勾缺）。 */
export function sectionCompletion(content, requiredSections) {
  const secs = requiredSections || [];
  if (!secs.length) return [];
  const heads = {};
  String(content || '').split('\n').forEach((line) => {
    const m = /^\s*#{1,6}\s+(.+?)\s*$/.exec(line);
    if (m) heads[m[1].toLowerCase()] = true;
  });
  return secs.map((s) => {
    const key = String(s).toLowerCase();
    return { name: String(s), done: !!heads[key] || String(content || '').toLowerCase().includes(key) };
  });
}

/** 正文 yaml/frontmatter 中的 tags 现读（支持 md `---` frontmatter + 纯 yaml；支持字符串数组与 `{tag,tag_type}` 对）。
 * F-08 ④：元数据路径=顶层 metadata.tags 优先（scheme 2026-11-26 统一口径），spec.metadata.tags
 * 与 spec.tags 兜底——不再只读 spec.metadata 致「未声明 spec.metadata.tags」全屏假空态。 */
export function tagsOfDoc(content) {
  const s = String(content || '');
  let yamlSrc = s;
  if (s.startsWith('---')) {
    const end = s.indexOf('\n---', 3);
    if (end > 3) yamlSrc = s.slice(3, end);
  }
  const t = loadYaml(yamlSrc);
  if (t && typeof t === 'object') {
    const spec = t.spec && typeof t.spec === 'object' ? t.spec : {};
    const topMeta = t.metadata && typeof t.metadata === 'object' ? t.metadata : {};
    const specMeta = spec.metadata && typeof spec.metadata === 'object' ? spec.metadata : {};
    const raw = topMeta.tags || specMeta.tags || spec.tags || t.tags;
    if (Array.isArray(raw)) {
      return raw.map((x) => {
        if (x && typeof x === 'object' && x.tag) {
          const tt = x.tag_type || x.tagType || '';
          return tt ? String(tt) + ':' + String(x.tag) : String(x.tag);
        }
        return String(x || '').trim();
      }).filter(Boolean);
    }
  }
  // 兜底：轻量正则解析 frontmatter inline `tags: [...]`
  const head = s.slice(0, 4000);
  const flat = head.match(/tags:\s*\[([^\]]*)\]/);
  if (flat) return flat[1].split(',').map((x) => x.trim().replace(/^['"]|['"]$/g, '')).filter(Boolean);
  return [];
}

/** markdown 判据（语义移植旧 console isMarkdownDoc）。
 * F-08 ②：Proposal/Task 声明载体=markdown-frontmatter（scheme 2026-11-26 显式）——
 * `---` frontmatter + Markdown 正文双形态进 md 渲染管线（不再整篇当 yaml 致「空正文」）。 */
export function isMarkdownDoc(kind, name) {
  if (kind === 'Goal' || kind === 'Rule' || kind === 'Runbook' || kind === 'DriftReport' ||
      kind === 'Proposal' || kind === 'Task' || kind === 'DesignDoc' || kind === 'DocReview' || kind === 'BugDoc') return true;
  return /(^|[-_])(README|design|report|notes)([-_.]|$)/i.test(String(name || ''));
}

/* ---------- 六片渲染 ---------- */
function tagPanel(groups) {
  if (!groups || !groups.length) {
    return tr('<div class="ob-empty">暂无标签——对象正文未声明 spec.metadata.tags（声明经 Apply 落库后此处现读）。</div>');
  }
  return groups.map((g) =>
    '<div class="ob-tagrow">' + chip(g.type, 'mute') + ' ' +
    g.tags.map((t) => chipGo('#' + t, tagRoute(t), 'mute', true, tr('看同 tag 文档'))).join(' ') + '</div>'
  ).join('');
}

function rulePanel(res) {
  if (res.degraded) return '<div class="ob-empty">' + esc(res.reason) + '</div>';
  if (!res.hits.length) {
    // F-08 ①：无规则=诚实终态（不再「判定中…」永卡——hydrateRules null 缺席也落到本函数外的终态词）。
    return tr('<div class="ob-empty">无适用规则——按 kind/scope/tag/正文引用四路检查零命中（检查范围=') + res.scanned + tr(' 份规则文档）。</div>');
  }
  return res.hits.map((h, i) =>
    '<details class="ob-rule"><summary>' +
    '<span class="lk mono" data-go="' + esc('designdoc/core_design/' + encodeURIComponent(String(h.ref.replace(/^[^/]+\/([^#]+).*/, '$1') + '.md'))) + '">' + esc(h.ref) + '</span></summary>' +
    '<div class="ob-rule-why small">' + h.why.map((w) => chip(w, 'mute')).join(' ') + '</div></details>'
  ).join('');
}

function ctxPanel(rv) {
  return '<div class="ob-ctx">' + chip(tr('已入快照'), 'ok') +
    ' <span class="small">resourceVersion r' + esc(String(rv || 0)) + tr('（rv 未变不重读正文——降级轮询路径同样生效）</span>') +
    ' ' + chipGo(tr('在文档库中打开'), 'docs', 'mute', true) + '</div>';
}

function opsPanel(ref) {
  // 代码 ref 直达仓库（2026-11-26 用户裁决——内部仓与外部仓都可 [ref] 打开代码仓库地址）。
  const repoChips = [ref.scope, ref.name]
    .map((x) => repoLinkChip(String(x || ''), String(x || '')))
    .filter(Boolean).join(' ');
  return '<div>' +
    '<button class="btn btn-secondary sm" data-go="' + esc(docRoute(ref.kind, ref.scope, ref.name)) + tr('">编辑器</button> ') +
    '<button class="btn btn-ghost sm" data-go="' + esc('s/' + encodeURIComponent(ref.scope)) + tr('">所属 scope</button> ') +
    (repoChips ? '<div class="ob-ctx" style="margin-top:8px">' + repoChips + '</div>' : '') +
    '</div>';
}

function schemePanel(kindReg, content) {
  if (!kindReg) return tr('<div class="ob-empty">scheme 暂不可达（kinds 聚合缺席）。</div>');
  const spec = kindReg.spec || {};
  const keys = Object.keys(spec);
  const cards = keys.length
    ? yamlToCards(Object.keys(spec).map((f) => f + ': ' + JSON.stringify(spec[f])).join('\n'))
    : '';
  const comp = sectionCompletion(content, (kindReg.requiredSections || [])).map((c) =>
    chip(c.name + (c.done ? ' ✓' : tr(' 缺')), c.done ? 'ok' : 'mute')
  ).join(' ');
  if (!cards && !comp) return '';
  return (comp ? '<div class="ob-seccomp">' + comp + '</div>' : '') +
    (cards ? '<div data-md-cards="1">' + cards + '</div>' : tr('<div class="ob-empty">该 kind 无字段声明（spec 缺席）。</div>'));
}

/** F-08 ③：scheme 字段卡活体注水（store.kinds 快照聚合缺席时现读 /v1/scheme/kinds——
 * 快照带 kinds 键零请求；缺键旧快照=拉真 Kind 定义渲染 scheme 字段卡，不再「暂不可达」）。 */
function hydrateSchemePanel(el, ctx, kind) {
  const box = el.querySelector('[data-panel-scheme]');
  if (!box) return;
  const paint = (kinds) => {
    const k = (kinds || []).find((x) => x && x.name === kind);
    if (!k) return; // 未注册 kind=保留快照词形（不编造定义）
    box.innerHTML = schemePanel(k, '');
  };
  const snapKinds = ctx.store.kinds();
  if (snapKinds.length) { paint(snapKinds); return; }
  if (ctx.api && typeof ctx.api.kinds === 'function') {
    ctx.api.kinds().then((r) => {
      if (!el.isConnected) return;
      if (r && r.ok && Array.isArray(r.data && r.data.kinds)) paint(r.data.kinds);
    }).catch(() => { /* 快照词形保留 */ });
  }
}

function contentPanel(kind, name, content) {
  if (!String(content || '').trim()) return tr('<div class="ob-empty">本文档暂无正文。</div>');
  if (isMarkdownDoc(kind, name)) {
    return '<div data-md-render="1">' + renderMarkdown(content) + '</div>' + rawDisclosure(tr('原文'), content);
  }
  return '<div data-md-cards="1">' + (yamlToCards(content) || tr('<div class="ob-empty">（空声明）</div>')) + '</div>' + rawDisclosure('YAML', content);
}

/* ---------- 正文懒取（docCache(rv) 命中=复用；未命中=现读后入缓存并重绘） ---------- */
async function hydrateDoc(el, ctx, ref) {
  const stamp = typeof location !== 'undefined' ? location.hash : '';
  const cached = ctx.store.cachedDoc(ref.kind, ref.scope, ref.name, ref.rv);
  if (cached) return cached.content;
  const r = await ctx.api.object(ref.kind, ref.scope, ref.name);
  if (!el.isConnected || (typeof location !== 'undefined' ? location.hash : '') !== stamp) return null;
  if (!r.ok) return null;
  const data = r.data && typeof r.data === 'object' ? r.data : {};
  const content = data.content != null ? String(data.content) : (typeof r.data === 'string' ? r.data : JSON.stringify(r.data, null, 2));
  ctx.store.cacheDoc(ref.kind, ref.scope, ref.name, ref.rv, content);
  return content;
}

/** 规则文档族懒取（Rule kind 在册=逐份现读上限 12；缺席=degraded 降级词形——不再永卡「判定中」）。 */
async function hydrateRules(el, ctx) {
  const stamp = typeof location !== 'undefined' ? location.hash : '';
  const ruleRefs = ctx.store.refs('Rule').slice(0, 12);
  if (!ruleRefs.length) return { hits: [], degraded: false, reason: '', scanned: 0 }; // F-08 ①：无 Rule 在册=诚实「无适用规则」终态（非永卡判定中）
  const rules = [];
  for (const rr of ruleRefs) {
    let content = '';
    const cached = ctx.store.cachedDoc(rr.kind, rr.scope, rr.name, rr.rv);
    if (cached) content = cached.content;
    else {
      const r = await ctx.api.object(rr.kind, rr.scope, rr.name);
      if (!el.isConnected || (typeof location !== 'undefined' ? location.hash : '') !== stamp) return null;
      if (r.ok) {
        const data = r.data && typeof r.data === 'object' ? r.data : {};
        content = data.content != null ? String(data.content) : '';
        if (content) ctx.store.cacheDoc(rr.kind, rr.scope, rr.name, rr.rv, content);
      }
    }
    rules.push({ repo: rr.scope, path: rr.name + '.md', tags: tagsOfDoc(content), content });
  }
  return rules;
}

/** 对象工作台渲染（六片）。 */
function renderScreen(el, ctx, ref) {
  const kinds = ctx.store.kinds();
  const kindReg = kinds.find((k) => k && (k.name === ref.kind)) || null;
  const cached = ctx.store.cachedDoc(ref.kind, ref.scope, ref.name, ref.rv);
  const content = cached ? cached.content : '';
  const tags = tagsOfDoc(content);
  const ruleState = rulesPending.get(ref.kind + '|' + ref.name) || null;

  el.innerHTML = '<div class="view-inner">' +
    screenHead(ref.name, tr('对象工作台（六片：标签/规则/上下文/操作/内容/scheme）'),
      kindBadge(ref.kind) + ' ' + chip('r' + String(ref.rv || 0), 'mute')) +
    (pickerHtml(el, ctx, ref) ? '<div style="margin-bottom:10px">' + pickerHtml(el, ctx, ref) + '</div>' : '') +
    '<div class="obj-grid">' +
    tr('<section class="card"><div class="card-h"><span class="h-title">标签</span></div><div class="card-b" data-panel="tags">') + tagPanel(tagGroups(tags, tagVocabFromKinds(kinds))) + '</div></section>' +
    tr('<section class="card"><div class="card-h"><span class="h-title">适用规则</span></div><div class="card-b" data-panel="rules">') +
    (ruleState ? rulePanel(ruleState) : tr('<div class="small faint">规则命中判定中…</div>')) + '</div></section>' +
    tr('<section class="card"><div class="card-h"><span class="h-title">上下文</span></div><div class="card-b">') + ctxPanel(ref.rv) + '</div></section>' +
    tr('<section class="card"><div class="card-h"><span class="h-title">操作</span></div><div class="card-b">') + opsPanel(ref) + '</div></section>' +
    tr('<section class="card obj-content"><div class="card-h"><span class="h-title">内容</span><span class="h-right">') + chip('scope: ' + ref.scope, 'mute') + '</span></div>' +
    '<div class="card-b" data-panel="content">' + (cached ? contentPanel(ref.kind, ref.name, content) : tr('<div class="small faint">正文现读中…</div>')) + '</div></section>' +
    '<section class="card"><div class="card-h"><span class="h-title">scheme</span></div><div class="card-b" data-panel-scheme>' + schemePanel(kindReg, content) + '</div></section>' +
    '</div></div>';

  // F-08 ⑤：切换器换选即深链（router data-go 不吃 select——change 事件直接 go）。
  const picker = el.querySelector('[data-obj-picker]');
  if (picker) {
    picker.addEventListener('change', () => {
      const v = String(picker.value || '');
      if (v) ctx.go(v);
    });
  }

  // F-08 ③：scheme 字段卡活体注水（快照 kinds 缺席=现读 /v1/scheme/kinds）。
  hydrateSchemePanel(el, ctx, ref.kind);

  if (!cached) {
    hydrateDoc(el, ctx, ref).then((c) => {
      if (c == null || !el.isConnected) return;
      renderScreen(el, ctx, ref); // rv 缓存已建——整屏重绘（首版允许）
    });
  }
  if (!ruleState) {
    hydrateRules(el, ctx).then((rules) => {
      if (rules == null || !el.isConnected) return;
      const fresh = ctx.store.cachedDoc(ref.kind, ref.scope, ref.name, ref.rv);
      const res = rulesFor({ kind: ref.kind, scope: ref.scope, name: ref.name, content: fresh ? fresh.content : '', tags }, rules);
      rulesPending.set(ref.kind + '|' + ref.name, res);
      const panel = el.querySelector('[data-panel="rules"]');
      if (panel) panel.innerHTML = rulePanel(res);
    }).catch(() => {
      // 现读失败=诚实终态（不永卡「判定中」——F-08 ①）。
      const panel = el.querySelector('[data-panel="rules"]');
      if (panel) panel.innerHTML = tr('<div class="ob-empty">规则命中判定暂不可读（现读失败）——重进本页即重判。</div>');
    });
  }
}

/** F-08 ⑤：对象快速切换器（顶栏 select——点进工作台不再盲开首个对象；换选即深链）。 */
function pickerHtml(el, ctx, ref) {
  const objs = ctx.store.objects();
  if (!objs.length) return '';
  const opts = objs.slice(0, 300).map((o) => {
    const key = o.kind + '/' + (o.scope || 'global') + '/' + o.name;
    const cur = ref && o.kind === ref.kind && o.name === ref.name && (o.scope || 'global') === (ref.scope || 'global');
    return '<option value="' + esc(objRoute(o.kind, o.scope || 'global', o.name)) + '"' + (cur ? ' selected' : '') + '>' +
      esc(o.kind + ' · ' + o.name) + '</option>';
  }).join('');
  return '<select class="inp" data-obj-picker style="max-width:280px" title="' + esc(tr('切换对象')) + '">' +
    '<option value="">— ' + esc(tr('切换对象（') + objs.length + tr(' 在册）')) + ' —</option>' + opts + '</select>';
}

const rulesPending = new Map(); // 会话内规则命中结果（键 kind|name——换对象现判不跨对象复用）

/** 路由入口：#/obj（默认首个）、#/obj/:kind（该 kind 首个）、#/obj/:kind/:scope/:name。 */
export function render(el, ctx) {
  const p = ctx.params || {};
  let objs = ctx.store.objects();
  let ref = null;
  if (p.kind && p.name) {
    ref = objs.find((o) => o.kind === p.kind && o.name === p.name && (!p.scope || p.scope === 'global' || o.scope === p.scope)) || null;
  } else if (p.kind) {
    ref = objs.find((o) => o.kind === p.kind) || null;
  } else {
    ref = objs[0] || null;
  }
  if (!ref) {
    const hint = p.kind ? 'kind「' + p.kind + '」' : tr('快照');
    el.innerHTML = '<div class="view-inner">' + screenHead(tr('对象工作台'), tr('六片固定：标签/规则/上下文/操作/内容/scheme')) +
      emptyBox(hint + tr('暂无在册对象'), tr('快照 objects 聚合为空或该 kind 无对象（不装假）。')) + '</div>';
    return;
  }
  renderScreen(el, ctx, ref);
}
