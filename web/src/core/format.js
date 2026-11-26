/* format.js —— 基元层：HTML 转义 / 链接白名单 / 时间数字格式化。
 * 安全红线（全仓适用）：一切动态文本入 HTML 前必经 esc()；链接只接受 http(s) 与相对路径
 * （#/、/、./），零 javascript:/data: 通路。零 DOM 依赖——node 单测可机械直测。 */

/** HTML 转义（入 HTML 的唯一文本通道）。 */
export function esc(s) {
  return String(s == null ? '' : s)
    .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;').replace(/'/g, '&#39;');
}

/** 链接白名单：http(s) 或相对路径（#/…、/…、./…）。不过=空串（调用方降级纯文本）。 */
export function safeHref(u) {
  const s = String(u || '').trim();
  if (!s) return '';
  if (/^https?:\/\//i.test(s)) return s;
  if (/^(#|\/|\.\.?\/)/.test(s)) return s;
  return '';
}

function pad(n) { return n < 10 ? '0' + n : '' + n; }

function toDate(v) {
  if (v instanceof Date) return v;
  if (typeof v === 'number') return new Date(v);
  return new Date(String(v == null ? '' : v).replace(' ', 'T'));
}

/** 时:分 */
export function fmtTime(v) { const d = toDate(v); return isNaN(d) ? '—' : pad(d.getHours()) + ':' + pad(d.getMinutes()); }

/** 时:分:秒 */
export function fmtTimeS(v) { const d = toDate(v); return isNaN(d) ? '—' : fmtTime(d) + ':' + pad(d.getSeconds()); }

/** YYYY-MM-DD */
export function fmtDate(v) { const d = toDate(v); return isNaN(d) ? '—' : d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate()); }

/** 绝对时刻（含本地时区标注——审计口径）。 */
export function fmtAbs(v) {
  const d = toDate(v);
  if (isNaN(d)) return '时间未知';
  let zone = '';
  try { zone = Intl.DateTimeFormat().resolvedOptions().timeZone || ''; } catch { zone = ''; }
  return fmtDate(d) + ' ' + fmtTimeS(d) + (zone ? ' ' + zone : '');
}

/** 相对时间（N 秒/分/小时/天前——不可解析=空串，不冒充）。 */
export function relAge(ts, nowMs) {
  const t = toDate(ts).getTime();
  if (isNaN(t)) return '';
  const d = Math.max(0, Math.round(((nowMs || Date.now()) - t) / 1000));
  if (d < 60) return d + 's 前';
  if (d < 3600) return Math.round(d / 60) + 'm 前';
  if (d < 86400) return Math.round(d / 3600) + 'h 前';
  return Math.round(d / 86400) + 'd 前';
}

/** <time> 元素（datetime 属性=ISO，title=绝对时刻，正文=相对时间）。 */
export function timeHtml(value) {
  const d = toDate(value);
  const iso = isNaN(d) ? '' : d.toISOString();
  return '<time class="mono" datetime="' + esc(iso) + '" title="' + esc(fmtAbs(d)) + '">' + esc(relAge(value) || '时间未知') + '</time>';
}

/** 数字千分位（不可解析=—）。 */
export function fmtNum(n) {
  if (n == null || isNaN(Number(n))) return '—';
  return Number(n).toLocaleString('zh-CN', { maximumFractionDigits: 1 });
}

/** 字节数人话（B/KB/MB）。 */
export function fmtBytes(n) {
  const v = Number(n);
  if (n == null || isNaN(v)) return '—';
  if (v < 1024) return v + 'B';
  if (v < 1048576) return (v / 1024).toFixed(1) + 'KB';
  return (v / 1048576).toFixed(1) + 'MB';
}

/** 安全截断（超长加省略号——词形保真，不在 esc 后截断避免实体撕裂）。 */
export function trunc(s, n) {
  const t = String(s == null ? '' : s);
  return t.length > n ? t.slice(0, n) + '…' : t;
}
