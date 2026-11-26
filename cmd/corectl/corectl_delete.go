package main

// corectl_delete.go —— `delete <kind>/<scope>/<name>` 对象删除（2026-11-26 假数据清理批）。
// 通路=DELETE /v1/objects/{kind}/{name}?scope=&expectRV=（与 apply 同一写面闸门：
// --allow-apply + write×scope 权限）；删除落三账（doc_history 终态行/apply_log deleted/
// DocDeleted 事件）。破坏性动作：默认要求 --yes 显式确认；--rv N=乐观锁期望（不符即拒）。
import (
	"flag"
	"fmt"
	"net/url"
	"strings"
)

func (a *app) cmdDelete(args []string) int {
	// 旗标先行归一（Go flag 包遇首个位置参数即停——`delete k/s/n --yes` 词形把 --yes 落入
	// 位置参数；把带 -- 前缀的词挪到前面即可两种词形都吃）。
	var flags, pos []string
	for _, x := range args {
		if strings.HasPrefix(x, "--") || strings.HasPrefix(x, "-") {
			flags = append(flags, x)
		} else {
			pos = append(pos, x)
		}
	}
	fs := flag.NewFlagSet("delete", flag.ContinueOnError)
	yes := fs.Bool("yes", false, "显式确认（破坏性动作缺省拒执行）")
	rv := fs.Int64("rv", 0, "乐观锁期望 resourceVersion（0=不设防）")
	fm := fs.String("field-manager", "corectl:delete", "fieldManager 留账词形")
	if err := fs.Parse(flags); err != nil {
		return 2
	}
	rest := pos
	if len(rest) != 1 || !strings.Contains(rest[0], "/") {
		fmt.Fprintln(a.err, "用法：corectl delete <kind>/<scope>/<name> [--rv N] --yes")
		return 2
	}
	parts := strings.SplitN(rest[0], "/", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		fmt.Fprintln(a.err, "用法：corectl delete <kind>/<scope>/<name>（三段全必填——对象键=(kind,scope,name)）")
		return 2
	}
	if !*yes {
		fmt.Fprintln(a.err, "拒绝执行：删除是账面动作（doc_history 留终态行）且不可在本面撤销——加 --yes 显式确认。")
		return 2
	}
	q := url.Values{}
	q.Set("scope", parts[1])
	if *rv > 0 {
		q.Set("expectRV", fmt.Sprintf("%d", *rv))
	}
	q.Set("fieldManager", *fm)
	q.Set("reasonRef", "corectl delete")
	raw, err := a.request("DELETE", "/v1/objects/"+url.PathEscape(parts[0])+"/"+url.PathEscape(parts[2])+"?"+q.Encode(), nil)
	if err != nil {
		return fail(a, err)
	}
	return a.emit(raw)
}
