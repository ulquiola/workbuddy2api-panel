package pool

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/misakano7545/workbuddy2api-panel/internal/auth"
)

// TestSetAliasPersists 本地备注名：Status 立即透出、state.json 落盘、重启往返、
// 空值清除、未知 uid 拒绝。别名与上游 nickname 独立（昵称同步不会覆盖它）。
func TestSetAliasPersists(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "state.json")
	p := New(fp)
	p.Add(&auth.Auth{UID: "u1", Nickname: "muhe"})

	if !p.SetAlias("u1", "主号") {
		t.Fatal("SetAlias 应返回 true")
	}
	st, ok := p.Status("u1")
	if !ok || st.Alias != "主号" {
		t.Fatalf("Status.Alias=%q ok=%v", st.Alias, ok)
	}
	if st.Nickname != "muhe" {
		t.Fatalf("别名不应改写上游昵称: %q", st.Nickname)
	}

	p.Flush()
	raw, err := os.ReadFile(fp)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"alias": "主号"`) {
		t.Fatalf("state.json 缺少 alias:\n%s", raw)
	}

	// 重启往返：别名从 state.json 恢复。
	p2 := New(fp)
	p2.Add(&auth.Auth{UID: "u1", Nickname: "muhe"})
	if st, _ := p2.Status("u1"); st.Alias != "主号" {
		t.Fatalf("重启后 alias 丢失: %q", st.Alias)
	}

	// 空值清除。
	if !p2.SetAlias("u1", "") {
		t.Fatal("清除别名应返回 true")
	}
	if st, _ := p2.Status("u1"); st.Alias != "" {
		t.Fatalf("清除失败: %q", st.Alias)
	}

	// 未知 uid 返回 false（与「值未变化」无关）。
	if p2.SetAlias("no-such-uid", "x") {
		t.Fatal("未知 uid 应返回 false")
	}
}