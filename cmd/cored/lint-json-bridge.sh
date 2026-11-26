#!/bin/sh
# lint-json-bridge.sh —— 净入 lint 脚本 → connectops-code-audit 命令模板 stdout 契约桥
# （G-09/G-12 Verify 硬 gate 装配伴生件——cmd/cored 同目录部署）。
#
# codeaudit 命令模板契约：stdout 空=零发现；非空={"findings":[{file,line,severity,message}]}
# JSON；非零退出/stderr=显式失败（error finding）。净入 lint 脚本（core_design/lint/*.sh）
# 词形=文本行 + exit 0/1——本桥不改脚本，适配在装配侧：
#   exit 0            → {"findings":[]}（脚本自身可能打了通过行——吞掉，按零发现投影）
#   exit 非 0         → 逐行转 error 级 findings（line=0 message=原行），exit 0 输出——
#                       失败语义已编码进 findings（severity=error→审核不过→run 卡 Verify），
#                       不靠退出码重复表达（模板契约下退出码只承载「桥自身故障」）。
# 用法：lint-json-bridge.sh <script.sh> <target>
# 越界=显式失败：参数缺失/脚本不可执行 → stderr 注记 + exit 1（连接器落 error finding）。
set -u

if [ "$#" -ne 2 ]; then
  echo "usage: lint-json-bridge.sh <script> <target>" >&2
  exit 1
fi
script="$1"
target="$2"
if [ ! -f "$script" ]; then
  echo "lint 脚本不可达: $script" >&2
  exit 1
fi

out="$(sh "$script" "$target" 2>/dev/null)"
rc=$?

if [ "$rc" -eq 0 ]; then
  echo '{"findings":[]}'
  exit 0
fi

# 非 0=有发现：逐行转 findings JSON（行文本截 300 字符防载荷膨胀；空行跳过）。
printf '{"findings":['
first=1
printf '%s\n' "$out" | while IFS= read -r line; do
  case "$line" in
    "") continue ;;
  esac
  msg=$(printf '%s' "$line" | cut -c1-300 | sed 's/\\/\\\\/g; s/"/\\"/g')
  if [ "$first" -eq 1 ]; then
    first=0
  else
    printf ','
  fi
  printf '{"file":"%s","line":0,"severity":"error","message":"%s"}' "$target" "$msg"
done
printf ']}'
echo ""
exit 0
