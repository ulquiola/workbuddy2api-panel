package panel

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/misakano7545/workbuddy2api-panel/internal/auth"
	"github.com/misakano7545/workbuddy2api-panel/internal/pool"
	"github.com/misakano7545/workbuddy2api-panel/internal/usage"
)

// TestOverviewIncludesExpiryBuckets 首页快照的数据源接线：overview 必须把池内按日
// 到期批次原样透出，并给出余额后台刷新的节奏/年龄（没有调度器时为 0）。
// 前端 renderExpirySnapshot 只吃这两个字段，缺一个就会退回「只列最早一批」的旧口径。
func TestOverviewIncludesExpiryBuckets(t *testing.T) {
	p := pool.New("")
	p.Add(&auth.Auth{UID: "u1", Nickname: "A", AccessToken: "at"})
	at := time.Now().Add(72 * time.Hour)
	p.SetCreditsSnapshot("u1", pool.CreditsSnapshot{
		Remain: 100, Total: 100,
		ExpiryBuckets: []pool.ExpiryBucket{
			{Date: at.Format("2006-01-02"), ExpireAt: at, Remain: 40},
		},
	})
	pn := New(Config{Version: "test", Pool: p})

	rec := httptest.NewRecorder()
	pn.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/panel/api/overview", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("overview code=%d body=%s", rec.Code, rec.Body)
	}
	var d struct {
		BalanceRefreshSec  int       `json:"balance_refresh_sec"`
		BalanceRefreshedAt time.Time `json:"balance_refreshed_at"`
		Accounts           []struct {
			UID                  string              `json:"uid"`
			CreditsExpiryBuckets []pool.ExpiryBucket `json:"credits_expiry_buckets"`
		} `json:"accounts"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &d); err != nil {
		t.Fatal(err)
	}
	if len(d.Accounts) != 1 || len(d.Accounts[0].CreditsExpiryBuckets) != 1 {
		t.Fatalf("overview 未透出 buckets: %s", rec.Body)
	}
	b := d.Accounts[0].CreditsExpiryBuckets[0]
	if b.Date != at.Format("2006-01-02") || b.Remain != 40 {
		t.Fatalf("bucket=%+v want %s/40", b, at.Format("2006-01-02"))
	}
	if d.BalanceRefreshSec != 0 {
		t.Fatalf("无调度器时 balance_refresh_sec=%d want 0", d.BalanceRefreshSec)
	}
}

// TestOverviewIncludesCreditBurn7d 首页必须透出近 7 天按账号的实测扣分：
// 前端到期提醒用它估算「按当前速度会用掉多少、可能浪费多少」。
// 没有扣分样本的账号不出现在 map 里（「没测到」≠「花了 0」）。
func TestOverviewIncludesCreditBurn7d(t *testing.T) {
	p := pool.New("")
	p.Add(&auth.Auth{UID: "u1", Nickname: "A", AccessToken: "at"})
	p.Add(&auth.Auth{UID: "u2", Nickname: "B", AccessToken: "at"})
	rec := usage.New("")
	now := time.Now()
	rec.Add(now, "global", "u1", "m", usage.Delta{Credit: 12, HasCredit: true}, true)
	rec.Add(now, "global", "u1", "m", usage.Delta{Credit: 0, HasCredit: true}, true)
	rec.Add(now, "global", "u2", "m", usage.Delta{}, true) // 无 credit 样本

	pn := New(Config{Version: "test", Pool: p, Usage: rec})
	rr := httptest.NewRecorder()
	pn.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/panel/api/overview", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("overview code=%d body=%s", rr.Code, rr.Body)
	}
	var d struct {
		CreditBurn7d map[string]usage.CreditBurn `json:"credit_burn_7d"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &d); err != nil {
		t.Fatal(err)
	}
	u1, ok := d.CreditBurn7d["u1"]
	if !ok || u1.Credits != 12 || u1.Samples != 2 {
		t.Fatalf("credit_burn_7d[u1]=%+v ok=%v want credits=12 samples=2", u1, ok)
	}
	if _, ok := d.CreditBurn7d["u2"]; ok {
		t.Fatal("没有 credit 样本的账号不该出现在 credit_burn_7d")
	}
}
