
package main

import (
	"fmt"
	"strings"
)

// 商品構造体
type Product struct {
	Name  string
	Price int
	Rating float64
}

// 評価を判定する関数
func judgeQuality(rating float64) string {
	switch {
	case rating >= 4.5:
		return "最高品質"
	case rating >= 4.0:
		return "優良品"
	case rating >= 3.0:
		return "標準品"
	case rating >= 2.0:
		return "改善必要"
	default:
		return "低品質"
	}
}

// 価格帯を判定する関数
func judgePriceRange(price int) string {
	switch {
	case price >= 10000:
		return "高価格帯"
	case price >= 5000:
		return "中価格帯"
	case price >= 1000:
		return "低価格帯"
	default:
		return "格安"
	}
}

// コストパフォーマンスを評価する関数
func evaluateCostPerformance(price int, rating float64) string {
	score := rating / (float64(price) / 1000)
	
	switch {
	case score >= 2.0:
		return "コスパ抜群！"
	case score >= 1.0:
		return "コスパ良好"
	case score >= 0.5:
		return "コスパ普通"
	default:
		return "コスパ悪い"
	}
}

// 商品を品定めする関数
func evaluateProduct(p Product) {
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("【商品名】 %s\n", p.Name)
	fmt.Printf("【価格】   ¥%d\n", p.Price)
	fmt.Printf("【評価】   %.1f / 5.0\n", p.Rating)
	fmt.Println(strings.Repeat("-", 50))
	
	quality := judgeQuality(p.Rating)
	priceRange := judgePriceRange(p.Price)
	costPerf := evaluateCostPerformance(p.Price, p.Rating)
	
	fmt.Printf("品質評価:     %s\n", quality)
	fmt.Printf("価格帯:       %s\n", priceRange)
	fmt.Printf("コスパ評価:   %s\n", costPerf)
	
	// 総合判定
	if p.Rating >= 4.0 && costPerf == "コスパ抜群！" {
		fmt.Println("\n★★★ おすすめ商品！ ★★★")
	} else if p.Rating >= 3.5 {
		fmt.Println("\n☆ 購入検討の価値あり")
	}
	
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()
}

func main() {
	fmt.Println("=== 商品品定めシステム ===\n")
	
	// サンプル商品データ
	products := []Product{
		{Name: "高級ヘッドホン", Price: 15000, Rating: 4.8},
		{Name: "ワイヤレスマウス", Price: 3000, Rating: 4.2},
		{Name: "USBケーブル", Price: 500, Rating: 3.5},
		{Name: "安物キーボード", Price: 2000, Rating: 2.1},
	}
	
	// 各商品を品定め
	for _, product := range products {
		evaluateProduct(product)
	}
	
	// ユーザー入力例
	fmt.Println("=== カスタム商品の評価 ===")
	customProduct := Product{
		Name:   "Bluetoothスピーカー",
		Price:  8000,
		Rating: 4.5,
	}
	evaluateProduct(customProduct)
}


**このコードの特徴:**

1. **構造体**: 商品情報を管理
1. **評価関数**:

- 品質判定（評価点数に基づく）
- 価格帯判定
- コストパフォーマンス評価

1. **総合評価**: 複数の要素を組み合わせた判定
1. **見やすい出力**: 罫線と記号で整形

実行すると、複数の商品を自動的に品定めして、おすすめ度を表示します！​​​​​​​​​​​​​​​​