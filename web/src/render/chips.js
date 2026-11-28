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

/** #16 agentSession 词形 → 深链目标：真 sessionId（session-*）走 session；
 * 任务键 hint（<scope>/<task>——spawn 受理回填）走 task 详情（防「看会话」断链）。 */
export function sessionTargetOf(sess) {
  const s = String(sess == null ? '' : sess).trim();
  if (!s) return '';
  if (/^session[-_]/i.test(s)) return 'session/' + encSeg(s);
  if (s.includes('/')) return 'task/' + encSeg(s);
  return 'session/' + encSeg(s); // 裸 id=既有词形兜底
}

/** tag 深链（V-01）：docs/tag/:tag——docs 屏第三段=tag 过滤预填。 */
export function tagRoute(tag) {
  return 'docs/tag/' + encSeg(tag);
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
 *     common_connector…）→ 默认 org 仓；部署侧私有仓名由前端注入 window.__CC.repoBase/repoNames 覆写（不入公开仓）.
 * 不匹配=空串（调用方降级纯文本——不编造外链）。 */
const DEFAULT_ORG = (typeof window !== 'undefined' && window.__CC && window.__CC.repoOrg) || 'friddle-self-recursive-agent';
const REPO_NAMES = [
  'core_system', 'core_design', 'core_system_instance', 'core_system_workservice', 'core_system_dsh_plugin',
  'common_connector',
  // 部署侧私有仓名经 window.__CC.repoNames 注入（R-PUB：不入公开仓）
];

/** repoRef → 外部仓库 URL（纯函数——node 单测可机械直测；不匹配=空串）。
 * 词形判定顺序：http(s) 透传 → pkg://git://（首段=仓名或 org——命中已知仓名家族即按
 * 默认 org 解析，否则首段=org）→ github.com 透传 → 裸仓名家族（默认 org）。 */
export function repoUrl(ref) {
  const r = String(ref || '').trim().replace(/^"+|"+$/g, '');
  if (!r) return '';
  if (/^https?:\/\//i.test(r)) return safeHref(r);
  const GH = (org, repo, sub, frag) => {
    const url = 'https://github.com/' + org + '/' + String(repo || '').replace(/\.git$/, '');
    const s = (sub || '').replace(/^\/+/, '');
    return s ? url + '/blob/main/' + s + (frag ? '#' + frag : '') : url + (frag ? '#' + frag : '');
  };
  let m = /^(?:pkg|git|source):\/\/([^/\s]+)(?:\/([^\s#]*))?/.exec(r);
  if (m) {
    const first = m[1], rest = (m[2] || '').replace(/^\/+/, '');
    if (REPO_NAMES.includes(first)) return GH(DEFAULT_ORG, first, rest); // 首段=仓名（A38 无前缀=默认 org）
    return GH(first, rest.split('/')[0] || '', rest.split('/').slice(1).join('/')); // 首段=org
  }
  m = /^github\.com\/([^/\s]+)\/([^/\s#]+)(?:\/([^\s#]*))?(?:#(\S+))?/.exec(r);
  if (m) return GH(m[1], m[2], m[3] || '', m[4] || '');
  // 裸仓名（core_design/goal.md、core_system_workservice/internal/...）→ 默认 org。
  const bare = /^([A-Za-z0-9._-]+)(?:\/([^\s#]+))?$/.exec(r);
  if (bare && REPO_NAMES.includes(bare[1])) return GH(DEFAULT_ORG, bare[1], bare[2] || '');
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

/* 文档状态徽（2026-11-26 二次裁决：所属 workflow 动了=Running→「同步中」，否则「已同步」——
 * 单一语义：同步中=该文档关联的工作流正在运行。
 * runs=snapshot.runs 聚合；匹配面=involved/target/route/workflow 词形包含对象名（窗口内可见事实）。
 * V-07：Running 徽可点——命中 run 可解析=运行详情深链 wf/:name/r/:n（不可解析=纯文本不冒充）。 */
export function syncBadge(runs, name) {
  const active = (runs || []).filter((r) => {
    const st = String((r && (r.status || r.state)) || '');
    if (!st || st === 'done' || st === 'failed' || st === 'canceled') return false;
    let hay = '';
    try { hay = JSON.stringify([r.involved, r.target, r.route, r.workflow]); } catch { hay = ''; }
    return hay.includes(name);
  });
  if (!active.length) return chip('已同步', 'ok');
  const top = active[0];
  const wf = String((top && (top.workflow || top.name)) || '');
  const no = top && (top.run != null ? top.run : (top.no != null ? top.no : ''));
  if (wf && no !== '' && no != null) {
    return chipGo('同步中', 'wf/' + encSeg(wf) + '/r/' + encSeg(no), 'acc', true, 'Workflow ' + wf + ' #' + no);
  }
  if (wf) return chipGo('同步中', 'wf/' + encSeg(wf), 'acc', true, 'Workflow ' + wf);
  return chip('同步中', 'acc');
}
