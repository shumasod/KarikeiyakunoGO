# ==============================================================================
# Stage 1: Builder
#   - Go 1.22 Alpine ベースでバイナリをコンパイル
#   - CGO 無効・静的リンクにより最小バイナリを生成
# ==============================================================================
FROM golang:1.22-alpine AS builder

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

WORKDIR /build

# 依存関係レイヤーをキャッシュ（ソースより先にコピー）
COPY go.mod ./
RUN go mod download

# ソース全体をコピーしてビルド
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
      -ldflags "-X main.version=${VERSION} \
                -X main.commit=${COMMIT} \
                -X main.buildDate=${BUILD_DATE} \
                -w -s" \
      -o garapon .

# ==============================================================================
# Stage 2: Runtime
#   - scratch に近い distroless/static を使用（CA証明書・タイムゾーンのみ）
#   - 最終イメージサイズを最小化
# ==============================================================================
FROM alpine:3.19 AS runtime

# タイムゾーンデータ（time.Now の JST 表示に必要）と CA 証明書
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -S garapon && adduser -S -G garapon garapon

ENV TZ=Asia/Tokyo

WORKDIR /app

COPY --from=builder /build/garapon .

# 非 root ユーザーで実行
USER garapon

EXPOSE 8081

HEALTHCHECK --interval=15s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8081/api/prizes || exit 1

ENTRYPOINT ["./garapon"]
