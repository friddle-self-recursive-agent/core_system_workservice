/* session-stages.js —— 三段式 Session 页面承载（session.md 权威；2026-11-27 用户裁决
 * 「有没有 scheme 来承载我这个三段式的描述。比如我想控制一些描述？页面上怎么承载」）。
 * 数据流：GET /v1/session-config（三段描述/背景预算/属性回整开关——PUT 保存，随 --allow-apply
 * 门控）+ GET /v1/session-config/writebacks（第三段回整账——「自动回整」来源标记）。
 * 三段进度=诚实推导：第一段=右栏上下文面在册（chunks 已现读）；第二段=会话运行中/等待 apply
 * （chat-state 词形）；第三段=回整账最近行（apply 触发，缺席=未触发不冒充）。
 * 纯函数（normalizeSessionConfig/deriveStageState/renderStagesCard）零 DOM 依赖——node 单测。 */

import { esc, trunc, fmtNum } from '../core/format.js';
import { chip } from '../render/chips.js';
import { t as tr } from '../core/i18n.js';

/* ---------- 纯逻辑（node 单测直测） ---------- */

/** GET 载荷 → 视图模型（字段逐项归一缺省——缺席=scheme 缺省，不冒充）。 */
export function normalizeSessionConfig(data) {
  const d = data && typeof data === 'object' ? data : {};
  const c = d.config && typeof d.config === 'object' ? d.config : {};
  const st = c.stages && typeof c.stages === 'object' ? c.stages : {};
  const def = {
    s1: '第一段·背景装配（LLM 描述命令 → tag 分析先查 CtxCache → 三路注入 → builtin RAG → 类似 session → developer_system 定位仓库 → 背景 ≤5k 逐块溯源）',
    s2: '第二段·预更新执行（反复 ask 询问；会话没有命令——只产文档预更新 draft；等待用户 apply）',
    s3: '第三段·属性回整（apply 触发：ref/tag/重要程度/metainfo 默认自动更新）',
  };
  const scope = (Array.isArray(c.metaScope) ? c.metaScope : []).map(String).filter(Boolean);
  return {
    persistent: d.persistent !== false,
    note: String(d.note || ''),
    stages: {
      s1: String(st.s1 || def.s1),
      s2: String(st.s2 || def.s2),
      s3: String(st.s3 || def.s3),
    },
    backgroundMaxChars: Number(c.backgroundMaxChars) || 5000,
    blockAttribution: c.blockAttribution !== false,
    metaAutoUpdate: c.metaAutoUpdate !== false,
    metaScope: scope.length ? scope : ['ref', 'tag', 'importance', 'metainfo'],
    provider: String(c.provider || 'openai-default'),
    model: String(c.model || 'reasoning-tier'),
    updatedAt: String(c.updatedAt || ''),
  };
}

/** 回整账行归一。 */
export function normalizeWritebacks(data) {
  const d = data && typeof data === 'object' ? data : {};
  return (Array.isArray(d.items) ? d.items : []).map((w) => {
    const row = w && typeof w === 'object' ? w : {};
    return {
      at: String(row.at || ''), resource: String(row.resource || ''),
      subject: String(row.subject || ''), sessionId: String(row.sessionId || ''),
      mode: String(row.mode || 'rule-default'),
      importanceDelta: Number(row.importanceDelta) || 0,
      tagsAdded: (Array.isArray(row.tagsAdded) ? row.tagsAdded : []).map(String),
      note: String(row.note || ''),
    };
  });
}

/** 三段进度诚实推导（纯函数）：ctx 在册=第一段完成；running=第二段进行中；
 * 回整账最近行=第三段已触发（行内 at 即触发时刻）。缺席态显式，不冒充。 */
export function deriveStageState(input) {
  const i = input && typeof input === 'object' ? input : {};
  const ctxReady = !!i.ctxReady;
  const running = i.running === true;
  const wb = (Array.isArray(i.writebacks) && i.writebacks.length) ? i.writebacks[0] : null;
  return [
    { n: 1, name: tr('第一段·背景装配'), done: ctxReady, active: !ctxReady,
      detail: ctxReady ? tr('上下文面已现读（缓存 chunks/分片见右栏）') : tr('等待会话上下文面装配') },
    { n: 2, name: tr('第二段·预更新执行'), done: !running && !!wb, active: running,
      detail: running ? tr('会话运行中（反复 ask 询问；产出=文档预更新 draft）') : (wb ? tr('预更新已 apply——进入回整') : tr('等待输入或等待 apply（会话没有命令——只产 draft）')) },
    { n: 3, name: tr('第三段·属性回整'), done: !!wb, active: false,
      detail: wb ? (wb.resource + (wb.importanceDelta ? tr(' · importance +') + wb.importanceDelta : '') + (wb.tagsAdded && wb.tagsAdded.length ? tr(' · tag+') + wb.tagsAdded.length : '')) : tr('apply 后自动触发（默认回整：ref/tag/重要程度/metainfo）') },
    { wb },
  ];
}

/** 三段设置卡 + 进度条 + 回整账（纯渲染——输入模型 → HTML）。 */
export function renderStagesCard(cfg, stages, writebacks, persistent) {
  const stageRow = (s) => '<div class="mini-row"><div class="mr-main"><div>' +
    (s.done ? chip(tr('完成'), 'ok') : s.active ? chip(tr('进行中'), 'acc') : chip(tr('待触发'), 'mute')) +
    ' <b>' + esc(String(s.name || '')) + '</b></div>' +
    '<div class="mr-sub small">' + esc(trunc(String(s.detail || ''), 90)) + '</div></div></div>';
  const wbRows = (writebacks || []).slice(0, 5).map((w) => '<div class="mini-row"><div class="mr-main"><div>' +
    chip(tr('自动回整'), 'llm tiny') + ' <span class="mono small">' + esc(trunc(w.resource, 48)) + '</span>' +
    (w.importanceDelta ? ' ' + chip('importance +' + w.importanceDelta, 'ok tiny') : '') +
    (w.tagsAdded && w.tagsAdded.length ? w.tagsAdded.map((t) => chip(t, 'mute tiny')).join(' ') : '') +
    '</div><div class="mr-sub small">' + esc(w.at || '') + ' · ' + esc(w.mode) +
    (w.note ? ' · ' + esc(trunc(w.note, 60)) : '') + '</div></div></div>').join('');
  const disp = (b) => (b ? tr('开') : tr('关'));
  return '<div class="card-h"><span class="h-title">' + esc(tr('三段式 Session')) + '</span><span class="h-right">' +
    (persistent ? chip(tr('已持久化'), 'ok tiny') : chip(tr('缺省（未持久化）'), 'warn tiny')) + '</span></div>' +
    '<div class="card-b">' +
    (stages || []).filter((s) => s && s.n).map(stageRow).join('') +
    '<div class="small faint" style="margin-top:6px">' + esc(tr('背景预算 ')) + esc(fmtNum(cfg.backgroundMaxChars)) + tr(' 字 · 逐块溯源 ') +
    esc(disp(cfg.blockAttribution)) + tr(' · 属性回整 ') + esc(disp(cfg.metaAutoUpdate)) +
    tr(' · Provider ') + esc(cfg.provider) + ' / ' + esc(cfg.model) + '</div>' +
    (wbRows ? '<h3 class="sec-title" style="margin:10px 0 4px">' + esc(tr('回整账（第三段）')) + '</h3>' + wbRows : '') +
    '<details class="ex-block" style="margin-top:8px"><summary class="ex-sum"><span class="ex-chev">▶</span>' +
    esc(tr('三段控制（描述/预算/开关——PUT /v1/session-config）')) + '</summary><div class="ex-body">' +
    '<div class="small faint" style="margin:4px 0">' + esc(tr('三段描述与开关=三段式行为的承载面（scheme runtime.context 同源；此处为数据库承载行——保存即生效）。')) + '</div>' +
    '<textarea class="inp" data-stage-desc="s1" rows="2" style="width:100%">' + esc(cfg.stages.s1) + '</textarea>' +
    '<textarea class="inp" data-stage-desc="s2" rows="2" style="width:100%;margin-top:4px">' + esc(cfg.stages.s2) + '</textarea>' +
    '<textarea class="inp" data-stage-desc="s3" rows="2" style="width:100%;margin-top:4px">' + esc(cfg.stages.s3) + '</textarea>' +
    '<div style="display:grid;grid-template-columns:1fr 1fr;gap:6px;margin-top:6px">' +
    '<label class="small">' + esc(tr('背景预算（字）')) + ' <input class="inp" data-stage-num="backgroundMaxChars" type="number" min="500" step="500" value="' + cfg.backgroundMaxChars + '"></label>' +
    '<label class="small">' + esc(tr('回整属性域')) + ' <input class="inp" data-stage-in="metaScope" value="' + esc(cfg.metaScope.join(',')) + '"></label>' +
    '<label class="small">' + esc(tr('会话 Provider')) + ' <input class="inp" data-stage-in="provider" value="' + esc(cfg.provider) + '"></label>' +
    '<label class="small">' + esc(tr('会话模型')) + ' <input class="inp" data-stage-in="model" value="' + esc(cfg.model) + '"></label>' +
    '</div>' +
    '<div style="display:flex;gap:10px;margin-top:6px" class="small">' +
    '<label><input type="checkbox" data-stage-chk="blockAttribution"' + (cfg.blockAttribution ? ' checked' : '') + '> ' + esc(tr('背景逐块溯源')) + '</label>' +
    '<label><input type="checkbox" data-stage-chk="metaAutoUpdate"' + (cfg.metaAutoUpdate ? ' checked' : '') + '> ' + esc(tr('第三段属性自动回整')) + '</label>' +
    '</div>' +
    '<div style="margin-top:8px;display:flex;gap:8px;align-items:center">' +
    '<button class="btn btn-secondary sm" data-stage-save>' + esc(tr('保存三段控制')) + '</button>' +
    '<span class="small faint" data-stage-msg></span></div>' +
    '</div></details>' +
    '</div>';
}

/* ---------- 挂载（右栏注水 + 保存接线） ---------- */

const STAGES = { model: null, writebacks: [], persistent: false };

function paint(box, cfg, stages, writebacks, persistent) {
  box.innerHTML = '<div class="card" data-stages-card>' + renderStagesCard(cfg, stages, writebacks, persistent) + '</div>';
  const save = box.querySelector('[data-stage-save]');
  if (save) save.addEventListener('click', async () => {
    const q = (sel) => box.querySelector(sel);
    const payload = {
      stages: {
        s1: q('[data-stage-desc="s1"]').value,
        s2: q('[data-stage-desc="s2"]').value,
        s3: q('[data-stage-desc="s3"]').value,
      },
      backgroundMaxChars: Number(q('[data-stage-num="backgroundMaxChars"]').value) || 5000,
      blockAttribution: q('[data-stage-chk="blockAttribution"]').checked,
      metaAutoUpdate: q('[data-stage-chk="metaAutoUpdate"]').checked,
      metaScope: q('[data-stage-in="metaScope"]').value.split(/[\s,，、]+/).map((x) => x.trim()).filter(Boolean),
      provider: q('[data-stage-in="provider"]').value.trim() || 'openai-default',
      model: q('[data-stage-in="model"]').value.trim() || 'reasoning-tier',
    };
    const msg = q('[data-stage-msg]');
    if (msg) msg.textContent = tr('保存中…');
    const r = await window.__CC_CONSOLE_CTX__.api.saveSessionConfig(payload);
    if (msg) msg.textContent = r.ok ? tr('已保存（') + ((r.data && r.data.updatedAt) || '') + tr('）') : tr('保存失败：') + String(r.error || r.status);
    refresh(box);
  });
}

async function refresh(box) {
  const api = (typeof window !== 'undefined' && window.__CC_CONSOLE_CTX__ && window.__CC_CONSOLE_CTX__.api) || null;
  if (!api || !api.sessionConfig) return;
  const [c, w] = await Promise.all([api.sessionConfig(), api.sessionWritebacks ? api.sessionWritebacks(5) : Promise.resolve({ ok: false })]);
  if (!c.ok) return;
  STAGES.model = normalizeSessionConfig(c.data);
  STAGES.persistent = STAGES.model.persistent;
  STAGES.writebacks = w && w.ok ? normalizeWritebacks(w.data) : STAGES.writebacks;
  const ctxCard = document.querySelector('[data-ctx-card]');
  const running = !!(window.__CC_CHAT_STATE__ && window.__CC_CHAT_STATE__.running === true);
  const stages = deriveStageState({ ctxReady: !!ctxCard, running, writebacks: STAGES.writebacks });
  paint(box, STAGES.model, stages, STAGES.writebacks, STAGES.persistent);
}

/** mountSessionStages 右栏三段卡挂载（chat.js 调用；轮询重绘由 chat refreshState 附带触发）。 */
export async function mountSessionStages(el) {
  const box = el.querySelector('[data-stages-panel]');
  if (!box) return;
  await refresh(box);
}

/** repaintSessionStages 轻量重绘（chat refreshState 每拍调用——不重新拉数，只重推进度）。 */
export function repaintSessionStages(el) {
  const box = el.querySelector('[data-stages-panel]');
  if (!box || !STAGES.model) return;
  const ctxCard = document.querySelector('[data-ctx-card]');
  const running = !!(window.__CC_CHAT_STATE__ && window.__CC_CHAT_STATE__.running === true);
  const stages = deriveStageState({ ctxReady: !!ctxCard, running, writebacks: STAGES.writebacks });
  paint(box, STAGES.model, stages, STAGES.writebacks, STAGES.persistent);
}
