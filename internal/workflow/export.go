// export.go —— workflow 投影导出面（G-08 收口——事件详情↔最近 loop↔session 链的可查询
// 落点，R-UX-3/4 可观测链）。路由面（apiserver）只消费本文件导出的 accessor，零内剖包内部：
//   - LoopSession / LoopWake     最近 loop（LoopRequested）会话关联投影——GET /v1/loop/current
//     的 session 锚数据源（唤醒桥受理回执回填后非空；无回执=null 不冒充）；
//   - AttachSessionID            唤醒桥受理回执回填入口（run 与 wake 同 eventID 关联——
//     dsh 会话 id/宿主进程锚由装配层桥回执获得，引擎侧只承接落环+写穿透）。
//
// 会话锚语义（诚实投影）：SessionID=装配侧桥回执提供的锚（宿主 sessionId 优先；缺席=宿主
// 进程句柄 pid:<n> 词形——受理已发生∧拉起已受理的最小非空锚；空串=未回填不冒充已拉起）。
package workflow

// SessionID 字段语义随 RunRecord/WakeRecord（run.go/wake.go）——本文件只放关联行为与
// 读面 accessor，字段定义不搬家（投影结构随各自投影面文件）。

// AttachSessionID 回填 run 会话锚（G-08 loop 会话关联）：按 eventID 命中环内最新 run，
// 就地覆写 SessionID 并写穿透（run 是可演化投影——AppendRun upsert 同 run id 覆写）。
// 命中返回 true；未命中（事件未落 run/环已挤出）=false（缺口可见，调用方 stderr 注记）。
func (r *RunLog) AttachSessionID(eventID, sessionID string) bool {
	if eventID == "" || sessionID == "" {
		return false
	}
	r.mu.Lock()
	for i := len(r.ring) - 1; i >= 0; i-- {
		if r.ring[i].EventID == eventID {
			r.ring[i].SessionID = sessionID
			row := runToRow(r.ring[i])
			r.mu.Unlock()
			if r.opts.Store != nil {
				persistAppend("run 会话回填", r.opts.Store.AppendRun(row))
			}
			return true
		}
	}
	r.mu.Unlock()
	return false
}

// AttachSessionID 回填唤醒会话锚（G-08）：按 eventID 命中环内最新唤醒投影，就地覆写
// SessionID 并写穿透（AppendWakeSession 演化回填——受理词形 IGNORE 幂等不覆盖回填）。
// 命中返回 true；未命中=false。
func (w *WakeLog) AttachSessionID(eventID, sessionID string) bool {
	if eventID == "" || sessionID == "" {
		return false
	}
	w.mu.Lock()
	for i := len(w.ring) - 1; i >= 0; i-- {
		if w.ring[i].EventID == eventID {
			w.ring[i].SessionID = sessionID
			wake, sess := w.ring[i].Wake, sessionID
			w.mu.Unlock()
			if w.opts.Store != nil {
				persistAppend("唤醒会话回填", w.opts.Store.AppendWakeSession(wake, sess))
			}
			return true
		}
	}
	w.mu.Unlock()
	return false
}

// LoopType loop 会话关联的事件类型锚（自循环拉起=LoopRequested——scheme.dispatch root-agent
// 行触发源；/v1/loop/current 读面判别词，routes 侧同字面消费零双源）。
const LoopType = "LoopRequested"

// LoopSession 最近 loop run 的会话关联投影（/v1/loop/current 读面 accessor）：环内尾向
// 扫描最近一条 LoopRequested run。ok=false=环内无 loop run（无数据→face 侧 null 不编造）。
// SessionID 空=loop run 在册但会话锚未回填（唤醒桥缺席/拉起未受理——face 侧 null 同词形）。
func (r *RunLog) LoopSession() (RunRecord, bool) {
	st := r.State()
	for i := len(st.Runs) - 1; i >= 0; i-- {
		if st.Runs[i].Type == LoopType {
			return st.Runs[i], true
		}
	}
	return RunRecord{}, false
}

// LoopWake 最近 loop 唤醒投影（/v1/loop/current 读面 accessor——wake 行半边）：环内尾向
// 扫描最近一条 LoopRequested 唤醒。ok=false=环内无 loop 唤醒（face 侧字段 null 不冒充）。
func (w *WakeLog) LoopWake() (WakeRecord, bool) {
	st := w.State()
	for i := len(st.Wakes) - 1; i >= 0; i-- {
		if st.Wakes[i].Type == LoopType {
			return st.Wakes[i], true
		}
	}
	return WakeRecord{}, false
}
