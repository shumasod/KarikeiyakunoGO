# 🎍 お年玉データ分析サイト

Go言語で作成した、お年玉データを分析してボーナスがもらえる楽しいWebアプリケーションです！

## 🌟 機能

- 📊 **データ分析**: 過去のお年玉データを詳しく分析
- 💰 **統計情報**: 合計・平均・最大・最小金額を表示
- 👥 **贈り主分析**: 誰から一番もらったかを確認
- 🎁 **ボーナス機能**: ランダムでお年玉がもらえる！
- 🔌 **REST API**: JSONデータとしてもアクセス可能

## 🚀 起動方法

### 1. 前提条件
- Go 1.21以上がインストールされていること

### 2. プログラムの実行

```bash
# プロジェクトディレクトリに移動
cd otoshidama-analyzer

# プログラムを実行
go run main.go
```

### 3. アクセス方法

ブラウザで以下のURLにアクセスしてください：
```
http://localhost:8080
```

## 📖 使い方

1. **ホームページ**: メインメニューから機能を選択
2. **データ分析を見る**: 詳細な統計情報と分析結果を表示
3. **ボーナスをもらう**: ボタンをクリックするとランダムで1,000円〜10,000円がもらえます！
4. **生データを見る**: JSON形式でデータを確認

## 🔌 API エンドポイント

- `GET /` - ホームページ
- `GET /analyze` - 分析結果ページ
- `GET /api/data` - JSON形式でデータ取得
- `GET /api/bonus` - ランダムボーナス取得

## 🎨 技術スタック

- **言語**: Go 1.21
- **Webフレームワーク**: 標準ライブラリ (net/http)
- **テンプレート**: html/template
- **フロントエンド**: HTML5, CSS3, JavaScript

## 📝 コードの説明

### 主要な構造体

```go
// お年玉データ
type OtoshidamaData struct {
    Year   int
    Amount int
    Giver  string
}

// 分析結果
type AnalysisResult struct {
    TotalAmount   int
    AverageAmount float64
    MaxAmount     int
    MinAmount     int
    Count         int
    YearlyData    []OtoshidamaData
    TopGivers     map[string]int
}
```

### 主要な関数

- `analyzeData()`: データを分析して統計情報を計算
- `getBonusOtoshidama()`: ランダムでボーナス金額を決定
- `homeHandler()`: ホームページを表示
- `analyzeHandler()`: 分析結果ページを表示
- `apiDataHandler()`: JSON APIでデータを返す
- `apiBonusHandler()`: ボーナス金額をJSONで返す

## 🎓 初心者向けポイント

### Goの基本を学べる要素

1. **構造体 (struct)**: データを整理して管理
2. **関数**: 処理を分割して再利用可能に
3. **Webサーバー**: `net/http`パッケージでHTTPサーバーを作成
4. **テンプレート**: HTMLテンプレートを使ったページ生成
5. **JSON**: データをJSON形式で送受信
6. **ルーティング**: URLごとに異なる処理を割り当て

## 🔧 カスタマイズ方法

### データを変更したい場合

`main.go`の`sampleData`変数を編集してください：

```go
var sampleData = []OtoshidamaData{
    {2025, 15000, "お父さん"},
    {2025, 10000, "お母さん"},
    // ... 自由に追加
}
```

### ポート番号を変更したい場合

`main()`関数内の`port`変数を変更してください：

```go
port := ":3000"  // 3000番ポートに変更
```

## 📚 学習リソース

- [Go公式サイト](https://go.dev/)
- [A Tour of Go](https://go.dev/tour/) - Go言語の基本を学べるインタラクティブなチュートリアル
- [Go by Example](https://gobyexample.com/) - 実例で学ぶGo言語

## 🎉 楽しみ方

1. データ分析ページで過去のお年玉を振り返る
2. ボーナス機能で何度もチャレンジ！
3. コードを読んで、Go言語の書き方を学ぶ
4. 自分でデータを追加してカスタマイズ

お年玉データ分析を楽しんでください！🎊
