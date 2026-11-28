// Package docsdocs —— connectops-docs：文档/发布域 OpsConnector（2026-11-28 用户裁决
// 「git提交这个不是功能。删掉功能。做成workflow的一个节点。即gh-release包括git push/release
// [scheme调用]」——workflow.md §1.3 / R-DOC-7 规则 2 修订）。
//
// 职责：把「docs DB → git 导出推送」从常驻定时器（旧 Run loop+docs_export_config 独立配置面
// 已删）改为 **release 域 op**——workflow 的 gh-release 节点（unit=connectops）执行时拍一发一：
//   - git-push：一拍导出（enumerate→写树→有变化才 commit→remote 声明才 push；幂等=水位
//     无变化零提交零推送）；过滤=scheme docTypes 声明唯一真相源（category=runtime 缺省排除）。
//   - release-publish：复合发布 op（gh-release 节点声明词形）=①git-push ②委托 connectops-gh
//     的 release-create（注册表晚期绑定——gh 域逻辑零复制）。gh 连接器缺席=显式失败不静默。
//
// 配置：ConnectorOps/global/connectops-docs spec.config {dir, remote}（声明值；dir=导出目标
// 工作树必填，remote=可选 push 远程，空=本地 commit 止步 no force-push）。装配层经 ApplyConfig
// 注入（registry.ApplyConfig 契约——对象声明现读）。注册=装配层显式 Register（需 docs 库句柄，
// 非 init 自注册——codeaudit 无状态才 init 自注册，两形态并存）。
package docsdocs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/friddle-self-recursive-agent/core_system_workservice/connectors"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// TypeName 注册名（=ConnectorOps 对象名 connectops-docs）。
const TypeName = "connectops-docs"

// Version 实现版本（协议演进只加不改）。
const Version = "0.1.0"

// Config 连接配置（ConnectorOps spec.config 声明值，零明文——remote 认证走部署侧 git 凭据）。
type Config struct {
	Dir    string `json:"dir"`    // 导出目标工作树（必填才可执行——显式缺口不猜测）
	Remote string `json:"remote"` // 可选 push 远程（空=本地 commit 止步）
}

// Connector docs/发布域 OpsConnector 实现（一个实例=一份导出目标绑定）。
type Connector struct {
	mu   sync.Mutex
	repo repository.Repository // docs 库只读半边（导出枚举源）
	sch  *scheme.Scheme        // 对象规范路径解析面（布局唯一具体分支）
	cfg  Config
	exp  *repository.Exporter // cfg 就绪后构建（惰性——ApplyConfig 时）
}

// New 构建（repo/sch 必填——导出=docs 库读侧派生任务；cfg 经 ApplyConfig 声明注入）。
func New(repo repository.Repository, sch *scheme.Scheme) *Connector {
	return &Connector{repo: repo, sch: sch}
}

// Type 契约实现（connectors.OpsConnector）。
func (c *Connector) Type() string { return TypeName }

// Version 契约实现。
func (c *Connector) Version() string { return Version }

// Ops op 集（gh-release 节点声明词形=release-publish；git-push 可独立声明为纯推送节点）。
func (c *Connector) Ops() []string { return []string{"git-push", "release-publish"} }

// ApplyConfig registry 可配置契约（装配层从 ConnectorOps spec.config 现读注入；dir 变更=重建
// 导出器——目标树绑定随声明走）。
func (c *Connector) ApplyConfig(raw []byte) error {
	var cfg Config
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return fmt.Errorf("connectops-docs: config 不可解析: %w", err)
		}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cfg = cfg
	if strings.TrimSpace(cfg.Dir) == "" {
		c.exp = nil // dir 未声明=导出面缺席（op 执行时显式报错——诚实缺口不猜测目标树）
		return nil
	}
	exp, err := repository.NewExporter(c.repo, c.sch, cfg.Dir, cfg.Remote)
	if err != nil {
		return err
	}
	c.exp = exp
	return nil
}

// Invoke 执行一次 op（调用即效果）：git-push=一拍导出；release-publish=git-push+委托
// connectops-gh release-create（params 原样透传——owner/repo/tag/title/body 随节点声明）。
func (c *Connector) Invoke(ctx context.Context, op string, params []byte) ([]byte, error) {
	if !c.hasOp(op) {
		return nil, fmt.Errorf("op %s 不在 %s 声明集 %v", op, TypeName, c.Ops())
	}
	docs, err := c.push(ctx)
	if err != nil {
		return nil, err
	}
	if op == "git-push" {
		return mustJSON(map[string]any{"op": op, "docs": docs}), nil
	}
	// release-publish 第二跳：connectops-gh release-create（晚期绑定——注册表缺席=显式失败）。
	gh, ok := connectors.LookupOpsConnector("connectops-gh")
	if !ok {
		return nil, fmt.Errorf("release-publish 委托缺席：connectops-gh 未注册（驱动包未导入/装配缺席——发布第二跳不可执行，显式失败不静默）")
	}
	rel, err := gh.Invoke(ctx, "release-create", params)
	if err != nil {
		return nil, fmt.Errorf("release-publish 第二跳 release-create 失败（docs push 已完成 commit=%s）: %w", docs.Commit, err)
	}
	return mustJSON(map[string]any{"op": op, "docs": docs, "release": json.RawMessage(rel)}), nil
}

// Probe 通道可达性探测（不产生任何提交）：dir 已声明=就绪；未声明=显式缺口。
func (c *Connector) Probe(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.exp == nil {
		return fmt.Errorf("connectops-docs: 导出目标未声明（ConnectorOps spec.config.dir 缺席——诚实缺口不猜测）")
	}
	return nil
}

// push 一拍导出（exporter.mu 串行化——并发调用安全；水位无变化=幂等零提交零推送）。
func (c *Connector) push(ctx context.Context) (repository.ExportResult, error) {
	c.mu.Lock()
	exp := c.exp
	c.mu.Unlock()
	if exp == nil {
		return repository.ExportResult{}, fmt.Errorf("connectops-docs: 导出目标未声明（ConnectorOps spec.config.dir 缺席——诚实缺口不猜测）")
	}
	return exp.Export(ctx)
}

func (c *Connector) hasOp(op string) bool {
	for _, o := range c.Ops() {
		if o == op {
			return true
		}
	}
	return false
}

func mustJSON(v map[string]any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte(`{"error":"result encode"}`)
	}
	return b
}
