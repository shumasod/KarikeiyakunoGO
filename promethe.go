package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTPリクエストの総数
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"path", "method", "status"},
	)

	// HTTPリクエストの処理時間を計測
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method"},
	)
)

func init() {
	// Prometheusにメトリクスを登録
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
}

// Prometheusメトリクスを収集するミドルウェア
func prometheusMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// レスポンスのステータスコードをキャプチャするためのResponseWriter
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// 実際のハンドラーを実行
		next(rw, r)

		// メトリクスを記録
		duration := time.Since(start).Seconds()
		httpRequestDuration.WithLabelValues(r.URL.Path, r.Method).Observe(duration)
		httpRequestsTotal.WithLabelValues(r.URL.Path, r.Method, fmt.Sprintf("%d", rw.statusCode)).Inc()
	}
}

// ResponseWriterのラッパー（ステータスコードをキャプチャするため）
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func main() {
	// ルートパスのハンドラーを設定（メトリクス収集付き）
	http.HandleFunc("/", prometheusMiddleware(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World!")
	}))

	// ヘルスチェックエンドポイント
	http.HandleFunc("/health", prometheusMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	}))

	// Prometheusメトリクスエンドポイント
	http.Handle("/metrics", promhttp.Handler())

	// サーバー起動のログ出力
	fmt.Println("サーバーを起動しています...")
	fmt.Println("http://localhost:8080 - メインページ")
	fmt.Println("http://localhost:8080/metrics - Prometheusメトリクス")
	fmt.Println("http://localhost:8080/health - ヘルスチェック")

	// ポート8080でサーバーを起動
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("サーバーの起動に失敗しました:", err)
	}
}
