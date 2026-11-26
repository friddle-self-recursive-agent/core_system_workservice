/* api.js —— REST 读/写面 fetch 封装（base 可配——dsh 宿主代理转发 cored 同形路径）。
 * 判别联合返回：{ok:true, data} | {ok:false, status, error}——调用方按 ok 分支渲染，
 * 读失败=显式降级词形（零 4xx 假装成功、不缓存失败态——下一拍自动重读）。
 * 端点契约（与并行后端 agent/web-backend-refactor 对接词形）：
 *   GET /v1/snapshot · GET /v1/stream(WS, ws.js) · GET /v1/events/{id} · GET /v1/loop/current
 *   以及既有只读 GET 面（objects/drafts/workflows/sessions/inbox/proposals/approvals/design-docs）
 *   与唯一写通路 POST /v1/apply、草稿 POST /v1/drafts。 */

export function createApi(base = '') {
  const B = String(base || '').replace(/\/+$/, '');
  const enc = encodeURIComponent;

  async function request(path, opts = {}) {
    const url = B + '/' + String(path).replace(/^\/+/, '');
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
    object: (kind, scope, name) =>
      request('v1/objects/' + enc(kind) + '/' + enc(name) + (scope && scope !== 'global' ? '?scope=' + enc(scope) : '')),
    kinds: () => request('v1/scheme/kinds'),
    drafts: () => request('v1/drafts'),
    draft: (kind, name, scope) =>
      request('v1/drafts/' + enc(kind) + '/' + enc(name) + (scope && scope !== 'global' ? '?scope=' + enc(scope) : '')),
    events: (limit = 200) => request('v1/events?limit=' + Number(limit)),
    workflows: () => request('v1/workflows'),
    workflowRuns: () => request('v1/workflows/runs'),
    sessions: () => request('v1/sessions'),
    session: (id) => request('v1/sessions/' + enc(String(id == null ? '' : id))),
    inbox: () => request('v1/inbox'),
    proposals: () => request('v1/proposals'),
    approvals: (limit = 20) => request('v1/approvals?limit=' + Number(limit)),
    designDocs: () => request('v1/design-docs'),
    designDoc: (repo, path) => request('v1/design-docs/' + enc(repo) + '/' + enc(path)),
    /* ---------- 写通路（唯一入口——客户端零旁路） ---------- */
    apply: (payload) => request('v1/apply', { method: 'POST', body: payload }),
    saveDraft: (payload) => request('v1/drafts', { method: 'POST', body: payload }),
  };
}
