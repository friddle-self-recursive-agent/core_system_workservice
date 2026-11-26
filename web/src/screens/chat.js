/* chat.js —— rootAgent 对话屏（2026-11-26 用户裁决·二次修正版）：
 *   入口永远是一个对话框（可选 session）；对话框目标=找文档、更新文档。
 *   交互=AI 主动提示（涉及文档 ref 可点 / 生成结论卡 / 完整文档模式观看）；
 *   消息与思考过程可 expand；数据通路=__CC.chatFaces.send（plugin chat-send 桥——
 *   sessionController.prompt queue，selfheal 同款受信通路）；无 reply 捕获=诚实词形
 *   （结论经原生界面返回，agent 自己 apply——文档状态徽与节点图跟进）。
 * 完整文档模式：data-doc-full="Kind/scope/name" → 拉取对象正文 → 渲染 md 弹层+写入草稿。 */

import { esc, trunc, relAge } from '../core/format.js';
import { chip, chipGo, screenHead } from '../render/chips.js';
import { renderMarkdown } from '../render/md.js';
import { renderMessage, expandBlock, linkifyRefs } from '../render/refs.js';
import { mountSessionCtx } from './sessionctx.js';
import { t as tr } from '../core/i18n.js';

function faces() {
  return (typeof window !== 'undefined' && window.__CC && window.__CC.chatFaces) || null;
}

/* 线程态（页面级内存——诚实口径：本地线程只记本控制台发出的消息；权威轨迹=宿主原生会话）。 */
const THREAD = { sessionId: '', msgs: [] };

function sessionOptions(ctx, cur) {
  const rows = (ctx.store.sessions() || []).map((s) => String(s && (s.id || s.sessionId) || '')).filter(Boolean);
  const seen = new Set(rows);
  const opts = rows.slice(0, 30).map((id) =>
    '<option value="' + esc(id) + '"' + (id === cur ? ' selected' : '') + '>' + esc(trunc(id, 30)) + '</option>').join('');
  const extra = cur && !seen.has(cur)
    ? '<option value="' + esc(cur) + '" selected>' + esc(trunc(cur, 30)) + '</option>' : '';
  return '<option value="">' + esc(tr('— 选择 session（或输入）—')) + '</option>' + extra + opts;
}

/* 对话 composer（session 可选 + 文本域 + 发送——大屏顶部与 chat 屏共用）。 */
export function composerHtml(ctx, curSession) {
  const f = faces();
  const canSend = !!(f && typeof f.send === 'function');
  return '<div class="cc-composer">' +
    '<div class="cc-comp-row">' +
    '<select class="inp cc-sess" data-cc-sess>' + sessionOptions(ctx, curSession) + '</select>' +
    '<input class="inp cc-sess-free" data-cc-sess-free placeholder="' + esc(tr('或手填 session id')) + '" value="' + esc(curSession || '') + '">' +
    '</div>' +
    '<div class="cc-comp-row">' +
    '<textarea class="inp cc-text" data-cc-text rows="2" placeholder="' +
    esc(tr('提交需求／找文档／更新文档——例如：进行代码审核，看看实际和文档差距有多大，并进行修复')) + '"></textarea>' +
    '<button class="btn primary cc-send" data-cc-send' + (canSend ? '' : ' disabled') + '>' + esc(tr('发送')) + '</button>' +
    '</div>' +
    '<div class="small faint">' +
    (canSend
      ? esc(tr('Enter 发送 · queue 语义（送达 agent 会话；结论由 agent 落文档后经状态徽与节点图可见）'))
      : esc(tr('对话通路未注入（chatfaces.js 缺席）——送达面显式缺席，不冒充已发送'))) +
    '</div></div>';
}

/* 线程渲染（消息数组 → 气泡流；expand 内建）。 */
function threadHtml(msgs, names) {
  if (!msgs.length) {
    return '<div class="card"><div class="ob-empty">' + esc(tr('线程为空——发第一条需求；agent 的完整轨迹在原生界面（新窗口）可见。')) + '</div></div>';
  }
  return '<div class="card thread">' + msgs.map((m) => renderMessage(m, names)).join('') + '</div>';
}

/* 完整文档模式弹层（data-doc-full → 现读对象正文 → 渲染 md + 写入草稿入口）。 */
function bindDocFull(el, ctx) {
  el.querySelectorAll('[data-doc-full]').forEach((btn) => {
    btn.addEventListener('click', async () => {
      const ref = String(btn.getAttribute('data-doc-full') || '');
      const parts = ref.split('/');
      if (parts.length < 3) return;
      const [kind, scope, name] = parts;
      const r = await ctx.api.object(decodeURIComponent(kind), decodeURIComponent(scope), decodeURIComponent(name));
      const body = r.ok && r.data && r.data.content != null ? String(r.data.content) : (tr('正文暂不可读（') + (r.ok ? tr('内容缺席') : String(r.error || r.status)) + '）');
      let ov = document.querySelector('[data-doc-ov]');
      if (!ov) {
        ov = document.createElement('div');
        ov.setAttribute('data-doc-ov', '1');
        document.body.appendChild(ov);
      }
      ov.className = 'doc-ov';
      ov.innerHTML = '<div class="doc-ov-card">' +
        '<div class="doc-ov-h"><b class="mono">' + esc(decodeURIComponent(kind) + '/' + decodeURIComponent(scope) + '/' + decodeURIComponent(name)) + '</b>' +
        '<button class="btn sm" data-doc-ov-close>' + esc(tr('关闭')) + '</button>' +
        '<button class="btn primary sm" data-doc-ov-edit>' + esc(tr('写入草稿（编辑器）')) + '</button></div>' +
        '<div class="doc-ov-b md-body">' + renderMarkdown(body) + '</div></div>';
      ov.querySelector('[data-doc-ov-close]').addEventListener('click', () => { ov.className = 'doc-ov hidden'; });
      ov.querySelector('[data-doc-ov-edit]').addEventListener('click', () => {
        ctx.go('doc/' + encodeURIComponent(decodeURIComponent(kind)) + '/' + encodeURIComponent(decodeURIComponent(scope)) + '/' + encodeURIComponent(decodeURIComponent(name)));
        ov.className = 'doc-ov hidden';
      });
    });
  });
}

/* 绑定 composer 行为（session 解析+发送+本地回执）。 */
export function bindComposer(el, ctx, rootEl) {
  const sendBtn = rootEl.querySelector('[data-cc-send]');
  const sessSel = rootEl.querySelector('[data-cc-sess]');
  const sessFree = rootEl.querySelector('[data-cc-sess-free]');
  const textTa = rootEl.querySelector('[data-cc-text]');
  if (!sendBtn || !textTa) return;
  const doSend = async () => {
    const text = textTa.value.trim();
    if (!text) return;
    const sessionId = (sessFree && sessFree.value.trim()) || (sessSel ? sessSel.value : '') || 'root';
    THREAD.sessionId = sessionId;
    const f = faces();
    if (!f || typeof f.send !== 'function') return;
    THREAD.msgs.push({ role: 'user', text, ts: new Date().toISOString().slice(11, 19) });
    textTa.value = '';
    paintThread(rootEl, ctx);
    const r = await f.send(sessionId, text);
    if (r && r.status === 200) {
      THREAD.msgs.push({
        role: 'assistant', name: 'rootAgent', seq: (r.data && r.data.requestId) || '',
        ts: new Date().toISOString().slice(11, 19),
        text: tr('已送达（queue）——agent 会话已受理。结论将由 agent 生成并自行落文档；文档状态徽与 workflow 节点图随后可见。原生轨迹：'),
        tool: 'send → sessionController.prompt · requestId=' + ((r.data && r.data.requestId) || '—'),
      });
    } else {
      THREAD.msgs.push({
        role: 'assistant', name: 'system',
        text: tr('送达失败（') + ((r && r.status) || '—') + '）' + ((r && r.data && r.data.error) ? ' · ' + r.data.error : ''),
      });
    }
    paintThread(rootEl, ctx);
  };
  sendBtn.addEventListener('click', doSend);
  textTa.addEventListener('keydown', (ev) => {
    if (ev.key === 'Enter' && !ev.shiftKey) { ev.preventDefault(); doSend(); }
  });
}

function paintThread(rootEl, ctx) {
  const box = rootEl.querySelector('[data-thread]');
  if (!box) return;
  const names = new Set((ctx.store.objects() || []).map((o) => o.name).filter(Boolean));
  box.innerHTML = threadHtml(THREAD.msgs, names);
  box.scrollTop = box.scrollHeight;
  bindDocFull(box, ctx);
}

/** #/chat/:agent 与 #/session/:id —— 对话屏（对话框是主体；原生界面收进 expand）。 */
function dialogScreen(el, ctx, sessionId) {
  const f = faces();
  const native = expandBlock(
    '<iframe class="host-embed" src="' + esc((function () { try { return location.origin + '/?embedded=1'; } catch { return '/'; } })()) + '" referrerpolicy="same-origin"></iframe>',
    tr('宿主原生界面（权威轨迹）'), { cls: 'host-ex' });
  el.innerHTML = '<div class="view-inner">' +
    screenHead(tr('对话'), tr('rootAgent——找文档、更新文档；ref 可点，结论可全文档观看'),
      chipGo(tr('新窗口原生界面'), '', 'mute', false)) +
    composerHtml(ctx, sessionId) +
    '<h2 class="sec-title" style="margin-top:14px">' + esc(tr('对话线程')) + '</h2>' +
    '<div data-thread>' + threadHtml(THREAD.msgs, null) + '</div>' +
    '<div class="small faint" style="margin-top:8px">' +
    esc(tr('本地线程只记本控制台发出的消息；权威轨迹=宿主原生会话（下方 expand 内嵌）。')) + '</div>' +
    native;
  if (sessionId) {
    const sess = (ctx.store.sessions() || []).find((s) => String((s && (s.id || s.sessionId)) || '') === sessionId);
    if (sess || true) {
      el.insertAdjacentHTML('beforeend', '');
    }
    // G-14 上下文面板（展开态由 sessionctx 自管）
    el.insertAdjacentHTML('beforeend',
      '<div class="view-inner"><h2 class="sec-title">' + esc(tr('上下文')) + '</h2>' +
      '<div data-ctx-panel><div class="card"><div class="card-b"><div class="small faint">' +
      esc(tr('上下文注水中…（GET /v1/sessions/')) + esc(encodeURIComponent(sessionId)) + '/context）</div></div></div></div></div>');
    mountSessionCtx(el, ctx, sessionId);
  }
  // 新窗口原生界面按钮
  const pop = el.querySelector('[data-go=""]');
  if (pop) {
    pop.removeAttribute('data-go');
    pop.addEventListener('click', () => { try { window.open(location.origin + '/?embedded=1', '_blank'); } catch { /* ignore */ } });
  }
  bindComposer(el, ctx, el);
  bindDocFull(el, ctx);
  paintThread(el, ctx);
}

export function render(el, ctx) {
  const id = ctx.params && ctx.params.id;
  dialogScreen(el, ctx, id ? String(id) : (THREAD.sessionId || ''));
}
