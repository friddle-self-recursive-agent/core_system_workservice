/* store.js —— 客户端单一状态源：snapshot 基线 + RFC6902 patch 增量 + 对象级 rv 缓存。
 * 数据纪律（同旧 console：零前端自建业务状态）——store 只承载服务端投影；
 * patch 应用失败（路径错位/版本跳跃）=保守回退信号：返回 false，调用方整段
 * re-fetch /v1/snapshot 后 init()（不猜不补——诚实回退）。
 * docCache：key='Kind|scope|name' value={rv, content}——快照 objects refs 携带
 * resourceVersion，rv 未变即复用缓存正文（WS 与降级轮询两条路径同样生效）。
 * 订阅通知=批处理合帧（microtask——一拍多 patch 只触发一次重渲染）。
 * 纯逻辑零 DOM——node 单测机械直测口径。 */

import fastJsonPatch from 'fast-json-patch';

// CJS interop：applyPatch 挂在 default 导出（cjs-module-lexer 无法静态识别具名导出）。
const { applyPatch } = fastJsonPatch;

/** 对象缓存键（与旧 console-obj 同形：'Kind|scope|name'）。 */
export function objKey(kind, scope, name) {
  return kind + '|' + (scope || 'global') + '|' + (name || '');
}

/** 对象 ref 归一化：兼容 objects={Kind:[…]}（按 kind 桶）与 objects=[{kind,…}]（平铺）两种快照形态。 */
function normalizeRef(x, kindHint) {
  if (x == null) return null;
  if (typeof x === 'string') {
    return kindHint ? { kind: String(kindHint), scope: 'global', name: x, rv: 0, raw: x } : null;
  }
  if (typeof x !== 'object') return null;
  const kind = x.kind || kindHint;
  const name = x.name;
  if (!kind || !name) return null;
  const rvRaw = x.resourceVersion != null ? x.resourceVersion : x.rv;
  return { kind: String(kind), scope: String(x.scope || 'global'), name: String(name), rv: Number(rvRaw) || 0, raw: x };
}

function arrOf(v) { return Array.isArray(v) ? v : []; }

/** 基线深拷贝（快照=服务端权威投影——init 后前端只经 patch 改写克隆体，不回写入参；
 * structuredClone 不可用（极旧运行环境）=JSON 往返降级）。 */
function cloneSnapshot(v) {
  try {
    return typeof structuredClone === 'function' ? structuredClone(v) : JSON.parse(JSON.stringify(v));
  } catch {
    return v;
  }
}

export function createStore() {
  let state = null;          // snapshot 基线（服务端权威投影——前端不改写）
  let seq = 0;               // 最近应用的 patch 序号（断点续传判据——缺位=服务端未携带）
  const subs = new Set();    // 订阅者（当前屏重渲染）
  let scheduled = false;     // 合帧标记
  const docCache = new Map(); // 对象正文 rv 缓存

  function notify() {
    if (scheduled) return;
    scheduled = true;
    queueMicrotask(() => {
      scheduled = false;
      subs.forEach((fn) => { try { fn(state); } catch (e) { console.error('[store] 订阅者失败', e); } });
    });
  }

  const api = {
    /** 快照是否就绪（未 init 前=骨架词形）。 */
    get isReady() { return !!state; },
    /** 建基线（整段快照——首连/patch 失败回退/降级轮询共用同一入口）。 */
    init(snapshot) {
      state = snapshot && typeof snapshot === 'object' ? cloneSnapshot(snapshot) : {};
      if (Number.isFinite(Number(state.seq))) seq = Number(state.seq);
      notify();
      return api;
    },
    /**
     * 增量应用 RFC6902 patch（validate:false 保守应用）。
     * 返回 true=已应用并通知；false=应用失败（状态保持原样——调用方整段 re-fetch）。
     */
    applyPatch(patch, meta) {
      if (!state) return false;
      if (!Array.isArray(patch) || !patch.length) return true; // 空 patch=无变化（成功语义）
      let ok = true;
      try {
        const res = applyPatch(state, patch, false, true);
        if (res && res.newDocument) state = res.newDocument;
      } catch (e) {
        ok = false;
        console.warn('[store] patch 应用失败——回退整段快照：', String((e && e.message) || e));
      }
      if (!ok) return false;
      if (meta && Number.isFinite(Number(meta.seq))) seq = Number(meta.seq);
      notify();
      return true;
    },
    /** 最近应用的 patch 序号。 */
    get seq() { return seq; },
    /** 连接态（connecting/connected/reconnecting/degraded——ws.js 广播，随状态帧通知）。 */
    setConn(s) { if (conn !== s) { conn = s; notify(); } },
    get conn() { return conn; },
    /** 点路径读（'a.b.0.c'——缺路径=undefined 不抛）。 */
    get(path) {
      if (!state) return undefined;
      return String(path || '').split('.').reduce((o, k) => (o == null ? undefined : o[k]), state);
    },
    /** 订阅状态帧（返回退订函数）。 */
    subscribe(fn) { subs.add(fn); return () => subs.delete(fn); },

    /* ---------- 容器化访问器（快照字段缺席=空数组/空值，不抛不猜） ---------- */
    objects() {
      const o = state && state.objects;
      if (!o) return [];
      if (Array.isArray(o)) return o.map((x) => normalizeRef(x)).filter(Boolean);
      return Object.keys(o).flatMap((k) => arrOf(o[k]).map((x) => normalizeRef(x, k))).filter(Boolean);
    },
    refs(kind) { return api.objects().filter((x) => x.kind === kind); },
    kinds() { return arrOf(state && state.kinds); },
    events() { return arrOf(state && state.events); },
    drafts() { return arrOf(state && state.drafts); },
    tasks() {
      const t = state && state.tasks;
      if (Array.isArray(t)) return t;
      return arrOf(t && t.tasks);
    },
    runs() {
      const r = state && (state.runs !== undefined ? state.runs : (state.workflowRuns));
      return Array.isArray(r) ? r : arrOf(r && r.runs);
    },
    workflows() {
      const w = state && state.workflows;
      return Array.isArray(w) ? w : arrOf(w && w.workflows);
    },
    proposals() {
      const p = state && state.proposals;
      return Array.isArray(p) ? p : arrOf(p && p.proposals);
    },
    sessions() {
      const s = state && state.sessions;
      return Array.isArray(s) ? s : arrOf(s && s.sessions);
    },
    attention() {
      const a = state && state.attention;
      if (!a) return [];
      if (Array.isArray(a)) return a;
      return arrOf(a.items || a.cards);
    },
    inbox() { return (state && state.inbox) || null; },

    /* ---------- 对象级 rv 缓存（正文懒取的家） ---------- */
    cacheDoc(kind, scope, name, rv, content) {
      docCache.set(objKey(kind, scope, name), {
        rv: Number(rv) || 0,
        content: content == null ? '' : String(content),
        at: Date.now(),
      });
    },
    /** 命中判据：rv 未变才返回缓存（rv 变了=过期不返回——调用方现读）。 */
    cachedDoc(kind, scope, name, rv) {
      const c = docCache.get(objKey(kind, scope, name));
      if (!c) return null;
      const want = Number(rv) || 0;
      if (want > 0 && c.rv !== want) return null;
      return c;
    },
    /** 快照中 rv 与缓存不符（或未缓存）的对象 refs——首连/降级轮询后正文预取清单。 */
    staleDocs(limit) {
      const cap = Number(limit) || 24;
      const out = [];
      for (const r of api.objects()) {
        if (!r.kind || !r.name) continue;
        const c = docCache.get(objKey(r.kind, r.scope, r.name));
        if (!c || c.rv !== r.rv) {
          out.push(r);
          if (out.length >= cap) break;
        }
      }
      return out;
    },
    clearDocCache() { docCache.clear(); },
  };

  let conn = 'connecting';
  return api;
}
