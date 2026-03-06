// Package service implements the core lottery business logic.
package service

import (
	"errors"
	"math/rand/v2"
	"sync"
	"time"

	"garapon/model"
)

const maxHistory = 50

// weightBounds defines [min, max] weight ranges for each prize except 参加賞.
// 参加賞 always receives the remainder so that all weights sum to exactly 1000.
//
// Derived invariants:
//   - 参加賞 min = 1000 - (15+60+120+250+300) = 255
//   - 参加賞 max = 1000 - (1+10+30+80+100)    = 779
var weightBounds = [][2]int{
	{1, 15},    // 特等
	{10, 60},   // 1等
	{30, 120},  // 2等
	{80, 250},  // 3等
	{100, 300}, // 4等
	// 参加賞: remainder
}

// initialPrizes is the canonical starting prize table.
var initialPrizes = []model.Prize{
	{Grade: model.GradeTokutou, Name: "特等賞", Description: "豪華旅行券 ¥100,000", Ball: model.BallColor{Name: "金色", Hex: "#FFD700"}, Weight: 5},
	{Grade: model.GradeIttou, Name: "1等賞", Description: "商品券 ¥10,000", Ball: model.BallColor{Name: "赤", Hex: "#FF3333"}, Weight: 30},
	{Grade: model.GradeNittou, Name: "2等賞", Description: "商品券 ¥5,000", Ball: model.BallColor{Name: "青", Hex: "#3366FF"}, Weight: 75},
	{Grade: model.GradeSantou, Name: "3等賞", Description: "商品券 ¥1,000", Ball: model.BallColor{Name: "緑", Hex: "#33AA33"}, Weight: 190},
	{Grade: model.GradeYontou, Name: "4等賞", Description: "お買い物割引券 ¥500", Ball: model.BallColor{Name: "黄色", Hex: "#FFCC00"}, Weight: 200},
	{Grade: model.GradeHazure, Name: "参加賞", Description: "記念品プレゼント", Ball: model.BallColor{Name: "白", Hex: "#F0F0F0"}, Weight: 500},
}

// LotteryService is the interface satisfied by all lottery implementations.
type LotteryService interface {
	Draw() (model.DrawResult, error)
	History() []model.DrawResult
	Stats() model.Stats
	Prizes() model.PrizesInfo
}

type lotteryService struct {
	prizes        []model.Prize
	prizeMu       sync.RWMutex
	history       []model.DrawResult
	historyMu     sync.Mutex
	ticketCount   int
	nextRotateAt  time.Time
	lastRotatedAt time.Time
	interval      time.Duration
}

// New creates a LotteryService and starts the background rotation goroutine.
func New(interval time.Duration) LotteryService {
	svc := &lotteryService{
		prizes:       clonePrizes(initialPrizes),
		interval:     interval,
		nextRotateAt: time.Now().Add(interval),
	}
	go svc.startRotation()
	return svc
}

// NewWithoutRotation creates a LotteryService without background rotation.
// Intended for use in tests that need deterministic, timer-free execution.
func NewWithoutRotation() LotteryService {
	return &lotteryService{
		prizes: clonePrizes(initialPrizes),
	}
}

func (s *lotteryService) startRotation() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for range ticker.C {
		s.rotate()
	}
}

// rotate regenerates all prize weights and updates rotation timestamps.
// It is safe to call concurrently.
func (s *lotteryService) rotate() {
	weights := generateWeights(len(s.prizes))
	s.prizeMu.Lock()
	for i := range s.prizes {
		s.prizes[i].Weight = weights[i]
	}
	s.lastRotatedAt = time.Now()
	s.nextRotateAt = s.lastRotatedAt.Add(s.interval)
	s.prizeMu.Unlock()
}

// generateWeights returns a new weight slice of length `count` that sums to 1000.
// The last element (参加賞) absorbs the remainder after the others are sampled.
func generateWeights(count int) []int {
	weights := make([]int, count)
	total := 0
	for i, b := range weightBounds {
		if i >= count-1 {
			break
		}
		w := b[0] + rand.IntN(b[1]-b[0]+1)
		weights[i] = w
		total += w
	}
	weights[count-1] = 1000 - total
	return weights
}

// Draw performs one lottery draw and appends the result to history.
func (s *lotteryService) Draw() (model.DrawResult, error) {
	s.prizeMu.RLock()
	snapshot := clonePrizes(s.prizes)
	s.prizeMu.RUnlock()

	total := 0
	for _, p := range snapshot {
		total += p.Weight
	}
	if total <= 0 {
		return model.DrawResult{}, errors.New("景品テーブルの重み合計が0です")
	}

	n := rand.IntN(total)
	selected := snapshot[len(snapshot)-1]
	cumulative := 0
	for _, p := range snapshot {
		cumulative += p.Weight
		if n < cumulative {
			selected = p
			break
		}
	}

	s.historyMu.Lock()
	s.ticketCount++
	num := s.ticketCount
	s.historyMu.Unlock()

	result := model.DrawResult{
		Prize:     selected,
		DrawnAt:   time.Now(),
		TicketNum: num,
	}

	s.historyMu.Lock()
	s.history = append([]model.DrawResult{result}, s.history...)
	if len(s.history) > maxHistory {
		s.history = s.history[:maxHistory]
	}
	s.historyMu.Unlock()

	return result, nil
}

// History returns a copy of the current draw history (most recent first).
func (s *lotteryService) History() []model.DrawResult {
	s.historyMu.Lock()
	defer s.historyMu.Unlock()
	if len(s.history) == 0 {
		return []model.DrawResult{}
	}
	cp := make([]model.DrawResult, len(s.history))
	copy(cp, s.history)
	return cp
}

// Stats returns aggregate statistics calculated from current history.
func (s *lotteryService) Stats() model.Stats {
	s.historyMu.Lock()
	defer s.historyMu.Unlock()
	counts := make(map[string]int)
	for _, r := range s.history {
		counts[string(r.Prize.Grade)]++
	}
	return model.Stats{
		TotalDraws:  len(s.history),
		GradeCount:  counts,
		LastUpdated: time.Now(),
	}
}

// Prizes returns the current prize table together with rotation metadata.
func (s *lotteryService) Prizes() model.PrizesInfo {
	s.prizeMu.RLock()
	defer s.prizeMu.RUnlock()
	return model.PrizesInfo{
		Prizes:              clonePrizes(s.prizes),
		NextRotationAt:      s.nextRotateAt,
		LastRotatedAt:       s.lastRotatedAt,
		RotationIntervalSec: int(s.interval.Seconds()),
	}
}

func clonePrizes(src []model.Prize) []model.Prize {
	dst := make([]model.Prize, len(src))
	copy(dst, src)
	return dst
}
