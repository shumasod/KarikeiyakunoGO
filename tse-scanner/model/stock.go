// Package model defines all shared data types for the TSE surge scanner.
package model

import "time"

// Stock represents a watchlist entry.
type Stock struct {
	Symbol string // e.g. "7203.T"
	Name   string // 日本語銘柄名
	Sector string // 業種
}

// Quote holds a real-time market snapshot for one stock.
type Quote struct {
	Symbol        string
	Name          string
	Sector        string
	Price         float64
	Change        float64
	ChangePercent float64 // %
	Volume        int64
	AvgVolume3M   int64   // 3ヶ月平均出来高
	DayHigh       float64
	DayLow        float64
	Open          float64
	PrevClose     float64
	WeekHigh52    float64
	WeekLow52     float64
	FetchedAt     time.Time
	Valid         bool // false if fetch failed
}

// Signal is one detected surge indicator.
type Signal struct {
	Label string  // 表示名（例: "出来高急増"）
	Score float64 // この指標が寄与するスコア
}

// Candidate is a Quote enriched with surge analysis results.
type Candidate struct {
	Quote
	VolumeRatio float64  // 出来高 / 3ヶ月平均出来高
	SurgeScore  float64  // 0–100 の急騰スコア
	Signals     []Signal // 発動したシグナル一覧
}

// DefaultWatchlist returns the built-in stock watchlist (東証プライム中心).
func DefaultWatchlist() []Stock {
	return []Stock{
		// 自動車
		{Symbol: "7203.T", Name: "トヨタ自動車", Sector: "自動車"},
		{Symbol: "7267.T", Name: "本田技研工業", Sector: "自動車"},
		{Symbol: "7201.T", Name: "日産自動車", Sector: "自動車"},
		{Symbol: "7270.T", Name: "SUBARU", Sector: "自動車"},
		// 電機・電子
		{Symbol: "6758.T", Name: "ソニーグループ", Sector: "電機"},
		{Symbol: "6752.T", Name: "パナソニックHD", Sector: "電機"},
		{Symbol: "6501.T", Name: "日立製作所", Sector: "電機"},
		{Symbol: "6503.T", Name: "三菱電機", Sector: "電機"},
		{Symbol: "6701.T", Name: "NEC", Sector: "電機"},
		// 半導体・精密
		{Symbol: "8035.T", Name: "東京エレクトロン", Sector: "半導体"},
		{Symbol: "6857.T", Name: "アドバンテスト", Sector: "半導体"},
		{Symbol: "4063.T", Name: "信越化学工業", Sector: "化学"},
		{Symbol: "6723.T", Name: "ルネサスエレクトロニクス", Sector: "半導体"},
		{Symbol: "6146.T", Name: "ディスコ", Sector: "半導体"},
		// IT・通信
		{Symbol: "9984.T", Name: "ソフトバンクグループ", Sector: "IT"},
		{Symbol: "9432.T", Name: "日本電信電話", Sector: "通信"},
		{Symbol: "9433.T", Name: "KDDI", Sector: "通信"},
		{Symbol: "9434.T", Name: "ソフトバンク", Sector: "通信"},
		{Symbol: "4689.T", Name: "LINEヤフー", Sector: "IT"},
		{Symbol: "3659.T", Name: "ネクソン", Sector: "ゲーム"},
		// 金融
		{Symbol: "8306.T", Name: "三菱UFJフィナンシャル", Sector: "銀行"},
		{Symbol: "8411.T", Name: "みずほフィナンシャル", Sector: "銀行"},
		{Symbol: "8316.T", Name: "三井住友フィナンシャル", Sector: "銀行"},
		{Symbol: "8604.T", Name: "野村ホールディングス", Sector: "証券"},
		{Symbol: "8591.T", Name: "オリックス", Sector: "金融"},
		// 小売
		{Symbol: "3382.T", Name: "セブン＆アイHD", Sector: "小売"},
		{Symbol: "8267.T", Name: "イオン", Sector: "小売"},
		{Symbol: "9843.T", Name: "ニトリHD", Sector: "小売"},
		{Symbol: "4452.T", Name: "花王", Sector: "日用品"},
		// ゲーム・エンタメ
		{Symbol: "7974.T", Name: "任天堂", Sector: "ゲーム"},
		{Symbol: "9766.T", Name: "コナミグループ", Sector: "ゲーム"},
		{Symbol: "7832.T", Name: "バンダイナムコHD", Sector: "ゲーム"},
		{Symbol: "9684.T", Name: "スクウェア・エニックス", Sector: "ゲーム"},
		// 医薬品
		{Symbol: "4568.T", Name: "第一三共", Sector: "医薬品"},
		{Symbol: "4519.T", Name: "中外製薬", Sector: "医薬品"},
		{Symbol: "4502.T", Name: "武田薬品工業", Sector: "医薬品"},
		{Symbol: "4578.T", Name: "大塚ホールディングス", Sector: "医薬品"},
		// 重工・機械
		{Symbol: "7011.T", Name: "三菱重工業", Sector: "重工"},
		{Symbol: "6326.T", Name: "クボタ", Sector: "機械"},
		{Symbol: "6301.T", Name: "小松製作所", Sector: "機械"},
		// 素材
		{Symbol: "3407.T", Name: "旭化成", Sector: "化学"},
		{Symbol: "4005.T", Name: "住友化学", Sector: "化学"},
		{Symbol: "5401.T", Name: "日本製鉄", Sector: "鉄鋼"},
		// エネルギー
		{Symbol: "5020.T", Name: "ENEOSホールディングス", Sector: "エネルギー"},
		// 食品
		{Symbol: "2914.T", Name: "日本たばこ産業", Sector: "食品"},
		{Symbol: "2802.T", Name: "味の素", Sector: "食品"},
		// 海運・交通
		{Symbol: "9101.T", Name: "日本郵船", Sector: "海運"},
		{Symbol: "9107.T", Name: "川崎汽船", Sector: "海運"},
		{Symbol: "9020.T", Name: "東日本旅客鉄道", Sector: "交通"},
		// 不動産
		{Symbol: "8801.T", Name: "三井不動産", Sector: "不動産"},
		{Symbol: "8802.T", Name: "三菱地所", Sector: "不動産"},
	}
}
