// transcript-map.test.mjs —— 会话屏 v4（C 批）纯函数锚：转录映射（mapTranscript/blocksOf）、
// 增量分派（nextBatch/nodeKey）、steer|queue 判别（modeForState）、活体行词形（liveRowHtml）、
// 会话行分组与未读（sessionRowHtml/dayGroupOf/sessionsHtml）、空态示例（emptyStateHtml）。
import { test } from 'node:test'
import assert from 'node:assert/strict'
import {
  blocksOf, mapTranscript, nextBatch, nodeKey, modeForState,
  liveRowHtml, sessionRowHtml, sessionsHtml, dayGroupOf, emptyStateHtml,
} from '../src/screens/chat.js'

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

test('blocksOf：text/reasoning 拼接 ∧ tool-call 参数词形（string 透传/object 序列化）', () => {
  const b = blocksOf([
    { type: 'text', text: 'a' }, { type: 'text', text: 'b' },
    { type: 'reasoning', text: 'r' },
    { type: 'tool-call', id: 't1', name: 'read', arguments: '{"p":1}' },
    { type: 'tool-call', id: 't2', name: 'go', arguments: { q: 2 } },
  ])
  assert.equal(b.text, 'a\nb')
  assert.equal(b.reasoning, 'r')
  assert.deepEqual(b.toolCalls.map((t) => [t.id, t.name, t.argsRaw]), [['t1', 'read', '{"p":1}'], ['t2', 'go', '{"q":2}']])
  assert.deepEqual(blocksOf(null), { text: '', reasoning: '', toolCalls: [] })
})

/* ---------- C-04 增量分派 ---------- */

test('nextBatch：首绘全量 ∧ 增量只出未渲染键 ∧ seq 回退不重复', () => {
  const nodes = [
    { kind: 'user', seq: 1, text: 'a' },
    { kind: 'assistant', seq: 2, text: 'b' },
    { kind: 'tool', seq: 3, name: 'read' },
  ]
  assert.equal(nextBatch(nodes, new Set(), true).length, 3, 'firstPaint=全量')
  const painted = new Set(nodes.map(nodeKey))
  assert.deepEqual(nextBatch(nodes, painted, false), [], '全已渲染=零增量')
  const fresh = [...nodes, { kind: 'assistant', seq: 4, text: 'c' }]
  assert.equal(nextBatch(fresh, painted, false).length, 1, '只出新增节点')
  assert.equal(nextBatch(fresh, painted, false)[0].seq, 4)
})

test('nodeKey：seq+kind 复合（同 seq 双族不互撞）∧ 非法值归 0', () => {
  assert.equal(nodeKey({ seq: 5, kind: 'user' }), '5:user')
  assert.notEqual(nodeKey({ seq: 5, kind: 'user' }), nodeKey({ seq: 5, kind: 'assistant' }))
  assert.equal(nodeKey(null), '0:')
})

/* ---------- C-01 steer|queue 判别 ---------- */

test('modeForState：running=true→steer（立即注入）∧ 其余→queue（排队落投）', () => {
  assert.equal(modeForState({ running: true }), 'steer')
  assert.equal(modeForState({ running: false }), 'queue')
  assert.equal(modeForState(null), 'queue')
  assert.equal(modeForState(undefined), 'queue')
  assert.equal(modeForState('x'), 'queue')
})

/* ---------- C-03 活体行词形 ---------- */

test('liveRowHtml：非 running=空 ∧ running 带工具词形+耗时 ∧ queueDepth chip', () => {
  assert.equal(liveRowHtml(null), '')
  assert.equal(liveRowHtml({ running: false }), '')
  const live = liveRowHtml({ running: true, currentTool: 'read', currentToolSince: Date.now() - 5000, queueDepth: 2 })
  assert.ok(live.includes('chat-live'), '活体行容器词形')
  assert.ok(live.includes('read'), '工具名在册')
  assert.ok(live.includes('秒'), '耗时词形在册')
  assert.ok(live.includes('排队 2'), '队列深度 chip')
  const steer = liveRowHtml({ running: true, steerDepth: 1 })
  assert.ok(steer.includes('插话 1'), 'steer 深度 chip')
  const bare = liveRowHtml({ running: true })
  assert.ok(bare.includes('agent 运行中'), '无工具=运行中泛词形')
})

/* ---------- C-06 会话行（未读点/分组） ---------- */

test('dayGroupOf：today（24h 内）∧ earlier（更旧/无效值）', () => {
  const now = Date.now()
  assert.equal(dayGroupOf(now - 3600 * 1000, now), 'today')
  assert.equal(dayGroupOf(now - 30 * 24 * 3600 * 1000, now), 'earlier')
  assert.equal(dayGroupOf(0, now), 'earlier')
})

test('sessionRowHtml：active 词形 ∧ 未读点（updatedAt>highwater 且非当前）∧ 运行中 chip', () => {
  const now = Date.now()
  const s = { sessionId: 's-1', updatedAt: now - 60000, running: true, blank: false }
  const row = sessionRowHtml(s, '', new Map())
  assert.ok(row.includes('data-sess-pick="s-1"'), 'pick 锚')
  assert.ok(row.includes('chat-dot live'), '运行中=live 点')
  assert.ok(row.includes('运行中'), '运行中 chip 词形')
  const unread = sessionRowHtml({ sessionId: 's-2', updatedAt: now, running: false, blank: false }, 's-1', new Map())
  assert.ok(unread.includes('chat-dot unread'), '未读=accent 点')
  const seen = sessionRowHtml({ sessionId: 's-3', updatedAt: now, running: false, blank: false }, 's-1', new Map([['s-3', now + 1]]))
  assert.ok(seen.includes('chat-dot"'), '已见=默认点（无 unread 词形）')
  const active = sessionRowHtml({ sessionId: 's-4', updatedAt: now, running: false, blank: false }, 's-4', new Map())
  assert.ok(active.includes(' active'), '当前会话 active 词形')
})

test('sessionsHtml：today/earlier 分组头 ∧ 空清单诚实词形', () => {
  const now = Date.now()
  const html = sessionsHtml([
    { sessionId: 'new', updatedAt: now - 60000 },
    { sessionId: 'old', updatedAt: now - 40 * 24 * 3600 * 1000 },
  ], '', new Map())
  assert.ok(html.includes('今天') && html.includes('更早'), '分组头词形')
  assert.ok(html.includes('data-sess-pick="new"') && html.includes('data-sess-pick="old"'), '两行都在')
  assert.ok(sessionsHtml([], '', new Map()).includes('会话清单为空'), '空清单词形')
  assert.ok(sessionsHtml(null, '', new Map()).includes('会话清单为空'))
})

/* ---------- C-10 空态示例 ---------- */

test('emptyStateHtml：问候 + 4 枚示例 chips（data-cc-sample 真实可发词形）', () => {
  const html = emptyStateHtml()
  const hits = html.match(/data-cc-sample/g) || []
  assert.equal(hits.length, 4, '示例 chips=4')
  assert.ok(html.includes('chat-empty-hi'), '问候词形')
})
