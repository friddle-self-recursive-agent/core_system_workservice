/* chips.js —— 共享 HTML 工厂：chip/kv/card/emptyBox/徽标/路由链接/骨架。
 * 约定：凡参数名 *Html 的，调用方传入「已转义/已构造」HTML；其余一律 esc 后入卡。
 * 点击路由统一走 data-go 委托（main.js 根容器接管——不散布 onclick）。 */

import { esc, safeHref } from '../core/format.js';

/** 状态 chip（cls: ok/warn/danger/esc/acc/llm/mute）。 */
export function chip(label, cls = 'mute', tiny = true, attrs = '') {
  return '<span class="chip ' + esc(cls) + (tiny ? ' tiny' : '') + '"' + (attrs ? ' ' + attrs : '') + '>' + esc(label) + '</span>';
}

/** 可点路由 chip（route=去 # 前缀的路径词形——data-go 委托接管）。 */
export function chipGo(label, route, cls = 'mute', tiny = true, title = '') {
  return '<span class="chip ' + esc(cls) + (tiny ? ' tiny' : '') + ' click" data-go="' + esc(route) + '"' +
    (title ? ' title="' + esc(title) + '"' : '') + '>' + esc(label) + '</span>';
}

/** kind 徽标（等宽字体——工程词形在徽标位可用）。 */
export function kindBadge(kind) {
  return '<span class="chip kind tiny">' + esc(kind) + '</span>';
}

/** 键值行（rows: [[k, vHtml]]——vHtml 必须已转义/已构造）。 */
export function kv(rows) {
  return '<div class="kv">' + (rows || []).filter((r) => r && r.length === 2).map(([k, v]) =>
    '<span class="k">' + esc(k) + '</span><span class="v">' + v + '</span>'
  ).join('') + '</div>';
}

/** 卡片（titleHtml 已构造；bodyHtml 已构造）。 */
export function card(titleHtml, bodyHtml, cls = '') {
  return '<div class="card ' + esc(cls) + '">' +
    (titleHtml ? '<div class="card-h"><span class="h-title">' + titleHtml + '</span></div>' : '') +
    '<div class="card-b">' + (bodyHtml || '') + '</div></div>';
}

/** 诚实空态（标题+副词+动作区——零命中不装假）。 */
export function emptyBox(title, sub = '', actionHtml = '') {
  return '<div class="empty"><div class="e-title">' + esc(title) + '</div>' +
    (sub ? '<div class="small faint">' + esc(sub) + '</div>' : '') +
    (actionHtml ? '<div class="e-act">' + actionHtml + '</div>' : '') + '</div>';
}

/** 骨架行（装载中词形——不冒充内容）。 */
export function skeletonRows(n = 3) {
  let html = '';
  for (let i = 0; i < (Number(n) || 3); i++) html += '<div class="skel-row"><div class="skel" style="width:' + (30 + (i * 17) % 50) + '%"></div></div>';
  return '<div class="card">' + html + '</div>';
}

/** 屏头（标题+副词+动作区）。 */
export function screenHead(title, sub = '', actionsHtml = '') {
  return '<div class="dh"><div class="dh-main"><div class="dh-title"><h1>' + esc(title) + '</h1></div>' +
    (sub ? '<div class="dh-sub">' + esc(sub) + '</div>' : '') + '</div>' +
    (actionsHtml ? '<div class="dh-actions">' + actionsHtml + '</div>' : '') + '</div>';
}

/* ---------- 路由词形（%2F 整段单次编码——路由侧整体解回） ---------- */

function encSeg(s) { return encodeURIComponent(String(s == null ? '' : s)); }

/** 对象工作台深链：obj/:kind/:scope/:name。 */
export function objRoute(kind, scope, name) {
  return 'obj/' + encSeg(kind) + '/' + encSeg(scope || 'global') + '/' + encSeg(name);
}

/** 编辑器深链：doc/:kind/:scope/:name。 */
export function docRoute(kind, scope, name) {
  return 'doc/' + encSeg(kind) + '/' + encSeg(scope || 'global') + '/' + encSeg(name);
}

/** 对象键字符串 → 路由（'Kind/scope/name' → obj/…；'#/' 词形剥前缀透传——不做词形判别不二次猜测）。
 * 不匹配=空串（调用方降级纯文本——不编造路由）。 */
export function involvedRoute(s) {
  const r = String(s || '').trim();
  if (!r) return '';
  if (r.startsWith('#/')) {
    const segs = r.slice(2).split('/').filter(Boolean);
    return segs.length ? segs.map(encSeg).join('/') : '';
  }
  const m = /^([A-Za-z][A-Za-z0-9]*)\/([^/\s]+)\/([^/\s]+)$/.exec(r);
  if (m) return 'obj/' + m.slice(1).map(encSeg).join('/');
  return '';
}

/** 安全外链（safeHref 白名单——不过=纯文本降级）。 */
export function safeLink(label, href, cls = 'lk') {
  const ok = safeHref(href);
  return ok
    ? '<a class="' + esc(cls) + '" href="' + esc(ok) + '" target="_blank" rel="noopener noreferrer">' + esc(label) + '</a>'
    : '<span class="mono">' + esc(label) + '</span>';
}
