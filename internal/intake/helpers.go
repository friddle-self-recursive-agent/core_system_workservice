// helpers.go —— intake 辅助函数族（R-ENG-7 行数收口纯搬家：控件扁平化/自然语言抽取/
// 字段校验状态/回复组装/YAML 生成——零语义变化，intake.go 拆分伴生件）。
package intake

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/docrepo"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
	"gopkg.in/yaml.v3"
)

// ───────────────────────── 辅助函数：控件扁平化、自然语言抽取、校验与 YAML 生成 ─────────────────────────

func flattenWidgets(kindName string, ws []scheme.FieldDescriptor) []scheme.FieldDescriptor {
	var out []scheme.FieldDescriptor
	for _, w := range ws {
		if w.Widget == "object" && len(w.Fields) > 0 {
			if kindName == "Proposal" && w.Path == "landing" {
				// Proposal landing 以 landing.target 作为主控件（selector doc|task），ref 为可选附属
				for _, sf := range w.Fields {
					if sf.Name == "target" {
						out = append(out, sf)
					}
				}
				continue
			}
			out = append(out, flattenWidgets(kindName, w.Fields)...)
			continue
		}
		out = append(out, w)
	}
	return out
}

func findWidget(ws []scheme.FieldDescriptor, path string) *scheme.FieldDescriptor {
	for i := range ws {
		if ws[i].Path == path || ws[i].Name == path {
			return &ws[i]
		}
	}
	return nil
}

func normalizeFieldValue(w *scheme.FieldDescriptor, val any) any {
	if val == nil || w == nil {
		return val
	}
	switch w.Widget {
	case "list":
		switch v := val.(type) {
		case []string:
			return v
		case []any:
			var out []string
			for _, item := range v {
				s := strings.TrimSpace(fmt.Sprint(item))
				if s != "" {
					out = append(out, s)
				}
			}
			return out
		case string:
			return parseListText(v)
		}
	case "switch":
		switch v := val.(type) {
		case bool:
			return v
		case string:
			return strings.EqualFold(strings.TrimSpace(v), "true") || strings.TrimSpace(v) == "是" || strings.TrimSpace(v) == "1"
		}
	}
	return val
}

func isEmptyValue(v any) bool {
	if v == nil {
		return true
	}
	switch val := v.(type) {
	case string:
		return strings.TrimSpace(val) == ""
	case []string:
		return len(val) == 0
	case []any:
		return len(val) == 0
	case map[string]any:
		return len(val) == 0
	}
	return false
}

var (
	secretKeyValRe = regexp.MustCompile(`(?i)\b(api[_-]?key|secret[_-]?key|access[_-]?token|password|passwd|private[_-]?key)\s*[:=]\s*([^\s,;]+)`)
	skTokenRe      = regexp.MustCompile(`\bsk-[A-Za-z0-9_-]{10,}\b`)
	validRefPrefs  = []string{"secret:", "env:", "vault:", "keychain:"}
)

func hasValidSecretPrefix(v string) bool {
	v = strings.TrimSpace(v)
	for _, p := range validRefPrefs {
		if strings.HasPrefix(v, p) {
			return true
		}
	}
	return false
}

// checkSecretRefDiscipline 敏感项引用式检查（proposal.md §4.5 第 4 条 / provider.md 规则 2）：
// 禁止手填明文凭据，secret 只允许引用式（secret:/env:/vault:/keychain:）。
func checkSecretRefDiscipline(fieldPath, text string) string {
	lowerPath := strings.ToLower(fieldPath)
	if strings.Contains(lowerPath, "secret") {
		if !hasValidSecretPrefix(text) {
			return fmt.Sprintf("敏感字段 %s 禁止手填明文，必须使用引用式前缀（secret:/env:/vault:/keychain:）", fieldPath)
		}
	}
	if skTokenRe.MatchString(text) {
		return fmt.Sprintf("字段 %s 含疑似明文密钥（sk-...），敏感项禁止手填，只允许 secret:/env:/vault:/keychain: 引用式", fieldPath)
	}
	if m := secretKeyValRe.FindStringSubmatch(text); len(m) >= 3 {
		rhs := strings.Trim(m[2], `"'`)
		if !hasValidSecretPrefix(rhs) && !strings.HasPrefix(rhs, "<") {
			return fmt.Sprintf("字段 %s 含疑似明文凭据（%s=...），敏感项禁止手填，只允许 secret:/env:/vault:/keychain: 引用式", fieldPath, m[1])
		}
	}
	return ""
}

func checkRefExists(ctx context.Context, objects ObjectReader, refKind, scope, refVal string) bool {
	if objects == nil || refKind == "" {
		return true
	}
	targetName := refVal
	if idx := strings.LastIndex(refVal, "/"); idx >= 0 {
		targetName = refVal[idx+1:]
	}
	objs, err := objects.List(ctx, refKind, "")
	if err != nil || len(objs) == 0 {
		return true // 库为空或不可读时不误拦
	}
	for _, o := range objs {
		if o.Name == targetName || (o.Scope+"/"+o.Name) == refVal {
			return true
		}
	}
	return false
}

// extractFieldsFromText 从用户一段自由描述或单轮回答中抽取字段（支持一段话多字段抽取 + 当前询问字段回填）。
func extractFieldsFromText(kindName string, flatW []scheme.FieldDescriptor, msg string, askingField string, current map[string]any) map[string]any {
	out := make(map[string]any)
	text := strings.TrimSpace(msg)
	if text == "" {
		return out
	}

	if kindName == "Proposal" {
		// 1. 显式标签或自然语言分句抽取（标题/目标/验收标准/影响面/优先级/产物落点）
		labeled := extractProposalLabeled(text)
		for k, v := range labeled {
			out[k] = v
		}
		// 2. 若未用显式「标题：」标签，尝试自然语言多句启发式抽取
		if len(out) == 0 && (strings.Contains(text, "目标") || strings.Contains(text, "验收") || strings.Contains(text, "影响面") || strings.Contains(text, "；") || strings.Count(text, "，") >= 2) {
			heur := extractProposalHeuristic(text)
			for k, v := range heur {
				out[k] = v
			}
		}
	} else {
		// 通用 Kind：支持 "字段名: 值" 或 "标签: 值" 多行/分号抽取
		for _, seg := range splitClauses(text) {
			idx := strings.IndexAny(seg, ":：")
			if idx <= 0 {
				continue
			}
			key := strings.TrimSpace(seg[:idx])
			val := strings.TrimSpace(seg[idx+len(string([]rune(seg[idx:])[0])):])
			if val == "" {
				continue
			}
			for _, w := range flatW {
				if strings.EqualFold(key, w.Name) || strings.EqualFold(key, w.Path) || key == w.Label {
					out[w.Path] = normalizeFieldValue(&w, val)
				}
			}
		}
	}

	// 3. 若未命中多字段标签（或当前正针对 askingField 提问而该字段尚未被抽取），将输入赋给当前询问字段
	targetAsk := askingField
	if targetAsk == "" {
		// 首轮直接输入一段简短描述时，赋给首个待补字段
		for _, w := range flatW {
			if w.AutoFill == "" && isEmptyValue(current[w.Path]) {
				targetAsk = w.Path
				break
			}
		}
	}
	if len(out) == 0 && targetAsk != "" {
		w := findWidget(flatW, targetAsk)
		if w != nil {
			out[w.Path] = normalizeFieldValue(w, text)
		}
	}
	return out
}

type labelPattern struct {
	field string
	re    *regexp.Regexp
}

var proposalLabelPatterns = []labelPattern{
	{field: "title", re: regexp.MustCompile(`(?i)(?:^|[；;\n。])\s*(?:标题|提案标题|名称|title)\s*[:：=]\s*`)},
	{field: "summary", re: regexp.MustCompile(`(?i)(?:^|[；;\n。])\s*(?:目标|摘要|背景与目标|做什么|summary)\s*[:：=]\s*`)},
	{field: "acceptance", re: regexp.MustCompile(`(?i)(?:^|[；;\n。])\s*(?:验收标准|验收条目|验收|acceptance)\s*[:：=]\s*`)},
	{field: "impact", re: regexp.MustCompile(`(?i)(?:^|[；;\n。])\s*(?:影响面|影响范围|涉及系统|impact)\s*[:：=]\s*`)},
	{field: "priority", re: regexp.MustCompile(`(?i)(?:^|[；;\n。])\s*(?:优先级|priority)\s*[:：=]\s*`)},
	{field: "landing.target", re: regexp.MustCompile(`(?i)(?:^|[；;\n。])\s*(?:产物落点|落点|产物|landing)\s*[:：=]\s*`)},
}

func extractProposalLabeled(text string) map[string]any {
	type hit struct {
		field    string
		start    int
		valStart int
	}
	var hits []hit
	for _, lp := range proposalLabelPatterns {
		loc := lp.re.FindStringIndex(text)
		if loc != nil {
			hits = append(hits, hit{field: lp.field, start: loc[0], valStart: loc[1]})
		}
	}
	if len(hits) == 0 {
		return nil
	}
	sort.Slice(hits, func(i, j int) bool { return hits[i].start < hits[j].start })
	out := make(map[string]any)
	// 若首个标签前还有前置文本且未含 title 标签，将其作为 title
	if hits[0].start > 0 {
		prefix := strings.Trim(strings.TrimSpace(text[:hits[0].start]), "，,；;。 ")
		if prefix != "" {
			out["title"] = prefix
		}
	}
	for i, h := range hits {
		end := len(text)
		if i+1 < len(hits) {
			end = hits[i+1].start
		}
		rawVal := strings.Trim(strings.TrimSpace(text[h.valStart:end]), "，,；;。 ")
		if rawVal == "" {
			continue
		}
		switch h.field {
		case "acceptance":
			items := parseListText(rawVal)
			if len(items) > 0 {
				out["acceptance"] = items
			}
		case "priority":
			out["priority"] = normalizePriorityWord(rawVal)
		case "landing.target":
			out["landing.target"] = normalizeLandingWord(rawVal)
		default:
			out[h.field] = rawVal
		}
	}
	return out
}

func extractProposalHeuristic(text string) map[string]any {
	out := make(map[string]any)
	// 检查优先级词
	lower := strings.ToLower(text)
	switch {
	case strings.Contains(lower, "critical") || strings.Contains(text, "紧急"):
		out["priority"] = "critical"
	case strings.Contains(lower, "high") || strings.Contains(text, "高优") || strings.Contains(text, "高危"):
		out["priority"] = "high"
	case strings.Contains(lower, "normal") || strings.Contains(text, "中优") || strings.Contains(text, "普通优先级"):
		out["priority"] = "normal"
	case strings.Contains(lower, "low") || strings.Contains(text, "低优"):
		out["priority"] = "low"
	}
	// 检查落点词
	if strings.Contains(text, "任务文件") || strings.Contains(lower, "target=task") || strings.Contains(lower, "landing=task") {
		out["landing.target"] = "task"
	}
	// 按分号/句号/逗号切分寻找目标、验收、影响面与标题
	clauses := splitClauses(text)
	for i, c := range clauses {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		switch {
		case strings.HasPrefix(c, "目标是") || strings.HasPrefix(c, "为了") || strings.HasPrefix(c, "用于"):
			out["summary"] = strings.TrimSpace(strings.TrimPrefix(c, "目标是"))
		case strings.Contains(c, "验收标准") || strings.Contains(c, "验收是"):
			cleaned := regexp.MustCompile(`^.*?(?:验收标准是?|验收是?)\s*[:：]?\s*`).ReplaceAllString(c, "")
			if items := parseListText(cleaned); len(items) > 0 {
				out["acceptance"] = items
			}
		case strings.HasPrefix(c, "影响面是") || strings.HasPrefix(c, "影响面") || strings.HasPrefix(c, "涉及"):
			cleaned := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(c, "影响面是"), "影响面"))
			cleaned = strings.TrimLeft(cleaned, ":： ")
			if cleaned != "" {
				out["impact"] = cleaned
			}
		default:
			if i == 0 && isEmptyValue(out["title"]) && len([]rune(c)) <= 50 {
				out["title"] = c
			} else if isEmptyValue(out["summary"]) && len([]rune(c)) > 8 {
				out["summary"] = c
			}
		}
	}
	return out
}

func normalizePriorityWord(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	switch s {
	case "低", "低优", "低危", "low":
		return "low"
	case "中", "中优", "普通", "normal", "medium":
		return "normal"
	case "高", "高优", "高危", "high":
		return "high"
	case "紧急", "最高", "critical":
		return "critical"
	default:
		return s // 保留原值以便步骤 3 冲突检测识别非法枚举！
	}
}

func normalizeLandingWord(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	switch {
	case s == "doc" || strings.Contains(s, "文档") || strings.Contains(s, "docsroot"):
		return "doc"
	case s == "task" || strings.Contains(s, "任务"):
		return "task"
	default:
		return s
	}
}

var listSplitRe = regexp.MustCompile(`(?:\r?\n|[；;]|(?:^|\s+)\d+[.、)）]\s*)`)

func parseListText(raw string) []string {
	parts := listSplitRe.Split(strings.TrimSpace(raw), -1)
	var out []string
	for _, p := range parts {
		p = strings.Trim(strings.TrimSpace(p), "-*•，,。；; ")
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 && strings.TrimSpace(raw) != "" {
		out = append(out, strings.TrimSpace(raw))
	}
	return out
}

func splitClauses(s string) []string {
	return regexp.MustCompile(`[\r\n；;。]+`).Split(s, -1)
}

func evaluateFields(ctx context.Context, kindName string, flatW []scheme.FieldDescriptor, fields map[string]any, sources map[string]string, objects ObjectReader) ([]FieldState, []string, []string) {
	states := make([]FieldState, 0, len(flatW))
	missing := []string{}
	conflicts := []string{}

	for _, w := range flatW {
		val := fields[w.Path]
		src := sources[w.Path]
		fs := FieldState{
			Path:     w.Path,
			Name:     w.Name,
			Label:    w.Label,
			Widget:   w.Widget,
			Type:     w.Type,
			Required: w.Required,
			Enum:     w.Enum,
			Default:  w.Default,
			Example:  w.Example,
			Value:    val,
			Source:   src,
		}

		// Proposal impact 虽在 scheme.yaml 为 ?string，但在提案对话录入六要素（proposal.md §4.5 / console-ui.md §12）
		// 中属待询问项（提问时提供默认值，跳过即填默认）。
		needAsk := w.Required || (kindName == "Proposal" && w.Path == "impact")
		if w.AutoFill == "system" {
			needAsk = false
		}

		if isEmptyValue(val) {
			if needAsk {
				fs.Status = "missing"
				missing = append(missing, w.Path)
			} else {
				fs.Status = "default"
			}
			states = append(states, fs)
			continue
		}

		// 校验枚举闭合集（R-UX-2）
		if w.Widget == "selector" && len(w.Enum) > 0 && !w.HasWildcard {
			sval := strings.TrimSpace(fmt.Sprint(val))
			hit := false
			for _, opt := range w.Enum {
				if opt == sval {
					hit = true
					break
				}
			}
			if !hit {
				fs.Status = "conflict"
				fs.Error = fmt.Sprintf("值 %q 不在合法枚举 [%s] 中", sval, strings.Join(w.Enum, "|"))
				conflicts = append(conflicts, w.Path)
				states = append(states, fs)
				continue
			}
		}

		// 校验 R-PUB-1 lint 与敏感项引用式纪律
		txt := valueToText(val)
		if vs := docrepo.Lint([]byte(txt)); len(vs) > 0 {
			fs.Status = "conflict"
			fs.Error = fmt.Sprintf("R-PUB-1 (%s) 违规：含 %q", vs[0].Rule, vs[0].Text)
			conflicts = append(conflicts, w.Path)
			states = append(states, fs)
			continue
		}
		if secMsg := checkSecretRefDiscipline(w.Path, txt); secMsg != "" {
			fs.Status = "conflict"
			fs.Error = secMsg
			conflicts = append(conflicts, w.Path)
			states = append(states, fs)
			continue
		}

		// 校验 ref 引用存在性（R-UX-2）
		if w.Widget == "ref" && w.RefKind != "" && objects != nil {
			sval := strings.TrimSpace(fmt.Sprint(val))
			if sval != "" && !checkRefExists(ctx, objects, w.RefKind, "global", sval) {
				fs.Status = "conflict"
				fs.Error = fmt.Sprintf("引用对象 %q (%s) 不在册", sval, w.RefKind)
				conflicts = append(conflicts, w.Path)
				states = append(states, fs)
				continue
			}
		}

		if src == "default" || src == "subject" || src == "system" {
			fs.Status = "default"
		} else {
			fs.Status = "filled"
		}
		states = append(states, fs)
	}
	return states, missing, conflicts
}

func composeReply(kindName string, flatW []scheme.FieldDescriptor, sess *SessionRecord, autoNotes []string, extracted map[string]any, states []FieldState, missing, conflicts []string, nextAsk string) string {
	var parts []string
	if sess.Turns == 1 && len(autoNotes) > 0 {
		parts = append(parts, fmt.Sprintf("已结合当前上下文自动补充缺省项：%s（右侧字段卡可随时修改，不会重复盘问已推断项）。", strings.Join(autoNotes, "、")))
	}
	if len(extracted) > 0 {
		var keys []string
		for k := range extracted {
			w := findWidget(flatW, k)
			if w != nil {
				keys = append(keys, w.Label)
			} else {
				keys = append(keys, k)
			}
		}
		sort.Strings(keys)
		parts = append(parts, fmt.Sprintf("已抽取并同步至字段卡：%s。", strings.Join(keys, "、")))
	}

	if len(conflicts) > 0 {
		w := findWidget(flatW, nextAsk)
		errText := ""
		for _, st := range states {
			if st.Path == nextAsk {
				errText = st.Error
				break
			}
		}
		label := nextAsk
		if w != nil {
			label = w.Label
		}
		hint := ""
		if w != nil && len(w.Enum) > 0 {
			hint = fmt.Sprintf("合法枚举值：[%s]（默认：%v，示例：%s）", strings.Join(w.Enum, " | "), defaultOrFirst(w), w.Example)
		} else if w != nil {
			hint = fmt.Sprintf("（示例：%s）", w.Example)
		}
		parts = append(parts, fmt.Sprintf("字段【%s (%s)】存在冲突或准入违规：%s。请修正该项 %s：", label, nextAsk, errText, hint))
		return strings.Join(parts, "\n")
	}

	if len(missing) == 0 {
		parts = append(parts, "全部 scheme 字段已齐备（字段卡全绿）！请点击【生成提案预览】，由提交节点生成标准 YAML 对象并执行 R-PUB lint 与 spec 准入 dry-run。")
		return strings.Join(parts, "\n")
	}

	w := findWidget(flatW, nextAsk)
	if w == nil {
		parts = append(parts, fmt.Sprintf("请补充字段 %s：", nextAsk))
		return strings.Join(parts, "\n")
	}
	defStr := "无（必填）"
	if w.Default != nil && fmt.Sprint(w.Default) != "" {
		defStr = fmt.Sprintf("%v（可点「跳过」采用默认）", w.Default)
	} else if w.Example != "" {
		defStr = fmt.Sprintf("可点「跳过」采用示例值「%s」", w.Example)
	}
	parts = append(parts, fmt.Sprintf("请提供【%s (%s)】（剩余待补 %d 项；默认：%s；示例：%s；也可一段话自由描述多项）：",
		w.Label, w.Path, len(missing), defStr, w.Example))
	return strings.Join(parts, "\n")
}

func defaultOrFirst(w *scheme.FieldDescriptor) any {
	if w.Default != nil {
		return w.Default
	}
	if len(w.Enum) > 0 {
		return w.Enum[0]
	}
	return ""
}

func valueToText(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case []string:
		return strings.Join(val, "\n")
	case []any:
		var s []string
		for _, it := range val {
			s = append(s, fmt.Sprint(it))
		}
		return strings.Join(s, "\n")
	default:
		b, _ := yaml.Marshal(val)
		return string(b)
	}
}

func deriveSlug(title, sessionID string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(title) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else if b.Len() > 0 && !strings.HasSuffix(b.String(), "-") {
			b.WriteByte('-')
		}
	}
	s := strings.Trim(b.String(), "-")
	if len(s) >= 3 {
		if len(s) > 36 {
			s = strings.Trim(s[:36], "-")
		}
		return "prop-" + s
	}
	suffix := sessionID
	if idx := strings.LastIndex(sessionID, "-"); idx >= 0 {
		suffix = sessionID[idx+1:]
	}
	return fmt.Sprintf("prop-%d-%s", time.Now().UTC().Unix()%100000, suffix)
}

func buildObjectYAML(k scheme.Kind, scope, name string, fields map[string]any) string {
	if scope == "" {
		scope = "global"
	}
	if name == "" {
		name = "unnamed"
	}
	if k.Name == "Proposal" {
		title, _ := fields["title"].(string)
		summary, _ := fields["summary"].(string)
		impact, _ := fields["impact"].(string)
		priority, _ := fields["priority"].(string)
		if priority == "" {
			priority = "low"
		}
		proposer, _ := fields["proposer"].(string)
		if proposer == "" {
			proposer = "console-operator"
		}
		status, _ := fields["status"].(string)
		if status == "" {
			status = "submitted"
		}
		target, _ := fields["landing.target"].(string)
		if target == "" {
			target = "doc"
		}
		landingRef, _ := fields["landing.ref"].(string)

		var acc []string
		switch v := fields["acceptance"].(type) {
		case []string:
			acc = v
		case []any:
			for _, it := range v {
				if s := strings.TrimSpace(fmt.Sprint(it)); s != "" {
					acc = append(acc, s)
				}
			}
		case string:
			acc = parseListText(v)
		}

		var sb strings.Builder
		sb.WriteString("apiVersion: core/v1\n")
		sb.WriteString("kind: Proposal\n")
		sb.WriteString(fmt.Sprintf("metadata: {name: %s, scope: %s}\n", yamlScalar(name), yamlScalar(scope)))
		sb.WriteString("spec:\n")
		if strings.TrimSpace(title) != "" {
			sb.WriteString(fmt.Sprintf("  title: %s\n", yamlScalar(title)))
		}
		if strings.TrimSpace(summary) != "" {
			sb.WriteString(fmt.Sprintf("  summary: %s\n", yamlScalar(summary)))
		}
		if len(acc) > 0 {
			sb.WriteString("  acceptance:\n")
			for _, item := range acc {
				sb.WriteString(fmt.Sprintf("    - %s\n", yamlScalar(item)))
			}
		}
		if strings.TrimSpace(impact) != "" {
			sb.WriteString(fmt.Sprintf("  impact: %s\n", yamlScalar(impact)))
		}
		sb.WriteString(fmt.Sprintf("  priority: %s\n", yamlScalar(priority)))
		sb.WriteString(fmt.Sprintf("  proposer: %s\n", yamlScalar(proposer)))
		sb.WriteString(fmt.Sprintf("  status: %s\n", yamlScalar(status)))
		if strings.TrimSpace(landingRef) != "" {
			sb.WriteString(fmt.Sprintf("  landing: {target: %s, ref: %s}\n", yamlScalar(target), yamlScalar(landingRef)))
		} else {
			sb.WriteString(fmt.Sprintf("  landing: {target: %s}\n", yamlScalar(target)))
		}
		// 动态附加字段（当用户修改 Scheme 为 Proposal 新增字段时，自动序列化入 YAML！）
		known := map[string]bool{
			"title": true, "summary": true, "acceptance": true, "impact": true,
			"priority": true, "proposer": true, "status": true,
			"landing": true, "landing.target": true, "landing.ref": true,
		}
		var extras []string
		for fk := range fields {
			if !known[fk] && !isEmptyValue(fields[fk]) {
				extras = append(extras, fk)
			}
		}
		sort.Strings(extras)
		for _, fk := range extras {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", fk, yamlScalar(fmt.Sprint(fields[fk]))))
		}
		return sb.String()
	}

	// 通用 Kind YAML 生成
	specMap := make(map[string]any)
	for fk, fv := range fields {
		if !isEmptyValue(fv) {
			specMap[fk] = fv
		}
	}
	doc := map[string]any{
		"apiVersion": "core/v1",
		"kind":       k.Name,
		"metadata": map[string]any{
			"name":  name,
			"scope": scope,
		},
		"spec": specMap,
	}
	out, err := yaml.Marshal(doc)
	if err != nil {
		return fmt.Sprintf("apiVersion: core/v1\nkind: %s\nmetadata: {name: %s, scope: %s}\nspec: {}\n", k.Name, name, scope)
	}
	return string(out)
}

func yamlScalar(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return `""`
	}
	if strings.ContainsAny(s, ":#-?&*!>|%@`'\"{[},\n\r\t ") {
		escaped := strings.ReplaceAll(s, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, `"`, `\"`)
		escaped = strings.ReplaceAll(escaped, "\n", `\n`)
		return `"` + escaped + `"`
	}
	return s
}

func appendFieldErrorUnique(list []FieldError, fe FieldError) []FieldError {
	for _, existing := range list {
		if existing.Field == fe.Field && existing.Rule == fe.Rule {
			return list
		}
	}
	return append(list, fe)
}

func cloneMap(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func cloneStrMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
