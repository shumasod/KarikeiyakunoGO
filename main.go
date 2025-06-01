package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// ルートパスのハンドラーを設定
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World!")
	})

	// サーバー起動のログ出力
	fmt.Println("サーバーを起動しています...")
	fmt.Println("http://localhost:8080 でアクセス可能です")

	// ポート8080でサーバーを起動
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("サーバーの起動に失敗しました:", err)
	}
}
