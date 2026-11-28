// meta-header.test.mjs —— 文档元数据公用 header（2026-11-28 用户裁决：metadatainfo 侧车上屏——
// tag/创建时间/提交 agent/提交人/相应 refs 公用一个 header 组件）。
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { frontMetaOf, metaHeaderHtml } from '../src/render/meta-header.js'

const FM = `---
name: style
docType: Rule
scope: core_system
title: 统一样式规范
tags:
  - {tag: console, tag_type: surface}
  - {tag: style, tag_type: domain}
importance: 93
ownerRefs: [{kind: ServiceDesc, name: core_system}]
---
# 正文
`

test('frontMetaOf：侧车词形全收（tags 对/裸串、importance、ownerRefs 对/串、docType/scope/title）', () => {
  const m = frontMetaOf(FM, 'Rule', 'core_system', 'style.md')
  assert.equal(m.name, 'style')
  assert.equal(m.docType, 'Rule')
  assert.equal(m.title, '统一规范标题占位'.slice(0, 0) || '统一样式规范')
  assert.equal(m.importance, 93)
  assert.deepEqual(m.tags.map(t => t.tag).sort(), ['console', 'style'])
  assert.equal(m.tags.find(t => t.tag === 'console').type, 'surface')
  assert.equal(m.ownerRefs.length, 1)
  assert.equal(m.ownerRefs[0].kind, 'ServiceDesc')
  const m2 = frontMetaOf('---\ntags: [a, domain:b]\nownerRefs:\n  - ConnectorOps/x\n---\nbody', 'Rule')
  assert.deepEqual(m2.tags.map(t => t.tag), ['a', 'b'])
  assert.deepEqual(m2.tags[1].type, 'domain')
  assert.equal(m2.ownerRefs[0].kind, 'ConnectorOps')
  assert.equal(m2.ownerRefs[0].name, 'x')
  assert.equal(frontMetaOf('apiVersion: core/v1\nkind: Provider\n').tags.length, 0, '纯 YAML 对象零 frontmatter=空 meta')
})

test('metaHeaderHtml：身份/账/tags/importance/归属四行；缺席整行不渲染；注入转义', () => {
  const m = frontMetaOf(FM, 'Rule', 'core_system', 'style.md')
  const full = metaHeaderHtml(m, { rv: 12, createdAt: '2026-11-26T10:00:00Z', lastAt: '2026-11-28T08:00:00Z', lastBy: 'core-system-developer-loop_89' })
  assert.match(full, /meta-header/)
  assert.match(full, /统一样式规范/, 'title 在场')
  assert.match(full, /r12/)
  assert.match(full, /最近提交/, '提交行在场')
  assert.match(full, /core-system-developer-loop_89/, '提交人/agent 词形在场')
  assert.match(full, /重要度 93/)
  assert.match(full, /#console/, 'tag chip 在场')
  assert.match(full, /data-go="[^"]*tag\/console/, 'tag 可点聚合页')
  assert.match(full, /data-go="obj\//, '归属引用可点对象工作台')
  const bare = metaHeaderHtml(frontMetaOf('kind: Provider\n'), { rv: 1 })
  assert.ok(!bare.includes('最近提交') && !bare.includes('重要度'), '缺席字段整行不渲染（零编造）')
  const xss = metaHeaderHtml(frontMetaOf('---\ntitle: "<img src=x>"\n---\n'), {})
  assert.ok(!xss.includes('<img src=x>'), 'esc 注入转义')
})
