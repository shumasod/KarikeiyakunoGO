package analyzer_test

import (
	"math"
	"testing"

	"tse-scanner/analyzer"
	"tse-scanner/model"
)

// ---- helpers ----

func newQuote(changePercent float64, volRatio float64, price, dayHigh, weekHigh52 float64) model.Quote {
	avgVol := int64(1000000)
	vol := int64(float64(avgVol) * volRatio)
	return model.Quote{
		Symbol:        "TEST.T",
		Name:          "テスト株式",
		Sector:        "テスト",
		Price:         price,
		ChangePercent: changePercent,
		Volume:        vol,
		AvgVolume3M:   avgVol,
		DayHigh:       dayHigh,
		WeekHigh52:    weekHigh52,
		Valid:          true,
	}
}

func hasSignal(signals []model.Signal, label string) bool {
	for _, s := range signals {
		if s.Label == label {
			return true
		}
	}
	return false
}

func totalScore(signals []model.Signal) float64 {
	total := 0.0
	for _, s := range signals {
		total += s.Score
	}
	return total
}

// ---- Analyze function tests ----

func TestAnalyze_ExcludesInvalidQuotes(t *testing.T) {
	quotes := []model.Quote{
		{Symbol: "A.T", Valid: false, ChangePercent: 10},
		newQuote(5.0, 3.0, 1000, 1010, 1200),
	}
	results := analyzer.Analyze(quotes, 0)
	if len(results) != 1 {
		t.Errorf("want 1 candidate, got %d", len(results))
	}
	if results[0].Symbol != "TEST.T" {
		t.Errorf("unexpected symbol: %s", results[0].Symbol)
	}
}

func TestAnalyze_ExcludesBelowMinScore(t *testing.T) {
	quotes := []model.Quote{
		newQuote(1.0, 1.1, 1000, 1100, 2000), // low score
		newQuote(5.0, 4.0, 1000, 1000, 1010), // high score
	}
	results := analyzer.Analyze(quotes, 50.0)
	for _, r := range results {
		if r.SurgeScore < 50.0 {
			t.Errorf("candidate %s has score %f below minScore 50", r.Symbol, r.SurgeScore)
		}
	}
}

func TestAnalyze_SortedByScoreDescending(t *testing.T) {
	quotes := []model.Quote{
		newQuote(1.0, 1.5, 1000, 1100, 2000),
		newQuote(5.0, 4.0, 1000, 1000, 1010),
		newQuote(3.0, 2.5, 1000, 1050, 1500),
	}
	results := analyzer.Analyze(quotes, 0)
	for i := 1; i < len(results); i++ {
		if results[i].SurgeScore > results[i-1].SurgeScore {
			t.Errorf("results not sorted: index %d (%f) > index %d (%f)",
				i, results[i].SurgeScore, i-1, results[i-1].SurgeScore)
		}
	}
}

func TestAnalyze_EmptyInput(t *testing.T) {
	results := analyzer.Analyze([]model.Quote{}, 0)
	if results == nil {
		t.Error("want empty slice, got nil")
	}
	if len(results) != 0 {
		t.Errorf("want 0 results, got %d", len(results))
	}
}

func TestAnalyze_AllInvalid(t *testing.T) {
	quotes := []model.Quote{
		{Symbol: "A.T", Valid: false},
		{Symbol: "B.T", Valid: false},
	}
	results := analyzer.Analyze(quotes, 0)
	if len(results) != 0 {
		t.Errorf("want 0 results, got %d", len(results))
	}
}

// ---- Score component tests ----

func TestScore_PriceComponent_Zero_WhenNoChange(t *testing.T) {
	q := newQuote(0, 1.0, 1000, 1100, 2000)
	results := analyzer.Analyze([]model.Quote{q}, 0)
	if len(results) == 0 {
		return // score 0, filtered out is fine
	}
	// Price score should not be present for 0% change
	if hasSignal(results[0].Signals, "騰落率") {
		t.Error("should not have 騰落率 signal for 0% change")
	}
}

func TestScore_PriceComponent_FullAt5Percent(t *testing.T) {
	// +5% change with no other signals → price score = 40
	q := newQuote(5.0, 1.0, 1000, 1100, 2000)
	results := analyzer.Analyze([]model.Quote{q}, 0)
	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}
	found := false
	for _, s := range results[0].Signals {
		if s.Label == "騰落率" {
			if math.Abs(s.Score-40.0) > 0.01 {
				t.Errorf("want price score 40.0 at 5%%, got %f", s.Score)
			}
			found = true
		}
	}
	if !found {
		t.Error("want 騰落率 signal")
	}
}

func TestScore_PriceComponent_CappedAt40(t *testing.T) {
	// +20% change → price score still capped at 40
	q := newQuote(20.0, 1.0, 1000, 1100, 2000)
	results := analyzer.Analyze([]model.Quote{q}, 0)
	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}
	for _, s := range results[0].Signals {
		if s.Label == "騰落率" && s.Score > 40.01 {
			t.Errorf("price score should be capped at 40, got %f", s.Score)
		}
	}
}

func TestScore_PriceComponent_NegativeChange_NoSignal(t *testing.T) {
	q := newQuote(-3.0, 1.0, 1000, 1100, 2000)
	results := analyzer.Analyze([]model.Quote{q}, 0)
	for _, r := range results {
		if hasSignal(r.Signals, "騰落率") {
			t.Error("should not have 騰落率 signal for negative change")
		}
	}
}

func TestScore_VolumeComponent_Zero_WhenBelowAverage(t *testing.T) {
	// Volume ratio = 0.5 (below average) → no volume score
	q := newQuote(3.0, 0.5, 1000, 1100, 2000)
	results := analyzer.Analyze([]model.Quote{q}, 0)
	if len(results) > 0 {
		if hasSignal(results[0].Signals, "出来高") {
			t.Error("should not have 出来高 signal when volume ratio < 1")
		}
	}
}

func TestScore_VolumeComponent_FullAt4x(t *testing.T) {
	// Volume ratio = 4.0 → volume score = 30
	q := newQuote(0, 4.0, 1000, 1100, 2000)
	results := analyzer.Analyze([]model.Quote{q}, 0)
	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}
	for _, s := range results[0].Signals {
		if s.Label == "出来高" {
			if math.Abs(s.Score-30.0) > 0.01 {
				t.Errorf("want volume score 30.0 at 4x, got %f", s.Score)
			}
		}
	}
}

func TestScore_VolumeComponent_NoAvgVolume_NeutralRatio(t *testing.T) {
	// When AvgVolume3M = 0, volumeRatio should return 1.0 (neutral, no score)
	q := newQuote(3.0, 1.0, 1000, 1100, 2000)
	q.AvgVolume3M = 0
	q.Volume = 999999
	results := analyzer.Analyze([]model.Quote{q}, 0)
	if len(results) > 0 {
		if hasSignal(results[0].Signals, "出来高") {
			t.Error("should not have 出来高 signal when AvgVolume3M=0")
		}
	}
}

func TestScore_DayHighComponent_AddedWhenNear(t *testing.T) {
	// Price = DayHigh → proximity = 1.0 → high score = 20
	q := newQuote(0, 1.0, 1000, 1000, 2000)
	results := analyzer.Analyze([]model.Quote{q}, 0)
	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}
	if !hasSignal(results[0].Signals, "高値圏") {
		t.Error("want 高値圏 signal when price == day high")
	}
}

func TestScore_DayHighComponent_SkippedWhenFarFromHigh(t *testing.T) {
	// Price = 90% of DayHigh → proximity = 0.9 < threshold 0.98
	q := newQuote(0, 1.0, 900, 1000, 2000)
	results := analyzer.Analyze([]model.Quote{q}, 0)
	for _, r := range results {
		if hasSignal(r.Signals, "高値圏") {
			t.Error("should not have 高値圏 signal when price far from day high")
		}
	}
}

func TestScore_52WeekHigh_AddedWhenNear(t *testing.T) {
	// Price >= 99% of 52-week high
	q := newQuote(0, 1.0, 990, 1100, 1000)
	results := analyzer.Analyze([]model.Quote{q}, 0)
	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}
	if !hasSignal(results[0].Signals, "52週高値") {
		t.Error("want 52週高値 signal when price near 52-week high")
	}
}

func TestScore_52WeekNewHigh_Signal(t *testing.T) {
	// Price >= 52-week high → 🏆新高値
	q := newQuote(0, 1.0, 1100, 1200, 1000)
	results := analyzer.Analyze([]model.Quote{q}, 0)
	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}
	if !hasSignal(results[0].Signals, "🏆52週新高値") {
		t.Error("want 🏆52週新高値 signal when price exceeds 52-week high")
	}
}

func TestScore_52WeekNearHigh_Signal(t *testing.T) {
	// Price = 99.5% of 52-week high → 🎯高値圏
	q := newQuote(0, 1.0, 995, 1100, 1000)
	results := analyzer.Analyze([]model.Quote{q}, 0)
	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}
	if !hasSignal(results[0].Signals, "🎯52週高値圏") {
		t.Error("want 🎯52週高値圏 signal when price near (but below) 52-week high")
	}
}

// ---- Signal detection tests ----

func TestSignals_BigRise(t *testing.T) {
	q := newQuote(6.0, 1.0, 1000, 1100, 2000)
	results := analyzer.Analyze([]model.Quote{q}, 0)
	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}
	if !hasSignal(results[0].Signals, "🚀大幅上昇") {
		t.Error("want 🚀大幅上昇 signal for >= 5% change")
	}
}

func TestSignals_RiseTrend(t *testing.T) {
	q := newQuote(4.0, 1.0, 1000, 1100, 2000)
	results := analyzer.Analyze([]model.Quote{q}, 0)
	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}
	if !hasSignal(results[0].Signals, "📈上昇トレンド") {
		t.Error("want 📈上昇トレンド signal for 3-5% change")
	}
	if hasSignal(results[0].Signals, "🚀大幅上昇") {
		t.Error("should not have 🚀大幅上昇 for 4% change")
	}
}

func TestSignals_VolumeSpike(t *testing.T) {
	q := newQuote(0, 3.0, 1000, 1100, 2000)
	results := analyzer.Analyze([]model.Quote{q}, 0)
	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}
	if !hasSignal(results[0].Signals, "⚡出来高急増") {
		t.Error("want ⚡出来高急増 signal for >= 2x volume")
	}
}

func TestSignals_DayHighProximity(t *testing.T) {
	q := newQuote(0, 1.0, 999, 1000, 2000) // 99.9% of day high
	results := analyzer.Analyze([]model.Quote{q}, 0)
	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}
	if !hasSignal(results[0].Signals, "🔼高値圏推移") {
		t.Error("want 🔼高値圏推移 signal when near day high")
	}
}

// ---- Score bounds ----

func TestScore_MaxIs100(t *testing.T) {
	// Perfect score scenario: +20%, 10x volume, price == day high == 52w high
	q := newQuote(20.0, 10.0, 1000, 1000, 1000)
	results := analyzer.Analyze([]model.Quote{q}, 0)
	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}
	if results[0].SurgeScore > 100.0 {
		t.Errorf("score should be capped at 100, got %f", results[0].SurgeScore)
	}
}

func TestScore_ZeroWhenAllNeutral(t *testing.T) {
	q := model.Quote{
		Symbol:      "ZERO.T",
		Price:       1000,
		DayHigh:     2000, // far from high
		WeekHigh52:  3000, // far from 52w high
		AvgVolume3M: 1000000,
		Volume:      500000, // below average
		Valid:       true,
	}
	results := analyzer.Analyze([]model.Quote{q}, 0)
	for _, r := range results {
		if r.SurgeScore != 0 {
			t.Errorf("want score 0 for neutral quote, got %f", r.SurgeScore)
		}
	}
}

// ---- VolumeRatio field ----

func TestCandidate_VolumeRatioSet(t *testing.T) {
	q := newQuote(5.0, 3.0, 1000, 1050, 2000)
	results := analyzer.Analyze([]model.Quote{q}, 0)
	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}
	if math.Abs(results[0].VolumeRatio-3.0) > 0.01 {
		t.Errorf("want VolumeRatio 3.0, got %f", results[0].VolumeRatio)
	}
}

func TestCandidate_VolumeRatioNeutralWhenNoAvg(t *testing.T) {
	q := newQuote(5.0, 1.0, 1000, 1050, 2000)
	q.AvgVolume3M = 0
	results := analyzer.Analyze([]model.Quote{q}, 0)
	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}
	if results[0].VolumeRatio != 1.0 {
		t.Errorf("want VolumeRatio 1.0 when AvgVolume3M=0, got %f", results[0].VolumeRatio)
	}
}

// ---- Boundary tests ----

func TestScore_Exactly5PercentChange(t *testing.T) {
	q := newQuote(5.0, 1.0, 1000, 1100, 2000)
	results := analyzer.Analyze([]model.Quote{q}, 0)
	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}
	// At exactly 5% we expect 🚀大幅上昇 (>= 5.0)
	if !hasSignal(results[0].Signals, "🚀大幅上昇") {
		t.Error("want 🚀大幅上昇 at exactly 5% change")
	}
}

func TestScore_Exactly3PercentChange(t *testing.T) {
	q := newQuote(3.0, 1.0, 1000, 1100, 2000)
	results := analyzer.Analyze([]model.Quote{q}, 0)
	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}
	if !hasSignal(results[0].Signals, "📈上昇トレンド") {
		t.Error("want 📈上昇トレンド at exactly 3% change")
	}
}

func TestScore_Exactly2xVolume(t *testing.T) {
	q := newQuote(0, 2.0, 1000, 1100, 2000)
	results := analyzer.Analyze([]model.Quote{q}, 0)
	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}
	if !hasSignal(results[0].Signals, "⚡出来高急増") {
		t.Error("want ⚡出来高急増 at exactly 2x volume")
	}
}

func TestScore_DayHighZero_SkipsComponent(t *testing.T) {
	q := newQuote(5.0, 3.0, 1000, 0, 0)
	results := analyzer.Analyze([]model.Quote{q}, 0)
	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}
	if hasSignal(results[0].Signals, "高値圏") {
		t.Error("should skip 高値圏 component when DayHigh=0")
	}
	if hasSignal(results[0].Signals, "52週高値") {
		t.Error("should skip 52週高値 component when WeekHigh52=0")
	}
}

func TestScore_PriceZero_SkipsDayHighComponent(t *testing.T) {
	q := model.Quote{
		Symbol:      "Z.T",
		Price:       0,
		DayHigh:     1000,
		WeekHigh52:  1000,
		AvgVolume3M: 1000000,
		Volume:      3000000,
		Valid:       true,
	}
	results := analyzer.Analyze([]model.Quote{q}, 0)
	for _, r := range results {
		if hasSignal(r.Signals, "高値圏") {
			t.Error("should skip 高値圏 when Price=0")
		}
	}
}
