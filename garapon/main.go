package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"garapon/handler"
	"garapon/service"
)

const (
	rotationInterval = 30 * time.Second
	listenAddr       = ":8081"
)

func main() {
	svc := service.New(rotationInterval)
	h := handler.New(svc)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	fmt.Println("🎰 ガラガラポン抽選システム起動中...")
	fmt.Printf("🌐 http://localhost%s にアクセスしてください\n", listenAddr)
	fmt.Printf("🔄 当選確率は %v ごとに自動変更されます\n", rotationInterval)

	if err := http.ListenAndServe(listenAddr, mux); err != nil {
		log.Fatalf("サーバー起動エラー: %v", err)
	}
}
