// Package analyzer computes surge scores and detects bullish signals.
//
// Scoring model (最大 100 点):
//
//	[A] 騰落率スコア  (0–40 点): 終値比上昇率に比例。+5% で満点。
//	[B] 出来高スコア  (0–30 点): 3ヶ月平均比の出来高倍率に比例。4 倍で満点。
//	[C] 高値圏スコア  (0–20 点): 当日高値と現値の乖離が小さいほど高い。
//	[D] 新高値スコア  (0–10 点): 52 週高値更新・接近で加点。
package analyzer

import (
	"math"
	"sort"

	"tse-scanner/model"
)

// thresholds for signal generation
const (
	thresholdVolumeSpike  = 2.0  // 出来高急増と判定する倍率（平均の 2 倍以上）
	thresholdBigRise      = 5.0  // 大幅上昇と判定する騰落率（%）
	thresholdRise         = 3.0  // 上昇トレンドと判定する騰落率（%）
	thresholdNearHigh     = 0.98 // 当日高値の 98% 以上を「高値圏」と判定
	thresholdNear52W      = 0.99 // 52 週高値の 99% 以上を「新高値圏」と判定
	thresholdAboveOpen    = 0.0  // 寄り付き以上なら「陽線」
	maxPriceChangeForFull = 5.0  // 騰落率スコア満点となる上昇率（%）
	maxVolumeRatioForFull = 4.0  // 出来高スコア満点となる倍率
)

// Analyze returns surge candidates from the given quotes, sorted by SurgeScore desc.
// Quotes with Valid=false or SurgeScore below minScore are excluded.
func Analyze(quotes []model.Quote, minScore float64) []model.Candidate {
	candidates := make([]model.Candidate, 0, len(quotes))

	for _, q := range quotes {
		if !q.Valid {
			continue
		}
		volRatio := volumeRatio(q)
		score, signals := scoreQuote(q, volRatio)

		if score < minScore {
			continue
		}
		candidates = append(candidates, model.Candidate{
			Quote:       q,
			VolumeRatio: volRatio,
			SurgeScore:  score,
			Signals:     signals,
		})
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].SurgeScore > candidates[j].SurgeScore
	})
	return candidates
}

// scoreQuote computes the composite surge score and collects triggered signals.
func scoreQuote(q model.Quote, volRatio float64) (float64, []model.Signal) {
	var signals []model.Signal
	total := 0.0

	// ---- [A] 騰落率スコア (0–40 pt) ----
	if q.ChangePercent > 0 {
		priceScore := math.Min(q.ChangePercent/maxPriceChangeForFull, 1.0) * 40.0
		total += priceScore
		signals = append(signals, model.Signal{Label: "騰落率", Score: priceScore})

		if q.ChangePercent >= thresholdBigRise {
			signals = append(signals, model.Signal{Label: "🚀大幅上昇", Score: 0})
		} else if q.ChangePercent >= thresholdRise {
			signals = append(signals, model.Signal{Label: "📈上昇トレンド", Score: 0})
		}
	}

	// ---- [B] 出来高スコア (0–30 pt) ----
	if volRatio > 1.0 {
		volScore := math.Min((volRatio-1.0)/(maxVolumeRatioForFull-1.0), 1.0) * 30.0
		total += volScore
		signals = append(signals, model.Signal{Label: "出来高", Score: volScore})

		if volRatio >= thresholdVolumeSpike {
			signals = append(signals, model.Signal{Label: "⚡出来高急増", Score: 0})
		}
	}

	// ---- [C] 高値圏スコア (0–20 pt) ----
	if q.DayHigh > 0 && q.Price > 0 {
		proximity := q.Price / q.DayHigh // 1.0 = 当日最高値ぴったり
		if proximity >= thresholdNearHigh {
			highScore := proximity * 20.0
			total += highScore
			signals = append(signals, model.Signal{Label: "高値圏", Score: highScore})
			signals = append(signals, model.Signal{Label: "🔼高値圏推移", Score: 0})
		}
	}

	// ---- [D] 52 週高値スコア (0–10 pt) ----
	if q.WeekHigh52 > 0 && q.Price > 0 {
		ratio52 := q.Price / q.WeekHigh52
		if ratio52 >= thresholdNear52W {
			w52Score := ratio52 * 10.0
			total += w52Score
			signals = append(signals, model.Signal{Label: "52週高値", Score: w52Score})
			if ratio52 >= 1.0 {
				signals = append(signals, model.Signal{Label: "🏆52週新高値", Score: 0})
			} else {
				signals = append(signals, model.Signal{Label: "🎯52週高値圏", Score: 0})
			}
		}
	}

	return math.Min(total, 100.0), signals
}

// volumeRatio returns current volume / 3-month average volume.
// Returns 1.0 if average volume is unavailable (neutral).
func volumeRatio(q model.Quote) float64 {
	if q.AvgVolume3M <= 0 {
		return 1.0
	}
	return float64(q.Volume) / float64(q.AvgVolume3M)
}
