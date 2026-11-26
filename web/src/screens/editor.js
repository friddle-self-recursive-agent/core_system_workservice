/* editor.js —— 文档编辑器（正文/原文双视图 · 草稿 diff · 历史 tab · Apply）。
 * 语义移植旧 console-app 编辑器屏（代码全部重写为 ESM+库）：
 *   - 默认视图=渲染正文（markdown/字段卡——R-UX-4 渲染优先）；「原文」=textarea 可编辑次级；
 *   - 编辑即累积草稿（POST /v1/drafts debounced——不动正式文档、不发事件）；
 *   - Apply 唯一入口=右栏吸底按钮（人点为准；LLM 只在提案，机制上无 Apply 权）；
 *   - 乐观锁基准=face rv（Apply 带 baseResourceVersion——409 走稿保留语义，不静默重基）；
 *   - 历史 tab=快照 events 过滤该对象（DocApplied/DocReview——诚实可得数据面）。
 * 工作副本=内存态（ED[key]）；store 变更触发的整屏重绘在编辑态被 main 抑制（textarea 焦点保真）。 */

import { esc, trunc } from '../core/format.js';
import { chip, chipGo, kv, emptyBox, screenHead, kindBadge, docRoute } from '../render/chips.js';
import { renderMarkdown, yamlToCards, rawDisclosure } from '../render/md.js';
import { diffLines, renderDiff, diffStats, hunksOf } from '../render/diff.js';

/* AI 修文档通道契约（2026-11-26 用户裁决：文档不再手编——对话框经 LLM 修文档然后 Apply）：
 * window.__CC.chatFaces.docFix(prompt, ref) => Promise<string>|string——返回建议全文；
 * 未注入=面板显式缺席（人仍可在「原文」视图手编——零旁路）。 */

function aiFaces() {
  return (typeof window !== 'undefined' && window.__CC && window.__CC.chatFaces) || null;
}
import { isMarkdownDoc } from './objects.js';
import { t as tr } from '../core/i18n.js';

/* 工作副本（key='kind|scope|name' → {working, base, loaded, faceRV, baseRV, faceAbsent, view, tab}）。 */
const ED = new Map();
const saveTimers = new Map(); // 草稿写穿 debounce
let notice = '';              // 本次会话操作回执（Apply/草稿保存结果——内存态）

function keyOf(kind, scope, name) { return kind + '|' + (scope || 'global') + '|' + (name || ''); }
function hasDiff(ed) { return !!ed && ed.loaded && ed.working !== ed.base; }

/** 对象键解析：#/doc/:kind/:scope/:name 直链；#/doc/:name 名字解析（跨 kind/scope 唯一才解析——歧义=空态）。 */
function resolveRef(ctx) {
  const p = ctx.params || {};
  const objs = ctx.store.objects();
  if (p.kind && p.name) {
    const hit = objs.find((o) => o.kind === p.kind && o.name === p.name && (!p.scope || p.scope === 'global' || o.scope === p.scope));
    if (hit) return { kind: hit.kind, scope: hit.scope, name: hit.name, rv: hit.rv };
    return { kind: p.kind, scope: p.scope || 'global', name: p.name, rv: 0 };
  }
  if (p.name) {
    const cands = objs.filter((o) => o.name === p.name);
    if (cands.length === 1) return { kind: cands[0].kind, scope: cands[0].scope, name: cands[0].name, rv: cands[0].rv };
    return null; // 歧义/缺席=空态（不发明解析）
  }
  return null;
}

/** 注水：面正文（乐观锁基准）+ 草稿缓存（工作副本来源；无稿零请求）。 */
async function hydrate(ctx, key) {
  const { kind, scope, name, rv } = key.ref;
  const stamp = typeof location !== 'undefined' ? location.hash : '';
  const faceP = ctx.api.object(kind, scope, name);
  const row = ctx.store.drafts().find((d) => d && d.kind === kind && (d.scope || 'global') === scope && d.name === name);
  const draftP = row ? ctx.api.draft(kind, name, scope) : Promise.resolve({ ok: false, status: 404, error: tr('无稿') });
  const [faceR, draftR] = await Promise.all([faceP, draftP]);
  if ((typeof location !== 'undefined' ? location.hash : '') !== stamp) return; // 路由已变=丢弃
  const prev = ED.get(key.key);
  let base = '';
  let faceRV = 0;
  let faceAbsent = false;
  if (faceR.ok) {
    const data = faceR.data && typeof faceR.data === 'object' ? faceR.data : {};
    base = data.content != null ? String(data.content) : (typeof faceR.data === 'string' ? faceR.data : JSON.stringify(faceR.data, null, 2));
    faceRV = Number(data.resourceVersion) || rv || 0;
  } else if (faceR.status === 404 && row) {
    faceAbsent = true; // 在册稿×面缺席=草稿基座窗（Apply=创建通路语义）
  } else {
    notice = tr('正文暂不可达（') + esc(String(faceR.error || faceR.status)) + tr('）——恢复后重进即现读');
  }
  const working = draftR && draftR.ok && draftR.data && draftR.data.content != null ? String(draftR.data.content) : base;
  ED.set(key.key, {
    ref: key.ref,
    working, base,
    faceAbsent, faceRV,
    baseRV: draftR && draftR.ok && draftR.data ? Number(draftR.data.baseResourceVersion) || faceRV : faceRV,
    draftFM: draftR && draftR.ok && draftR.data ? String(draftR.data.fieldManager || '') : '',
    view: (prev && prev.view) || 'render',
    tab: (prev && prev.tab) || 'draft',
    loaded: true,
  });
  paintBody(key.key);
  paintBadges(key.key);
  paintRail(key.key, ctx);
}

/** 渲染正文绘制（markdown 文档类=markdown 渲染；对象正文(YAML spec)=结构化字段卡）。 */
function paintBody(key) {
  const el = document.querySelector('[data-ed-body]');
  const ed = ED.get(key);
  if (!el || !ed) return;
  el.innerHTML = ed.view === 'render'
    ? (isMarkdownDoc(ed.ref.kind, ed.ref.name)
      ? '<div data-md-render="1">' + renderMarkdown(ed.working) + '</div>' + rawDisclosure(tr('原文'), ed.working)
      : '<div data-md-cards="1">' + (yamlToCards(ed.working) || tr('<div class="ob-empty">（空声明）</div>')) + '</div>' + rawDisclosure('YAML', ed.working))
    : '';
  const ta = document.querySelector('[data-ed-textarea]');
  if (ta) {
    ta.hidden = ed.view !== 'raw';
    if (!ta.hidden && ta.value !== ed.working && document.activeElement !== ta) ta.value = ed.working;
  }
  const renderWrap = document.querySelector('[data-ed-view="render"]');
  if (renderWrap) renderWrap.hidden = ed.view !== 'render';
}

/** 徽章/dock 原位更新（编辑中不整屏重绘——textarea 焦点保真）。 */
function paintBadges(key) {
  const ed = ED.get(key);
  if (!ed) return;
  const has = hasDiff(ed);
  const hunks = has ? hunksOf(diffLines(ed.base, ed.working)) : 0;
  const chipEl = document.querySelector('[data-ed-draft-chip]');
  if (chipEl) {
    chipEl.className = has ? 'chip warn tiny' : 'chip ok tiny';
    chipEl.textContent = has ? tr('● 未 Apply（草稿 ') + hunks + tr(' 处变更）') : tr('无草稿');
  }
  const dock = document.querySelector('[data-ed-apply]');
  if (dock) {
    dock.disabled = !has;
    dock.textContent = has ? 'Apply（' + hunks + tr(' 处变更）') : tr('Apply（无变更）');
  }
}

/** 右栏（草稿 diff / 历史 tab）。 */
function paintRail(key, ctx) {
  const box = document.querySelector('[data-ed-rail]');
  const ed = ED.get(key);
  if (!box || !ed) return;
  if (ed.tab === 'ai') {
    const f = aiFaces();
    const canAi = !!(f && typeof f.docFix === 'function');
    box.innerHTML = tr('<div class="card-h"><span class="h-title">AI 修文档</span></div><div class="card-b">') +
      '<div class="small faint">' + esc(tr('流程：写修改意图 → AI 出全文建议 → 写入草稿 → 右下 Apply（乐观锁+预检照走——AI 无特权）。')) + '</div>' +
      '<textarea class="inp" data-ai-prompt rows="3" style="width:100%;margin-top:8px" placeholder="' +
      esc(tr('修改意图，如：补充 X 节；修正 Y 处口径…')) + '"></textarea>' +
      '<button class="btn primary sm" data-ai-ask style="margin-top:8px"' + (canAi ? '' : ' disabled') + '>' + esc(tr('请求 AI 建议')) + '</button>' +
      '<div data-ai-out style="margin-top:10px">' +
      (canAi
        ? '<div class="small faint">' + esc(tr('建议将现读于此——写入草稿后走既有 Apply 通路。')) + '</div>'
        : '<div class="small faint">' + esc(tr('AI 通道未注入（__CC.chatFaces.docFix 缺席）——可在「原文」视图手编，或经会话屏对话后人工贴回。'))) + '</div>' +
      '</div>';
    const askBtn = box.querySelector('[data-ai-ask]');
    if (askBtn) askBtn.addEventListener('click', async () => {
      const promptTa = box.querySelector('[data-ai-prompt]');
      const out = box.querySelector('[data-ai-out]');
      if (!promptTa || !out) return;
      const intent = promptTa.value.trim();
      if (!intent) return;
      askBtn.disabled = true;
      out.innerHTML = '<div class="small faint">' + esc(tr('AI 现读中…')) + '</div>';
      try {
        const sug = await f.docFix(intent, { kind: ed.ref.kind, scope: ed.ref.scope, name: ed.ref.name, base: ed.working });
        const text = String(sug == null ? '' : sug);
        if (!text.trim()) throw new Error(tr('空建议'));
        out.innerHTML = '<div class="small faint">' + esc(tr('AI 建议已到（')) + text.length + ' B）</div>' +
          '<pre class="mono small" style="max-height:220px;overflow:auto;border:1px solid var(--bd,#2a2f3a);padding:8px;border-radius:8px">' + esc(trunc(text, 4000)) + '</pre>' +
          '<button class="btn primary sm" data-ai-apply style="margin-top:8px">' + esc(tr('写入草稿')) + '</button>';
        out.querySelector('[data-ai-apply]').addEventListener('click', () => {
          ed.working = text;
          paintBody(ed.ref.kind + '/' + (ed.ref.scope || 'global') + '/' + ed.ref.name);
          paintBadges(ed.ref.kind + '/' + (ed.ref.scope || 'global') + '/' + ed.ref.name);
          out.innerHTML = '<div class="small faint">' + esc(tr('已写入草稿——右下 Apply 落库（乐观锁照走）。')) + '</div>';
        });
      } catch (e) {
        out.innerHTML = '<div class="small faint">' + esc(tr('AI 建议失败（')) + esc(String((e && e.message) || e)) + '）</div>';
      } finally {
        askBtn.disabled = false;
      }
    });
    return;
  }
  if (ed.tab === 'history') {
    const rows = ctx.store.events()
      .filter((e) => e && String(e.involved || '').includes(ed.ref.name));
    box.innerHTML = tr('<div class="card-h"><span class="h-title">历史</span></div><div class="card-b">') +
      (rows.length
        ? rows.slice(0, 30).map((e) => '<div class="mini-row"><div class="mr-main"><div>' + esc(String(e.type || '')) + '</div>' +
          '<div class="mr-sub small mono">' + esc(trunc(String(e.involved || '—'), 50)) + '</div></div>' +
          '<span class="w-ev-time mono">' + esc(String((e.ts || e.at || e.time || '').slice(0, 16))) + '</span></div>').join('')
        : tr('<div class="small faint">该对象暂无归档事件行（DocApplied/DocReview 等——空态显式）。</div>')) +
      '</div>';
    return;
  }
  if (!hasDiff(ed)) {
    box.innerHTML = tr('<div class="card-h"><span class="h-title">草稿</span></div><div class="card-b">') +
      tr('<div class="small faint">无草稿——编辑「原文」即累积草稿（写穿 /v1/drafts，不动正式文档、不发事件）。</div>') +
      (ed.draftFM ? tr('<div class="small faint" style="margin-top:6px">稿来源 fieldManager：') + esc(ed.draftFM) + '</div>' : '') + '</div>';
    return;
  }
  const lines = diffLines(ed.base, ed.working);
  const stats = diffStats(lines);
  box.innerHTML = tr('<div class="card-h"><span class="h-title">草稿</span><span class="h-right">') +
    chip('+' + stats.add, 'ok') + ' ' + chip('-' + stats.del, 'esc') + '</span></div><div class="card-b">' +
    (ed.draftFM ? tr('<div class="small faint" style="margin-bottom:6px">稿来源 fieldManager：') + esc(ed.draftFM) + '</div>' : '') +
    renderDiff(lines, 40) +
    tr('<div class="small faint" style="margin-top:6px">基准=r') + ed.baseRV + tr('（乐观锁期望值——409 不静默重基）· 稿 ') + ed.working.length + tr('B vs 落库值 ') + ed.base.length + 'B</div>' +
    '</div>';
}

/** 草稿写穿（debounced 800ms——编辑的家；失败=显式注记不静默）。 */
function scheduleSave(ctx, key) {
  const prev = saveTimers.get(key.key);
  if (prev) clearTimeout(prev);
  saveTimers.set(key.key, setTimeout(async () => {
    const ed = ED.get(key.key);
    if (!ed || !hasDiff(ed)) return;
    const r = await ctx.api.saveDraft({
      kind: ed.ref.kind, scope: ed.ref.scope, name: ed.ref.name,
      content: ed.working, baseResourceVersion: ed.baseRV, fieldManager: 'console:web',
    });
    if (!r.ok) {
      notice = tr('草稿保存失败（') + esc(String(r.error || r.status)) + tr('）——工作副本仍在内存，退出前不丢');
      paintNotice();
    }
  }, 800));
}

function paintNotice() {
  const slot = document.querySelector('[data-ed-notice]');
  if (slot) slot.innerHTML = notice ? '<div class="small faint">' + notice + '</div>' : '';
}

/** 编辑器屏本体。 */
function editorScreen(el, ctx, key) {
  const ed = ED.get(key.key);
  const view = ed ? ed.view : 'render';
  el.innerHTML = '<div class="view-inner"><div class="doc-layout"><div class="doc-main">' +
    '<div class="doc-headbar"><span class="doc-path mono">' + esc(key.ref.scope + '/' + key.ref.name) + '</span>' + kindBadge(key.ref.kind) +
    tr('<span class="chip ok tiny" data-ed-draft-chip>无草稿</span>') +
    '<div class="md-switch" style="margin-left:auto">' +
    '<span class="chip click small' + (view === 'render' ? ' acc' : ' mute') + tr('" data-ed-view-btn="render">正文</span>') +
    '<span class="chip click small' + (view === 'raw' ? ' acc' : ' mute') + tr('" data-ed-view-btn="raw">原文</span>') +
    '</div></div>' +
    '<div class="doc-meta-row">' + chip('scope', 'mute') + chipGo(key.ref.scope, 's/' + encodeURIComponent(key.ref.scope), 'mute') +
    '<span class="mono small">' + (ed ? (ed.faceAbsent ? tr('文档缺席（在册稿基座——Apply=创建通路）') : 'r' + ed.faceRV + tr('（乐观锁基准）')) : tr('注水中…')) + '</span></div>' +
    '<div data-ed-notice></div>' +
    '<div class="doc-render-wrap" data-ed-view="render" data-ed-body' + (view === 'render' ? '' : ' hidden') + '></div>' +
    tr('<textarea class="doc-editor" data-ed-textarea spellcheck="false" placeholder="原文编辑——变更即累积草稿"') + (view === 'raw' ? '' : ' hidden') + '></textarea>' +
    tr('<div class="small faint">默认展示渲染后的成品；点「原文」查看与编辑源文件。编辑即累积草稿（POST /v1/drafts——不动正式文档、不发事件）；Apply 唯一入口=右栏吸底按钮。</div>') +
    '</div>' +
    '<div class="doc-collab">' +
    '<div class="collab-tabs">' +
    '<div class="collab-tab' + ((ed && ed.tab) === 'draft' ? ' active' : '') + tr('" data-ed-tab="draft">草稿</div>') +
    '<div class="collab-tab' + ((ed && ed.tab) === 'history' ? ' active' : '') + tr('" data-ed-tab="history">历史</div>') +
    '<div class="collab-tab' + ((ed && ed.tab) === 'ai' ? ' active' : '') + tr('" data-ed-tab="ai">AI 修文档</div>') +
    '</div>' +
    '<div class="collab-body" data-ed-rail></div>' +
    tr('<div class="apply-dock"><button class="btn primary block" data-ed-apply disabled>Apply（无变更）</button></div>') +
    '</div></div></div>';
  paintNotice();
  if (ed && ed.loaded) { paintBody(key.key); paintBadges(key.key); paintRail(key.key, ctx); }
  else hydrate(ctx, key);

  // 视图切换（显隐容器——工作副本零重建）
  el.querySelectorAll('[data-ed-view-btn]').forEach((b) => {
    b.addEventListener('click', () => {
      const cur = ED.get(key.key);
      if (!cur) return;
      cur.view = b.dataset.edViewBtn;
      el.querySelectorAll('[data-ed-view-btn]').forEach((x) => {
        x.className = 'chip click small' + (x.dataset.edViewBtn === cur.view ? ' acc' : ' mute');
      });
      paintBody(key.key);
    });
  });
  // tab 切换
  el.querySelectorAll('[data-ed-tab]').forEach((t) => {
    t.addEventListener('click', () => {
      const cur = ED.get(key.key);
      if (!cur) return;
      cur.tab = t.dataset.edTab;
      el.querySelectorAll('[data-ed-tab]').forEach((x) => x.classList.toggle('active', x === t));
      paintRail(key.key, ctx);
    });
  });
  // textarea 编辑 → 工作副本 + 徽章 + 草稿写穿
  const ta = el.querySelector('[data-ed-textarea]');
  if (ta) {
    if (ed && ed.loaded && document.activeElement !== ta) ta.value = ed.working;
    ta.addEventListener('input', () => {
      const cur = ED.get(key.key);
      if (!cur) return;
      cur.working = ta.value;
      paintBadges(key.key);
      paintRail(key.key, ctx);
      scheduleSave(ctx, key);
    });
  }
  // Apply（人点为准——confirm 两步，409=稿保留语义不静默重基）
  const dock = el.querySelector('[data-ed-apply]');
  if (dock) {
    dock.addEventListener('click', async () => {
      const cur = ED.get(key.key);
      if (!cur || !hasDiff(cur)) return;
      const stats = diffStats(diffLines(cur.base, cur.working));
      if (!window.confirm(tr('Apply 落库 ') + cur.ref.kind + '/' + cur.ref.name + '（+' + stats.add + '/-' + stats.del + tr('）？乐观锁基准 r') + cur.baseRV + '。')) return;
      dock.disabled = true;
      const r = await ctx.api.apply({
        kind: cur.ref.kind, scope: cur.ref.scope, name: cur.ref.name,
        content: cur.working, baseResourceVersion: cur.baseRV, fieldManager: 'console:web',
        reasonRef: 'console:web-editor',
      });
      dock.disabled = false;
      if (!r.ok) {
        notice = tr('Apply 失败（') + esc(String(r.error || r.status)) + '）' + (r.status === 409 ? tr('——乐观锁冲突：稿已保留，请对齐最新 rv 后重试') : '');
        paintNotice();
        return;
      }
      notice = tr('Apply 成功——文档已落库（rv=') + esc(String((r.data && r.data.rv) || '?')) + '）。';
      cur.base = cur.working;
      paintBadges(key.key);
      paintRail(key.key, ctx);
      paintNotice();
      ctx.api.snapshot().then((s) => { if (s.ok) ctx.store.init(s.data); });
    });
  }
}

export function render(el, ctx) {
  const ref = resolveRef(ctx);
  if (!ref) {
    el.innerHTML = '<div class="view-inner">' + screenHead(tr('文档编辑器'), '') +
      emptyBox(tr('对象解析歧义或缺席'), tr('同名对象跨 kind/scope 不唯一——请从系统树进入。'), chipGo(tr('打开系统树'), 'tree', 'primary', false)) + '</div>';
    return;
  }
  const key = { ref, key: keyOf(ref.kind, ref.scope, ref.name) };
  editorScreen(el, ctx, key);
}
