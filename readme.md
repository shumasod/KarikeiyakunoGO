いいですね。
Shuma向けに「**API + IaC連携ツール**」として、実務でそのまま使える構成を考えます。

---

# 🎯 作るもの

## 🔥 Infrastructure Control API

**目的：**

* AWS / Kubernetes の状態をAPI経由で取得
* CLIからも操作可能
* DBに監査ログ保存
* 将来的にTerraform実行連携も可能

---

# 🏗 アーキテクチャ

* Web: Gin
* DB: GORM
* CLI: Cobra
* Config: Viper
* Log: Zap
* AWS: aws-sdk-go-v2
* K8s: client-go

---

# 📦 実装内容

### 機能

1. AWS EC2一覧取得API
2. K8s Pod一覧取得API
3. CLIから同じ操作可能
4. 操作履歴をDB保存
5. JSON構造化ログ出力

---

# 🗂 ディレクトリ構成

```
infra-control/
 ├ cmd/
 │   └ root.go
 ├ internal/
 │   ├ api/
 │   ├ service/
 │   ├ model/
 │   └ logger/
 ├ main.go
 └ config.yaml
```

---

# 🧠 モデル定義（監査ログ）

```go
// internal/model/audit.go
package model

import "time"

type AuditLog struct {
    ID        uint      `gorm:"primaryKey"`
    Action    string
    Target    string
    CreatedAt time.Time
}
```

---

# ⚙ Logger初期化（Zap）

```go
// internal/logger/logger.go
package logger

import "go.uber.org/zap"

var Log *zap.Logger

func Init() {
    logger, _ := zap.NewProduction()
    Log = logger
}
```

---

# ☁ AWS EC2取得サービス

```go
// internal/service/aws.go
package service

import (
    "context"
    "github.com/aws/aws-sdk-go-v2/service/ec2"
)

func ListEC2(client *ec2.Client) ([]string, error) {
    result, err := client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})
    if err != nil {
        return nil, err
    }

    var instances []string
    for _, r := range result.Reservations {
        for _, i := range r.Instances {
            instances = append(instances, *i.InstanceId)
        }
    }
    return instances, nil
}
```

---

# ☸ Kubernetes Pod取得

```go
// internal/service/k8s.go
package service

import (
    "context"
    "k8s.io/client-go/kubernetes"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ListPods(clientset *kubernetes.Clientset) ([]string, error) {
    pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
    if err != nil {
        return nil, err
    }

    var names []string
    for _, p := range pods.Items {
        names = append(names, p.Name)
    }

    return names, nil
}
```

---

# 🌐 API実装（Gin）

```go
// internal/api/server.go
package api

import (
    "github.com/gin-gonic/gin"
)

func Start() {
    r := gin.Default()

    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    r.Run(":8080")
}
```

---

# 🖥 CLI（Cobra）

```go
// cmd/root.go
package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "infra-control",
    Short: "Infrastructure control CLI",
}

var ec2Cmd = &cobra.Command{
    Use:   "ec2",
    Short: "List EC2 instances",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("EC2 list command executed")
    },
}

func Execute() {
    rootCmd.AddCommand(ec2Cmd)
    rootCmd.Execute()
}
```

---

# 🚀 main.go

```go
package main

import (
    "infra-control/cmd"
    "infra-control/internal/logger"
)

func main() {
    logger.Init()
    cmd.Execute()
}
```

---

# 🔥 発展させるなら

* Terraform CLI実行ラッパー実装
* SLO監視API
* Kubernetesイベント監査
* Slack通知
* RBAC機能
* OIDC認証

---

# 💡 これ実務でどう使う？

例えば：

* EC2の増減をAPI経由で可視化
* 本番K8s Pod状態を即時確認
* IaC実行をAPIでトリガー
* 監査ログ保存（SOC2対応）

---

# 🧠 Shumaレベルでやるなら

さらに：

* Clean Architecture適用
* DI（Wire）
* OpenTelemetry統合
* Prometheusメトリクス出力
* GitHub Actions CI

---

もしよければ次は：

* 🏗 Clean Architecture版に進化させる？
* 🐳 Docker + ECSデプロイ？
* ☸ Kubernetes Operator化？
* 🧪 テスト付きフル実装？

どこまで本気で作りますか？ 😎
