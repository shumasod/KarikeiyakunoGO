package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tse-scanner/analyzer"
	"tse-scanner/display"
	"tse-scanner/fetcher"
	"tse-scanner/model"
)

// バージョン情報は make build 時に -ldflags で注入される
var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

func main() {
	var (
		interval     = flag.Duration("interval", 60*time.Second, "更新間隔（例: 30s, 1m, 5m）")
		minScore     = flag.Float64("min-score", 20.0, "表示する最小急騰スコア（0–100）")
		topN         = flag.Int("top", 20, "最大表示件数")
		showVersion  = flag.Bool("version", false, "バージョン情報を表示")
	)
	flag.Parse()

	if *showVersion {
		fmt.Printf("tse-scanner %s (commit: %s, built: %s)\n", version, commit, buildDate)
		return
	}

	if *interval < 10*time.Second {
		log.Fatal("interval は 10 秒以上に設定してください（レート制限回避のため）")
	}

	watchlist := model.DefaultWatchlist()
	client := fetcher.New()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown on SIGINT / SIGTERM
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		fmt.Println("\n終了中...")
		cancel()
	}()

	scan := func() {
		quotes, err := client.FetchQuotes(ctx, watchlist)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("データ取得エラー: %v", err)
			return
		}

		candidates := analyzer.Analyze(quotes, *minScore)
		if len(candidates) > *topN {
			candidates = candidates[:*topN]
		}
		display.Render(candidates, time.Now(), *interval, len(watchlist))
	}

	scan() // 初回即時実行

	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			scan()
		case <-ctx.Done():
			fmt.Println("終了しました。")
			return
		}
	}
}
