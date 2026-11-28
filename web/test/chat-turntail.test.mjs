// chat-turntail.test.mjs —— turn/end reason 归一化修复（2026-11-28：会话账 reason=嵌套对象
// {kind:'completed'} 被 naive String() 成 "[object Object]" 恒画「处理失败」——真 completed/aborted
// 词形全错。session-9bc0f53f 实证：后端 completed、UI 处理失败。
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { normalizeTurnReason, turnTailOf } from '../src/screens/chat-nodes.js'

test('normalizeTurnReason：嵌套对象取 kind，字符串直通', () => {
  assert.equal(normalizeTurnReason({ kind: 'completed' }), 'completed')
  assert.equal(normalizeTurnReason({ kind: 'aborted', reason: { kind: 'disposed' } }), 'aborted')
  assert.equal(normalizeTurnReason('interrupted'), 'interrupted')
  assert.equal(normalizeTurnReason(undefined), '')
})

test('turnTailOf：completed=已完成、aborted/disposed=已停止、真错误才=处理失败', () => {
  assert.match(turnTailOf({ kind: 'completed' }, 5000).word, /已完成/)
  assert.equal(turnTailOf({ kind: 'aborted', reason: { kind: 'disposed' } }).word, '已停止')
  assert.equal(turnTailOf('interrupted').word, '已停止')
  assert.equal(turnTailOf('disposed').word, '已停止')
  assert.equal(turnTailOf('max-tokens').cls, 'maxtok')
  assert.equal(turnTailOf({ kind: 'error' }).word, '处理失败')
  assert.equal(turnTailOf('').word, '已完成')
})
