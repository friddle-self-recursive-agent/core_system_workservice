/* objects.js —— 通用对象工作台（R-UX-10 2026-11-28 内容优先：内容 → 相关任务与会话 →
 * 标签 / 适用规则 → 操作；scheme=按钮弹窗按需——人不主动看 scheme，需要修改时才看，修改主通路=对话）。
 * 语义移植旧 console-obj.js（代码全部重写为 ESM+库）。缺片显式「无」不装假（诚实空态词形）。
 * 单一真相：对象枚举/kinds=快照（store）；对象正文=docCache(rv 缓存)+api.object 现读。
 * 规则命中=确定性纯函数 rulesFor()（tag/kind/scope/正文引用四路——命中来源显式标注；
 * 零命中=「无适用规则」诚实空态；规则索引缺席=降级词形不冒充）。
 * 安全红线：一切动态文本 esc/或经 md 渲染管线；零原文直插。 */

import { esc, trunc } from '../core/format.js';
import { chip, chipGo, kv, emptyBox, screenHead, kindBadge, objRoute, docRoute, repoLinkChip, tagRoute } from '../render/chips.js';
import { loadYaml, renderMarkdown, rawDisclosure, splitFrontmatter, yamlTextView, dumpYaml } from '../render/md.js';
import { paintMetaHeader } from '../render/meta-header.js';
import { htmlModal } from './modal.js';
import { openDocChat } from './doc-chat.js';
import { t as tr } from '../core/i18n.js';

/* ---------- 相关任务与会话（R-UX-10 2026-11-28 用户裁决：内容优先+周边信息直达） ----------
 * 确定性纯函数（机械直测）：对象名在任务字段（name/spec.title/summary/source/sourceRef）提及
 * 即命中；run 的 involved 提及即命中。状态人话、session 可点直达。零编造——不在册=诚实空态。 */

/** 相关任务行匹配（纯函数）：Task 对象同名精确 + 他类对象字段提及；行=可点任务+可点会话。
 * 任务行双词形兼容（F-05 同口径）：快照扁平行（phase/agentSession 顶层）∪ API 嵌套行（status/spec 包裹）。 */
export function relatedTasksOf(name, tasks) {
  const n = String(name || '');
  if (!n) return [];
  const out = [];
  (tasks || []).forEach((t) => {
    if (!t) return;
    const spec = t.spec || {};
    const st = t.status || t; // 嵌套缺席=扁平行兜底
    const id = (t.scope || 'global') + '/' + String(t.name || '');
    const hay = [t.name, spec.title, spec.summary, spec.source, spec.sourceRef, t.title]
      .map((x) => String(x || '').toLowerCase()).join(' \n');
    const isSelf = String(t.name || '') === n;
    const mentioned = !isSelf && n.length >= 4 && hay.includes(n.toLowerCase());
    if (!isSelf && !mentioned) return;
    const sess = String(st.agentSession || st.sessionId || '').trim();
    out.push({
      id, self: isSelf,
      title: String(spec.title || t.title || t.name || id),
      phase: String(st.phase || ''),
      agent: String(spec.agent || t.agent || ''),
      session: sess,
      lastAction: String(st.lastAction || ''),
    });
  });
  return out.sort((a, b) => (b.self ? 1 : 0) - (a.self ? 1 : 0)).slice(0, 8);
}

/** 相关 run 行匹配（纯函数）：run.involved 提及对象名即命中；行=可点 workflow run。 */
export function relatedRunsOf(name, runs) {
  const n = String(name || '').toLowerCase();
  if (!n || n.length < 4) return [];
  return (runs || []).filter((r) => String((r && r.involved) || '').toLowerCase().includes(n))
    .slice(0, 8).map((r) => ({
      workflow: String((r && (r.workflow || r.name)) || ''),
      run: r && (r.run != null ? r.run : (r.no != null ? r.no : '')),
      phase: String((r && r.phase) || ''),
      status: String((r && r.status) || ''),
    }));
}

const PHASE_CHIP = {
  running: ['正在做', 'acc'], verifying: ['验证中', 'acc'],
  queued: ['排队中', 'mute'], ready: ['待启动', 'mute'], pending: ['等待拉起', 'mute'],
  blocked: ['受阻', 'warn'], failed: ['失败', 'esc'], done: ['已完成', 'ok'],
  converged: ['已收敛', 'ok'], executing: ['执行中', 'acc'], suspended: ['挂起', 'warn'],
  canceled: ['已取消', 'mute'],
};
function phaseChipText(p) {
  const v = PHASE_CHIP[String(p || '')];
  return v ? chip(v[0], v[1]) : chip(String(p || tr('状态未知')), 'mute');
}

function relatedPanelHtml(tasks, runs) {
  if (!tasks.length && !runs.length) {
    return tr('<div class="ob-empty">无相关任务与工作流运行（快照任务/运行记录提及本对象名零命中——不编造）。</div>');
  }
  const taskRows = tasks.map((x) =>
    '<div class="mini-row"><div class="mr-main"><div>' + phaseChipText(x.phase) + ' <b>' + esc(trunc(x.title, 44)) + '</b> ' +
    (x.agent ? chip(tr('负责 ') + x.agent, 'mute') : '') + '</div>' +
    '<div class="mr-sub small mono">' + esc(trunc(x.id, 48)) +
    (x.lastAction ? ' · ' + esc(trunc(x.lastAction, 60)) : '') + '</div></div>' +
    '<span class="w-ev-time">' +
    (x.session ? chipGo(tr('看会话'), 'session/' + encodeURIComponent(x.session), 'acc', true, tr('查看执行会话')) : '') + ' ' +
    chipGo(tr('任务详情'), 'task/' + encodeURIComponent(x.id), 'mute', true) + '</span></div>'
  ).join('');
  const runRows = runs.map((x) =>
    '<div class="mini-row"><div class="mr-main"><div>' + phaseChipText(x.phase || x.status) + ' <b>' +
    esc(x.workflow) + '</b> ' + chip('run #' + String(x.run != null ? x.run : '?'), 'mono') + '</div></div>' +
    '<span class="w-ev-time">' + chipGo(tr('看运行'), 'wf/' + encodeURIComponent(x.workflow) + '/r/' + encodeURIComponent(String(x.run != null ? x.run : '')), 'mute', true) + '</span></div>'
  ).join('');
  return (taskRows ? '<div class="small faint" style="margin-bottom:4px">' + esc(tr('相关任务')) + '</div>' + taskRows : '') +
    (runRows ? '<div class="small faint" style="margin:8px 0 4px">' + esc(tr('相关工作流运行')) + '</div>' + runRows : '');
}

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

/* ---------- 展示片渲染 ---------- */
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

/* （R-UX-10：上下文片退役——版本 rv 徽章已在头部，文档库入口=编辑器 chip。） */

function opsPanel(ref, content) {
  // 代码 ref 直达仓库（2026-11-26 用户裁决——内部仓与外部仓都可 [ref] 打开代码仓库地址）。
  // N-02 接线补充：connector.download 词形（pkg://<仓>/<子路径>）优先——对象 connector 声明
  // 携带真实仓库路径；缺失=scope/name 词形降级（仓家族名才出 chip，不编造外链）。
  // R-UX-10：编辑器/所属 scope 已前置到内容卡与相关卡头部——操作片只剩仓库直达。
  let downloadRef = '';
  const m = /download:\s*"?(pkg:\/\/[^"\s}]+)/.exec(String(content || ''));
  if (m) downloadRef = m[1];
  const repoChips = [downloadRef, ref.scope, ref.name]
    .map((x) => repoLinkChip(String(x || ''), String(x || '')))
    .filter(Boolean).join(' ');
  return repoChips ? '<div class="ob-ctx">' + repoChips + '</div>' : tr('<div class="ob-empty">无仓库直达声明（connector.download/scope 词形零命中）。</div>');
}

function schemePanel(kindReg, content) {
  if (!kindReg) return tr('<div class="ob-empty">scheme 暂不可达（kinds 聚合缺席）。</div>');
  const comp = sectionCompletion(content, (kindReg.requiredSections || [])).map((c) =>
    chip(c.name + (c.done ? ' ✓' : tr(' 缺')), c.done ? 'ok' : 'mute')
  ).join(' ');
  const row = {};
  if (kindReg.name != null) row.name = kindReg.name;
  if (kindReg.requiredSections != null) row.requiredSections = kindReg.requiredSections;
  if (kindReg.spec != null) row.spec = kindReg.spec;
  const yamlText = dumpYaml(row);
  if (!yamlText && !comp) return '';
  return (comp ? '<div class="ob-seccomp" style="margin-bottom:8px">' + comp + '</div>' : '') +
    (yamlText ? yamlTextView(yamlText, tr('kind 定义（YAML 纯文本）')) : tr('<div class="ob-empty">该 kind 无字段声明（spec 缺席）。</div>'));
}

/** R-UX-10：scheme=弹窗按需（人不主动看 scheme，需要修改时才看——修改主通路=对话）。
 * 点击「scheme」按钮现读 kind 定义（快照 kinds 优先，缺席=现读 /v1/scheme/kinds）→ htmlModal 宽体。 */
function bindSchemeModal(el, ctx, kind, content) {
  const btn = el.querySelector('[data-obj-scheme]');
  if (!btn) return;
  btn.addEventListener('click', async () => {
    const paint = (kinds) => {
      const k = (kinds || []).find((x) => x && x.name === kind) || null;
      htmlModal({ title: tr('scheme · ') + kind, wide: true, html: schemePanel(k, content || '') });
    };
    const snapKinds = ctx.store.kinds();
    if (snapKinds.length) { paint(snapKinds); return; }
    if (ctx.api && typeof ctx.api.kinds === 'function') {
      try {
        const r = await ctx.api.kinds();
        paint(r && r.ok && r.data && Array.isArray(r.data.kinds) ? r.data.kinds : []);
      } catch { paint([]); }
    } else paint([]);
  });
}

export function contentPanel(kind, name, content) {
  if (!String(content || '').trim()) return tr('<div class="ob-empty">本文档暂无正文。</div>');
  if (isMarkdownDoc(kind, name)) {
    // R-UX-10：正文先行——frontmatter scheme chips 不再占首屏（scheme 经按钮弹窗按需看）。
    const { body } = splitFrontmatter(content);
    return '<div data-md-render="1">' + renderMarkdown(body) + '</div>' + rawDisclosure(tr('原文'), content);
  }
  // 2026-11-28 用户裁决：scheme 看优先=text-area 纯文本 YAML 全文（结构化视图信息密度太低——
  // 结构化只在编辑器编辑态出现；修改主通路=对话修改）。
  return yamlTextView(content, tr('YAML 声明（纯文本全文）'));
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

/** 对象工作台渲染（R-UX-10 内容优先版：内容 → 相关任务与会话 → 标签/规则 → 操作；scheme=按钮弹窗按需）。 */
function renderScreen(el, ctx, ref) {
  const kinds = ctx.store.kinds();
  const cached = ctx.store.cachedDoc(ref.kind, ref.scope, ref.name, ref.rv);
  const content = cached ? cached.content : '';
  const tags = tagsOfDoc(content);
  const ruleState = rulesPending.get(ref.kind + '|' + ref.name) || null;
  const related = relatedPanelHtml(
    relatedTasksOf(ref.name, ctx.store.tasks()),
    relatedRunsOf(ref.name, ctx.store.runs()));

  el.innerHTML = '<div class="view-inner">' +
    screenHead(ref.name, tr('内容优先：正文与相关任务会话直达；scheme 经按钮按需查看'),
      kindBadge(ref.kind) + ' ' + chip('r' + String(ref.rv || 0), 'mute') +
      ' <button type="button" class="btn btn-chip sm" data-obj-chatfix title="' + esc(tr('对话修改（LLM 改文档→Apply——主通路）')) + '">' + esc(tr('对话修改')) + '</button>' +
      ' <button type="button" class="btn btn-chip sm" data-obj-scheme title="' + esc(tr('scheme 定义（弹窗按需查看）')) + '">scheme</button>') +
    (pickerHtml(el, ctx, ref) ? '<div style="margin-bottom:10px">' + pickerHtml(el, ctx, ref) +
      (ref.kind === 'Provider' ? '<button type="button" class="btn btn-chip sm" data-add-provider style="margin-left:8px" title="' + esc(tr('新建 Provider（表单手填=降级通路——默认对话修改）')) + '">+ ' + esc(tr('添加 Provider')) + '</button>' : '') +
      '</div>' : '') +
    '<div data-provider-create></div>' +
    (ref.kind === 'Agent' ? '<div data-agent-binding style="margin-bottom:10px"><div class="small faint">' + esc(tr('Provider 绑定现读中…')) + '</div></div>' : '') +
    tr('<section class="card obj-content" style="margin-bottom:12px"><div class="card-h"><span class="h-title">内容</span><span class="h-right">') +
    chip('scope: ' + ref.scope, 'mute') + ' ' + chipGo(tr('编辑器'), docRoute(ref.kind, ref.scope, ref.name), 'mute', true) + '</span></div>' +
    '<div class="card-b"><div data-meta-header-slot style="margin-bottom:8px">' + (cached ? '' : '') + '</div>' +
    '<div data-panel="content">' + (cached ? contentPanel(ref.kind, ref.name, content) : tr('<div class="small faint">正文现读中…</div>')) + '</div></div></section>' +
    tr('<section class="card" style="margin-bottom:12px"><div class="card-h"><span class="h-title">相关任务与会话</span><span class="h-right">') +
    chipGo(tr('所属 scope'), 's/' + encodeURIComponent(ref.scope), 'mute', true) + '</span></div>' +
    '<div class="card-b">' + related + '</div></section>' +
    '<div class="obj-grid">' +
    tr('<section class="card"><div class="card-h"><span class="h-title">标签</span></div><div class="card-b" data-panel="tags">') + tagPanel(tagGroups(tags, tagVocabFromKinds(kinds))) + '</div></section>' +
    tr('<section class="card"><div class="card-h"><span class="h-title">适用规则</span></div><div class="card-b" data-panel="rules">') +
    (ruleState ? rulePanel(ruleState) : tr('<div class="small faint">规则命中判定中…</div>')) + '</div></section>' +
    '</div>' +
    '<div class="card" style="margin-top:12px"><div class="card-h"><span class="h-title">' + esc(tr('操作')) + '</span></div><div class="card-b">' +
    opsPanel(ref, cached ? content : '') + '</div></div>' +
    '</div>';

  // F-08 ⑤：切换器换选即深链（router data-go 不吃 select——change 事件直接 go）。
  const picker = el.querySelector('[data-obj-picker]');
  if (picker) {
    picker.addEventListener('change', () => {
      const v = String(picker.value || '');
      if (v) ctx.go(v);
    });
  }

  // R-UX-10：scheme=按钮弹窗按需；对话修改=主通路预填跳会话屏。
  bindSchemeModal(el, ctx, ref.kind, cached ? content : '');
  // 元数据 header（2026-11-28 用户裁决：metadatainfo 侧车上屏——tag/创建/提交人/refs 公用组件）。
  paintMetaHeader(el.querySelector('[data-meta-header-slot]'), ctx, ref, cached ? content : '');
  const chatBtn = el.querySelector('[data-obj-chatfix]');
  if (chatBtn) chatBtn.addEventListener('click', async () => {
    // R-UX-10 二次裁决（2026-11-28）：不跳转——原地嵌入对话修改面板（同对象续跑同会话，
    // 首条自带当前 YAML+意图，助手回 yaml 块→对比→人审 Apply）。
    let content = cached ? content : '';
    if (!content) {
      const r = await ctx.api.object(ref.kind, ref.scope, ref.name);
      content = (r.ok && r.data && r.data.content != null) ? String(r.data.content) : '';
    }
    openDocChat({
      title: tr('对话修改 · ') + ref.kind + '/' + ref.name,
      refKey: 'obj|' + ref.kind + '|' + (ref.scope || 'global') + '|' + ref.name,
      ref: { kind: ref.kind, scope: ref.scope, name: ref.name, rv: ref.rv },
      current: content,
      ctx,
      onApplied: (modified) => {
        ctx.store.cacheDoc(ref.kind, ref.scope, ref.name, (ref.rv || 0) + 1, String(modified || ''));
        renderScreen(el, ctx, ref);
      },
    });
  });

  // Provider 工作台增强（2026-11-27 用户指令：页面上添加 Provider+配置 openai key）。
  if (ref.kind === 'Provider') hydrateProviderCreate(el, ctx);
  // Agent 默认 Provider 绑定编辑（scheme Agent spec.defaultProvider 缺省 openai-default——不同 agent 绑不同 provider）。
  if (ref.kind === 'Agent') hydrateAgentProviderBinding(el, ctx, ref);

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
    el.innerHTML = '<div class="view-inner">' + screenHead(tr('对象工作台'), tr('内容优先：正文/相关任务与会话/标签/规则/操作；scheme 弹窗按需')) +
      emptyBox(hint + tr('暂无在册对象'), tr('暂无在册对象。')) + '</div>';
    return;
  }
  renderScreen(el, ctx, ref);
}

/** 添加 Provider 表单（2026-11-27 用户指令；R-UX-10 二次收紧：表单=手填降级通路——点击弹窗，不占页面）——
 * 组装 Provider 对象 YAML（auth.secret 引用式零明文）apply 创建；默认 provider 名可自定。 */
async function hydrateProviderCreate(el, ctx) {
  const slot = el.querySelector('[data-provider-create]');
  const btn = el.querySelector('[data-add-provider]');
  if (!slot || !btn) return;
  btn.addEventListener('click', () => {
    const form = document.createElement('div');
    form.innerHTML = '<div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:10px">' +
      prow('name', 'openai-2', tr('对象名（如 openai-2）')) +
      prow('account', 'openai-main', tr('账号标识 (auth.account)')) +
      prow('secretSrc', 'secret:llm/openai-key', tr('密钥引用 (auth.secret.source)——零明文')) +
      prow('baseUrl', 'https://api.openai.com/v1', tr('Base 地址 (config.baseUrl)')) +
      prow('lowModel', 'flash-tier', tr('Low 档模型')) +
      prow('stdModel', 'standard-tier', tr('Standard 档模型')) +
      prow('highModel', 'reasoning-tier', tr('High 档模型')) +
      '</div><div style="margin-top:10px;display:flex;gap:8px;align-items:center">' +
      '<button type="button" class="btn btn-chip sm" data-prov-create>' + esc(tr('创建 Provider')) + '</button>' +
      '<span class="small faint" data-prov-create-msg></span></div>';
    const createBtn = form.querySelector('[data-prov-create]');
    if (createBtn) createBtn.addEventListener('click', () => createProvider(el, ctx, form));
    htmlModal({ title: tr('添加 Provider（OpenAI 兼容）——手填降级通路'), wide: true, node: form, holder: slot });
  });
}

function prow(key, def, label) {
  return '<div><div class="small faint">' + esc(label) + '</div>' +
    '<input class="inp mono" data-prov-new="' + key + '" value="' + esc(def) + '" style="width:100%" /></div>';
}

async function createProvider(el, ctx, slot) {
  const get = (k) => { const n = slot.querySelector('[data-prov-new="' + k + '"]'); return n ? String(n.value || '').trim() : ''; };
  const msg = slot.querySelector('[data-prov-create-msg]');
  const name = get('name');
  if (!name) { if (msg) msg.textContent = tr('对象名必填'); return; }
  const yaml = [
    'apiVersion: core/v1',
    'kind: Provider',
    'metadata:',
    '  name: ' + name,
    '  scope: global',
    'spec:',
    '  auth:',
    '    account: ' + (get('account') || 'openai-main'),
    '    secret:',
    '      source: ' + (get('secretSrc') || 'secret:llm/openai-key'),
    '  config:',
    '    baseUrl: ' + (get('baseUrl') || 'https://api.openai.com/v1'),
    '    models:',
    '      low: ' + (get('lowModel') || 'flash-tier'),
    '      standard: ' + (get('stdModel') || 'standard-tier'),
    '      high: ' + (get('highModel') || 'reasoning-tier'),
  ].join('\n');
  if (msg) msg.textContent = tr('正在创建 Provider…');
  const r = await ctx.api.apply({
    kind: 'Provider', scope: 'global', name, resourceVersion: 0,
    content: yaml, fieldManager: 'console:provider-create', reasonRef: 'provider-create:' + name,
  });
  if (msg) msg.textContent = r.ok
    ? tr('Provider 已创建生效 (rv=' + ((r.data && r.data.rv) != null ? r.data.rv : '?') + ')——刷新后可在切换器看到')
    : (tr('创建失败：') + String(r.error || r.status));
}

/** Agent 默认 Provider 绑定（scheme Agent spec.defaultProvider）——R-UX-10 二次收紧：
 * 默认=只读现势一行（值+修改按钮）；表单编辑收「修改绑定」弹窗（手改降级通路）。
 * 现读对象 YAML defaultProvider 行→下拉选在册 Provider→apply 更新（rv 乐观锁）。 */
async function hydrateAgentProviderBinding(el, ctx, ref) {
  const slot = el.querySelector('[data-agent-binding]');
  if (!slot) return;
  const res = await ctx.api.object('Agent', ref.scope, ref.name);
  if (!res.ok || !el.isConnected) { slot.innerHTML = '<div class="small faint">' + esc(tr('Provider 绑定暂不可读')) + '</div>'; return; }
  const raw = String((res.data && res.data.content) || '');
  const rv = (res.data && res.data.resourceVersion != null) ? Number(res.data.resourceVersion) : 0;
  const m = raw.match(/defaultProvider:\s*([^\n#]+)/);
  const cur = m ? m[1].trim().replace(/^["']|["']$/g, '') : 'openai-default';
  const provs = (ctx.store.objects() || []).filter((o) => o.kind === 'Provider');
  // 默认只读行（当前绑定值+修改按钮——表单不占页面）
  slot.innerHTML = '<div class="ob-ctx">' + chip(tr('默认 Provider: ') + cur, 'ok') + ' ' +
    '<button type="button" class="btn btn-chip sm" data-bind-open>' + esc(tr('修改绑定')) + '</button> ' +
    '<span class="small faint">rv=' + esc(String(rv)) + '</span></div>';
  const openBtn = slot.querySelector('[data-bind-open]');
  if (!openBtn) return;
  openBtn.addEventListener('click', () => {
    const opts = provs.map((p) => '<option value="' + esc(p.name) + '"' + (p.name === cur ? ' selected' : '') + '>' + esc(p.name) + '</option>').join('');
    const form = document.createElement('div');
    form.innerHTML = '<div style="display:flex;gap:10px;align-items:center;flex-wrap:wrap">' +
      '<span class="small faint">' + esc(tr('spec.defaultProvider（缺省 openai-default；不同 Agent 可绑不同 Provider）')) + ':</span>' +
      '<select class="inp" data-bind-pick style="max-width:220px"><option value="">' + esc(tr('— 选择在册 Provider（') + provs.length + tr(' 在册）')) + ' —</option>' + opts + '</select>' +
      '<input class="inp mono" data-bind-free style="width:170px" placeholder="' + esc(tr('或手填 provider 名')) + '" />' +
      '<button type="button" class="btn btn-chip sm" data-bind-save>' + esc(tr('保存绑定')) + '</button>' +
      '<span class="small faint" data-bind-msg></span></div>';
    const pick = form.querySelector('[data-bind-pick]');
    const free = form.querySelector('[data-bind-free]');
    if (pick) pick.addEventListener('change', () => { if (free) free.value = String(pick.value || ''); });
    const save = form.querySelector('[data-bind-save]');
    if (save) save.addEventListener('click', async () => {
      const msg = form.querySelector('[data-bind-msg]');
      const val = String((free && free.value) || '').trim().replace(/^["']|["']$/g, '');
      if (!val) { if (msg) msg.textContent = tr('provider 名必填'); return; }
      const next = m && m[1] !== undefined
        ? raw.replace(/defaultProvider:\s*[^\n#]*/, 'defaultProvider: ' + val)
        : raw.replace(/(spec:\n)/, '$1  defaultProvider: ' + val + '\n');
      if (msg) msg.textContent = tr('正在保存绑定…');
      const r = await ctx.api.apply({
        kind: 'Agent', scope: ref.scope, name: ref.name, resourceVersion: rv,
        content: next, fieldManager: 'console:agent-provider-bind', reasonRef: 'agent-bind:' + ref.name,
      });
      if (msg) msg.textContent = r.ok
        ? tr('绑定已保存生效 (rv=' + ((r.data && r.data.rv) != null ? r.data.rv : '?') + ')——该 Agent 后续调用改用此 Provider')
        : (tr('保存失败：') + String(r.error || r.status));
    });
    htmlModal({ title: tr('修改 Provider 绑定 · ') + ref.name, wide: true, node: form, holder: slot });
  });
}
