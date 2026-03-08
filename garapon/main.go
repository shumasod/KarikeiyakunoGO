package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"garapon/handler"
	"garapon/service"
)

// バージョン情報は make build 時に -ldflags で注入される
var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

const (
	rotationInterval = 30 * time.Second
	listenAddr       = ":8081"
)

func main() {
	showVersion := flag.Bool("version", false, "バージョン情報を表示して終了")
	flag.Parse()

	if *showVersion {
		fmt.Printf("garapon %s (commit: %s, built: %s)\n", version, commit, buildDate)
		return
	}

	svc := service.New(rotationInterval)
	h := handler.New(svc)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	fmt.Printf("🎰 ガラガラポン抽選システム v%s 起動中...\n", version)
	fmt.Printf("🌐 http://localhost%s にアクセスしてください\n", listenAddr)
	fmt.Printf("🔄 当選確率は %v ごとに自動変更されます\n", rotationInterval)

	if err := http.ListenAndServe(listenAddr, mux); err != nil {
		log.Fatalf("サーバー起動エラー: %v", err)
	}
}
