package upstream

import (
	"encoding/json"
	"strings"
	"testing"
)

func mergeMsgs(t *testing.T, s string) []any {
	t.Helper()
	var v []any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		t.Fatalf("parse messages: %v", err)
	}
	return v
}

func mergeShape(messages []any) []string {
	out := make([]string, 0, len(messages))
	for _, m := range messages {
		msg, _ := m.(map[string]any)
		role, _ := msg["role"].(string)
		if role == "tool" {
			id, _ := msg["tool_call_id"].(string)
			out = append(out, "tool("+id+")")
			continue
		}
		var ids []string
		if tcs, ok := msg["tool_calls"].([]any); ok {
			for _, tci := range tcs {
				if tc, ok := tci.(map[string]any); ok {
					id, _ := tc["id"].(string)
					ids = append(ids, id)
				}
			}
		}
		if len(ids) > 0 {
			out = append(out, role+"("+strings.Join(ids, ",")+")")
		} else {
			out = append(out, role+"(-)")
		}
	}
	return out
}

func assertMergeShape(t *testing.T, got, want []string) {
	t.Helper()
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("shape=%v want %v", got, want)
	}
}

// TestMergeAdjacentToolCalls 背靠背的 assistant.tool_calls 必须合成一条（上游 11148）。
func TestMergeAdjacentToolCalls(t *testing.T) {
	const (
		a00  = `{"role":"assistant","content":null,"tool_calls":[{"id":"c00","type":"function","function":{"name":"f","arguments":"{}"}}]}`
		a01  = `{"role":"assistant","content":null,"tool_calls":[{"id":"c01","type":"function","function":{"name":"f","arguments":"{}"}}]}`
		t00  = `{"role":"tool","tool_call_id":"c00","content":"r0"}`
		t01  = `{"role":"tool","tool_call_id":"c01","content":"r1"}`
		user = `{"role":"user","content":"x"}`
	)

	in := mergeMsgs(t, "["+a00+","+a01+","+t00+","+t01+"]")
	out, changed := mergeAdjacentToolCalls(in)
	if !changed {
		t.Fatal("相邻两条 assistant(tool_calls) 未合并")
	}
	assertMergeShape(t, mergeShape(out), []string{"assistant(c00,c01)", "tool(c00)", "tool(c01)"})

	// 中间隔着消息 = 不是同一批声明，不合并。
	in2 := mergeMsgs(t, "["+a00+","+user+","+a01+"]")
	if _, changed2 := mergeAdjacentToolCalls(in2); changed2 {
		t.Fatal("非相邻消息不应合并")
	}

	// 后一条 content 非空：无法无损拼接，不合并。
	in3 := mergeMsgs(t, `[`+a00+`,{"role":"assistant","content":"text","tool_calls":[{"id":"c01"}]}]`)
	if _, changed3 := mergeAdjacentToolCalls(in3); changed3 {
		t.Fatal("后一条有正文时不应合并")
	}

	// content:[] 也按空正文处理。
	in4 := mergeMsgs(t, `[{"role":"assistant","content":[],"tool_calls":[{"id":"c1"}]},{"role":"assistant","content":[],"tool_calls":[{"id":"c2"}]}]`)
	if out4, changed4 := mergeAdjacentToolCalls(in4); !changed4 || len(out4) != 1 {
		t.Fatalf("content:[] 应合并: changed=%v len=%d", changed4, len(out4))
	}

	// 纯正文 assistant 折进前一条带 tool_calls 的空正文 assistant。
	in5 := mergeMsgs(t, `[`+a00+`,{"role":"assistant","content":"说明"},`+t00+`]`)
	out5, changed5 := mergeAdjacentToolCalls(in5)
	if !changed5 {
		t.Fatal("正文应折进前一条工具调用")
	}
	first, _ := out5[0].(map[string]any)
	if first["content"] != "说明" {
		t.Fatalf("content=%#v", first["content"])
	}
	assertMergeShape(t, mergeShape(out5), []string{"assistant(c00)", "tool(c00)"})
}

// TestPrepareBodyMergesSplitParallelToolCalls 出站管线必须先生成合并后的 messages。
func TestPrepareBodyMergesSplitParallelToolCalls(t *testing.T) {
	body := `{"model":"deepseek-v4.1-flash","messages":[
		{"role":"user","content":"跑两个命令"},
		{"role":"assistant","content":null,"tool_calls":[{"id":"call_00_a","type":"function","function":{"name":"exec_command","arguments":"{\"cmd\":\"pwd\"}"}}]},
		{"role":"assistant","content":null,"tool_calls":[{"id":"call_01_b","type":"function","function":{"name":"exec_command","arguments":"{\"cmd\":\"ls\"}"}}]},
		{"role":"tool","tool_call_id":"call_00_a","content":"/opt"},
		{"role":"tool","tool_call_id":"call_01_b","content":"a b c"}
	]}`
	out := PrepareBodyOptWithEfforts([]byte(body), false, nil)
	var obj map[string]any
	if err := json.Unmarshal(out, &obj); err != nil {
		t.Fatalf("unmarshal 出站 body: %v", err)
	}
	messages, _ := obj["messages"].([]any)
	assertMergeShape(t, mergeShape(messages), []string{
		"user(-)", "assistant(call_00_a,call_01_b)", "tool(call_00_a)", "tool(call_01_b)",
	})
}
