/* diff.js —— 行级 diff（npm diff 库）→ ±着色 HTML。
 * 语义对齐旧 console-app.diffLines：变更块 ±2 行上下文压缩、其余折叠 gap；
 * 一切行文本 esc 后入 HTML（安全红线——diff 内容是用户文档，零原文直插）。
 * 纯字符串处理零 DOM——node 单测机械直测口径。 */

import { diffLines as libDiffLines } from 'diff';
import { esc, trunc } from '../core/format.js';

/**
 * 行级 diff：返回 [{t:'+'|'-'|' ' , s:行文本}, …]（含 gap 折行 {t:'gap'}）。
 * 两侧一致=单条 '（无差异）' 上下文行（诚实空 diff 词形）。
 */
export function diffLines(oldStr, newStr) {
  const parts = libDiffLines(String(oldStr == null ? '' : oldStr), String(newStr == null ? '' : newStr));
  const rows = [];
  for (const p of parts) {
    const lines = String(p.value == null ? '' : p.value).split('\n');
    if (lines.length && lines[lines.length - 1] === '') lines.pop();
    const t = p.added ? '+' : p.removed ? '-' : ' ';
    for (const s of lines) rows.push({ t, s });
  }
  return compress(rows);
}

/** 只保留变更块 ±2 行上下文，其余折叠为 gap 行（首屏禁全量裸转储）。 */
function compress(rows) {
  const keep = new Array(rows.length).fill(false);
  rows.forEach((r, i) => {
    if (r.t !== ' ') {
      for (let j = Math.max(0, i - 2); j <= Math.min(rows.length - 1, i + 2); j++) keep[j] = true;
    }
  });
  const out = [];
  let last = -3;
  rows.forEach((r, i) => {
    if (!keep[i]) return;
    if (i - last > 1 && r.t === ' ') out.push({ t: 'gap', s: '⋯' });
    out.push(r);
    last = i;
  });
  if (!out.length) out.push({ t: 'ctx', s: '（无差异）' });
  return out;
}

/** diff 行 → HTML（±着色；maxLines 截断并标注总行数——词形诚实）。 */
export function renderDiff(lines, maxLines = 200) {
  const cap = Number(maxLines) || 200;
  const html = lines.slice(0, cap).map((l) => {
    if (l.t === 'gap') return '<div class="diff-line ctx">   ⋯</div>';
    const sign = l.t === '+' ? '+' : l.t === '-' ? '-' : ' ';
    const cls = l.t === '+' ? 'add' : l.t === '-' ? 'del' : 'ctx';
    return '<div class="diff-line ' + cls + '"><span class="d-sign">' + sign + '</span>' + esc(trunc(l.s, 500)) + '</div>';
  }).join('');
  const tail = lines.length > cap ? '<div class="diff-line ctx">… 共 ' + lines.length + ' 行差异</div>' : '';
  return '<div class="diff">' + html + tail + '</div>';
}

/** 增删行数统计（{add, del}）。 */
export function diffStats(lines) {
  let add = 0;
  let del = 0;
  (lines || []).forEach((l) => { if (l.t === '+') add += 1; else if (l.t === '-') del += 1; });
  return { add, del };
}

/** 变更块数（「Apply（N 处变更）」的 N=块数非行数——诚实计数）。 */
export function hunksOf(lines) {
  let n = 0;
  let inHunk = false;
  (lines || []).forEach((l) => {
    if (l.t === '+' || l.t === '-') {
      if (!inHunk) { n += 1; inHunk = true; }
    } else {
      inHunk = false;
    }
  });
  return n;
}
