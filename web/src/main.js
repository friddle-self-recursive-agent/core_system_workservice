/* main.js —— 入口：装配 store+ws+router，挂载 app。
 * bundle 形态：IIFE + window.CoreConsoleWeb（dsh plugin 打包器消费——零模块加载器假设；避开宿主原生 app 的 CoreConsole 全局——四次裁决期黑屏定位）。
 * 数据流（单向）：WS/轮询 → store → onChange → 当前屏整屏重绘（首版允许——编辑器
 * textarea 焦点态除外）；屏内异步注水以 location.hash 为 stamp，路由已变即丢弃。
 * 自动挂载：页面存在 [data-core-console] 或 #app 容器即引导；否则由宿主调 CoreConsoleWeb.mount。 */

import '../styles/base.css';
import '../styles/components.css';
import '../styles/screens.css';
import '../styles/chat.css';
import '../styles/home.css';
import '../styles/theme-dsh.css';

import { esc } from './core/format.js';
import { createStore } from './core/store.js';
import { createApi } from './core/api.js';
import { createRouter } from './core/router.js';
import { connectStream, wsUrlOf } from './core/ws.js';

import * as home from './screens/home.js';
import * as welcome from './screens/welcome.js';
import * as events from './screens/events.js';
import * as objects from './screens/objects.js';
import * as tasks from './screens/tasks.js';
import * as editor from './screens/editor.js';
import * as chat from './screens/chat.js';
import * as approvals from './screens/approvals.js';
import * as workflows from './screens/workflows.js';
import * as misc from './screens/misc.js';
import * as tags from './screens/tags.js';
import * as audit from './screens/audit.js';
import { getLocale, onLocaleChange, setLocale, t as tr } from './core/i18n.js';

/* 部署基路径注入（window.__CORE_CONSOLE_BASES__: api/ws/apply/drafts——直连同源部署缺省
 * 全空=内置词位不变；dsh 插件代理部署由壳页预置代理词位，见 plugin routes/console-web.ts）。 */
const BASES = (typeof window !== 'undefined' && window.__CORE_CONSOLE_BASES__) || {};
const store = createStore();
const api = createApi(BASES.api || '', { apply: BASES.apply, drafts: BASES.drafts });
const router = createRouter({ defaultRoute: 'home' });
const ctx = { store, api, router, go: (p) => router.go(p), params: {}, route: '' };
/* 2026-11-27 D 批修复：console ctx 全局出借（sessionctx/chat-refs 消费面——此前只建不借，
 * 上下文条恒空壳「加载中」，引用条/详情弹窗无 api 面。），零新依赖。 */
if (typeof window !== 'undefined') window.__CC_CONSOLE_CTX__ = ctx;

/* 两级控制台（2026-11-26 用户裁决）：系统控制台=系统视角（总览大屏+操作面）；
 * 业务控制台=占位级（后续 ServiceAgent 登录的业务面，先不加东西）。
 * Batch 3 IA 重构（2026-11-26 全页扫描——以人为本 4 主线 + 系统底层折叠）：
 *   主线四入口按用户心智（首页/工作与流程/审批中心/文档与资产），研发钻取面收进
 *   「系统底层」折叠组（默认收起）。标签存源词形（中文原文）——locale 切换靠
 *   sidebarHtml 渲染期 tr(label) 现译。 */
/* 三级菜单树（2026-11-26 四次裁决：学 dsh 原生侧栏——一级=分组头可折叠，二级=条目，
 * 三级=缩进子组；语言切换挪顶栏右上角）。标签存源词形；icon=单色 glyph。 */
/* #22（2026-11-28）：#/chat 裸路径的默认落点——与首页/侧栏「会话」入口同一 agent，
 * 人手输 #/chat 不再落「未知路径」死胡同（R-UX-3 同族）。 */
const DEFAULT_CHAT_AGENT = 'core-headless';
const NAV_TREE = [
  {
    level: 'sys', label: 'Core Console', icon: '▣',
    items: [
      ['home', '开始', '✧'],
      ['welcome', '总览', '⌂'],
      ['tasks', '工作与流程', '⊞'],
      ['workflows', '流程运行', '⟳'],
      ['approvals', '审批中心', '✓'],
      ['docs', '文档与资产', '▤'],
    ],
    sub: { label: '系统底层', items: [
      ['chat/' + DEFAULT_CHAT_AGENT, '会话', '💬'],
      ['tree', '系统树', '⑃'],
      ['tags', '标签', '#'],
      ['designdoc', '设计文档', '✎'],
      ['obj', '对象工作台', '▭'],
      ['connectors', '连接', '⌁'],
      ['events', '事件流', '⚡'],
      ['proposals', '提案', '✦'],
      ['report', '关注报告', '≈'],
      ['audit', '代码审核', '✓'],
    ] },
  },
];
let curLevel = 'sys';
try { curLevel = (typeof localStorage !== 'undefined' && localStorage.getItem('cc.level')) || 'sys'; } catch { /* ignore */ }

/* 语言切换（顶栏右上角——四次裁决：不在侧栏底角）。 */
function langToggleHtml() {
  const cur = getLocale();
  const btn = (id, label) => '<button class="btn btn-ghost sm' + (cur === id ? ' active' : '') + '" data-lang="' + id + '">' + label + '</button>';
  return '<span class="lang-toggle">' + btn('zh', '中文') + btn('en', 'EN') + '</span>';
}

/* 折叠态（section/子组两级） */
const foldState = { '系统底层': false }; // Batch 3 IA：研发钻取组默认收起（主线四入口不被淹没）

/* 审批徽标数（草稿 ∪ ApprovalNeeded——人一眼看到欠账）。 */
function approvalBadge() {
  try {
    const n = (store.drafts() || []).length + (store.events() || []).filter((e) => e && e.type === 'ApprovalNeeded').length;
    return n > 0 ? '<span class="nav-badge">' + n + '</span>' : '';
  } catch { return ''; }
}

/* 侧栏：品牌行 + 三级分组树（一级 section 头=级别折叠；二级条目 icon+徽；三级缩进子组）。 */
function navTree() { return NAV_TREE; }

/* 一级 Tab（2026-11-27 用户裁决「一级在左二级在右/默认在左」）：
 * 一级分组从侧栏移出，横排做顶栏 Tab（默认系统控制台在最左）；
 * 侧栏只渲染当前一级下的二级条目。 */
function levelTabsHtml() {
  return navTree().map((sec) => {
    const on = curLevel === sec.level;
    return '<button class="level-tab' + (on ? ' active' : '') + '" data-level="' + esc(sec.level) + '">' +
      '<span class="nav-ico">' + sec.icon + '</span>' + esc(tr(sec.label)) + '</button>';
  }).join('');
}

function sidebarHtml() {
  const sec = navTree().find((s) => s.level === curLevel) || navTree()[0];
  if (!sec) return '';
  const items = sec.items.map(([r, label, icon]) => {
    const badge = r === 'approvals' ? approvalBadge() : '';
    return '<div class="nav-item l2" data-nav="' + esc(r.split('/')[0]) + '" data-go="' + esc(r) + '">' +
      '<span class="nav-ico">' + icon + '</span>' + esc(tr(label)) + badge + '</div>';
  }).join('');
  const sub = sec.sub
    ? '<div class="nav-sub-h" data-fold="' + esc(sec.sub.label) + '"><span class="nav-ico fold-chev">' + (foldState[sec.sub.label] ? '▸' : '▾') + '</span>' + esc(tr(sec.sub.label)) + '</div>' +
      (foldState[sec.sub.label] ? '' : '<nav class="side-nav l3">' + sec.sub.items.map(([r, label, icon]) =>
        '<div class="nav-item l3" data-nav="' + esc(r.split('/')[0]) + '" data-go="' + esc(r) + '">' +
        '<span class="nav-ico">' + icon + '</span>' + esc(tr(label)) + '</div>').join('') + '</nav>')
    : '';
  return '<nav class="side-nav l2">' + items + '</nav>' + sub;
}

function sideFootHtml() {
  // L-01：侧栏底角状态灯退役（顶栏单灯保留——重复状态灯收口；底角留版本词形占位零状态位）。
  return '<div class="side-foot"><span class="small faint">Core Console</span></div>';
}

/* 路由→面包屑词形（NAV 树现译——§19 R5：裸 route 词「welcome」不给人看）。 */
function crumbLabel(route) {
  const pattern = String(route || '');
  if (pattern.startsWith('wf/') && pattern.includes('/r/')) return tr('运行详情');
  const bare = pattern.split('/')[0];
  for (const sec of NAV_TREE) {
    for (const [r, label] of sec.items) if (r.split('/')[0] === bare) return tr(label);
    if (sec.sub) for (const [r, label] of sec.sub.items) if (r.split('/')[0] === bare) return tr(label);
  }
  const extra = { task: tr('任务详情'), event: tr('事件详情'), doc: tr('文档编辑器'), designdoc: tr('设计文档'), wf: tr('工作流'), session: tr('会话'), service: tr('Service 大屏'), s: tr('系统详情'), inbox: tr('收件箱'), prop: tr('需求提案'), tags: tr('标签'), audit: tr('代码审核') };
  return extra[bare] || bare;
}

function shellHtml() {
  return '<div class="app">' +
    '<aside class="sidebar">' +
    '<div class="side-logo"><span class="logo-mark">C</span>' +
    '<span><span class="logo-name">Core Console</span><br><span class="logo-sub">' + esc(tr('系统工作台')) + '</span></span></div>' +
    '<div data-nav-tree>' + sidebarHtml() + '</div>' +
    sideFootHtml() +
    '</aside>' +
    '<div class="main-wrap"><header class="topbar">' +
    '<nav class="level-tabs" data-level-tabs>' + levelTabsHtml() + '</nav>' +
    '<span class="topbar-crumb" data-topbar-crumb></span>' +
    '<span class="topbar-right">' +
    /* L-01：连接状态灯只保留顶栏一处（侧栏底角双灯退役——重复状态灯收口）。 */
    '<span class="lamp off" data-conn-lamp-top></span>' +
    '<span class="small faint" data-conn-label-top>' + esc(tr('连接中…')) + '</span>' +
    langToggleHtml() +
    '</span></header>' +
    '<main class="main"><div class="view" id="view"></div></main></div></div>';
}

/* 导航高亮（route→nav 前缀映射；§19 R5：report 不再映射 tasks——双高亮残留）。 */
function paintNav(route) {
  document.querySelectorAll('[data-nav]').forEach((n) => {
    const r = String(n.dataset.nav || '');
    const on = route === r ||
      (r === 'tasks' && route === 'task') ||
      (r === 'events' && route === 'event') || (r === 'workflows' && route === 'wf') ||
      (r === 'proposals' && route === 'prop') ||
      (r === 'docs' && route === 'doc') ||
      (r === 'welcome' && (route === 'service')) ||
      (r === 'chat' && (route === 'session' || route.startsWith('chat')));
    n.classList.toggle('active', on);
  });
  // 一级归属随路由推进（Tab + 侧栏同步；一级已移到顶栏，侧栏不再有 nav-section）
  const wantLevel = (function () {
    const hit = navTree().find((s) => s.items.some(([r]) => r === route || r.split('/')[0] === route));
    return hit ? hit.level : null;
  })();
  if (wantLevel && wantLevel !== curLevel) {
    curLevel = wantLevel;
    try { localStorage.setItem('cc.level', curLevel); } catch { /* ignore */ }
    const tree = document.querySelector('[data-nav-tree]'); if (tree) tree.innerHTML = sidebarHtml();
    const tabs = document.querySelector('[data-level-tabs]'); if (tabs) tabs.innerHTML = levelTabsHtml();
  }
}

/* 屏包装（params 注入+当前 route 记录+导航高亮；crumb=NAV 现译词形）。
 * F-02 错误边界：屏渲染抛错=显式错误卡（屏名+错误词形），导航照常 paintNav——
 * 不再静默卡 loading/整屏崩溃导航残留在上一页（loop-87 审计 F-02 收口）。 */
function screenOf(fn) {
  return (el, params) => {
    ctx.params = params;
    ctx.route = router.current.route;
    try {
      fn(el, ctx);
    } catch (err) {
      try {
        const msg = String((err && err.message) || err || 'unknown');
        el.innerHTML = '<div class="view-inner"><div class="dh"><div class="dh-main"><div class="dh-title"><h1>Screen Error</h1></div>' +
          '<div class="dh-sub">该屏渲染失败——错误已显式呈现（不静默卡加载；返回或换屏即恢复）。</div></div></div>' +
          '<div class="card"><div class="card-b"><div class="small faint mono" style="white-space:pre-wrap">' +
          String(msg).replace(/[&<>"']/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c])) +
          '</div><div style="margin-top:10px"><button class="btn btn-secondary sm" data-go="welcome">回总览</button> ' +
          '<button class="btn btn-ghost sm" data-retry-screen="1">重试本屏</button></div></div></div></div>';
        const retry = el.querySelector('[data-retry-screen]');
        if (retry) retry.addEventListener('click', () => { try { fn(el, ctx); } catch { /* 二次失败保持错误卡 */ } });
      } catch { /* 错误卡也失败=保留原 DOM */ }
    }
    paintNav(router.current.route);
    try {
      const crumb = document.querySelector('[data-topbar-crumb]');
      if (crumb) crumb.textContent = crumbLabel(router.current.pattern || router.current.route);
    } catch { /* 无 crumb 环境 */ }
  };
}

/* 连接态绘制（侧栏+顶栏双位同步）。 */
let lastConnState = 'connecting';
function paintConn(state) {
  lastConnState = state;
  const text = { connected: tr('实时（WS patch）'), reconnecting: tr('重连中…'), degraded: tr('降级轮询（30s）'), connecting: tr('连接中…') }[state] || state;
  const cls = { connected: 'ok', reconnecting: 'warn', degraded: 'warn', connecting: 'off' }[state] || 'off';
  [['[data-conn-lamp]', '[data-conn-label]'], ['[data-conn-lamp-top]', '[data-conn-label-top]']].forEach(([ls, ltx]) => {
    document.querySelectorAll(ls).forEach((lamp) => { lamp.className = 'lamp ' + cls + (state === 'connected' ? ' pulse-ok' : ''); });
    document.querySelectorAll(ltx).forEach((label) => { label.textContent = text; });
  });
}

/* 快照后正文预取（rv 变化才 fetch——docCache 语义；并发 4、单轮上限 24）。 */
let hydrating = false;
async function hydrateDocs() {
  if (hydrating) return;
  hydrating = true;
  try {
    const stale = store.staleDocs(24);
    let i = 0;
    const worker = async () => {
      while (i < stale.length) {
        const r = stale[i++];
        const res = await api.object(r.kind, r.scope, r.name);
        if (res.ok) {
          const data = res.data && typeof res.data === 'object' ? res.data : {};
          const content = data.content != null ? String(data.content) : (typeof res.data === 'string' ? res.data : JSON.stringify(res.data, null, 2));
          store.cacheDoc(r.kind, r.scope, r.name, r.rv, content);
        }
      }
    };
    await Promise.all(Array.from({ length: Math.min(4, stale.length) }, worker));
    if (stale.length && (router.current.route === 'tags' || router.current.route === 'docs' || router.current.route === 'obj')) {
      router.render();
    }
  } finally { hydrating = false; }
}

export function mount(root, opts = {}) {
  if (!root) return null;
  root.classList.add('cc-app');
  root.innerHTML = shellHtml();
  const view = root.querySelector('#view');

  // store → 当前屏重绘（合帧已由 store 保证；编辑态焦点保真由 editor 屏自管）
  store.subscribe(() => {
    if (view.isConnected) router.render();
    // L-07 首屏徽标时机：审批徽标随状态帧刷新（首屏徽标不再等 locale 切换才出现）。
    try {
      const tree = root.querySelector('[data-nav-tree]');
      if (tree) tree.innerHTML = sidebarHtml();
    } catch { /* 无侧栏环境 */ }
  });

  // 点击路由委托（data-go 全仓统一入口——chip/行/按钮共用）
  root.addEventListener('click', (ev) => {
    const lv = ev.target.closest('[data-level]');
    if (lv) { ev.preventDefault(); curLevel = lv.dataset.level; try { localStorage.setItem('cc.level', curLevel); } catch { /* ignore */ }
      const tree = root.querySelector('[data-nav-tree]'); if (tree) tree.innerHTML = sidebarHtml();
      const tabs = root.querySelector('[data-level-tabs]'); if (tabs) tabs.innerHTML = levelTabsHtml();
      router.go('home'); return; }
    const fold = ev.target.closest('[data-fold]');
    if (fold) { ev.preventDefault(); const g = fold.dataset.fold; foldState[g] = !foldState[g];
      const tree = root.querySelector('[data-nav-tree]'); if (tree) tree.innerHTML = sidebarHtml(); return; }
    const langBtn = ev.target.closest('[data-lang]');
    if (langBtn) { ev.preventDefault(); setLocale(langBtn.dataset.lang); return; }
    const t = ev.target.closest('[data-go]');
    if (t) { ev.preventDefault(); router.go(t.dataset.go); }
  });

  // locale 变更：侧栏整段重渲染（NAV/文案现译）+连接灯恢复+当前屏重绘+标题
  onLocaleChange(() => {
    const tree = root.querySelector('[data-nav-tree]');
    if (tree) tree.innerHTML = sidebarHtml();
    const tabs = root.querySelector('[data-level-tabs]');
    if (tabs) tabs.innerHTML = levelTabsHtml();
    paintConn(lastConnState);
    if (view.isConnected) router.render();
    try { document.title = 'Core Console'; } catch { /* 无 document 环境 */ }
  });

  router
    .route('home', screenOf(home.render))
    .route('welcome', screenOf(welcome.render))
    .route('obj', screenOf(objects.render))
    .route('obj/:kind', screenOf(objects.render))
    .route('obj/:kind/:scope/:name', screenOf(objects.render))
    .route('events', screenOf(events.render))
    .route('event/:id', screenOf(events.render))
    .route('tasks', screenOf(tasks.render))
    .route('task/:id', screenOf(tasks.render))
    .route('report', screenOf(tasks.renderReport))
    .route('doc/:name', screenOf(editor.render))
    .route('doc/:kind/:scope/:name', screenOf(editor.render))
    /* #22：#/chat 裸路径 → 默认 agent 对话（重定向到可分享的 #/chat/core-headless，
     * 不把 :agent 参数拼写甩给人）；真实对话仍走 chat/:agent 屏。 */
    .route('chat', () => router.go('chat/' + DEFAULT_CHAT_AGENT))
    .route('chat/:agent', screenOf(chat.render))
    .route('session/:id', screenOf(chat.render))
    .route('approvals', screenOf(approvals.render))
    .route('approvals/:focus', screenOf(approvals.render))
    .route('inbox', screenOf(approvals.render))
    .route('proposals', screenOf(misc.render))
    .route('proposals/:name', screenOf(misc.render))
    .route('prop/:name', screenOf(misc.render))
    .route('workflows', screenOf(workflows.render))
    .route('wf/:name', screenOf(workflows.render))
    .route('wf/:name/r/:n', screenOf(workflows.render))
    .route('tree', screenOf(misc.render))
    .route('docs', screenOf(misc.render))
    .route('docs/tag/:tag', screenOf(misc.render))
    .route('tags', screenOf(tags.render))
    .route('designdoc', screenOf(misc.render))
    .route('designdoc/:repo/:path', screenOf(misc.render))
    .route('connectors', screenOf(misc.render))
    .route('audit', screenOf(audit.render))
    .route('s/:name', screenOf(misc.render))
    .route('service/:name', screenOf(welcome.renderService))
    .route('biz', screenOf(misc.renderBiz));
  router.mount(view);

  // WS 主通路（patch 失败=整段 re-fetch 回退）；onState 经 store 广播（同帧触发重绘）
  const stream = connectStream({
    url: BASES.ws || wsUrlOf(opts.location || (typeof location !== 'undefined' ? location : null)),
    base: BASES.api || '',
    onSnapshot: (s) => { store.setConn('connected'); store.init(s); hydrateDocs(); },
    onPatch: (patch, msg) => {
      if (!store.applyPatch(patch, msg)) {
        api.snapshot().then((r) => { if (r.ok) { store.init(r.data); hydrateDocs(); } });
      }
    },
    onState: (s) => { paintConn(s); },
  });

  return { store, api, router, stream };
}

/* 自动挂载：宿主容器约定 [data-core-console] 或 #app。 */
if (typeof document !== 'undefined') {
  const boot = () => {
    const root = document.querySelector('[data-core-console]') || document.getElementById('app');
    if (root && !root.classList.contains('cc-app')) mount(root);
  };
  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', boot);
  else boot();
}

/* 全局形态（IIFE——dsh plugin 面向 window.CoreConsoleWeb 装配；宿主原生 CoreConsole 名被占用）。 */
if (typeof window !== 'undefined') {
  window.CoreConsoleWeb = { mount, store, api, router, connectStream, wsUrlOf };
}
