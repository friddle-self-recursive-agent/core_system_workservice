/* workflow-config.js —— 以人为本 Workflow 全节点可视化配置工作台 + OpenAIProvider 点击即配
 * + 数据库独立文档导出配置面板（2026-11-27 用户裁决）：
 *   1. 「Workflow 文档 + Scheme = 声明层，代码 = 通用解释器」：原独立临时 BugDoc 闭环失效，
 *      所有节点参数、日志过滤、ConnectorOps 收集、审批模式（human|agent|auto）、验收条件增删、
 *      收尾 postActions（临时 BugDoc 标记失效 invalidate / 删除 delete / 保留不删除 keep，可增删）
 *      全部声明在 Workflow YAML 与 Scheme Adapter 层，前端可视化编辑并经 POST /v1/apply 落库。
 *   2. 点击任意 Provider 徽章（默认 openai-default）弹出 OpenAIProvider 配置抽屉：
 *      配置账号/密钥引用、baseUrl、模型清单及按任务/ConnectorOps 智商档位（low/standard/high）映射。
 *   3. 文档导出独立配置面板（GET/PUT /v1/docs-export/config）：数据库 docs_export_config 独立存储。 */

import { esc } from '../core/format.js';
import { chip, chipGo, docRoute } from '../render/chips.js';
import { t as tr } from '../core/i18n.js';

/** 默认 core-system-developer-loop 8 节点解释器声明骨架 */
export function defaultLoopSpec(name = 'core-system-developer-loop', existing = {}) {
  const s = existing && typeof existing === 'object' ? existing : {};
  const logSrc = s.logSource || {};
  const collector = logSrc.collector || {};
  const timer = logSrc.timer || (s.triggers && s.triggers.timer) || {};
  const approval = s.approval || {};
  const sysBind = s.systemBindings || {};
  const devSys = sysBind.developerSystem || {};
  const depSys = sysBind.deploySystem || {};
  const devConv = devSys.conventions || {};
  const ver = s.verification || {};
  const reg = s.regression || {};
  const llm = s.llm || {};

  const defaultConditions = [
    { id: 'lint-rules', kind: 'code', title: '代码 Lint 与规则审核 (R-PUB-1 / R-DOC-6 / R-ENG-7 ≤800行)', command: 'connectops-code-audit/audit', required: true },
    { id: 'unit-tests', kind: 'test', title: '后端与前端单元/集成测试全绿 (go test ./... & npm test)', command: 'go test ./... && npm test', required: true },
    { id: 'live-probe', kind: 'agent', title: '活体探针与页面无假数据/无报错验证', command: 'agent-verify', required: true },
  ];
  const defaultPostActions = [
    { id: 'bugdoc-lifecycle', kind: 'doc-lifecycle', targetDocType: 'BugDoc', action: reg.bugDocPolicy || 'invalidate', enabled: true, title: '临时 BugDoc 收尾处理（标记失效 invalidate / 删除 delete / 保留 keep，可增删此项）' },
    { id: 'status-insync', kind: 'status-writeback', targetPhase: 'InSync', enabled: true, title: '回写 ServiceInstance.status.phase = InSync（清零漂移）' },
  ];
  const defaultPhases = [
    { name: 'log-event', phase: 'Idle', title: '1. 监听日志与 Event', unit: 'connectops', connector: collector.connector || 'connectops-log-collector', op: collector.op || 'collect-logs', provider: llm.provider || 'openai-default', modelTier: 'low', entryEvents: ['BugReport', 'MetricThreshold', 'ExternalChanged', 'ReconcileTick', 'LoopRequested'] },
    { name: 'agent-judge', phase: 'Analyzing', title: '2. Agent 判断（评估修复价值/产出方案或 BugDoc）', unit: 'agent', op: 'judge-issue', provider: llm.provider || 'openai-default', modelTier: 'high', refRules: ['rule/rule_eng', 'rule/rule_doc'] },
    { name: 'approval-gate', phase: 'AwaitingApproval', title: '3. 审批 / 自动判断审批（控制节点）', unit: 'gate', op: 'approval-check' },
    { name: 'doc-apply', phase: 'Planning', title: '4. Doc Apply（对话改文档/BugDoc 第二切入点）', unit: 'agent', op: 'apply-doc', provider: llm.provider || 'openai-default', modelTier: 'standard', refDocs: ['goal'], refRules: ['rule/rule_doc', 'rule/rule_eng'], refTags: ['audit', 'fix'], tagRules: ['rule/tag_default'], entryEvents: ['DocApplied'] },
    { name: 'developer-system', phase: 'Executing', title: '5. developer-system 研发（service 内部 agent 执行）', unit: 'service', connector: devSys.connector || 'connectops-developer', op: devSys.op || 'develop', service: devSys.service || 'developer-system', contextPassing: { previousRound: true, fields: ['judge', 'plan', 'diff', 'approval'] }, permissionRole: 'role/developer-system-ops' },
    { name: 'verify', phase: 'Verifying', title: '6. 验收（developer-system Agent 验收 session）', unit: 'service', service: devSys.service || 'developer-system', op: 'acceptance', contextPassing: { previousRound: true, fields: ['diff', 'verifyConditions'] } },
    { name: 'deploy-system', phase: 'Deploying', title: '7. deploy-system 发版（service 内部 agent 执行）', unit: 'service', connector: depSys.connector || 'connectops-deploy', op: depSys.op || 'release', service: depSys.service || 'deploy-system', contextPassing: { previousRound: true, fields: ['verifyReport', 'release'] }, permissionRole: 'role/deploy-system-ops' },
    { name: 'report-archive', phase: 'Reporting', title: '8. 总结归档（解释执行 postActions + 回写 InSync）', unit: 'agent', op: 'archive-report', provider: llm.provider || 'openai-default', modelTier: 'low', tools: ['findings-read', 'doc-draft', 'tag-write'] },
  ];

  return {
    name,
    enabled: s.enabled != null ? Boolean(s.enabled) : false,
    immutableNodes: Boolean(s.immutableNodes),   // 2026-11-27 用户裁决：节点可加删（内置主干也不恒锁定）——前端节点表加节点/删节点/改参数，draft→apply 落账
    undeletable: name === 'core-system-developer-loop' || name === 'default-task' ? true : Boolean(s.undeletable),
    servicesInvolved: Array.isArray(s.services && s.services.involved) ? s.services.involved.join(', ') : 'self-recursive-core, developer-system, deploy-system',
    servicesMode: (s.services && s.services.mode) || 'act',
    provider: llm.provider || 'openai-default',
    defaultModel: llm.model || 'reasoning-tier',
    globalRuleRefs: Array.isArray(s.globalRuleRefs) ? s.globalRuleRefs.join(', ') : 'rule/rule_eng, rule/rule_doc, rule/rule_ux, rule/rule_ctx, rule/rule_pub',
    logEvents: Array.isArray(logSrc.events) ? logSrc.events.join(', ') : 'BugReport, MetricThreshold, ExternalChanged, ReconcileTick, LoopRequested',
    logMinLevel: logSrc.minLevel || 'warn',
    logServices: Array.isArray(logSrc.services) ? logSrc.services.join(', ') : 'self-recursive-core',
    logKeyword: logSrc.keywordMatch || 'ERROR|PANIC|FAIL|drift|unhealthy',
    timerEnabled: timer.enabled != null ? Boolean(timer.enabled) : true,
    timerInterval: timer.interval || timer.resync || '30m',
    timerCron: timer.cron || '*/30 * * * *',
    collectorConnector: collector.connector || 'connectops-log-collector',
    collectorOp: collector.op || 'collect-logs',
    collectorQuery: collector.query || 'level>=warn',
    docApplyEntryEnabled: true,
    docApplyDocTypes: 'Goal, Desc, Rule, WorkflowDesign, ImplDoc, VerifyDoc, BugDoc',
    approvalMode: approval.mode || 'agent',
    approvalCriteria: approval.autoCriteria || '单测全绿、符合全局 rule_eng/rule_doc 规范且不涉及破坏性高危删库操作时自动批准放行',
    approvalMaxRisk: approval.maxRiskLevel || 'low',
    devService: devSys.service || 'developer-system',
    devResolution: devSys.resolution || 'default',
    devConnector: devSys.connector || 'connectops-developer',
    devOp: devSys.op || 'develop',
    devBranchPrefix: devConv.branchPrefix || 'agent/',
    devWorktreeDir: devConv.worktreeDir || 'worktrees/',
    devMaxLines: devConv.maxFileLines || 800,
    deployService: depSys.service || 'deploy-system',
    deployResolution: depSys.resolution || 'default',
    deployConnector: depSys.connector || 'connectops-deploy',
    deployOp: depSys.op || 'release',
    verifyMode: ver.mode || 'both',
    verifyConnector: ver.connector || 'connectops-code-audit',
    verifyOp: ver.op || 'audit',
    verifyOnFail: ver.onFail || 'retry-develop',
    conditions: Array.isArray(ver.conditions) && ver.conditions.length ? ver.conditions : defaultConditions,
    bugDocPolicy: reg.bugDocPolicy || (reg.cleanTempBugDoc === true ? 'delete' : 'invalidate'),
    statusWriteback: reg.statusWriteback != null ? Boolean(reg.statusWriteback) : true,
    postActions: Array.isArray(reg.postActions) && reg.postActions.length ? reg.postActions : defaultPostActions,
    phases: (Array.isArray(s.phases) && s.phases.length && typeof s.phases[0] === 'object'
      ? s.phases.map((n) => (n && n.unit === 'session' ? { ...n, unit: 'agent' } : n))   // 旧 session 词形退役→agent（workflow.md §1.3）
      : defaultPhases),
  };
}

/** 将结构化配置序列化为标准 Workflow YAML 文档（解释器唯一消费源） */
export function buildWorkflowYaml(cfg, scope = 'global') {
  const splitList = (v) => String(v || '').split(',').map((x) => x.trim()).filter(Boolean);
  const inv = splitList(cfg.servicesInvolved);
  const rules = splitList(cfg.globalRuleRefs);
  const logEvs = splitList(cfg.logEvents);
  const logSvcs = splitList(cfg.logServices);
  const docTypes = splitList(cfg.docApplyDocTypes);
  const q = (s) => JSON.stringify(String(s == null ? '' : s));

  const condLines = (cfg.conditions || []).map((c, idx) =>
    `      - {id: ${q(c.id || ('cond-' + (idx + 1)))}, kind: ${q(c.kind || 'code')}, title: ${q(c.title || '')}, command: ${q(c.command || '')}, required: ${Boolean(c.required)}}`
  ).join('\n');

  const actionLines = (cfg.postActions || []).map((a, idx) =>
    `      - {id: ${q(a.id || ('act-' + (idx + 1)))}, kind: ${q(a.kind || 'doc-lifecycle')}, targetDocType: ${q(a.targetDocType || 'BugDoc')}, action: ${q(a.action || cfg.bugDocPolicy || 'invalidate')}, enabled: ${Boolean(a.enabled)}, title: ${q(a.title || '')}}`
  ).join('\n');

  /* 节点行序列化（2026-11-27 用户裁决：执行单元三型 connectops|agent|service + gate 控制节点；
   * 节点表可加删——phases 从节点表实时生成，YAML=解释器权威源）。 */
  const nodeLine = (n) => {
    const p = [`name: ${n.name || 'node'}`, `phase: ${n.phase || 'Idle'}`, `title: ${q(n.title || n.name || '')}`, `unit: ${n.unit || 'connectops'}`];
    if (n.connector) p.push(`connector: ${n.connector}`);
    if (n.op) p.push(`op: ${n.op}`);
    if (n.service) p.push(`service: ${n.service}`);
    if (n.provider) p.push(`provider: ${n.provider}`);
    if (n.model) p.push(`model: ${q(n.model)}`);
    if (n.modelTier) p.push(`modelTier: ${n.modelTier}`);
    if (n.permissionRole) p.push(`permissionRole: ${n.permissionRole}`);
    ['refDocs', 'refRules', 'refTags', 'tagRules', 'tools', 'entryEvents'].forEach((k) => {
      if (Array.isArray(n[k]) && n[k].length) p.push(`${k}: [${n[k].map(q).join(', ')}]`);
    });
    if (n.contextPassing && (n.contextPassing.previousRound || (Array.isArray(n.contextPassing.fields) && n.contextPassing.fields.length))) {
      p.push(`contextPassing: {previousRound: ${n.contextPassing.previousRound !== false}, fields: [${(n.contextPassing.fields || []).map(q).join(', ')}]}`);
    }
    return `    - {${p.join(', ')}}`;
  };
  const phaseLines = (cfg.phases || []).map(nodeLine).join('\n');
  const nodeMappingLines = (cfg.phases || []).map((n) =>
    `      ${n.name}: {tier: ${n.modelTier || 'standard'}${n.model ? ', model: ' + q(n.model) : ''}}`
  ).join('\n');

  return `apiVersion: core/v1
kind: Workflow
metadata:
  name: ${cfg.name}
  scope: ${scope}
  labels:
    core/origin: builtin
spec:
  enabled: ${Boolean(cfg.enabled)}
  immutableNodes: ${Boolean(cfg.immutableNodes)}
  undeletable: ${Boolean(cfg.undeletable)}
  connector: connector-ops/${cfg.name}
  services:
    involved: [${inv.map(q).join(', ')}]
    mode: ${cfg.servicesMode || 'act'}
  llm:
    provider: ${cfg.provider || 'openai-default'}
    model: ${q(cfg.defaultModel || 'reasoning-tier')}
    tier: high
    nodeMapping:
${nodeMappingLines}
  globalRuleRefs: [${rules.map(q).join(', ')}]
  logSource:
    events: [${logEvs.map(q).join(', ')}]
    minLevel: ${cfg.logMinLevel || 'warn'}
    services: [${logSvcs.map(q).join(', ')}]
    keywordMatch: ${q(cfg.logKeyword || '')}
    timer:
      enabled: ${Boolean(cfg.timerEnabled)}
      interval: ${q(cfg.timerInterval || '30m')}
      cron: ${q(cfg.timerCron || '*/30 * * * *')}
    collector:
      connector: ${cfg.collectorConnector || 'connectops-log-collector'}
      op: ${cfg.collectorOp || 'collect-logs'}
      query: ${q(cfg.collectorQuery || 'level>=warn')}
      autoDevelopOnMissing: true
  triggers:
    events:
      types: [${logEvs.concat(cfg.docApplyEntryEnabled ? ['DocApplied'] : []).map(q).join(', ')}]
      minLevel: ${cfg.logMinLevel || 'warn'}
    timer:
      resync: ${q(cfg.timerInterval || '30m')}
      enabled: ${Boolean(cfg.timerEnabled)}
      interval: ${q(cfg.timerInterval || '30m')}
      cron: ${q(cfg.timerCron || '*/30 * * * *')}
  entryPoints:
    - {node: log-event, phase: Idle, enabled: true, events: [${logEvs.map(q).join(', ')}], description: "入口 1：监听日志/运行期 Bug/定时巡检事件从头流转"}
    - {node: doc-apply, phase: Planning, enabled: ${Boolean(cfg.docApplyEntryEnabled)}, events: ["DocApplied"], docTypes: [${docTypes.map(q).join(', ')}], description: "入口 2：用户对话修改文档或产出临时 BugDoc 触发 DocApplied 直接从节点 4 切入"}
  ingest:
    dedup: {window: 5m}
    group: {by: service, max: 20}
    filterLevels: [warn, error, fatal]
  approval:
    mode: ${cfg.approvalMode || 'agent'}
    autoCriteria: ${q(cfg.approvalCriteria || '')}
    maxRiskLevel: ${cfg.approvalMaxRisk || 'low'}
  systemBindings:
    developerSystem:
      service: ${cfg.devService || 'developer-system'}
      resolution: ${cfg.devResolution || 'default'}
      connector: ${cfg.devConnector || 'connectops-developer'}
      op: ${cfg.devOp || 'develop'}
      conventions:
        branchPrefix: ${q(cfg.devBranchPrefix || 'agent/')}
        worktreeDir: ${q(cfg.devWorktreeDir || 'worktrees/')}
        tmpDir: "tmp/"
        maxFileLines: ${Number(cfg.devMaxLines) || 800}
    deploySystem:
      service: ${cfg.deployService || 'deploy-system'}
      resolution: ${cfg.deployResolution || 'default'}
      connector: ${cfg.deployConnector || 'connectops-deploy'}
      op: ${cfg.deployOp || 'release'}
  verification:
    mode: ${cfg.verifyMode || 'both'}
    connector: ${cfg.verifyConnector || 'connectops-code-audit'}
    op: ${cfg.verifyOp || 'audit'}
    onFail: ${cfg.verifyOnFail || 'retry-develop'}
    conditions:
${condLines || '      []'}
  phases:
${phaseLines}
  transitions:
    - {from: Idle,             on: advance,            to: Analyzing}
    - {from: Idle,             on: "entry:DocApplied", to: Planning}
    - {from: Analyzing,        on: advance,            to: AwaitingApproval}
    - {from: Analyzing,        on: worth_fixing,       to: AwaitingApproval}
    - {from: Analyzing,        on: ignore,             to: Reporting}
    - {from: AwaitingApproval, on: advance,            to: Planning}
    - {from: AwaitingApproval, on: ApprovalGranted,    to: Planning}
    - {from: AwaitingApproval, on: ApprovalRejected,   to: Failed}
    - {from: Planning,         on: advance,            to: Executing}
    - {from: Executing,        on: advance,            to: Verifying}
    - {from: Verifying,        on: advance,            to: Deploying}
    - {from: Verifying,        on: retry-develop,      to: Executing}
    - {from: Deploying,        on: advance,            to: Reporting}
  onEscalate: {action: pause, notify: bind-role}
  onFailed:
    backoff: 30s*2^n
    maxAttempts: 3
    rollbackPolicy: by-step
    escalate: {to: root, notify: bind-role}
  brakes:
    cooldown: 2m
    maxMutationsPerTick: 10
    avoidRepeatFix: true
  regression:
    statusWriteback: ${Boolean(cfg.statusWriteback)}
    bugDocPolicy: ${cfg.bugDocPolicy || 'invalidate'}
    cleanTempBugDoc: ${cfg.bugDocPolicy === 'delete'}
    targetPhaseOnConverge: InSync
    postActions:
${actionLines || '      []'}
    reports: [DriftReport, SessionLog]
`;
}

/* ---------- 节点流转编辑（2026-11-27 用户裁决：执行单元三型 connectops/agent/service +
 * gate 控制节点；节点可加删「我要加节点或者我要去掉验收节点」；agent 点进即配 LLM/ref 文档/
 * ref 规则/ref tag/tag 规则/工具；service=调目标系统内部 agent 执行+上一轮详细上下文+
 * permissionRole）。纯渲染 nodeCardHtml(n, i, defProvider) → HTML；绑定在 bindWorkflowStudio。 */

const NODE_UNITS = ['connectops', 'agent', 'service', 'gate'];

const UNIT_CHIP = { connectops: 'mute', agent: 'acc', service: 'warn', gate: 'llm' };

/** 单节点配置卡（点进即配：unit select 切换后由绑定层重渲染该列表）。 */
function nodeCardHtml(n, i, defProvider) {
  const u = n.unit || 'connectops';
  const grid = (inner) => `<div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(150px,1fr));gap:6px;margin-top:6px">${inner}</div>`;
  const inp = (lab, key, val) => `<div><div class="small faint">${esc(lab)}</div><input class="inp mono" data-node="${key}" data-ni="${i}" value="${esc(val == null ? '' : String(val))}" style="width:100%"></div>`;
  const base = grid(
    inp('节点名 (name)', 'name', n.name || ('node-' + (i + 1))) +
    inp('相位 (phase)', 'phase', n.phase || 'Idle') +
    `<div><div class="small faint">${esc(tr('职责说明 (title)'))}</div><input class="inp" data-node="title" data-ni="${i}" value="${esc(n.title || '')}" style="width:100%"></div>` +
    `<div><div class="small faint">${esc(tr('执行单元 (unit)'))}</div><select class="inp" data-node="unit" data-ni="${i}" style="width:100%">${NODE_UNITS.map((x) => `<option value="${x}" ${u === x ? 'selected' : ''}>${x}${x === 'gate' ? tr('（控制节点）') : ''}</option>`).join('')}</select></div>`
  );
  const rows = [base];
  if (u === 'agent') {
    rows.push(grid(
      inp('Provider', 'provider', n.provider || defProvider || 'openai-default') +
      inp('模型档位 (modelTier low/standard/high)', 'modelTier', n.modelTier || 'standard') +
      inp('ref 文档 (refDocs 逗号分隔)', 'refDocs', (n.refDocs || []).join(', ')) +
      inp('ref 规则 (refRules)', 'refRules', (n.refRules || []).join(', ')) +
      inp('默认 ref tag (refTags)', 'refTags', (n.refTags || []).join(', ')) +
      inp('tag 规则 (tagRules)', 'tagRules', (n.tagRules || []).join(', ')) +
      inp('随身工具 (tools)', 'tools', (n.tools || []).join(', ')) +
      inp('入口事件 (entryEvents)', 'entryEvents', (n.entryEvents || []).join(', '))
    ));
  } else if (u === 'service') {
    rows.push(grid(
      inp('目标系统 (service)', 'service', n.service || '') +
      inp('算子 (op)', 'op', n.op || '') +
      inp('ConnectorOps', 'connector', n.connector || '') +
      inp('权限 Role (permissionRole——P7 随文档授予)', 'permissionRole', n.permissionRole || '') +
      inp('上一轮上下文字段 (contextPassing.fields)', 'ctxFields', ((n.contextPassing && n.contextPassing.fields) || []).join(', ')) +
      `<div style="display:flex;align-items:end;padding-bottom:4px"><label class="small"><input type="checkbox" data-node="ctxPrev" data-ni="${i}" ${(!n.contextPassing || n.contextPassing.previousRound !== false) ? 'checked' : ''}> ${esc(tr('传上一轮详细上下文'))}</label></div>`
    ));
  } else if (u === 'connectops') {
    rows.push(grid(
      inp('ConnectorOps', 'connector', n.connector || '') +
      inp('算子 (op)', 'op', n.op || '') +
      inp('Provider', 'provider', n.provider || defProvider || 'openai-default') +
      inp('模型档位 (modelTier)', 'modelTier', n.modelTier || 'low') +
      inp('入口事件 (entryEvents)', 'entryEvents', (n.entryEvents || []).join(', '))
    ));
  } else {
    rows.push(grid(inp('闸算子 (op)', 'op', n.op || 'approval-check')));
  }
  return `<div class="card" style="margin:0 0 8px 0" data-node-card="${i}"><div class="card-b">` +
    `<div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:6px">` +
    `<div>${chip(u === 'gate' ? tr('控制节点') : u, UNIT_CHIP[u] || 'mute')} <b class="mono small">${esc(n.name || '')}</b></div>` +
    `<button type="button" class="btn tiny mute" data-node-del="${i}">${esc(tr('删除节点'))}</button></div>` +
    rows.join('') + `</div></div>`;
}

/** 四大设计表 + 以人为本可视化配置工作台 HTML */
export function workflowStudioHtml(wf, usedName) {
  const cfg = defaultLoopSpec(usedName, wf && wf.spec);
  const scope = (wf && wf.primaryScope) || 'global';
  const yamlText = (wf && wf.rawYaml) ? String(wf.rawYaml) : buildWorkflowYaml(cfg, scope);

  const condRowsHtml = (cfg.conditions || []).map((c, idx) =>
    `<div class="mini-row" data-wf-cond-row="${idx}" style="display:flex;gap:8px;align-items:center;flex-wrap:wrap">` +
      `<input type="checkbox" data-cond-req="${idx}" ${c.required !== false ? 'checked' : ''} title="必过硬门" />` +
      `<select class="inp" data-cond-kind="${idx}" style="width:110px">` +
        ['code', 'test', 'agent', 'probe', 'custom'].map((k) => `<option value="${k}" ${c.kind === k ? 'selected' : ''}>${k}</option>`).join('') +
      `</select>` +
      `<input class="inp" data-cond-title="${idx}" value="${esc(c.title || '')}" placeholder="验收条件描述" style="flex:2;min-width:200px" />` +
      `<input class="inp mono" data-cond-cmd="${idx}" value="${esc(c.command || '')}" placeholder="执行命令或 ConnectorOps/op" style="flex:1;min-width:160px" />` +
      `<button type="button" class="btn tiny mute" data-cond-del="${idx}">删除</button>` +
    `</div>`
  ).join('');

  const postActRowsHtml = (cfg.postActions || []).map((a, idx) =>
    `<div class="mini-row" data-wf-act-row="${idx}" style="display:flex;gap:8px;align-items:center;flex-wrap:wrap">` +
      `<input type="checkbox" data-act-en="${idx}" ${a.enabled !== false ? 'checked' : ''} title="启用此收尾动作" />` +
      `<select class="inp" data-act-kind="${idx}" style="width:140px">` +
        ['doc-lifecycle', 'status-writeback', 'notify', 'custom'].map((k) => `<option value="${k}" ${a.kind === k ? 'selected' : ''}>${k}</option>`).join('') +
      `</select>` +
      `<select class="inp" data-act-policy="${idx}" style="width:170px">` +
        `<option value="invalidate" ${(a.action || cfg.bugDocPolicy) === 'invalidate' ? 'selected' : ''}>标记失效 (invalidate)</option>` +
        `<option value="delete" ${(a.action || cfg.bugDocPolicy) === 'delete' ? 'selected' : ''}>删除文档 (delete)</option>` +
        `<option value="keep" ${(a.action || cfg.bugDocPolicy) === 'keep' ? 'selected' : ''}>保留不删除 (keep)</option>` +
      `</select>` +
      `<input class="inp" data-act-title="${idx}" value="${esc(a.title || '')}" placeholder="收尾动作说明" style="flex:2;min-width:220px" />` +
      `<button type="button" class="btn tiny mute" data-act-del="${idx}">删除动作</button>` +
    `</div>`
  ).join('');

  const designNodesHtml = (cfg.phases || []).map((n, i) =>
    `<tr>` +
      `<td class="mono small">${i + 1}. ${esc(n.name || '')}</td>` +
      `<td>${chip(n.phase || '—', 'acc')}</td>` +
      `<td><b>${esc(n.title || '')}</b></td>` +
      `<td class="mono small">${chip(n.unit === 'gate' ? tr('控制节点') : (n.unit || 'connectops'), UNIT_CHIP[n.unit] || 'mute')}</td>` +
      `<td class="mono small">${n.connector ? `<span class="lk" data-open-conn="${esc(n.connector)}">${esc(n.connector)}</span>` : '—'}${n.op ? ` · ${esc(n.op)}` : ''}</td>` +
      `<td><span class="chip acc tiny click" data-open-provider="${esc(n.provider || cfg.provider)}" title="点击配置 OpenAIProvider 账号/baseUrl/模型映射">Provider: ${esc(n.provider || cfg.provider)}</span> ` +
      `${chip('tier:' + (n.modelTier || 'standard'), n.modelTier === 'high' ? 'warn' : 'mute')}</td>` +
      `<td class="small mono">${Array.isArray(n.globalRuleRefs) && n.globalRuleRefs.length ? esc(n.globalRuleRefs.join(', ')) : '—'}</td>` +
    `</tr>`
  ).join('');

  return `
  <h2 class="sec-title">${esc(tr('设计表与节点流转架构（Workflow + Scheme = 声明层，代码 = 通用解释器）'))}</h2>
  <div class="card">
    <div class="card-h">
      <span class="h-title">${esc(tr(String((cfg.phases || []).length) + ' 节点主干、双触发入口与按任务智商分级模型设计表'))}</span>
      <button type="button" class="btn tiny acc" data-open-provider="${esc(cfg.provider)}">${esc(tr('配置默认 OpenAIProvider (' + cfg.provider + ')'))}</button>
    </div>
    <div class="card-b" style="overflow-x:auto">
      <div class="small faint" style="margin-bottom:8px">
        ${esc(tr('入口 1（运行期日志/Bug/巡检事件）从节点 1 (Idle) 顺次流转；入口 2（对话修改设计文档或产出临时 BugDoc 触发 DocApplied）直接从节点 4 (Planning / Doc Apply) 切入，直走 developer-system → 验收 → deploy-system → 总结归档。点击任意 Provider 徽章可直接配置 OpenAIProvider 账号、Base 地址与模型分级。'))}
      </div>
      <table class="tbl">
        <thead><tr><th>节点</th><th>相位</th><th>职责说明</th><th>执行单元</th><th>ConnectorOps / 算子</th><th>OpenAIProvider / 模型分级</th><th>引用 Global Rules</th></tr></thead>
        <tbody>${designNodesHtml}</tbody>
      </table>
    </div>
  </div>

  <div data-provider-drawer></div>

  <h2 class="sec-title">${esc(tr('以人为本可视化配置工作台（Scheme Adapter 层配置，保存即 Apply 落库）'))}</h2>
  <div class="card" data-wf-studio>
    <div class="card-h">
      <span class="h-title">${esc(usedName)} · ${esc(tr('全节点参数、自动审批、验收条件与临时 BugDoc 收尾策略配置'))}</span>
      <div>
        ${cfg.undeletable ? chip(tr('内置保护：不可删除'), 'warn') : ''}
        ${cfg.immutableNodes ? chip(tr('主干节点拓扑锁定（参数可配）'), 'acc') : chip(tr('节点可加删（draft→apply）'), 'ok')}
      </div>
    </div>
    <div class="card-b">
      <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(280px,1fr));gap:12px;margin-bottom:14px">
        <div class="card" style="margin:0"><div class="card-b">
          <b>${esc(tr('A. 总控开关与全局规则引用'))}</b>
          <div style="margin-top:8px">
            <label style="display:flex;align-items:center;gap:8px;cursor:pointer">
              <input type="checkbox" data-cfg="enabled" ${cfg.enabled ? 'checked' : ''} />
              <span><b>${esc(tr('启用此工作流 (spec.enabled)'))}</b> ${esc(tr('（出厂默认不启动；core-system 自我迭代显式开启）'))}</span>
            </label>
          </div>
          <div style="margin-top:8px">
            <div class="small faint">${esc(tr('涉及服务 (services.involved)'))}</div>
            <input class="inp" data-cfg="servicesInvolved" value="${esc(cfg.servicesInvolved)}" style="width:100%" />
          </div>
          <div style="margin-top:8px">
            <div class="small faint">${esc(tr('Session/Agent 默认引用全局规则 (globalRuleRefs)'))}</div>
            <input class="inp mono" data-cfg="globalRuleRefs" value="${esc(cfg.globalRuleRefs)}" style="width:100%" />
          </div>
        </div></div>

        <div class="card" style="margin:0"><div class="card-b">
          <b>${esc(tr('B. 节点 1（监听日志与 Event 过滤 + ConnectorOps 收集）'))}</b>
          <div style="margin-top:8px;display:grid;grid-template-columns:1fr 110px;gap:6px">
            <div>
              <div class="small faint">${esc(tr('监听事件类型'))}</div>
              <input class="inp mono" data-cfg="logEvents" value="${esc(cfg.logEvents)}" style="width:100%" />
            </div>
            <div>
              <div class="small faint">${esc(tr('最低日志级别'))}</div>
              <select class="inp" data-cfg="logMinLevel" style="width:100%">
                ${['debug', 'info', 'warn', 'error', 'fatal'].map((l) => `<option value="${l}" ${cfg.logMinLevel === l ? 'selected' : ''}>${l}</option>`).join('')}
              </select>
            </div>
          </div>
          <div style="margin-top:6px;display:grid;grid-template-columns:1fr 1fr;gap:6px">
            <div>
              <div class="small faint">${esc(tr('过滤 Service'))}</div>
              <input class="inp mono" data-cfg="logServices" value="${esc(cfg.logServices)}" style="width:100%" />
            </div>
            <div>
              <div class="small faint">${esc(tr('日志关键词/正则过滤'))}</div>
              <input class="inp mono" data-cfg="logKeyword" value="${esc(cfg.logKeyword)}" style="width:100%" />
            </div>
          </div>
          <div style="margin-top:6px;display:grid;grid-template-columns:1fr 110px auto;gap:6px;align-items:end">
            <div>
              <div class="small faint">${esc(tr('日志收集 ConnectorOps'))}</div>
              <input class="inp mono" data-cfg="collectorConnector" value="${esc(cfg.collectorConnector)}" style="width:100%" />
            </div>
            <div>
              <div class="small faint">${esc(tr('收集算子 op'))}</div>
              <input class="inp mono" data-cfg="collectorOp" value="${esc(cfg.collectorOp)}" style="width:100%" />
            </div>
            <button type="button" class="btn tiny acc" data-develop-collector title="当日志收集 ConnectorOps 缺失或为骨架时，直接拉出代码开发任务并打开编辑">${esc(tr('拉出代码开发/编辑'))}</button>
          </div>
          <div style="margin-top:6px;display:grid;grid-template-columns:auto 100px 1fr;gap:6px;align-items:center">
            <label><input type="checkbox" data-cfg="timerEnabled" ${cfg.timerEnabled ? 'checked' : ''} /> ${esc(tr('定时触发'))}</label>
            <input class="inp mono" data-cfg="timerInterval" value="${esc(cfg.timerInterval)}" placeholder="30m" />
            <input class="inp mono" data-cfg="timerCron" value="${esc(cfg.timerCron)}" placeholder="*/30 * * * *" />
          </div>
        </div></div>

        <div class="card" style="margin:0"><div class="card-b">
          <b>${esc(tr('C. 节点 3（审批 / 自动判断审批）& 节点 4（Doc Apply 第二入口）'))}</b>
          <div style="margin-top:8px;display:grid;grid-template-columns:1fr 110px;gap:6px">
            <div>
              <div class="small faint">${esc(tr('节点 3 审批模式 (approval.mode)'))}</div>
              <select class="inp" data-cfg="approvalMode" style="width:100%">
                <option value="agent" ${cfg.approvalMode === 'agent' ? 'selected' : ''}>agent（Agent 按规则自动判断审批）</option>
                <option value="auto" ${cfg.approvalMode === 'auto' ? 'selected' : ''}>auto（全自动放行不挂起）</option>
                <option value="human" ${cfg.approvalMode === 'human' ? 'selected' : ''}>human（人工审批挂起）</option>
              </select>
            </div>
            <div>
              <div class="small faint">${esc(tr('自动放行最高风险'))}</div>
              <select class="inp" data-cfg="approvalMaxRisk" style="width:100%">
                ${['low', 'medium', 'high'].map((r) => `<option value="${r}" ${cfg.approvalMaxRisk === r ? 'selected' : ''}>${r}</option>`).join('')}
              </select>
            </div>
          </div>
          <div style="margin-top:6px">
            <div class="small faint">${esc(tr('Agent 自动审批判断准则 (autoCriteria)'))}</div>
            <input class="inp" data-cfg="approvalCriteria" value="${esc(cfg.approvalCriteria)}" style="width:100%" />
          </div>
          <div style="margin-top:8px;padding-top:8px;border-top:1px solid var(--ds-border)">
            <label style="display:flex;align-items:center;gap:6px;cursor:pointer">
              <input type="checkbox" data-cfg="docApplyEntryEnabled" ${cfg.docApplyEntryEnabled ? 'checked' : ''} />
              <span><b>${esc(tr('启用节点 4 (Doc Apply) 多入口切入'))}</b> ${esc(tr('（对话改文档/临时 BugDoc 触发 DocApplied 直切节点 4）'))}</span>
            </label>
            <input class="inp mono" data-cfg="docApplyDocTypes" value="${esc(cfg.docApplyDocTypes)}" style="width:100%;margin-top:4px" />
          </div>
        </div></div>

        <div class="card" style="margin:0"><div class="card-b">
          <b>${esc(tr('D. 节点 5 & 7（developer-system / deploy-system 逐层绑定）'))}</b>
          <div style="margin-top:8px;display:grid;grid-template-columns:1fr 130px;gap:6px">
            <div>
              <div class="small faint">${esc(tr('研发系统 (developer-system)'))}</div>
              <input class="inp mono" data-cfg="devService" value="${esc(cfg.devService)}" style="width:100%" />
            </div>
            <div>
              <div class="small faint">${esc(tr('来源策略 (逐层询问)'))}</div>
              <select class="inp" data-cfg="devResolution" style="width:100%">
                <option value="default" ${cfg.devResolution === 'default' ? 'selected' : ''}>default (默认开源习惯)</option>
                <option value="install" ${cfg.devResolution === 'install' ? 'selected' : ''}>install (模板安装)</option>
                <option value="import" ${cfg.devResolution === 'import' ? 'selected' : ''}>import (导入已有系统)</option>
              </select>
            </div>
          </div>
          <div style="margin-top:6px;display:grid;grid-template-columns:1fr 1fr 80px;gap:6px">
            <input class="inp mono" data-cfg="devBranchPrefix" value="${esc(cfg.devBranchPrefix)}" placeholder="分支前缀 agent/" title="开源习惯：分支前缀" />
            <input class="inp mono" data-cfg="devWorktreeDir" value="${esc(cfg.devWorktreeDir)}" placeholder="worktrees/" title="开源习惯：worktree 目录" />
            <input class="inp mono" data-cfg="devMaxLines" value="${esc(String(cfg.devMaxLines))}" placeholder="800" title="单文件行数上限" />
          </div>
          <div style="margin-top:8px;display:grid;grid-template-columns:1fr 130px;gap:6px">
            <div>
              <div class="small faint">${esc(tr('发布系统 (deploy-system)'))}</div>
              <input class="inp mono" data-cfg="deployService" value="${esc(cfg.deployService)}" style="width:100%" />
            </div>
            <div>
              <div class="small faint">${esc(tr('来源策略 (逐层询问)'))}</div>
              <select class="inp" data-cfg="deployResolution" style="width:100%">
                <option value="default" ${cfg.deployResolution === 'default' ? 'selected' : ''}>default (默认宿主发布)</option>
                <option value="install" ${cfg.deployResolution === 'install' ? 'selected' : ''}>install (模板安装)</option>
                <option value="import" ${cfg.deployResolution === 'import' ? 'selected' : ''}>import (导入外部发布)</option>
              </select>
            </div>
          </div>
        </div></div>
      </div>

      <div class="card" style="margin-bottom:12px"><div class="card-b">
        <div style="display:flex;justify-content:space-between;align-items:center;flex-wrap:wrap;gap:8px;margin-bottom:8px">
          <b>${esc(tr('E. 节点 6（验收 Verify 配置 —— 怎么验收、验收条件可自由添加/删除）'))}</b>
          <div style="display:flex;gap:8px;align-items:center">
            <select class="inp" data-cfg="verifyMode">
              <option value="both" ${cfg.verifyMode === 'both' ? 'selected' : ''}>both（代码审核 + Agent 验收）</option>
              <option value="code" ${cfg.verifyMode === 'code' ? 'selected' : ''}>code（仅代码规则审核）</option>
              <option value="agent" ${cfg.verifyMode === 'agent' ? 'selected' : ''}>agent（仅 Agent 探针验收）</option>
              <option value="human" ${cfg.verifyMode === 'human' ? 'selected' : ''}>human（人工确认验收）</option>
            </select>
            <select class="inp" data-cfg="verifyOnFail">
              <option value="retry-develop" ${cfg.verifyOnFail === 'retry-develop' ? 'selected' : ''}>未通过：打回 developer-system 重修</option>
              <option value="block" ${cfg.verifyOnFail === 'block' ? 'selected' : ''}>未通过：阻断并上报</option>
              <option value="ask" ${cfg.verifyOnFail === 'ask' ? 'selected' : ''}>未通过：转人工裁决</option>
            </select>
            <button type="button" class="btn tiny acc" data-cond-add>${esc(tr('+ 添加验收条件'))}</button>
          </div>
        </div>
        <div data-wf-cond-list>${condRowsHtml}</div>
      </div></div>

      <div class="card" style="margin-bottom:12px"><div class="card-b">
        <div style="display:flex;justify-content:space-between;align-items:center;flex-wrap:wrap;gap:8px;margin-bottom:8px">
          <b>${esc(tr('F. 节点 8（总结归档 Reporting 通用解释器收尾动作 —— 临时 BugDoc 默认可标记失效/删除/选择不删除，动作项可自由增删）'))}</b>
          <div style="display:flex;gap:8px;align-items:center">
            <span class="small faint">${esc(tr('临时 BugDoc 默认策略：'))}</span>
            <select class="inp" data-cfg="bugDocPolicy">
              <option value="invalidate" ${cfg.bugDocPolicy === 'invalidate' ? 'selected' : ''}>标记失效归档 (invalidate)</option>
              <option value="delete" ${cfg.bugDocPolicy === 'delete' ? 'selected' : ''}>删除临时 BugDoc (delete)</option>
              <option value="keep" ${cfg.bugDocPolicy === 'keep' ? 'selected' : ''}>保留不删除 (keep)</option>
            </select>
            <label style="display:flex;align-items:center;gap:4px"><input type="checkbox" data-cfg="statusWriteback" ${cfg.statusWriteback ? 'checked' : ''} /> ${esc(tr('回写 InSync'))}</label>
            <button type="button" class="btn tiny acc" data-act-add>${esc(tr('+ 添加收尾动作'))}</button>
          </div>
        </div>
        <div data-wf-act-list>${postActRowsHtml}</div>
      </div></div>

      <details class="ex-block" open>
        <summary class="ex-sum"><span class="ex-chev">▶</span><b>${esc(tr('Workflow 完整 YAML 声明（解释器权威源，表单联动实时生成，也可直接编辑 YAML）'))}</b></summary>
        <div class="ex-body">
          <textarea class="inp mono" data-wf-yaml rows="18" style="width:100%;font-size:12px;line-height:1.45">${esc(yamlText)}</textarea>
        </div>
      </details>

      <div class="card" style="margin-bottom:12px"><div class="card-b">
        <div style="display:flex;justify-content:space-between;align-items:center;flex-wrap:wrap;gap:8px;margin-bottom:8px">
          <b>${esc(tr('G. 节点流转编辑（执行单元三型 connectops/agent/service；agent 点进即配——LLM/ref 文档/ref 规则/ref tag/tag 规则/工具；service=调目标系统且内部由其 agent 执行+上一轮详细上下文+权限 Role；节点可加删——如删验收节点）'))}</b>
          <span class="small faint">${esc(tr('改动随「保存并 Apply」以 Workflow 文档对象落账'))}</span>
        </div>
        <div data-wf-node-list>${(cfg.phases || []).map((n, i) => nodeCardHtml(n, i, cfg.provider)).join('')}</div>
        <button type="button" class="btn tiny acc" data-node-add>${esc(tr('+ 添加节点'))}</button>
      </div></div>

      <div style="margin-top:12px;display:flex;gap:10px;align-items:center;flex-wrap:wrap">
        <button type="button" class="btn acc" data-wf-save>${esc(tr('保存并 Apply 工作流配置'))}</button>
        <button type="button" class="btn mute" data-wf-sync-yaml>${esc(tr('从上方表单重新生成 YAML'))}</button>
        <span class="small faint" data-wf-save-msg></span>
      </div>
    </div>
  </div>`;
}

/** 绑定可视化配置工作台交互（增删验收条件、增删收尾动作、拉出 ConnectorOps 开发任务、打开 OpenAIProvider 配置、Apply 保存） */
export function bindWorkflowStudio(el, ctx, wf, usedName, onSaved) {
  const studio = el.querySelector('[data-wf-studio]');
  if (!studio) return;
  const scope = (wf && wf.primaryScope) || 'global';
  let baseRv = Number((wf && wf.primaryRv) || 0);
  const cfg = defaultLoopSpec(usedName, wf && wf.spec);

  const readFormIntoCfg = () => {
    studio.querySelectorAll('[data-cfg]').forEach((inp) => {
      const k = inp.getAttribute('data-cfg');
      if (inp.type === 'checkbox') cfg[k] = Boolean(inp.checked);
      else cfg[k] = inp.value;
    });
    const conds = [];
    studio.querySelectorAll('[data-wf-cond-row]').forEach((row) => {
      const idx = row.getAttribute('data-wf-cond-row');
      conds.push({
        id: 'cond-' + (conds.length + 1),
        required: Boolean(row.querySelector(`[data-cond-req="${idx}"]`)?.checked),
        kind: row.querySelector(`[data-cond-kind="${idx}"]`)?.value || 'code',
        title: row.querySelector(`[data-cond-title="${idx}"]`)?.value || '',
        command: row.querySelector(`[data-cond-cmd="${idx}"]`)?.value || '',
      });
    });
    cfg.conditions = conds;
    const acts = [];
    studio.querySelectorAll('[data-wf-act-row]').forEach((row) => {
      const idx = row.getAttribute('data-wf-act-row');
      acts.push({
        id: 'act-' + (acts.length + 1),
        enabled: Boolean(row.querySelector(`[data-act-en="${idx}"]`)?.checked),
        kind: row.querySelector(`[data-act-kind="${idx}"]`)?.value || 'doc-lifecycle',
        targetDocType: 'BugDoc',
        action: row.querySelector(`[data-act-policy="${idx}"]`)?.value || cfg.bugDocPolicy || 'invalidate',
        title: row.querySelector(`[data-act-title="${idx}"]`)?.value || '',
      });
    });
    cfg.postActions = acts;
    /* 节点表 → cfg.phases（DOM 顺序=节点序；unit 分型读字段——执行单元三型，2026-11-27） */
    const nodeCards = studio.querySelectorAll('[data-node-card]');
    if (nodeCards.length) {
      const nodes = [];
      nodeCards.forEach((card) => {
        const get = (k) => card.querySelector(`[data-node="${k}"]`);
        const lst = (k) => String(get(k) ? get(k).value : '').split(/[，,]+/).map((x) => x.trim()).filter(Boolean);
        const u = get('unit') ? get('unit').value : 'connectops';
        const n = {
          name: (get('name') && get('name').value.trim()) || ('node-' + (nodes.length + 1)),
          phase: (get('phase') && get('phase').value.trim()) || 'Idle',
          title: get('title') ? get('title').value : '',
          unit: u,
        };
        if (u === 'agent') {
          n.provider = get('provider') ? get('provider').value.trim() : cfg.provider;
          n.modelTier = get('modelTier') ? get('modelTier').value.trim() : 'standard';
          ['refDocs', 'refRules', 'refTags', 'tagRules', 'tools', 'entryEvents'].forEach((k) => { n[k] = lst(k); });
        } else if (u === 'service') {
          n.service = get('service') ? get('service').value.trim() : '';
          n.op = get('op') ? get('op').value.trim() : '';
          n.connector = get('connector') ? get('connector').value.trim() : '';
          n.permissionRole = get('permissionRole') ? get('permissionRole').value.trim() : '';
          n.contextPassing = {
            previousRound: Boolean(get('ctxPrev') && get('ctxPrev').checked),
            fields: lst('ctxFields'),
          };
        } else if (u === 'connectops') {
          n.connector = get('connector') ? get('connector').value.trim() : '';
          n.op = get('op') ? get('op').value.trim() : '';
          n.provider = get('provider') ? get('provider').value.trim() : cfg.provider;
          n.modelTier = get('modelTier') ? get('modelTier').value.trim() : 'low';
          n.entryEvents = lst('entryEvents');
        } else {
          n.op = get('op') ? get('op').value.trim() : 'approval-check';
        }
        nodes.push(n);
      });
      cfg.phases = nodes;
    }
  };

  const syncYamlBox = () => {
    readFormIntoCfg();
    const box = studio.querySelector('[data-wf-yaml]');
    if (box) box.value = buildWorkflowYaml(cfg, scope);
  };

  studio.querySelectorAll('[data-cfg]').forEach((inp) => {
    inp.addEventListener('change', syncYamlBox);
  });

  const bindRowDeletes = () => {
    studio.querySelectorAll('[data-cond-del]').forEach((btn) => {
      btn.onclick = () => {
        btn.closest('[data-wf-cond-row]')?.remove();
        syncYamlBox();
      };
    });
    studio.querySelectorAll('[data-act-del]').forEach((btn) => {
      btn.onclick = () => {
        btn.closest('[data-wf-act-row]')?.remove();
        syncYamlBox();
      };
    });
    studio.querySelectorAll('[data-wf-cond-row] input, [data-wf-cond-row] select, [data-wf-act-row] input, [data-wf-act-row] select').forEach((inp) => {
      inp.onchange = syncYamlBox;
    });
  };
  bindRowDeletes();

  studio.querySelector('[data-cond-add]')?.addEventListener('click', () => {
    const list = studio.querySelector('[data-wf-cond-list]');
    if (!list) return;
    const idx = Date.now();
    const div = document.createElement('div');
    div.className = 'mini-row';
    div.setAttribute('data-wf-cond-row', String(idx));
    div.style.cssText = 'display:flex;gap:8px;align-items:center;flex-wrap:wrap';
    div.innerHTML =
      `<input type="checkbox" data-cond-req="${idx}" checked />` +
      `<select class="inp" data-cond-kind="${idx}" style="width:110px"><option value="code">code</option><option value="test">test</option><option value="agent">agent</option><option value="probe">probe</option><option value="custom">custom</option></select>` +
      `<input class="inp" data-cond-title="${idx}" value="自定义验收条件" style="flex:2;min-width:200px" />` +
      `<input class="inp mono" data-cond-cmd="${idx}" value="connectops-code-audit/audit" style="flex:1;min-width:160px" />` +
      `<button type="button" class="btn tiny mute" data-cond-del="${idx}">删除</button>`;
    list.appendChild(div);
    bindRowDeletes();
    syncYamlBox();
  });

  studio.querySelector('[data-act-add]')?.addEventListener('click', () => {
    const list = studio.querySelector('[data-wf-act-list]');
    if (!list) return;
    const idx = Date.now();
    const div = document.createElement('div');
    div.className = 'mini-row';
    div.setAttribute('data-wf-act-row', String(idx));
    div.style.cssText = 'display:flex;gap:8px;align-items:center;flex-wrap:wrap';
    div.innerHTML =
      `<input type="checkbox" data-act-en="${idx}" checked />` +
      `<select class="inp" data-act-kind="${idx}" style="width:140px"><option value="doc-lifecycle">doc-lifecycle</option><option value="status-writeback">status-writeback</option><option value="notify">notify</option><option value="custom">custom</option></select>` +
      `<select class="inp" data-act-policy="${idx}" style="width:170px"><option value="invalidate">标记失效 (invalidate)</option><option value="delete">删除文档 (delete)</option><option value="keep">保留不删除 (keep)</option></select>` +
      `<input class="inp" data-act-title="${idx}" value="临时 BugDoc 收尾动作" style="flex:2;min-width:220px" />` +
      `<button type="button" class="btn tiny mute" data-act-del="${idx}">删除动作</button>`;
    list.appendChild(div);
    bindRowDeletes();
    syncYamlBox();
  });

  studio.querySelector('[data-wf-sync-yaml]')?.addEventListener('click', syncYamlBox);

  /* 节点流转编辑绑定（加节点/删节点/unit 切换重渲染该卡字段组——agent 点进即配） */
  const nodeList = studio.querySelector('[data-wf-node-list]');
  const bindNodeEvents = () => {
    if (!nodeList) return;
    nodeList.querySelectorAll('[data-node-del]').forEach((btn) => {
      btn.onclick = () => {
        readFormIntoCfg();
        cfg.phases.splice(Number(btn.getAttribute('data-node-del')), 1);
        renderNodeList();
        syncYamlBox();
      };
    });
    nodeList.querySelectorAll('[data-node]').forEach((inp) => {
      inp.onchange = () => {
        readFormIntoCfg();
        renderNodeList();
        syncYamlBox();
      };
    });
  };
  const renderNodeList = () => {
    if (!nodeList) return;
    nodeList.innerHTML = (cfg.phases || []).map((n, i) => nodeCardHtml(n, i, cfg.provider)).join('');
    bindNodeEvents();
  };
  studio.querySelector('[data-node-add]')?.addEventListener('click', () => {
    readFormIntoCfg();
    cfg.phases.push({ name: 'node-' + (cfg.phases.length + 1), phase: 'Idle', title: '新节点', unit: 'connectops', provider: cfg.provider, modelTier: 'standard' });
    renderNodeList();
    syncYamlBox();
  });
  bindNodeEvents();

  // 拉出日志收集 ConnectorOps 代码开发/编辑任务
  studio.querySelector('[data-develop-collector]')?.addEventListener('click', async () => {
    readFormIntoCfg();
    const connName = String(cfg.collectorConnector || 'connectops-log-collector').trim();
    const opName = String(cfg.collectorOp || 'collect-logs').trim();
    const msg = studio.querySelector('[data-wf-save-msg]');
    if (msg) msg.textContent = tr('正在检查/生成 ConnectorOps 骨架并拉出代码开发任务…');
    const existing = await ctx.api.object('ConnectorOps', 'global', connName);
    if (!existing.ok) {
      const connYaml = `apiVersion: core/v1\nkind: ConnectorOps\nmetadata:\n  name: ${connName}\n  scope: global\nspec:\n  workservice: workflow-engine\n  workflow: ${usedName}\n  provider: ${cfg.provider || 'openai-default'}\n  llm:\n    provider: ${cfg.provider || 'openai-default'}\n    tier: low\n    model: "flash-tier"\n  ops: [${opName}]\n  config:\n    query: ${JSON.stringify(cfg.collectorQuery || 'level>=warn')}\n    services: ${JSON.stringify(cfg.logServices || 'self-recursive-core')}\n`;
      await ctx.api.apply({ kind: 'ConnectorOps', scope: 'global', name: connName, resourceVersion: 0, content: connYaml, fieldManager: 'console:workflow-studio', reasonRef: 'develop-collector:' + connName });
    }
    if (msg) msg.innerHTML = esc(tr('已就绪 ConnectorOps/global/' + connName + ' —— ')) + chipGo(tr('打开 ConnectorOps 编辑器开发算子'), docRoute('ConnectorOps', 'global', connName), 'acc', true);
  });

  // 保存并 Apply Workflow 配置
  studio.querySelector('[data-wf-save]')?.addEventListener('click', async () => {
    const yamlBox = studio.querySelector('[data-wf-yaml]');
    const content = yamlBox ? yamlBox.value : buildWorkflowYaml(cfg, scope);
    const msg = studio.querySelector('[data-wf-save-msg]');
    if (msg) msg.textContent = tr('正在 Apply 工作流配置…');
    let rv = baseRv;
    const cur = await ctx.api.object('Workflow', scope, usedName);
    if (cur.ok && cur.data && cur.data.resourceVersion != null) rv = Number(cur.data.resourceVersion) || 0;
    const res = await ctx.api.apply({
      kind: 'Workflow', scope, name: usedName, resourceVersion: rv,
      content, fieldManager: 'console:workflow-studio', reasonRef: 'workflow-config:' + usedName,
    });
    if (!res.ok) {
      if (msg) msg.textContent = tr('Apply 失败：') + String(res.error || res.status);
      return;
    }
    baseRv = Number(res.data && res.data.rv) || (rv + 1);
    if (msg) msg.textContent = tr('已保存并 Apply 生效（rv=' + baseRv + '）');
    if (typeof onSaved === 'function') onSaved();
  });

  // 点击任意 Provider 徽章弹出 OpenAIProvider 配置抽屉
  el.querySelectorAll('[data-open-provider]').forEach((btn) => {
    btn.addEventListener('click', () => openProviderDrawer(el, ctx, btn.getAttribute('data-open-provider') || 'openai-default'));
  });
  el.querySelectorAll('[data-open-conn]').forEach((btn) => {
    btn.addEventListener('click', () => {
      const cName = btn.getAttribute('data-open-conn');
      if (cName && ctx.router) ctx.router.go(docRoute('ConnectorOps', 'global', cName));
    });
  });
}

/** OpenAIProvider 点击即配抽屉（配置账号/密钥引用、BaseUrl、模型列表、不同任务/ConnectorOps 智商档位映射） */
export async function openProviderDrawer(el, ctx, providerName = 'openai-default') {
  const slot = el.querySelector('[data-provider-drawer]');
  if (!slot) return;
  slot.innerHTML = `<div class="card" style="border:1px solid var(--ds-accent)"><div class="card-b"><div class="small faint">${esc(tr('正在加载 OpenAIProvider (' + providerName + ') 配置…'))}</div></div></div>`;
  const res = await ctx.api.object('Provider', 'global', providerName);
  const rv = (res.ok && res.data && res.data.resourceVersion != null) ? Number(res.data.resourceVersion) : 0;
  const raw = (res.ok && res.data && res.data.content) ? String(res.data.content) : '';
  const extract = (re, def) => { const m = raw.match(re); return m ? m[1].trim().replace(/^["']|["']$/g, '') : def; };
  const account = extract(/account:\s*([^\n#]+)/, 'magpie-openai-gateway');
  const secretSrc = extract(/source:\s*([^\n}#]+)/, 'secret:llm/magpie-gateway-key');
  const baseUrl = extract(/baseUrl:\s*([^\n#]+)/, 'https://magpie.ListenAll.cn/v1');
  const lowModel = extract(/low:\s*([^\n}#]+)/, 'flash-tier');
  const stdModel = extract(/standard:\s*([^\n}#]+)/, 'standard-tier');
  const highModel = extract(/high:\s*([^\n}#]+)/, 'reasoning-tier');

  slot.innerHTML = `
  <div class="card" style="border:1px solid var(--ds-accent);margin-bottom:14px">
    <div class="card-h">
      <span class="h-title">${esc(tr('OpenAIProvider 配置面板'))} · <span class="mono">${esc(providerName)}</span> (rv=${rv})</span>
      <button type="button" class="btn tiny mute" data-prov-close>${esc(tr('收起'))}</button>
    </div>
    <div class="card-b">
      <div class="small faint" style="margin-bottom:8px">${esc(tr('所有 Session、ConnectorOps 与 Workflow 节点默认绑定此 OpenAIProvider。支持按任务智商要求分级路由：重复性/日志过滤任务用低智商轻量模型（low），常规开发用标准模型（standard），架构设计与复杂根因诊断用高智商推理模型（high）。'))}</div>
      <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(220px,1fr));gap:10px">
        <div>
          <div class="small faint">${esc(tr('账号标识 (auth.account)'))}</div>
          <input class="inp mono" data-prov="account" value="${esc(account)}" style="width:100%" />
        </div>
        <div>
          <div class="small faint">${esc(tr('密钥引用 (auth.secret.source)'))}</div>
          <input class="inp mono" data-prov="secretSrc" value="${esc(secretSrc)}" style="width:100%" />
        </div>
        <div>
          <div class="small faint">${esc(tr('OpenAI 兼容 Base 地址 (config.baseUrl)'))}</div>
          <input class="inp mono" data-prov="baseUrl" value="${esc(baseUrl)}" style="width:100%" />
        </div>
      </div>
      <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(220px,1fr));gap:10px;margin-top:10px">
        <div>
          <div class="small faint">${esc(tr('Low 档模型（重复性任务/日志收集/代码审核算子）'))}</div>
          <input class="inp mono" data-prov="lowModel" value="${esc(lowModel)}" style="width:100%" />
        </div>
        <div>
          <div class="small faint">${esc(tr('Standard 档模型（Doc Apply/常规开发算子）'))}</div>
          <input class="inp mono" data-prov="stdModel" value="${esc(stdModel)}" style="width:100%" />
        </div>
        <div>
          <div class="small faint">${esc(tr('High 档模型（架构设计/Agent 判断/自动审批/核心研发）'))}</div>
          <input class="inp mono" data-prov="highModel" value="${esc(highModel)}" style="width:100%" />
        </div>
      </div>
      <div style="margin-top:10px;display:flex;gap:10px;align-items:center">
        <button type="button" class="btn acc" data-prov-save>${esc(tr('保存并 Apply OpenAIProvider'))}</button>
        ${chipGo(tr('在对象编辑器中查看 YAML'), docRoute('Provider', 'global', providerName), 'mute', true)}
        <span class="small faint" data-prov-msg></span>
      </div>
    </div>
  </div>`;

  slot.querySelector('[data-prov-close]')?.addEventListener('click', () => { slot.innerHTML = ''; });
  slot.querySelector('[data-prov-save]')?.addEventListener('click', async () => {
    const val = (k) => slot.querySelector(`[data-prov="${k}"]`)?.value.trim() || '';
    const yamlContent = `apiVersion: core/v1
kind: Provider
metadata:
  name: ${providerName}
  scope: global
spec:
  kind: openai
  trustLevel: trusted
  auth:
    scheme: bearer
    account: ${JSON.stringify(val('account') || 'magpie-openai-gateway')}
    secret:
      source: ${JSON.stringify(val('secretSrc') || 'secret:llm/magpie-gateway-key')}
      rotation: manual
  config:
    baseUrl: ${JSON.stringify(val('baseUrl') || 'https://magpie.ListenAll.cn/v1')}
    defaultModel: ${JSON.stringify(val('highModel') || 'reasoning-tier')}
    models: [${JSON.stringify(val('lowModel'))}, ${JSON.stringify(val('stdModel'))}, ${JSON.stringify(val('highModel'))}]
    modelTiers:
      low: ${JSON.stringify(val('lowModel') || 'flash-tier')}
      standard: ${JSON.stringify(val('stdModel') || 'standard-tier')}
      high: ${JSON.stringify(val('highModel') || 'reasoning-tier')}
    taskModelMapping:
      connectops-log-collector: ${JSON.stringify(val('lowModel') || 'flash-tier')}
      connectops-code-audit: ${JSON.stringify(val('lowModel') || 'flash-tier')}
      connectops-deploy: ${JSON.stringify(val('lowModel') || 'flash-tier')}
      doc-apply: ${JSON.stringify(val('stdModel') || 'standard-tier')}
      agent-judge: ${JSON.stringify(val('highModel') || 'reasoning-tier')}
      approval-gate: ${JSON.stringify(val('highModel') || 'reasoning-tier')}
      connectops-developer: ${JSON.stringify(val('highModel') || 'reasoning-tier')}
`;
    const msg = slot.querySelector('[data-prov-msg]');
    if (msg) msg.textContent = tr('正在 Apply OpenAIProvider…');
    const r = await ctx.api.apply({
      kind: 'Provider', scope: 'global', name: providerName, resourceVersion: rv,
      content: yamlContent, fieldManager: 'console:provider-drawer', reasonRef: 'openai-provider:' + providerName,
    });
    if (msg) msg.textContent = r.ok ? tr('OpenAIProvider 已保存生效 (rv=' + (r.data && r.data.rv) + ')') : (tr('保存失败：') + String(r.error || r.status));
  });
}

/** 数据库独立文档导出配置面板（GET/PUT /v1/docs-export/config —— 默认不导出 BugDoc 等运行时文档） */
export async function hydrateDocsExportConfig(el, ctx) {
  const slot = el.querySelector('[data-wf-export-cfg]');
  if (!slot || !ctx.api.exportConfig) return;
  const r = await ctx.api.exportConfig();
  if (!r.ok || !el.isConnected) return;
  const cfg = r.data || {};
  const excl = Array.isArray(cfg.excludeDocTypes) ? cfg.excludeDocTypes.join(', ') : 'BugDoc, DriftReport, PlanDoc, SessionLog';
  slot.innerHTML = `
  <div class="card">
    <div class="card-h"><span class="h-title">${esc(tr('文档 Git 导出独立配置（数据库 docs_export_config，非代码硬编码）'))}</span></div>
    <div class="card-b" style="display:flex;gap:12px;align-items:center;flex-wrap:wrap">
      <label style="display:flex;align-items:center;gap:6px;cursor:pointer">
        <input type="checkbox" data-exp-inc ${cfg.includeRuntimeDocs ? 'checked' : ''} />
        <span>${esc(tr('导出时包含运行时临时文档 (includeRuntimeDocs)'))}</span>
      </label>
      <span class="small faint">${esc(tr('默认排除的运行时 docTypes：'))}</span>
      <input class="inp mono" data-exp-types value="${esc(excl)}" style="min-width:280px;flex:1" />
      <button type="button" class="btn tiny acc" data-exp-save>${esc(tr('保存导出配置'))}</button>
      <span class="small faint" data-exp-msg></span>
    </div>
  </div>`;
  slot.querySelector('[data-exp-save]')?.addEventListener('click', async () => {
    const inc = Boolean(slot.querySelector('[data-exp-inc]')?.checked);
    const types = String(slot.querySelector('[data-exp-types]')?.value || '').split(',').map((x) => x.trim()).filter(Boolean);
    const msg = slot.querySelector('[data-exp-msg]');
    const res = await ctx.api.saveExportConfig({ includeRuntimeDocs: inc, excludeDocTypes: types, excludeKinds: cfg.excludeKinds || [] });
    if (msg) msg.textContent = res.ok ? tr('导出配置已更新入库') : (tr('保存失败：') + String(res.error || res.status));
  });
}
