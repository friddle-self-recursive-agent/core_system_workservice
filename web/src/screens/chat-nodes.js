/* chat-nodes.js —— 转录映射纯函数 v5（2026-11-27 D 批：DeepSeek harness 展示类型对账）。
 * 上游权威=deepseek-ai/deepseek-harness：事件词表 @deepseek-ai/dsh-session types.d.ts
 * （SessionEventMap）；语义词形 @deepseek-ai/dsh-llm message.d.ts（MessageSourceMap +
 * ContextForm：instructions/catalog/snapshot/notice≤120/relay/recall）；文案=ui-chat
 * src/client/locale.ts zh 段；组件词形=ui-chat conversation-nodes/* + ui-tool toolviews/*。
 * 对账表=core_system_instance/acceptance/console/chat-ds.md（D1..D13 逐条带依据）。
 * 零 DOM——node 单测机械直测；输出节点={kind,...} 供 chat.js nodeHtml 渲染。 */

import { trunc } from '../core/format.js';
import { t as tr } from '../core/i18n.js';

/** content 块 → {text, reasoning, toolCalls[]}（user/assistant 两族共用块词形）。 */
export function blocksOf(content) {
  const out = { text: '', reasoning: '', toolCalls: [] };
  if (!Array.isArray(content)) return out;
  for (const b of content) {
    if (!b || typeof b !== 'object') continue;
    if (b.type === 'text' && typeof b.text === 'string') out.text += (out.text ? '\n' : '') + b.text;
    else if (b.type === 'reasoning' && typeof b.text === 'string') out.reasoning += (out.reasoning ? '\n' : '') + b.text;
    else if (b.type === 'tool-call') out.toolCalls.push({ id: String(b.id || b.callId || ''), name: String(b.name || ''), argsRaw: typeof b.arguments === 'string' ? b.arguments : JSON.stringify(b.arguments ?? b.args ?? '') });
  }
  return out;
}

/** 节点渲染去重键（seq+kind——同 seq 双族不互撞；chat.js/chat-views 共用）。 */
export function nodeKey(n) {
  return String(Number(n && n.seq) || 0) + ':' + String((n && n.kind) || '');
}

/** tool/result 载荷 → 结果文本：上游词形 data.message.content 块（ToolResultBlock），
 * 兼容旧 data.content/data.output 直串词形——2026-11-27 实测「执行中…」卡死的病根。 */
export function resultTextOf(data) {
  const d = (data && typeof data === 'object') ? data : {};
  if (typeof d.content === 'string' && d.content) return d.content;
  const msg = (d.message && typeof d.message === 'object') ? d.message : null;
  if (msg && Array.isArray(msg.content)) {
    const text = msg.content.filter((b) => b && b.type === 'text' && typeof b.text === 'string').map((b) => b.text).join('\n');
    if (text) return text;
  }
  if (d.output != null) return typeof d.output === 'string' ? d.output : JSON.stringify(d.output);
  return '';
}

/* ---------- D1/D2：轮触发行卡 + 上下文注入行（source.kind → 词族） ---------- */

/** source → 词族 id（上游 turn-trigger.ts 词表 + ContextForm）。 */
export function triggerOf(source) {
  const s = (source && typeof source === 'object') ? source : {};
  const kind = String(s.kind || '');
  if (kind === 'tool-jobs') return { id: 'job' };
  if (kind === 'cordis-host-runner') return { id: 'plugin' };
  if (kind === 'schedule') return { id: 'schedule' };
  if (kind === 'subagent-settled') return { id: 'subagent', stopReason: String(s.stopReason || '') };
  if (kind === 'goal') return { id: 'goal' };
  if (kind === 'agent-message') return { id: 'agent' };
  if (kind === 'team-message') return { id: 'team' };
  if (kind === 'webhook') return String(s.provider || '') === 'github' ? { id: 'github' } : { id: 'webhook' };
  if (kind === 'retry') return { id: 'retry' };
  const form = String(s.form || '');
  if (form === 'recall' || kind === 'session-recall') return { id: 'recall' };
  return { id: 'inject', form };
}

const TRIGGER_TITLES = {
  job: () => tr('后台任务状态更新'),
  plugin: () => tr('插件状态更新'),
  schedule: () => tr('自动化任务'),
  goal: () => tr('继续执行目标'),
  agent: () => tr('收到任务消息'),
  team: () => tr('收到团队消息'),
  github: () => tr('收到 GitHub 事件'),
  webhook: () => tr('收到外部事件'),
  retry: () => tr('已请求继续生成'),
  inject: () => tr('上下文注入'),
  recall: () => tr('跨会话召回'),
};

const SUBAGENT_TITLES = {
  completed: () => tr('子智能体已完成任务'),
  aborted: () => tr('子智能体任务已停止'),
  error: () => tr('子智能体执行失败'),
  'max-tokens': () => tr('子智能体已达 token 上限'),
  refusal: () => tr('子智能体拒绝执行'),
};

/** 触发/注入行标题（D1/D2；缺词族=通用注入行不冒充身份）。 */
export function noticeTitle(tg) {
  if (tg.id === 'subagent') return (SUBAGENT_TITLES[tg.stopReason] || (() => tr('子任务状态更新')))();
  return (TRIGGER_TITLES[tg.id] || TRIGGER_TITLES.inject)();
}

/** notice 摘要（上游契约：form=notice 携 ≤120 字 summary——boundContextSummary 同族：
 * 超界=截断+省略号，总长恒 ≤120；其余 form 字段不可读=无摘要（opaque body，documented default）。 */
export function noticeSummary(source) {
  const s = (source && typeof source === 'object') ? source : {};
  if (String(s.form || '') === 'notice' && typeof s.summary === 'string' && s.summary) {
    const t = String(s.summary);
    return t.length > 120 ? t.slice(0, 119) + '…' : t;
  }
  return '';
}

/** producer 标签（ContextInjectionRow 的 producer.label 词形——来源可读才显示）。 */
export function producerLabel(source) {
  const s = (source && typeof source === 'object') ? source : {};
  if (typeof s.plugin === 'string' && s.plugin) return s.plugin;
  const kind = String(s.kind || '');
  return kind && kind !== 'user' ? kind : '';
}

/* ---------- D3：轮汇总（已完成，用时 X）+ D13 时钟 ---------- */

/** 时长分段（上游 formatRunDuration：时/分/秒，分钟从 60s 起；<1s=''——调用方落「已完成」）。 */
export function formatRunDuration(ms) {
  const total = Math.floor(Number(ms) || 0);
  if (total < 1000) return '';
  const h = Math.floor(total / 3600000);
  const m = Math.floor((total % 3600000) / 60000);
  const s = Math.floor((total % 60000) / 1000);
  const parts = [];
  if (h) parts.push(h + tr('小时'));
  if (m) parts.push(m + tr('分'));
  parts.push(s + tr('秒'));
  return parts.join('');
}

const pad2 = (n) => String(n).padStart(2, '0');

/** 消息时钟（上游 formatMessageClock：当天 HH:MM；同年 M月D日 HH:MM；跨年 Y年M月D日 HH:MM）。 */
export function formatClock(time, now) {
  const d = new Date(Number(time) || 0);
  const n = new Date(Number(now) || Date.now());
  if (!Number(time) || isNaN(d.getTime())) return '';
  const clock = pad2(d.getHours()) + ':' + pad2(d.getMinutes());
  if (d.getFullYear() === n.getFullYear() && d.getMonth() === n.getMonth() && d.getDate() === n.getDate()) return clock;
  const md = d.getFullYear() === n.getFullYear()
    ? tr('{m}月{d}日').replace('{m}', String(d.getMonth() + 1)).replace('{d}', String(d.getDate()))
    : tr('{y}年{m}月{d}日').replace('{y}', String(d.getFullYear())).replace('{m}', String(d.getMonth() + 1)).replace('{d}', String(d.getDate()));
  return md + ' ' + clock;
}

/** turn/end → 汇总词形 {word, duration, hint, cls}（上游 TurnProcessNodeView）。 */
export function turnTailOf(reason, durationMs) {
  const r = String(reason || '');
  if (r === 'interrupted') return { word: tr('已停止'), duration: '', hint: '', cls: 'stopped' };
  if (r === 'max-tokens') return { word: tr('已达到输出 token 上限'), duration: '', hint: tr('回答被截断，已有输出保留在对话中。发送「继续」可让模型接着输出。'), cls: 'maxtok' };
  if (r && r !== 'complete' && r !== 'completed' && r !== 'end_turn' && r !== 'stop-sequence') {
    return { word: tr('处理失败'), duration: '', hint: '', cls: 'failed' };
  }
  const dur = formatRunDuration(durationMs);
  return { word: dur ? tr('已完成，用时 ') + dur : tr('已完成'), duration: dur, hint: '', cls: 'ok' };
}

/* ---------- D4：待办清单 ---------- */

/** todo/write → {done, total, active}（上游 todo-row：{done}/{total} 已完成 + 首个进行中项）。 */
export function todoModel(todos) {
  const rows = (Array.isArray(todos) ? todos : []).filter((x) => x && typeof x.content === 'string');
  const done = rows.filter((x) => x.status === 'completed').length;
  const active = rows.find((x) => x.status === 'in_progress');
  return { rows, done, total: rows.length, active: active ? active.content : '' };
}

/* ---------- D5/D6/D7：工具行五态 + 专名标题 + 关键参数摘要 + meta 计数 ---------- */

const TOOL_TITLES = {
  bash: () => tr('运行命令'), pwsh: () => tr('运行命令'),
  read: () => tr('读取'), read_image: () => tr('读取图片'),
  write: () => tr('写入'), edit: () => tr('编辑'), apply_patch: () => tr('编辑'),
  grep: () => tr('搜索文件内容'), glob: () => tr('查找文件'),
  web_search: () => tr('网页搜索'), web_fetch: () => tr('网页获取'),
  todo_write: () => tr('更新任务清单'), ask_user_question: () => tr('提问'),
};

/** 工具专名标题（D6；未知=''——渲染层落「已调用 {name}」沿用词形）。 */
export function toolTitle(name) {
  const fn = TOOL_TITLES[String(name || '').toLowerCase()];
  return fn ? fn() : '';
}

const SUMMARY_KEYS = ['file_path', 'path', 'command', 'url', 'query', 'pattern', 'note'];

/** 关键参数摘要（D6：折叠行显示第一关键参数而非裸 JSON；完整 JSON 保留在展开体）。 */
export function toolSummary(name, argsRaw) {
  let obj = null;
  try { obj = JSON.parse(String(argsRaw || '')); } catch { /* 非 JSON=原样首行 */ }
  if (obj && typeof obj === 'object' && !Array.isArray(obj)) {
    for (const k of SUMMARY_KEYS) {
      const v = obj[k];
      if (typeof v === 'string' && v.trim()) return trunc(v.trim(), 90);
    }
  }
  const s = String(argsRaw || '').trim();
  return s ? trunc(s.split('\n')[0], 90) : '';
}

/** 工具行状态（D5 五态——上游 tool-call-model：running/error/stopped/ok；preparing 不入账）。 */
export function toolStateOf(hasResult, errorObj, meta, name) {
  if (!hasResult) return 'running';
  if (meta && typeof meta === 'object' && meta.stopped === true) return 'stopped';
  if (errorObj) return 'error';
  const nm = String(name || '').toLowerCase();
  if ((nm === 'bash' || nm === 'pwsh') && meta && typeof meta === 'object') {
    const code = Number(meta.exitCode);
    if (Number.isFinite(code) && code !== 0) return 'error';
    if (typeof meta.signal === 'string' && meta.signal) return 'error';
  }
  return 'ok';
}

/** meta 计数摘要（D7 分级：search/read/diff/web 四族——字段缺席=空串不冒充）。 */
export function metaSummary(name, meta) {
  if (!meta || typeof meta !== 'object') return '';
  const nm = String(name || '').toLowerCase();
  if (nm === 'grep' || nm === 'glob') {
    if (meta.shape === 'paths' && Number.isFinite(Number(meta.total))) {
      return tr('{shown} 个路径').replace('{shown}', String(meta.truncated != null ? Number(meta.shown ?? meta.total) : Number(meta.total)));
    }
    if (Number.isFinite(Number(meta.total))) {
      if (Number(meta.total) === 0) return tr('无结果');
      const files = Number.isFinite(Number(meta.files)) ? Number(meta.files) : null;
      return tr('{shown} 处匹配 · {files} 个文件')
        .replace('{shown}', String(Number(meta.shown ?? meta.total)))
        .replace('{files}', files != null ? String(files) : '—');
    }
    return '';
  }
  if (nm === 'read' && Array.isArray(meta.lines) && meta.lines.length) {
    return tr('{n} 行').replace('{n}', String(meta.lines.length));
  }
  if ((nm === 'edit' || nm === 'apply_patch' || nm === 'write') && Array.isArray(meta.diffs)) {
    return meta.diffs.length
      ? tr('{n} 个文件变更').replace('{n}', String(meta.diffs.length))
      : '';
  }
  if (nm === 'web_fetch' && Number.isFinite(Number(meta.statusCode))) {
    return 'HTTP ' + String(meta.statusCode);
  }
  return '';
}

/* ---------- D10：用量 ---------- */

/** usage → 总 token 数（上游 TokenUsage 四字段求和；缺席=0）。 */
export function usageTotal(usage) {
  const u = (usage && typeof usage === 'object') ? usage : {};
  return ['inputTokens', 'outputTokens', 'cacheReadTokens', 'cacheWriteTokens']
    .reduce((sum, k) => sum + (Number.isFinite(Number(u[k])) ? Number(u[k]) : 0), 0);
}

/** usage → title 明细行（D10 词形：缓存命中/未缓存输入/缓存读取/缓存写入/输出/其中推理）。 */
export function usageDetail(usage) {
  const u = (usage && typeof usage === 'object') ? usage : {};
  const lines = [];
  if (Number.isFinite(Number(u.inputTokens))) lines.push(tr('未缓存输入') + ' ' + u.inputTokens);
  if (Number.isFinite(Number(u.cacheReadTokens))) lines.push(tr('缓存读取') + ' ' + u.cacheReadTokens);
  if (Number.isFinite(Number(u.cacheWriteTokens))) lines.push(tr('缓存写入') + ' ' + u.cacheWriteTokens);
  if (Number.isFinite(Number(u.outputTokens))) lines.push(tr('输出') + ' ' + u.outputTokens);
  if (Number.isFinite(Number(u.reasoningTokens))) lines.push(tr('（其中推理 {tokens}）').replace('{tokens}', String(u.reasoningTokens)));
  return lines.join(' · ');
}

/* ---------- 映射主函数 ---------- */

/** journal 事件 → 渲染节点族 v5（seq 升序；工具组按首次 call 顺序；callId 去重）。
 * 词族：user/notice(触发+注入)/assistant(含 usage)/tool(五态)/turntail/todo/retry/system。 */
export function mapTranscript(events) {
  const nodes = [];
  const tools = new Map(); // callId → {seq, time, name, args, result, error, meta}
  const order = [];
  const rows = Array.isArray(events) ? [...events] : [];
  rows.sort((a, b) => (Number(a && a.seq) || 0) - (Number(b && b.seq) || 0));
  const turnStarts = new Map(); // turn → time（D3 用时）
  let emittedSinceTurnStart = 0; // 轮首判别：turn/start 后尚未出节点→该注入词=轮触发
  let lastBoundaryTurnStart = false; // 最近边界词=turn/start 且其后零出账（无 turn/start 证据≠轮触发）
  for (const ev of rows) {
    if (!ev || typeof ev !== 'object') continue;
    const type = String(ev.type || '');
    const data = (ev.data && typeof ev.data === 'object') ? ev.data : {};
    const seq = Number(ev.seq) || 0;
    const time = Number(ev.time) || 0;
    if (type === 'turn/start') {
      turnStarts.set(Number(data.turn) || 0, time);
      emittedSinceTurnStart = 0;
      lastBoundaryTurnStart = true;
      continue;
    }
    if (type === 'user/message') {
      const src = (data.source && typeof data.source === 'object') ? data.source : null;
      const kind = String((src && src.kind) || '');
      if (kind === 'compact-checkpoint') continue; // 压缩检查点=替换面事件，不渲染
      const b = blocksOf(data.content);
      if (kind !== 'user') {
        const tg = triggerOf(src);
        const waking = lastBoundaryTurnStart && emittedSinceTurnStart === 0;
        nodes.push({
          kind: 'notice', seq, time, text: b.text, waking,
          trigger: tg, title: noticeTitle(tg),
          summary: noticeSummary(src), producer: producerLabel(src),
        });
        emittedSinceTurnStart += 1;
        lastBoundaryTurnStart = false;
        continue;
      }
      nodes.push({ kind: 'user', seq, time, text: b.text, injected: false });
      emittedSinceTurnStart += 1;
      lastBoundaryTurnStart = false;
    } else if (type === 'assistant/message') {
      const b = blocksOf(data.content);
      for (const tc of b.toolCalls) order.push(tc.id);
      // 工具轮次（只有 tool-call、无 text/reasoning）不生成助手节点——工具本身渲染
      if (!b.text && !b.reasoning && b.toolCalls.length) { emittedSinceTurnStart += 1; lastBoundaryTurnStart = false; continue; }
      nodes.push({ kind: 'assistant', seq, time, text: b.text, reasoning: b.reasoning, usage: data.usage ?? null });
      emittedSinceTurnStart += 1;
      lastBoundaryTurnStart = false;
    } else if (type === 'tool/call') {
      const id = String(data.callId || data.id || '');
      if (!tools.has(id)) tools.set(id, { seq, time, name: String(data.name || ''), args: typeof data.arguments === 'string' ? data.arguments : JSON.stringify(data.arguments ?? data.args ?? ''), result: '', error: null, meta: null });
      order.push(id);
      emittedSinceTurnStart += 1;
      lastBoundaryTurnStart = false;
    } else if (type === 'tool/result') {
      const id = String(data.callId || data.id || '');
      const t = tools.get(id);
      if (t) {
        t.result = resultTextOf(data);
        if (data.error && typeof data.error === 'object') t.error = { name: String(data.error.name || ''), code: String(data.error.code || '') };
        if (data.meta && typeof data.meta === 'object') t.meta = data.meta;
        if (!t.error && data.message && typeof data.message === 'object') {
          const bad = Array.isArray(data.message.content)
            && data.message.content.some((b) => b && typeof b === 'object' && b.isError === true);
          if (bad) t.error = { name: '', code: '' };
        }
      }
    } else if (type === 'turn/end') {
      const turn = Number(data.turn) || 0;
      const startedAt = turnStarts.get(turn) || 0;
      nodes.push({ kind: 'turntail', seq, time, reason: String(data.reason ?? ''), durationMs: startedAt ? Math.max(0, time - startedAt) : 0 });
      emittedSinceTurnStart = 0;
      lastBoundaryTurnStart = false;
    } else if (type === 'todo/write') {
      nodes.push({ kind: 'todo', seq, time, todos: Array.isArray(data.todos) ? data.todos : [] });
      emittedSinceTurnStart += 1;
      lastBoundaryTurnStart = false;
    } else if (type === 'llm/retry') {
      nodes.push({ kind: 'retry', seq, time, data });
      emittedSinceTurnStart += 1;
      lastBoundaryTurnStart = false;
    } else if (type === 'system/message' || type === 'developer/message') {
      nodes.push({ kind: 'system', seq, time, text: String(data.text ?? data.content ?? ''), dev: type === 'developer/message' });
      emittedSinceTurnStart += 1;
      lastBoundaryTurnStart = false;
    }
  }
  // callId 去重：assistant 消息里的 tool-call block 与独立 tool/call 事件可能同时携带
  // 同一 callId，不去重会渲染两个「已调用 X」节点（2026-11-27 实测）。
  const seenToolIds = new Set();
  let lastTurnEndSeq = 0;
  for (const n of nodes) if (n.kind === 'turntail' && Number(n.seq) > lastTurnEndSeq) lastTurnEndSeq = Number(n.seq);
  for (const id of order) {
    if (seenToolIds.has(id)) continue;
    seenToolIds.add(id);
    const t = tools.get(id);
    if (!t) continue;
    // 中断合成（上游 tool.ts 同族）：turn/end 已过而调用未闭合=本轮中断遗留——已停止，不冒充运行中
    const interrupted = !t.result && lastTurnEndSeq > 0 && t.seq < lastTurnEndSeq;
    const state = interrupted ? 'stopped' : toolStateOf(!!t.result, t.error, t.meta, t.name);
    nodes.push({
      kind: 'tool', seq: t.seq, time: t.time, name: t.name || tr('工具'), args: t.args,
      result: t.result, error: t.error, meta: t.meta,
      state,
      title: toolTitle(t.name), summary: toolSummary(t.name, t.args),
      metaLine: metaSummary(t.name, t.meta),
    });
  }
  nodes.sort((a, b) => a.seq - b.seq);
  return nodes;
}
