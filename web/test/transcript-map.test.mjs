// transcript-map.test.mjs —— 会话屏转录映射器（chat.js mapTranscript 纯函数）：
// journal 词形 → 渲染节点族（user 泡/assistant 泡+思考/tool 组按 callId 归组；delta 剥除）。
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { mapTranscript } from '../src/screens/chat.js'

test('mapTranscript：user/assistant/tool 三族映射 ∧ delta 剥除 ∧ callId 归组带结果', () => {
  const nodes = mapTranscript([
    { seq: 8, type: 'assistant/live-chunk', data: {} },
    { seq: 7, type: 'user/message', data: { id: 'm1', source: { kind: 'user' }, content: [{ type: 'text', text: 'hi' }] } },
    { seq: 9, type: 'tool/call', data: { callId: 'c1', name: 'read', arguments: '{"p":1}' } },
    { seq: 10, type: 'tool/result', data: { callId: 'c1', content: 'file body' } },
    { seq: 11, type: 'assistant/message', data: { content: [{ type: 'reasoning', text: 'think' }, { type: 'text', text: 'done' }] } },
  ])
  assert.equal(nodes.length, 3)
  assert.equal(nodes[0].kind, 'user')
  assert.equal(nodes[0].text, 'hi')
  assert.equal(nodes[1].kind, 'tool')
  assert.equal(nodes[1].name, 'read')
  assert.ok(nodes[1].result.includes('file body'))
  assert.equal(nodes[2].kind, 'assistant')
  assert.equal(nodes[2].text, 'done')
  assert.equal(nodes[2].reasoning, 'think')
})

test('mapTranscript：compact-checkpoint 不渲染 ∧ 注入词形标记 ∧ 空页=空', () => {
  const nodes = mapTranscript([
    { seq: 1, type: 'user/message', data: { source: { kind: 'compact-checkpoint' }, content: [{ type: 'text', text: 'x' }] } },
    { seq: 2, type: 'user/message', data: { source: { kind: 'system' }, content: [{ type: 'text', text: '注入' }] } },
  ])
  assert.equal(nodes.length, 1)
  assert.equal(nodes[0].injected, true)
  assert.deepEqual(mapTranscript([]), [])
  assert.deepEqual(mapTranscript(null), [])
})
