/* router.js —— hash 路由（#/seg/seg）。纯解析/匹配函数与 DOM 装配分离——前者机械直测。
 * 路径段整段单次编码（%2F 形——task id / designdoc path 同词形）：先按 '/' 切段再逐段解码。
 * 词形兼容旧 console 全量路由：#/welcome #/obj(/:kind(/:scope/:name)) #/events #/event/:id
 * #/tasks #/task/:id #/report #/doc/:name #/doc/:kind/:scope/:name #/chat/:agent #/session/:id
 * #/approvals #/inbox #/proposals #/workflows #/wf/:name(/r/:n) #/tree #/docs
 * #/designdoc(/:repo/:path) #/connectors #/s/:name。 */

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

  function notFoundHtml(segs) {
    return '<div class="view-inner">' +
      '<div class="empty"><div class="e-title">未知路径（' + segs.map((s) => s).join('/') + '）</div>' +
      '<div class="small faint">路由词形见控制台导航——或从 Welcome 进入。</div>' +
      '<div class="e-act"><button class="btn primary" data-go="welcome">回 Welcome</button></div></div></div>';
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
      if (!segs.length) segs.push(opts.defaultRoute || 'welcome');
      const hit = resolve(segs);
      if (!hit) {
        current = { route: '404', pattern: '', segs, params: {} };
        viewEl.innerHTML = notFoundHtml(segs);
        return null;
      }
      current = { route: hit.r.pattern.split('/')[0], pattern: hit.r.pattern, segs, params: hit.params };
      hit.r.fn(viewEl, hit.params);
      return hit.r;
    },
  };
  return api;
}
