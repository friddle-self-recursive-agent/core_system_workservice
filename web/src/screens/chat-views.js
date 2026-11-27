/* chat-views.js —— 转录节点 → HTML 词形（2026-11-27 D 批：chat.js 行数红线拆分）。
 * 上游词形依据同 chat-nodes.js（chat-ds.md 对账表）：触发行卡=D1/注入行=D2/轮汇总=D3/
 * 待办=D4/工具五态+专名+meta=D5-7/用量 pill=D10/思考对齐=D12/时钟=D13。
 * 纯函数零 DOM（node 单测机械直测）；一切动态文本过 esc——零原文直插。 */

import { esc, trunc, fmtNum } from '../core/format.js';
import { linkifyRefs } from '../render/refs.js';
import { renderMarkdown } from '../render/md.js';
import { t as tr } from '../core/i18n.js';
import { nodeKey, turnTailOf, todoModel, usageTotal, usageDetail, formatClock } from './chat-nodes.js';

/* 触发/注入行卡图标（上游 TRIGGER_ICONS 词族——最小内联线形 SVG 14px，stroke 随 CSS）。 */
const N_ICONS = {
  job: '<rect x="2" y="3.5" width="12" height="2.2" rx="1.1"/><rect x="2" y="7.4" width="8.5" height="2.2" rx="1.1"/><rect x="2" y="11.3" width="5.5" height="2.2" rx="1.1"/>',
  plugin: '<rect x="3" y="3" width="10" height="10" rx="2.5"/><path d="M8 6.2v3.6M6.2 8h3.6"/>',
  schedule: '<circle cx="8" cy="8" r="5.5"/><path d="M8 5.2V8l2 1.6"/>',
  subagent: '<circle cx="5.6" cy="5.8" r="2.1"/><path d="M2.4 13c.5-2.4 1.8-3.4 3.2-3.4s2.7 1 3.2 3.4"/><circle cx="11" cy="6.6" r="1.7"/><path d="M10.4 9.7c1.6.4 2.5 1.5 2.9 3.3"/>',
  goal: '<circle cx="8" cy="8" r="5.5"/><circle cx="8" cy="8" r="2.2"/>',
  agent: '<path d="M2.5 8 13.5 3 11 13 8 9.5 2.5 8z"/><path d="M8 9.5 13.5 3"/>',
  team: '<circle cx="5.6" cy="5.8" r="2.1"/><path d="M2.4 13c.5-2.4 1.8-3.4 3.2-3.4s2.7 1 3.2 3.4"/><circle cx="11" cy="6.6" r="1.7"/><path d="M10.4 9.7c1.6.4 2.5 1.5 2.9 3.3"/>',
  github: '<circle cx="4.5" cy="4" r="1.7"/><circle cx="4.5" cy="12" r="1.7"/><circle cx="11.5" cy="8" r="1.7"/><path d="M4.5 5.7v4.6M6.2 8h3.6"/>',
  webhook: '<circle cx="8" cy="8" r="5.5"/><path d="M2.5 8h11M8 2.5c2.6 2.6 2.6 8.4 0 11-2.6-2.6-2.6-8.4 0-11"/>',
  retry: '<path d="M3 8a5 5 0 1 1 1.5 3.5"/><path d="M3 12.5V8.5h4"/>',
  inject: '<rect x="2.5" y="5.5" width="11" height="8" rx="2"/><path d="M8 2v5M6 5l2 2 2-2"/>',
  recall: '<path d="M5 2.5h6a1 1 0 0 1 1 1V14l-4-2.6L4 14V3.5a1 1 0 0 1 1-1z"/>',
};

function noticeIcon(id) {
  return '<svg class="chat-n-ico" viewBox="0 0 16 16" aria-hidden="true">' + (N_ICONS[id] || N_ICONS.inject) + '</svg>';
}

function clockAttr(n) {
  const c = formatClock(n.time);
  return c ? ' title="' + esc(c) + '"' : '';
}

/** 触发行卡/注入行（D1/D2）：details 折叠——head=图标+标题+·+producer+·+摘要+右时间+chevron；
 * 展开体=轮触发说明行（waking）+全文。 */
export function noticeHtml(n, names) {
  const head = '<summary>' + noticeIcon(n.trigger && n.trigger.id) +
    '<span class="chat-n-title">' + esc(n.title) + '</span>' +
    (n.producer ? '<span class="chat-n-sep">·</span><span class="chat-n-producer mono">' + esc(n.producer) + '</span>' : '') +
    (n.summary ? '<span class="chat-n-sep">·</span><span class="chat-n-summary">' + esc(n.summary) + '</span>' : '') +
    (n.time ? '<time class="chat-n-time">' + esc(formatClock(n.time)) + '</time>' : '') +
    '<span class="ex-chev">▸</span></summary>';
  const body = '<div class="ex-body">' +
    (n.waking ? '<div class="chat-n-exp">' + esc(tr('这条通知触发了本轮回复。')) + '</div>' : '') +
    '<div class="chat-n-text">' + (n.text ? linkifyRefs(n.text, names) : '<span class="faint">…</span>') + '</div></div>';
  return '<details class="chat-notice' + (n.waking ? ' chat-trigger' : '') + '" data-node="' + esc(nodeKey(n)) + '"' + clockAttr(n) + '>' +
    head + body + '</details>';
}

/** 思考折叠（D12 对齐：标题「思考」+ 收起首行预览剥 **；展开=reasoning 全文）。 */
export function reasoningFold(reasoning, names) {
  const preview = String(reasoning || '').split('\n')
    .map((l) => l.replace(/\*\*/g, ''))
    .find((l) => l.trim()) || '';
  return '<details class="chat-fold"><summary><span class="ex-chev">▸</span>' + esc(tr('思考')) +
    (preview ? '<span class="chat-think-preview">' + esc(trunc(preview.trim(), 72)) + '</span>' : '') + '</summary>' +
    '<div class="ex-body">' + linkifyRefs(reasoning, names) + '</div></details>';
}

/** 用量 pill（D10）：>0 才渲染；title=明细（缓存命中族词形）。 */
export function usagePill(usage) {
  const total = usageTotal(usage);
  if (!total) return '';
  const detail = usageDetail(usage);
  return '<span class="chat-usage" title="' + esc(detail || tr('本轮用量')) + '">' +
    esc(tr('用量 {total}').replace('{total}', fmtNum(total))) + '</span>';
}

/** 工具行（D5/D6/D7）：details 五态 data-state + 专名标题|已调用 X + 关键参数摘要 + meta 计数
 * + 状态词（运行中 shimmer/失败+退出码徽标/已停止）+ 展开体（meta.lines 行号预览+args+结果）。 */
export function toolHtml(n) {
  const st = n.state || 'ok';
  let stateHtml = '';
  if (st === 'running') stateHtml = '<span class="chat-tool-run">' + esc(tr('运行中')) + '</span>';
  else if (st === 'error') {
    stateHtml = '<span class="chat-tool-err">' + esc(tr('失败')) + '</span>';
    const code = Number(n.meta && n.meta.exitCode);
    if (n.meta && Number.isFinite(code) && code !== 0) {
      stateHtml += '<span class="chat-tool-exit">' + esc(tr('退出码 {code}').replace('{code}', String(n.meta.exitCode))) + '</span>';
    } else if (n.meta && typeof n.meta.signal === 'string' && n.meta.signal) {
      stateHtml += '<span class="chat-tool-exit">' + esc(tr('信号 {signal}').replace('{signal}', n.meta.signal)) + '</span>';
    }
  } else if (st === 'stopped') stateHtml = '<span class="chat-tool-stop">' + esc(tr('已停止')) + '</span>';
  const head = '<summary><span class="ex-chev">▸</span>' +
    (n.title
      ? '<span class="chat-tool-title">' + esc(n.title) + '</span>'
      : '<span>' + esc(tr('已调用')) + '</span> <span class="chat-tool-name">' + esc(n.name) + '</span>') +
    (n.summary ? '<span class="chat-tool-arg">' + esc(n.summary) + '</span>' : '') +
    (n.metaLine ? '<span class="chat-tool-metaline">' + esc(n.metaLine) + '</span>' : '') +
    ' ' + stateHtml + '</summary>';
  const linesPrev = (n.meta && Array.isArray(n.meta.lines) && n.meta.lines.length)
    ? '<div class="chat-tool-lines">' + n.meta.lines.slice(0, 24).map((l) =>
      '<span class="chat-tool-ln">' + esc(String((l && l.number) != null ? l.number : '')) + '</span>' +
      '<span class="chat-tool-lt">' + esc(String((l && l.text) != null ? l.text : '')) + '</span>').join('') + '</div>'
    : '';
  const errLine = (n.error && (n.error.name || n.error.code))
    ? '\n' + tr('失败原因：') + [n.error.name, n.error.code].filter(Boolean).join('/')
    : '';
  const detail = linesPrev +
    '<div class="chat-tool-io">' + esc(trunc(String(n.args || ''), 800)) +
    (n.result ? '\n' + tr('──── 结果 ────') + '\n' + esc(trunc(String(n.result), 1200)) : '') +
    errLine + '</div>';
  return '<details class="chat-fold chat-tool" data-node="' + esc(nodeKey(n)) + '" data-state="' + esc(st) + '"' + clockAttr(n) + '>' +
    head + '<div class="ex-body">' + detail + '</div></details>';
}

/** 轮汇总行（D3）：状态词+右时间；max-tokens 附 hint。 */
export function turnTailHtml(n) {
  const tt = turnTailOf(n.reason, n.durationMs);
  return '<div class="chat-turntail-wrap" data-node="' + esc(nodeKey(n)) + '"' + clockAttr(n) + '>' +
    '<div class="chat-turntail ' + esc(tt.cls) + '"><span class="chat-tt-dot"></span>' +
    '<span class="chat-tt-word">' + esc(tt.word) + '</span>' +
    (n.time ? '<time class="chat-n-time">' + esc(formatClock(n.time)) + '</time>' : '') + '</div>' +
    (tt.hint ? '<div class="chat-tt-hint">' + esc(tt.hint) + '</div>' : '') + '</div>';
}

/** 待办清单行（D4）：{done}/{total} 已完成 + 首个进行中项；展开=三态点清单。 */
export function todoHtml(n) {
  const m = todoModel(n.todos);
  if (!m.total) return '';
  const dot = (s) => '<span class="chat-todo-dot ' + (s === 'completed' ? 'done' : s === 'in_progress' ? 'act' : 'idle') + '"></span>';
  return '<details class="chat-fold chat-todo" data-node="' + esc(nodeKey(n)) + '"' + clockAttr(n) + '>' +
    '<summary><span class="ex-chev">▸</span>' + esc(tr('更新任务清单')) + ' ' +
    '<span class="chat-todo-count">' + esc(tr('{done}/{total} 已完成').replace('{done}', String(m.done)).replace('{total}', String(m.total))) + '</span>' +
    (m.active ? '<span class="chat-todo-active">' + esc(trunc(m.active, 44)) + '</span>' : '') + '</summary>' +
    '<div class="ex-body chat-todo-list">' + m.rows.map((r) =>
      '<div class="chat-todo-row">' + dot(r.status) + '<span>' + esc(r.content) + '</span></div>').join('') + '</div></details>';
}

/** 重试行（D8）：诚实最小词形——attempt/reason 字段缺席=只显示标题。 */
export function retryHtml(n) {
  const d = (n.data && typeof n.data === 'object') ? n.data : {};
  const fail = (d.failure && typeof d.failure === 'object') ? String(d.failure.message || '') : String(d.message || '');
  const label = (typeof d.label === 'string' && d.label) ? d.label : tr('正在重试模型请求');
  const extra = [];
  if (Number.isFinite(Number(d.attempt)) && Number.isFinite(Number(d.maximum))) {
    extra.push(tr('{label}（{retry}/{maximum}）').replace('{label}', label).replace('{retry}', String(d.attempt)).replace('{maximum}', String(d.maximum)));
    if (Number.isFinite(Number(d.seconds))) extra.push(String(d.seconds) + 's');
  } else extra.push(label);
  return '<div class="chat-sysrow chat-retry" data-node="' + esc(nodeKey(n)) + '"' + clockAttr(n) + '>' +
    esc(extra.join(' · ')) + (fail ? ' · ' + esc(tr('失败原因：') + trunc(fail, 120)) : '') + '</div>';
}

/** 节点 → HTML（chat.js nodeHtml v5 权威）。 */
export function nodeHtml(n, names) {
  if (n.kind === 'user') {
    const pending = n.pending
      ? '<div class="chat-pending">' + esc(tr('送达中…')) + '</div>'
      : (n.injected ? '<div class="chat-pending">' + esc(tr('注入上下文')) + '</div>' : '');
    const err = n.error
      ? '<div class="chat-err"><span>' + esc(tr('送达失败（') + String(n.error) + tr('）')) + '</span>' +
        '<button class="btn btn-ghost sm" data-cc-retry="' + esc(String(n.requestId || '')) + '">' + esc(tr('重试')) + '</button></div>'
      : '';
    return '<div class="chat-msg-user" data-node="' + esc(nodeKey(n)) + '"' + clockAttr(n) + '>' +
      '<div class="chat-bubble">' + (n.text ? renderMarkdown(n.text) : '<span class="faint">…</span>') + '</div>' +
      (pending || err ? '<div>' + pending + err + '</div>' : '') +
      '</div>';
  }
  if (n.kind === 'notice') return noticeHtml(n, names);
  if (n.kind === 'assistant') {
    // 无正文助手节点不渲染（工具轮次已在 mapTranscript 过滤，这里是纯空流防御）。
    if (!n.text && !n.reasoning) return '';
    const reasoning = n.reasoning ? reasoningFold(n.reasoning, names) : '';
    const body = n.text ? '<div class="chat-md">' + renderMarkdown(n.text) + '</div>' : '';
    return '<div class="chat-msg-ai" data-node="' + esc(nodeKey(n)) + '"' + clockAttr(n) + '>' + reasoning + body + usagePill(n.usage) + '</div>';
  }
  if (n.kind === 'tool') return toolHtml(n);
  if (n.kind === 'turntail') return turnTailHtml(n);
  if (n.kind === 'todo') return todoHtml(n);
  if (n.kind === 'retry') return retryHtml(n);
  const label = (n.dev ? tr('开发者') + ' · ' : '') + trunc(String(n.text || ''), 200);
  return '<div class="chat-sysrow" data-node="' + esc(nodeKey(n)) + '"' + clockAttr(n) + '>' + esc(label) + '</div>';
}
