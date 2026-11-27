/* home.js —— 首页=对话（2026-11-27 用户裁决「做成 DeepSeek 的 chat 模式」）：
 * #/ 默认渲染本屏：DeepSeek 新对话版型的大 composer（居中 hero / 圆角大卡 / 圆形品牌蓝发送 /
 * 真实示例 chips / 图片附加）+ 最近会话（真实清单，空=不渲染）。
 * 发送链路：chatFaces.create（宿主发 sessionId）→ chatFaces.send（queue 落投）→ 进入会话屏
 * #/session/<id> 看真实转录。零假开关：DeepThink/Search 是 DeepSeek 能力位，本系统没有对应
 * 真通路，故底栏左侧=真实可发示例 chips + 图片（真上传），不摆假 toggle。
 * 组件复用：heroComposerHtml/bindHeroComposer 同时供总览大屏（welcome）使用——同一版型单点。 */
import { esc, trunc, relAge } from '../core/format.js';
import { t as tr } from '../core/i18n.js';

function faces() {
  return (typeof window !== 'undefined' && window.__CC && window.__CC.chatFaces) || null;
}

/** 真实示例（点击填入，不自动发送）——键=短标签，值=完整可发提示。 */
const SAMPLES = [
  ['代码审核', '进行代码审核，看看实际和文档差距有多大，并进行修复'],
  ['找文档', '找出所有 drift 的文档并生成对账清单'],
  ['更新文档', '把当前循环的状态写成一份 SessionLog 摘要并更新相关文档'],
];

/** DeepSeek 式大 composer（home 与总览大屏共用；不含标题/最近会话）。 */
export function heroComposerHtml() {
  return '<div class="home-box" data-home-box>' +
    '<textarea class="home-text" data-home-text rows="2" placeholder="' +
    esc(tr('提交需求、找文档、更新文档……')) + '" aria-label="' + esc(tr('对话输入')) + '"></textarea>' +
    '<div class="home-bar">' +
    '<div class="home-tools">' +
    SAMPLES.map(([label, full]) =>
      '<button type="button" class="home-tool" data-home-sample="' + esc(full) + '">' + esc(tr(label)) + '</button>').join('') +
    '<label class="home-tool home-img" title="' + esc(tr('附加图片（≤8MB）')) + '">🖼 ' + esc(tr('图片')) +
    '<input type="file" accept="image/png,image/jpeg,image/gif,image/webp,image/svg+xml" data-home-img hidden></label>' +
    '</div>' +
    '<button type="button" class="home-send" data-home-send aria-label="' + esc(tr('发送')) + '" title="' + esc(tr('发送')) + '">' +
    '<span aria-hidden="true">↑</span></button>' +
    '</div></div>' +
    '<div class="home-note" data-home-note aria-live="polite"></div>';
}

/** 最近会话（真实 sessions 清单；空清单=空串，不造假行）。 */
export function recentSessionsHtml(items) {
  const rows = (Array.isArray(items) ? items : [])
    .filter((s) => s && s.sessionId)
    .sort((a, b) => (Number(b.updatedAt) || 0) - (Number(a.updatedAt) || 0))
    .slice(0, 6);
  if (!rows.length) return '';
  return '<div class="home-recent-h">' + esc(tr('最近会话')) + '</div>' +
    '<div class="home-recent-list">' +
    rows.map((s) => {
      const id = String(s.sessionId);
      const title = String(s.title || '').trim() || id;
      return '<button type="button" class="home-recent-chip" data-go="' + esc(sessionRoute(id)) + '" title="' + esc(id) + '">' +
        '<span class="home-recent-dot' + (s.running === true ? ' live' : '') + '"></span>' +
        '<span class="home-recent-t">' + esc(trunc(title, 30)) + '</span>' +
        (s.updatedAt ? '<span class="home-recent-time">' + esc(relAge(Number(s.updatedAt)) || '') + '</span>' : '') +
        '</button>';
    }).join('') +
    '<button type="button" class="home-recent-chip all" data-go="chat/core-headless">' + esc(tr('全部会话')) + ' →</button>' +
    '</div>';
}

/** 发送成功后的会话屏路由（会话 id 单段编码——#/session/<id>）。 */
export function sessionRoute(sessionId) {
  return 'session/' + encodeURIComponent(String(sessionId));
}

function note(el, msg) {
  const n = el.querySelector('[data-home-note]');
  if (n) n.textContent = String(msg || '');
}

/** textarea 自增长（2-8 行）。 */
function autoGrow(ta) {
  if (!ta) return;
  ta.style.height = 'auto';
  const max = 24 * 8 + 12;
  ta.style.height = Math.min(ta.scrollHeight, max) + 'px';
}

/** 发送：create（宿主发 id）→ send（queue）→ 进入会话屏。 */
export async function submitHome(el, ctx) {
  const ta = el.querySelector('[data-home-text]');
  const btn = el.querySelector('[data-home-send]');
  const text = String((ta && ta.value) || '').trim();
  if (!text) return false;
  const f = faces();
  if (!f || typeof f.send !== 'function' || typeof f.create !== 'function') {
    note(el, tr('对话通路未注入（chatfaces.js 缺席）——显式缺席，不冒充已发送'));
    return false;
  }
  if (btn) btn.disabled = true;
  note(el, tr('正在创建会话…'));
  let sid = '';
  let crStatus = 0;
  try {
    const cr = await f.create();
    crStatus = Number((cr && cr.status) || 0);
    if (crStatus === 200 && cr.data && cr.data.sessionId) sid = String(cr.data.sessionId);
  } catch { /* 下方显式失败词形 */ }
  if (!sid) {
    if (btn) btn.disabled = false;
    note(el, tr('新会话创建失败') + '（' + (crStatus || '—') + '）');
    return false;
  }
  note(el, tr('正在发送…'));
  let r;
  try {
    r = await f.send(sid, text);
  } catch (e) {
    r = { status: 0, data: { error: String((e && e.message) || e) } };
  }
  if (r && r.status === 200) {
    if (ta) ta.value = '';
    note(el, tr('已送达，正在打开会话…'));
    ctx.go(sessionRoute(sid));
    return true;
  }
  if (btn) btn.disabled = false;
  note(el, tr('发送失败') + '（' + String((r && r.status) || 0) +
    ((r && r.data && r.data.error) ? ' · ' + String(r.data.error) : '') + '）');
  return false;
}

/** 图片附加：上传 → markdown 引用追加到 composer（agent prompt 契约不变）。 */
function bindImage(el) {
  const input = el.querySelector('[data-home-img]');
  const ta = el.querySelector('[data-home-text]');
  if (!input || !ta) return;
  input.addEventListener('change', async () => {
    const file = (input.files || [])[0];
    input.value = '';
    if (!file || !String(file.type || '').startsWith('image/')) return;
    if (file.size > (8 << 20)) { note(el, tr('图片超过 8MB 上限')); return; }
    const f = faces();
    if (!f || typeof f.uploadImage !== 'function') { note(el, tr('图片上传通路未注入')); return; }
    note(el, tr('图片上传中…'));
    try {
      const url = await f.uploadImage(file);
      const tag = '![image](' + url + ')';
      ta.value = ta.value.trimEnd() + (ta.value.trimEnd() ? '\n' : '') + tag;
      autoGrow(ta);
      note(el, tr('图片已附加（随消息发送）'));
    } catch (e) {
      note(el, tr('图片上传失败') + '（' + String((e && e.message) || e) + '）');
    }
  });
}

/** 绑定 composer 行为（home 与总览大屏共用）。 */
export function bindHeroComposer(el, ctx) {
  const ta = el.querySelector('[data-home-text]');
  const btn = el.querySelector('[data-home-send]');
  if (!ta || !btn) return;
  ta.addEventListener('input', () => autoGrow(ta));
  ta.addEventListener('keydown', (ev) => {
    if (ev.key === 'Enter' && !ev.shiftKey) { ev.preventDefault(); submitHome(el, ctx); }
  });
  btn.addEventListener('click', () => submitHome(el, ctx));
  el.querySelectorAll('[data-home-sample]').forEach((b) => {
    b.addEventListener('click', () => {
      const full = String(b.getAttribute('data-home-sample') || '');
      if (!full) return;
      ta.value = full;
      autoGrow(ta);
      ta.focus();
    });
  });
  bindImage(el);
  autoGrow(ta);
}

/** 最近会话注水（读失败=不渲染——不冒充空/不造假行）。 */
async function loadRecent(el) {
  const box = el.querySelector('[data-home-recent]');
  const f = faces();
  if (!box || !f || typeof f.sessions !== 'function') return;
  try {
    const r = await f.sessions();
    if (!el.isConnected) return;
    const items = r && r.status === 200 && r.data && Array.isArray(r.data.items) ? r.data.items : [];
    box.innerHTML = recentSessionsHtml(items);
  } catch { /* 保持空——不冒充空态 */ }
}

/** #/（默认首页=DeepSeek 式对话入口）。 */
export function render(el, ctx) {
  el.innerHTML = '<div class="home">' +
    '<div class="home-hero">' +
    '<h1 class="home-title">' + esc(tr('有什么可以帮忙的？')) + '</h1>' +
    '<div class="home-sub">' + esc(tr('直接说需求——找文档、更新文档、提交变更；结论落到文档与流程。')) + '</div>' +
    heroComposerHtml() +
    '<div class="home-recent" data-home-recent></div>' +
    '</div></div>';
  bindHeroComposer(el, ctx);
  loadRecent(el);
}
