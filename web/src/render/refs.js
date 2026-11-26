/* refs.js —— 对话消息的 ref 链接化与 expand 折叠渲染核（2026-11-26 用户裁决：
 * chat 中提到的 ref 文档/provider/connector/内部资源必须可点击；对话或思考过程可 expand）。
 * 纯函数零 DOM（node 测试可跑）；输出 HTML 片段一律过 esc——零原文直插。 */

import { esc } from '../core/format.js';
import { t as tr } from '../core/i18n.js';

/** 已知 kind 词表（ref 识别用——与 snapshot schemeKinds 对齐的常用集）。 */
const REF_KINDS = 'Rule|Goal|Desc|Agent|ServiceDesc|ServiceInstance|WorkService|ConnectorOps|Provider|Ref|Role|Group|Workflow|Task|Proposal|Notification|DesignDoc|BugDoc|Template|Event';

/** ref 深链词形（与 objRoute 同构——此处自带避免循环依赖）。 */
export function refRoute(kind, scope, name) {
  const k = encodeURIComponent(kind);
  const s = encodeURIComponent(scope || 'global');
  const n = encodeURIComponent(name);
  return 'obj/' + k + '/' + s + '/' + n;
}

/**
 * linkifyRefs(text) —— 消息文本中 ref 词形链接化：
 *   Kind/scope/name（三段式）与裸 name（命中已知对象名集 names 可选参数）→ 可点 chip。
 *   连接器词形 connector:<name> / provider:<name> → 连接层深链。
 * 非匹配文本照原样 esc 输出——零直插。
 */
export function linkifyRefs(text, names) {
  const s = String(text || '');
  const nameSet = names instanceof Set ? names : (Array.isArray(names) ? new Set(names) : null);
  const re = new RegExp('(' + REF_KINDS + ')\\/([A-Za-z0-9_.\\/-]+?)\\/([A-Za-z0-9_.\\-\u4e00-\u9fff]+)', 'g');
  const connRe = /\b(connector|provider):([A-Za-z0-9_.-]+)/g;
  let html = '';
  let last = 0;
  const spans = [];
  let m;
  while ((m = re.exec(s)) !== null) {
    spans.push({ i: m.index, len: m[0].length, kind: m[1], scope: m[2], name: m[3] });
  }
  while ((m = connRe.exec(s)) !== null) {
    spans.push({ i: m.index, len: m[0].length, conn: m[1], cname: m[2] });
  }
  spans.sort((a, b) => a.i - b.i);
  for (const sp of spans) {
    if (sp.i < last) continue;
    html += esc(s.slice(last, sp.i));
    if (sp.conn) {
      html += '<span class="ref-chip" data-go="connectors">' + esc(sp.conn + ':' + sp.cname) + '</span>';
    } else {
      const route = refRoute(sp.kind, sp.scope, sp.name);
      html += '<span class="ref-chip" data-go="' + esc(route) + '" title="' + esc(tr('查看对象')) + '">' +
        esc(sp.kind + '/' + sp.scope + '/' + sp.name) + '</span>';
    }
    last = sp.i + sp.len;
  }
  html += esc(s.slice(last));
  // 裸对象名（names 集合命中）→ 可点 chip（第二遍——简单文本替换，esc 已完成故用占位法）
  if (nameSet && nameSet.size) {
    for (const nm of nameSet) {
      const needle = esc(nm);
      if (!needle || !html.includes(needle)) continue;
      html = html.split(needle).join(
        '<span class="ref-chip" data-go="' + esc(refRoute('Doc', 'global', nm)) + '" data-ref-name="' + esc(nm) + '">' + needle + '</span>');
    }
  }
  return html;
}

/**
 * expandBlock(innerHtml, label, opts) —— 折叠块（默认收起，点标题展开——思考过程/长内容/工具过程）。
 * opts.open=true 初始展开；opts.cls 附加类。
 */
export function expandBlock(innerHtml, label, opts) {
  const o = opts || {};
  return '<details class="ex-block' + (o.cls ? ' ' + esc(o.cls) : '') + (o.open ? ' open' : '') + '"' + (o.open ? ' open' : '') + '>' +
    '<summary class="ex-sum"><span class="ex-chev">▸</span>' + esc(label || tr('展开')) + '</summary>' +
    '<div class="ex-body">' + (innerHtml || '') + '</div>' +
    '</details>';
}

/**
 * renderMessage(msg, names) —— 单条消息卡：
 *   msg={role:'user'|'assistant'|'system', text, thinking?, tool?, refs?, kind?, ts?}
 *   kind='conclusion' → 结论卡（结论+涉及文档 ref+「完整文档模式」按钮 data-doc-full="Kind/scope/name"）。
 */
export function renderMessage(msg, names) {
  const m = msg || {};
  const role = String(m.role || 'assistant');
  const isUser = role === 'user';
  const text = String(m.text || '');
  const html =
    '<div class="msg ' + (isUser ? 'msg-user' : 'msg-ai') + '">' +
    '<div class="msg-avatar">' + (isUser ? '👤' : '◆') + '</div>' +
    '<div class="msg-body">' +
    '<div class="msg-head"><b class="msg-name">' + esc(isUser ? tr('我') : String(m.name || 'rootAgent')) + '</b>' +
    (m.seq != null ? '<span class="small faint mono">seq ' + esc(String(m.seq)) + '</span>' : '') +
    (m.ts ? '<span class="small faint">' + esc(m.ts) + '</span>' : '') + '</div>' +
    (m.kind === 'conclusion'
      ? '<div class="concl-card">' +
        '<div class="concl-t">' + esc(tr('生成结论')) + '</div>' +
        '<div class="concl-text">' + linkifyRefs(text, names) + '</div>' +
        (Array.isArray(m.refs) && m.refs.length
          ? '<div class="concl-refs">' + m.refs.map((r) => {
            const parts = String(r).split('/');
            return parts.length >= 3
              ? '<span class="ref-chip" data-doc-full="' + esc(r) + '">' + esc(tr('完整文档模式：')) + esc(parts[2]) + '</span>'
              : '';
          }).join('') : '') +
        '<div class="concl-acts">' +
        '<button class="btn sm" data-doc-full="' + esc(String((m.refs && m.refs[0]) || '')) + '">' + esc(tr('完整文档模式')) + '</button>' +
        '<button class="btn sm" data-go="chat/core-headless">' + esc(tr('写入草稿')) + '</button>' +
        '</div></div>'
      : '<div class="msg-text">' + linkifyRefs(text, names) + '</div>') +
    (m.thinking ? expandBlock('<div class="msg-text faint">' + linkifyRefs(m.thinking, names) + '</div>', tr('思考过程')) : '') +
    (m.tool ? expandBlock('<div class="msg-text mono faint">' + esc(String(m.tool)) + '</div>', tr('工具过程')) : '') +
    '</div></div>';
  return html;
}
