/* modal.js —— 弹窗与提示组件（2026-11-28 用户裁决「不要用系统弹窗。要用标准的 deepseek style
 * 实现弹窗。默认请用 modal」——style.md §3.3）：原生 window.confirm/alert/prompt 全站禁用
 * （style.md §3.3 grep 清零判据）——确认类交互=本组件 confirmModal（.chat-modal 族 DeepSeek
 * 样式，Promise 语义），结果/错误提示=toast（非阻断右上浮层）。两件共用 token 体系，调用点
 * 零原生框。 */

import { esc } from '../core/format.js';
import { t as tr } from '../core/i18n.js';

/* confirmModal 确认弹窗：{title, body, confirmText, cancelText, danger} → Promise<boolean>。
 * 确认=主操作变体（danger=true=ghost 形 --danger 字——style.md §3.1 破坏性纪律）；取消=次操作；
 * Esc/点 mask=取消（resolve false）；不引入新样式类——全量复用 .chat-modal 族既有 token。
 * 同屏允许多次调用（每次独立挂载，动作后即拆——workflow-config 节点弹窗 [data-nm-modal]
 * 同族不同锚，互不干扰）。 */
export function confirmModal(opts) {
  const o = opts || {};
  return new Promise((resolve) => {
    const modal = document.createElement('div');
    modal.className = 'chat-modal';
    modal.setAttribute('data-cc-modal', '1');
    const confirmCls = o.danger ? 'btn btn-ghost btn-danger' : 'btn btn-primary';
    modal.innerHTML = '<div class="chat-modal-mask" data-cc-modal-mask></div>' +
      '<div class="chat-modal-panel" style="width:420px;max-width:calc(100vw - 48px)">' +
      '<div class="chat-modal-head"><div class="chat-modal-title">' + esc(tr(o.title || tr('确认操作'))) + '</div></div>' +
      '<div class="chat-modal-body"><div class="small" style="line-height:1.6">' + (o.body || '') + '</div></div>' +
      '<div style="display:flex;justify-content:flex-end;gap:8px;padding:12px 16px 16px">' +
      '<button type="button" class="btn btn-secondary sm" data-cc-modal-cancel>' + esc(tr(o.cancelText || '取消')) + '</button>' +
      '<button type="button" class="' + confirmCls + ' sm" data-cc-modal-ok>' + esc(tr(o.confirmText || '确定')) + '</button>' +
      '</div></div>';
    const done = (v) => {
      modal.remove();
      document.removeEventListener('keydown', onKey, true);
      resolve(v);
    };
    const onKey = (e) => { if (e.key === 'Escape') done(false); };
    modal.querySelector('[data-cc-modal-ok]').addEventListener('click', () => done(true));
    modal.querySelector('[data-cc-modal-cancel]').addEventListener('click', () => done(false));
    modal.querySelector('[data-cc-modal-mask]').addEventListener('click', () => done(false));
    document.addEventListener('keydown', onKey, true);
    document.body.appendChild(modal);
    const ok = modal.querySelector('[data-cc-modal-ok]');
    if (ok) ok.focus();
  });
}

/* toast 非阻断提示：toast(message, kind)——kind: error（--danger 语义条，缺省）| accent（成功/强调）。
 * 右上浮层 3.5s 自动消散（style.md §3.3——结果/错误不抢主操作焦点）；同帧多条纵向堆叠；
 * 容器懒建一次全局复用（.chat-toast-wrap fixed）。消息文本调用方已拼好（可含 esc 后的片段）。 */
export function toast(message, kind) {
  let wrap = document.querySelector('.chat-toast-wrap');
  if (!wrap) {
    wrap = document.createElement('div');
    wrap.className = 'chat-toast-wrap';
    document.body.appendChild(wrap);
  }
  const t = document.createElement('div');
  t.className = 'chat-toast' + (kind === 'accent' ? ' is-accent' : ' is-error');
  t.innerHTML = message;
  wrap.appendChild(t);
  setTimeout(() => {
    t.classList.add('is-leaving');
    setTimeout(() => t.remove(), 260);
  }, 3500);
}
