package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"garapon/model"
	"garapon/service"
)

// ============================================================
// Mock service
// ============================================================

type mockService struct {
	drawResult model.DrawResult
	drawErr    error
	history    []model.DrawResult
	stats      model.Stats
	prizes     model.PrizesInfo
}

var _ service.LotteryService = (*mockService)(nil) // compile-time check

func (m *mockService) Draw() (model.DrawResult, error) { return m.drawResult, m.drawErr }
func (m *mockService) History() []model.DrawResult     { return m.history }
func (m *mockService) Stats() model.Stats              { return m.stats }
func (m *mockService) Prizes() model.PrizesInfo        { return m.prizes }

// defaultMock returns a mock that returns a valid 参加賞 result.
func defaultMock() *mockService {
	return &mockService{
		drawResult: model.DrawResult{
			Prize: model.Prize{
				Grade:       model.GradeHazure,
				Name:        "参加賞",
				Description: "記念品プレゼント",
				Ball:        model.BallColor{Name: "白", Hex: "#F0F0F0"},
				Weight:      500,
			},
			DrawnAt:   time.Now(),
			TicketNum: 1,
		},
		history: []model.DrawResult{},
		stats:   model.Stats{GradeCount: map[string]int{}},
		prizes: model.PrizesInfo{
			Prizes:              []model.Prize{{Grade: model.GradeHazure, Weight: 1000}},
			NextRotationAt:      time.Now().Add(30 * time.Second),
			RotationIntervalSec: 30,
		},
	}
}

// helper: perform a request and return the recorder
func do(h *Handler, method, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	switch path {
	case "/api/draw":
		h.Draw(w, req)
	case "/api/history":
		h.History(w, req)
	case "/api/stats":
		h.Stats(w, req)
	case "/api/prizes":
		h.Prizes(w, req)
	case "/":
		h.Home(w, req)
	}
	return w
}

// ============================================================
// GET /api/draw — 正常系
// ============================================================

func TestDraw_GET_Returns200(t *testing.T) {
	h := New(defaultMock())
	w := do(h, http.MethodGet, "/api/draw")
	if w.Code != http.StatusOK {
		t.Errorf("ステータス: got %d, want %d", w.Code, http.StatusOK)
	}
}

func TestDraw_GET_ContentTypeIsJSON(t *testing.T) {
	h := New(defaultMock())
	w := do(h, http.MethodGet, "/api/draw")
	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type: got %q, want application/json", ct)
	}
}

func TestDraw_GET_ResponseDecodesAsDrawResult(t *testing.T) {
	h := New(defaultMock())
	w := do(h, http.MethodGet, "/api/draw")
	var result model.DrawResult
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

// ============================================================
// GET /api/draw — 異常系
// ============================================================

// POST は 405 を返すことを確認
func TestDraw_POST_Returns405(t *testing.T) {
	h := New(defaultMock())
	w := do(h, http.MethodPost, "/api/draw")
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("ステータス: got %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

// PUT は 405 を返すことを確認
func TestDraw_PUT_Returns405(t *testing.T) {
	h := New(defaultMock())
	w := do(h, http.MethodPut, "/api/draw")
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("ステータス: got %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

// Allow ヘッダーが 405 に含まれることを確認
func TestDraw_MethodNotAllowed_SetsAllowHeader(t *testing.T) {
	h := New(defaultMock())
	w := do(h, http.MethodPost, "/api/draw")
	if w.Header().Get("Allow") != http.MethodGet {
		t.Errorf("Allow ヘッダー: got %q, want GET", w.Header().Get("Allow"))
	}
}

// サービスがエラーを返したとき 500 を返すことを確認
func TestDraw_ServiceError_Returns500(t *testing.T) {
	mock := defaultMock()
	mock.drawErr = errors.New("景品テーブルの重み合計が0です")
	h := New(mock)
	w := do(h, http.MethodGet, "/api/draw")
	if w.Code != http.StatusInternalServerError {
		t.Errorf("ステータス: got %d, want %d", w.Code, http.StatusInternalServerError)
	}
	var errResp model.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("エラーレスポンスのパース失敗: %v", err)
	}
	if errResp.Error == "" {
		t.Error("エラーメッセージが空")
	}
}

// ============================================================
// GET /api/history — 正常系・異常系
// ============================================================

func TestHistory_GET_Returns200(t *testing.T) {
	h := New(defaultMock())
	w := do(h, http.MethodGet, "/api/history")
	if w.Code != http.StatusOK {
		t.Errorf("ステータス: got %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHistory_GET_ReturnsJSONArray(t *testing.T) {
	mock := defaultMock()
	mock.history = []model.DrawResult{mock.drawResult, mock.drawResult}
	h := New(mock)
	w := do(h, http.MethodGet, "/api/history")
	var hist []model.DrawResult
	if err := json.Unmarshal(w.Body.Bytes(), &hist); err != nil {
		t.Fatalf("JSONパースエラー: %v", err)
	}
	if len(hist) != 2 {
		t.Errorf("履歴件数: got %d, want 2", len(hist))
	}
}

func TestHistory_POST_Returns405(t *testing.T) {
	h := New(defaultMock())
	w := do(h, http.MethodPost, "/api/history")
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("ステータス: got %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

// ============================================================
// GET /api/stats — 正常系・異常系
// ============================================================

func TestStats_GET_Returns200(t *testing.T) {
	h := New(defaultMock())
	w := do(h, http.MethodGet, "/api/stats")
	if w.Code != http.StatusOK {
		t.Errorf("ステータス: got %d, want %d", w.Code, http.StatusOK)
	}
}

func TestStats_GET_ResponseDecodesAsStats(t *testing.T) {
	mock := defaultMock()
	mock.stats = model.Stats{
		TotalDraws:  5,
		GradeCount:  map[string]int{"参加賞": 5},
		LastUpdated: time.Now(),
	}
	h := New(mock)
	w := do(h, http.MethodGet, "/api/stats")
	var s model.Stats
	if err := json.Unmarshal(w.Body.Bytes(), &s); err != nil {
		t.Fatalf("JSONパースエラー: %v", err)
	}
	if s.TotalDraws != 5 {
		t.Errorf("TotalDraws: got %d, want 5", s.TotalDraws)
	}
}

func TestStats_POST_Returns405(t *testing.T) {
	h := New(defaultMock())
	w := do(h, http.MethodPost, "/api/stats")
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("ステータス: got %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

// ============================================================
// GET /api/prizes — 正常系・異常系
// ============================================================

func TestPrizes_GET_Returns200(t *testing.T) {
	h := New(defaultMock())
	w := do(h, http.MethodGet, "/api/prizes")
	if w.Code != http.StatusOK {
		t.Errorf("ステータス: got %d, want %d", w.Code, http.StatusOK)
	}
}

func TestPrizes_GET_HasRotationFields(t *testing.T) {
	h := New(defaultMock())
	w := do(h, http.MethodGet, "/api/prizes")
	var info model.PrizesInfo
	if err := json.Unmarshal(w.Body.Bytes(), &info); err != nil {
		t.Fatalf("JSONパースエラー: %v", err)
	}
	if len(info.Prizes) == 0 {
		t.Error("Prizes が空")
	}
	if info.RotationIntervalSec != 30 {
		t.Errorf("RotationIntervalSec: got %d, want 30", info.RotationIntervalSec)
	}
	if info.NextRotationAt.IsZero() {
		t.Error("NextRotationAt がゼロ値")
	}
}

func TestPrizes_POST_Returns405(t *testing.T) {
	h := New(defaultMock())
	w := do(h, http.MethodPost, "/api/prizes")
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("ステータス: got %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

// ============================================================
// GET / — Home
// ============================================================

func TestHome_GET_Returns200(t *testing.T) {
	h := New(defaultMock())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.Home(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("ステータス: got %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHome_GET_ContentTypeIsHTML(t *testing.T) {
	h := New(defaultMock())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.Home(w, req)
	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type: got %q, want text/html", ct)
	}
}

func TestHome_GET_BodyContainsExpectedElements(t *testing.T) {
	h := New(defaultMock())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.Home(w, req)
	body := w.Body.String()
	for _, want := range []string{"<!DOCTYPE html>", "ガラガラポン", "countdown", "/api/draw"} {
		if !strings.Contains(body, want) {
			t.Errorf("HTMLに %q が含まれていない", want)
		}
	}
}

// 未知のパス（例: /unknown）は 404 を返すことを確認
func TestHome_UnknownPath_Returns404(t *testing.T) {
	h := New(defaultMock())
	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	w := httptest.NewRecorder()
	h.Home(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("ステータス: got %d, want %d", w.Code, http.StatusNotFound)
	}
}

// ============================================================
// RegisterRoutes — 統合確認
// ============================================================

func TestRegisterRoutes_AllEndpointsReachable(t *testing.T) {
	h := New(defaultMock())
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	endpoints := []struct {
		method string
		path   string
		want   int
	}{
		{http.MethodGet, "/", http.StatusOK},
		{http.MethodGet, "/api/draw", http.StatusOK},
		{http.MethodGet, "/api/history", http.StatusOK},
		{http.MethodGet, "/api/stats", http.StatusOK},
		{http.MethodGet, "/api/prizes", http.StatusOK},
	}

	for _, tc := range endpoints {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != tc.want {
			t.Errorf("%s %s: ステータス got %d, want %d", tc.method, tc.path, w.Code, tc.want)
		}
	}
}
