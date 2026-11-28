#!/bin/bash
# core-watchdog —— core_agent 看门狗（独立程序；2026-11-28 用户裁定：发现 core_agent[systemctl]
# 不在位 → 拉最新 core-workflow[构建换装] 并让服务起来；现阶段 1 小时一拍 systemd timer）。
# 两段职责：
#   ① 服务不在位（is-active 非 active）→ 六仓 pull --ff-only → 机上构建 cored → md5 不同才
#      换装（.bak 留档）→ 顺序 restart 缺席服务；构建失败=只重启现有 bin（诚实降级）。
#   ② 自循环活性：最新会话账 mtime <40min=活；超 40min 且无 headless 进程且无 .loop-hold
#      （用户暂停闸，巡检误清教训）→ corectl loop 点火（只走 wake 桥正道，loop-96 交接契约
#      「只经 corectl loop；不得 setsid 自点火」）。
# 日志：~/.local/state/core-watchdog.log（一拍一行起，落账可查）。
set -u
LOG=/home/friddle/.local/state/core-watchdog.log
ROOT=/home/friddle/Project/self-recursive-agent
CTL=/home/friddle/bin/corectl
GO=/usr/local/go/bin/go
BIN=/home/friddle/bin/cored
SESS=/home/friddle/.dsh/sessions/--home-friddle--
SERVICES="dsh-core-system dsh-kube-inner"
mkdir -p "$(dirname "$LOG")"
log() { echo "[$(date '+%Y-%m-%dT%H:%M:%S%z')] $*" >> "$LOG"; }

# ① 服务在位检查
DOWN=""
for s in $SERVICES; do
  systemctl is-active --quiet "$s" || DOWN="$DOWN $s"
done

if [ -n "$DOWN" ]; then
  log "DOWN:${DOWN} —— 拉最新+构建+按需换装+重启"
  for r in core_design core_system_workservice core_system_dsh_plugin core_system_instance; do
    git -C "$ROOT/$r" pull --ff-only -q >>"$LOG" 2>&1 || log "pull $r 失败（保留现版继续）"
  done
  if (cd "$ROOT/core_system_workservice" && "$GO" build -o /tmp/cored-wd-new ./cmd/cored) >>"$LOG" 2>&1; then
    old=$(md5sum "$BIN" 2>/dev/null | cut -d' ' -f1)
    new=$(md5sum /tmp/cored-wd-new 2>/dev/null | cut -d' ' -f1)
    if [ -n "$old" ] && [ "$old" != "$new" ]; then
      cp -a "$BIN" "$BIN.bak-wd-$(date +%s)" \
        && mv /tmp/cored-wd-new "$BIN.new" && chown friddle:friddle "$BIN.new" && mv "$BIN.new" "$BIN" \
        && log "bin 换装 md5 $old -> $new" \
        || log "bin 换装失败（保留现版）"
    else
      log "bin 无变化（md5 同），不换装"
      rm -f /tmp/cored-wd-new
    fi
  else
    log "构建失败 —— 只重启现有 bin"
  fi
  for s in $DOWN; do
    sudo systemctl restart "$s" && log "restart $s ok" || log "restart $s 失败"
  done
  sleep 5 # 服务起来再查循环
fi

# ② 自循环活性（账目真相源=会话账 mtime；投影只作参考——2026-11-28 盲区教训）
recent=$(find "$SESS" -name 'session.v4.jsonl.zstd' -newermt '-40 minutes' 2>/dev/null | head -1)
if [ -z "$recent" ] && ! pgrep -f 'profile core-headless' >/dev/null 2>&1 \
   && [ ! -f /home/friddle/.loop-hold ]; then
  if HOME=/home/friddle "$CTL" loop --dedup "wd-$(date +%s)" >>"$LOG" 2>&1; then
    log "自循环停滞（>40min 无账目推进且无 headless）→ LoopRequested 点火已发"
  else
    log "LoopRequested 点火失败（wake 桥缺席？服务刚重启未就绪，下拍重试）"
  fi
fi
log "拍完成 DOWN=[${DOWN:-none}] recent_session=[${recent:+alive}]"
