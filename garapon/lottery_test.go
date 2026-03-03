package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// テスト間でグローバル状態をリセットするヘルパー
func resetState() {
	historyMu.Lock()
	defer historyMu.Unlock()
	history = nil
	ticketCount = 0
}

// ==============================
// 抽選ロジックのテスト
// ==============================

// 抽選結果が有効な景品の等級であることを確認
func TestDrawLottery_ReturnsValidPrize(t *testing.T) {
	resetState()
	result := drawLottery()

	validGrades := map[PrizeGrade]bool{
		GradeTokutou: true,
		GradeIttou:   true,
		GradeNittou:  true,
		GradeSantou:  true,
		GradeYontou:  true,
		GradeHazure:  true,
	}

	if !validGrades[result.Prize.Grade] {
		t.Errorf("無効な等級が返された: %q", result.Prize.Grade)
	}
}

// 抽選のたびにチケット番号が1ずつ増えることを確認
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

// 抽選結果が履歴に正しく追加されることを確認
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

// 履歴は最新のものが先頭になることを確認
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

// 履歴の上限（50件）を超えないことを確認
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

// 抽選結果のTimestampがゼロ値でないことを確認
func TestDrawLottery_HasTimestamp(t *testing.T) {
	resetState()
	result := drawLottery()

	if result.DrawnAt.IsZero() {
		t.Error("DrawnAt がゼロ値になっている")
	}
}

// 並列抽選でチケット番号が重複しないことを確認（並行安全性）
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

// 景品テーブルの重みの合計が1000であることを確認
func TestPrizesWeightSum(t *testing.T) {
	total := 0
	for _, p := range prizes {
		total += p.Weight
	}
	if total != 1000 {
		t.Errorf("景品の重み合計: got %d, want 1000", total)
	}
}

// すべての景品が必須フィールドを持つことを確認
func TestPrizesHaveRequiredFields(t *testing.T) {
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

// 景品の等級がすべて異なることを確認（重複なし）
func TestPrizesGradesAreUnique(t *testing.T) {
	seen := make(map[PrizeGrade]bool)
	for _, p := range prizes {
		if seen[p.Grade] {
			t.Errorf("等級が重複している: %s", p.Grade)
		}
		seen[p.Grade] = true
	}
}

// ==============================
// 統計のテスト
// ==============================

// 空の履歴の場合は TotalDraws=0 であることを確認
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

// TotalDraws が実際の抽選回数と一致することを確認
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

// GradeCount の合計が TotalDraws と一致することを確認
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

// LastUpdated がゼロ値でないことを確認
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

// GET /api/draw は 405 を返すことを確認
func TestApiDrawHandler_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/draw", nil)
	w := httptest.NewRecorder()
	apiDrawHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("GETリクエストのステータス: got %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

// POST /api/draw は有効な DrawResult を JSON で返すことを確認
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

// GET /api/history は抽選履歴の JSON 配列を返すことを確認
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

// GET /api/stats は統計情報の JSON を返すことを確認
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

// GET /api/prizes は全景品の JSON 配列を返すことを確認
func TestApiPrizesHandler_ReturnsAllPrizes(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/prizes", nil)
	w := httptest.NewRecorder()
	apiPrizesHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("ステータス: got %d, want %d", w.Code, http.StatusOK)
	}

	var ps []Prize
	if err := json.Unmarshal(w.Body.Bytes(), &ps); err != nil {
		t.Fatalf("JSONパースエラー: %v", err)
	}
	if len(ps) != len(prizes) {
		t.Errorf("景品数: got %d, want %d", len(ps), len(prizes))
	}
}

// GET / は HTML を返すことを確認
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
}
