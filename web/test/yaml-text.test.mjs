// yaml-text.test.mjs —— scheme 展示纯文本优先（2026-11-28 用户裁决）：
// 看=text-area 纯文本 YAML 全文，结构化只留编辑态显式切换；markdown 散文不受影响。
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { yamlTextView, dumpYaml } from '../src/render/md.js'
import { contentPanel } from '../src/screens/objects.js'
import { defaultEditorView } from '../src/screens/editor.js'

test('yamlTextView：只读 textarea 全文，esc 注入，行数有界', () => {
  const html = yamlTextView('kind: Provider\nmetadata:\n  name: "<img src=x>"', 'YAML 声明')
  assert.match(html, /<textarea[^>]*readonly/, '只读 textarea')
  assert.match(html, /&lt;img/, '注入以实体出现')
  assert.ok(!html.includes('<img src=x'), '零原文直插')
  const long = yamlTextView(Array.from({ length: 200 }, (_, i) => 'k' + i + ': v').join('\n'))
  const rows = Number(/rows="(\d+)"/.exec(long)[1])
  assert.ok(rows <= 48, '行数上界 48，实得 ' + rows)
  assert.match(yamlTextView('a: 1'), /rows="4"/, '行数下界 4')
})

test('dumpYaml：对象 → YAML 文本；null/异常=空串', () => {
  const s = dumpYaml({ name: 'Provider', spec: { auth: { account: 'a' } } })
  assert.match(s, /^name: Provider/m)
  assert.match(s, /account: a/m)
  assert.equal(dumpYaml(null), '')
  assert.equal(dumpYaml(undefined), '')
})

test('contentPanel：YAML 对象默认纯文本 textarea（结构化卡退役），markdown 保持渲染', () => {
  const yamlHtml = contentPanel('Provider', 'openai-2', 'apiVersion: core/v1\nkind: Provider\nmetadata:\n  name: openai-2\n')
  assert.match(yamlHtml, /<textarea[^>]*readonly/, 'YAML 默认=只读 textarea 纯文本')
  assert.ok(!yamlHtml.includes('md-cards') && !yamlHtml.includes('scheme-chip'), '看态零结构化')
  const mdHtml = contentPanel('Rule', 'r-doc.md', '---\ntags: [a]\n---\n# 规则正文\n')
  assert.match(mdHtml, /data-md-render/, 'markdown 保持渲染（散文=文本形态）')
  assert.match(mdHtml, /md-raw/, '原文折叠保留')
  assert.match(contentPanel('Goal', 'g', '   \n'), /暂无正文/, '空正文诚实空态')
})

test('defaultEditorView：YAML 对象默认 raw（纯文本），markdown 默认 render', () => {
  assert.equal(defaultEditorView('Provider', 'openai-2'), 'raw')
  assert.equal(defaultEditorView('Workflow', 'gh-issue-trigger'), 'raw')
  assert.equal(defaultEditorView('Rule', 'r-x.md'), 'render')
  assert.equal(defaultEditorView('Proposal', 'p-1'), 'render')
})
