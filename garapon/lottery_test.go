package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// テスト間でグローバル状態をリセットするヘルパー
func resetState() {
	historyMu.Lock()
	history = nil
	ticketCount = 0
	historyMu.Unlock()

	// 景品の重みを初期値に戻す
	prizeMu.Lock()
	prizes[0].Weight = 5
	prizes[1].Weight = 30
	prizes[2].Weight = 75
	prizes[3].Weight = 190
	prizes[4].Weight = 200
	prizes[5].Weight = 500
	lastRotatedAt = time.Time{}
	nextRotateAt = time.Time{}
	prizeMu.Unlock()
}

// ==============================
// 抽選ロジックのテスト
// ==============================

func TestDrawLottery_ReturnsValidPrize(t *testing.T) {
	resetState()
	result := drawLottery()

	validGrades := map[PrizeGrade]bool{
		GradeTokutou: true, GradeIttou: true, GradeNittou: true,
		GradeSantou: true, GradeYontou: true, GradeHazure: true,
	}
	if !validGrades[result.Prize.Grade] {
		t.Errorf("無効な等級が返された: %q", result.Prize.Grade)
	}
}

func TestDrawLottery_IncrementTicketNum(t *testing.T) {
	resetState()
	r1 := drawLottery()
	r2 := drawLottery()
	r3 := drawLottery()

	if r1.TicketNum != 1 {
		t.Errorf("1回目のチケット番号: got %d, want 1", r1.TicketNum)
	}
	if r2.TicketNum != 2 {
		t.Errorf("2回目のチケット番号: got %d, want 2", r2.TicketNum)
	}
	if r3.TicketNum != 3 {
		t.Errorf("3回目のチケット番号: got %d, want 3", r3.TicketNum)
	}
}

func TestDrawLottery_AddsToHistory(t *testing.T) {
	resetState()
	drawLottery()
	drawLottery()

	historyMu.Lock()
	count := len(history)
	historyMu.Unlock()

	if count != 2 {
		t.Errorf("履歴件数: got %d, want 2", count)
	}
}

func TestDrawLottery_HistoryIsRecentFirst(t *testing.T) {
	resetState()
	drawLottery() // ticket #1
	drawLottery() // ticket #2

	historyMu.Lock()
	first := history[0].TicketNum
	historyMu.Unlock()

	if first != 2 {
		t.Errorf("先頭履歴のチケット番号: got %d, want 2（最新）", first)
	}
}

func TestDrawLottery_HistoryMaxSize(t *testing.T) {
	resetState()
	for i := 0; i < 60; i++ {
		drawLottery()
	}
	historyMu.Lock()
	count := len(history)
	historyMu.Unlock()

	if count > 50 {
		t.Errorf("履歴件数が上限を超えた: got %d, want <= 50", count)
	}
}

func TestDrawLottery_HasTimestamp(t *testing.T) {
	resetState()
	result := drawLottery()
	if result.DrawnAt.IsZero() {
		t.Error("DrawnAt がゼロ値になっている")
	}
}

func TestDrawLottery_ConcurrentSafe(t *testing.T) {
	resetState()
	const n = 100
	var wg sync.WaitGroup
	results := make([]DrawResult, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = drawLottery()
		}(i)
	}
	wg.Wait()

	seen := make(map[int]bool)
	for _, r := range results {
		if seen[r.TicketNum] {
			t.Errorf("チケット番号が重複している: %d", r.TicketNum)
		}
		seen[r.TicketNum] = true
	}
}

// ==============================
// 景品テーブルのテスト
// ==============================

func TestPrizesWeightSum(t *testing.T) {
	resetState()
	prizeMu.RLock()
	total := 0
	for _, p := range prizes {
		total += p.Weight
	}
	prizeMu.RUnlock()

	if total != 1000 {
		t.Errorf("景品の重み合計: got %d, want 1000", total)
	}
}

func TestPrizesHaveRequiredFields(t *testing.T) {
	prizeMu.RLock()
	defer prizeMu.RUnlock()

	for _, p := range prizes {
		if p.Grade == "" {
			t.Errorf("Gradeが空の景品: %+v", p)
		}
		if p.Name == "" {
			t.Errorf("Nameが空の景品: %+v", p)
		}
		if p.Description == "" {
			t.Errorf("Descriptionが空の景品: %+v", p)
		}
		if p.Ball.Hex == "" {
			t.Errorf("Ball.Hexが空の景品: %+v", p)
		}
		if p.Ball.Name == "" {
			t.Errorf("Ball.Nameが空の景品: %+v", p)
		}
		if p.Weight <= 0 {
			t.Errorf("Weightが0以下の景品: %+v", p)
		}
	}
}

func TestPrizesGradesAreUnique(t *testing.T) {
	prizeMu.RLock()
	defer prizeMu.RUnlock()

	seen := make(map[PrizeGrade]bool)
	for _, p := range prizes {
		if seen[p.Grade] {
			t.Errorf("等級が重複している: %s", p.Grade)
		}
		seen[p.Grade] = true
	}
}

// ==============================
// 重みローテーションのテスト
// ==============================

// 生成された重みの合計が必ず 1000 であることを確認
func TestGenerateNewWeights_SumIs1000(t *testing.T) {
	for i := 0; i < 200; i++ {
		weights := generateNewWeights()
		total := 0
		for _, w := range weights {
			total += w
		}
		if total != 1000 {
			t.Errorf("重み合計: got %d, want 1000 (試行 %d)", total, i)
		}
	}
}

// 生成された重みがすべて正の値であることを確認
func TestGenerateNewWeights_AllPositive(t *testing.T) {
	for i := 0; i < 200; i++ {
		weights := generateNewWeights()
		for j, w := range weights {
			if w <= 0 {
				t.Errorf("重み[%d]が0以下: %d (試行 %d)", j, w, i)
			}
		}
	}
}

// 生成された重みの個数が景品数と一致することを確認
func TestGenerateNewWeights_CountMatchesPrizes(t *testing.T) {
	weights := generateNewWeights()
	if len(weights) != len(prizes) {
		t.Errorf("重みの個数: got %d, want %d", len(weights), len(prizes))
	}
}

// rotatePrizes 後も重みの合計が 1000 であることを確認
func TestRotatePrizes_SumStays1000(t *testing.T) {
	resetState()
	rotatePrizes()

	prizeMu.RLock()
	total := 0
	for _, p := range prizes {
		total += p.Weight
	}
	prizeMu.RUnlock()

	if total != 1000 {
		t.Errorf("ローテーション後の重み合計: got %d, want 1000", total)
	}
}

// rotatePrizes が lastRotatedAt を更新することを確認
func TestRotatePrizes_UpdatesLastRotatedAt(t *testing.T) {
	resetState()
	before := time.Now()
	rotatePrizes()

	prizeMu.RLock()
	lra := lastRotatedAt
	prizeMu.RUnlock()

	if lra.Before(before) {
		t.Errorf("lastRotatedAt が更新されていない: got %v, want >= %v", lra, before)
	}
}

// rotatePrizes が nextRotateAt を rotationInterval 後に設定することを確認
func TestRotatePrizes_UpdatesNextRotateAt(t *testing.T) {
	resetState()
	before := time.Now()
	rotatePrizes()

	prizeMu.RLock()
	nra := nextRotateAt
	lra := lastRotatedAt
	prizeMu.RUnlock()

	expectedMin := lra.Add(rotationInterval - time.Second)
	expectedMax := before.Add(rotationInterval + time.Second)

	if nra.Before(expectedMin) || nra.After(expectedMax) {
		t.Errorf("nextRotateAt が期待範囲外: got %v", nra)
	}
}

// rotatePrizes を複数回呼んでも重みの合計が常に 1000 であることを確認
func TestRotatePrizes_MultipleRotations(t *testing.T) {
	resetState()
	for i := 0; i < 50; i++ {
		rotatePrizes()

		prizeMu.RLock()
		total := 0
		for _, p := range prizes {
			total += p.Weight
		}
		prizeMu.RUnlock()

		if total != 1000 {
			t.Errorf("第%d回ローテーション後の重み合計: got %d, want 1000", i+1, total)
		}
	}
}

// ==============================
// 統計のテスト
// ==============================

func TestCalcStats_EmptyHistory(t *testing.T) {
	resetState()
	stats := calcStats()

	if stats.TotalDraws != 0 {
		t.Errorf("空履歴のTotalDraws: got %d, want 0", stats.TotalDraws)
	}
	if len(stats.GradeCount) != 0 {
		t.Errorf("空履歴のGradeCount: got %v, want empty", stats.GradeCount)
	}
}

func TestCalcStats_TotalDraws(t *testing.T) {
	resetState()
	drawLottery()
	drawLottery()
	drawLottery()

	stats := calcStats()
	if stats.TotalDraws != 3 {
		t.Errorf("TotalDraws: got %d, want 3", stats.TotalDraws)
	}
}

func TestCalcStats_GradeCountSumEqualsTotalDraws(t *testing.T) {
	resetState()
	for i := 0; i < 10; i++ {
		drawLottery()
	}
	stats := calcStats()
	gradeTotal := 0
	for _, cnt := range stats.GradeCount {
		gradeTotal += cnt
	}
	if gradeTotal != stats.TotalDraws {
		t.Errorf("GradeCount合計(%d) != TotalDraws(%d)", gradeTotal, stats.TotalDraws)
	}
}

func TestCalcStats_HasLastUpdated(t *testing.T) {
	resetState()
	stats := calcStats()
	if stats.LastUpdated.IsZero() {
		t.Error("LastUpdated がゼロ値になっている")
	}
}

// ==============================
// HTTP ハンドラのテスト
// ==============================

func TestApiDrawHandler_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/draw", nil)
	w := httptest.NewRecorder()
	apiDrawHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("GETリクエストのステータス: got %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestApiDrawHandler_PostReturnsDrawResult(t *testing.T) {
	resetState()
	req := httptest.NewRequest(http.MethodPost, "/api/draw", nil)
	w := httptest.NewRecorder()
	apiDrawHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("ステータス: got %d, want %d", w.Code, http.StatusOK)
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type: got %q, want application/json", ct)
	}
	var result DrawResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("JSONパースエラー: %v", err)
	}
	if result.TicketNum != 1 {
		t.Errorf("TicketNum: got %d, want 1", result.TicketNum)
	}
	if result.Prize.Grade == "" {
		t.Error("Prize.Grade が空")
	}
}

func TestApiHistoryHandler_ReturnsHistory(t *testing.T) {
	resetState()
	drawLottery()
	drawLottery()

	req := httptest.NewRequest(http.MethodGet, "/api/history", nil)
	w := httptest.NewRecorder()
	apiHistoryHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("ステータス: got %d, want %d", w.Code, http.StatusOK)
	}
	var hist []DrawResult
	if err := json.Unmarshal(w.Body.Bytes(), &hist); err != nil {
		t.Fatalf("JSONパースエラー: %v", err)
	}
	if len(hist) != 2 {
		t.Errorf("履歴件数: got %d, want 2", len(hist))
	}
}

func TestApiStatsHandler_ReturnsStats(t *testing.T) {
	resetState()
	drawLottery()
	drawLottery()

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	w := httptest.NewRecorder()
	apiStatsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("ステータス: got %d, want %d", w.Code, http.StatusOK)
	}
	var stats Stats
	if err := json.Unmarshal(w.Body.Bytes(), &stats); err != nil {
		t.Fatalf("JSONパースエラー: %v", err)
	}
	if stats.TotalDraws != 2 {
		t.Errorf("TotalDraws: got %d, want 2", stats.TotalDraws)
	}
}

// /api/prizes は PrizesInfo 形式（prizes 配列 + ローテーション情報）を返すことを確認
func TestApiPrizesHandler_ReturnsPrizesInfo(t *testing.T) {
	resetState()
	nextRotateAt = time.Now().Add(rotationInterval)

	req := httptest.NewRequest(http.MethodGet, "/api/prizes", nil)
	w := httptest.NewRecorder()
	apiPrizesHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("ステータス: got %d, want %d", w.Code, http.StatusOK)
	}
	var info PrizesInfo
	if err := json.Unmarshal(w.Body.Bytes(), &info); err != nil {
		t.Fatalf("JSONパースエラー: %v", err)
	}
	if len(info.Prizes) != len(prizes) {
		t.Errorf("景品数: got %d, want %d", len(info.Prizes), len(prizes))
	}
	if info.RotationIntervalSec != int(rotationInterval.Seconds()) {
		t.Errorf("RotationIntervalSec: got %d, want %d", info.RotationIntervalSec, int(rotationInterval.Seconds()))
	}
	if info.NextRotationAt.IsZero() {
		t.Error("NextRotationAt がゼロ値")
	}
}

// /api/prizes が返す prizes の重み合計が 1000 であることを確認
func TestApiPrizesHandler_WeightSumIs1000(t *testing.T) {
	resetState()
	req := httptest.NewRequest(http.MethodGet, "/api/prizes", nil)
	w := httptest.NewRecorder()
	apiPrizesHandler(w, req)

	var info PrizesInfo
	if err := json.Unmarshal(w.Body.Bytes(), &info); err != nil {
		t.Fatalf("JSONパースエラー: %v", err)
	}
	total := 0
	for _, p := range info.Prizes {
		total += p.Weight
	}
	if total != 1000 {
		t.Errorf("レスポンス内の重み合計: got %d, want 1000", total)
	}
}

func TestHomeHandler_ReturnsHTML(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	homeHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("ステータス: got %d, want %d", w.Code, http.StatusOK)
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type: got %q, want text/html", ct)
	}
	body := w.Body.String()
	if !strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("レスポンスに <!DOCTYPE html> が含まれていない")
	}
	if !strings.Contains(body, "ガラガラポン") {
		t.Error("レスポンスに 'ガラガラポン' が含まれていない")
	}
	if !strings.Contains(body, "countdown") {
		t.Error("レスポンスにカウントダウン要素が含まれていない")
	}
}
