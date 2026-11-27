package models

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestNodeWithSortSerializesSortAndNodeNames(t *testing.T) {
	node := NodeWithSort{Node: Node{Name: "Remark", LinkName: "Original"}, Sort: 7}
	body, err := json.Marshal(node)
	if err != nil {
		t.Fatal(err)
	}
	var fields struct {
		Name, LinkName, NameMode, EffectiveName string
		Sort                                    int
	}
	if err := json.Unmarshal(body, &fields); err != nil {
		t.Fatal(err)
	}
	if fields.Sort != 7 || fields.Name != "Remark" || fields.LinkName != "Original" || fields.NameMode != NodeNameModeLink || fields.EffectiveName != "Original" {
		t.Fatalf("unexpected sorted node JSON: %s", body)
	}
}

func TestNodeNamePlanLegacyRules(t *testing.T) {
	node := Node{
		ID: 1, Name: "Effective", LinkName: "Original", NameMode: NodeNameModeRemark,
		Link: "ss://first,ss://second", UnlockSummary: `{"providers":[{"provider":"OpenAI","status":"available","region":"US"},{"provider":"netflix","status":"partial"}]}`,
	}
	for _, tt := range []struct {
		rule string
		want []string
	}{
		{"", []string{"Effective", "Effective"}},
		{"$Name/$LinkName/$Unlock(openai)/$Unlock(netflix)/$Unlock(youtube_premium)/$UnlockStatus/$UnlockRegion", []string{
			"Effective/Original/解锁-US/部分/未检测/available/US", "Effective/Original/解锁-US/部分/未检测/available/US",
		}},
		{"$Name $DuplicateIndex", []string{"Effective", "Effective 1"}},
	} {
		plan := BuildNodeNamePlan([]Node{node}, "", tt.rule, nil)
		if got := plan.SplitNamesAt(0); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("rule %q: got %v, want %v", tt.rule, got, tt.want)
		}
	}
}

func TestLegacyNodeNameRuleKeepsLiteralServiceMarkTextAndArguments(t *testing.T) {
	for _, rule := range []string{
		`$Name/$ServiceMarks(v1)`,
		`$Name/$TagGroup(team$ServiceMarks({sample)`,
		`$Name/$Unlock(provider$ServiceMarks({sample)`,
		`$Name/$TagGroup(team$ServiceMarks({"services":[{"provider":"openai","label":"GPT","acceptedStatuses":["available"]}]})`,
	} {
		compiled, invalid := compileNodeNameRule(rule)
		if invalid != nil || compiled.legacy != rule {
			t.Errorf("legacy rule %q was reinterpreted: %+v, %+v", rule, compiled, invalid)
		}
	}
}

func TestLegacyUnlockVariableIgnoresUnrelatedSummaryFields(t *testing.T) {
	node := Node{
		LinkName:      "Original",
		UnlockSummary: `{"providers":[{"provider":"openai","status":"available","region":"US","detail":{"legacy":true}}],"updatedAt":1730000000}`,
	}
	const rule = `$Unlock(openai)/$Unlock/$UnlockStatus`
	if got := BuildNodeExportName(node, node.LinkName, rule, "", 1, 0); got != "解锁-US/未检测/未检测" {
		t.Fatalf("legacy values = %q", got)
	}
}

const testServiceMarksRule = `$Name$ServiceMarks({"prefix":" · ","services":[{"provider":"youtube_premium","label":"YT","acceptedStatuses":["available"]},{"provider":"openai","label":"GPT","acceptedStatuses":["available","partial"],"region":true}]})`

func TestNodeNameServiceMarks(t *testing.T) {
	for _, tt := range []struct {
		name    string
		summary string
		want    string
	}{
		{"all", `{"providers":[{"provider":" OpenAI ","status":"partial","region":"US"},{"provider":"YouTube-Premium","status":"available","region":"JP"}]}`, "香港 01 · YT · GPT-US"},
		{"some", `{"providers":[{"provider":"openai","status":" AVAILABLE "},{"provider":"youtube_premium","status":"restricted"}]}`, "香港 01 · GPT"},
		{"none", `{"providers":[{"provider":"openai","status":"reachable"}]}`, "香港 01"},
		{"missing", "", "香港 01"},
		{"damaged data", `{"providers":[`, "香港 01"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			node := Node{ID: 1, LinkName: "香港 01", Link: "ss://x", UnlockSummary: tt.summary}
			plan := BuildNodeNamePlan([]Node{node}, "", testServiceMarksRule, nil)
			if got := plan.NodeNameAt(0, node.ID); got != tt.want {
				t.Fatalf("planned name = %q, want %q", got, tt.want)
			}
			if got := BuildNodeExportName(node, node.LinkName, testServiceMarksRule, "ss", 1, 0); got != tt.want {
				t.Fatalf("single name = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNodeNameServiceMarksPreserveLiteralTextAndDuplicateIndexes(t *testing.T) {
	node := Node{ID: 1, LinkName: "Original", Link: "ss://x,ss://y", UnlockSummary: `{"providers":[{"provider":"openai","status":"available"}]}`}
	rule := `$ServiceMarks({"suffix":" / ","services":[{"provider":"openai","label":"A/\"未送中(字面量)\" $DuplicateIndex $Name","acceptedStatuses":["available"]}]})$Name`
	want := `A/"未送中(字面量)" $DuplicateIndex $Name / Original`
	plan := BuildNodeNamePlan([]Node{node}, "", rule, nil)
	if got := plan.SplitNamesAt(0); !reflect.DeepEqual(got, []string{want, want}) {
		t.Fatalf("literal mark names = %v", got)
	}
	plan = BuildNodeNamePlan([]Node{node}, "", rule+` $DuplicateIndex`, nil)
	if got := plan.SplitNamesAt(0); !reflect.DeepEqual(got, []string{want, want + " 1"}) {
		t.Fatalf("numbered mark names = %v", got)
	}
	if got := plan.NodeNameMap[node.ID]; got != want {
		t.Fatalf("node name reference = %q, want %q", got, want)
	}
}

func TestValidateNodeNameServiceMarks(t *testing.T) {
	for _, tt := range []struct {
		rule   string
		reason string
	}{
		{`$ServiceMarks({"services":[]})`, "services"},
		{`$ServiceMarks({"services":[{"label":"GPT","acceptedStatuses":["available"]}]})`, "provider"},
		{`$ServiceMarks({"services":[{"provider":"openai","label":" ","acceptedStatuses":["available"]}]})`, "label"},
		{`$ServiceMarks({"services":[{"provider":"openai","label":"GPT","acceptedStatuses":["all"]}]})`, "statuses"},
		{`$ServiceMarks({"services":[{"provider":"openai","label":"bad\nname","acceptedStatuses":["available"]}]})`, "text"},
		{`$ServiceMarks({"services":[{"provider":"openai","label":"GPT","acceptedStatuses":["available"]}],"separator":","})`, "text"},
		{`$ServiceMarks({"services":[{"provider":"openai","label":"A","acceptedStatuses":["available"]},{"provider":" OpenAI ","label":"B","acceptedStatuses":["partial"]}]})`, "duplicate"},
		{`$ServiceMarks({"services":[]}`, "syntax"},
	} {
		rule := testServiceMarksRule + tt.rule
		if got := ValidateNodeNameRule(rule); got == nil || got.Block != 2 || got.Reason != tt.reason {
			t.Errorf("validate %q = %+v, want block 2 with reason %q", tt.rule, got, tt.reason)
		}
		if got := BuildNodeExportName(Node{LinkName: "Original"}, "Original", rule, "", 1, 0); got != "Original" {
			t.Errorf("damaged template rendered %q, want effective name", got)
		}
	}
}

func BenchmarkNodeNamePlan(b *testing.B) {
	nodes := make([]Node, 1000)
	for i := range nodes {
		nodes[i] = Node{
			ID: i + 1, LinkName: fmt.Sprintf("Node %d", i), Link: "ss://x,ss://y",
			UnlockSummary: `{"providers":[{"provider":"youtube_premium","status":"available"},{"provider":"openai","status":"partial","region":"US"}]}`,
		}
	}
	for _, tt := range []struct{ name, rule string }{
		{"legacy", `$Name $Unlock(openai) $DuplicateIndex`},
		{"marks", testServiceMarksRule + ` $DuplicateIndex`},
	} {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				BuildNodeNamePlan(nodes, "", tt.rule, nil)
			}
		})
	}
}
