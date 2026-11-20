# core_system_workservice — 核心代码

> 三仓之一：声明在 [core_system](https://github.com/friddle-self-recursive-agent/core_system)（ServiceDesc），实现形式在 [core_system_instance](https://github.com/friddle-self-recursive-agent/core_system_instance)（ServiceInstance），**本仓放代码**：Igniter、Core 进程（cored）、dsh 插件、内置 Connector。builtin 系统=代码+connector 内置（R1.3），即本仓。

## 骨架（当前状态：可编译的骨架，不含实现）

```
cmd/igniter      点火器：装载 core/ → 启动 cored → 消失（预算 <1000 行，instance 仓 loop/bootstrap.md）
cmd/cored        Core 常驻进程：Apply/Event/Loop 引擎/Scheme/Policy（instance 仓 impl/self-recursive-core.md）
internal/
  scheme/        类型注册表载入与自检（core_system/core/scheme.yaml）
  repository/    文档库 git 后端（get/apply/list/watch，串行写队列）
  apply/         唯一写路径：Policy→SSA合并→乐观锁→git commit + ops DB 双写
  eventbus/      事件队列：dedupKey 聚合、dispatch 路由、死信
  workflow/      Workflow 引擎（v0.8 吸收 Loop：触发规则+态机+刹车+成本，A41）
  policy/        准入规则引擎 + 3 条硬规则
  opslog/        SQLite ops 日志（六表，instance 仓 runtime/ops-log.md）
  plugin/        dsh 插件宿主接口（五方法契约）
plugins/         control(root直控) / console(总控界面) / execlocal(本地执行) / reporter(上报)
connectors/      onedev / gitfs / nas / oss / dsh —— 类型层驱动实现 ConnectorProvider（v0.3 拆分，A25）
internal/workflow/ 执行态机 Workflow（builtin：default-task/simple-shot，A26）
```

## 对应关系

| builtin 服务 | 代码位置 |
|---|---|
| self-recursive-core | cmd/igniter, cmd/cored, internal/* |
| builtin-developer-system | connectors/onedev, plugins/execlocal |
| builtin-deploy-system | plugins/execlocal, connectors/gitfs |
| builtin-storage-system | connectors/gitfs, connectors/nas, connectors/oss |

## 开发约定

- 设计先行：任何本仓代码必须能指回 core_system 的一条声明或 instance 仓的一份设计（P1：没有文档的对象是黑箱）。
- 契约测试：Provider 五方法三例（首跑/重跑幂等/失败路径）先于实现；Ops 是运行时挂载（环境+op子集+Workflow），不写死在驱动里。
- 骨架期的每个 `// TODO` 都对应 open-questions 或落地顺序里的一项（instance 仓 feasibility.md §12）。
