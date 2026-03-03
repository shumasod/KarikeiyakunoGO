package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"sync"
	"time"
)

// 景品の等級
type PrizeGrade string

const (
	GradeTokutou PrizeGrade = "特等"
	GradeIttou   PrizeGrade = "1等"
	GradeNittou  PrizeGrade = "2等"
	GradeSantou  PrizeGrade = "3等"
	GradeYontou  PrizeGrade = "4等"
	GradeHazure  PrizeGrade = "参加賞"
)

// ボールの色
type BallColor struct {
	Name string `json:"name"`
	Hex  string `json:"hex"`
}

// 景品情報
type Prize struct {
	Grade       PrizeGrade `json:"grade"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Ball        BallColor  `json:"ball"`
	Weight      int        `json:"weight"`
}

// 抽選結果
type DrawResult struct {
	Prize     Prize     `json:"prize"`
	DrawnAt   time.Time `json:"drawn_at"`
	TicketNum int       `json:"ticket_num"`
}

// 統計情報
type Stats struct {
	TotalDraws  int            `json:"total_draws"`
	GradeCount  map[string]int `json:"grade_count"`
	LastUpdated time.Time      `json:"last_updated"`
}

// 景品一覧レスポンス（ローテーション情報付き）
type PrizesInfo struct {
	Prizes              []Prize   `json:"prizes"`
	NextRotationAt      time.Time `json:"next_rotation_at"`
	LastRotatedAt       time.Time `json:"last_rotated_at"`
	RotationIntervalSec int       `json:"rotation_interval_sec"`
}

// ローテーション間隔
const rotationInterval = 30 * time.Second

// 各景品（参加賞を除く）の重みの範囲 [min, max]
// 参加賞 は 1000 - 他の合計 で自動計算
// Max合計 = 15+60+120+250+300 = 745 → 参加賞 min = 255
// Min合計 = 1+10+30+80+100  = 221 → 参加賞 max = 779
var weightBounds = [][2]int{
	{1, 15},    // 特等
	{10, 60},   // 1等
	{30, 120},  // 2等
	{80, 250},  // 3等
	{100, 300}, // 4等
}

// 景品テーブル（prizeMu で保護）
var prizes = []Prize{
	{
		Grade:       GradeTokutou,
		Name:        "特等賞",
		Description: "豪華旅行券 ¥100,000",
		Ball:        BallColor{Name: "金色", Hex: "#FFD700"},
		Weight:      5,
	},
	{
		Grade:       GradeIttou,
		Name:        "1等賞",
		Description: "商品券 ¥10,000",
		Ball:        BallColor{Name: "赤", Hex: "#FF3333"},
		Weight:      30,
	},
	{
		Grade:       GradeNittou,
		Name:        "2等賞",
		Description: "商品券 ¥5,000",
		Ball:        BallColor{Name: "青", Hex: "#3366FF"},
		Weight:      75,
	},
	{
		Grade:       GradeSantou,
		Name:        "3等賞",
		Description: "商品券 ¥1,000",
		Ball:        BallColor{Name: "緑", Hex: "#33AA33"},
		Weight:      190,
	},
	{
		Grade:       GradeYontou,
		Name:        "4等賞",
		Description: "お買い物割引券 ¥500",
		Ball:        BallColor{Name: "黄色", Hex: "#FFCC00"},
		Weight:      200,
	},
	{
		Grade:       GradeHazure,
		Name:        "参加賞",
		Description: "記念品プレゼント",
		Ball:        BallColor{Name: "白", Hex: "#F0F0F0"},
		Weight:      500,
	},
}

var (
	prizeMu       sync.RWMutex
	nextRotateAt  time.Time
	lastRotatedAt time.Time

	history     []DrawResult
	historyMu   sync.Mutex
	ticketCount int
)

// 新しい重みセットをランダム生成（合計は必ず 1000）
func generateNewWeights() []int {
	weights := make([]int, len(prizes))
	total := 0
	for i, b := range weightBounds {
		w := b[0] + rand.IntN(b[1]-b[0]+1)
		weights[i] = w
		total += w
	}
	weights[len(prizes)-1] = 1000 - total // 参加賞
	return weights
}

// 景品の重みをローテーション
func rotatePrizes() {
	weights := generateNewWeights()

	prizeMu.Lock()
	for i := range prizes {
		prizes[i].Weight = weights[i]
	}
	lastRotatedAt = time.Now()
	nextRotateAt = lastRotatedAt.Add(rotationInterval)
	prizeMu.Unlock()
}

// バックグラウンドでローテーションを開始
func startPrizeRotation() {
	nextRotateAt = time.Now().Add(rotationInterval)
	go func() {
		ticker := time.NewTicker(rotationInterval)
		defer ticker.Stop()
		for range ticker.C {
			rotatePrizes()
		}
	}()
}

// 抽選を行う
func drawLottery() DrawResult {
	prizeMu.RLock()
	total := 0
	for _, p := range prizes {
		total += p.Weight
	}
	n := rand.IntN(total)
	cumulative := 0
	selected := prizes[len(prizes)-1]
	for _, p := range prizes {
		cumulative += p.Weight
		if n < cumulative {
			selected = p
			break
		}
	}
	prizeMu.RUnlock()

	historyMu.Lock()
	ticketCount++
	num := ticketCount
	historyMu.Unlock()

	result := DrawResult{
		Prize:     selected,
		DrawnAt:   time.Now(),
		TicketNum: num,
	}

	historyMu.Lock()
	history = append([]DrawResult{result}, history...)
	if len(history) > 50 {
		history = history[:50]
	}
	historyMu.Unlock()

	return result
}

// 統計を計算する
func calcStats() Stats {
	historyMu.Lock()
	defer historyMu.Unlock()

	counts := make(map[string]int)
	for _, r := range history {
		counts[string(r.Prize.Grade)]++
	}
	return Stats{
		TotalDraws:  len(history),
		GradeCount:  counts,
		LastUpdated: time.Now(),
	}
}

// ホームページ
func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>🎰 商店街ガラガラポン抽選会</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: 'Hiragino Kaku Gothic Pro', 'Meiryo', sans-serif;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
            min-height: 100vh;
            color: white;
            overflow-x: hidden;
        }
        .header {
            text-align: center;
            padding: 30px 20px 10px;
            background: linear-gradient(180deg, rgba(255,200,0,0.15) 0%, transparent 100%);
        }
        .header h1 {
            font-size: 2.5em;
            color: #FFD700;
            text-shadow: 0 0 20px rgba(255,215,0,0.6), 2px 2px 4px rgba(0,0,0,0.8);
            margin-bottom: 8px;
            letter-spacing: 3px;
        }
        .header p { color: #aaa; font-size: 1em; }
        .main-content {
            display: flex;
            flex-wrap: wrap;
            gap: 30px;
            padding: 30px;
            max-width: 1200px;
            margin: 0 auto;
            justify-content: center;
        }
        .machine-section { flex: 0 0 auto; display: flex; flex-direction: column; align-items: center; }
        .machine-wrapper { position: relative; width: 320px; }
        .drum-container { position: relative; width: 280px; height: 280px; margin: 0 auto; }
        .drum {
            width: 100%; height: 100%; border-radius: 50%;
            background: radial-gradient(circle at 35% 35%, #888, #444 60%, #222);
            border: 8px solid #666;
            box-shadow: 0 0 0 4px #888, 0 0 30px rgba(0,0,0,0.8), inset 0 0 40px rgba(0,0,0,0.5);
            position: relative; overflow: hidden; transition: transform 0.1s;
        }
        .drum.spinning { animation: drumSpin 0.2s linear infinite; }
        @keyframes drumSpin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
        .drum-grid {
            position: absolute; inset: 10px; border-radius: 50%;
            background:
                repeating-linear-gradient(0deg, transparent, transparent 18px, rgba(255,255,255,0.08) 18px, rgba(255,255,255,0.08) 19px),
                repeating-linear-gradient(90deg, transparent, transparent 18px, rgba(255,255,255,0.08) 18px, rgba(255,255,255,0.08) 19px);
        }
        .drum-balls { position: absolute; inset: 20px; border-radius: 50%; overflow: hidden; }
        .mini-ball {
            position: absolute; width: 28px; height: 28px; border-radius: 50%;
            box-shadow: inset -3px -3px 6px rgba(0,0,0,0.4), inset 2px 2px 4px rgba(255,255,255,0.3);
        }
        .handle-area { display: flex; justify-content: flex-end; margin-top: -40px; padding-right: 10px; }
        .handle { width: 60px; height: 120px; position: relative; }
        .handle-bar {
            width: 12px; height: 80px;
            background: linear-gradient(90deg, #888, #ccc, #888);
            border-radius: 6px; margin: 0 auto; box-shadow: 2px 2px 6px rgba(0,0,0,0.5);
        }
        .handle-knob {
            width: 36px; height: 36px; border-radius: 50%;
            background: radial-gradient(circle at 35% 35%, #ffcc00, #cc8800);
            border: 3px solid #aa6600; margin: 0 auto;
            box-shadow: 0 4px 8px rgba(0,0,0,0.5);
            cursor: pointer; transition: transform 0.1s;
        }
        .handle-knob:hover { transform: scale(1.1); }
        .handle-knob:active { transform: scale(0.95); }
        .outlet {
            width: 100px; height: 50px; margin: 10px auto 0;
            background: linear-gradient(180deg, #333, #555);
            border-radius: 0 0 20px 20px; border: 4px solid #666; border-top: none;
            position: relative; display: flex; align-items: center; justify-content: center; overflow: visible;
        }
        .outlet-label { font-size: 0.7em; color: #aaa; letter-spacing: 1px; }
        .result-ball {
            width: 80px; height: 80px; border-radius: 50%;
            position: absolute; top: -120px; left: 50%; transform: translateX(-50%);
            box-shadow: inset -8px -8px 15px rgba(0,0,0,0.4), inset 4px 4px 10px rgba(255,255,255,0.4), 0 8px 20px rgba(0,0,0,0.5);
            display: none; animation: ballDrop 0.6s ease-out forwards;
        }
        @keyframes ballDrop {
            0%   { top: -140px; opacity: 0; transform: translateX(-50%) scale(0.5); }
            50%  { top: -100px; opacity: 1; transform: translateX(-50%) scale(1.1); }
            100% { top: -110px; opacity: 1; transform: translateX(-50%) scale(1); }
        }
        .result-ball.show { display: block; }
        .machine-base {
            width: 300px; height: 20px;
            background: linear-gradient(180deg, #888, #555);
            border-radius: 0 0 10px 10px; margin: 0 auto; box-shadow: 0 6px 12px rgba(0,0,0,0.5);
        }
        .draw-btn {
            margin-top: 30px; padding: 18px 60px; font-size: 1.4em; font-weight: bold;
            background: linear-gradient(135deg, #FFD700, #FF8C00); color: #1a1a1a;
            border: none; border-radius: 50px; cursor: pointer;
            box-shadow: 0 6px 20px rgba(255,140,0,0.5); transition: all 0.2s; letter-spacing: 2px;
        }
        .draw-btn:hover:not(:disabled) { transform: translateY(-3px); box-shadow: 0 10px 30px rgba(255,140,0,0.7); }
        .draw-btn:active:not(:disabled) { transform: translateY(0); }
        .draw-btn:disabled { opacity: 0.6; cursor: not-allowed; }
        .result-section { flex: 1; min-width: 300px; }
        .result-panel {
            background: rgba(255,255,255,0.05); border: 1px solid rgba(255,255,255,0.1);
            border-radius: 20px; padding: 30px; margin-bottom: 20px; min-height: 200px;
            display: flex; flex-direction: column; align-items: center; justify-content: center;
            text-align: center; transition: all 0.3s;
        }
        .result-panel.highlight {
            border-color: rgba(255,215,0,0.5); background: rgba(255,215,0,0.05);
            box-shadow: 0 0 30px rgba(255,215,0,0.2);
        }
        .result-grade {
            font-size: 3em; font-weight: bold; margin-bottom: 10px;
            opacity: 0; transform: scale(0.5);
            transition: all 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
        }
        .result-grade.show { opacity: 1; transform: scale(1); }
        .result-name { font-size: 1.4em; color: #ddd; margin-bottom: 8px; }
        .result-desc { font-size: 1.1em; color: #aaa; }
        .wait-msg { color: #555; font-size: 1.1em; }
        .prize-table-section {
            background: rgba(255,255,255,0.05); border: 1px solid rgba(255,255,255,0.1);
            border-radius: 20px; padding: 25px; margin-bottom: 20px;
            transition: background 0.5s;
        }
        .prize-table-section.flash {
            background: rgba(255,215,0,0.1);
        }
        .prize-table-header {
            display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px;
        }
        .prize-table-header h2 { color: #FFD700; font-size: 1.2em; letter-spacing: 2px; }
        .live-badge {
            font-size: 0.65em; background: #cc2200; color: white;
            padding: 2px 7px; border-radius: 4px; letter-spacing: 1px;
            animation: livePulse 1.2s ease-in-out infinite;
        }
        @keyframes livePulse { 0%,100%{opacity:1;} 50%{opacity:0.4;} }
        .rotation-timer {
            font-size: 0.82em; color: #888; margin-bottom: 12px;
            display: flex; align-items: center; gap: 6px;
        }
        .countdown-num {
            color: #FFD700; font-weight: bold; font-size: 1.15em;
            min-width: 24px; display: inline-block; text-align: center;
        }
        .countdown-num.soon { color: #FF6644; animation: urgentPulse 0.5s ease-in-out infinite; }
        @keyframes urgentPulse { 0%,100%{transform:scale(1);} 50%{transform:scale(1.2);} }
        .prize-row {
            display: flex; align-items: center; gap: 12px; padding: 8px 0;
            border-bottom: 1px solid rgba(255,255,255,0.05);
        }
        .prize-row:last-child { border-bottom: none; }
        .ball-icon {
            width: 24px; height: 24px; border-radius: 50%; flex-shrink: 0;
            box-shadow: inset -2px -2px 4px rgba(0,0,0,0.4), inset 1px 1px 3px rgba(255,255,255,0.3);
        }
        .prize-grade-label { font-weight: bold; min-width: 50px; font-size: 0.95em; }
        .prize-prize-name { color: #ddd; flex: 1; font-size: 0.9em; }
        .prize-prob { color: #888; font-size: 0.8em; min-width: 50px; text-align: right; }
        .history-section {
            background: rgba(255,255,255,0.05); border: 1px solid rgba(255,255,255,0.1);
            border-radius: 20px; padding: 25px;
        }
        .history-section h2 { color: #FFD700; margin-bottom: 15px; font-size: 1.2em; letter-spacing: 2px; }
        #history-list { max-height: 300px; overflow-y: auto; }
        .history-item {
            display: flex; align-items: center; gap: 10px; padding: 8px 0;
            border-bottom: 1px solid rgba(255,255,255,0.05);
            animation: fadeIn 0.3s ease;
        }
        @keyframes fadeIn { from{opacity:0;transform:translateY(-10px);} to{opacity:1;transform:translateY(0);} }
        .history-num { color: #666; font-size: 0.8em; min-width: 40px; }
        .history-grade { font-weight: bold; font-size: 0.9em; min-width: 60px; }
        .history-prize { color: #aaa; font-size: 0.85em; flex: 1; }
        .history-time { color: #555; font-size: 0.75em; }
        .stats-bar { display: flex; gap: 15px; flex-wrap: wrap; margin-top: 15px; justify-content: center; }
        .stat-chip { background: rgba(255,255,255,0.08); padding: 6px 14px; border-radius: 20px; font-size: 0.85em; color: #ccc; }
        .stat-chip span { color: #FFD700; font-weight: bold; }
        .confetti-container { position: fixed; top: 0; left: 0; width: 100%; height: 100%; pointer-events: none; z-index: 1000; }
        .confetti-piece { position: absolute; width: 10px; height: 10px; top: -20px; animation: confettiFall linear forwards; }
        @keyframes confettiFall { to { top: 110vh; transform: rotate(720deg); } }
        .grade-tokutou{color:#FFD700;} .grade-ittou{color:#FF6666;} .grade-nittou{color:#6699FF;}
        .grade-santou{color:#66CC66;} .grade-yontou{color:#FFDD44;} .grade-hazure{color:#AAAAAA;}
        .no-history { color: #555; font-size: 0.9em; text-align: center; padding: 20px 0; }
        @media (max-width: 700px) {
            .header h1 { font-size: 1.8em; }
            .main-content { padding: 15px; gap: 20px; }
            .draw-btn { font-size: 1.1em; padding: 14px 40px; }
        }
    </style>
</head>
<body>
<div class="header">
    <h1>🎰 商店街ガラガラポン抽選会 🎰</h1>
    <p>ハンドルを回して景品をゲットしよう！</p>
</div>
<div class="main-content">
    <div class="machine-section">
        <div class="machine-wrapper">
            <div class="drum-container">
                <div class="drum" id="drum">
                    <div class="drum-grid"></div>
                    <div class="drum-balls" id="drumBalls"></div>
                </div>
                <div class="handle-area">
                    <div class="handle">
                        <div class="handle-bar"></div>
                        <div class="handle-knob" onclick="startDraw()" title="ハンドルを回す！"></div>
                    </div>
                </div>
            </div>
            <div class="outlet">
                <span class="outlet-label">排出口</span>
                <div class="result-ball" id="resultBall"></div>
            </div>
            <div class="machine-base"></div>
        </div>
        <button class="draw-btn" id="drawBtn" onclick="startDraw()">🎲 ガラガラ回す！</button>
        <div class="stats-bar">
            <div class="stat-chip">総抽選数: <span id="totalDraws">0</span>回</div>
        </div>
    </div>

    <div class="result-section">
        <div class="result-panel" id="resultPanel">
            <p class="wait-msg">ボタンを押して抽選してみよう！</p>
        </div>

        <div class="prize-table-section" id="prizeTableSection">
            <div class="prize-table-header">
                <h2>🎁 景品一覧</h2>
                <span class="live-badge">LIVE</span>
            </div>
            <div class="rotation-timer">
                🔄 次の確率変更まで
                <span class="countdown-num" id="countdown">--</span>秒
            </div>
            <div id="prizeTable"><p style="color:#555;font-size:0.9em;">読込中...</p></div>
        </div>

        <div class="history-section">
            <h2>📋 抽選履歴</h2>
            <div id="history-list"><p class="no-history">まだ抽選していません</p></div>
        </div>
    </div>
</div>
<div class="confetti-container" id="confettiContainer"></div>

<script>
const gradeClasses = {
    "特等":"grade-tokutou","1等":"grade-ittou","2等":"grade-nittou",
    "3等":"grade-santou","4等":"grade-yontou","参加賞":"grade-hazure"
};

let currentPrizes = [];
let nextRotationAt = null;
let isDrawing = false;
let totalDraws = 0;

function lighten(hex) {
    const r = parseInt(hex.slice(1,3),16);
    const g = parseInt(hex.slice(3,5),16);
    const b = parseInt(hex.slice(5,7),16);
    return ` + "`" + `rgb(${Math.min(255,r+80)},${Math.min(255,g+80)},${Math.min(255,b+80)})` + "`" + `;
}

function weightToProb(weight) {
    return (weight / 10).toFixed(1) + '%';
}

async function fetchPrizes() {
    try {
        const res = await fetch('/api/prizes');
        const info = await res.json();
        nextRotationAt = new Date(info.next_rotation_at);

        const weightsChanged = currentPrizes.length > 0 &&
            currentPrizes.some((p, i) => p.weight !== info.prizes[i].weight);

        currentPrizes = info.prizes;
        renderPrizeTable();
        populateDrum();

        if (weightsChanged) {
            flashPrizeTable();
        }
    } catch(e) {
        console.error('景品取得エラー:', e);
    }
}

function renderPrizeTable() {
    if (!currentPrizes.length) return;
    document.getElementById('prizeTable').innerHTML = currentPrizes.map(p => ` + "`" + `
        <div class="prize-row">
            <div class="ball-icon" style="background:radial-gradient(circle at 35% 35%, ${lighten(p.ball.hex)}, ${p.ball.hex} 70%);"></div>
            <span class="prize-grade-label ${gradeClasses[p.grade] || ''}">${p.grade}</span>
            <span class="prize-prize-name">${p.description}</span>
            <span class="prize-prob">${weightToProb(p.weight)}</span>
        </div>
    ` + "`" + `).join('');
}

function flashPrizeTable() {
    const sec = document.getElementById('prizeTableSection');
    sec.classList.remove('flash');
    void sec.offsetWidth;
    sec.classList.add('flash');
    setTimeout(() => sec.classList.remove('flash'), 800);
}

function populateDrum() {
    if (!currentPrizes.length) return;
    const container = document.getElementById('drumBalls');
    const colors = currentPrizes.map(p => p.ball.hex);
    const positions = [
        {top:'15%',left:'20%'},{top:'15%',left:'55%'},{top:'30%',left:'10%'},
        {top:'30%',left:'40%'},{top:'30%',left:'68%'},{top:'50%',left:'15%'},
        {top:'50%',left:'45%'},{top:'50%',left:'72%'},{top:'65%',left:'25%'},
        {top:'65%',left:'55%'},{top:'78%',left:'15%'},{top:'78%',left:'42%'},
        {top:'78%',left:'65%'},{top:'20%',left:'75%'},{top:'42%',left:'30%'},
    ];
    container.innerHTML = positions.map((pos, i) => {
        const c = colors[i % colors.length];
        return ` + "`" + `<div class="mini-ball" style="top:${pos.top};left:${pos.left};background:radial-gradient(circle at 35% 35%, ${lighten(c)}, ${c} 70%);"></div>` + "`" + `;
    }).join('');
}

function updateCountdown() {
    if (!nextRotationAt) return;
    const remaining = Math.max(0, Math.ceil((nextRotationAt.getTime() - Date.now()) / 1000));
    const el = document.getElementById('countdown');
    if (!el) return;
    el.textContent = remaining;
    el.className = 'countdown-num' + (remaining <= 5 ? ' soon' : '');
    if (remaining === 0) {
        setTimeout(fetchPrizes, 600);
    }
}

async function startDraw() {
    if (isDrawing) return;
    isDrawing = true;

    const btn = document.getElementById('drawBtn');
    const drum = document.getElementById('drum');
    const resultBall = document.getElementById('resultBall');
    const resultPanel = document.getElementById('resultPanel');

    btn.disabled = true;
    btn.textContent = '🎲 抽選中...';
    drum.classList.add('spinning');
    resultBall.classList.remove('show');
    resultPanel.classList.remove('highlight');
    resultPanel.innerHTML = '<p style="color:#888;font-size:1.1em;">🎰 ガラガラ回転中...</p>';

    let result;
    try {
        const res = await fetch('/api/draw', { method: 'POST' });
        result = await res.json();
    } catch(e) {
        resultPanel.innerHTML = '<p style="color:#f66;">エラーが発生しました</p>';
        drum.classList.remove('spinning');
        btn.disabled = false;
        btn.textContent = '🎲 ガラガラ回す！';
        isDrawing = false;
        return;
    }

    setTimeout(() => {
        drum.classList.remove('spinning');

        const ballColor = result.prize.ball.hex;
        resultBall.style.background = ` + "`" + `radial-gradient(circle at 35% 35%, ${lighten(ballColor)}, ${ballColor} 70%)` + "`" + `;
        resultBall.classList.add('show');

        const grade = result.prize.grade;
        const gradeClass = gradeClasses[grade] || 'grade-hazure';

        resultPanel.classList.add('highlight');
        resultPanel.innerHTML = ` + "`" + `
            <div class="result-grade ${gradeClass}" id="rg">${result.prize.grade}</div>
            <div class="result-name">${result.prize.name}</div>
            <div class="result-desc">${result.prize.description}</div>
        ` + "`" + `;
        setTimeout(() => { const rg = document.getElementById('rg'); if(rg) rg.classList.add('show'); }, 50);

        if (['特等','1等','2等'].includes(grade)) {
            launchConfetti(grade === '特等' ? 80 : grade === '1等' ? 50 : 30);
        }

        totalDraws++;
        document.getElementById('totalDraws').textContent = totalDraws;
        addHistory(result);
        btn.disabled = false;
        btn.textContent = '🎲 もう一度回す！';
        isDrawing = false;
    }, 1500);
}

function addHistory(result) {
    const list = document.getElementById('history-list');
    const noHistory = list.querySelector('.no-history');
    if (noHistory) noHistory.remove();

    const grade = result.prize.grade;
    const now = new Date();
    const item = document.createElement('div');
    item.className = 'history-item';
    item.innerHTML = ` + "`" + `
        <div style="width:16px;height:16px;flex-shrink:0;border-radius:50%;background:radial-gradient(circle at 35% 35%, ${lighten(result.prize.ball.hex)}, ${result.prize.ball.hex} 70%);"></div>
        <span class="history-num">#${result.ticket_num}</span>
        <span class="history-grade ${gradeClasses[grade] || ''}">${grade}</span>
        <span class="history-prize">${result.prize.description}</span>
        <span class="history-time">${now.toLocaleTimeString('ja-JP')}</span>
    ` + "`" + `;
    list.insertBefore(item, list.firstChild);
    const items = list.querySelectorAll('.history-item');
    if (items.length > 20) items[items.length-1].remove();
}

function launchConfetti(count) {
    const container = document.getElementById('confettiContainer');
    const colors = ['#FFD700','#FF6666','#6699FF','#66CC66','#FF99CC','#FFCC44'];
    for (let i = 0; i < count; i++) {
        setTimeout(() => {
            const piece = document.createElement('div');
            piece.className = 'confetti-piece';
            const size = 6 + Math.random() * 10;
            piece.style.cssText = ` + "`" + `
                left:${Math.random()*100}%;width:${size}px;height:${size}px;
                background:${colors[Math.floor(Math.random()*colors.length)]};
                border-radius:${Math.random()>0.5?'50%':'2px'};
                animation-duration:${1.5+Math.random()*2}s;
                animation-delay:${Math.random()*0.5}s;
                opacity:${0.6+Math.random()*0.4};
            ` + "`" + `;
            container.appendChild(piece);
            setTimeout(() => piece.remove(), 4000);
        }, i * 30);
    }
}

// カウントダウンは毎秒更新
setInterval(updateCountdown, 1000);
// 景品テーブルは5秒ごとに同期（ローテーション後のズレを補正）
setInterval(fetchPrizes, 5000);
// 初期ロード
fetchPrizes();
</script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, tmpl)
}

// 抽選APIハンドラ
func apiDrawHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	result := drawLottery()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// 履歴APIハンドラ
func apiHistoryHandler(w http.ResponseWriter, r *http.Request) {
	historyMu.Lock()
	defer historyMu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

// 統計APIハンドラ
func apiStatsHandler(w http.ResponseWriter, r *http.Request) {
	stats := calcStats()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// 景品一覧APIハンドラ（ローテーション情報付き）
func apiPrizesHandler(w http.ResponseWriter, r *http.Request) {
	prizeMu.RLock()
	ps := make([]Prize, len(prizes))
	copy(ps, prizes)
	nra := nextRotateAt
	lra := lastRotatedAt
	prizeMu.RUnlock()

	info := PrizesInfo{
		Prizes:              ps,
		NextRotationAt:      nra,
		LastRotatedAt:       lra,
		RotationIntervalSec: int(rotationInterval.Seconds()),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func main() {
	startPrizeRotation()

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/api/draw", apiDrawHandler)
	http.HandleFunc("/api/history", apiHistoryHandler)
	http.HandleFunc("/api/stats", apiStatsHandler)
	http.HandleFunc("/api/prizes", apiPrizesHandler)

	port := ":8081"
	fmt.Println("🎰 ガラガラポン抽選システム起動中...")
	fmt.Printf("🌐 http://localhost%s にアクセスしてください\n", port)
	fmt.Printf("🔄 当選確率は %v ごとに自動変更されます\n", rotationInterval)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
