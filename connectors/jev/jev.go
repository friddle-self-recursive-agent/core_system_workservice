// Package jev —— connectops-jev：判定域 OpsConnector（2026-11-28 用户裁定「judge要支持
// connectOps到jev来做判断，默认connectOpsJev。还是一样的习惯」——workflow.md §1.3 判断分叉
// 判定域缺省词形）。
//
// 职责：workflow 分叉的判定通路——judge 节点声明 unit=connectops + connector=connectops-jev
// + op=judge 时，ops 桥经本连接器调判定端点，result JSON 的 decision 字段=判断词
// （引擎 resume 时查表 (phase, decision) 命中即沿判断边走——opsrun.go/ops_bridge.go 分叉机制）。
//
// 习惯同族（connectops-gh/codeaudit 同款）：
//   - init 契约层自注册（RegisterOpsConnector——缺省实例零配置可用）；
//   - ApplyConfig registry 可配置契约（ConnectorOps spec.config 声明值注入——零明文，tokenEnv 引用式）；
//   - 定义与代码同仓习惯→本包在 workservice/connectors（builtin 域——判定域是平台件，非用法仓）。
//
// 通道：http 单通道（判定端点 REST——POST {baseUrl}/judge，params 原样透传，期望 JSON
// {decision, reason?}；decision 缺席=下游机械 advance 兜底不编造）。secret 口径：tokenEnv
// 引用式取值（provider.md 规则 2/6），声明面永不落明文。
package jev

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/connectors"
)

// TypeName 注册名（=ConnectorOps 对象名 connectops-jev——2026-11-28 用户裁定缺省词形）。
const TypeName = "connectops-jev"

// Version 实现版本（协议演进只加不改）。
const Version = "0.1.0"

// defaultTimeout Invoke 上限缺省（判定端点常规秒级——judge LLM 稍长，30s 起步）。
const defaultTimeout = 30 * time.Second

// Config 连接配置（ConnectorOps spec.config 声明值，零明文——token 只留 env 键名）。
type Config struct {
	BaseURL  string `json:"baseUrl"`  // 判定端点入口（必填才可执行——显式缺口不猜测）
	TokenEnv string `json:"tokenEnv"` // bearer token env 键名（缺省 JEVD_TOKEN）
	Timeout  int    `json:"timeout"`  // 秒（缺省 30）
}

// Connector 判定域 OpsConnector 实现（一个实例=一份判定端点绑定）。
type Connector struct {
	mu  sync.Mutex
	cfg Config
	hc  *http.Client
}

// init 契约层自注册（codeaudit 同款——缺省实例 baseUrl 缺席=Invoke 时显式失败不静默）。
func init() { connectors.RegisterOpsConnector(New()) }

// normalize 缺省填平（New/ApplyConfig 共用单点——声明缺席不产生零值字段）。
func normalize(cfg *Config) {
	if cfg.TokenEnv == "" {
		cfg.TokenEnv = "JEVD_TOKEN"
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = int(defaultTimeout.Seconds())
	}
}

// New 构建（无凭证安全：token 只在 Invoke/Probe 时从 env 解析）。
func New() *Connector {
	cfg := Config{}
	normalize(&cfg)
	return &Connector{cfg: cfg, hc: &http.Client{Timeout: time.Duration(cfg.Timeout) * time.Second}}
}

// Type 契约实现（connectors.OpsConnector）。
func (c *Connector) Type() string { return TypeName }

// Version 契约实现。
func (c *Connector) Version() string { return Version }

// Ops op 集（judge=判定 op——result.decision 即判断词）。
func (c *Connector) Ops() []string { return []string{"judge"} }

// ApplyConfig registry 可配置契约（ConnectorOps spec.config 声明值注入——并发安全原子换装）。
func (c *Connector) ApplyConfig(raw []byte) error {
	var cfg Config
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return fmt.Errorf("connectops-jev: config 不可解析: %w", err)
		}
	}
	normalize(&cfg)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cfg = cfg
	c.hc = &http.Client{Timeout: time.Duration(cfg.Timeout) * time.Second}
	return nil
}

// Invoke 执行一次 op（调用即效果）：POST {baseUrl}/judge，params 原样透传（owner/repo/
// subject/上下文随节点声明∪run 溯源派生——ops 桥同款），期望 JSON {decision, reason?}。
func (c *Connector) Invoke(ctx context.Context, op string, params []byte) ([]byte, error) {
	if op != "judge" {
		return nil, fmt.Errorf("op %s 不在 %s 声明集 %v", op, TypeName, c.Ops())
	}
	c.mu.Lock()
	baseURL := c.cfg.BaseURL
	tokenEnv := c.cfg.TokenEnv
	hc := c.hc
	c.mu.Unlock()
	if strings.TrimSpace(baseURL) == "" {
		return nil, fmt.Errorf("connectops-jev: 判定端点未声明（ConnectorOps spec.config.baseUrl 缺席——诚实缺口不猜测）")
	}
	token := os.Getenv(tokenEnv)
	url := strings.TrimRight(baseURL, "/") + "/judge"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(params))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("jev 判定端点调用失败 (%s): %w", url, err)
	}
	defer resp.Body.Close()
	out, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return out, fmt.Errorf("jev 判定端点 HTTP %d（%s）", resp.StatusCode, strings.TrimSpace(string(out)))
	}
	var parsed struct {
		Decision string `json:"decision"`
	}
	if json.Unmarshal(out, &parsed) != nil || strings.TrimSpace(parsed.Decision) == "" {
		// 透传原结果（decision 缺席=下游机械 advance 兜底——判定端点无词不编造）。
		return out, nil
	}
	return out, nil
}

// Probe 通道可达性探测（不发送任何判定内容）：baseUrl 已声明=就绪；未声明=显式缺口。
func (c *Connector) Probe(ctx context.Context) error {
	c.mu.Lock()
	baseURL := c.cfg.BaseURL
	c.mu.Unlock()
	if strings.TrimSpace(baseURL) == "" {
		return fmt.Errorf("connectops-jev: 判定端点未声明（ConnectorOps spec.config.baseUrl 缺席——诚实缺口不猜测）")
	}
	return nil
}
