// scheme-chips.test.mjs —— frontmatter 拆分 + scheme 简易标签（2026-11-27）。
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { splitFrontmatter, schemeChips } from '../src/render/md.js'

test('splitFrontmatter：拆 --- 头与正文，缺头/未闭合不误拆', () => {
  assert.deepEqual(splitFrontmatter('---\nname: goal\n---\n# 正文'), { fm: 'name: goal', body: '# 正文' })
  assert.deepEqual(splitFrontmatter('# 无头'), { fm: '', body: '# 无头' })
  assert.deepEqual(splitFrontmatter('---\nname: x\n未闭合'), { fm: '', body: '---\nname: x\n未闭合' })
})

test('schemeChips：标量出 chip、复杂字段折叠、纯文本不渲染 markdown、注入转义', () => {
  const html = schemeChips('name: goal\ndocType: Goal\nspec:\n  title: 标题\nname2: "<img src=x onerror=alert(1)>"')
  assert.match(html, /scheme-chip/, '有简易标签')
  assert.match(html, /scheme-more/, '复杂字段折叠')
  assert.ok(!html.includes('<h1') && !html.includes('<img'), 'scheme 纯文本、注入被转义')
  assert.match(html, /&lt;img/, '注入以实体出现')
})
