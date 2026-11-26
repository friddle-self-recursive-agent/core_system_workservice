// routes_sessions.go —— 会话上下文契约面（G-14：会话页可读 ContextPack/TagCtxShard/
// takeBudget 且可编辑分叉——前后端契约面；R-UX-3 用户明确要求）。
//
//	GET  /v1/sessions/{id}/context          会话上下文投影（ContextPack 词形 + 分片 + 策略）
//	PUT  /v1/sessions/{id}/context          编辑分叉（takeBudget/tagSig 保守覆盖；写前快照、可撤销）
//	POST /v1/sessions/{id}/context/revert   撤销（恢复最近一次写前快照）
//
// 载体判定（G-14 任务书要求先探明）：TagCtxShard/ContextPack 在本仓=contextidx 包的
// **进程内派生物**（零 DocType、零落库、删掉可重建——contextidx.go 红线行），非 docs DB
// 对象——故编辑分叉不走 drafts/apply 两步写通路，落本文件保守 PUT 面：
//   - 写前快照：每次 PUT 前≈当前覆盖压入撤销栈（≤10 层），POST revert 弹栈恢复；
//   - 可撤销：revert 后 revision 单调推进（乐观锁观察者可见变更，不静默回跳）；
//   - 乐观锁：PUT 带 baseRevision 与当前 revision 不符=409（editor.js Apply 同语义——
//     不静默重基）；覆盖存进程内存（重启即失——诚实词形，不冒充持久化）。
//
// 诚实纪律：本代无三段式 Session 装配器（context.md §4）——pack.degraded 恒 true，
// chunks=索引 TopK 候选投影（GET /v1/context/find 同一查询），**不冒充会话真实装配包**；
// 会话不在活性账=session:null（不冒充在册）；分片缓存面缺席=shards:null。
// 写面门控：PUT/revert 随 --allow-apply 同开关（POST /v1/sessions/heartbeat 同口径——
// 写面动作与注入面同开关）；GET 恒开但 contextIdx 缺席=503（与 /v1/context/* 同防御面）。
package apiserver

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/contextidx"
)

// 会话上下文策略缺省（context.md §3 ContextPolicy.budget/packLimits/cache 声明值——
// 本代 ContextPolicy 对象未落地，先以常量缺省承载契约面；落地后改读对象面零词形变化）。
const (
	DefaultTakeBudget = 30000 // takeBudget 缺省（ContextPack.spec.query.takeBudget 设计示例值）
	MinTakeBudget     = 2000  // ContextPolicy.budget.minTakeTokens
	MaxTakeBudget     = 60000 // ContextPolicy.budget.defaultTakeMaxTokens
	MaxTagsPerType    = 5     // ContextPolicy.packLimits.maxTagsPerType（每 tagType ≤5）
	MaxTagsTotal      = 20    // ContextPolicy.packLimits.maxTagsTotal（总计 ≤20）
	ShardMaxTokens    = 6000  // ContextPolicy.cache.shardMaxTokens（分片大小）
	CtxChunkK         = 12    // chunks 候选投影 TopK（读面收敛——防拖库同口径）
	MaxTagKeyLen      = 128   // 单 tagKey 长度上限（tagType:tag 词形）
	MaxNoteLen        = 512   // 分叉备注长度上限
	MaxSessionIDLen   = 128   // 会话锚长度上限
	CtxRevertDepth    = 10    // 撤销栈深度（写前快照层数上限）
)

// sessionCtxOverride 会话上下文覆盖（编辑分叉的落点——进程内运行态，重启即失）。
type sessionCtxOverride struct {
	TakeBudget int64     `json:"takeBudget"`       // 提取预算（0–100 标度外的 token 数——context.md §3）
	TagSig     []string  `json:"tagSig,omitempty"` // 受限 tag 集（"tagType:tag" 词形，与 /v1/context/find tags 同形）
	Note       string    `json:"note,omitempty"`   // 分叉备注（人读）
	By         string    `json:"by,omitempty"`     // 写者词形（fieldManager 惯例，如 console:web）
	Revision   int64     `json:"revision"`         // 会话内单调递增（乐观锁基准；revert 亦推进）
	UpdatedAt  time.Time `json:"updatedAt"`        // 最近覆盖时刻（UTC）
}

// sessionCtxStore 覆盖账（并发安全；undo 栈=写前快照——G-14 保守实现）。
type sessionCtxStore struct {
	mu        sync.Mutex
	overrides map[string]*sessionCtxOverride   // sessionId → 现行覆盖
	undo      map[string][]*sessionCtxOverride // sessionId → 写前快照栈（栈顶=最近写前）
}

func newSessionCtxStore() *sessionCtxStore {
	return &sessionCtxStore{overrides: map[string]*sessionCtxOverride{}, undo: map[string][]*sessionCtxOverride{}}
}

// put 落覆盖（写前快照压栈；revision 接力单调推进）。返回新覆盖。
func (s *sessionCtxStore) put(id string, ov *sessionCtxOverride) *sessionCtxOverride {
	s.mu.Lock()
	defer s.mu.Unlock()
	prev := s.overrides[id]
	if len(s.undo[id]) >= CtxRevertDepth {
		s.undo[id] = s.undo[id][1:] // 栈满=丢最旧（有界——防拖内存）
	}
	s.undo[id] = append(s.undo[id], prev) // prev 可能 nil=「无覆盖」快照（revert 恢复为无覆盖）
	rev := int64(1)
	if prev != nil {
		rev = prev.Revision + 1
	}
	ov.Revision = rev
	ov.UpdatedAt = time.Now().UTC()
	s.overrides[id] = ov
	return ov
}

// revert 弹栈恢复（栈空=无快照，返回 nil,false 不动作）。恢复行 revision 同样推进。
func (s *sessionCtxStore) revert(id string) (*sessionCtxOverride, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := len(s.undo[id])
	if n == 0 {
		return nil, false
	}
	snap := s.undo[id][n-1]
	s.undo[id] = s.undo[id][:n-1]
	cur := s.overrides[id]
	rev := int64(1)
	if cur != nil {
		rev = cur.Revision + 1
	}
	if snap == nil {
		delete(s.overrides, id) // 恢复「无覆盖」基线
		return nil, true
	}
	snap.Revision = rev
	snap.UpdatedAt = time.Now().UTC()
	s.overrides[id] = snap
	return snap, true
}

// get 现行覆盖（只读投影）。
func (s *sessionCtxStore) get(id string) *sessionCtxOverride {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.overrides[id]
}

// canRevert 是否存在可撤销快照。
func (s *sessionCtxStore) canRevert(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.undo[id]) > 0
}

// shardCacheFace 分片缓存读面（contextidx.Index 满足——Shards() 可得 TagCtxShard 缓存；
// 可选能力：face 断言失败=分片区 null 不冒充，接口化便于测试注入有无两态）。
type shardCacheFace interface {
	Shards() *contextidx.ShardCache
}

// registerSessionCtxRoutes 会话上下文契约面注册（G-14；store=进程内覆盖账——
// handler 组闭包自持，与装配层零耦合）。
func registerSessionCtxRoutes(d *deps, mux *http.ServeMux) {
	store := newSessionCtxStore()
	mux.HandleFunc("GET /v1/sessions/{id}/context", func(w http.ResponseWriter, r *http.Request) {
		if d.o.contextIdx == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "会话上下文契约面未接线（contextidx 未装配——与 /v1/context/* 同防御面）"})
			return
		}
		writeJSON(w, http.StatusOK, sessionCtxPayload(d, store, r.PathValue("id")))
	})
	mux.HandleFunc("PUT /v1/sessions/{id}/context", func(w http.ResponseWriter, r *http.Request) {
		if d.applier == nil {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "写面关断（--allow-apply 未开启）——会话上下文编辑分叉随写面门控"})
			return
		}
		id := r.PathValue("id")
		if len(id) > MaxSessionIDLen {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "会话锚超长（≤" + strconv.Itoa(MaxSessionIDLen) + "）"})
			return
		}
		var req struct {
			TakeBudget   int64    `json:"takeBudget"`
			TagSig       []string `json:"tagSig"`
			Note         string   `json:"note"`
			By           string   `json:"by"`
			BaseRevision *int64   `json:"baseRevision"`
		}
		if err := json.NewDecoder(io.LimitReader(r.Body, 64<<10)).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "会话上下文覆盖载荷解析失败: " + err.Error()})
			return
		}
		if req.TakeBudget < MinTakeBudget || req.TakeBudget > MaxTakeBudget {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "takeBudget 越界（ContextPolicy 预算域 " +
				strconv.Itoa(MinTakeBudget) + "–" + strconv.Itoa(MaxTakeBudget) + "）"})
			return
		}
		tags, badTag := normalizeTagSig(req.TagSig)
		if badTag != "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "tagSig 词形非法: " + badTag + "（tagType:tag；总 ≤" + strconv.Itoa(MaxTagsTotal) + "、每 tagType ≤" + strconv.Itoa(MaxTagsPerType) + "）"})
			return
		}
		if len(req.Note) > MaxNoteLen {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "note 超长（≤" + strconv.Itoa(MaxNoteLen) + "）"})
			return
		}
		if cur := store.get(id); req.BaseRevision != nil && cur != nil && *req.BaseRevision != cur.Revision {
			writeJSON(w, http.StatusConflict, map[string]any{"error": "乐观锁冲突：覆盖已被他端推进（baseRevision 不符——稿不落，请对齐 revision 重试）",
				"revision": cur.Revision})
			return
		}
		by := strings.TrimSpace(req.By)
		if by == "" {
			by = "unknown"
		}
		ov := store.put(id, &sessionCtxOverride{TakeBudget: req.TakeBudget, TagSig: tags, Note: strings.TrimSpace(req.Note), By: by})
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "override": ov, "canRevert": store.canRevert(id)})
	})
	mux.HandleFunc("POST /v1/sessions/{id}/context/revert", func(w http.ResponseWriter, r *http.Request) {
		if d.applier == nil {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "写面关断（--allow-apply 未开启）——会话上下文撤销随写面门控"})
			return
		}
		id := r.PathValue("id")
		ov, ok := store.revert(id)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "无可撤销快照（撤销栈空——写前快照只在每次 PUT 时压栈）"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "override": ov, "canRevert": store.canRevert(id)})
	})
}

// sessionCtxPayload 会话上下文投影组装（GET 载荷——face 缺席=null 不冒充）。
func sessionCtxPayload(d *deps, store *sessionCtxStore, id string) map[string]any {
	ov := store.get(id)
	// 会话行：活性账现读（face 缺席/不在账=null——不冒充在册；Heartbeats 全量捞取后键匹配，
	// 活性行集有界（心跳登记≤会话数）——读面降级不放大）。
	var sess map[string]any
	if d.o.sessionLiveness != nil {
		rows, err := d.o.sessionLiveness.Heartbeats(500)
		if err == nil {
			for _, row := range rows {
				if row.SessionID == id {
					sess = map[string]any{
						"sessionId": row.SessionID, "state": row.State, "phase": row.Phase,
						"lastAction": row.LastAction, "taskName": row.TaskName, "longTask": row.LongTask,
						"heartbeats": row.Heartbeats, "healAttempts": row.HealAttempts,
						"lastBeat": row.LastBeat.UTC().Format(time.RFC3339Nano),
					}
					break
				}
			}
		}
	}
	// 生效值：覆盖 > 缺省（tagSigSource 标来源——装配面词形与覆盖词形分立）。
	takeBudget := int64(DefaultTakeBudget)
	tags := []string{}
	tagSigSource := "none"
	if ov != nil {
		takeBudget = ov.TakeBudget
		tags = ov.TagSig
		tagSigSource = "override"
	}
	// chunks：索引 TopK 候选投影（/v1/context/find 同一查询同一公式——tagSig 缺席=空集；
	// 本代无三段式装配器，degraded 恒 true 诚实标注——context.md 规则 1/13 降级词形）。
	chunks := []map[string]any{}
	if len(tags) > 0 {
		for _, h := range d.o.contextIdx.Find(contextidx.FindQuery{Tags: tags, K: CtxChunkK}) {
			chunks = append(chunks, map[string]any{
				"kind": h.Doc.Kind, "scope": h.Doc.Scope, "name": h.Doc.Name,
				"score": h.Score, "why": h.Why,
			})
		}
	}
	// shards：生效 tagSig 逐 tagKey 投影 TagCtxShard（face 断言失败=null；访问时重建语义
	// ——shard.go Get 缺席/stale 即重建，读面零额外失效路径）。
	var shards []map[string]any
	var shardsStats map[string]any
	if sc, ok := d.o.contextIdx.(shardCacheFace); ok {
		cache := sc.Shards()
		shards = make([]map[string]any, 0, len(tags))
		for _, key := range tags {
			typ, tag := splitTagKey(key)
			sh := cache.Get(typ, tag)
			top := sh.DocKeys
			if len(top) > 8 {
				top = top[:8] // 投影瘦身：全量走对象面（不冒充实全清单）
			}
			shards = append(shards, map[string]any{
				"tagKey": sh.TagKey, "docs": len(sh.DocKeys), "topDocs": top,
				"stale": sh.Stale, "version": sh.Version,
				"builtAt": sh.BuiltAt.UTC().Format(time.RFC3339Nano),
			})
		}
		total, stale := cache.Stats()
		shardsStats = map[string]any{"total": total, "stale": stale, "rebuildThreshold": cache.Threshold()}
	}
	// pack：ContextPack 词形契约投影（tokens=null——本代无真实 token 计数不冒充）。
	pack := map[string]any{
		"takeBudget":        takeBudget,
		"takeBudgetDefault": int64(DefaultTakeBudget),
		"tagSig":            tagSigRows(tags),
		"tagSigSource":      tagSigSource,
		"chunks":            chunks,
		"tokens":            nil,
		"degraded":          true,
		"degradedReason":    "本代无三段式 Session 装配器——chunks=索引 TopK 候选投影（context.md 检索序第一段词形），非会话真实装配包",
	}
	out := map[string]any{
		"sessionId": id, "session": sess, "pack": pack,
		"shards": shards, "shardsStats": shardsStats,
		"policy": map[string]any{
			"shardMaxTokens": ShardMaxTokens, "minTakeBudget": MinTakeBudget,
			"maxTakeBudget": MaxTakeBudget, "maxTagsPerType": MaxTagsPerType,
			"maxTagsTotal": MaxTagsTotal, "chunkCandidates": CtxChunkK,
		},
		"override":    ovRef(ov),
		"canRevert":   store.canRevert(id),
		"writeGate":   map[string]any{"open": d.applier != nil, "hint": "--allow-apply"},
		"generatedAt": time.Now().UTC().Format(time.RFC3339Nano),
	}
	return out
}

// splitTagKey tagKey 拆解（"tagType:tag" 词形；裸 tag=type 侧空——/v1/context/find
// tags 参数同词形，shard.go shardPseudoDoc 同款切分口径）。
func splitTagKey(key string) (string, string) {
	typ, tag, found := strings.Cut(key, ":")
	if !found {
		return "", key
	}
	return typ, tag
}

// tagSigRows 受限 tag 集投影（[{tagType,tag,tagKey}]——UI chips 直用词形）。
func tagSigRows(tags []string) []map[string]any {
	rows := make([]map[string]any, 0, len(tags))
	for _, key := range tags {
		typ, tag := splitTagKey(key)
		rows = append(rows, map[string]any{"tagType": typ, "tag": tag, "tagKey": key})
	}
	return rows
}

// ovRef 覆盖投影（nil=nil——无覆盖不冒充零值覆盖）。
func ovRef(ov *sessionCtxOverride) map[string]any {
	if ov == nil {
		return nil
	}
	return map[string]any{
		"takeBudget": ov.TakeBudget, "tagSig": ov.TagSig, "note": ov.Note,
		"by": ov.By, "revision": ov.Revision,
		"updatedAt": ov.UpdatedAt.Format(time.RFC3339Nano),
	}
}

// normalizeTagSig tag 集归一校验（保序去重；总 ≤20、每 tagType ≤5、单键 ≤128 且无
// 空白/控制字符——ContextPolicy.packLimits 装配面限制的写面落点）。违规=返回首个坏键。
func normalizeTagSig(tags []string) ([]string, string) {
	if len(tags) > MaxTagsTotal {
		return nil, "总计超限（" + strconv.Itoa(len(tags)) + " > " + strconv.Itoa(MaxTagsTotal) + "）"
	}
	seen := map[string]bool{}
	perType := map[string]int{}
	out := make([]string, 0, len(tags))
	for _, raw := range tags {
		key := strings.TrimSpace(raw)
		if key == "" {
			continue // 空项跳过（宽松收纳——写面只拦真违规）
		}
		if len(key) > MaxTagKeyLen || strings.ContainsAny(key, " \t\r\n\x00") {
			return nil, key
		}
		typ, tag := splitTagKey(key)
		if strings.TrimSpace(tag) == "" {
			return nil, key // "tagType:" 空名=坏键（裸 tag 合法——type 侧空）
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		perType[typ]++
		if perType[typ] > MaxTagsPerType {
			return nil, "tagType=" + typ + " 超限（> " + strconv.Itoa(MaxTagsPerType) + "）"
		}
		out = append(out, key)
	}
	return out, ""
}
