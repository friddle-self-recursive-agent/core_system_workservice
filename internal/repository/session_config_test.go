package repository

import "testing"

func TestBumpSidecarImportance(t *testing.T) {
	side := "tags:\n  - {tag: context, tag_type: domain}\nimportance: 82\n"
	out, delta, ok := BumpSidecarImportance(side, 5)
	if !ok || delta != 5 {
		t.Fatalf("bump: ok=%v delta=%d", ok, delta)
	}
	if !contains(out, "importance: 87") {
		t.Fatalf("bumped content: %q", out)
	}
	// 封顶 100
	out2, d2, _ := BumpSidecarImportance("importance: 99\n", 5)
	if d2 != 1 || !contains(out2, "importance: 100") {
		t.Fatalf("cap: delta=%d out=%q", d2, out2)
	}
	// 无 importance 行=原样
	same, d3, ok3 := BumpSidecarImportance("tags:\n  - {tag: a, tag_type: doc}\n", 5)
	if ok3 || d3 != 0 || same != "tags:\n  - {tag: a, tag_type: doc}\n" {
		t.Fatalf("no-importance line: ok=%v delta=%d", ok3, d3)
	}
}

func TestMergeSidecarTags(t *testing.T) {
	side := "tags:\n  - {tag: context, tag_type: domain}\nimportance: 82\n"
	out, added := MergeSidecarTags(side, []string{"audit", "context", "session-x"})
	if len(added) != 2 || added[0] != "audit" {
		t.Fatalf("added=%v", added)
	}
	if !contains(out, "{tag: audit, tag_type: session}") || !contains(out, "{tag: session-x, tag_type: session}") {
		t.Fatalf("merged=%q", out)
	}
	if !contains(out, "importance: 82") {
		t.Fatalf("importance line lost: %q", out)
	}
	// 无 tags 块：importance 前插新块
	out2, added2 := MergeSidecarTags("importance: 10\n", []string{"k1"})
	if len(added2) != 1 || !contains(out2, "tags:\n  - {tag: k1, tag_type: session}\nimportance: 10") {
		t.Fatalf("new block: added=%v out=%q", added2, out2)
	}
}

func contains(s, sub string) bool { return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0) }

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
