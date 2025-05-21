# KarikeiyakunoGO

package main

import (
	"fmt"
	"os"
	"text/template"
	"time"
)

// 仮契約の情報を格納する構造体
type ContractInfo struct {
	ProjectName    string
	Version        string
	Description    string
	Features       []string
	Installation   string
	Usage          []string
	License        string
	LastUpdated    string
	ContactEmail   string
	Disclaimer     string
}

func main() {
	// 現在の日時を取得
	currentTime := time.Now().Format("2006年01月02日")
	
	// 仮契約のプロジェクト情報
	contractInfo := ContractInfo{
		ProjectName:    "仮契約管理システム",
		Version:        "v0.1.0",
		Description:    "このシステムは企業間の仮契約プロセスを管理するためのツールです。契約書の作成、レビュー、承認、署名までの一連のワークフローを自動化します。",
		Features: []string{
			"契約テンプレートの管理",
			"電子署名機能",
			"承認ワークフローの設定",
			"契約期限の通知アラート",
			"契約履歴の追跡",
			"PDFエクスポート機能",
		},
		Installation: "go get github.com/yourcompany/provisional-contract",
		Usage: []string{
			"契約テンプレートの作成: `contract template create <名前>`",
			"新規契約の作成: `contract new --template=<テンプレート名>`",
			"契約の送信: `contract send --id=<契約ID> --to=<メールアドレス>`",
			"契約のステータス確認: `contract status --id=<契約ID>`",
		},
		License:       "MIT",
		LastUpdated:   currentTime,
		ContactEmail:  "support@example.com",
		Disclaimer:    "このソフトウェアは開発中であり、予告なく変更される場合があります。実際の法的契約には専門家の助言を求めてください。",
	}

	// READMEテンプレートの定義
	const readmeTemplate = `# {{.ProjectName}}

## バージョン
{{.Version}}

## 概要
{{.Description}}

## 主な機能
{{range .Features}}* {{.}}
{{end}}

## インストール方法
` + "```" + `
{{.Installation}}
` + "```" + `

## 使用方法
{{range .Usage}}* {{.}}
{{end}}

## ライセンス
{{.License}}

## 免責事項
{{.Disclaimer}}

---
最終更新日: {{.LastUpdated}}
お問い合わせ: {{.ContactEmail}}
`

	// テンプレートを解析
	tmpl, err := template.New("readme").Parse(readmeTemplate)
	if err != nil {
		fmt.Printf("テンプレートの解析エラー: %v\n", err)
		return
	}

	// README.mdファイルを作成
	file, err := os.Create("README.md")
	if err != nil {
		fmt.Printf("ファイル作成エラー: %v\n", err)
		return
	}
	defer file.Close()

	// テンプレートを実行してファイルに書き込み
	err = tmpl.Execute(file, contractInfo)
	if err != nil {
		fmt.Printf("テンプレート実行エラー: %v\n", err)
		return
	}

	fmt.Println("README.mdファイルが正常に生成されました。")
}
