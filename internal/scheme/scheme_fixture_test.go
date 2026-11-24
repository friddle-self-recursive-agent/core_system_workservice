// scheme_fixture_test.go —— 权威声明逐字快照（生成自 core_design/core/scheme.yaml v0.13，
// 含 loop-10 Role 子字段必填锐化 + loop-52 kind:Proposal 注册）。改动权威声明须同步再生成此夹具；
// 真仓对账由 validate_test.go TestValidateDocAuthoritativeScheme 直读相对路径兜底漂移。
package scheme

// fixtureSchemeYAML 权威 scheme.yaml 逐字快照（漂移卫兵：checker 的声明解析必须吃下真实全文）。
const fixtureSchemeYAML = `
# core/scheme.yaml —— 完整类型注册表（v0.13，R12 补全）
# 依据：concepts/ 全部概念文档 §3。标注：! = 必填；? = 可选；a|b = 枚举；已采纳的墓碑概念不注册。
# 机读语义（checker=scheme.ValidateDoc，R-TYPE 有规则即有 checker）：无标注=不强制必填但出现即校验形状
# （类型/枚举/嵌套）；子字段必填=decl 后缀 !（先例 ref!）；未知字段放行（只增不破坏）；散文声明透明放行。
# 层级判例：System=deploy_system → 实例=sybran 开发实例 → WorkService=jenkins/onedev/bug-management
apiVersion: core/v1
kind: Scheme
metadata:
  name: default
  scope: builtin
spec:
  version: core/v1

  # ============================================================
  # 一、对象类型（kinds，18 项；已采纳的墓碑概念不注册，见文末不注册清单）
  # storage=纯路径模板：占位符 <name>=对象名、<s>=系统段、<i>=实例段；
  # 交替分支以 | 分隔；分支须含 <name> 或为固定单例路径，否则机器不可寻址。
  # ============================================================
  kinds:

    - kind: Scheme                    # 自注册
      builtin: true
      storage: core/scheme.yaml
      spec: {version: "!string", kinds: "[]KindReg", docTypes: "[]DocTypeReg",
             eventTypes: "[]EventTypeReg", dispatch: "[]DispatchRule", runtime: "?RuntimeReg",
             conventions: "?{status, cascade}"}

    - kind: DocType                   # 文档类型注册（本系统的 CRD 机制）
      builtin: true
      storage: doctypes/<name>.yaml
      spec:
        appliesTo: "{kinds: []string, scopes: []string}!"
        requiredSections: "[]string"          # ≤3（doctype.md 规则 2）
        writers: "[]string"                   # 默认最窄
        admission: "normal|strict"
        mandatory: "required|default|optional"  # 注册级；Service 可加严

    - kind: ServiceDesc
      builtin: true
      storage: systems/<name>/system.yaml
      spec:
        goalRef: "!ref"
        descRef: "!ref"
        docs: "?[]ref"
        externalRefs: "?[]{capabilities: []string}"     # 按能力引用 WorkService
        instances: "[]{name: string, agentRefs: []ref}"
        builtinServices: "?[]ref"                        # 仅根 core_system
      status: {}                                          # Desc 无 Status（P2）
      relations: {owns: ServiceInstance, workflows: Workflow}

    - kind: ServiceInstance
      builtin: true
      storage: systems/<s>/instances/<name>/instance.yaml
      spec: {}                                            # v1 跟随 owner Desc
      fields:
        agentRefs: "[]ref"
        docs: "{impl: ref, verify: ref}"
        metricsInfo: "{store: ref, queries: map}"
      status:
        observedGeneration: "int"
        phase: "InSync|Drifting|Remediating|Failed|Escalated"
        conditions: "[]{type, status, reason, message, lastTransitionTime}"
        drift: "{summary, reportRef}"
      cascade: 删 systems/<s>/ 目录即级联（目录即对象，A51）

    - kind: Agent
      builtin: true
      storage: global/agents/<name>/agent.yaml
      spec:
        runtime: "?string (=dsh 外部运行时)"
        autoLevel: "?ask|auto|all"                
        connectorOps: "[]{ops: ref, as: string}"  
        providers: "?[]ref"                       
        refs: "?[]ref"
        roles: "?[]ref"                           
        docsPolicy: "{read: []scopePattern, onDemand: bool}"
        subAgents: "[]{name, spawn: machine-instance}"   
        retry: "{maxAttempts, backoff, escalateTo: root}" 
      status:
        phase: "Idle|Analyzing|Planning|AwaitingApproval|Executing|Verifying|Reporting|Failed"
        workflow: "?string"
        currentSession: "{id, startedAt, docsLoaded}"
        conditions: "[]Condition"

    - kind: WorkService              # 六面体
      builtin: true
      storage: global/workservices/<name>/workservice.yaml   # 作用域=Global（workservice.md §头表）——物化落 global/ 桶（layout.md §2 路径=scope 准绳；无 scope 段根级模板不可寻址=v0.44 收口前缺口）
      spec:
        category: "developer|storage|tool|business|…"   # tool 型=本机工具（dsh/http）
        connector:                                  # 连接面（原 ConnectorProvider 并入）
          types: "[]string"
          impl: "{kind: tool|image|source, ref: ref}"
          download: "!string"                       # image:// git:// pkg://
          scheme: "{ops: []string, params: map}"
          usage: "?string"
          llmPrompt: "!string"
          nameResolution: "无前缀=默认org | 带前缀=github.com/<org>/<name>"   
        observation:                                # 观测面（意图；数据在 MetricsInfo）
          metrics: "[]{name, query, thresholds: ?[]{op: gt|gte|lt|lte, limit: number}}"   # thresholds=告警面声明（v0.33 数据面首段：装配层阈值桥绑定锚——声明面外谓词不绑定不即兴语义）
          logs: "?{ref: ref}"
        data:                                       # 数据面（Ref 解析权威之一）
          entities: "[]{name, store: ref, access}"
        flows:                                      # 流程面
          "[]{name, stages: []string}"
        runtime:                                    # 运行状态观测面
          probes: "[]{httpGet|tcp|version, period}"
        tools:                                      # 工具与调用描述面（Ops 的描述来源）
          "[]{op, params, idempotent, undo: full|partial|none, cleanup: ?op}"  # cleanup 成对
      status: {conditions: "[]Condition"}           # runtime 面探针回写

    - kind: ConnectorOps             # 工具调用面（v0.13 收窄：描述全在 WorkService）
      builtin: true
      storage: "global/connector-ops/<name>.yaml | systems/<s>/connector-ops/<name>.yaml | systems/<s>/instances/<i>/connector-ops/<name>.yaml"
      spec:
        workservice: "!ref"                          
        workflow: "?string (=default-task)"
        config: "?map"                               # 逻辑声明（endpointRef 等，A46）
        auth: "?{provider: ref}"
        ops: "[]string"                              # ⊆ WorkService.tools（Lint）
        flow: "?{nodes: []{name: !string, title: ?string, unit: session|agent, op: ?string, connector: ?string, provider: ?string}}"   # 节点流转骨架（workflow.md §1.3 执行模型）：节点执行单元=session|agent（global/服务内=session、跨服务=agent 调 agent），op/connector/provider=节点接线声明投影。骨架由 cored 绑定物化器在 workflow apply（创建/变更）时写入/更新（metadata.labels[core/origin]=workflow-bind=物化器所有面标注——非该标注的对象不盲写）；真实实现代码（connectOps）=apply-workflow agent 余量，运行时不做代码即兴

    - kind: Workflow                 # v0.8 吸收 Workflow
      builtin: true
      storage: "global/workflows/<name>.yaml | systems/<s>/workflows/<name>.yaml"
      spec:
        enabled: "?bool (=true)"                       # 规则开关：false=停用（不触发不消耗）
        services: "{involved: []ref, mode: observe|act}"   # 涉及哪些 service、怎么涉及
        llm: "?{provider: ref, model: string}"         # 该 workflow 用的 LLM
        triggers: "{events: {types: []string, minLevel}, timer: {resync}}"
        ingest: "{dedup: {window}, group: {by, max}, filterLevels}"
        connector: "?ref (=<同名 ConnectorOps>)"        # 默认绑定的 ConnectorOps（缺省同名；apply 时开始实现节点流转；节点执行单元=session|agent，global/服务内=session、跨服务=agent 调 agent——concepts/workflow.md §1.3）
        phases: "[]Phase"                              # 含 Paused（挂起态，见 onEscalate）；节点执行单元=session|agent（workflow.md §1.3）；引擎词汇基座=Agent.status.phase 八相位 + workflow 域交付锚 Deploying（concepts/workflow.md §1 映射表）
        transitions: "?[]{from: Phase, on: string, to: Phase}"   # 显式迁移表（workflow.md 规则 14；v0.25 首段=builtin 表=序列物化相邻行+审批边表行先例；v0.27 对象自声明表解析首段=object 源投影（规则 16）——v0.28 相位投影与迁移判定接线收口）
        onEscalate: "{action: pause|finish-tick|continue(默认 pause), notify: bind-role}"   # AUTO 处理不了的事件到达时的暂停策略（可配置）
        onFailed: "{backoff, maxAttempts, escalate: {to: root, notify: role-holders}}"
        brakes: "{cooldown, maxMutationsPerTick, avoidRepeatFix}"    # 强制默认（可收紧不可关闭；v0.32 刹车首段：cooldown 强制默认 2m ∧ 变更上限强制默认 10——宽松方向声明一律钳回默认=不可关闭机读落点；冷却门键=锚+involved 防震荡语义粒度=同一对象；变更上限=推进拍锚级只计实际迁移；cooldown 锚首段=run 受理时刻；avoidRepeatFix 随 DriftReport 回归声明面顺延）
        regression: "{statusWriteback: bool, reports: [DriftReport, SessionLog]}"
      builtinSet: [core-workflow, default-task, simple-shot, watch-only, dev-delivery, apply-workflow]   # 改 builtin 项=高敏权限项；六项相位词汇映射见 concepts/workflow.md §1（运行时注册表同源）；预算/成本归 metrics 域（暂不做）

    - kind: Proposal                 # 需求提案（concepts/proposal.md——console 提需求→审批→派工的一等对象；能力 B，2026-11-24 用户裁决）
      builtin: true
      storage: "global/proposals/<name>.yaml | systems/<s>/proposals/<name>.yaml"
      spec:
        title: "!string"                     # 标题
        summary: "!string"                   # 目标（做什么/为什么——批准后随派单载荷交给执行体）
        acceptance: "[]string"               # 验收标准（可验证条目——Goal.acceptance 同词汇）
        impact: "?string"                    # 影响面（涉及系统/对象/风险说明）
        priority: "low|normal|high|critical" # 优先级（低危/高危审批分流判定输入——分流配置=部署侧 --proposal-auto-approve）
        proposer: "!string"                  # 提出人（主体标识——审计回链 apply_log.subject）
        status: "!draft|submitted|approved|rejected|landed"  # 状态闭合集（迁移纪律=写面主体：draft=本地草稿未入册→submitted=提交待审→approved|rejected=判定→landed=产物落盘回写；checker=枚举闭合集，迁移序=主体纪律）
        landing: "?{target: doc|task, ref: ?string}"  # 产物指针（批准后执行体落点声明：doc=docsRoot 对象草稿 / task=部署侧任务文件；产物真实落盘=apply-workflow agent 真实拉起段，declared-deferred）

    - kind: Provider                 # 账号包
      builtin: true
      storage: global/providers/<name>.yaml
      spec:
        kind: "llm|im|email|cloud|network|generic"   # network=网络出口（代理/DNS/域名转发，声明/运行时分离见 concepts/provider.md 规则 6）
        auth: "{scheme, secret: {source: 'secret:|vault:|env:|keychain:', rotation}}"   # 明文禁（引用式取值四前缀——provider.md 规则 2；secret:=部署侧秘密库 runtime.storage.secrets，v0.34 起与 vault:/env:/keychain: 并列）
        config: "?map"
        connector: "?ref"                            # 可为空

    - kind: Notification             # 一 Scheme 一 Instance
      builtin: true
      storage: global/notifications/<name>.yaml
      spec:
        channel: "{kind: im|email|webhook, provider: ref}"
        on: "{events: []string, levels: []string}"
        route: "{to: role-holders, dedup: {window}, minLevel}"   # v0.35 阶段门段：minLevel=等级下限（低于下限零受理 filtered 计数可见）；窗口内同 dedupKey 至多一封（deduped 计数可见）——通知只由相位终态事实触发（notification.md 规则 5，apply 写入不触发用户通知）
        template: "?{ref}"                       # 文案模板=引用包装对象（{ref: <ref>}——notification.md §3 例与存量对象同形态；v0.48 净入 lint 对账修正：?ref 标量声明与概念例/存量对象劈叉，声明面单点修正收口)
      status: {delivered: "{last24h, failed}"}

    - kind: User                     # 用户系统 builtin
      builtin: true
      storage: global/users/<name>.yaml
      spec:
        display: "!string"
        contact: "{im: {provider, userId}, email}"
        roles: "[]ref"
        groups: "[]ref"
      status: {lastSeen: timestamp}

    - kind: Group                    # 组只绑 role
      builtin: true
      storage: global/groups/<name>.yaml
      spec: {roles: "[]ref"}                         # 禁权限字段

    - kind: Role                     # 权限 + 对象挂载
      builtin: true
      storage: global/roles/<name>.yaml
      spec:
        on: "?[]ref"                                 
        permissions: "[]{scope: 'global|systems/<n>/*'!, actions: [read|write|approve|call|publish|*]!}"   # 条目子字段必填=decl 后缀 !（先例 ref!）；条目结构完整=授权面与写面闸门共同语义（permission.LoadObjects 同判）；publish=事件发布域动作（event.md 规则 8 v0.20——发布权限域与 apply 写权限域分立）
      status: {holders: "[]User"}                    # user-system 投影
      builtinSet: [global/admin]                     # 全权限不可删

    - kind: Ref                      # 声明式逻辑引用
      builtin: true
      storage: "global/refs/<name>.yaml | systems/<s>/refs/<name>.yaml"
      spec:
        resolve: "{system: ref!, type: code|doc|url|resource|service, params: map}"
        purpose: "?string"
        ttl: "?duration"
      status: {lastResolution: "{at, by, value, cacheTtl}"}
      rules: [spec 禁物理地址, 解析链 ≤2, owner 必挂]

    - kind: TaskScheme               # 四层执行第三层
      builtin: true
      storage: systems/<s>/schemes/<name>.yaml
      spec:
        trigger: "{events: []string, match: map}"
        steps: "[]{id, connector: opsRef, op, args}"
        params: "[]string"
        onFailure: "{retry, backoff, then: escalate}"

    - kind: WorkServiceTemplate      # 安装模板 builtin Scheme
      builtin: true
      storage: global/templates/<name>.yaml
      spec:
        provides: "{capabilities: []string, category}"
        resources: "{images: [{name, source, digest!}], storage, ports, min}"
        install: "[]{action}"                       # digest 必填；幂等；三段 undo
        postInstall: "[{action: register-workservice}]"   # 产出六面体骨架
        asks: "[]{field, question}"

    - kind: Skill                    # 工具使用缓存
      builtin: true
      storage: systems/<s>/instances/<i>/skills/<name>.yaml
      spec:
        sourceOps: "!ref"
        when: "?{task}"
        prompt: "!string"
        examples: "[]{op, args, outcome}"           # 脱敏
        origin: "{sessionRef, cachedAt}"
        cachePolicy: "{ttl: 30d, invalidateOn: {providerVersionChange: true}}"

    # 不注册（按裁决）：
    #   Event —— 短生命周期数据非文档对象，队列由 dsh core plugin 实现
    #   ConnectorProvider —— 并入 WorkService.connector
    #   Bind —— 并入 Role.on；Owner/Finalizer —— 删除
    #   Status —— 降级为字段约定

  # ============================================================
  # 二、文档类型（docTypes 完整注册）
  # ============================================================
  docTypes:
    - {name: Goal,           mandatory: required, sections: [intent, acceptance],          writers: [human, "role:owner"], appliesTo: ServiceDesc, admission: normal}
    - {name: Desc,           mandatory: required, sections: [overview, boundaries],        writers: [human, "role:owner"], appliesTo: ServiceDesc, admission: normal}
    - {name: Rule,           mandatory: default,  sections: [rules],                       writers: [human, "role:owner"], appliesTo: ServiceDesc, admission: normal}   # 条目带 ruleId?/type: policy|lint；type=lint ⇒ 必挂 checker（rule.md R-TYPE）
    - {name: WorkflowDesign, mandatory: optional, sections: [flows],                       writers: [human],              appliesTo: ServiceDesc, admission: normal}
    - {name: Metrics,        mandatory: default,  sections: [indicators, thresholds],      writers: [human, "role:owner"], appliesTo: ServiceDesc, admission: normal}
    - {name: DataDesign,     mandatory: optional, sections: [entities, flows],             writers: [human],              appliesTo: ServiceDesc, admission: normal}
    - {name: DataMetrics,    mandatory: optional, sections: [flows, slo],                  writers: [human],              appliesTo: ServiceDesc, admission: normal}
    - {name: Auth,           mandatory: default,  sections: [subjects, grants],            writers: [human],              appliesTo: ServiceDesc, admission: strict}
    - {name: DeployDesign,   mandatory: optional, sections: [envs, rollout],               writers: [human],              appliesTo: ServiceDesc, admission: normal}
    - {name: CodeDesign,     mandatory: optional, sections: [modules],                     writers: [human],              appliesTo: ServiceDesc, admission: normal}
    - {name: ImplDoc,        mandatory: required, sections: [how, mapping],                writers: [agent],              appliesTo: ServiceInstance, admission: normal}
    - {name: VerifyDoc,      mandatory: required, sections: [guarantees, probes],          writers: [agent, "role:owner"], appliesTo: ServiceInstance, admission: normal}
    - {name: DriftReport,    mandatory: auto,     sections: [drifts, actions],             writers: [workflow-engine],    appliesTo: ServiceInstance, admission: normal}
    - {name: Runbook,        mandatory: optional, sections: [],                            writers: [human],              appliesTo: ServiceInstance, admission: normal}
    - {name: Requirements,   mandatory: required, sections: [archive],                     writers: [human],              appliesTo: core_system 存档, admission: strict}
    - {name: PlanDoc,        mandatory: required, sections: [intent, steps],               writers: [agent],              appliesTo: ServiceInstance, admission: normal}
    - {name: SessionLog,     mandatory: required, sections: [phases, docsLoaded, outcome], writers: [agent],              appliesTo: ServiceInstance, admission: normal}
    - {name: Memory,         mandatory: optional, sections: [content, refs],               writers: [agent],              appliesTo: ServiceInstance, admission: normal}   # Agent 记忆条目（默认记忆系统=Service 文档管理）
    # mandatory=注册级；Service 可加严；Builtin 无特殊档

  # ============================================================
  # 三、事件类型
  # 发布面（v0.20，event.md 规则 8）：全部注册类型可经写面注入端点发布——类型注册校验
  # +level 缺省=defaultLevel 投影+同管线去重/派发/归档（零旁路）；发布动作=publish × events
  # 资源域，与 apply 写权限域分立；注入通道的 source 由服务侧判定恒 human（本表 source 列
  # 是引擎内部来源声明，不是注入载荷参数）。
  # ============================================================
  eventTypes:
    # ExternalChanged（event.md 规则 12 v0.30 watch 桥首段）：装配层驱动 connectors.Watch 契约通道自产——source 服务侧判定恒 connector，dedupKey=驱动侧因由声明（<obj>:<change>:hash 投影），detail 不进总线；落 run 同拍即拍（workflow.md 规则 15 v0.30 段）。
    - {type: ExternalChanged,  source: connector, defaultLevel: info,  dedup: "<obj>:<change>:hash", route: owner-workflow}
    # MetricThreshold（event.md 规则 13 v0.31 阈值桥首段）：装配层阈值桥数据面样本通道服务侧判定自产——source 服务侧判定恒 metrics，dedupKey=<metric>:<window>（同指标同窗口越限聚合防风暴）；阈值判定是服务侧动作（数据面只产样本不带判定）；落 run 同拍即拍（workflow.md 规则 15）。
    - {type: MetricThreshold,  source: metrics,   defaultLevel: warn,  dedup: "<metric>:<window>",   route: owner-workflow}
    - {type: DocApplied,       source: core,      defaultLevel: info,  dedup: "apply 落库必发；engine 自写豁免",     route: owner-workflow}
    - {type: DriftFixed,       source: engine,    defaultLevel: info,  dedup: "<workflow>:<tick>",       route: observation-bus}
    - {type: ApprovalNeeded,   source: "permission|core",defaultLevel: warn,  dedup: "<rule>:<obj>",        route: notification/*}   # 双产源：permission=人审档拒绝路径（v0.21）；core=Proposal 提交高危分流自产（v0.51，dedup=proposal:<obj>——concepts/proposal.md 规则 5）
    - {type: ApprovalGranted,  source: human,     defaultLevel: info,  dedup: "approval:<obj>",      route: dead-letter}   # 审批信号族（workflow.md 规则 13/event.md 规则 11）：run 域迁移事件——involved=Workflow 引用即锚；dispatch 派发面无目标
    - {type: ApprovalRejected, source: human,     defaultLevel: warn,  dedup: "approval:<obj>",      route: dead-letter}   # 同上：驳回=闸锚终态迁移
    - {type: Escalate,         source: engine,    defaultLevel: error, dedup: "<workflow>:<reason>",     route: owner-workflow+root+notification/*}   # owner-workflow 收到后按其 onEscalate 策略决定暂停
    - {type: HeartbeatLost,    source: core,      defaultLevel: error, dedup: "<agent>",             route: parent-agent+notification/*}
    - {type: InstallRequested, source: human,     defaultLevel: info,  dedup: "<template>:<target>", route: workflows/deploy-system-workflow}
    - {type: TemplateFound,    source: connector, defaultLevel: info,  dedup: "<org>:<template>",    route: workflows/deploy-system-workflow}
    - {type: SessionClosed,    source: agent,     defaultLevel: info,  dedup: "<session>",           route: instance-docs}
    - {type: ReconcileTick,    source: core,      defaultLevel: info,  dedup: "reconcile:<at>",      route: workflows/core-workflow}   # 定时巡检到点（level-triggered 保险丝，event.md §7；引擎内部源二段=cored 定时器自产）

  # ============================================================
  # 四、事件路由（穷举缺省 + 死信兜底）
  # ============================================================
  dispatch:
    - {on: ExternalChanged,  for: "owner:*",        dispatch: owner-workflow}
    - {on: MetricThreshold,  for: "owner:*",        dispatch: owner-workflow}
    - {on: DocApplied,       for: "owner:*",        dispatch: owner-workflow}
    - {on: DriftFixed,       for: "any",            dispatch: observation-bus}
    - {on: ApprovalNeeded,   for: "any",            dispatch: notification/*}
    - {on: ApprovalGranted,  for: "any",            dispatch: dead-letter}   # 审批信号族：run 域迁移事件（workflow.md 规则 13）——dispatch 派发面无目标，不另起新 run
    - {on: ApprovalRejected, for: "any",            dispatch: dead-letter}   # 同上
    - {on: Escalate,         for: "any",            dispatch: root-agent, cc: notification/*}
    - {on: HeartbeatLost,    for: "owner:agent",    dispatch: parent-agent, cc: notification/*}
    - {on: InstallRequested, for: "any",            dispatch: workflows/deploy-system-workflow}
    - {on: TemplateFound,    for: "any",            dispatch: workflows/deploy-system-workflow}
    - {on: SessionClosed,    for: "owner:instance", dispatch: instance-docs}
    - {on: ReconcileTick,    for: "any",            dispatch: workflows/core-workflow}   # 巡检到点→一次全量对账 run（对账锚=系统级循环本体）；v0.29 事件驱动推进源首段：该行落 run 的同拍即一次全量推进拍（source=core 三重闸——workflow.md 规则 15 v0.29 段，推进总闸同闸不旁路）
    - {on: "*",              for: "scope:builtin/*", dispatch: workflows/builtin-root}
    - {on: "*",              for: "any",            dispatch: dead-letter}   # 死信必须可见（词汇=hub 合成死信/eventbus routed_to 同一词）

  # ============================================================
  # 五、运行时
  # ============================================================
  runtime:
    baseAgent: dsh                  # 外部运行时（deepseek 插件机制）；core 即 dsh plugin
    executionLayers: [intent, agent, scheme, code]
    workflows: {default: default-task, builtin: [core-workflow, default-task, simple-shot, watch-only, dev-delivery, apply-workflow]}
    autoLevels: {rootAgent: "ask（默认，可配置 auto/all）", serviceAgent: auto, all: "需显式授权+权限系统放行"}  
    escalation: root                
    permissionSystem:
      defaultAdminRole: global/admin
      binding: [user->role, user->group->role]      # 只绑 role
      scopes: [global, system]
      agentJudgement: [agent, role, parent-agent]   # 灵活；跨系统 B 侧裁判
      hardRules: [deny delete agent/builtinRoot, deny disable scheme, deny delete role/global/admin]
    providerResolution: {defaultOrg: friddle-self-recursive-agent, prefixed: "github.com/<org>/<name>"}
    userSystem: builtin             
    events:
      levels: [debug, info, warn, error, fatal]     
      debug: 只进内存队列不落盘
      fatal: 必接 Notification
      ingestDefaults: {dedupWindow: 5m, groupMax: 100}
    sessions: {record: [PlanDoc, SessionLog]}       
    skills: {cachePolicy: {ttl: 30d, invalidateOnProviderVersionChange: true}, promote: "高频→Global 走人审"}
    storage:
      docs: "git 布局 | sqlite docs DB（二选一单权威，--docs-backend 装配；缺省 git=既有行为面零翻转。sqlite 后端：doc 表=文档内容权威（(kind,scope,name) 键+resourceVersion 乐观锁），git 布局降级为导入导出格式——空库一次性种子导入（保留 rv），非空库不导入=幂等迁移路径；Repository 接口同一套写语义（串行写+乐观锁+apply_log 双写），后端可换=接口既定）"
      opsLog: SQLite WAL（六表；DB=操作流水可丢弃可重建；每日对账）
      secrets: "SQLite 0600 独立文件（部署侧凭据真实值权威运行时面——secret: 引用的解析源；独立于 ops DB（秘密非缓存不可重建）与 docs DB（脱敏红线——值只进不出，任何查询/导出面只出名称零值出）；缺省关断，--secrets-db 显式装配）"
      refResolution: {cacheTtl: 10m, 失败降级旧缓存+Event}

  # ============================================================
  # 六、字段约定（conventions；A51/A52 后 Concepts 降级至此）
  # ============================================================
  conventions:
    status:                          
      observedGeneration: int        # 收敛到 Spec 第几代
      phase: string                  # 对象级粗状态
      conditions: "[]{type, status: True|False|Unknown, reason(枚举注册), message, lastTransitionTime}"
      drift: "{summary, reportRef}"  # Instance 必有
      writers: {status: workflow-engine(fieldManager), spec: human/agent-by-role}   # P2/P6
    tags:                            # 检索元数据（LLM 自动打标；Session/Doc/模板更新时）
      spec.metadata.tags: "?[]string"          # 关键 tag（含 tag 相关性派生排序）
      spec.metadata.importance: "?int(1-5)"    # 重要程度（LLM 评判）
      spec.metadata.digest: "?string"          # 一句话标题+描述（检索列表展示用）
    cascade:                         
      rule: 目录即对象——删目录=级联删除其下一切
      crossDir: 跨目录引用走 Ref或跨系统提案
      externalCleanup: WorkService.tools 成对 cleanup op（Lint 强制；无系统级 Finalizer）
    metadata:
      reserved: [metadata, spec, status]
      labels: 自由注释性元数据，无选择语义
      frontmatter: name/docType/scope/labels/roles
`
