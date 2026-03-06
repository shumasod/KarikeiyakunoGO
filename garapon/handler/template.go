package handler

// indexHTML is the single-page application served at GET /.
// It communicates with the backend exclusively via the JSON API.
const indexHTML = `<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>🎰 商店街ガラガラポン抽選会</title>
    <style>
        *{margin:0;padding:0;box-sizing:border-box;}
        body{font-family:'Hiragino Kaku Gothic Pro','Meiryo',sans-serif;
             background:linear-gradient(135deg,#1a1a2e 0%,#16213e 50%,#0f3460 100%);
             min-height:100vh;color:white;overflow-x:hidden;}
        .header{text-align:center;padding:30px 20px 10px;
                background:linear-gradient(180deg,rgba(255,200,0,0.15) 0%,transparent 100%);}
        .header h1{font-size:2.5em;color:#FFD700;letter-spacing:3px;margin-bottom:8px;
                   text-shadow:0 0 20px rgba(255,215,0,0.6),2px 2px 4px rgba(0,0,0,0.8);}
        .header p{color:#aaa;font-size:1em;}
        .main-content{display:flex;flex-wrap:wrap;gap:30px;padding:30px;
                      max-width:1200px;margin:0 auto;justify-content:center;}

        /* ---- Machine ---- */
        .machine-section{flex:0 0 auto;display:flex;flex-direction:column;align-items:center;}
        .machine-wrapper{position:relative;width:320px;}
        .drum-container{position:relative;width:280px;height:280px;margin:0 auto;}
        .drum{width:100%;height:100%;border-radius:50%;
              background:radial-gradient(circle at 35% 35%,#888,#444 60%,#222);
              border:8px solid #666;overflow:hidden;position:relative;
              box-shadow:0 0 0 4px #888,0 0 30px rgba(0,0,0,0.8),inset 0 0 40px rgba(0,0,0,0.5);}
        .drum.spinning{animation:drumSpin 0.2s linear infinite;}
        @keyframes drumSpin{from{transform:rotate(0deg);}to{transform:rotate(360deg);}}
        .drum-grid{position:absolute;inset:10px;border-radius:50%;
                   background:repeating-linear-gradient(0deg,transparent,transparent 18px,rgba(255,255,255,0.08) 18px,rgba(255,255,255,0.08) 19px),
                               repeating-linear-gradient(90deg,transparent,transparent 18px,rgba(255,255,255,0.08) 18px,rgba(255,255,255,0.08) 19px);}
        .drum-balls{position:absolute;inset:20px;border-radius:50%;overflow:hidden;}
        .mini-ball{position:absolute;width:28px;height:28px;border-radius:50%;
                   box-shadow:inset -3px -3px 6px rgba(0,0,0,0.4),inset 2px 2px 4px rgba(255,255,255,0.3);}
        .handle-area{display:flex;justify-content:flex-end;margin-top:-40px;padding-right:10px;}
        .handle{width:60px;height:120px;}
        .handle-bar{width:12px;height:80px;background:linear-gradient(90deg,#888,#ccc,#888);
                    border-radius:6px;margin:0 auto;box-shadow:2px 2px 6px rgba(0,0,0,0.5);}
        .handle-knob{width:36px;height:36px;border-radius:50%;
                     background:radial-gradient(circle at 35% 35%,#ffcc00,#cc8800);
                     border:3px solid #aa6600;margin:0 auto;cursor:pointer;
                     box-shadow:0 4px 8px rgba(0,0,0,0.5);transition:transform 0.1s;}
        .handle-knob:hover{transform:scale(1.1);}
        .handle-knob:active{transform:scale(0.95);}
        .outlet{width:100px;height:50px;margin:10px auto 0;
                background:linear-gradient(180deg,#333,#555);
                border-radius:0 0 20px 20px;border:4px solid #666;border-top:none;
                position:relative;display:flex;align-items:center;justify-content:center;overflow:visible;}
        .outlet-label{font-size:0.7em;color:#aaa;letter-spacing:1px;}
        .result-ball{width:80px;height:80px;border-radius:50%;
                     position:absolute;top:-120px;left:50%;transform:translateX(-50%);display:none;
                     box-shadow:inset -8px -8px 15px rgba(0,0,0,0.4),inset 4px 4px 10px rgba(255,255,255,0.4),0 8px 20px rgba(0,0,0,0.5);
                     animation:ballDrop 0.6s ease-out forwards;}
        @keyframes ballDrop{
            0%  {top:-140px;opacity:0;transform:translateX(-50%) scale(0.5);}
            50% {top:-100px;opacity:1;transform:translateX(-50%) scale(1.1);}
            100%{top:-110px;opacity:1;transform:translateX(-50%) scale(1);}}
        .result-ball.show{display:block;}
        .machine-base{width:300px;height:20px;background:linear-gradient(180deg,#888,#555);
                      border-radius:0 0 10px 10px;margin:0 auto;box-shadow:0 6px 12px rgba(0,0,0,0.5);}
        .draw-btn{margin-top:30px;padding:18px 60px;font-size:1.4em;font-weight:bold;
                  background:linear-gradient(135deg,#FFD700,#FF8C00);color:#1a1a1a;
                  border:none;border-radius:50px;cursor:pointer;letter-spacing:2px;
                  box-shadow:0 6px 20px rgba(255,140,0,0.5);transition:all 0.2s;}
        .draw-btn:hover:not(:disabled){transform:translateY(-3px);box-shadow:0 10px 30px rgba(255,140,0,0.7);}
        .draw-btn:active:not(:disabled){transform:translateY(0);}
        .draw-btn:disabled{opacity:0.6;cursor:not-allowed;}
        .stats-bar{display:flex;gap:15px;flex-wrap:wrap;margin-top:15px;justify-content:center;}
        .stat-chip{background:rgba(255,255,255,0.08);padding:6px 14px;border-radius:20px;
                   font-size:0.85em;color:#ccc;}
        .stat-chip span{color:#FFD700;font-weight:bold;}

        /* ---- Result panel ---- */
        .result-section{flex:1;min-width:300px;}
        .result-panel{background:rgba(255,255,255,0.05);border:1px solid rgba(255,255,255,0.1);
                      border-radius:20px;padding:30px;margin-bottom:20px;min-height:200px;
                      display:flex;flex-direction:column;align-items:center;justify-content:center;
                      text-align:center;transition:all 0.3s;}
        .result-panel.highlight{border-color:rgba(255,215,0,0.5);background:rgba(255,215,0,0.05);
                                 box-shadow:0 0 30px rgba(255,215,0,0.2);}
        .result-grade{font-size:3em;font-weight:bold;margin-bottom:10px;
                      opacity:0;transform:scale(0.5);
                      transition:all 0.4s cubic-bezier(0.175,0.885,0.32,1.275);}
        .result-grade.show{opacity:1;transform:scale(1);}
        .result-name{font-size:1.4em;color:#ddd;margin-bottom:8px;}
        .result-desc{font-size:1.1em;color:#aaa;}
        .wait-msg{color:#555;font-size:1.1em;}

        /* ---- Prize table ---- */
        .prize-table-section{background:rgba(255,255,255,0.05);border:1px solid rgba(255,255,255,0.1);
                              border-radius:20px;padding:25px;margin-bottom:20px;transition:background 0.5s;}
        .prize-table-section.flash{background:rgba(255,215,0,0.1);}
        .prize-table-header{display:flex;align-items:center;justify-content:space-between;margin-bottom:8px;}
        .prize-table-header h2{color:#FFD700;font-size:1.2em;letter-spacing:2px;}
        .live-badge{font-size:0.65em;background:#cc2200;color:white;padding:2px 7px;
                    border-radius:4px;letter-spacing:1px;animation:livePulse 1.2s ease-in-out infinite;}
        @keyframes livePulse{0%,100%{opacity:1;}50%{opacity:0.4;}}
        .rotation-timer{font-size:0.82em;color:#888;margin-bottom:12px;display:flex;align-items:center;gap:6px;}
        .countdown-num{color:#FFD700;font-weight:bold;font-size:1.15em;min-width:24px;
                       display:inline-block;text-align:center;}
        .countdown-num.soon{color:#FF6644;animation:urgentPulse 0.5s ease-in-out infinite;}
        @keyframes urgentPulse{0%,100%{transform:scale(1);}50%{transform:scale(1.2);}}
        .prize-row{display:flex;align-items:center;gap:12px;padding:8px 0;
                   border-bottom:1px solid rgba(255,255,255,0.05);}
        .prize-row:last-child{border-bottom:none;}
        .ball-icon{width:24px;height:24px;border-radius:50%;flex-shrink:0;
                   box-shadow:inset -2px -2px 4px rgba(0,0,0,0.4),inset 1px 1px 3px rgba(255,255,255,0.3);}
        .prize-grade-label{font-weight:bold;min-width:50px;font-size:0.95em;}
        .prize-prize-name{color:#ddd;flex:1;font-size:0.9em;}
        .prize-prob{color:#888;font-size:0.8em;min-width:50px;text-align:right;}

        /* ---- History ---- */
        .history-section{background:rgba(255,255,255,0.05);border:1px solid rgba(255,255,255,0.1);
                         border-radius:20px;padding:25px;}
        .history-section h2{color:#FFD700;margin-bottom:15px;font-size:1.2em;letter-spacing:2px;}
        #history-list{max-height:300px;overflow-y:auto;}
        .history-item{display:flex;align-items:center;gap:10px;padding:8px 0;
                      border-bottom:1px solid rgba(255,255,255,0.05);animation:fadeIn 0.3s ease;}
        @keyframes fadeIn{from{opacity:0;transform:translateY(-10px);}to{opacity:1;transform:translateY(0);}}
        .history-num{color:#666;font-size:0.8em;min-width:40px;}
        .history-grade{font-weight:bold;font-size:0.9em;min-width:60px;}
        .history-prize{color:#aaa;font-size:0.85em;flex:1;}
        .history-time{color:#555;font-size:0.75em;}
        .no-history{color:#555;font-size:0.9em;text-align:center;padding:20px 0;}

        /* ---- Grade colors ---- */
        .grade-tokutou{color:#FFD700;}.grade-ittou{color:#FF6666;}.grade-nittou{color:#6699FF;}
        .grade-santou{color:#66CC66;}.grade-yontou{color:#FFDD44;}.grade-hazure{color:#AAAAAA;}

        /* ---- Confetti ---- */
        .confetti-container{position:fixed;top:0;left:0;width:100%;height:100%;pointer-events:none;z-index:1000;}
        .confetti-piece{position:absolute;top:-20px;animation:confettiFall linear forwards;}
        @keyframes confettiFall{to{top:110vh;transform:rotate(720deg);}}

        @media(max-width:700px){
            .header h1{font-size:1.8em;}
            .main-content{padding:15px;gap:20px;}
            .draw-btn{font-size:1.1em;padding:14px 40px;}}
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
'use strict';

const GRADE_CLASS = {
    "特等":"grade-tokutou","1等":"grade-ittou","2等":"grade-nittou",
    "3等":"grade-santou","4等":"grade-yontou","参加賞":"grade-hazure"
};

let currentPrizes = [];
let nextRotationAt = null;
let isDrawing = false;
let totalDraws = 0;

/* ---------- helpers ---------- */
function lighten(hex) {
    const r = parseInt(hex.slice(1,3),16);
    const g = parseInt(hex.slice(3,5),16);
    const b = parseInt(hex.slice(5,7),16);
    return ` + "`" + `rgb(${Math.min(255,r+80)},${Math.min(255,g+80)},${Math.min(255,b+80)})` + "`" + `;
}
function weightToProb(w) { return (w / 10).toFixed(1) + '%'; }
function gradeClass(grade) { return GRADE_CLASS[grade] || 'grade-hazure'; }

/* ---------- API ---------- */
async function apiFetch(path, options) {
    const res = await fetch(path, options);
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || res.statusText);
    return data;
}

/* ---------- Prizes polling ---------- */
async function fetchPrizes() {
    try {
        const info = await apiFetch('/api/prizes');
        nextRotationAt = new Date(info.next_rotation_at);
        const changed = currentPrizes.length > 0 &&
            currentPrizes.some((p, i) => p.weight !== info.prizes[i].weight);
        currentPrizes = info.prizes;
        renderPrizeTable();
        populateDrum();
        if (changed) flashPrizeTable();
    } catch(e) { console.error('景品取得エラー:', e); }
}

function renderPrizeTable() {
    if (!currentPrizes.length) return;
    document.getElementById('prizeTable').innerHTML =
        currentPrizes.map(p => ` + "`" + `
        <div class="prize-row">
            <div class="ball-icon" style="background:radial-gradient(circle at 35% 35%,${lighten(p.ball.hex)},${p.ball.hex} 70%);"></div>
            <span class="prize-grade-label ${gradeClass(p.grade)}">${p.grade}</span>
            <span class="prize-prize-name">${p.description}</span>
            <span class="prize-prob">${weightToProb(p.weight)}</span>
        </div>` + "`" + `).join('');
}

function populateDrum() {
    if (!currentPrizes.length) return;
    const colors = currentPrizes.map(p => p.ball.hex);
    const positions = [
        {top:'15%',left:'20%'},{top:'15%',left:'55%'},{top:'30%',left:'10%'},
        {top:'30%',left:'40%'},{top:'30%',left:'68%'},{top:'50%',left:'15%'},
        {top:'50%',left:'45%'},{top:'50%',left:'72%'},{top:'65%',left:'25%'},
        {top:'65%',left:'55%'},{top:'78%',left:'15%'},{top:'78%',left:'42%'},
        {top:'78%',left:'65%'},{top:'20%',left:'75%'},{top:'42%',left:'30%'},
    ];
    document.getElementById('drumBalls').innerHTML = positions.map((pos,i) => {
        const c = colors[i % colors.length];
        return ` + "`" + `<div class="mini-ball" style="top:${pos.top};left:${pos.left};background:radial-gradient(circle at 35% 35%,${lighten(c)},${c} 70%);"></div>` + "`" + `;
    }).join('');
}

function flashPrizeTable() {
    const s = document.getElementById('prizeTableSection');
    s.classList.remove('flash');
    void s.offsetWidth;
    s.classList.add('flash');
    setTimeout(() => s.classList.remove('flash'), 800);
}

/* ---------- Countdown ---------- */
function updateCountdown() {
    if (!nextRotationAt) return;
    const remaining = Math.max(0, Math.ceil((nextRotationAt.getTime() - Date.now()) / 1000));
    const el = document.getElementById('countdown');
    if (!el) return;
    el.textContent = remaining;
    el.className = 'countdown-num' + (remaining <= 5 ? ' soon' : '');
    if (remaining === 0) setTimeout(fetchPrizes, 600);
}

/* ---------- Draw ---------- */
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
        result = await apiFetch('/api/draw');
    } catch(e) {
        resultPanel.innerHTML = ` + "`" + `<p style="color:#f66;">エラー: ${e.message}</p>` + "`" + `;
        drum.classList.remove('spinning');
        btn.disabled = false;
        btn.textContent = '🎲 ガラガラ回す！';
        isDrawing = false;
        return;
    }

    setTimeout(() => {
        drum.classList.remove('spinning');
        const ballColor = result.prize.ball.hex;
        resultBall.style.background = ` + "`" + `radial-gradient(circle at 35% 35%,${lighten(ballColor)},${ballColor} 70%)` + "`" + `;
        resultBall.classList.add('show');

        const grade = result.prize.grade;
        resultPanel.classList.add('highlight');
        resultPanel.innerHTML = ` + "`" + `
            <div class="result-grade ${gradeClass(grade)}" id="rg">${grade}</div>
            <div class="result-name">${result.prize.name}</div>
            <div class="result-desc">${result.prize.description}</div>` + "`" + `;
        setTimeout(() => { const rg=document.getElementById('rg'); if(rg) rg.classList.add('show'); }, 50);

        if (['特等','1等','2等'].includes(grade)) {
            launchConfetti(grade==='特等'?80:grade==='1等'?50:30);
        }

        totalDraws++;
        document.getElementById('totalDraws').textContent = totalDraws;
        addHistoryItem(result);
        btn.disabled = false;
        btn.textContent = '🎲 もう一度回す！';
        isDrawing = false;
    }, 1500);
}

/* ---------- History ---------- */
function addHistoryItem(result) {
    const list = document.getElementById('history-list');
    const noHist = list.querySelector('.no-history');
    if (noHist) noHist.remove();

    const grade = result.prize.grade;
    const ballColor = result.prize.ball.hex;
    const now = new Date().toLocaleTimeString('ja-JP');
    const item = document.createElement('div');
    item.className = 'history-item';
    item.innerHTML = ` + "`" + `
        <div style="width:16px;height:16px;flex-shrink:0;border-radius:50%;
             background:radial-gradient(circle at 35% 35%,${lighten(ballColor)},${ballColor} 70%);"></div>
        <span class="history-num">#${result.ticket_num}</span>
        <span class="history-grade ${gradeClass(grade)}">${grade}</span>
        <span class="history-prize">${result.prize.description}</span>
        <span class="history-time">${now}</span>` + "`" + `;
    list.insertBefore(item, list.firstChild);
    const items = list.querySelectorAll('.history-item');
    if (items.length > 20) items[items.length-1].remove();
}

/* ---------- Confetti ---------- */
function launchConfetti(count) {
    const container = document.getElementById('confettiContainer');
    const colors = ['#FFD700','#FF6666','#6699FF','#66CC66','#FF99CC','#FFCC44'];
    for (let i = 0; i < count; i++) {
        setTimeout(() => {
            const p = document.createElement('div');
            p.className = 'confetti-piece';
            const size = 6 + Math.random() * 10;
            p.style.cssText = ` + "`" + `
                left:${Math.random()*100}%;width:${size}px;height:${size}px;
                background:${colors[Math.floor(Math.random()*colors.length)]};
                border-radius:${Math.random()>0.5?'50%':'2px'};
                animation-duration:${1.5+Math.random()*2}s;
                animation-delay:${Math.random()*0.5}s;
                opacity:${0.6+Math.random()*0.4};` + "`" + `;
            container.appendChild(p);
            setTimeout(() => p.remove(), 4000);
        }, i * 30);
    }
}

/* ---------- Bootstrap ---------- */
setInterval(updateCountdown, 1000);
setInterval(fetchPrizes, 5000);
fetchPrizes();
</script>
</body>
</html>`
