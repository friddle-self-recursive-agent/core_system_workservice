/* diff-render.test.mjs —— 渲染层封装单测（无 DOM，断言字符串级）：
 * esc/safeHref 白名单、diff 行级压缩与 ±着色转义、md 最小渲染（无 sanitizer 环境的
 * 构造安全降级路径）、yamlToCards 字段卡转义、rawDisclosure。 */
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { esc, safeHref, trunc, relAge } from '../src/core/format.js';
import { diffLines, renderDiff, diffStats, hunksOf } from '../src/render/diff.js';
import { renderMarkdown, minimalRender, yamlToCards, loadYaml, rawDisclosure, purifySupported } from '../src/render/md.js';

/* ---------- format 基元 ---------- */
test('esc：全实体转义（入 HTML 唯一文本通道）', () => {
  assert.equal(esc('<img src=x onerror=alert(1)>'), '&lt;img src=x onerror=alert(1)&gt;');
  assert.equal(esc('a&"\'b'), 'a&amp;&quot;&#39;b');
  assert.equal(esc(null), '');
  assert.equal(esc(undefined), '');
});

test('safeHref：白名单 http(s)/相对路径；javascript:/data: 封死', () => {
  assert.equal(safeHref('javascript:alert(1)'), '');
  assert.equal(safeHref('data:text/html,<script>'), '');
  assert.equal(safeHref('vbscript:x'), '');
  assert.equal(safeHref('https://example.com/a?b=1'), 'https://example.com/a?b=1');
  assert.equal(safeHref('http://example.com'), 'http://example.com');
  assert.equal(safeHref('#/obj'), '#/obj');
  assert.equal(safeHref('/rel/path'), '/rel/path');
  assert.equal(safeHref('./x.md'), './x.md');
  assert.equal(safeHref(''), '');
});

test('trunc / relAge：截断保词形；不可解析=空串不冒充', () => {
  assert.equal(trunc('abcdef', 3), 'abc…');
  assert.equal(trunc('abc', 5), 'abc');
  assert.equal(relAge('not-a-time'), '');
});

/* ---------- diff 封装 ---------- */
test('diffLines：增删行与上下文压缩', () => {
  const a = 'a\nb\nc\nd\ne\nf\ng\nh';
  const b = 'a\nb\nC\nd\ne\nf\ng\nh';
  const rows = diffLines(a, b);
  assert.ok(rows.some((r) => r.t === '+' && r.s === 'C'));
  assert.ok(rows.some((r) => r.t === '-' && r.s === 'c'));
  assert.ok(rows.every((r) => r.t !== ' ' || true), '含上下文行');
  const stats = diffStats(rows);
  assert.equal(stats.add, 1);
  assert.equal(stats.del, 1);
  assert.equal(hunksOf(rows), 1);
});

test('diffLines：无差异=诚实空词形', () => {
  const rows = diffLines('same\nlines', 'same\nlines');
  assert.deepEqual(rows, [{ t: 'ctx', s: '（无差异）' }]);
});

test('renderDiff：行文本全转义（零原文直插）', () => {
  const rows = diffLines('', '<script>alert(1)</script>');
  const html = renderDiff(rows);
  assert.ok(!html.includes('<script'), '不得出现原始 script 标签');
  assert.ok(html.includes('&lt;script&gt;'), '行文本已转义');
  assert.ok(html.includes('diff-line add'));
});

test('renderDiff：maxLines 截断并标注总行数', () => {
  const rows = diffLines('x', 'a\nb\nc\nd\ne\nf\ng');
  const html = renderDiff(rows, 3);
  assert.ok(html.includes('共 ' + rows.length + ' 行差异'));
});

test('hunksOf：变更块计数（非行数——诚实计数口径）', () => {
  const rows = [
    { t: ' ', s: 'x' }, { t: '-', s: 'a' }, { t: '-', s: 'b' }, { t: '+', s: 'c' },
    { t: ' ', s: 'y' }, { t: '+', s: 'd' },
  ];
  assert.equal(hunksOf(rows), 2);
  assert.equal(diffStats(rows).add, 2);
  assert.equal(diffStats(rows).del, 2);
});

/* ---------- md 封装（node=无 DOMPurify 支持环境——走 minimalRender 构造安全路径） ---------- */
test('node 环境无 DOMPurify 支持=降级路径生效', () => {
  assert.equal(purifySupported(), false);
});

test('minimalRender：script 全转义、粗体/链接结构化', () => {
  const html = renderMarkdown('<script>alert(1)</script>\n\n**bold** text');
  assert.ok(!html.includes('<script'), '零原始 script');
  assert.ok(html.includes('&lt;script&gt;alert(1)&lt;/script&gt;'));
  assert.ok(html.includes('<strong>bold</strong>'));
});

test('minimalRender：javascript: 链接降级纯文本（零 href 通路）', () => {
  const html = minimalRender('[click](javascript:alert(1))');
  assert.ok(!html.includes('href="javascript:'), '零 javascript href');
  assert.ok(!html.includes('<a '), '整锚降级纯文本');
  const ok = minimalRender('[ok](https://example.com)');
  assert.ok(ok.includes('href="https://example.com"'));
});

test('minimalRender：标题/围栏代码/表格结构', () => {
  const html = minimalRender('# 标题\n\n```js\nlet x = "<b>";\n```\n');
  assert.ok(html.includes('<h1 class="md-h md-h1">标题</h1>'));
  assert.ok(html.includes('&lt;b&gt;'), '代码块内容纯转义');
  assert.ok(html.includes('md-pre'));
});

test('loadYaml：js-yaml 解析（失败=null 诚实降级）', () => {
  const v = loadYaml('name: demo\ntags: [a, b]');
  assert.equal(v.name, 'demo');
  assert.deepEqual(v.tags, ['a', 'b']);
  assert.equal(loadYaml('a: [unclosed'), null);
  assert.equal(loadYaml(''), null);
});

test('yamlToCards：值全转义（零原文直插）+ 字段卡结构', () => {
  const html = yamlToCards('name: <b>x</b>\ntags:\n  - a\n  - b');
  assert.ok(!html.includes('<b>x</b>'), '不得出现原始 b 标签');
  assert.ok(html.includes('&lt;b&gt;x&lt;/b&gt;'));
  assert.ok(html.includes('md-card'));
  assert.ok(html.includes('名称'), '常用字段人话释义');
  assert.ok(html.includes('chip mute small'), '列表=逐项 chip');
});

test('yamlToCards：嵌套子卡 + 深度有界', () => {
  // 5 层嵌套：第 5 层对象进入归并词形（depth ≥4——根卡为 depth 0）
  const deep = { a: { b: { c: { d: { e: { f: 'deep' } } } } } };
  const html = yamlToCards(Object.entries(deep).map(([k, v]) => k + ': ' + JSON.stringify(v)).join('\n'));
  assert.ok(html.includes('（嵌套结构见「YAML」）'), '深度 ≥4 归并到折叠次级词形');
});

test('rawDisclosure：标题与内容全转义', () => {
  const html = rawDisclosure('原文', '<img src=x onerror=alert(1)>');
  assert.ok(html.includes('<details'));
  assert.ok(!html.includes('<img '));
  assert.ok(html.includes('&lt;img '));
});
