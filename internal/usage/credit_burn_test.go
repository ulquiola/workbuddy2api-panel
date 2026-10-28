package usage

import (
	"testing"
	"time"
)

// 近 7 天按账号聚合只统计窗口内、且确实拿到 credit 样本的请求：
// 观测到 0 分与「没测到」必须区分——前者是已知的 0 消耗，后者不该被当成 0。
func TestCreditByAccountWindow(t *testing.T) {
	r := New("")
	now := time.Now()
	r.Add(now, "global", "u1", "m", Delta{Credit: 10, HasCredit: true}, true)
	r.Add(now, "global", "u1", "m", Delta{Credit: 0, HasCredit: true}, true)
	r.Add(now, "cn", "u2", "m", Delta{}, true) // 没有 credit 样本
	r.Add(now.Add(-8*24*time.Hour), "global", "u1", "m", Delta{Credit: 99, HasCredit: true}, true)

	burns := r.CreditByAccountWindow(Window{Hours: 7 * 24})
	u1, ok := burns["u1"]
	if !ok {
		t.Fatal("u1 应在窗口内（10 分、2 个样本）")
	}
	if u1.Credits != 10 || u1.Samples != 2 {
		t.Fatalf("u1=%+v want credits=10 samples=2（8 天前的 99 分不得计入）", u1)
	}
	if _, ok := burns["u2"]; ok {
		t.Fatal("没有 credit 样本的账号不该出现在 map 里")
	}
}
