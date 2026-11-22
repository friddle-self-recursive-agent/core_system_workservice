# core_system_workservice — 核心代码

> 三仓之一：声明在 [core_system](https://github.com/friddle-self-recursive-agent/core_system)（ServiceDesc），实现形式在 [core_system_instance](https://github.com/friddle-self-recursive-agent/core_system_instance)（ServiceInstance），**本仓放代码**：Igniter、Core 进程（cored）、dsh 插件、内置 Connector。builtin 系统=代码+connector 内置（R1.3），即本仓。

## 骨架（当前状态：可编译的骨架，不含实现）

```
cmd/igniter      点火器：装载 core/ → 启动 cored → 消失（预算 <1000 行，instance 仓 loop/bootstrap.md）
cmd/cored        Core 常驻进程：Apply/Event/Loop 引擎/Scheme/Policy（instance 仓 impl/self-recursive-core.md）
internal/
  scheme/        类型注册表载入与自检（core_system/core/scheme.yaml）
  repository/    文档库 git 后端（get/apply/list/watch，串行写队列）
  apply/         唯一写路径：scheme 准入→权限准入→乐观锁→git commit + ops DB 双写 + DocApplied 必发
  draft/         草稿缓存（编辑与发布分离：代改代审先落稿，apply 落真实差异后清空）
  eventbus/      事件队列：dedupKey 聚合、dispatch 路由、死信
  workflow/      Workflow 引擎（v0.8 吸收 Loop：触发规则+态机+刹车+成本，A41）
  permission/    权限系统 v0.9（Role/Group 对象 + Check/ByRole + A44 硬规则）
  opslog/        SQLite ops 日志（apply_log/event_log 双轨，instance 仓 runtime/ops-log.md）
connectors/      gitfs / dsh —— builtin 幂等驱动实现（五方法契约；类型绑定在 WorkService 六面体 connector 面，A53）
                 用法型驱动（onedev/nas/oss 等）不进本仓：放部署侧仓 sybran_usage_connector（rule.md R-ENG-3）
internal/workservice/ WorkService 六面体（连接/观测/数据/流程/运行状态/工具，A53）
internal/refresolve/ 声明式 Ref 运行时解析（v0.10：逻辑声明→owning system 返回真实地址，A46）
internal/workflow/ Workflow 引擎（v0.8 吸收 Loop：触发规则+态机+刹车+成本，A41）
# v0.9（A43）：系统本身无 plugin 概念——dsh 为外部运行时，core 以 dsh plugin 实现；
# core/console/exec-local 等承载为 dsh 侧实现说明（见 instance 仓 runtime/），不在本仓建 plugin 目录
```

## 对应关系

| builtin 服务 | 代码位置 |
|---|---|
| self-recursive-core | cmd/igniter, cmd/cored, internal/* |
| builtin-developer-system | connectors/dsh；onedev 驱动在 sybran_usage_connector 仓 |
| builtin-deploy-system | plugins/execlocal, connectors/gitfs |
| builtin-storage-system | connectors/gitfs；nas/oss 驱动在 sybran_usage_connector 仓 |

## 开发约定

- 设计先行：任何本仓代码必须能指回 core_system 的一条声明或 instance 仓的一份设计（P1：没有文档的对象是黑箱）。
- 驱动分仓（rule.md R-ENG-3）：本仓 connectors/ 只收 builtin 驱动（gitfs/dsh）；用法型驱动放部署侧仓（如 sybran_usage_connector），依赖方向恒为用法仓→本仓。
- 契约测试：Provider 五方法三例（首跑/重跑幂等/失败路径）先于实现；Ops 是运行时挂载（环境+op子集+Workflow），不写死在驱动里。
- 骨架期的每个 `// TODO` 都对应 open-questions 或落地顺序里的一项（instance 仓 feasibility.md §12）。
