/* audit.js —— 代码审核结果屏（V-18，2026-11-26 概念可见性批）：#/audit。
 * 实测依据：POST /v1/connectops/connectops-code-audit/findings 受 G-03 fail-closed 权限门
 * 拦截（匿名/只读代理恒 403 且代理只放 GET）→ 后端新增只读投影 GET /v1/code-audit/findings
 * （+ snapshot.codeAudit 摘要），数据直取 connectops-code-audit 进程内 LRU（含 Verify gate），
 * 零写面、零假发现、零冒充通过。纯函数 renderAuditPayload 零 DOM——node 单测直测。 */

import { esc, trunc, relAge } from '../core/format.js';
import { chip, chipGo, emptyBox, screenHead } from '../render/chips.js';
import { t as tr } from '../core/i18n.js';

function auditCardHtml(a) {
  const findings = Array.isArray(a && a.findings) ? a.findings : [];
  const unbound = Array.isArray(a && a.unbound) ? a.unbound : [];
  const skipped = Array.isArray(a && a.skipped) ? a.skipped : [];
  const errN = findings.filter((f) => f && f.severity === 'error').length;
  const warnN = findings.length - errN;
  const ok = !!(a && a.ok);
  const tool = a && a.tool && a.tool.name ? (String(a.tool.name) + (a.tool.version ? '@' + String(a.tool.version) : '')) : 'connectops-code-audit';
  const ranAt = String((a && a.ranAt) || '');

  const findHtml = findings.length
    ? '<div style="overflow-x:auto;margin-top:8px"><table class="tbl"><thead><tr>' +
      '<th>' + esc(tr('规则')) + '</th><th>' + esc(tr('级别')) + '</th><th>' + esc(tr('位置')) + '</th><th>' + esc(tr('说明')) + '</th><th>checker</th>' +
      '</tr></thead><tbody>' +
      findings.slice(0, 100).map((f) => {
        const sev = String((f && f.severity) || 'error');
        const loc = String((f && f.file) || '—') + (f && Number(f.line) > 0 ? ':' + String(f.line) : '');
        return '<tr>' +
          '<td class="mono small">' + esc(String((f && f.rule) || '—')) + '</td>' +
          '<td>' + chip(sev, sev === 'error' ? 'esc' : 'warn') + '</td>' +
          '<td class="mono small">' + esc(trunc(loc, 56)) + '</td>' +
          '<td class="small">' + esc(trunc(String((f && f.message) || '—'), 120)) + '</td>' +
          '<td class="mono small faint">' + esc(String((f && f.checkRef) || '—')) + '</td>' +
          '</tr>';
      }).join('') + '</tbody></table></div>'
    : '<div class="small faint" style="margin-top:6px">' + esc(tr('findings 为空（零代码级违规发现）。')) + '</div>';

  const unboundHtml = unbound.length
    ? '<div style="margin-top:8px"><div class="small" style="font-weight:600;margin-bottom:4px">' + esc(tr('未绑定检查器（unbound——显式阻断，不跳过）：')) + '</div>' +
      unbound.map((u) => '<div class="mini-row"><div class="mr-main"><div>' +
        chip('unbound', 'esc') + ' <b class="mono">' + esc(String((u && u.rule) || '*')) + '</b> ' +
        (u && u.checkRef ? chip(String(u.checkRef), 'mute') : '') + '</div>' +
        '<div class="mr-sub small">' + esc(String((u && u.reason) || '—')) + '</div></div></div>').join('') +
      '</div>'
    : '';

  const skippedHtml = skipped.length
    ? '<details class="ex-block" style="margin-top:8px"><summary class="ex-sum"><span class="ex-chev">▸</span>' +
      esc(tr('非 code 规则对账（agent/human 路由 · ')) + skipped.length + '）</summary><div class="ex-body">' +
      skipped.map((s) => '<div class="small mono">' + chip(String((s && s.mode) || '—'), 'mute') + ' ' +
        esc(String((s && s.rule) || '—')) + ' — <span class="faint">' + esc(String((s && s.reason) || '')) + '</span></div>').join('') +
      '</div></details>'
    : '';

  return '<section class="card" style="margin-bottom:12px"><div class="card-h">' +
    '<span class="h-title mono">' + esc(String((a && a.auditId) || '—')) + '</span>' +
    '<span class="h-right">' +
    chip(ok ? tr('通过') : tr('阻断'), ok ? 'ok' : 'esc') + ' ' +
    chip('error ' + errN, errN ? 'esc' : 'mute') + ' ' +
    chip('warn ' + warnN, warnN ? 'warn' : 'mute') +
    (unbound.length ? ' ' + chip('unbound ' + unbound.length, 'esc') : '') +
    '</span></div><div class="card-b">' +
    '<div class="small faint">' + esc(tool) +
    (ranAt ? ' · ' + esc(ranAt.replace('T', ' ').slice(0, 19)) + (relAge(ranAt) ? '（' + esc(relAge(ranAt)) + '）' : '') : '') + '</div>' +
    findHtml + unboundHtml + skippedHtml +
    '</div></section>';
}

/** 审核结果载荷渲染（纯函数——node 单测机械直测）。 */
export function renderAuditPayload(data) {
  if (!data || typeof data !== 'object') {
    return '<div class="card"><div class="card-b">' +
      emptyBox(tr('代码审核只读面未装配'), tr('后端未暴露 /v1/code-audit/findings（旧版 cored 或驱动未装配——诚实降级，不冒充通过）。')) +
      '</div></div>';
  }
  const audits = Array.isArray(data.audits) ? data.audits : [];
  if (!audits.length) {
    return '<div class="card"><div class="card-b">' +
      emptyBox(tr('暂无代码审核记录'), tr('connectops-code-audit 进程内 LRU 为空——Verify 相位或审核 op 执行后此处实时展示 findings/unbound/skipped（零假数据）。')) +
      '</div></div>';
  }
  return '<div class="small faint" style="margin-bottom:8px">' +
    esc(tr('数据源：')) + esc(String(data.source || 'connectops-code-audit')) +
    esc(tr(' 只读投影（最近 ')) + audits.length + esc(tr(' 次审核；只读不改代码、不替代审批）。')) +
    '</div>' + audits.map(auditCardHtml).join('');
}

/** 代码审核结果屏（#/audit）。 */
export function render(el, ctx) {
  const snap = ctx.store && ctx.store.codeAudit ? ctx.store.codeAudit() : null;
  el.innerHTML = '<div class="view-inner">' +
    screenHead(tr('代码审核'), tr('connectops-code-audit 只读 findings 投影（R-TYPE-3：code 规则机械审核；error/unbound 阻断，warn 进报告）'),
      chipGo(tr('工作流'), 'workflows', 'mute', false) + ' ' +
      chipGo(tr('规则设计'), 'designdoc/' + encodeURIComponent('core_design') + '/' + encodeURIComponent('rule/rule_type.md'), 'mute', false)) +
    '<div data-audit-body>' +
    (snap ? renderAuditPayload(snap) : '<div class="card"><div class="card-b"><div class="small faint">' + esc(tr('代码审核 findings 现读中…（GET /v1/code-audit/findings）')) + '</div></div></div>') +
    '</div></div>';

  const stamp = typeof location !== 'undefined' ? location.hash : '';
  const fetchFn = ctx.api && typeof ctx.api.codeAuditFindings === 'function'
    ? ctx.api.codeAuditFindings()
    : Promise.resolve({ ok: false, status: 404, error: '接口缺席' });

  fetchFn.then((r) => {
    if (!el.isConnected || (typeof location !== 'undefined' ? location.hash : '') !== stamp) return;
    const box = el.querySelector('[data-audit-body]');
    if (!box) return;
    if (!r.ok) {
      box.innerHTML = '<div class="card"><div class="card-b">' +
        emptyBox(tr('代码审核只读面暂不可达（HTTP ') + String(r.status || '?') + '）',
          String(r.error || tr('需要带 /v1/code-audit/findings 只读投影的 cored 版本——不冒充零发现。'))) +
        '</div></div>';
      return;
    }
    box.innerHTML = renderAuditPayload(r.data);
  }).catch(() => {
    const box = el.querySelector('[data-audit-body]');
    if (box && !snap) box.innerHTML = '<div class="card"><div class="card-b"><div class="small faint">' + esc(tr('代码审核 findings 读取异常——不冒充通过。')) + '</div></div></div>';
  });
}
