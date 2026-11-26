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

/* ---------- 代码 ref → 仓库地址（2026-11-26 用户裁决：代码最好也能 [ref] 直接打开代码仓库地址，
 * 对外的也可以）。映射口径（R-PUB-1：公开面只出现 org 名与公开仓——真实部署私有仓走部署侧
 * 同名映射或 OneDev，不在前端硬编码私网地址）：
 *   - pkg://<org>/<repo>/<path>   → https://github.com/<org>/<repo>/(tree|blob)/main/<path>
 *   - git://<org>/<repo>          → https://github.com/<org>/<repo>
 *   - github.com/<org>/<repo>[…]  → 透传 https 化
 *   - 常见内部仓名（core_design/core_system_instance/core_system_workservice/core_system_dsh_plugin/
 *     common_connector/sybran_usage_connector…）→ 默认 org 仓（前端注入 window.__CC.repoBase 可覆写）。
 * 不匹配=空串（调用方降级纯文本——不编造外链）。 */
const DEFAULT_ORG = (typeof window !== 'undefined' && window.__CC && window.__CC.repoOrg) || 'friddle-self-recursive-agent';
const REPO_NAMES = [
  'core_system', 'core_system_instance', 'core_system_workservice', 'core_system_dsh_plugin',
  'common_connector', 'sybran_usage_connector', 'sybran-developer-system', 'sybran-business-system',
];

/** repoRef → 外部仓库 URL（纯函数——node 单测可机械直测；不匹配=空串）。 */
export function repoUrl(ref) {
  const r = String(ref || '').trim().replace(/^"+|"+$/g, '');
  if (!r) return '';
  if (/^https?:\/\//i.test(r)) return safeHref(r);
  let m = /^(?:pkg|git|source):\/\/([^/\s]+)(?:\/([^\s#]*))?/.exec(r);
  if (m) {
    const org = m[1], rest = (m[2] || '').replace(/^\/+/, '');
    const url = 'https://github.com/' + org;
    return rest ? url + '/blob/main/' + rest : url;
  }
  m = /^github\.com\/([^/\s]+)\/([^/\s#]+)(?:\/([^\s#]*))?(?:#(\S+))?/.exec(r);
  if (m) {
    const url = 'https://github.com/' + m[1] + '/' + m[2].replace(/\.git$/, '');
    const sub = (m[3] || '').replace(/^\/+/, '');
    const frag = m[4] ? '#' + m[4] : '';
    return sub ? url + '/blob/main/' + sub + frag : url + frag;
  }
  // 裸仓名（core_design/goal.md、core_system_workservice/internal/...）→ 默认 org。
  const bare = /^([A-Za-z0-9._-]+)(?:\/([^\s#]+))?$/.exec(r);
  if (bare && REPO_NAMES.includes(bare[1])) {
    const url = 'https://github.com/' + DEFAULT_ORG + '/' + bare[1];
    return bare[2] ? url + '/blob/main/' + bare[2] : url;
  }
  return '';
}

/** repoRef → 可点外链 chip（不匹配=空串——调用方降级纯文本 chip）。 */
export function repoLinkChip(ref, label) {
  const url = repoUrl(ref);
  if (!url) return '';
  const text = String(label || ref).trim() || ref;
  return '<a class="chip mute tiny click lk" href="' + esc(url) + '" target="_blank" rel="noopener noreferrer" ' +
    'title="' + esc(url) + '">' + esc(truncTxt(text, 44)) + '</a>';
}

function truncTxt(s, n) { return s.length > n ? s.slice(0, n) + '…' : s; }

/* 文档状态徽（2026-11-26 二次裁决：所属 workflow 动了=Running，否则 InSync）。
 * runs=snapshot.runs 聚合；匹配面=involved/target/route/workflow 词形包含对象名（窗口内可见事实）。 */
export function syncBadge(runs, name) {
  const active = (runs || []).some((r) => {
    const st = String((r && (r.status || r.state)) || '');
    if (!st || st === 'done' || st === 'failed' || st === 'canceled') return false;
    let hay = '';
    try { hay = JSON.stringify([r.involved, r.target, r.route, r.workflow]); } catch { hay = ''; }
    return hay.includes(name);
  });
  return active ? chip('Running', 'acc') : chip('InSync', 'ok');
}
