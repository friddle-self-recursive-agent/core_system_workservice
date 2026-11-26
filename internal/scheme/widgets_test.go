package scheme

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// coreSchemePath CORE_SYSTEM_SCHEME 环境变量优先（与 cored 启动旗标同源）；缺省=仓库相对
// 布局（workservice 与 core_design 同级 clone——Mac 侧布局）。
func coreSchemePath() string {
	if v := strings.TrimSpace(os.Getenv("CORE_SYSTEM_SCHEME")); v != "" {
		return v
	}
	return filepath.Join("..", "..", "..", "core_design", "core", "scheme.yaml")
}

func TestDescribeKindWidgets_RUX2AllEightWidgetTypes(t *testing.T) {
	schPath := coreSchemePath()
	sch, err := LoadFile(schPath)
	if err != nil {
		t.Fatalf("LoadFile(%s): %v", schPath, err)
	}

	prop, err := sch.Lookup("Proposal")
	if err != nil {
		t.Fatalf("Lookup(Proposal): %v", err)
	}
	byName := map[string]FieldDescriptor{}
	for _, w := range prop.Widgets {
		byName[w.Name] = w
	}
	if w := byName["title"]; w.Widget != "input" || !w.Required {
		t.Fatalf("Proposal.title widget=%s required=%v", w.Widget, w.Required)
	}
	if w := byName["summary"]; w.Widget != "textarea" || !w.Required {
		t.Fatalf("Proposal.summary widget=%s required=%v", w.Widget, w.Required)
	}
	if w := byName["acceptance"]; w.Widget != "list" || w.Elem == nil || w.Elem.Widget != "input" {
		t.Fatalf("Proposal.acceptance widget=%+v", w)
	}
	if w := byName["impact"]; w.Widget != "textarea" {
		t.Fatalf("Proposal.impact widget=%s", w.Widget)
	}
	if w := byName["priority"]; w.Widget != "selector" || len(w.Enum) != 4 || w.Default != "low" {
		t.Fatalf("Proposal.priority widget=%+v", w)
	}
	if w := byName["landing"]; w.Widget != "object" || len(w.Fields) < 1 || w.Fields[0].Widget != "selector" {
		t.Fatalf("Proposal.landing widget=%+v", w)
	}

	// 验证 18 kind 全量覆盖 R-UX-2 八类控件：selector/input/textarea/list/ref/switch/map/number
	seen := map[string]bool{}
	var walk func(ws []FieldDescriptor)
	walk = func(ws []FieldDescriptor) {
		for _, w := range ws {
			seen[w.Widget] = true
			if w.Elem != nil {
				walk([]FieldDescriptor{*w.Elem})
			}
			if len(w.Fields) > 0 {
				walk(w.Fields)
			}
		}
	}
	for _, k := range sch.ListKinds() {
		walk(k.Widgets)
	}
	for _, need := range []string{"selector", "input", "textarea", "list", "ref", "switch", "map", "number"} {
		if !seen[need] {
			t.Errorf("R-UX-2 缺少控件类型 %q（已见：%v）", need, seen)
		}
	}
}

func TestSchemeReplaceHotReload(t *testing.T) {
	schPath := coreSchemePath()
	sch, err := LoadFile(schPath)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	modSch, err := LoadFile(schPath)
	if err != nil {
		t.Fatalf("LoadFile2: %v", err)
	}
	for i := range modSch.Spec.Kinds {
		if modSch.Spec.Kinds[i].Name == "Proposal" {
			modSch.Spec.Kinds[i].Spec["priority"] = "low|normal|high|critical|blocker"
			modSch.Spec.Kinds[i].Spec["extraNote"] = "?string"
		}
	}

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				_, _ = sch.Lookup("Proposal")
				_ = sch.ListKinds()
				_ = sch.Summary()
			}
		}()
	}
	sch.Replace(modSch)
	wg.Wait()

	k, err := sch.Lookup("Proposal")
	if err != nil {
		t.Fatalf("Lookup after Replace: %v", err)
	}
	foundBlocker := false
	foundExtra := false
	for _, w := range k.Widgets {
		if w.Name == "priority" {
			for _, e := range w.Enum {
				if e == "blocker" {
					foundBlocker = true
				}
			}
		}
		if w.Name == "extraNote" && w.Widget == "input" {
			foundExtra = true
		}
	}
	if !foundBlocker || !foundExtra {
		t.Fatalf("Replace 后控件未自动反映：foundBlocker=%v foundExtra=%v", foundBlocker, foundExtra)
	}
}
