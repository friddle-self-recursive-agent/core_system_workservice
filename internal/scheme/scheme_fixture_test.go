// scheme_fixture_test.go —— 权威声明逐字快照（生成自 core_design/core/scheme.yaml v0.13，
// 含 loop-10 Role 子字段必填锐化）。改动权威声明须同步再生成此夹具；
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
  # 一、对象类型（kinds，17 项；已采纳的墓碑概念不注册，见文末不注册清单）
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
      storage: global/workservices/<name>/workservice.yaml
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
          metrics: "[]{name, query}"
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

    - kind: Workflow                 # v0.8 吸收 Workflow
      builtin: true
      storage: "global/workflows/<name>.yaml | systems/<s>/workflows/<name>.yaml"
      spec:
        enabled: "?bool (=true)"                       # 规则开关：false=停用（不触发不消耗）
        services: "{involved: []ref, mode: observe|act}"   # 涉及哪些 service、怎么涉及
        llm: "?{provider: ref, model: string}"         # 该 workflow 用的 LLM
        triggers: "{events: {types: []string, minLevel}, timer: {resync}}"
        ingest: "{dedup: {window}, group: {by, max}, filterLevels}"
        phases: "[]Phase"                              # 含 Paused（挂起态，见 onEscalate）；引擎词汇基座=Agent.status.phase 八相位 + workflow 域交付锚 Deploying（concepts/workflow.md §1 映射表）
        onEscalate: "{action: pause|finish-tick|continue(默认 pause), notify: bind-role}"   # AUTO 处理不了的事件到达时的暂停策略（可配置）
        onFailed: "{backoff, maxAttempts, escalate: {to: root, notify: role-holders}}"
        brakes: "{cooldown, maxMutationsPerTick, avoidRepeatFix}"    # 强制默认（可收紧不可关闭）
        regression: "{statusWriteback: bool, reports: [DriftReport, SessionLog]}"
      builtinSet: [core-workflow, default-task, simple-shot, watch-only, dev-delivery, apply-workflow]   # 改 builtin 项=高敏权限项；六项相位词汇映射见 concepts/workflow.md §1（运行时注册表同源）；预算/成本归 metrics 域（暂不做）

    - kind: Provider                 # 账号包
      builtin: true
      storage: global/providers/<name>.yaml
      spec:
        kind: "llm|im|email|cloud|network|generic"   # network=网络出口（代理/DNS/域名转发，声明/运行时分离见 concepts/provider.md 规则 6）
        auth: "{scheme, secret: {source: 'vault:|env:|keychain:', rotation}}"   # 明文禁
        config: "?map"
        connector: "?ref"                            # 可为空

    - kind: Notification             # 一 Scheme 一 Instance
      builtin: true
      storage: global/notifications/<name>.yaml
      spec:
        channel: "{kind: im|email|webhook, provider: ref}"
        on: "{events: []string, levels: []string}"
        route: "{to: role-holders, dedup: {window}}"   
        template: "?ref"
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
    - {type: ExternalChanged,  source: connector, defaultLevel: info,  dedup: "<obj>:<change>:hash", route: owner-workflow}
    - {type: MetricThreshold,  source: metrics,   defaultLevel: warn,  dedup: "<metric>:<window>",   route: owner-workflow}
    - {type: DocApplied,       source: core,      defaultLevel: info,  dedup: "apply 落库必发；engine 自写豁免",     route: owner-workflow}
    - {type: DriftFixed,       source: engine,    defaultLevel: info,  dedup: "<workflow>:<tick>",       route: observation-bus}
    - {type: ApprovalNeeded,   source: permission,defaultLevel: warn,  dedup: "<rule>:<obj>",        route: notification/*}
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
    - {on: Escalate,         for: "any",            dispatch: root-agent, cc: notification/*}
    - {on: HeartbeatLost,    for: "owner:agent",    dispatch: parent-agent, cc: notification/*}
    - {on: InstallRequested, for: "any",            dispatch: workflows/deploy-system-workflow}
    - {on: TemplateFound,    for: "any",            dispatch: workflows/deploy-system-workflow}
    - {on: SessionClosed,    for: "owner:instance", dispatch: instance-docs}
    - {on: ReconcileTick,    for: "any",            dispatch: workflows/core-workflow}   # 巡检到点→一次全量对账 run（对账锚=系统级循环本体）
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
      docs: 本地文件夹 git（权威；机器预装 git 是默认假设；Repository 接口可换后端）
      opsLog: SQLite WAL（六表；git=内容版本，DB=操作流水；每日对账）
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
