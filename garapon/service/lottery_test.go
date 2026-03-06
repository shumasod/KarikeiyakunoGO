package service

import (
	"math"
	"sync"
	"testing"
	"time"

	"garapon/model"
)

// helper: cast interface to concrete type for white-box testing
func asImpl(svc LotteryService) *lotteryService {
	return svc.(*lotteryService)
}

// ============================================================
// 初期化
// ============================================================

func TestNew_HasInitialPrizes(t *testing.T) {
	svc := NewWithoutRotation()
	info := svc.Prizes()
	if len(info.Prizes) != len(initialPrizes) {
		t.Errorf("景品数: got %d, want %d", len(info.Prizes), len(initialPrizes))
	}
}

func TestNew_InitialWeightSum(t *testing.T) {
	svc := NewWithoutRotation()
	info := svc.Prizes()
	total := 0
	for _, p := range info.Prizes {
		total += p.Weight
	}
	if total != 1000 {
		t.Errorf("初期重み合計: got %d, want 1000", total)
	}
}

func TestNew_WithRotation_StartsRotationTimer(t *testing.T) {
	svc := New(30 * time.Second)
	impl := asImpl(svc)
	if impl.nextRotateAt.IsZero() {
		t.Error("nextRotateAt がゼロ値のまま（タイマー未設定）")
	}
}

// ============================================================
// Draw — 正常系
// ============================================================

func TestDraw_ReturnsValidGrade(t *testing.T) {
	svc := NewWithoutRotation()
	validGrades := map[model.PrizeGrade]bool{
		model.GradeTokutou: true, model.GradeIttou: true, model.GradeNittou: true,
		model.GradeSantou: true, model.GradeYontou: true, model.GradeHazure: true,
	}
	for i := 0; i < 100; i++ {
		r, err := svc.Draw()
		if err != nil {
			t.Fatalf("Draw error: %v", err)
		}
		if !validGrades[r.Prize.Grade] {
			t.Errorf("無効な等級: %q", r.Prize.Grade)
		}
	}
}

func TestDraw_TicketNumIsSequential(t *testing.T) {
	svc := NewWithoutRotation()
	for i := 1; i <= 5; i++ {
		r, err := svc.Draw()
		if err != nil {
			t.Fatalf("Draw error: %v", err)
		}
		if r.TicketNum != i {
			t.Errorf("チケット番号: got %d, want %d", r.TicketNum, i)
		}
	}
}

func TestDraw_HasTimestamp(t *testing.T) {
	svc := NewWithoutRotation()
	before := time.Now()
	r, _ := svc.Draw()
	after := time.Now()
	if r.DrawnAt.Before(before) || r.DrawnAt.After(after) {
		t.Errorf("DrawnAt が範囲外: %v", r.DrawnAt)
	}
}

func TestDraw_AddsToHistory(t *testing.T) {
	svc := NewWithoutRotation()
	svc.Draw()
	svc.Draw()
	h := svc.History()
	if len(h) != 2 {
		t.Errorf("履歴件数: got %d, want 2", len(h))
	}
}

func TestDraw_HistoryIsRecentFirst(t *testing.T) {
	svc := NewWithoutRotation()
	svc.Draw() // ticket #1
	svc.Draw() // ticket #2
	h := svc.History()
	if h[0].TicketNum != 2 {
		t.Errorf("先頭のチケット番号: got %d, want 2", h[0].TicketNum)
	}
}

// ============================================================
// Draw — 境界値分析
// ============================================================

// 履歴の上限（50件）を超えないことを確認
func TestDraw_HistoryBoundary_MaxSize(t *testing.T) {
	svc := NewWithoutRotation()
	for i := 0; i < maxHistory+10; i++ {
		svc.Draw()
	}
	h := svc.History()
	if len(h) > maxHistory {
		t.Errorf("履歴件数 %d が上限 %d を超えた", len(h), maxHistory)
	}
	if len(h) != maxHistory {
		t.Errorf("履歴件数: got %d, want %d", len(h), maxHistory)
	}
}

// ちょうど上限件数（50件）は保持されることを確認
func TestDraw_HistoryBoundary_ExactMaxSize(t *testing.T) {
	svc := NewWithoutRotation()
	for i := 0; i < maxHistory; i++ {
		svc.Draw()
	}
	h := svc.History()
	if len(h) != maxHistory {
		t.Errorf("履歴件数: got %d, want %d", len(h), maxHistory)
	}
}

// ============================================================
// Draw — 異常系
// ============================================================

// 重みがすべて0の場合はエラーを返すことを確認
func TestDraw_ZeroWeights_ReturnsError(t *testing.T) {
	svc := NewWithoutRotation()
	impl := asImpl(svc)
	impl.prizeMu.Lock()
	for i := range impl.prizes {
		impl.prizes[i].Weight = 0
	}
	impl.prizeMu.Unlock()

	_, err := svc.Draw()
	if err == nil {
		t.Error("重み合計0のときエラーが返されなかった")
	}
}

// ============================================================
// Draw — 同時実行の整合性（競合チェック）
// ============================================================

func TestDraw_Concurrency_NoTicketDuplication(t *testing.T) {
	svc := NewWithoutRotation()
	const n = 200
	results := make([]model.DrawResult, n)
	var wg sync.WaitGroup

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			r, err := svc.Draw()
			if err != nil {
				t.Errorf("goroutine %d: Draw error: %v", idx, err)
				return
			}
			results[idx] = r
		}(i)
	}
	wg.Wait()

	seen := make(map[int]bool, n)
	for _, r := range results {
		if seen[r.TicketNum] {
			t.Errorf("チケット番号の重複: %d", r.TicketNum)
		}
		seen[r.TicketNum] = true
	}
}

// 並列 Draw 中に rotate() を呼んでもデータ競合が起きないことを確認
func TestDraw_Concurrency_WithConcurrentRotation(t *testing.T) {
	svc := NewWithoutRotation()
	impl := asImpl(svc)
	var wg sync.WaitGroup

	// 100 draws と 20 rotations を同時実行
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			svc.Draw() //nolint
		}()
	}
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			impl.rotate()
		}()
	}
	wg.Wait()

	// 整合性チェック: 重み合計は 1000 のまま
	info := svc.Prizes()
	total := 0
	for _, p := range info.Prizes {
		total += p.Weight
	}
	if total != 1000 {
		t.Errorf("並列実行後の重み合計: got %d, want 1000", total)
	}
}

// ============================================================
// History
// ============================================================

func TestHistory_ReturnsEmptySlice_WhenNoDraws(t *testing.T) {
	svc := NewWithoutRotation()
	h := svc.History()
	if h == nil {
		t.Error("履歴が nil（空スライスを期待）")
	}
	if len(h) != 0 {
		t.Errorf("履歴件数: got %d, want 0", len(h))
	}
}

// 返された履歴を変更しても内部状態に影響しないことを確認
func TestHistory_ReturnsCopy(t *testing.T) {
	svc := NewWithoutRotation()
	svc.Draw()
	h1 := svc.History()
	h1[0].TicketNum = 9999 // 外部から変更

	h2 := svc.History()
	if h2[0].TicketNum == 9999 {
		t.Error("History() が内部スライスへの参照を返している（コピーではない）")
	}
}

// ============================================================
// Stats
// ============================================================

func TestStats_Empty(t *testing.T) {
	svc := NewWithoutRotation()
	s := svc.Stats()
	if s.TotalDraws != 0 {
		t.Errorf("TotalDraws: got %d, want 0", s.TotalDraws)
	}
	if len(s.GradeCount) != 0 {
		t.Errorf("GradeCount: got %v, want empty", s.GradeCount)
	}
}

func TestStats_TotalDrawsIsCorrect(t *testing.T) {
	svc := NewWithoutRotation()
	for i := 0; i < 7; i++ {
		svc.Draw()
	}
	s := svc.Stats()
	if s.TotalDraws != 7 {
		t.Errorf("TotalDraws: got %d, want 7", s.TotalDraws)
	}
}

func TestStats_GradeCountSumEqualsTotalDraws(t *testing.T) {
	svc := NewWithoutRotation()
	for i := 0; i < 20; i++ {
		svc.Draw()
	}
	s := svc.Stats()
	gradeSum := 0
	for _, cnt := range s.GradeCount {
		gradeSum += cnt
	}
	if gradeSum != s.TotalDraws {
		t.Errorf("GradeCount合計(%d) != TotalDraws(%d)", gradeSum, s.TotalDraws)
	}
}

func TestStats_HasLastUpdated(t *testing.T) {
	svc := NewWithoutRotation()
	before := time.Now()
	s := svc.Stats()
	if s.LastUpdated.Before(before) {
		t.Errorf("LastUpdated が古すぎる: %v", s.LastUpdated)
	}
}

// ============================================================
// generateWeights — 正常系・境界値
// ============================================================

func TestGenerateWeights_SumIs1000(t *testing.T) {
	for i := 0; i < 1000; i++ {
		w := generateWeights(len(initialPrizes))
		total := 0
		for _, v := range w {
			total += v
		}
		if total != 1000 {
			t.Errorf("重み合計: got %d, want 1000 (試行 %d)", total, i)
		}
	}
}

func TestGenerateWeights_AllPositive(t *testing.T) {
	for i := 0; i < 1000; i++ {
		w := generateWeights(len(initialPrizes))
		for j, v := range w {
			if v <= 0 {
				t.Errorf("重み[%d]が0以下: %d (試行 %d)", j, v, i)
			}
		}
	}
}

// 参加賞（末尾）の重みが数学的な境界内に収まることを確認（境界値分析）
func TestGenerateWeights_HazureWeightWithinBounds(t *testing.T) {
	// Max sum of non-hazure: 15+60+120+250+300 = 745  → hazure min = 255
	// Min sum of non-hazure: 1+10+30+80+100   = 221  → hazure max = 779
	const hazureMin = 255
	const hazureMax = 779

	for i := 0; i < 1000; i++ {
		w := generateWeights(len(initialPrizes))
		hazure := w[len(w)-1]
		if hazure < hazureMin || hazure > hazureMax {
			t.Errorf("参加賞の重みが境界外: got %d, want [%d, %d] (試行 %d)",
				hazure, hazureMin, hazureMax, i)
		}
	}
}

// ============================================================
// rotate — 正常系・境界値・多重実行
// ============================================================

func TestRotate_WeightSumStays1000(t *testing.T) {
	svc := NewWithoutRotation()
	impl := asImpl(svc)
	impl.rotate()

	info := svc.Prizes()
	total := 0
	for _, p := range info.Prizes {
		total += p.Weight
	}
	if total != 1000 {
		t.Errorf("ローテーション後の重み合計: got %d, want 1000", total)
	}
}

func TestRotate_AllWeightsPositive(t *testing.T) {
	svc := NewWithoutRotation()
	impl := asImpl(svc)
	for i := 0; i < 100; i++ {
		impl.rotate()
		info := svc.Prizes()
		for _, p := range info.Prizes {
			if p.Weight <= 0 {
				t.Errorf("第%d回ローテーション後に重みが0以下: %s=%d", i+1, p.Grade, p.Weight)
			}
		}
	}
}

func TestRotate_UpdatesLastRotatedAt(t *testing.T) {
	svc := NewWithoutRotation()
	impl := asImpl(svc)
	before := time.Now()
	impl.rotate()
	info := svc.Prizes()
	if info.LastRotatedAt.Before(before) {
		t.Errorf("lastRotatedAt が更新されていない: %v", info.LastRotatedAt)
	}
}

func TestRotate_UpdatesNextRotateAt(t *testing.T) {
	svc := New(10 * time.Second)
	impl := asImpl(svc)
	impl.rotate()
	info := svc.Prizes()
	expectedAfter := info.LastRotatedAt.Add(9 * time.Second) // ±1s margin
	if info.NextRotationAt.Before(expectedAfter) {
		t.Errorf("nextRotateAt が期待より早い: %v", info.NextRotationAt)
	}
}

func TestRotate_50Times_Invariant(t *testing.T) {
	svc := NewWithoutRotation()
	impl := asImpl(svc)
	for i := 0; i < 50; i++ {
		impl.rotate()
		info := svc.Prizes()
		total := 0
		for _, p := range info.Prizes {
			total += p.Weight
		}
		if total != 1000 {
			t.Errorf("第%d回ローテーション後の重み合計: got %d, want 1000", i+1, total)
		}
	}
}

// ============================================================
// Prizes
// ============================================================

func TestPrizes_ReturnsCopy(t *testing.T) {
	svc := NewWithoutRotation()
	info1 := svc.Prizes()
	info1.Prizes[0].Weight = 9999 // 外部変更

	info2 := svc.Prizes()
	if info2.Prizes[0].Weight == 9999 {
		t.Error("Prizes() が内部スライスへの参照を返している")
	}
}

func TestPrizes_WeightSumIs1000(t *testing.T) {
	svc := NewWithoutRotation()
	info := svc.Prizes()
	total := 0
	for _, p := range info.Prizes {
		total += p.Weight
	}
	if total != 1000 {
		t.Errorf("返された重み合計: got %d, want 1000", total)
	}
}

// ============================================================
// 確率更新ロジックの統計的妥当性（カイ二乗検定相当）
// ============================================================

// 10,000 回抽選して各等級の出現頻度が期待値 ±4σ 内に収まることを確認
func TestDraw_StatisticalDistribution(t *testing.T) {
	svc := NewWithoutRotation()
	const n = 10_000

	counts := make(map[model.PrizeGrade]int)
	for i := 0; i < n; i++ {
		r, err := svc.Draw()
		if err != nil {
			t.Fatalf("Draw error: %v", err)
		}
		counts[r.Prize.Grade]++
	}

	// expected weights from initialPrizes
	expected := map[model.PrizeGrade]int{
		model.GradeTokutou: 5,
		model.GradeIttou:   30,
		model.GradeNittou:  75,
		model.GradeSantou:  190,
		model.GradeYontou:  200,
		model.GradeHazure:  500,
	}

	for grade, weight := range expected {
		p := float64(weight) / 1000.0
		expectedN := float64(n) * p
		sigma := math.Sqrt(float64(n) * p * (1 - p))
		tolerance := 4 * sigma // 4σ: 理論的な偽陽性率 0.0063%

		got := float64(counts[grade])
		if got < expectedN-tolerance || got > expectedN+tolerance {
			t.Errorf("%s: 出現回数 %.0f が期待範囲 [%.0f, %.0f] 外（期待値 %.0f, σ %.2f）",
				grade, got, expectedN-tolerance, expectedN+tolerance, expectedN, sigma)
		}
	}
}

// rotate() 後の確率が境界値を満たすことを統計的に確認
func TestRotate_StatisticalBounds(t *testing.T) {
	svc := NewWithoutRotation()
	impl := asImpl(svc)

	// 200 回ローテーションして毎回重みが境界内か確認
	for i := 0; i < 200; i++ {
		impl.rotate()
		info := svc.Prizes()
		for _, p := range info.Prizes {
			if p.Weight <= 0 {
				t.Errorf("第%d回後: %s の重みが0以下 (%d)", i+1, p.Grade, p.Weight)
			}
			if p.Weight > 1000 {
				t.Errorf("第%d回後: %s の重みが1000超 (%d)", i+1, p.Grade, p.Weight)
			}
		}
	}
}
