/* main.js —— 入口：装配 store+ws+router，挂载 app。
 * bundle 形态：IIFE + window.CoreConsole（dsh plugin 打包器消费——零模块加载器假设）。
 * 数据流（单向）：WS/轮询 → store → onChange → 当前屏整屏重绘（首版允许——编辑器
 * textarea 焦点态除外）；屏内异步注水以 location.hash 为 stamp，路由已变即丢弃。
 * 自动挂载：页面存在 [data-core-console] 或 #app 容器即引导；否则由宿主调 CoreConsole.mount。 */

import '../styles/base.css';
import '../styles/components.css';
import '../styles/screens.css';

import { esc } from './core/format.js';
import { createStore } from './core/store.js';
import { createApi } from './core/api.js';
import { createRouter } from './core/router.js';
import { connectStream, wsUrlOf } from './core/ws.js';

import * as welcome from './screens/welcome.js';
import * as events from './screens/events.js';
import * as objects from './screens/objects.js';
import * as tasks from './screens/tasks.js';
import * as editor from './screens/editor.js';
import * as chat from './screens/chat.js';
import * as approvals from './screens/approvals.js';
import * as workflows from './screens/workflows.js';
import * as misc from './screens/misc.js';

/* 部署基路径注入（window.__CORE_CONSOLE_BASES__: api/ws/apply/drafts——直连同源部署缺省
 * 全空=内置词位不变；dsh 插件代理部署由 v2 壳页预置代理词位，见 plugin routes/console-v2.ts）。 */
const BASES = (typeof window !== 'undefined' && window.__CORE_CONSOLE_BASES__) || {};
const store = createStore();
const api = createApi(BASES.api || '', { apply: BASES.apply, drafts: BASES.drafts });
const router = createRouter({ defaultRoute: 'welcome' });
const ctx = { store, api, router, go: (p) => router.go(p), params: {}, route: '' };

const NAV = [
  ['welcome', 'Welcome'], ['obj', '对象工作台'], ['tree', '系统树'], ['docs', '文档库'],
  ['tasks', '任务中心'], ['report', '关注报告'], ['workflows', '工作流'], ['events', '事件'],
  ['approvals', '审批'], ['inbox', '收件箱'], ['proposals', '提案'], ['connectors', '连接层'],
  ['chat/root-agent', '与 root 对话'],
];

function shellHtml() {
  return '<div class="app">' +
    '<aside class="sidebar"><div class="side-logo"><span class="logo-mark">C</span>' +
    '<span><span class="logo-name">Core Console</span><br><span class="logo-sub">web 重构起点</span></span></div>' +
    '<nav class="side-nav">' + NAV.map(([r, label]) =>
      '<div class="nav-item" data-nav="' + esc(r) + '" data-go="' + esc(r) + '">' + esc(label) + '</div>').join('') +
    '</nav><div class="side-foot"><span class="lamp off" data-conn-lamp></span>' +
    '<span class="small faint" data-conn-label>连接中…</span></div></aside>' +
    '<main class="main"><div class="view" id="view"></div></main></div>';
}

function paintConn(state) {
  const lamp = document.querySelector('[data-conn-lamp]');
  const label = document.querySelector('[data-conn-label]');
  const text = { connected: '实时（WS patch）', reconnecting: '重连中…', degraded: '降级轮询（30s）', connecting: '连接中…' }[state] || state;
  const cls = { connected: 'ok', reconnecting: 'warn', degraded: 'warn', connecting: 'off' }[state] || 'off';
  if (lamp) lamp.className = 'lamp ' + cls + (state === 'connected' ? ' pulse-ok' : '');
  if (label) label.textContent = text;
}

function paintNav(route) {
  document.querySelectorAll('[data-nav]').forEach((n) => {
    const on = n.dataset.nav === route || (route === 'task' && n.dataset.nav === 'tasks') ||
      (route === 'event' && n.dataset.nav === 'events') || (route === 'wf' && n.dataset.nav === 'workflows') ||
      (route === 'doc' && n.dataset.nav === 'docs') || (route === 'chat' && String(n.dataset.nav).startsWith('chat/'));
    n.classList.toggle('active', on);
  });
}

function screenOf(fn) {
  return (el, params) => {
    ctx.params = params;
    ctx.route = router.current.route;
    fn(el, ctx);
    paintNav(router.current.route);
  };
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
  } finally { hydrating = false; }
}

export function mount(root, opts = {}) {
  if (!root) return null;
  root.classList.add('cc-app');
  root.innerHTML = shellHtml();
  const view = root.querySelector('#view');

  // store → 当前屏重绘（合帧已由 store 保证；编辑态焦点保真由 editor 屏自管）
  store.subscribe(() => { if (view.isConnected) router.render(); });

  // 点击路由委托（data-go 全仓统一入口——chip/行/按钮共用）
  root.addEventListener('click', (ev) => {
    const t = ev.target.closest('[data-go]');
    if (t) { ev.preventDefault(); router.go(t.dataset.go); }
  });

  router
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
    .route('chat/:agent', screenOf(chat.render))
    .route('session/:id', screenOf(chat.render))
    .route('approvals', screenOf(approvals.render))
    .route('inbox', screenOf(approvals.render))
    .route('proposals', screenOf(misc.render))
    .route('workflows', screenOf(workflows.render))
    .route('wf/:name', screenOf(workflows.render))
    .route('wf/:name/r/:n', screenOf(workflows.render))
    .route('tree', screenOf(misc.render))
    .route('docs', screenOf(misc.render))
    .route('designdoc', screenOf(misc.render))
    .route('designdoc/:repo/:path', screenOf(misc.render))
    .route('connectors', screenOf(misc.render))
    .route('s/:name', screenOf(misc.render));
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

/* 全局形态（IIFE——dsh plugin 面向 window.CoreConsole 装配）。 */
if (typeof window !== 'undefined') {
  window.CoreConsole = { mount, store, api, router, connectStream, wsUrlOf };
}
