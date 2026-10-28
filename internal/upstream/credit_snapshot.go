// credit_snapshot.go 积分快照的唯一聚合入口。
//
// 面板有两个消费面：账号池的「到期提醒」（按自然日聚合的 batches）与「积分构成」
// （逐包明细）。两者此前各走一条解析/聚合路径——global 的到期批次按 PackageCode
// 归并后套 summary 的 code 总额，把同一 code 下多个独立日包（ResourceId 不同、
// 到期日不同）合并成「最早一天到期全部余额」，30 个 30 分包被显示成 1 个 900 包。
//
// 现在 CN 与 global 都先解析成同一套 creditGrant，再由 buildCreditSnapshot 一次
// 产出 totals / buckets / packages。包的身份是 ResourceId（缺失时回退
// PackageCode+到期时刻+发放时刻），同一个资源的多个计量周期只取最早未来那条；
// 不同 ResourceId 的包即使 PackageCode 相同也分别计入。
package upstream

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/misakano7545/workbuddy2api-panel/internal/auth"
	"github.com/misakano7545/workbuddy2api-panel/internal/logfmt"
)

// CreditSnapshot 一次额度查询的完整快照。
//
// Buckets 与 Packages 来自同一批 creditGrant，面板「到期提醒」与「积分构成」因此
// 永远同口径；Remain/Total 对 global 取 summary 的权威值，对 CN 取逐包求和。
type CreditSnapshot struct {
	Remain            int64
	Total             int64
	Expiring          int64
	EarliestAt        time.Time
	EarliestRemaining int64
	Buckets           []ExpiryBucket
	Packages          []CreditPackage
	// packagesErr 明细拉取失败：只影响「积分构成」，不影响余额/批次。
	packagesErr error
}

// creditGrant 单个积分包的规范化形态（CN/global 共用）。
type creditGrant struct {
	ResourceID     string
	PackageCode    string
	Name           string
	SubProductCode string
	SubProductName string
	End            time.Time
	HasEnd         bool
	CreatedAt      time.Time
	Remain         int64
	Used           int64
	Size           int64
	Cycle          bool
}

// creditAccountRow 是 CN get-user-resource 与 global paid/free 明细的共同字段形态。
// 两端返回的 Accounts[] 同构（global 多 ResourceId/CreateTime/PackageCode 等字段）。
type creditAccountRow struct {
	PackageName         string `json:"PackageName"`
	PackageCode         string `json:"PackageCode"`
	SubProductCode      string `json:"SubProductCode"`
	SubProductName      string `json:"SubProductName"`
	DeductionEndTime    int64  `json:"DeductionEndTime"`
	ExpiredTime         string `json:"ExpiredTime"`
	CycleEndTime        string `json:"CycleEndTime"`
	CreateTime          int64  `json:"CreateTime"`
	CapacityRemain      int64  `json:"CapacityRemain"`
	CapacityUsed        int64  `json:"CapacityUsed"`
	CapacitySize        int64  `json:"CapacitySize"`
	CycleCapacityRemain int64  `json:"CycleCapacityRemain"`
	CycleCapacityUsed   int64  `json:"CycleCapacityUsed"`
	CycleCapacitySize   int64  `json:"CycleCapacitySize"`
	ResourceID          string `json:"ResourceId"`
}

// CreditSnapshot 查询一次账号额度，返回余额、快到期、按日批次与逐包构成。
// global 的 summary 失败时回落到 CN 分页路径（老路径兼容尝试）。
func (c *Client) CreditSnapshot(a *auth.Auth, soon time.Duration) (CreditSnapshot, error) {
	return c.creditSnapshot(a, soon, true)
}

// creditSnapshot 是 CreditSnapshot 的内部实现。notify=false 供「积分构成」这类
// 只读页面使用：它同样需要逐包数据，但不该把一次页面访问写进积分流水观察。
func (c *Client) creditSnapshot(a *auth.Auth, soon time.Duration, notify bool) (CreditSnapshot, error) {
	var (
		snap CreditSnapshot
		err  error
	)
	switch {
	case a.UsesEnterpriseBilling():
		snap, err = c.enterpriseCreditSnapshot(a, soon)
	case a.IsGlobal():
		snap, err = c.globalCreditSnapshot(a, soon)
		if err != nil {
			log.Printf("WARN: billing %s: global summary 失败，回落个人资源包路径: %v",
				logfmt.Label(a.UID, a.Nickname), err)
			snap, err = c.cnCreditSnapshot(a, soon)
		}
	default:
		snap, err = c.cnCreditSnapshot(a, soon)
	}
	if notify {
		// 余额观察者只在成功时留痕；失败不是观测值（与 notifyCredits 的既有语义一致）。
		c.notifyCredits(a, snap.Remain, err)
	}
	return snap, err
}

// enterpriseCreditSnapshot 企业额度：周期末即作废，只有一条批次。
// 企业额度不在个人资源包体系内，故不产出逐包构成（保持面板既有行为）。
func (c *Client) enterpriseCreditSnapshot(a *auth.Auth, soon time.Duration) (CreditSnapshot, error) {
	remain, total, expiring, earliestAt, earliestRemaining, err := c.enterpriseResource(a, soon)
	if err != nil {
		return CreditSnapshot{}, err
	}
	snap := CreditSnapshot{
		Remain:            remain,
		Total:             total,
		Expiring:          expiring,
		EarliestAt:        earliestAt,
		EarliestRemaining: earliestRemaining,
	}
	if earliestRemaining > 0 && !earliestAt.IsZero() && earliestAt.After(time.Now()) {
		snap.Buckets = []ExpiryBucket{{
			Date:     earliestAt.In(softRateResetLoc).Format("2006-01-02"),
			ExpireAt: earliestAt,
			Remain:   earliestRemaining,
		}}
	}
	return snap, nil
}

// globalCreditSnapshot 国际版：summary 给权威 remain/total，paid/free 明细给逐包批次。
// 明细失败不影响总额：余下照常返回，只是没有批次/逐包构成。
func (c *Client) globalCreditSnapshot(a *auth.Auth, soon time.Duration) (CreditSnapshot, error) {
	raw, err := c.billingJSON(a, "POST", billingSummaryPath, map[string]any{})
	if err != nil {
		return CreditSnapshot{}, err
	}
	var sum summaryReply
	if uerr := json.Unmarshal(raw, &sum); uerr != nil {
		return CreditSnapshot{}, fmt.Errorf("billing summary parse: %w", uerr)
	}
	now := time.Now()
	var remain, total int64
	codes := make([]string, 0, len(sum.Packages))
	for _, p := range sum.Packages {
		r := int64(p.CycleRemainCapacity)
		s := int64(p.CycleTotalCapacity)
		if r < 0 {
			r = 0
		}
		if r > s && s > 0 {
			r = s
		}
		remain += r
		total += s
		if p.PackageCode != "" {
			codes = append(codes, p.PackageCode)
		}
	}

	rows, detailOK := c.globalPackageDetails(a, codes)
	snap := buildCreditSnapshot(remain, total, mergeGlobalGrants(rows), soon, now)
	if detailOK == 0 && len(codes) > 0 {
		snap.packagesErr = fmt.Errorf("套餐明细拉取失败")
	}
	return snap, nil
}

// cnCreditSnapshot 国内版：get-user-resource 一次解析出逐包构成与到期批次。
func (c *Client) cnCreditSnapshot(a *auth.Auth, soon time.Duration) (CreditSnapshot, error) {
	now := time.Now()
	body := map[string]any{
		"PageNumber":               1,
		"PageSize":                 100,
		"ProductCode":              "p_tcaca",
		"Status":                   []int{0, 3},
		"PackageEndTimeRangeBegin": now.Format(packageEndLayout),
		"PackageEndTimeRangeEnd":   now.Add(365 * 101 * 24 * time.Hour).Format(packageEndLayout),
	}
	// 余额查询做瞬时错误有界重试（签到后紧接着的 user-resource 偶发 500 会让
	// 该账号错过本次解冻/到期快照更新，只能等下一个刷新周期）。
	var data json.RawMessage
	err := c.retryBillingTransient(func() error {
		var e error
		data, e = c.billingMeterJSON(a, c.billingMeterPaths(a), http.MethodPost, body)
		return e
	})
	if err != nil {
		return CreditSnapshot{}, err
	}
	// 注意层级：doJSON 已经解过 apiEnvelope 并返回 env.Data，所以这里从
	// Response 开始解析——不能再套一层 Code/Data，否则 Accounts 恒为空。
	var resp struct {
		Response struct {
			Data struct {
				Accounts []creditAccountRow `json:"Accounts"`
			} `json:"Data"`
		} `json:"Response"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return CreditSnapshot{}, fmt.Errorf("resource parse: %w", err)
	}
	rows := resp.Response.Data.Accounts
	grants := make([]creditGrant, 0, len(rows))
	var remain, total int64
	for _, row := range rows {
		g := creditGrantFromRow(row)
		remain += g.Remain
		total += g.Size
		grants = append(grants, g)
	}
	return buildCreditSnapshot(remain, total, grants, soon, now), nil
}

// creditGrantFromRow 把上游单行规范化成 creditGrant。
//
// remain/used/size 统一走 packageRemainUsed（与 cmd/credit、历史 CreditPackages
// 同一口径）；到期时刻统一走 packageEndAt（DeductionEndTime 优先）。
func creditGrantFromRow(r creditAccountRow) creditGrant {
	remain, used, size := packageRemainUsed(respAccount{
		CapacityRemain:      r.CapacityRemain,
		CapacityUsed:        r.CapacityUsed,
		CapacitySize:        r.CapacitySize,
		CycleCapacityRemain: r.CycleCapacityRemain,
		CycleCapacityUsed:   r.CycleCapacityUsed,
		CycleCapacitySize:   r.CycleCapacitySize,
	})
	if remain < 0 {
		remain = 0
	}
	if used < 0 {
		used = 0
	}
	if size < remain {
		size = remain
	}
	end, hasEnd := packageEndAt(r.DeductionEndTime, r.ExpiredTime, r.CycleEndTime)
	created := time.Time{}
	if r.CreateTime > 0 {
		created = time.UnixMilli(r.CreateTime)
	}
	rid := strings.TrimSpace(r.ResourceID)
	if rid == "" {
		// 上游没给 ResourceId 时用「包码 + 扣费截止 + 发放时刻」兜底：同一包的
		// 多个计量周期通常共享这三者，不同日包至少发放时刻不同。
		rid = fmt.Sprintf("%s|%d|%d", r.PackageCode, r.DeductionEndTime, r.CreateTime)
	}
	return creditGrant{
		ResourceID:     rid,
		PackageCode:    r.PackageCode,
		Name:           r.PackageName,
		SubProductCode: r.SubProductCode,
		SubProductName: r.SubProductName,
		End:            end,
		HasEnd:         hasEnd,
		CreatedAt:      created,
		Remain:         remain,
		Used:           used,
		Size:           size,
		Cycle:          r.CycleCapacitySize > 0,
	}
}

// mergeGlobalGrants 按 ResourceId 归并 global 明细。
//
// 同一 ResourceId 的多个计量周期只保留一条：优先最早到期且仍有余额的周期（FEFO
// 真正会先扣完的那期），其次最早到期的周期，全部过期则保留最新一条供构成展示。
// 不同 ResourceId 即使 PackageCode 相同也各自保留——日赠包正是这种形态。
func mergeGlobalGrants(rows []creditAccountRow) []creditGrant {
	now := time.Now()
	groups := make(map[string][]creditGrant, len(rows))
	order := make([]string, 0, len(rows))
	for _, row := range rows {
		g := creditGrantFromRow(row)
		if _, seen := groups[g.ResourceID]; !seen {
			order = append(order, g.ResourceID)
		}
		groups[g.ResourceID] = append(groups[g.ResourceID], g)
	}
	out := make([]creditGrant, 0, len(groups))
	for _, key := range order {
		out = append(out, pickGlobalGrant(groups[key], now))
	}
	return out
}

func pickGlobalGrant(grants []creditGrant, now time.Time) creditGrant {
	sort.SliceStable(grants, func(i, j int) bool {
		if grants[i].HasEnd != grants[j].HasEnd {
			return grants[i].HasEnd
		}
		return grants[i].End.Before(grants[j].End)
	})
	var firstFuture, firstFuturePositive *creditGrant
	for i := range grants {
		g := &grants[i]
		if !g.HasEnd || !g.End.After(now) {
			continue
		}
		if firstFuture == nil {
			firstFuture = g
		}
		if g.Remain > 0 {
			firstFuturePositive = g
			break
		}
	}
	if firstFuturePositive != nil {
		return *firstFuturePositive
	}
	if firstFuture != nil {
		return *firstFuture
	}
	for _, g := range grants {
		if !g.HasEnd && g.Remain > 0 {
			return g
		}
	}
	return grants[len(grants)-1]
}

// buildCreditSnapshot 由同一批 grants 产出 totals/buckets/packages。
// grants 里没有有效未来到期时间或余额为 0 的包不进批次；逐包构成仍全部保留。
func buildCreditSnapshot(remain, total int64, grants []creditGrant, soon time.Duration, now time.Time) CreditSnapshot {
	snap := CreditSnapshot{Remain: remain, Total: total}
	byDate := make(map[string]*ExpiryBucket)
	for _, g := range grants {
		if !g.HasEnd || !g.End.After(now) || g.Remain <= 0 {
			continue
		}
		if snap.EarliestAt.IsZero() || g.End.Before(snap.EarliestAt) {
			snap.EarliestAt = g.End
		}
		day := g.End.In(softRateResetLoc).Format("2006-01-02")
		b, seen := byDate[day]
		if !seen {
			b = &ExpiryBucket{Date: day, ExpireAt: g.End}
			byDate[day] = b
		} else if g.End.Before(b.ExpireAt) {
			b.ExpireAt = g.End
		}
		b.Remain += g.Remain
		if soon > 0 && !g.End.After(now.Add(soon)) {
			snap.Expiring += g.Remain
		}
	}
	dates := make([]string, 0, len(byDate))
	for d := range byDate {
		dates = append(dates, d)
	}
	sort.Strings(dates)
	if len(dates) > 0 {
		snap.Buckets = make([]ExpiryBucket, 0, len(dates))
		for _, d := range dates {
			snap.Buckets = append(snap.Buckets, *byDate[d])
		}
		// earliestRemaining 是该自然日全部正余额包的合计（不是单个包的剩余量）。
		snap.EarliestRemaining = snap.Buckets[0].Remain
	}

	snap.Packages = make([]CreditPackage, 0, len(grants))
	for _, g := range grants {
		cp := CreditPackage{
			Name:           g.Name,
			Remain:         g.Remain,
			Used:           g.Used,
			Size:           g.Size,
			PackageCode:    g.PackageCode,
			SubProductCode: g.SubProductCode,
			SubProductName: g.SubProductName,
			Cycle:          g.Cycle,
		}
		if g.HasEnd {
			cp.EndTime = g.End.Format(packageEndLayout)
			cp.ExpiresAt = g.End.UnixMilli()
		}
		if !g.CreatedAt.IsZero() {
			cp.CreatedAt = g.CreatedAt.Format(time.RFC3339)
		}
		snap.Packages = append(snap.Packages, cp)
	}
	// 面额降序：大包一眼可见，正是差异最可能出现的地方。
	sort.SliceStable(snap.Packages, func(i, j int) bool {
		return snap.Packages[i].Size > snap.Packages[j].Size
	})
	return snap
}
