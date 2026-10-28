package panel

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/misakano7545/workbuddy2api-panel/internal/auth"
	"github.com/misakano7545/workbuddy2api-panel/internal/pool"
)

// TestAccountAliasEndpoint 本地备注名接口：鉴权、trim、长度上限、清除、未知账号。
func TestAccountAliasEndpoint(t *testing.T) {
	p := New(Config{Version: "test", APIKey: "test-key", Pool: pool.New("")})
	p.cfg.Pool.Add(&auth.Auth{UID: "u1", AccessToken: "at", Nickname: "muhe"})

	post := func(path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest("POST", path, strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer test-key")
		rec := httptest.NewRecorder()
		p.ServeHTTP(rec, req)
		return rec
	}

	if rec := post("/panel/api/accounts/u1/alias", `{"alias":" 主号 "}`); rec.Code != 200 {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body)
	}
	if st, _ := p.cfg.Pool.Status("u1"); st.Alias != "主号" {
		t.Fatalf("alias=%q want 主号（应 trim）", st.Alias)
	}
	if st, _ := p.cfg.Pool.Status("u1"); st.Nickname != "muhe" {
		t.Fatalf("备注不应改写昵称: %q", st.Nickname)
	}

	// 超过 32 个字符（按 rune 计）拒绝。
	if rec := post("/panel/api/accounts/u1/alias", `{"alias":"`+strings.Repeat("长", 33)+`"}`); rec.Code != 400 {
		t.Fatalf("超长应 400, got %d body=%s", rec.Code, rec.Body)
	}

	// 空值清除。
	if rec := post("/panel/api/accounts/u1/alias", `{"alias":""}`); rec.Code != 200 {
		t.Fatalf("清除 code=%d body=%s", rec.Code, rec.Body)
	}
	if st, _ := p.cfg.Pool.Status("u1"); st.Alias != "" {
		t.Fatalf("清除失败: %q", st.Alias)
	}

	// 未知账号 404。
	if rec := post("/panel/api/accounts/nope/alias", `{"alias":"x"}`); rec.Code != 404 {
		t.Fatalf("未知账号应 404, got %d body=%s", rec.Code, rec.Body)
	}

	// 未鉴权 401。
	req := httptest.NewRequest("POST", "/panel/api/accounts/u1/alias", strings.NewReader(`{"alias":"x"}`))
	rec := httptest.NewRecorder()
	p.ServeHTTP(rec, req)
	if rec.Code != 401 {
		t.Fatalf("未鉴权应 401, got %d", rec.Code)
	}
}