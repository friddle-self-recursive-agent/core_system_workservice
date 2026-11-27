// workflow-detail.test.mjs —— 工作流详情真实数据机械锚（2026-11-27 用户目标）：
// ① Task involved 解析；② 节点会话只认真实 session-*（旧 task.agentSession=taskId 不当 session）；
// ③ history/run/task-journal 三级锚；④ 无运行=不伪造任务/会话；⑤ ConnectorOps 骨架真实投影。
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { wfTaskKey, wfRealSession, wfRunNodesHtml, wfNodeGraphHtml, wfFlowNodesHtml } from '../src/screens/workflows.js'
import { defaultLoopSpec, buildWorkflowYaml, workflowStudioHtml } from '../src/screens/workflow-config.js'

test('wfTaskKey：只解析 Task/scope/name 三段 involved', () => {
  assert.equal(wfTaskKey('Task/self-recursive-core/console-audit-20261126'), 'self-recursive-core/console-audit-20261126')
  assert.equal(wfTaskKey('Proposal/global/x'), '')
  assert.equal(wfTaskKey(''), '')
})

test('wfRunNodesHtml：history→run→task journal 真实 session；旧 taskId 不冒充 session', () => {
  const task = { agentSession: 'self-recursive-core/console-audit-20261126' }
  const rows = [{ taskId: 'self-recursive-core/console-audit-20261126', sessionId: 'session-real-1' }]
  const run = {
    phase: 'Executing', phases: ['Analyzing', 'Executing'],
    involved: 'Task/self-recursive-core/console-audit-20261126',
    history: [{ from: 'Analyzing', to: 'Executing', on: 'advance', at: '2026-11-27T00:00:00Z', sessionId: 'session-hist-1' }],
  }
  const html = wfRunNodesHtml(run, task, rows)
  assert.match(html, /session-hist-1/, 'history.sessionId 优先')
  assert.match(html, /data-go="session\/session-hist-1"/, '真实 session 深链在册')
  assert.doesNotMatch(html, /self-recursive-core\/console-audit-20261126/, '旧 taskId 不当 session 展示')
  assert.match(html, /Task journal 反查/, '锚来源注记在册')

  const run2 = { phase: 'Analyzing', phases: ['Analyzing', 'Executing'], involved: 'Task/self-recursive-core/console-audit-20261126', history: [] }
  const html2 = wfRunNodesHtml(run2, task, rows)
  assert.match(html2, /session-real-1/, 'history 缺席时回退 task journal 反查')
  assert.doesNotMatch(html2, /data-go="chat\//, '不把 taskId 当 agent 会话面')
})

test('wfRunNodesHtml：无真实会话=会话未记录（不编造）', () => {
  const html = wfRunNodesHtml({ phase: 'Idle', phases: ['Idle'], involved: 'Task/a/b' }, { agentSession: 'a/b' }, [])
  assert.match(html, /会话未记录/)
  assert.doesNotMatch(html, /session-/)
})

test('wfNodeGraphHtml：声明相位+真实运行统计；零运行仍展示定义链', () => {
  const wf = { phases: ['Idle', 'Analyzing', 'AwaitingApproval', 'Executing'] }
  const empty = wfNodeGraphHtml(wf, [])
  assert.match(empty, /data-wf-node="Idle"/)
  assert.match(empty, /data-wf-node="Executing"/)
  assert.doesNotMatch(empty, /当前 [1-9]/, '零运行不伪造当前节点')
  const runs = [{ phase: 'Executing', status: 'running', history: [{ to: 'Analyzing' }, { to: 'AwaitingApproval' }, { to: 'Executing' }] }]
  const html = wfNodeGraphHtml(wf, runs)
  assert.match(html, /当前 1/)
  assert.match(html, /到过 1/)
  assert.match(html, /当前 1/)
})

test('wfFlowNodesHtml：ConnectorOps 骨架真实节点；缺席=显式未绑定', () => {
  const wf = { executionFlow: [{ name: 'analyze', title: 'Analyze', unit: 'session' }, { name: 'verify', title: 'Verify', unit: 'session', op: 'audit/run' }] }
  const html = wfFlowNodesHtml(wf)
  assert.match(html, /analyze/)
  assert.match(html, /verify/)
  assert.match(html, /audit\/run/)
  const empty = wfFlowNodesHtml({ executionFlow: null })
  assert.match(empty, /执行面未绑定/)
  assert.doesNotMatch(empty, /wf-flow-node/)
})


test('wfRealSession：同 Task 多次重跑按 startedAt 匹配最近真实会话', () => {
  const run = { involved: 'Task/dev/x', startedAt: '2026-11-27T00:00:12Z' }
  const rows = [
    { taskId: 'dev/x', sessionId: 'session-old', startedAt: Date.parse('2026-11-27T00:00:00Z') },
    { taskId: 'dev/x', sessionId: 'session-new', startedAt: Date.parse('2026-11-27T00:00:20Z') },
  ]
  assert.equal(wfRealSession(run, { agentSession: 'dev/x' }, rows), 'session-new')
  assert.equal(wfRealSession({ involved: 'Task/dev/x', startedAt: '2026-11-27T00:00:01Z' }, null, rows), 'session-old')
  assert.equal(wfRealSession({ involved: 'Task/dev/x' }, null, []), '')
})


test('wfRealSession：Task 多会话但 run 时刻与候选差>10m=不猜', () => {
  const run = { involved: 'Task/dev/x', startedAt: '2026-11-27T01:00:00Z' }
  const rows = [
    { taskId: 'dev/x', sessionId: 'session-old', startedAt: Date.parse('2026-11-26T23:00:00Z') },
    { taskId: 'dev/x', sessionId: 'session-new', startedAt: Date.parse('2026-11-27T00:30:00Z') },
  ]
  assert.equal(wfRealSession(run, null, rows), '')
})

test('wfRunNodesHtml：未到达节点不挂 Task 会话', () => {
  const run = { phase: 'Idle', phases: ['Idle', 'Executing'], involved: 'Task/dev/x', startedAt: '2026-11-27T00:00:00Z' }
  const rows = [{ taskId: 'dev/x', sessionId: 'session-real', startedAt: Date.parse('2026-11-27T00:00:01Z') }]
  const html = wfRunNodesHtml(run, null, rows)
  assert.match(html, /无会话（未到达）/)
  assert.equal((html.match(/data-go="session\/session-real"/g) || []).length, 1, '只有当前/已到节点挂会话深链')
})

test('workflowStudioHtml & buildWorkflowYaml：执行单元三型、节点可加删、双入口 DocApplied、自动审批、BugDoc 收尾可增删、OpenAIProvider 点击即配（2026-11-27 用户裁决）', () => {
  const cfg = defaultLoopSpec('core-system-developer-loop', {})
  assert.equal(cfg.enabled, false, '出厂默认不自动启动')
  assert.equal(cfg.immutableNodes, false, '节点拓扑可加删（我要加节点/去掉验收节点）')
  assert.equal(cfg.undeletable, true, '内置禁止删除')
  assert.equal(cfg.bugDocPolicy, 'invalidate', '临时 BugDoc 默认策略为标记失效归档')
  // 执行单元三型（connectops/agent/service + gate 控制节点；旧 session 词形退役）
  const byName = Object.fromEntries(cfg.phases.map((n) => [n.name, n]))
  assert.equal(byName['doc-apply'].unit, 'agent', 'doc-apply=判断执行单元 agent')
  assert.equal(byName['developer-system'].unit, 'service', 'developer-system=service 调目标系统')
  assert.equal(byName['verify'].unit, 'service', '验收=developer-system 的 Agent 验收 session')
  assert.equal(byName['verify'].op, 'acceptance')
  assert.equal(byName['deploy-system'].unit, 'service')
  assert.equal(byName['report-archive'].unit, 'agent', 'Report 也是 agent')
  assert.ok(byName['report-archive'].tools.length, 'Report 随身工具')
  assert.ok(byName['developer-system'].contextPassing.previousRound, 'service 节点传上一轮详细上下文')
  assert.ok(byName['doc-apply'].refDocs.length && byName['doc-apply'].refTags.length, 'agent 节点 ref 文档/ref tag 配置')
  const yaml = buildWorkflowYaml(cfg, 'global')
  assert.match(yaml, /name: core-system-developer-loop/)
  assert.match(yaml, /on: "entry:DocApplied", to: Planning/)
  assert.match(yaml, /bugDocPolicy: invalidate/)
  assert.match(yaml, /provider: openai-default/)
  assert.match(yaml, /unit: service/, 'YAML 含 service 执行单元')
  assert.match(yaml, /op: acceptance/, '验收 op=acceptance')
  assert.match(yaml, /contextPassing: \{previousRound: true/, '上下文透传声明')
  assert.match(yaml, /permissionRole: role\/developer-system-ops/, '权限 Role 随文档授予')
  const html = workflowStudioHtml({}, 'core-system-developer-loop')
  assert.match(html, /data-open-provider="openai-default"/)
  assert.match(html, /data-develop-collector/)
  assert.match(html, /data-cond-add/)
  assert.match(html, /data-act-add/)
  assert.match(html, /data-node-add/, '节点可加：添加节点按钮')
  assert.match(html, /data-node-del/, '节点可删：删除节点按钮')
  assert.match(html, /data-node="refDocs"/, 'agent 节点点进即配 ref 文档')
  assert.match(html, /data-node="ctxFields"/, 'service 节点上一轮上下文字段配置')
})

