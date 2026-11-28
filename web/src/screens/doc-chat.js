/* doc-chat.js —— 原地对话修改面板（2026-11-28 用户裁决：对象工作台「对话修改」不要跳转，
 * 直接嵌入相应流程进行修改——R-UX-10 修改主通路的承载面）。
 * 通路=chatFaces 真会话（create/send/transcript/state——同会话续跑：SESSIONS[refKey] 持锚）；
 * 闭环=首条消息自带当前 YAML+修改意图 → 助手回 ```yaml 块 → 提取 → diff 预览 → 人审 Apply
 * （rv 乐观锁 409=提示刷新）。docFix 契约插件侧缺席（诚实不依赖）。纯函数 node 直测。 */
import { esc, trunc, relAge } from '../core/format.js';
import { chip } from '../render/chips.js';
import { diffLines, renderDiff, hunksOf } from '../render/diff.js';
import { turnTailOf } from './chat-nodes.js';
import { htmlModal, toast } from './modal.js';
import { t as tr } from '../core/i18n.js';

/* 会话锚（内存态——同对象重开面板=同会话续跑；刷新页面重开=新会话，账在会话屏仍可查）。 */
const SESSIONS = new Map();

function faces() {
  return (typeof window !== 'undefined' && window.__CC && window.__CC.chatFaces) || null;
}

/** 助手回复 → 最后一个 ```yaml/``` 代码块（提取失败=空串由调用方诚实词形）。 */
export function extractYamlBlock(text) {
  const s = String(text || '');
  const blocks = [...s.matchAll(/```(?:yaml|yml)?\s*\n([\s\S]*?)```/g)].map((m) => m[1]);
  for (let i = blocks.length - 1; i >= 0; i--) {
    const b = blocks[i].trim();
    if (/^(apiVersion|kind|metadata):/m.test(b)) return b;
  }
  return '';
}

/** 默认首条修改指令（当前 YAML 全文+意图——要求整块回 yaml）。纯函数。 */
export function buildEditPrompt(ref, current, intent) {
  return tr('请修改对象 ') + (ref ? ref.kind + '/' + (ref.scope || 'global') + '/' + ref.name : '') +
    tr('（当前 YAML 声明为权威源，全文如下）。修改意图：') + String(intent || '') + '\n\n' +
    '```yaml\n' + String(current || '') + '\n```\n' +
    tr('输出要求：给出修改后的完整 YAML 声明，放在一个 ```yaml 代码块里；不改无关字段；解释放代码块外。');
}

/** 转录事件 → 面板行（纯函数）：user 气泡/assistant 文本（reasoning 折叠）/turn 尾。 */
export function transcriptRows(events) {
  const rows = [];
  (events || []).forEach((e) => {
    if (!e || !e.type) return;
    const d = e.data || {};
    if (e.type === 'user/message' && d.source && d.source.kind === 'user') {
      const text = (d.content || []).filter((b) => b && b.type === 'text').map((b) => b.text).join('\n');
      rows.push({ kind: 'user', seq: e.seq, text });
    } else if (e.type === 'assistant/message') {
      const msg = d.message || d;
      const blocks = Array.isArray(msg.content) ? msg.content : [];
      const reasoning = blocks.filter((b) => b.type === 'reasoning').map((b) => b.text).join('\n');
      const text = blocks.filter((b) => b.type === 'text').map((b) => b.text).join('\n');
      rows.push({ kind: 'assistant', seq: e.seq, text, reasoning });
    } else if (e.type === 'turn/end') {
      rows.push({ kind: 'tail', seq: e.seq, tail: turnTailOf(d.reason, 0) });
    }
  });
  return rows;
}

/** 面板行 HTML（纯函数——panelHtml 与轮询重绘共用同一渲染）。 */
export function rowsHtmlOf(rows) {
  const list = rows || [];
  if (!list.length) return '<div class="small faint">' + esc(tr('尚无对话——下方输入修改意图发送。')) + '</div>';
  return list.map((r) => {
    if (r.kind === 'user') {
      return '<div class="dc-row dc-user"><div class="dc-bubble">' + esc(r.text) + '</div></div>';
    }
    if (r.kind === 'assistant') {
      const rfold = r.reasoning
        ? '<details class="chat-fold dc-fold"><summary><span class="ex-chev">▸</span>' + esc(tr('思考')) + '</summary><div class="ex-body">' + esc(r.reasoning) + '</div></details>'
        : '';
      return '<div class="dc-row dc-assistant">' + rfold +
        (r.text ? '<div class="dc-text">' + esc(r.text) + '</div>' : '') + '</div>';
    }
    if (r.kind === 'tail' && r.tail) {
      return '<div class="dc-row dc-tail"><span class="dc-tail-' + esc(r.tail.cls || 'ok') + '">' +
        esc(r.tail.word || '') + '</span></div>';
    }
    return '';
  }).join('');
}

/** 面板 HTML（纯函数）。 */
export function panelHtml(state) {
  const s = state || {};
  return '<div class="dc-meta">' + (s.refLine || '') + '</div>' +
    '<div class="dc-transcript" data-dc-transcript>' + rowsHtmlOf(s.rows) + '</div>' +
    '<div data-dc-diff></div>' +
    '<div class="dc-composer">' +
    '<textarea class="inp" data-dc-text rows="2" placeholder="' + esc(tr('修改意图（如：把 standard 模型换成 reasoning-tier）——发送后助手回修改后 YAML，对比确认即 Apply')) + '"></textarea>' +
    '<div class="dc-acts">' +
    '<button type="button" class="btn btn-chip sm" data-dc-preview>' + esc(tr('对比并 Apply')) + '</button>' +
    '<button type="button" class="btn btn-chip sm" data-dc-send>' + esc(tr('发送')) + '</button>' +
    '<button type="button" class="btn btn-ghost sm" data-dc-stop hidden>' + esc(tr('停止')) + '</button>' +
    '<span class="small faint" data-dc-msg>' + esc(s.note || '') + '</span>' +
    '</div></div>';
}

/** openDocChat 主入口：htmlModal 宽体内嵌对话修改（同对象续跑同会话）。
 * opts={title, refKey, ref, current, ctx, extractApply?, onApplied?}——缺省 extract=extractYamlBlock。 */
export function openDocChat(opts) {
  const o = opts || {};
  const f = faces();
  const node = document.createElement('div');
  node.className = 'doc-chat';
  if (!f || typeof f.send !== 'function') {
    node.innerHTML = '<div class="ob-empty">' + esc(tr('对话通路未注入（chatfaces.js 缺席）——原地对话修改不可用；可在「原文」视图手编，或经会话屏对话。')) + '</div>';
    htmlModal({ title: o.title || tr('对话修改'), wide: true, node, holder: document.body });
    return;
  }
  const painted = new Set();
  let sessionId = SESSIONS.get(o.refKey || '') || '';
  let running = false;
  let current = String(o.current || '');
  const state = { rows: [], refLine: '', note: '' };
  const park = document.createElement('div'); // 关闭后节点归位处（事件绑定随节点保真，重开零重绑）
  node.innerHTML = panelHtml(state);
  htmlModal({ title: o.title || tr('对话修改'), wide: true, node, holder: park }).then(() => {
    clearInterval(timer); // 关闭=停轮询（会话保留续跑）
  });

  const paintRows = () => {
    const box = node.querySelector('[data-dc-transcript]');
    if (box) box.innerHTML = rowsHtmlOf(state.rows);
    box && (box.scrollTop = box.scrollHeight);
  };

  const refresh = async () => {
    if (!sessionId) return;
    const r = await f.transcript(sessionId, 40);
    if (r.status !== 200 || !Array.isArray(r.data && r.data.events)) return;
    state.rows = transcriptRows(r.data.events);
    paintRows();
    const st = await f.state(sessionId);
    running = !!(st.status === 200 && st.data && st.data.running === true);
    const send = node.querySelector('[data-dc-send]');
    const stop = node.querySelector('[data-dc-stop]');
    if (send) send.disabled = running;
    if (stop) stop.hidden = !running;
  };

  const ensureSession = async () => {
    if (sessionId) return sessionId;
    const r = await f.create();
    if (r.status === 200 && r.data && r.data.sessionId) {
      sessionId = String(r.data.sessionId);
      SESSIONS.set(o.refKey || '', sessionId);
      return sessionId;
    }
    throw new Error(tr('新会话失败（') + String(r.status) + tr('）'));
  };

  const send = async () => {
    const ta = node.querySelector('[data-dc-text]');
    const intent = String((ta && ta.value) || '').trim();
    const msg = node.querySelector('[data-dc-msg]');
    if (!intent) { if (msg) msg.textContent = tr('请先写修改意图'); return; }
    if (running) { if (msg) msg.textContent = tr('上一轮仍在运行——可点停止'); return; }
    try {
      const sid = await ensureSession();
      const first = !state.rows.length;
      const text = first ? buildEditPrompt(o.ref, current, intent) : intent;
      const r = await f.send(sid, text, 'queue');
      if (r.status !== 200) { if (msg) msg.textContent = tr('发送失败：') + String((r.data && r.data.error) || r.status); return; }
      if (ta) ta.value = '';
      await refresh();
    } catch (e) { if (msg) msg.textContent = String((e && e.message) || e); }
  };

  const showDiff = () => {
    const lastAssistant = [...state.rows].reverse().find((r) => r.kind === 'assistant' && r.text);
    const modified = (o.extractApply || extractYamlBlock)(lastAssistant ? lastAssistant.text : '');
    const diffBox = node.querySelector('[data-dc-diff]');
    if (!modified) {
      diffBox.innerHTML = '<div class="small faint" style="margin-top:8px">' + esc(tr('最近回复未含 ```yaml 代码块——请让助手给出修改后的完整 YAML 声明。')) + '</div>';
      return;
    }
    const hunks = hunksOf(diffLines(current, modified));
    diffBox.innerHTML = '<div class="card" style="margin-top:10px"><div class="card-h"><span class="h-title">' +
      esc(tr('修改预览（') + hunks.length + tr(' 处变更）')) + '</span><span class="h-right">' +
      '<button type="button" class="btn btn-primary sm" data-dc-apply>' + esc(tr('批准（Apply）')) + '</button> ' +
      '<button type="button" class="btn btn-secondary sm" data-dc-dismiss>' + esc(tr('放弃')) + '</button>' +
      '</span></div><div class="card-b" style="max-height:300px;overflow:auto">' + renderDiff(diffLines(current, modified)) + '</div></div>';
    diffBox.querySelector('[data-dc-dismiss]').addEventListener('click', () => { diffBox.innerHTML = ''; });
    diffBox.querySelector('[data-dc-apply]').addEventListener('click', async () => {
      const msg = node.querySelector('[data-dc-msg]');
      if (msg) msg.textContent = tr('正在 Apply…');
      try {
        const face = await o.ctx.api.object(o.ref.kind, o.ref.scope, o.ref.name);
        const rv = (face.ok && face.data && face.data.resourceVersion != null) ? Number(face.data.resourceVersion) : (o.ref.rv || 0);
        const r = await o.ctx.api.apply({
          kind: o.ref.kind, scope: o.ref.scope || 'global', name: o.ref.name, resourceVersion: rv,
          content: modified, fieldManager: 'console:doc-chat', reasonRef: 'doc-chat:' + o.ref.name,
        });
        if (r.ok) {
          current = modified;
          diffBox.innerHTML = '';
          if (msg) msg.textContent = tr('已 Apply（rv=') + ((r.data && r.data.rv) != null ? r.data.rv : '?') + tr('）');
          toast(tr('对话修改已 Apply'), 'accent');
          if (typeof o.onApplied === 'function') o.onApplied(modified);
        } else {
          if (msg) msg.textContent = tr('Apply 失败：') + String(r.error || r.status) + (String(r.status) === '409' ? tr('（版本已变——重试将取新版本）') : '');
          toast(tr('Apply 失败（') + String(r.error || r.status) + '）', 'error');
        }
      } catch (e) { if (msg) msg.textContent = String((e && e.message) || e); }
    });
  };

  node.querySelector('[data-dc-send]').addEventListener('click', send);
  node.querySelector('[data-dc-stop]').addEventListener('click', async () => {
    if (sessionId) await f.stop(sessionId);
  });
  // 回车发送（Shift+Enter 换行）
  node.querySelector('[data-dc-text]').addEventListener('keydown', (ev) => {
    if (ev.key === 'Enter' && !ev.shiftKey) { ev.preventDefault(); send(); }
  });
  node.querySelector('[data-dc-preview]').addEventListener('click', showDiff);

  const timer = setInterval(() => { refresh().catch(() => {}); }, 1800);
  refresh().catch(() => {});
}
