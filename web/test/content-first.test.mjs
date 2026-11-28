// content-first.test.mjs —— R-UX-10 内容优先批（2026-11-28 用户裁决）：
// 对象页相关任务/会话/运行匹配纯函数 + 会话背景卡 + scheme 弹窗形态 + 正文先行。
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { relatedTasksOf, relatedRunsOf, contentPanel } from '../src/screens/objects.js'
import { sessionTasksOf, sessBacklinksHtml } from '../src/screens/chat-links.js'

const store = {
  tasks: () => [
    { scope: 'global', name: 'bug-gh-issue-3-502-retry-20261127', spec: { title: '修复 issue#3 502 重试', summary: 'gh-issue-trigger 502 重试缺口', source: 'gh-issue-trigger', sourceRef: 'wf/gh-issue-trigger' }, status: { phase: 'running', agentSession: 'sess-abc' } },
    { scope: 'global', name: 'unrelated-task', spec: { title: '别的活' }, status: { phase: 'done' } },
    { scope: 'global', name: 'mentioning-task', spec: { title: '对齐 core_system_instance 文档' }, status: { phase: 'queued' } },
  ],
  events: () => [],
}

test('relatedTasksOf：Task 同名自命中（置顶+带 session），他任务字段提及命中', () => {
  const rows = relatedTasksOf('bug-gh-issue-3-502-retry-20261127', store.tasks())
  assert.equal(rows.length, 1)
  assert.equal(rows[0].self, true)
  assert.equal(rows[0].session, 'sess-abc')
  assert.equal(rows[0].phase, 'running')
  const m = relatedTasksOf('core_system_instance', store.tasks())
  assert.equal(m.length, 1)
  assert.equal(m[0].id, 'global/mentioning-task')
  assert.equal(relatedTasksOf('', store.tasks()).length, 0)
})

test('relatedRunsOf：run.involved 提及命中，短名不猜', () => {
  const runs = [
    { workflow: 'gh-issue-trigger', run: 43, phase: 'converged', involved: 'Task/bug-gh-issue-3-502-retry-20261127' },
    { workflow: 'other-wf', run: 1, phase: 'running', involved: 'Proposal/other' },
  ]
  const rows = relatedRunsOf('bug-gh-issue-3-502-retry-20261127', runs)
  assert.equal(rows.length, 1)
  assert.equal(rows[0].workflow, 'gh-issue-trigger')
  assert.equal(rows[0].run, 43)
  assert.equal(relatedRunsOf('ab', runs).length, 0, '短名(<4)不猜')
})

test('sessionTasksOf/背景卡：title/phase/summary/来源全呈现，任务详情可点', () => {
  const rows = sessionTasksOf(store, 'sess-abc')
  assert.equal(rows.length, 1)
  assert.match(rows[0].title, /502/)
  const html = sessBacklinksHtml(store, 'sess-abc')
  assert.match(html, /执行的任务（本会话由任务拉起——背景如下）/)
  assert.match(html, /来源：/)
  assert.match(html, /data-go="task\//, '任务详情可点')
  assert.match(sessBacklinksHtml(store, 'sess-none'), /无在册关联/, '零编造空态')
  assert.equal(sessBacklinksHtml(store, ''), '')
})

test('contentPanel：markdown 正文先行（零 scheme-chips 首屏），YAML 保持纯文本', () => {
  const mdHtml = contentPanel('Rule', 'r-x.md', '---\ntags: [a]\ndocType: Rule\n---\n# 规则正文\n正文内容')
  assert.match(mdHtml, /data-md-render/)
  assert.ok(!mdHtml.includes('scheme-chip'), 'R-UX-10：markdown 首屏零 frontmatter chips')
  assert.match(mdHtml, /规则正文/, '正文在场')
  const yamlHtml = contentPanel('Provider', 'p', 'kind: Provider\n')
  assert.match(yamlHtml, /<textarea[^>]*readonly/, 'YAML 纯文本保持')
})

test('relatedTasksOf：快照扁平行（无 status/spec 包裹）兜底读顶层字段', () => {
  const flat = [{ scope: 'self-recursive-core', name: 'align-docs-reality-20261127', title: '对齐任务', phase: 'done', agentSession: 'sess-flat', lastAction: '完成' }]
  const rows = relatedTasksOf('align-docs-reality-20261127', flat)
  assert.equal(rows.length, 1)
  assert.equal(rows[0].phase, 'done', '扁平 phase 兜底')
  assert.equal(rows[0].session, 'sess-flat', '扁平 agentSession 兜底')
  assert.equal(rows[0].title, '对齐任务', '扁平 title 兜底')
})

test('defaultSideOpen：#/session/:id 深链=清单默认收起；#/chat 入口=展开', async () => {
  const { defaultSideOpen } = await import('../src/screens/chat.js')
  assert.equal(defaultSideOpen('session-9bc0'), false, '具体会话深链=收起')
  assert.equal(defaultSideOpen(''), true, 'agent 入口=展开')
  assert.equal(defaultSideOpen(undefined), true)
})
