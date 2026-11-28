// doc-chat.test.mjs —— 原地对话修改面板纯函数（2026-11-28 用户裁决：对象工作台「对话修改」
// 不跳转——直接嵌入相应流程进行修改；提取 yaml 块→diff→人审 Apply）。
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { extractYamlBlock, buildEditPrompt, transcriptRows, rowsHtmlOf } from '../src/screens/doc-chat.js'

test('extractYamlBlock：取最后一个含 apiVersion/kind 的 ```yaml 块', () => {
  const reply = '解释文字在前。\n```yaml\nkind: Provider\nmetadata:\n  name: a\n```\n中间说明。\n```yaml\napiVersion: core/v1\nkind: Workflow\nmetadata:\n  name: b\n```\n尾部。'
  const got = extractYamlBlock(reply)
  assert.match(got, /kind: Workflow/, '取最后一个合法声明块')
  assert.equal(extractYamlBlock('没有代码块'), '', '无块=空串诚实')
  assert.equal(extractYamlBlock('```\nfoo: 1\n```'), '', '非声明块不收')
})

test('buildEditPrompt：自带当前 YAML 全文+意图+输出要求', () => {
  const p = buildEditPrompt({ kind: 'Provider', scope: 'global', name: 'gh' }, 'kind: Provider\n', '换模型')
  assert.match(p, /Provider\/global\/gh/)
  assert.match(p, /修改意图：换模型/)
  assert.match(p, /```yaml\nkind: Provider/, '当前声明在场')
  assert.match(p, /```yaml 代码块/, '输出要求在场')
})

test('transcriptRows/rowsHtmlOf：user 气泡右、assistant 文本+思考折叠、turn 尾词形归一', () => {
  const rows = transcriptRows([
    { type: 'user/message', seq: 1, data: { content: [{ type: 'text', text: '改一下' }], source: { kind: 'user' } } },
    { type: 'assistant/message', seq: 2, data: { message: { content: [{ type: 'reasoning', text: '想想' }, { type: 'text', text: '改好了\n```yaml\nkind: X\n```' }] } } },
    { type: 'turn/end', seq: 3, data: { reason: { kind: 'completed' } } },
    { type: 'turn/end', seq: 4, data: { reason: { kind: 'aborted', reason: { kind: 'disposed' } } } },
  ])
  assert.equal(rows.length, 4)
  const html = rowsHtmlOf(rows)
  assert.match(html, /dc-user/, '用户气泡')
  assert.match(html, /改好了/, '助手文本')
  assert.match(html, /思考/, 'reasoning 折叠')
  assert.match(html, /已停止/, 'aborted→已停止（不再误画处理失败）')
  assert.ok(!/处理失败/.test(html), '嵌套 reason 零误判')
})
