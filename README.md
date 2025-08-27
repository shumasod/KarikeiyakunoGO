// README Generator for Provisional Contract Management System
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/template"
	"time"
)

// ProjectInfo represents the metadata for a project's README generation
type ProjectInfo struct {
	ProjectName    string    `json:"project_name"`
	Version        string    `json:"version"`
	Description    string    `json:"description"`
	Features       []string  `json:"features"`
	Installation   string    `json:"installation"`
	Usage          []string  `json:"usage"`
	License        string    `json:"license"`
	LastUpdated    time.Time `json:"last_updated"`
	ContactEmail   string    `json:"contact_email"`
	Disclaimer     string    `json:"disclaimer"`
	Repository     string    `json:"repository"`
}

// Config represents application configuration
type Config struct {
	OutputPath   string
	TemplatePath string
}

// READMEGenerator handles README file generation
type READMEGenerator struct {
	config   Config
	template *template.Template
}

const defaultREADMETemplate = `# {{.ProjectName}}

[![Version](https://img.shields.io/badge/version-{{.Version}}-blue.svg)]({{.Repository}})
[![License](https://img.shields.io/badge/license-{{.License}}-green.svg)](#license)

## 📖 概要

{{.Description}}

## ✨ 主な機能

{{range .Features -}}
- {{.}}
{{end}}

## 🚀 インストール方法

` + "```bash" + `
{{.Installation}}
` + "```" + `

## 💡 使用方法

{{range .Usage -}}
- {{.}}
{{end}}

## 📄 ライセンス

このプロジェクトは {{.License}} ライセンスの下で公開されています。

## ⚠️ 免責事項

{{.Disclaimer}}

## 📞 お問い合わせ

- Email: {{.ContactEmail}}
- Repository: {{.Repository}}

---

**最終更新日**: {{.LastUpdated.Format "2006年01月02日 15:04:05"}}
`

// NewREADMEGenerator creates a new README generator instance
func NewREADMEGenerator(config Config) (*READMEGenerator, error) {
	var tmpl *template.Template
	var err error

	if config.TemplatePath != "" {
		tmpl, err = template.ParseFiles(config.TemplatePath)
	} else {
		tmpl, err = template.New("readme").Parse(defaultREADMETemplate)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return &READMEGenerator{
		config:   config,
		template: tmpl,
	}, nil
}

// Generate creates a README.md file based on the provided project information
func (rg *READMEGenerator) Generate(info *ProjectInfo) error {
	outputPath := rg.config.OutputPath
	if outputPath == "" {
		outputPath = "README.md"
	}

	// Ensure directory exists
	if dir := filepath.Dir(outputPath); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", outputPath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Printf("Warning: failed to close file: %v", closeErr)
		}
	}()

	if err := rg.template.Execute(file, info); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// createProjectInfo returns the default project information
func createProjectInfo() *ProjectInfo {
	return &ProjectInfo{
		ProjectName: "仮契約管理システム",
		Version:     "v0.2.0",
		Description: "企業間の仮契約プロセスを管理するための包括的なツールです。契約書の作成からレビュー、承認、電子署名まで、一連のワークフローを効率的に自動化し、契約管理の透明性と追跡可能性を向上させます。",
		Features: []string{
			"🔧 契約テンプレートの管理と作成",
			"✍️ 電子署名機能（複数の署名プロバイダー対応）",
			"🔄 承認ワークフローの設定とカスタマイズ",
			"⏰ 契約期限の通知アラートとリマインダー",
			"📊 契約履歴の追跡と分析",
			"📱 PDFエクスポート機能（透かし対応）",
			"🔐 セキュアな文書管理と暗号化",
			"📈 ダッシュボードでの進捗可視化",
		},
		Installation: "go install github.com/yourcompany/provisional-contract@latest",
		Usage: []string{
			"**契約テンプレートの作成**: `contract template create <名前> --type=<契約種別>`",
			"**新規契約の作成**: `contract new --template=<テンプレート名> --parties=<当事者>`",
			"**契約の送信**: `contract send --id=<契約ID> --to=<メールアドレス> --message=<メッセージ>`",
			"**契約のステータス確認**: `contract status --id=<契約ID> --format=<json|table>`",
			"**契約のレビュー**: `contract review --id=<契約ID> --action=<approve|reject|request-changes>`",
		},
		License:      "MIT",
		LastUpdated:  time.Now(),
		ContactEmail: "support@provisional-contract.example.com",
		Repository:   "https://github.com/yourcompany/provisional-contract",
		Disclaimer:   "⚠️ このソフトウェアは開発中であり、機能や仕様が予告なく変更される場合があります。実際の法的契約においては必ず専門家の助言を求め、適切な法的レビューを実施してください。本ソフトウェアの使用により生じる一切の損害について、開発者は責任を負いません。",
	}
}

// validateProjectInfo validates the project information
func validateProjectInfo(info *ProjectInfo) error {
	if info.ProjectName == "" {
		return fmt.Errorf("project name is required")
	}
	if info.Version == "" {
		return fmt.Errorf("version is required")
	}
	if info.Description == "" {
		return fmt.Errorf("description is required")
	}
	if info.ContactEmail == "" {
		return fmt.Errorf("contact email is required")
	}
	return nil
}

func main() {
	// Initialize configuration
	config := Config{
		OutputPath: "README.md",
		// TemplatePath: "custom_template.tmpl", // Optional: use custom template
	}

	// Create README generator
	generator, err := NewREADMEGenerator(config)
	if err != nil {
		log.Fatalf("Failed to create README generator: %v", err)
	}

	// Create project information
	projectInfo := createProjectInfo()

	// Validate project information
	if err := validateProjectInfo(projectInfo); err != nil {
		log.Fatalf("Invalid project information: %v", err)
	}

	// Generate README file
	if err := generator.Generate(projectInfo); err != nil {
		log.Fatalf("Failed to generate README: %v", err)
	}

	fmt.Printf("✅ README.md ファイルが正常に生成されました。\n")
	fmt.Printf("📁 出力先: %s\n", config.OutputPath)
}
