// stream.go —— GET /v1/snapshot（聚合快照）+ GET /v1/stream（WebSocket 推送面，v0.66）。
// 设计依据：R-ENG 任务书 §2（coder/websocket；本面只读）。协议：
//
//   - 连接即推 {type:"snapshot", data:<同 /v1/snapshot>, seq:N}；
//   - 此后底层状态变更（总线订阅——全类型一条订阅覆盖 DocApplied/TaskStateChanged/
//     ApprovalNeeded/Escalate/SessionStalled 等已注册事件族 + 定时 5s 兜底重算）推
//     {type:"patch", patch:[<RFC6902 add/replace/remove 数组，diff.go 保守器产出>], seq:N}；
//     patch=前后两次快照 diff（generatedAt 不参与 diff——重算时刻抖动不产生 patch 噪音）；
//   - 心跳：30s 服务侧 ping；读循环持续在位（控制帧处理+关闭感知），断连即清理；
//   - seq 单调递增：patch 只在其基线快照之上应用（快照帧携带当前 seq——消费方丢弃
//     seq≤快照 seq 的迟到 patch）。
//
// 鉴权：继承 apiserver 现状（本面只读；--allow-apply 关断时照常可连，载荷与 GET 面同
// 口径——快照构造器即 GET /v1/snapshot 同一函数）。face 缺席=载荷字段 null（不冒充），
// 端点恒注册不 503（快照是聚合投影不是能力面）。
package apiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
)

// stream 面参数（任务书口径：5s 兜底重算、30s 心跳；写帧超时=短时限防慢消费方拖住泵；
// 去抖=事件拍后短窗再读现势——订阅方落投影与总线派发并发的合并窗）。
const (
	streamRescanInterval = 5 * time.Second
	streamPingInterval   = 30 * time.Second
	streamWriteTimeout   = 5 * time.Second
	streamDebounceDelay  = 50 * time.Millisecond
)

// WithEventBus 注入事件总线（stream 面变更感知源；nil=仅 5s 兜底重算——headless/测试
// 装配不受影响）。
func WithEventBus(b *eventbus.Bus) Option {
	return func(o *options) { o.bus = b }
}

// streamHub 推送面泵：持基线快照与连接集；事件/兜底拍驱动重算 diff 与广播。
// 惰性启动（首个 WS 连接触发）；零连接期不重算（重算只在有消费方时有意义）。
type streamHub struct {
	d   *deps
	bus *eventbus.Bus

	mu      sync.Mutex
	conns   map[*streamConn]chan struct{}
	last    map[string]any // 基线快照（不含 generatedAt 参与项；与客户端持有态同基线）
	seq     int64
	started bool
	notify  chan struct{}
}

// streamConn 单连接（写互斥——coder/websocket 同连接写非并发安全；读循环+心跳在 serve 内管理）。
type streamConn struct {
	ws  *websocket.Conn
	wmu sync.Mutex
}

// registerStreamRoutes stream 面路由注册（snapshot=一次 GET；stream=WS 升级）。
func registerStreamRoutes(d *deps, mux *http.ServeMux) {
	hub := &streamHub{
		d:      d,
		bus:    d.o.bus,
		conns:  map[*streamConn]chan struct{}{},
		notify: make(chan struct{}, 1),
	}
	mux.HandleFunc("GET /v1/snapshot", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, snapshotPayload(d))
	})
	mux.HandleFunc("GET /v1/stream", hub.serve)
}

// serve WS 升级 + 基线快照下发 + 读循环/心跳（连接生命期管理）。
func (h *streamHub) serve(w http.ResponseWriter, r *http.Request) {
	ws, err := websocket.Accept(w, r, nil)
	if err != nil {
		return // 升级失败：Accept 已按 HTTP 语义回执（非 WS 握手=400）
	}
	sc := &streamConn{ws: ws}
	stop := make(chan struct{})
	h.addConn(sc, stop)
	defer h.removeConn(sc)
	// 连接即推基线快照（seq 随帧——消费方据此丢弃迟到 patch）。
	seq, snap := h.joinSnapshot()
	if err := sc.writeFrame(map[string]any{"type": "snapshot", "data": snap, "seq": seq}); err != nil {
		return // 基线下发失败=连接不可用（读循环侧亦将退出）
	}
	go sc.pingLoop(stop)
	// 读循环：客户端消息照收即弃（只读面协议无请求词）；错误/关闭=断连（defer 清理——
	// removeConn 统一关 stop，pingLoop 随之退场）。
	for {
		if _, _, err := ws.Read(r.Context()); err != nil {
			break
		}
	}
}

// addConn 登记连接（首连启动泵与总线订阅——惰性装配，无连接零后台开销）。
func (h *streamHub) addConn(sc *streamConn, stop chan struct{}) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.conns[sc] = stop
	if !h.started {
		h.started = true
		go h.pump()
	}
}

// removeConn 断连清理（幂等；连接缺席=空操作）。
func (h *streamHub) removeConn(sc *streamConn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if stop, ok := h.conns[sc]; ok {
		close(stop)
		delete(h.conns, sc)
	}
}

// joinSnapshot 取基线快照（基线缺席=现建并登记——与后续 patch 同基线零错位）。
// 返回 (当前 seq, 快照)。
func (h *streamHub) joinSnapshot() (int64, map[string]any) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.last == nil {
		h.last = stripGeneratedAt(snapshotPayload(h.d))
	}
	return h.seq, h.last
}

// pump 泵循环：事件信号与 5s 兜底拍驱动重算（信号合并——连发事件一次重算）。
func (h *streamHub) pump() {
	if h.bus != nil {
		// 全类型订阅一条（eventbus typ=""——覆盖已注册事件族全集：DocApplied/
		// TaskStateChanged/ApprovalNeeded/Escalate/SessionStalled 等；退场随进程生命周期——
		// 泵与面同寿命，无独立 stop 面）。
		ch, _ := h.bus.Subscribe("")
		go func() {
			for range ch {
				select {
				case h.notify <- struct{}{}:
				default: // 信号已在队=合并（重算读现势态，连发事件一次重算足矣）
				}
			}
		}()
	}
	ticker := time.NewTicker(streamRescanInterval)
	defer ticker.Stop()
	for {
		select {
		case <-h.notify:
			// 变更感知合并窗（短去抖）：总线发布与订阅方落投影（留痕/run/送达…）是并发
			// goroutine——立刻重算会与落投影竞速（零差异假阴）；短窗后读现势态一次到位。
			// 与 5s 兜底拍同构的「拍后读」语义，不改任何投影面行为。
			time.Sleep(streamDebounceDelay)
		case <-ticker.C: // 兜底重算（事件丢失/无总线装配的双保险）
		}
		h.recompute()
	}
}

// recompute 重算快照 → diff 基线 → 有差异即广播 patch（seq 进位；零连接=零重算）。
func (h *streamHub) recompute() {
	h.mu.Lock()
	if len(h.conns) == 0 {
		h.mu.Unlock()
		return // 无消费方不重算（基线保持——新连接 join 时现建，patch 错位不可能）
	}
	prev := h.last
	fresh := stripGeneratedAt(snapshotPayload(h.d))
	ops := DiffJSON(prev, fresh)
	if len(ops) > 0 {
		h.seq++
	}
	h.last = fresh
	frame := map[string]any{"type": "patch", "patch": ops, "seq": h.seq}
	targets := make([]*streamConn, 0, len(h.conns))
	for sc := range h.conns {
		targets = append(targets, sc)
	}
	h.mu.Unlock()
	if len(ops) == 0 {
		return // 零差异零帧（兜底拍不打扰）
	}
	b, err := json.Marshal(frame)
	if err != nil {
		return // 构造面载荷恒可 marshal（防御性——失败=本轮丢弃不崩泵）
	}
	for _, sc := range targets {
		sc.writeRaw(b) // 写失败=该连接读循环侧感知退出（广播不逐连接重试）
	}
}

// writeFrame marshal+写一帧（互斥写）。
func (c *streamConn) writeFrame(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	c.writeRaw(b)
	return nil
}

// writeRaw 原始帧写（互斥+短超时；慢消费方不拖住泵——写错误由读循环侧感知清场）。
func (c *streamConn) writeRaw(b []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), streamWriteTimeout)
	defer cancel()
	c.wmu.Lock()
	defer c.wmu.Unlock()
	_ = c.ws.Write(ctx, websocket.MessageText, b)
}

// pingLoop 30s 服务侧心跳（ping 失败=连接坏——读循环侧感知退出；stop=断连信号）。
func (c *streamConn) pingLoop(stop <-chan struct{}) {
	ticker := time.NewTicker(streamPingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), streamWriteTimeout)
			c.wmu.Lock()
			err := c.ws.Ping(ctx)
			c.wmu.Unlock()
			cancel()
			if err != nil {
				return
			}
		}
	}
}
