// transcript-map.test.mjs —— 会话屏 v4（C 批）纯函数锚：转录映射（mapTranscript/blocksOf）、
// 增量分派（nextBatch/nodeKey）、steer|queue 判别（modeForState）、活体行词形（liveRowHtml）、
// 会话行分组与未读（sessionRowHtml/dayGroupOf/sessionsHtml）、空态示例（emptyStateHtml）。
import { test } from 'node:test'
import assert from 'node:assert/strict'
import {
  blocksOf, mapTranscript, nextBatch, nodeKey, modeForState,
  liveRowHtml, sessionRowHtml, sessionsHtml, dayGroupOf, emptyStateHtml,
} from '../src/screens/chat.js'
import { turnTailOf, todoModel, toolTitle, toolSummary, metaSummary, usageTotal, usageDetail, formatClock } from '../src/screens/chat-nodes.js'

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

test('mapTranscript：compact-checkpoint 不渲染 ∧ 非 user 源=notice 行 ∧ 空页=空', () => {
  const nodes = mapTranscript([
    { seq: 1, type: 'user/message', data: { source: { kind: 'compact-checkpoint' }, content: [{ type: 'text', text: 'x' }] } },
    { seq: 2, type: 'user/message', data: { source: { kind: 'system' }, content: [{ type: 'text', text: '注入' }] } },
  ])
  assert.equal(nodes.length, 1)
  assert.equal(nodes[0].kind, 'notice', '非 user 源=notice 行（D1/D2 词族）')
  assert.equal(nodes[0].text, '注入')
  assert.equal(nodes[0].waking, false, '无 turn/start 前缀=过程注入（非轮触发）')
  assert.deepEqual(mapTranscript([]), [])
  assert.deepEqual(mapTranscript(null), [])
})

/* ---------- D 批（2026-11-27：chat-ds.md 对账锚） ---------- */

test('D1 轮触发行：turn/start 后首条注入=trigger 词形（tool-jobs→后台任务状态更新）', () => {
  const nodes = mapTranscript([
    { seq: 1, type: 'turn/start', data: { turn: 1 } },
    { seq: 2, type: 'user/message', data: { source: { kind: 'tool-jobs', form: 'notice', summary: 'bash (sleep 120) finished [status: completed]' }, content: [{ type: 'text', text: 'background job j1 finished [status: completed]. Output:\nALL DONE' }] } },
  ])
  assert.equal(nodes.length, 1)
  const n = nodes[0]
  assert.equal(n.kind, 'notice')
  assert.equal(n.waking, true, 'turn/start 后未出节点=轮触发')
  assert.equal(n.title, '后台任务状态更新')
  assert.equal(n.summary, 'bash (sleep 120) finished [status: completed]')
  assert.ok(n.text.includes('ALL DONE'))
})

test('D1 词族映射：schedule/subagent-settled(终态变体)/goal/webhook(github)', () => {
  const t = (src) => mapTranscript([{ seq: 1, type: 'user/message', data: { source: src, content: [{ type: 'text', text: 'x' }] } }])[0].title
  assert.equal(t({ kind: 'schedule' }), '自动化任务')
  assert.equal(t({ kind: 'subagent-settled', stopReason: 'completed' }), '子智能体已完成任务')
  assert.equal(t({ kind: 'subagent-settled', stopReason: 'max-tokens' }), '子智能体已达 token 上限')
  assert.equal(t({ kind: 'subagent-settled' }), '子任务状态更新', '未知终态=通用词不冒充')
  assert.equal(t({ kind: 'goal' }), '继续执行目标')
  assert.equal(t({ kind: 'webhook', provider: 'github' }), '收到 GitHub 事件')
  assert.equal(t({ kind: 'webhook' }), '收到外部事件')
  assert.equal(t({ kind: 'cordis-host-runner' }), '插件状态更新')
})

test('D2 注入行：notice form 摘要 ≤120 ∧ producer 标签', () => {
  const rows = mapTranscript([
    { seq: 1, type: 'user/message', data: { source: { kind: 'plugin', plugin: 'my-plug', form: 'notice', summary: 'x'.repeat(200) }, content: [{ type: 'text', text: 'body' }] } },
  ])
  const n = rows[0]
  assert.equal(n.kind, 'notice')
  assert.equal(n.summary.length, 120, 'summary 按上游契约封顶 120')
  assert.equal(n.producer, 'my-plug')
})

test('D3 轮汇总：已完成，用时 ∧ 失败/停止/max-tokens 词形', () => {
  const nodes = mapTranscript([
    { seq: 1, type: 'turn/start', data: { turn: 1 }, time: 1000 },
    { seq: 2, type: 'assistant/message', data: { content: [{ type: 'text', text: 'hi' }] }, time: 2000 },
    { seq: 3, type: 'turn/end', data: { turn: 1, reason: 'complete' }, time: 105000 },
  ])
  const tail = nodes.find((n) => n.kind === 'turntail')
  assert.ok(tail, 'turn/end → turntail 节点')
  assert.equal(tail.durationMs, 104000)
  assert.ok(turnTailOf(tail.reason, tail.durationMs).word.includes('已完成，用时'), '词形=已完成，用时')
  assert.equal(turnTailOf('complete', 104000).word, '已完成，用时 1分44秒')
  assert.equal(turnTailOf('complete', 500).word, '已完成', '<1s 无用时段')
  assert.equal(turnTailOf('interrupted', 0).word, '已停止')
  assert.equal(turnTailOf('error', 0).word, '处理失败')
  const mt = turnTailOf('max-tokens', 0)
  assert.equal(mt.cls, 'maxtok')
  assert.ok(mt.hint.includes('继续'), 'max-tokens 附 hint')
})

test('D4 待办：todo/write → todo 节点 ∧ {done}/{total} 模型', () => {
  const nodes = mapTranscript([
    { seq: 5, type: 'todo/write', data: { todos: [{ content: 'a', status: 'completed' }, { content: 'b', status: 'in_progress' }, { content: 'c', status: 'pending' }] } },
  ])
  const n = nodes[0]
  assert.equal(n.kind, 'todo')
  const m = todoModel(n.todos)
  assert.deepEqual([m.done, m.total, m.active], [1, 3, 'b'])
  assert.deepEqual(todoModel(null), { rows: [], done: 0, total: 0, active: '' })
})

test('D5+病根：tool/result 结果词形=message.content 块 ∧ error→失败态 ∧ meta.stopped→已停止', () => {
  const nodes = mapTranscript([
    { seq: 1, type: 'tool/call', data: { callId: 'c1', name: 'bash', arguments: '{"command":"ls"}' } },
    { seq: 2, type: 'tool/result', data: { callId: 'c1', message: { role: 'tool', content: [{ type: 'text', text: 'file1\nfile2' }] } } },
    { seq: 3, type: 'tool/call', data: { callId: 'c2', name: 'grep', arguments: '{"pattern":"x"}' } },
    { seq: 4, type: 'tool/result', data: { callId: 'c2', message: { role: 'tool', content: [{ type: 'text', text: 'boom' }] }, error: { name: 'ToolError', code: 'E1' } } },
    { seq: 5, type: 'tool/call', data: { callId: 'c3', name: 'bash', arguments: '{"command":"sleep"}' } },
    { seq: 6, type: 'tool/result', data: { callId: 'c3', message: { role: 'tool', content: [{ type: 'text', text: 'part' }] }, meta: { stopped: true } } },
    { seq: 7, type: 'tool/call', data: { callId: 'c4', name: 'bash', arguments: '{"command":"exit3"}' } },
    { seq: 8, type: 'tool/result', data: { callId: 'c4', message: { role: 'tool', content: [{ type: 'text', text: 'out' }] }, meta: { exitCode: 3 } } },
  ])
  const byId = Object.fromEntries(nodes.filter((n) => n.kind === 'tool').map((n) => [n.summary, n]))
  assert.equal(byId['ls'].state, 'ok')
  assert.equal(byId['ls'].result, 'file1\nfile2', '结果词形=message.content 块（2026-11-27 执行中卡死病根）')
  assert.equal(byId['x'].state, 'error')
  assert.equal(byId['sleep'].state, 'stopped')
  assert.equal(byId['exit3'].state, 'error', 'bash 退出码≠0=失败态（上游 terminalFailed 同族）')
  assert.equal(byId['exit3'].meta.exitCode, 3)
  assert.equal(byId['x'].error.code, 'E1')
})

test('D6/D7 专名标题+关键参数摘要+meta 计数', () => {
  assert.equal(toolTitle('bash'), '运行命令')
  assert.equal(toolTitle('read'), '读取')
  assert.equal(toolTitle('grep'), '搜索文件内容')
  assert.equal(toolTitle('todo_write'), '更新任务清单')
  assert.equal(toolTitle('nope'), '')
  assert.equal(toolSummary('read', '{"file_path":"/a/b/c.md","limit":5}'), '/a/b/c.md')
  assert.equal(toolSummary('bash', 'not-json'), 'not-json')
  assert.equal(metaSummary('grep', { total: 12, files: 3 }), '12 处匹配 · 3 个文件')
  assert.equal(metaSummary('grep', { shape: 'paths', total: 5 }), '5 个路径')
  assert.equal(metaSummary('grep', { total: 0 }), '无结果')
  assert.equal(metaSummary('read', { lines: [{ number: 1, text: 'a' }, { number: 2, text: 'b' }] }), '2 行')
  assert.equal(metaSummary('read', {}), '')
  assert.equal(metaSummary('web_fetch', { statusCode: 200 }), 'HTTP 200')
})

test('D10 用量：usage 求和 ∧ 明细词形', () => {
  const u = { inputTokens: 100, outputTokens: 40, cacheReadTokens: 10, cacheWriteTokens: 1, reasoningTokens: 7 }
  assert.equal(usageTotal(u), 151)
  assert.equal(usageTotal(null), 0)
  const d = usageDetail(u)
  assert.ok(d.includes('缓存读取 10'))
  assert.ok(d.includes('（其中推理 7）'))
})

test('D13 时钟：当天 HH:MM ∧ 同年 M月D日 ∧ 跨年 Y年M月D日', () => {
  const now = new Date('2026-11-27T18:00:00')
  const t = (iso) => formatClock(new Date(iso).getTime(), now.getTime())
  assert.match(t('2026-11-27T12:46:00'), /12:46/)
  assert.match(t('2026-10-20T12:46:00'), /10月20日 12:46/)
  assert.match(t('2025-10-20T12:46:00'), /2025年10月20日 12:46/)
  assert.equal(formatClock(0, now.getTime()), '')
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
  assert.ok(steer.includes('追加 1'), 'steer 深度 chip（词形=追加，黑话「插话/steer」不上界面）')
  const bare = liveRowHtml({ running: true })
  assert.ok(bare.includes('深度求索中'), '无工具=深度求索中泛词形（D 批上游对齐）')
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

/* ---------- 2026-11-27 bug 回归锚 ---------- */

test('mapTranscript：tool-only assistant 轮次不产生空回复气泡', () => {
  const nodes = mapTranscript([
    { seq: 1, type: 'user/message', data: { source: { kind: 'user' }, content: [{ type: 'text', text: '查日志' }] } },
    { seq: 2, type: 'assistant/message', data: { content: [{ type: 'tool-call', id: 'c1', name: 'bash', arguments: '{"cmd":"ls"}' }] } },
    { seq: 3, type: 'tool/call', data: { callId: 'c1', name: 'bash', arguments: '{"cmd":"ls"}' } },
    { seq: 4, type: 'tool/result', data: { callId: 'c1', content: 'a.txt' } },
    { seq: 5, type: 'assistant/message', data: { content: [{ type: 'text', text: '结论' }] } },
  ])
  const assistants = nodes.filter((n) => n.kind === 'assistant')
  assert.equal(assistants.length, 1, '只有最终有正文的助手轮次保留')
  assert.equal(assistants[0].text, '结论')
  assert.equal(nodes.filter((n) => n.kind === 'tool').length, 1, '工具调用去重后只有一个节点')
})

test('D5 补：turn/end 后未闭合调用=已停止（中断合成，上游同族）', () => {
  const nodes = mapTranscript([
    { seq: 1, type: 'turn/start', data: { turn: 1 }, time: 1 },
    { seq: 2, type: 'tool/call', data: { callId: 'x', name: 'bash', arguments: '{"command":"tail"}' }, time: 2 },
    { seq: 3, type: 'turn/end', data: { turn: 1, reason: 'error' }, time: 3 },
    { seq: 4, type: 'turn/start', data: { turn: 2 }, time: 4 },
    { seq: 5, type: 'tool/call', data: { callId: 'y', name: 'bash', arguments: '{"command":"live"}' }, time: 5 },
  ])
  const tools = Object.fromEntries(nodes.filter((n) => n.kind === 'tool').map((n) => [n.summary, n]))
  assert.equal(tools['tail'].state, 'stopped', '轮已终=中断遗留')
  assert.equal(tools['live'].state, 'running', '新轮未闭合=真运行中')
})
