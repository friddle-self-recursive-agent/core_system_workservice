// chat-refs.test.mjs —— B 批（2026-11-27，chat-ds.md §二）纯函数锚：引用抽取/引用条/弹窗目标/
// 代码块头横幅（mdCodeBlock）。
import { test } from 'node:test'
import assert from 'node:assert/strict'
import {
  looksLikePath, extractFilePaths, extractDocRefs, extractRefs, chunkRefs,
  refsStripHtml, refTarget,
} from '../src/screens/chat-refs.js'
import { mdCodeBlock } from '../src/render/md.js'

/* ---------- B1 路径词形 ---------- */

test('looksLikePath：相对/家目录/扩展名 ∧ 空白与杂字符拒绝', () => {
  assert.equal(looksLikePath('concepts/cli.md'), true)
  assert.equal(looksLikePath('~/Project/x/y.go'), true)
  assert.equal(looksLikePath('../a/b.ts'), true)
  assert.equal(looksLikePath('readme'), false)
  assert.equal(looksLikePath('foo bar.md'), false, '含空白拒绝')
  assert.equal(looksLikePath('a/b<c'), false, '杂字符拒绝')
  assert.equal(looksLikePath(''), false)
})

test('extractFilePaths：保序去重封顶 ∧ 尾标点剥离', () => {
  const rows = extractFilePaths('先看 concepts/cli.md，再看 concepts/cli.md 与 core/scheme.yaml。', 2)
  assert.deepEqual(rows, ['concepts/cli.md', 'core/scheme.yaml'])
  assert.equal(extractFilePaths('x（见 a/b/c.md）').length >= 1, true)
  assert.deepEqual(extractFilePaths('', 8), [])
})

test('extractDocRefs：三段词识别（Doc/DesignDoc/BugDoc）∧ 去重封顶', () => {
  const rows = extractDocRefs('读 Doc/global/cli 与 DesignDoc/core/global/scheme，再读 Doc/global/cli。', 12)
  assert.equal(rows.length, 2, '同 ref 去重')
  assert.deepEqual([rows[0].kind, rows[0].scope, rows[0].name], ['Doc', 'global', 'cli'])
  assert.equal(rows[1].type, 'obj')
  assert.deepEqual(extractDocRefs('no refs here'), [])
})

test('extractRefs：节点文本+tool args file_path 收集 ∧ docs/files 分桶', () => {
  const refs = extractRefs([
    { kind: 'assistant', text: '见 Doc/global/cli 与 concepts/cli.md' },
    { kind: 'tool', args: '{"file_path":"/tmp/x/loop-91.md","limit":120}' },
    { kind: 'tool', args: 'not-json' },
    null,
  ])
  assert.ok(refs.docs.some((d) => d.name === 'cli' && d.kind === 'Doc'))
  assert.ok(refs.files.some((f) => f.path === 'concepts/cli.md'))
  assert.ok(refs.files.some((f) => f.path === '/tmp/x/loop-91.md'))
})

test('chunkRefs：ContextPack chunks → obj 引用 ∧ 去重封顶8 ∧ score 保留', () => {
  const rows = chunkRefs([
    { kind: 'Doc', scope: 'global', name: 'a', score: 87 },
    { kind: 'Doc', scope: 'global', name: 'a', score: 80 },
    { kind: '', name: 'x' },
  ])
  assert.equal(rows.length, 1)
  assert.equal(rows[0].score, 87)
  assert.equal(rows[0].ctx, true)
})

/* ---------- B2 引用条 ---------- */

test('refsStripHtml：零引用=空串 ∧ 图2 词形（名称+打开 pill）∧ 分组头', () => {
  assert.equal(refsStripHtml(null), '')
  assert.equal(refsStripHtml({}), '')
  const html = refsStripHtml({
    docs: [{ type: 'obj', kind: 'Doc', scope: 'global', name: 'cli', label: 'Doc/global/cli' }],
    files: [{ type: 'file', path: 'concepts/cli.md', label: 'concepts/cli.md' }],
    chunks: [{ type: 'obj', kind: 'Doc', scope: 'global', name: 'z', label: 'Doc/global/z', score: 61, ctx: true }],
  })
  assert.ok(html.includes('chat-refs'), '条容器')
  assert.ok(html.includes('concepts/cli.md'))
  assert.ok(html.includes('打开'), '图2 词形 pill')
  assert.ok(html.includes('引用'))
  assert.ok(html.includes('上下文'))
  assert.ok(html.includes('61'), 'chunk 分数在册')
})

test('refTarget：g/i 解析 ∧ 越界=null 不冒充', () => {
  const refs = { docs: [{ type: 'obj', kind: 'Doc', scope: 'global', name: 'cli', label: 'L' }], files: [], chunks: [{ type: 'obj', kind: 'Doc', scope: 'global', name: 'z', label: 'Z' }] }
  assert.equal(refTarget(refs, 'cited', 0).name, 'cli')
  assert.equal(refTarget(refs, 'cited', 1), null)
  assert.equal(refTarget(refs, 'chunk', 0).name, 'z')
  assert.equal(refTarget(refs, 'chunk', 9), null)
  assert.equal(refTarget(refs, 'other', 0), null)
  assert.equal(refTarget(refs, 'cited', 'nan'), null)
})

/* ---------- D11 代码块 ---------- */

test('mdCodeBlock：语言标签|缺省「代码块」∧ 文本转义 ∧ 复制/换行按钮词形', () => {
  const html = mdCodeBlock('let x = "<b>";\nrun()', 'js')
  assert.ok(html.includes('md-code-block'), '卡容器')
  assert.ok(html.includes('>js<'), '语言标签')
  assert.ok(html.includes('&lt;b&gt;'), '内容纯转义')
  assert.ok(html.includes('md-code-copy'), '复制钮')
  assert.ok(html.includes('md-code-wrap'), '换行钮')
  assert.ok(html.includes('wrap'), '默认换行态')
  const plain = mdCodeBlock('x', '')
  assert.ok(plain.includes('代码块'), '缺省=代码块标签')
})

test('looksLikePath 活体纠偏：纯数字段串（HTTP 码）拒绝 ∧ 无扩展名目录段拒绝', () => {
  assert.equal(looksLikePath('200/200/200/405/401'), false, 'HTTP 码串不是路径')
  assert.equal(looksLikePath('189/0'), false)
  assert.equal(looksLikePath('acceptance/loop-'), false, '截断无扩展名拒绝')
  assert.equal(looksLikePath('core/scheme.yaml'), true)
  assert.equal(looksLikePath('~/x/loop-next.md'), true)
  assert.equal(looksLikePath('/abs/dir/file.go'), true)
})
