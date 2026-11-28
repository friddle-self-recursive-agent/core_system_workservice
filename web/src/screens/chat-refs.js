/* chat-refs.js —— 对话引用条 + 详情弹窗（2026-11-27 B 批，chat-ds.md §二：用户点名词形）。
 * B1 引用抽取（纯函数）：转录节点收集①ref 三段词 Kind/scope/name②工具 args 与消息文本中的
 * 文件路径③ContextPack chunks（索引候选）——去重保序封顶。
 * B2 引用条：图2 词形 chip=路径名+[打开] pill；空=整条不渲染（不冒充）。
 * B3 详情弹窗：obj→api.object 直读 content（markdown 渲染+frontmatter chips+深链）；file→
 * design-docs 仓树按 path 匹配直读，未命中=诚实降级（不冒充文档）。
 * B4：chunks 与引用条同源合流（「上下文」组）。
 * 纯函数零 DOM 可 node 直测；mount* 系浏览器侧（事件委托+异步注水）。 */

import { esc, trunc } from '../core/format.js';
import { REF_KINDS, refRoute } from '../render/refs.js';
import { renderMarkdown, splitFrontmatter, rawDisclosure } from '../render/md.js';
import { metaHeaderHtml, frontMetaOf } from '../render/meta-header.js';
import { t as tr } from '../core/i18n.js';

/* ---------- B1：引用抽取（纯函数） ---------- */

/** 文件路径词形判定（保守：含 / 或已知扩展名；无空白；≤160 字；末段须带字母扩展名——
 * 2026-11-27 活体纠偏：纯数字段词形（HTTP 码串 200/200/405）不是路径，拒绝）。 */
export function looksLikePath(s) {
  const t = String(s || '').trim().replace(/[),.;:」』）】]+$/, '');
  if (!t || t.length > 160 || /\s/.test(t) || t.includes('`') || t.includes('*')) return false;
  if (/[“”'"<>|\\]/.test(t)) return false;
  if (!/[A-Za-z]/.test(t)) return false; // 全数字/符号串拒绝
  if (/^(?:~\/|\.{1,2}\/|(?:[\w@.-]+\/)+)[\w@.-]+$/.test(t)) {
    // 目录词形须至少一段含字母；末段带字母扩展名（open 面需要文件）或首段是家/相对根
    const segs = t.split('/');
    const last = segs[segs.length - 1];
    if (!/\.[A-Za-z][A-Za-z0-9]{0,7}$/.test(last) && !/^(?:~\/|\.{1,2}\/)/.test(t)) return false;
    return true;
  }
  return /\.(md|markdown|yaml|yml|json|ts|tsx|js|mjs|go|py|sh|toml|css|html|sql)$/i.test(t);
}

/** 文本 → 文件路径数组（保序去重；≤cap）。 */
export function extractFilePaths(text, cap = 8) {
  const s = String(text || '');
  const out = [];
  const seen = new Set();
  const re = /(?:~\/|\.{1,2}\/)?(?:[\w@.-]+\/)+[\w@.-]+|[\w@.-]+\.(?:md|markdown|yaml|yml|json|ts|tsx|js|mjs|go|py|sh|toml|css|html|sql)/g;
  let m;
  while ((m = re.exec(s)) !== null && out.length < cap * 3) {
    const p = m[0].replace(/[),.;:」』）】]+$/, '');
    if (!looksLikePath(p) || seen.has(p)) continue;
    seen.add(p);
    out.push(p);
    if (out.length >= cap) break;
  }
  return out;
}

/** 文本 → ref 三段对象数组（Kind/scope/name——REF_KINDS 词表；≤cap）。 */
export function extractDocRefs(text, cap = 12) {
  const s = String(text || '');
  const re = new RegExp('(' + REF_KINDS + ')\\/([A-Za-z0-9_.\\/-]+?)\\/([A-Za-z0-9_.\\-\u4e00-\u9fff]+)', 'g');
  const out = [];
  const seen = new Set();
  let m;
  while ((m = re.exec(s)) !== null && out.length < cap) {
    const key = m[0];
    if (seen.has(key)) continue;
    seen.add(key);
    out.push({ type: 'obj', kind: m[1], scope: m[2], name: m[3], label: key });
  }
  return out;
}

const ARGS_PATH_KEYS = ['file_path', 'path', 'note_path', 'cwd'];

/** 转录节点 → 引用集（B1 主抽取：docs=ref 三段（先文件类 kind 优先）+files=文件路径）。
 * 返回 {docs:[], files:[]}——已去重封顶（docs≤12 files≤8）。 */
export function extractRefs(nodes) {
  const docs = [];
  const docSeen = new Set();
  const files = [];
  const fileSeen = new Set();
  const addDoc = (d) => {
    const key = d.kind + '/' + d.scope + '/' + d.name;
    if (docSeen.has(key) || docs.length >= 12) return;
    docSeen.add(key);
    docs.push({ ...d, label: key });
  };
  const addFile = (p) => {
    if (!p || fileSeen.has(p) || files.length >= 8) return;
    fileSeen.add(p);
    files.push({ type: 'file', path: p, label: p });
  };
  for (const n of (Array.isArray(nodes) ? nodes : [])) {
    if (!n || typeof n !== 'object') continue;
    if (typeof n.text === 'string' && n.text) {
      extractDocRefs(n.text).forEach(addDoc);
      extractFilePaths(n.text).forEach(addFile);
    }
    if (n.kind === 'tool' && n.args) {
      try {
        const a = JSON.parse(String(n.args));
        if (a && typeof a === 'object') {
          for (const k of ARGS_PATH_KEYS) {
            if (typeof a[k] === 'string' && looksLikePath(a[k])) { addFile(a[k].trim()); break; }
          }
        }
      } catch { /* 非 JSON args=跳过（诚实） */ }
    }
  }
  return { docs, files };
}

/** 上下文 chunks → 引用集（B4 同源：kind/scope/name → obj 引用；有 score 带分）。 */
export function chunkRefs(chunks) {
  const rows = (Array.isArray(chunks) ? chunks : []).filter((c) => c && c.kind && c.name);
  const out = [];
  const seen = new Set();
  for (const c of rows) {
    const key = c.kind + '/' + (c.scope || 'global') + '/' + c.name;
    if (seen.has(key)) continue;
    seen.add(key);
    out.push({ type: 'obj', kind: c.kind, scope: c.scope || 'global', name: c.name, label: key, score: Number(c.score) || 0, ctx: true });
    if (out.length >= 8) break;
  }
  return out;
}

/* ---------- B2：引用条（纯 HTML） ---------- */

const refChipHtml = (r, i, group) =>
  '<button class="chat-ref" data-cc-ref-g="' + esc(group) + '" data-cc-ref-i="' + i + '" title="' + esc(r.label) + '">' +
    '<span class="chat-ref-name">' + esc(trunc(r.label, 56)) + '</span>' +
    (r.score ? '<span class="chat-ref-score">' + esc(String(Math.round(r.score))) + '</span>' : '') +
    '<span class="chat-ref-open">' + esc(tr('打开')) + '</span></button>';

/** 引用条 HTML（图2 词形；零引用=空串整条不渲染）。 */
export function refsStripHtml(refs) {
  const r = refs && typeof refs === 'object' ? refs : {};
  const cited = [].concat(Array.isArray(r.docs) ? r.docs : [], Array.isArray(r.files) ? r.files : []);
  const chunks = Array.isArray(r.chunks) ? r.chunks : [];
  if (!cited.length && !chunks.length) return '';
  const groups = [];
  if (cited.length) {
    groups.push('<span class="chat-refs-h">' + esc(tr('引用')) + '</span>' +
      cited.map((x, i) => refChipHtml(x, i, 'cited')).join(''));
  }
  if (chunks.length) {
    groups.push('<span class="chat-refs-h">' + esc(tr('上下文')) + '</span>' +
      chunks.map((x, i) => refChipHtml(x, i, 'chunk')).join(''));
  }
  return '<div class="chat-refs">' + groups.join('') + '</div>';
}

/** 引用条模型 → 弹窗目标解析（g/i → 引用描述符；越界=null 不冒充）。 */
export function refTarget(refs, g, i) {
  const r = refs && typeof refs === 'object' ? refs : {};
  const gi = Number(i);
  if (!Number.isFinite(gi)) return null;
  if (g === 'cited') {
    const cited = [].concat(Array.isArray(r.docs) ? r.docs : [], Array.isArray(r.files) ? r.files : []);
    return cited[gi] || null;
  }
  if (g === 'chunk') {
    return (Array.isArray(r.chunks) ? r.chunks : [])[gi] || null;
  }
  return null;
}

/* ---------- B3：详情弹窗（浏览器挂载） ---------- */

/** 弹窗骨架（零 inline style——词形类挂 chat.css）。 */
function modalShell() {
  return '<div class="chat-modal" data-cc-modal hidden>' +
    '<div class="chat-modal-mask" data-cc-modal-close></div>' +
    '<div class="chat-modal-panel" role="dialog" aria-modal="true">' +
    '<div class="chat-modal-head"><span class="chat-modal-title mono" data-cc-modal-title></span>' +
    '<span class="chat-modal-acts"><button class="btn btn-ghost sm" data-cc-modal-go hidden>' + esc(tr('在文档库中打开')) + '</button>' +
    '<button class="btn btn-ghost sm" data-cc-modal-close>' + esc(tr('关闭')) + '</button></span></div>' +
    '<div class="chat-modal-body" data-cc-modal-body></div></div></div>';
}

async function fetchObjContent(ctxc, ref) {
  const r = await ctxc.api.object(ref.kind, ref.scope, ref.name);
  if (!r || (!r.ok && r.status !== 200)) return { ok: false, note: tr('对象面暂不可读（') + String((r && (r.error || r.status)) || tr('网络')) + tr('）') };
  const d = (r.data && typeof r.data === 'object') ? r.data : {};
  const content = d.content != null ? String(d.content) : '';
  if (!content.trim()) return { ok: false, note: tr('本文档暂无正文。') };
  return { ok: true, content };
}

async function fetchFileContent(ctxc, ref, tree) {
  const p = String(ref.path || '');
  // 仓树匹配：精确 path 或「仓名/ 后缀命中」——逐仓白名单，未命中=诚实降级
  const rows = (tree && Array.isArray(tree.docs)) ? tree.docs : [];
  const hit = rows.find((d) => d && d.path === p)
    || rows.find((d) => d && (d.path === p || d.path.endsWith('/' + p) || p.endsWith('/' + String(d.path))));
  if (!hit || !hit.repo) return { ok: false, note: tr('未在文档库命中（design-docs 白名单仓树无此路径）——来源摘录见对话原文。') };
  const r = await ctxc.api.designDoc(hit.repo, hit.path);
  if (!r || (!r.ok && r.status !== 200)) return { ok: false, note: tr('文档直读失败（') + String((r && (r.error || r.status)) || tr('网络')) + tr('）') };
  const d = (r.data && typeof r.data === 'object') ? r.data : {};
  const content = d.content != null ? String(d.content) : (typeof r.data === 'string' ? r.data : '');
  if (!content.trim()) return { ok: false, note: tr('本文档暂无正文。') };
  return { ok: true, content, repo: hit.repo, path: hit.path };
}

function renderDocBody(content, metaRef) {
  // R-UX-10：正文先行——frontmatter 不占首屏（原文折叠可查；scheme=对话修改通路）。
  // 元数据 header（2026-11-28）：metadatainfo 侧车上屏（弹窗内点击走全仓 data-go 委托）。
  const meta = (metaRef && metaRef.kind) || (metaRef && metaRef.scope != null) || (metaRef && metaRef.name)
    ? frontMetaOf(content, metaRef && metaRef.kind, metaRef && metaRef.scope, metaRef && metaRef.name) : frontMetaOf(content);
  return metaHeaderHtml(meta, {}) +
    '<div data-md-render="1">' + renderMarkdown(splitFrontmatter(content).body) + '</div>' +
    rawDisclosure(tr('原文'), content);
}

/** 挂载引用条事件 + 详情弹窗（el=屏根；一次性）。返回 paint(refs)——每轮转录刷新调。 */
export function mountRefs(el, ctxc) {
  const slot = el.querySelector('[data-cc-refs]');
  if (!slot || !ctxc) return () => {};
  if (!el.querySelector('[data-cc-modal]')) slot.insertAdjacentHTML('afterend', modalShell());
  const modal = el.querySelector('[data-cc-modal]');
  const body = modal.querySelector('[data-cc-modal-body]');
  const title = modal.querySelector('[data-cc-modal-title]');
  const goBtn = modal.querySelector('[data-cc-modal-go]');
  let tree = null;
  let treeTried = false;
  const closeModal = () => { modal.hidden = true; body.innerHTML = ''; };
  modal.querySelector('[data-cc-modal-close].btn').addEventListener('click', closeModal);
  modal.querySelector('.chat-modal-mask').addEventListener('click', closeModal);
  el.addEventListener('keydown', (ev) => { if (ev.key === 'Escape' && !modal.hidden) closeModal(); });
  const openTarget = async (ref) => {
    if (!ref) return;
    modal.hidden = false;
    title.textContent = ref.label || (ref.type === 'file' ? ref.path : '');
    goBtn.hidden = true;
    body.innerHTML = '<div class="small faint">' + esc(tr('现读中…')) + '</div>';
    try {
      const res = ref.type === 'file'
        ? await fetchFileContent(ctxc, ref, tree)
        : await fetchObjContent(ctxc, ref);
      if (modal.hidden) return; // 已关=丢弃过期注水
      if (res.ok) {
        body.innerHTML = renderDocBody(res.content, ref.type === 'obj' ? { kind: ref.kind, scope: ref.scope, name: ref.name } : null);
        if (res.repo && res.path) goBtn.dataset.ccGo = 'designdoc/' + encodeURIComponent(res.repo) + '/' + encodeURIComponent(res.path);
        else if (ref.type === 'obj') goBtn.dataset.ccGo = refRoute(ref.kind, ref.scope, ref.name);
        else goBtn.dataset.ccGo = '';
        goBtn.hidden = !goBtn.dataset.ccGo;
      } else {
        body.innerHTML = '<div class="chat-modal-miss">' +
          '<div class="mono small">' + esc(ref.label || '') + '</div>' +
          '<div class="small faint">' + esc(res.note) + '</div></div>';
      }
    } catch (e) {
      body.innerHTML = '<div class="small faint">' + esc(tr('现读失败（') + String((e && e.message) || e) + tr('）')) + '</div>';
    }
  };
  el.addEventListener('click', (ev) => {
    const chip = ev.target.closest('[data-cc-ref-g]');
    if (chip && slot.contains(chip)) {
      openTarget(refTarget((el.__chatRefs || {}), chip.getAttribute('data-cc-ref-g'), chip.getAttribute('data-cc-ref-i')));
      return;
    }
    const mention = ev.target.closest('[data-cc-mention]');
    if (mention) {
      openTarget({ type: 'file', path: String(mention.getAttribute('data-cc-mention') || ''), label: String(mention.getAttribute('data-cc-mention') || '') });
      return;
    }
    const go = ev.target.closest('[data-cc-modal-go]');
    if (go && go.dataset.ccGo && ctxc.go) ctxc.go(go.dataset.ccGo);
  });
  // paint(refs)：每轮转录刷新调用——条重绘+模型存根；chunks 首次懒拉（B4 同源）
  const paint = (refs) => {
    el.__chatRefs = refs || {};
    if (slot) slot.innerHTML = refsStripHtml(refs);
    if (!treeTried && ctxc.api && typeof ctxc.api.designDocs === 'function') {
      treeTried = true;
      ctxc.api.designDocs().then((r) => { if (r && (r.ok || r.status === 200)) tree = r.data || r; }).catch(() => { /* 仓树缺席=降级词形 */ });
    }
  };
  return paint;
}

/** 内联 code 文件提及分类（B3 联动：path 词形 code → 可点 chip）。浏览器侧 DOM pass。 */
export function classifyMentions(root) {
  if (!root || typeof root.querySelectorAll !== 'function') return root;
  root.querySelectorAll('.chat-md code:not(.md-mention)').forEach((c) => {
    const t = String(c.textContent || '').trim();
    if (t && looksLikePath(t) && t.length <= 160) {
      c.classList.add('md-mention');
      c.setAttribute('data-cc-mention', t);
    }
  });
  return root;
}
