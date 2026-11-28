/* router.js —— hash 路由（#/seg/seg）。纯解析/匹配函数与 DOM 装配分离——前者机械直测。
 * 路径段整段单次编码（%2F 形——task id / designdoc path 同词形）：先按 '/' 切段再逐段解码。
 * 词形兼容旧 console 全量路由：#/welcome #/obj(/:kind(/:scope/:name)) #/events #/event/:id
 * #/tasks #/task/:id #/report #/doc/:name #/doc/:kind/:scope/:name #/chat/:agent #/session/:id
 * #/approvals #/inbox #/proposals #/workflows #/wf/:name(/r/:n) #/tree #/docs
 * #/designdoc(/:repo/:path) #/connectors #/s/:name。
 * #22（2026-11-28 R-UX-3 同族）：需参数裸路径（#/task #/doc #/wf…）兜底=「缺参数+可点出口」，
 * 不再把 :agent 这类工程参数拼写甩给人。 */

import { esc } from './format.js';

/* 需参数裸路径的人话出口（键=裸路径首段；只在「已注册模式的前缀命中」时启用——
 * 不给未注册路由编造出口）。#/chat 的常规落点是显式 chat 路由（main.js 重定向默认 agent），
 * 此处保留同词形出口供「只注册 chat/:agent」的嵌入宿主/兜底路径使用。 */
const PARAM_EXITS = {
  chat: { go: 'chat/core-headless', label: '进入会话（默认 agent）' },
  session: { go: 'chat/core-headless', label: '回会话列表' },
  task: { go: 'tasks', label: '回工作与流程' },
  event: { go: 'events', label: '回事件流' },
  doc: { go: 'docs', label: '回文档与资产' },
  wf: { go: 'workflows', label: '回流程运行' },
  prop: { go: 'proposals', label: '回提案' },
  s: { go: 'home', label: '回首页' },
  service: { go: 'welcome', label: '回总览' },
};

/** 解析 hash → 已解码段数组（空 hash=[]；查询串剥除；非法编码段原样保留不抛）。 */
export function parseHash(hash) {
  let h = String(hash || '').replace(/^#\/?/, '');
  h = h.split('?')[0];
  if (h === '') return [];
  return h.split('/').map((s) => {
    try { return decodeURIComponent(s); } catch { return s; }
  });
}

/** 模式匹配：':x'=参数段；字面段须全等。匹配失败=null；成功=params 对象。 */
export function matchSegs(pattern, segs) {
  const ps = String(pattern).split('/');
  if (ps.length !== segs.length) return null;
  const params = {};
  for (let i = 0; i < ps.length; i++) {
    const p = ps[i];
    const s = segs[i];
    if (p.charAt(0) === ':') params[p.slice(1)] = s;
    else if (p !== s) return null;
  }
  return params;
}

/** 路由器工厂（createRouter 不触 DOM——只有 mount/render 访问 location）。 */
export function createRouter(opts = {}) {
  const routes = [];
  let viewEl = null;
  let current = { route: '', pattern: '', segs: [], params: {} };

  function resolve(segs) {
    for (const r of routes) {
      const params = matchSegs(r.pattern, segs);
      if (params) return { r, params };
    }
    return null;
  }

  /** 裸路径「缺参数」诊断：segs 是某已注册模式的前缀且下一段为 :param → 报缺失参数名；
   * 完全不相干的词形仍走未知路径（不冒充）。 */
  function missingParamOf(segs) {
    for (const r of routes) {
      const ps = r.pattern.split('/');
      if (ps.length <= segs.length) continue;
      let prefix = true;
      for (let i = 0; i < segs.length; i++) {
        if (ps[i].charAt(0) === ':' || ps[i] !== segs[i]) { prefix = false; break; }
      }
      if (prefix && ps[segs.length].charAt(0) === ':') {
        return ps.slice(segs.length).filter((p) => p.charAt(0) === ':').map((p) => p.slice(1));
      }
    }
    return null;
  }

  function notFoundHtml(segs, missing) {
    const exit = missing && PARAM_EXITS[segs[0]];
    const act = (exit
      ? '<button class="btn btn-primary" data-go="' + esc(exit.go) + '">' + esc(exit.label) + '</button> '
      : '') + '<button class="btn' + (exit ? ' btn-secondary' : ' btn-primary') + '" data-go="home">回首页</button>';
    if (missing) {
      return '<div class="view-inner">' +
        '<div class="empty"><div class="e-title">缺参数：' + missing.map((p) => '<code>' + esc(p) + '</code>').join('、') + '</div>' +
        '<div class="small faint">这个地址少了参数——不用手拼，点下面入口进。</div>' +
        '<div class="e-act">' + act + '</div></div></div>';
    }
    return '<div class="view-inner">' +
      '<div class="empty"><div class="e-title">未知路径（' + esc(segs.join('/')) + '）</div>' +
      '<div class="small faint">路由词形见控制台导航——或从首页对话进入。</div>' +
      '<div class="e-act">' + act + '</div></div></div>';
  }

  const api = {
    /** 注册屏（pattern 如 'wf/:name/r/:n'；fn(viewEl, params, ctx)）。 */
    route(pattern, fn) { routes.push({ pattern: String(pattern), fn }); return api; },
    resolve,
    get current() { return current; },
    /** 导航（location.hash 变化触发 render；同址=强制重渲染）。 */
    go(path) {
      if (typeof location === 'undefined') return;
      const h = '#/' + String(path || '').replace(/^\/+/, '');
      if (location.hash === h) api.render();
      else location.hash = h;
    },
    /** 装配：挂容器 + 监听 hashchange + 首渲染。 */
    mount(el) {
      viewEl = el;
      if (typeof window !== 'undefined') {
        window.addEventListener('hashchange', () => api.render());
      }
      api.render();
      return api;
    },
    /** 解析当前 hash → 匹配屏执行（未命中=404 词形，不冒充内容）。 */
    render() {
      if (!viewEl || typeof location === 'undefined') return null;
      const segs = parseHash(location.hash);
      if (!segs.length) segs.push(opts.defaultRoute || 'home');
      const hit = resolve(segs);
      if (!hit) {
        current = { route: '404', pattern: '', segs, params: {} };
        viewEl.innerHTML = notFoundHtml(segs, missingParamOf(segs));
        return null;
      }
      current = { route: hit.r.pattern.split('/')[0], pattern: hit.r.pattern, segs, params: hit.params };
      hit.r.fn(viewEl, hit.params);
      return hit.r;
    },
  };
  return api;
}
