/* api.js —— REST 读/写面 fetch 封装（base 可配——dsh 宿主代理转发 cored 同形路径）。
 * 判别联合返回：{ok:true, data} | {ok:false, status, error}——调用方按 ok 分支渲染，
 * 读失败=显式降级词形（零 4xx 假装成功、不缓存失败态——下一拍自动重读）。
 * 端点契约（与并行后端 agent/web-backend-refactor 对接词形）：
 *   GET /v1/snapshot · GET /v1/stream(WS, ws.js) · GET /v1/events/{id} · GET /v1/loop/current
 *   以及既有只读 GET 面（objects/drafts/workflows/sessions/inbox/proposals/approvals/design-docs）
 *   与唯一写通路 POST /v1/apply、草稿 POST /v1/drafts。 */

export function createApi(base = '', paths = {}) {
  const B = String(base || '').replace(/\/+$/, '');
  const enc = encodeURIComponent;
  /* 写通路词位（部署基路径注入——dsh 插件代理部署：apply/drafts 走宿主注入主体路由
   * /core-console/{apply,drafts}，不经只读代理前缀；直连同源部署=缺省 v1 词位）。 */
  const P = {
    apply: paths.apply || 'v1/apply',
    drafts: paths.drafts || 'v1/drafts',
  };

  async function request(path, opts = {}) {
    /* `/`开头=绝对词位（部署注入的宿主路由，如 /core-console/apply）不拼 base；其余相对 B。 */
    const p = String(path);
    const url = p.startsWith('/') ? p : B + '/' + p.replace(/^\/+/, '');
    try {
      const res = await fetch(url, {
        method: opts.method || 'GET',
        headers: opts.body !== undefined ? { 'Content-Type': 'application/json' } : undefined,
        body: opts.body !== undefined ? JSON.stringify(opts.body) : undefined,
        credentials: 'same-origin',
      });
      const text = await res.text();
      let data = null;
      try { data = text ? JSON.parse(text) : null; } catch { data = { raw: text }; }
      if (!res.ok) {
        const err = data && (data.error || data.message);
        return { ok: false, status: res.status, error: err ? String(err) : 'HTTP ' + res.status };
      }
      return { ok: true, data };
    } catch (e) {
      return { ok: false, status: 0, error: String((e && e.message) || e) };
    }
  }

  return {
    request,
    /* ---------- 新契约端点 ---------- */
    snapshot: () => request('v1/snapshot'),
    event: (id) => request('v1/events/' + enc(String(id == null ? '' : id))),
    loopCurrent: () => request('v1/loop/current'),
    /* ---------- 既有只读面 ---------- */
    /* R1（§19.2）：scope 恒显式（cored 词形「写入/读取需显式 scope」——global 亦不省，
     * 省略=400「scope 与 storage 模板不相容」噪声 15 条/页+正文恒注水中）。 */
    object: (kind, scope, name) =>
      request('v1/objects/' + enc(kind) + '/' + enc(name) + '?scope=' + enc(scope || 'global')),
    kinds: () => request('v1/scheme/kinds'),
    drafts: () => request('v1/drafts'),
    draft: (kind, name, scope) =>
      request('v1/drafts/' + enc(kind) + '/' + enc(name) + (scope && scope !== 'global' ? '?scope=' + enc(scope) : '')),
    events: (limit = 200) => request('v1/events?limit=' + Number(limit)),
    workflows: () => request('v1/workflows'),
    workflowRuns: () => request('v1/workflows/runs'),
    sessions: () => request('v1/sessions'),
    session: (id) => request('v1/sessions/' + enc(String(id == null ? '' : id))),
    /* 会话上下文契约面（G-14——routes_sessions.go：ContextPack 词形投影+编辑分叉） */
    sessionContext: (id) => request('v1/sessions/' + enc(String(id == null ? '' : id)) + '/context'),
    saveSessionContext: (id, payload) =>
      request('v1/sessions/' + enc(String(id == null ? '' : id)) + '/context', { method: 'PUT', body: payload }),
    revertSessionContext: (id) =>
      request('v1/sessions/' + enc(String(id == null ? '' : id)) + '/context/revert', { method: 'POST' }),
    inbox: () => request('v1/inbox'),
    proposals: () => request('v1/proposals'),
    approvals: (limit = 20) => request('v1/approvals?limit=' + Number(limit)),
    designDocs: () => request('v1/design-docs'),
    designDoc: (repo, path) => request('v1/design-docs/' + enc(repo) + '/' + enc(path)),
    /* V-16/V-17/V-18/V-21（2026-11-26 概念可见性批）：上下文索引 / 逐次正文史 / 历史快照 / 送达投影 / 代码审核只读面 */
    contextIndex: () => request('v1/context/index'),
    docHistory: (kind, scope, name) =>
      request('v1/doc-history/' + enc(kind) + '/' + enc(name) + '?scope=' + enc(scope || 'global')),
    revContent: (kind, scope, name, rv) =>
      request('v1/rev-content/' + enc(kind) + '/' + enc(name) + '?scope=' + enc(scope || 'global') + '&rv=' + enc(String(rv))),
    deliveries: () => request('v1/notifications/deliveries'),
    codeAuditFindings: () => request('v1/code-audit/findings'),
    /* ---------- 写通路（唯一入口——客户端零旁路） ---------- */
    apply: (payload) => request(P.apply, { method: 'POST', body: payload }),
    saveDraft: (payload) => request(P.drafts, { method: 'POST', body: payload }),
    /* 事件注入通路（§19 R3——人工批闸/点火/取消族；部署=宿主路由 /core-console/event-publish
     * （绝对词位——经 plugin 注入主体，source=human 服务端强制）。 */
    eventPub: (payload) => request('/core-console/event-pub', { method: 'POST', body: payload }),
  };
}
