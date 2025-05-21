package main

import (
    "fmt"
    "log"
    "net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "こんにちは, %s へようこそ!", r.URL.Path[1:])
}

func main() {
    http.HandleFunc("/", handler)
    log.Println("サーバーを起動します。http://localhost:8080/ にアクセスしてください")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
