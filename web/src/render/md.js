/* md.js —— markdown / 结构化字段卡渲染（marked + dompurify + js-yaml）。
 * 安全管线（R-UX-4 红线）：marked 输出必经 DOMPurify.sanitize（标签/属性白名单、
 * URI 白名单 http(s)+相对路径——javascript:/data: 通路封死）；运行环境无 DOM
 * （node 单测）=降级 minimalRender（先全转义再结构包装——安全由构造保证）。
 * yamlToCards：js-yaml 解析 → 字段卡 HTML（一切值 esc 后入卡、嵌套深度有界 ≤4、
 * 列表 chip 有界 ≤24）——语义移植旧 console-md.js（代码全部重写）。
 * 裸 YAML 转储恒为折叠次级（rawDisclosure——首屏禁裸转储）。
 * 2026-11-27 D11：代码块头横幅（语言标签|缺省「代码块」+复制+自动换行）——上游 ui-primitives
 * CodeBlock/CodeToolbar 词形；复制/换行走类名事件委托（chat.js 屏侧绑定），零 data:* 注入面。 */

import { marked } from 'marked';
import DOMPurify from 'dompurify';
import yaml from 'js-yaml';
import { esc, safeHref, trunc } from '../core/format.js';
import { t as tr } from '../core/i18n.js';

marked.setOptions({ gfm: true, breaks: true, async: false });

/* sanitize 白名单：文档面常用标签 + 安全属性；ALLOW_DATA_ATTR=false（零 data:* 注入面）。
 * D11：button/type 入白名单（代码块头横幅两钮——类名驱动，零属性参数）。 */
const PURIFY_CFG = {
  ALLOWED_TAGS: ['a', 'p', 'br', 'hr', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6', 'ul', 'ol', 'li',
    'blockquote', 'pre', 'code', 'em', 'strong', 'del', 'table', 'thead', 'tbody', 'tr', 'th', 'td', 'span', 'img', 'button'],
  ALLOWED_ATTR: ['href', 'src', 'alt', 'title', 'type'],
  ALLOW_DATA_ATTR: false,
  // URI 白名单沿用 DOMPurify 默认（http(s)/mailto/相对——javascript:/data: 默认已封）。
};

/** 运行环境是否有可用 sanitizer（浏览器=true；node 单测=false）。 */
export function purifySupported() {
  return !!(DOMPurify && DOMPurify.isSupported);
}

/** sanitize 出口（浏览器路径唯一 HTML 通道）。 */
export function sanitizeHtml(html) {
  if (!purifySupported()) return '';
  return DOMPurify.sanitize(String(html == null ? '' : html), PURIFY_CFG);
}

/* ---------- D11 代码块（纯函数——node 单测直测；marked renderer 与 minimalRender 共用） ---------- */

/** 围栏代码块 → 头横幅卡 HTML（语言标签缺省「代码块」+自动换行+复制；文本恒 esc）。 */
export function mdCodeBlock(text, lang) {
  const label = String(lang || '').trim() || tr('代码块');
  return '<div class="md-code-block wrap"><div class="md-code-head">' +
    '<span class="md-code-lang">' + esc(trunc(label, 24)) + '</span>' +
    '<span class="md-code-acts">' +
    '<button type="button" class="md-code-btn md-code-wrap">' + esc(tr('自动换行')) + '</button>' +
    '<button type="button" class="md-code-btn md-code-copy">' + esc(tr('复制')) + '</button>' +
    '</span></div>' +
    '<pre class="md-pre"><code>' + esc(String(text == null ? '' : text)) + '</code></pre></div>';
}

/** marked 15 renderer.code 覆写（token 词形 {text, lang}）。 */
marked.use({
  renderer: {
    code(tok) {
      const t = (tok && typeof tok === 'object') ? tok : { text: tok };
      return mdCodeBlock(t.text, t.lang);
    },
  },
});

/* ---------- 行内格式（minimalRender 用——输入恒为已转义文本，只做词形包装） ---------- */
function inline(s) {
  let out = String(s)
    .replace(/`([^`]+)`/g, '<code class="md-code">$1</code>')
    .replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
    .replace(/(^|[^*])\*([^*\n]+)\*/g, '$1<em>$2</em>');
  out = out.replace(/\[([^\]]+)\]\(([^)\s]+)\)/g, (m, text, href) => {
    const ok = safeHref(href); // 链接白名单——不过=整锚降级纯文本
    return ok ? '<a class="md-a" href="' + esc(ok) + '">' + text + '</a>' : text + ' <span class="mono faint small">(' + href + ')</span>';
  });
  return out;
}

/** 最小 markdown 渲染（纯转义构造安全——无 sanitizer 环境的降级路径）。
 * 语义移植旧 console-md：标题/围栏代码/引用/表格/列表(两层)/分隔线/行内格式。 */
export function minimalRender(src) {
  const lines = String(src == null ? '' : src).split(/\r?\n/);
  const out = [];
  const para = [];
  const flush = () => { if (para.length) { out.push('<p class="md-p">' + inline(esc(para.join(' '))) + '</p>'); para.length = 0; } };
  const isList = (l) => /^(\s*)([-*+]|\d+[.)])\s+/.test(l);
  let i = 0;
  while (i < lines.length) {
    const line = lines[i];
    if (!line.trim()) { flush(); i += 1; continue; }
    if (/^\s*```/.test(line)) { // 围栏代码块（D11：头横幅卡——语言|缺省「代码块」+复制+换行）
      flush();
      const langM = /^\s*```(\S*)/.exec(line);
      const buf = [];
      i += 1;
      while (i < lines.length && !/^\s*```/.test(lines[i])) { buf.push(lines[i]); i += 1; }
      i += 1;
      out.push(mdCodeBlock(buf.join('\n'), langM ? langM[1] : ''));
      continue;
    }
    const h = /^(#{1,6})\s+(.*)$/.exec(line);
    if (h) {
      flush();
      const lv = Math.min(h[1].length, 6);
      out.push('<h' + lv + ' class="md-h md-h' + lv + '">' + inline(esc(h[2])) + '</h' + lv + '>');
      i += 1; continue;
    }
    if (/^\s*(-{3,}|\*{3,}|_{3,})\s*$/.test(line)) { flush(); out.push('<hr class="md-hr">'); i += 1; continue; }
    if (/^\s*>\s?/.test(line)) { // 引用块
      flush();
      const q = [];
      while (i < lines.length && /^\s*>\s?/.test(lines[i])) { q.push(lines[i].replace(/^\s*>\s?/, '')); i += 1; }
      out.push('<blockquote class="md-quote">' + inline(esc(q.join(' '))) + '</blockquote>');
      continue;
    }
    if (/^\s*\|.+\|\s*$/.test(line) && i + 1 < lines.length && /^\s*\|[\s:|-]+\|\s*$/.test(lines[i + 1])) { // 表格
      flush();
      const heads = line.trim().replace(/^\||\|$/g, '').split('|').map((x) => x.trim());
      i += 2;
      const body = [];
      while (i < lines.length && /^\s*\|.+\|\s*$/.test(lines[i])) {
        body.push(lines[i].trim().replace(/^\||\|$/g, '').split('|').map((x) => x.trim()));
        i += 1;
      }
      out.push('<table class="md-tbl"><thead><tr>' + heads.map((x) => '<th>' + inline(esc(x)) + '</th>').join('') + '</tr></thead><tbody>' +
        body.map((row) => '<tr>' + heads.map((_, ci) => '<td>' + inline(esc(row[ci] == null ? '' : row[ci])) + '</td>').join('') + '</tr>').join('') + '</tbody></table>');
      continue;
    }
    if (isList(line)) { // 列表（两层嵌套缩进）
      flush();
      const items = [];
      while (i < lines.length && isList(lines[i])) {
        const m = /^(\s*)([-*+]|\d+[.)])\s+(.*)$/.exec(lines[i]);
        items.push({ depth: m[1].length >= 2 ? 1 : 0, ord: /\d/.test(m[2]), text: m[3] });
        i += 1;
      }
      let html = '<' + (items[0].ord ? 'ol' : 'ul') + ' class="md-list md-list-root">';
      const stack = [{ tag: items[0].ord ? 'ol' : 'ul' }];
      let openLi = false;
      for (const it of items) {
        if (openLi) { html += '</li>'; openLi = false; }
        const want = it.ord ? 'ol' : 'ul';
        while (stack.length > it.depth + 1) html += '</' + stack.pop().tag + '>';
        if (stack.length === it.depth + 1 && stack[stack.length - 1].tag !== want) {
          html += '</' + stack.pop().tag + '><' + want + ' class="md-list">';
          stack.push({ tag: want });
        }
        while (stack.length < it.depth + 1) { html += '<' + want + ' class="md-list">'; stack.push({ tag: want }); }
        html += '<li class="md-li">' + inline(esc(it.text));
        openLi = true;
      }
      if (openLi) html += '</li>';
      while (stack.length) html += '</' + stack.pop().tag + '>';
      out.push('<div class="md-wrap">' + html + '</div>');
      continue;
    }
    para.push(line.trim());
    i += 1;
  }
  flush();
  return out.join('');
}

/** markdown 渲染入口：浏览器=marked+DOMPurify；无 DOM 环境=minimalRender（同源降级）。 */
export function renderMarkdown(src) {
  const s = String(src == null ? '' : src);
  if (purifySupported()) {
    let raw = '';
    try { raw = marked.parse(s); } catch { raw = ''; }
    const clean = sanitizeHtml(raw);
    return clean || minimalRender(s); // 全滤空=降级纯转义（词形不丢）
  }
  return minimalRender(s);
}

/* ---------- 结构化字段卡（R-UX-2 只读态映射——scheme 控件的展示半边） ---------- */

/** js-yaml 解析（失败=null——调用方诚实降级，不装假卡）。 */
export function loadYaml(src) {
  try {
    const v = yaml.load(String(src == null ? '' : src), { json: true });
    return v && typeof v === 'object' ? v : null;
  } catch {
    return null;
  }
}

/** 常用字段人话释义（缺省=无释义行不冒充）。 */
const FIELD_GLOSS = {
  name: tr('名称'), scope: tr('作用域'), resourceVersion: tr('版本号'), goalRef: tr('目标文档'), descRef: tr('说明文档'),
  docs: tr('相关文档'), externalRefs: tr('外部依赖'), instances: tr('服务实例'), builtinServices: tr('内置服务'),
  agentRefs: tr('对话入口（Agent）'), phase: tr('运行状态'), observedGeneration: tr('已应用的版本'), labels: tr('标签'),
  mounts: tr('已挂载连接'), connectorOps: tr('已挂载连接'), provider: tr('凭据源'), session: tr('当前会话'),
  tags: tr('标签'), title: tr('标题'), dependsOn: tr('依赖'), agent: tr('负责'),
};

/** 值渲染：标量=esc 文本（链接白名单）；列表=逐项 chip（≤24）；映射=子卡递归（深度 ≤4）。 */
function cardVal(v, depth) {
  if (v == null) return '<span class="faint">—</span>';
  if (Array.isArray(v)) {
    if (!v.length) return tr('<span class="faint">（空）</span>');
    return '<div class="md-chips">' + v.slice(0, 24).map((x) => '<span class="chip mute small md-chip">' + cardVal(x, depth + 1) + '</span>').join('') +
      (v.length > 24 ? tr('<span class="faint small">…共 ') + v.length + tr(' 项</span>') : '') + '</div>';
  }
  if (typeof v === 'object') {
    if (depth >= 4) return tr('<span class="mono faint small">（嵌套结构见「YAML」）</span>');
    const keys = Object.keys(v);
    if (!keys.length) return tr('<span class="faint">（空）</span>');
    return '<div class="md-subcards">' + keys.map((k) =>
      '<div class="md-subcard"><div class="md-subcard-k">' + esc(k) + '</div><div class="md-subcard-v">' + cardVal(v[k], depth + 1) + '</div></div>'
    ).join('') + '</div>';
  }
  const s = String(v);
  if (/^(agent|system|session):/.test(s)) return '<span class="lk md-lk">' + esc(s) + '</span>';
  if (/^https?:\/\//i.test(s)) {
    const href = safeHref(s);
    return href ? '<a class="md-a" href="' + esc(href) + '">' + esc(trunc(s, 80)) + '</a>' : esc(trunc(s, 80));
  }
  return esc(trunc(s, 300));
}

function cardHtml(key, v, depth) {
  const g = FIELD_GLOSS[String(key)];
  return '<div class="md-card"><div class="md-card-k">' + esc(String(key)) +
    (g ? '<span class="md-gloss">' + esc(g) + '</span>' : '') +
    '</div><div class="md-card-v">' + cardVal(v, depth) + '</div></div>';
}

/** YAML → 字段卡 HTML（解析失败/空=空串——调用方显式空态词形）。 */
export function yamlToCards(src) {
  const tree = loadYaml(src);
  if (!tree || !Object.keys(tree).length) return '';
  return '<div class="md-cards">' + Object.keys(tree).map((k) => cardHtml(k, tree[k], 0)).join('') + '</div>';
}

/** 折叠次级（裸原文恒折叠——首屏禁裸转储；内容 esc）。 */
export function rawDisclosure(title, content) {
  return '<details class="md-raw"><summary>' + esc(title || tr('原文')) + '</summary>' +
    '<pre class="mono md-raw-pre">' + esc(content == null ? '' : content) + '</pre></details>';
}

/* ---------- YAML 纯文本视图（2026-11-28 用户裁决：scheme 展示纯文本优先） ----------
 * 所有 scheme/对象(YAML) 的「看」默认=text-area 纯文本全文（信息密度优先——结构化视图
 * 只留编辑态）；markdown 散文不受影响（渲染即文本形态）。 */

/** 只读等宽 textarea 全文视图（行数自适应 4..48 有界；全 esc——零原文直插）。 */
export function yamlTextView(content, title) {
  const s = String(content == null ? '' : content);
  const rows = Math.min(Math.max(s.split('\n').length + 1, 4), 48);
  return (title ? '<div class="small faint" style="margin-bottom:4px">' + esc(title) + '</div>' : '') +
    '<textarea class="inp mono" readonly rows="' + rows + '" spellcheck="false" ' +
    'style="width:100%;font-size:12px;line-height:1.5;resize:vertical;white-space:pre">' + esc(s) + '</textarea>';
}

/** 值 → YAML 文本（kind 定义行等 JSON 结构纯文本化；js-yaml dump，失败=空串由调用方出诚实空态）。 */
export function dumpYaml(value) {
  if (value == null) return '';
  try {
    const s = yaml.dump(value, { lineWidth: 240, noRefs: true });
    return typeof s === 'string' ? s : '';
  } catch { return ''; }
}

/* ---------- frontmatter 拆分 + scheme 简易标签（2026-11-27 用户裁决） ----------
 * 文档 = frontmatter(YAML 属性) + markdown 正文。属性在库/metainfo 里，
 * 不是正文内容 → 顶栏用轻量 chip 展示，复杂的收进折叠（默认不展开）；
 * scheme 描述字段一律 esc 纯文本（不走 markdown），正文才走 renderMarkdown。
 */
export function splitFrontmatter(src) {
  const s = String(src == null ? '' : src);
  if (!s.startsWith('---')) return { fm: '', body: s };
  const end = s.indexOf('\n---', 3);
  if (end < 0) return { fm: '', body: s };
  return { fm: s.slice(4, end), body: s.slice(end + 4).replace(/^\n/, '') };
}

function isChipValue(v) {
  if (v == null) return false;
  if (Array.isArray(v)) return v.length <= 4 && v.every((x) => x == null || typeof x !== 'object' || Array.isArray(x));
  if (typeof v === 'object') return false;
  return String(v).length <= 80;
}

/** scheme → 简易标签 HTML（chip 平铺 + 复杂字段 details 折叠，全 esc 纯文本）。 */
export function schemeChips(frontmatter) {
  const fm = String(frontmatter || '').trim();
  if (!fm) return '';
  const tree = loadYaml(fm);
  if (!tree || !Object.keys(tree).length) return '<span class="scheme-chip mono">' + esc(trunc(fm, 80)) + '</span>';
  const chips = [], complex = [];
  Object.keys(tree).forEach((k) => {
    const v = tree[k];
    if (isChipValue(v)) {
      const val = Array.isArray(v) ? v.join(', ') : String(v);
      chips.push('<span class="scheme-chip"><span class="scheme-chip-k">' + esc(k) + '</span>' +
        '<span class="scheme-chip-v">' + esc(trunc(val, 80)) + '</span></span>');
    } else complex.push([k, v]);
  });
  let out = '<div class="scheme-chips">' + chips.join('') + '</div>';
  if (complex.length) {
    const lines = complex.map(([k, v]) => k + ': ' + (typeof v === 'string' ? v : JSON.stringify(v, null, 2))).join('\n');
    out += '<details class="scheme-more"><summary>' + esc(tr('scheme 详细（') + complex.length + tr(' 项，默认折叠）')) + '</summary>' +
      '<pre class="mono scheme-more-pre">' + esc(trunc(lines, 4000)) + '</pre></details>';
  }
  return out;
}
