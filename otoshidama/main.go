
package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math/rand/v2"
	"net/http"
)

// お年玉データの構造体
type OtoshidamaData struct {
	Year   int    `json:"year"`
	Amount int    `json:"amount"`
	Giver  string `json:"giver"`
}

// 分析結果の構造体
type AnalysisResult struct {
	TotalAmount   int                `json:"total_amount"`
	AverageAmount float64            `json:"average_amount"`
	MaxAmount     int                `json:"max_amount"`
	MinAmount     int                `json:"min_amount"`
	Count         int                `json:"count"`
	YearlyData    []OtoshidamaData   `json:"yearly_data"`
	TopGivers     map[string]int     `json:"top_givers"`
}

// サンプルデータ
var sampleData = []OtoshidamaData{
	{2020, 5000, "おじいちゃん"},
	{2020, 3000, "おばあちゃん"},
	{2020, 2000, "叔父さん"},
	{2021, 5000, "おじいちゃん"},
	{2021, 3000, "おばあちゃん"},
	{2021, 3000, "叔父さん"},
	{2022, 10000, "おじいちゃん"},
	{2022, 5000, "おばあちゃん"},
	{2022, 3000, "叔父さん"},
	{2023, 10000, "おじいちゃん"},
	{2023, 5000, "おばあちゃん"},
	{2023, 5000, "叔父さん"},
	{2024, 10000, "おじいちゃん"},
	{2024, 5000, "おばあちゃん"},
	{2024, 5000, "叔父さん"},
}

// データを分析する関数
func analyzeData(data []OtoshidamaData) AnalysisResult {
	if len(data) == 0 {
		return AnalysisResult{}
	}

	total := 0
	max := data[0].Amount
	min := data[0].Amount
	giverTotals := make(map[string]int)

	for _, record := range data {
		total += record.Amount
		if record.Amount > max {
			max = record.Amount
		}
		if record.Amount < min {
			min = record.Amount
		}
		giverTotals[record.Giver] += record.Amount
	}

	average := float64(total) / float64(len(data))

	return AnalysisResult{
		TotalAmount:   total,
		AverageAmount: average,
		MaxAmount:     max,
		MinAmount:     min,
		Count:         len(data),
		YearlyData:    data,
		TopGivers:     giverTotals,
	}
}

// ボーナスお年玉を決定する関数
func getBonusOtoshidama() int {
	bonuses := []int{1000, 2000, 3000, 5000, 10000}
	return bonuses[rand.IntN(len(bonuses))]
}

// ホームページのハンドラ
func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>お年玉データ分析サイト</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        .header {
            text-align: center;
            color: white;
            margin-bottom: 40px;
            padding: 20px;
        }
        .header h1 {
            font-size: 3em;
            margin-bottom: 10px;
            text-shadow: 2px 2px 4px rgba(0,0,0,0.3);
        }
        .header p {
            font-size: 1.2em;
            opacity: 0.9;
        }
        .card {
            background: white;
            border-radius: 15px;
            padding: 30px;
            margin-bottom: 30px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.2);
        }
        .button-group {
            display: flex;
            gap: 20px;
            justify-content: center;
            flex-wrap: wrap;
        }
        button {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            padding: 15px 40px;
            font-size: 1.1em;
            border-radius: 50px;
            cursor: pointer;
            transition: transform 0.2s, box-shadow 0.2s;
            font-weight: bold;
        }
        button:hover {
            transform: translateY(-2px);
            box-shadow: 0 5px 20px rgba(102, 126, 234, 0.4);
        }
        .bonus-button {
            background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
            font-size: 1.3em;
            padding: 20px 50px;
        }
        .emoji {
            font-size: 2em;
            margin-right: 10px;
        }
        .feature-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin-top: 20px;
        }
        .feature-item {
            text-align: center;
            padding: 20px;
            background: #f8f9fa;
            border-radius: 10px;
        }
        .feature-item h3 {
            color: #667eea;
            margin-top: 10px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎍 お年玉データ分析サイト 🎍</h1>
            <p>あなたのお年玉データを分析して、ボーナスをゲット!</p>
        </div>

        <div class="card">
            <h2 style="text-align: center; margin-bottom: 30px;">📊 機能メニュー</h2>
            <div class="button-group">
                <button onclick="location.href='/analyze'">
                    <span class="emoji">📈</span>データ分析を見る
                </button>
                <button onclick="location.href='/api/data'" target="_blank">
                    <span class="emoji">💾</span>生データを見る (JSON)
                </button>
            </div>
        </div>

        <div class="card" style="text-align: center;">
            <h2 style="margin-bottom: 20px;">🎁 ボーナスお年玉をもらう!</h2>
            <p style="margin-bottom: 30px; color: #666;">データを見ると、ランダムでボーナスお年玉がもらえます!</p>
            <button class="bonus-button" onclick="getBonus()">
                <span class="emoji">🎊</span>ボーナスをもらう!
            </button>
            <div id="bonus-result" style="margin-top: 30px; font-size: 1.5em; font-weight: bold; color: #f5576c;"></div>
        </div>

        <div class="card">
            <h2 style="text-align: center; margin-bottom: 20px;">✨ このサイトでできること</h2>
            <div class="feature-grid">
                <div class="feature-item">
                    <div class="emoji">📊</div>
                    <h3>データ分析</h3>
                    <p>過去のお年玉データを詳しく分析</p>
                </div>
                <div class="feature-item">
                    <div class="emoji">💰</div>
                    <h3>統計情報</h3>
                    <p>合計・平均・最大・最小金額を表示</p>
                </div>
                <div class="feature-item">
                    <div class="emoji">👥</div>
                    <h3>贈り主分析</h3>
                    <p>誰から一番もらったかを表示</p>
                </div>
                <div class="feature-item">
                    <div class="emoji">🎁</div>
                    <h3>ボーナス</h3>
                    <p>ランダムでお年玉がもらえる!</p>
                </div>
            </div>
        </div>
    </div>

    <script>
        function getBonus() {
            fetch('/api/bonus')
                .then(response => response.json())
                .then(data => {
                    const resultDiv = document.getElementById('bonus-result');
                    resultDiv.innerHTML = '🎉 おめでとうございます!<br>' + data.amount.toLocaleString() + '円のボーナスをゲット!';
                    
                    // 紙吹雪アニメーション(簡易版)
                    for (let i = 0; i < 30; i++) {
                        createConfetti();
                    }
                });
        }

        function createConfetti() {
            const confetti = document.createElement('div');
            confetti.style.position = 'fixed';
            confetti.style.width = '10px';
            confetti.style.height = '10px';
            confetti.style.backgroundColor = ['#f093fb', '#f5576c', '#667eea', '#ffd700'][Math.floor(Math.random() * 4)];
            confetti.style.left = Math.random() * 100 + '%';
            confetti.style.top = '-10px';
            confetti.style.borderRadius = '50%';
            confetti.style.pointerEvents = 'none';
            confetti.style.zIndex = '1000';
            document.body.appendChild(confetti);

            let pos = -10;
            const fall = setInterval(() => {
                if (pos > window.innerHeight) {
                    clearInterval(fall);
                    confetti.remove();
                } else {
                    pos += 5;
                    confetti.style.top = pos + 'px';
                }
            }, 20);
        }
    </script>
</body>
</html>
`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, tmpl)
}

// 分析ページのハンドラ
func analyzeHandler(w http.ResponseWriter, r *http.Request) {
	result := analyzeData(sampleData)

	tmpl := `
<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>データ分析結果</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        .header {
            text-align: center;
            color: white;
            margin-bottom: 40px;
        }
        .card {
            background: white;
            border-radius: 15px;
            padding: 30px;
            margin-bottom: 30px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.2);
        }
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        .stat-box {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 20px;
            border-radius: 10px;
            text-align: center;
        }
        .stat-box h3 {
            font-size: 0.9em;
            margin-bottom: 10px;
            opacity: 0.9;
        }
        .stat-box p {
            font-size: 2em;
            font-weight: bold;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 20px;
        }
        th, td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #ddd;
        }
        th {
            background-color: #667eea;
            color: white;
        }
        tr:hover {
            background-color: #f5f5f5;
        }
        .back-button {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            padding: 15px 40px;
            font-size: 1.1em;
            border-radius: 50px;
            cursor: pointer;
            display: block;
            margin: 30px auto;
            font-weight: bold;
        }
        .back-button:hover {
            transform: translateY(-2px);
            box-shadow: 0 5px 20px rgba(102, 126, 234, 0.4);
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📊 お年玉データ分析結果</h1>
        </div>

        <div class="card">
            <h2>💰 統計サマリー</h2>
            <div class="stats-grid">
                <div class="stat-box">
                    <h3>合計金額</h3>
                    <p>¥{{.TotalAmount}}</p>
                </div>
                <div class="stat-box">
                    <h3>平均金額</h3>
                    <p>¥{{printf "%.0f" .AverageAmount}}</p>
                </div>
                <div class="stat-box">
                    <h3>最高額</h3>
                    <p>¥{{.MaxAmount}}</p>
                </div>
                <div class="stat-box">
                    <h3>最低額</h3>
                    <p>¥{{.MinAmount}}</p>
                </div>
                <div class="stat-box">
                    <h3>データ件数</h3>
                    <p>{{.Count}}件</p>
                </div>
            </div>
        </div>

        <div class="card">
            <h2>👥 贈り主別合計</h2>
            <table>
                <thead>
                    <tr>
                        <th>贈り主</th>
                        <th>合計金額</th>
                    </tr>
                </thead>
                <tbody>
                    {{range $giver, $amount := .TopGivers}}
                    <tr>
                        <td>{{$giver}}</td>
                        <td>¥{{$amount}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </div>

        <div class="card">
            <h2>📅 年別データ</h2>
            <table>
                <thead>
                    <tr>
                        <th>年</th>
                        <th>金額</th>
                        <th>贈り主</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .YearlyData}}
                    <tr>
                        <td>{{.Year}}年</td>
                        <td>¥{{.Amount}}</td>
                        <td>{{.Giver}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </div>

        <button class="back-button" onclick="location.href='/'">🏠 ホームに戻る</button>
    </div>
</body>
</html>
`
	t, err := template.New("analyze").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t.Execute(w, result)
}

// APIハンドラ - データ取得
func apiDataHandler(w http.ResponseWriter, r *http.Request) {
	result := analyzeData(sampleData)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// APIハンドラ - ボーナス取得
func apiBonusHandler(w http.ResponseWriter, r *http.Request) {
	bonus := getBonusOtoshidama()
	response := map[string]int{"amount": bonus}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	// ルーティング設定
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/analyze", analyzeHandler)
	http.HandleFunc("/api/data", apiDataHandler)
	http.HandleFunc("/api/bonus", apiBonusHandler)

	// サーバー起動
	port := ":8080"
	fmt.Printf("🚀 サーバーを起動しています...\n")
	fmt.Printf("🌐 ブラウザで http://localhost%s にアクセスしてください\n", port)
	fmt.Printf("📊 お年玉データ分析サイトへようこそ!\n\n")

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
