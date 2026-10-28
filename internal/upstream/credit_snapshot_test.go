package upstream

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/misakano7545/workbuddy2api-panel/internal/auth"
)

// 同一 PackageCode 下多个独立日包（ResourceId 不同）必须各自成为一批。
// 这是生产 900/900 卡片 bug 的回归：按 PackageCode 归并会把 30 个 30 分包
// 合成「最早一天到期 900」，并吞掉其余 29 个批次。
func TestGlobalSnapshotKeepsSameCodeDailyGrants(t *testing.T) {
	now := time.Now().In(softRateResetLoc)
	day := func(n int) time.Time { return now.Add(time.Duration(n) * 24 * time.Hour) }
	ms := func(t time.Time) int64 { return t.UnixMilli() }
	row := func(id string, end time.Time, remain int) string {
		return fmt.Sprintf(`{"PackageCode":"PKG-A","PackageName":"Bonus Pack","ResourceId":%q,`+
			`"CreateTime":%d,"DeductionEndTime":%d,"CycleCapacitySize":%d,`+
			`"CycleCapacityRemain":%d,"CycleCapacityUsed":0}`,
			id, ms(end.Add(-30*24*time.Hour)), ms(end), remain, remain)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/get-user-resource-summary"):
			fmt.Fprint(w, `{"code":0,"data":{"Packages":[`+
				`{"PackageCode":"PKG-A","CycleTotalCapacity":"90","CycleRemainCapacity":"90"},`+
				`{"PackageCode":"PKG-L","CycleTotalCapacity":"100","CycleRemainCapacity":"100"}`+
				`]}}`)
		case strings.HasSuffix(r.URL.Path, "/get-user-resource-free-packages"):
			rows := []string{
				row("r1", day(1), 30),
				row("r2", day(3), 30),
				row("r3", day(5), 30),
				fmt.Sprintf(`{"PackageCode":"PKG-L","PackageName":"Free Plan","ResourceId":"rl",`+
					`"DeductionEndTime":%d,"CycleCapacitySize":100,"CycleCapacityRemain":100,"CycleCapacityUsed":0}`,
					ms(now.AddDate(8, 0, 0))),
			}
			fmt.Fprintf(w, `{"code":0,"data":{"Accounts":[%s],"TotalCount":%d}}`,
				strings.Join(rows, ","), len(rows))
		case strings.HasSuffix(r.URL.Path, "/get-user-resource-paid-packages"):
			fmt.Fprint(w, `{"code":0,"data":{"Accounts":[],"TotalCount":0}}`)
		default:
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	c := &Client{HTTP: srv.Client(), BillingBaseGlobal: srv.URL, GlobalEnabled: true}
	remain, total, expiring, earliestAt, earliestRemaining, buckets, err :=
		c.UserResourceSnapshot(globalAuth(), 168*time.Hour)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if remain != 190 || total != 190 {
		t.Fatalf("remain/total=%d/%d want 190/190", remain, total)
	}
	if len(buckets) != 4 {
		t.Fatalf("buckets=%+v want 4（3 个日包 + 1 个长期包）", buckets)
	}
	if buckets[0].Remain != 30 || buckets[1].Remain != 30 || buckets[2].Remain != 30 {
		t.Fatalf("日包被合并: %+v", buckets)
	}
	if earliestRemaining != 30 {
		t.Fatalf("earliestRemaining=%d want 30（最早一天只有 1 个包）", earliestRemaining)
	}
	// +1/+3/+5 天共 3 批在 168h 窗口内。
	if expiring != 90 {
		t.Fatalf("expiring=%d want 90", expiring)
	}
	if earliestAt.IsZero() || earliestAt.Sub(day(1)) > time.Minute {
		t.Fatalf("earliestAt=%v want ≈%v", earliestAt, day(1))
	}

	packs, pRemain, pTotal, err := c.CreditPackages(globalAuth())
	if err != nil {
		t.Fatalf("packages: %v", err)
	}
	if len(packs) != 4 {
		t.Fatalf("packages=%d want 4（同 code 不能合并）", len(packs))
	}
	if pRemain != 190 || pTotal != 190 {
		t.Fatalf("packages remain/total=%d/%d want 190/190", pRemain, pTotal)
	}
}

// 同一 ResourceId 的多个计量周期只取最早未来且仍有余额的一期，
// 不把同一笔余额重复计入批次。
func TestGlobalSnapshotDedupesResourceCycles(t *testing.T) {
	now := time.Now().In(softRateResetLoc)
	ms := func(t time.Time) int64 { return t.UnixMilli() }
	row := func(end time.Time, remain int) string {
		return fmt.Sprintf(`{"PackageCode":"PKG-C","PackageName":"Subscription","ResourceId":"same",`+
			`"DeductionEndTime":%d,"CycleCapacitySize":100,"CycleCapacityRemain":%d,"CycleCapacityUsed":0}`,
			ms(end), remain)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/get-user-resource-summary"):
			fmt.Fprint(w, `{"code":0,"data":{"Packages":[`+
				`{"PackageCode":"PKG-C","CycleTotalCapacity":"40","CycleRemainCapacity":"40"}]}}`)
		case strings.HasSuffix(r.URL.Path, "/get-user-resource-free-packages"):
			fmt.Fprintf(w, `{"code":0,"data":{"Accounts":[%s,%s],"TotalCount":2}}`,
				row(now.Add(2*24*time.Hour), 40), row(now.Add(10*24*time.Hour), 60))
		case strings.HasSuffix(r.URL.Path, "/get-user-resource-paid-packages"):
			fmt.Fprint(w, `{"code":0,"data":{"Accounts":[],"TotalCount":0}}`)
		default:
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	c := &Client{HTTP: srv.Client(), BillingBaseGlobal: srv.URL, GlobalEnabled: true}
	_, _, _, _, earliestRemaining, buckets, err := c.UserResourceSnapshot(globalAuth(), 0)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if len(buckets) != 1 || buckets[0].Remain != 40 || earliestRemaining != 40 {
		t.Fatalf("buckets=%+v earliest=%d want 单批 40", buckets, earliestRemaining)
	}
	packs, _, _, err := c.CreditPackages(globalAuth())
	if err != nil {
		t.Fatalf("packages: %v", err)
	}
	if len(packs) != 1 || packs[0].Remain != 40 {
		t.Fatalf("packages=%+v want 单包 40", packs)
	}
}

// 明细整体失败时余额仍可用（与既有降级语义一致），但「积分构成」必须显式报错，
// 不能静默显示 0 个包。
func TestGlobalSnapshotDetailFailureKeepsTotals(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/get-user-resource-summary") {
			fmt.Fprint(w, `{"code":0,"data":{"Packages":[`+
				`{"PackageCode":"PKG-A","CycleTotalCapacity":"100","CycleRemainCapacity":"100"}]}}`)
			return
		}
		w.WriteHeader(500)
	}))
	defer srv.Close()

	c := &Client{HTTP: srv.Client(), BillingBaseGlobal: srv.URL, GlobalEnabled: true}
	remain, total, _, _, _, buckets, err := c.UserResourceSnapshot(globalAuth(), 0)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if remain != 100 || total != 100 || len(buckets) != 0 {
		t.Fatalf("remain/total/buckets=%d/%d/%d want 100/100/0", remain, total, len(buckets))
	}
	if _, _, _, err := c.CreditPackages(globalAuth()); err == nil {
		t.Fatal("明细失败时 CreditPackages 必须报错，不能静默返回 0 个包")
	}
}

// CN 路径：逐包构成与到期批次必须来自同一次解析、同一套 remain/到期口径。
func TestCNSnapshotPackagesAgreeWithBuckets(t *testing.T) {
	now := time.Now().In(softRateResetLoc)
	soon := now.Add(48 * time.Hour).Format(packageEndLayout)
	later := now.Add(10 * 24 * time.Hour).Format(packageEndLayout)
	payload := `{"code":0,"data":{"Response":{"Data":{"Accounts":[` +
		`{"PackageName":"soon","PackageCode":"P1","CycleCapacitySize":10,"CycleCapacityRemain":7,"CycleCapacityUsed":3,"CycleEndTime":"` + soon + `"},` +
		`{"PackageName":"later","PackageCode":"P2","CycleCapacitySize":20,"CycleCapacityRemain":15,"CycleCapacityUsed":5,"CycleEndTime":"` + later + `"}` +
		`]}}}}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/v2/billing/meter/get-user-resource") {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		fmt.Fprint(w, payload)
	}))
	defer srv.Close()

	c := &Client{HTTP: srv.Client(), BillingBaseCN: srv.URL}
	a := &auth.Auth{UID: "u1", AccessToken: "at"}
	remain, total, expiring, _, earliestRemaining, buckets, err := c.UserResourceSnapshot(a, 168*time.Hour)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if remain != 22 || total != 30 || expiring != 7 || earliestRemaining != 7 {
		t.Fatalf("remain/total/expiring/earliest=%d/%d/%d/%d want 22/30/7/7",
			remain, total, expiring, earliestRemaining)
	}
	if len(buckets) != 2 || buckets[0].Remain != 7 || buckets[1].Remain != 15 {
		t.Fatalf("buckets=%+v want 7/15", buckets)
	}
	packs, pRemain, pTotal, err := c.CreditPackages(a)
	if err != nil {
		t.Fatalf("packages: %v", err)
	}
	if len(packs) != 2 || pRemain != remain || pTotal != total {
		t.Fatalf("packages=%+v remain/total=%d/%d want 2 包 / %d/%d",
			packs, pRemain, pTotal, remain, total)
	}
	if packs[0].Size != 20 || packs[1].Size != 10 {
		t.Fatalf("包应按面额降序: %+v", packs)
	}
}
