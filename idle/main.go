package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// ANSI カラーコード
const (
	reset   = "\033[0m"
	bold    = "\033[1m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	white   = "\033[37m"

	bgRed     = "\033[41m"
	bgGreen   = "\033[42m"
	bgYellow  = "\033[43m"
	bgBlue    = "\033[44m"
	bgMagenta = "\033[45m"

	clearScreen = "\033[2J\033[H"
	hideCursor  = "\033[?25l"
	showCursor  = "\033[?25h"
)

// アイドルの名前と色
var idols = []struct {
	name  string
	color string
}{
	{"愛花", red},
	{"美妮", magenta},
	{"ここ", yellow},
	{"桃梨", green},
	{"笑美", cyan},
}

// 4フレームのダンスアニメーション（各アイドル用）
// フレームはアイドルの動きを表す（7行）
var danceFrames = [4][7]string{
	// フレーム1: バンザイポーズ
	{
		` \(^o^)/`,
		`  |  | `,
		`  |  | `,
		`  |  | `,
		` /    \`,
		`/      \`,
		`        `,
	},
	// フレーム2: 右手を上げる
	{
		`  (^o^)/`,
		`  |  /  `,
		`  | /   `,
		`  |/    `,
		`  |     `,
		` / \    `,
		`        `,
	},
	// フレーム3: ジャンプポーズ
	{
		` \(^o^)/`,
		`  |  | `,
		` /|  |\ `,
		`/ |  | \`,
		`  | /|  `,
		`  |/ |  `,
		`        `,
	},
	// フレーム4: 左手を上げる
	{
		` \(^o^) `,
		`  \  |  `,
		`   \ |  `,
		`    \|  `,
		`     |  `,
		`    / \ `,
		`        `,
	},
}

// 音符アニメーション
var musicNotes = []string{"♪", "♫", "♬", "♩", "♭", "♮"}

// ステージの装飾
func printStage(frame int) {
	noteIdx := frame % len(musicNotes)
	note := musicNotes[noteIdx]

	// タイトル
	fmt.Printf("%s%s%s\n", bold+yellow, "  ★☆★ アイドルダンスショー ★☆★", reset)
	fmt.Printf("%s  %s %s %s %s %s %s %s %s%s\n",
		cyan, note, note, note, note, note, note, note, note, reset)
	fmt.Println()
}

// ステージ枠の上部
func printStageTop() {
	fmt.Printf("%s╔══════════════════════════════════════════════════════════╗%s\n", white, reset)
}

// ステージ枠の下部
func printStageBottom() {
	fmt.Printf("%s╚══════════════════════════════════════════════════════════╝%s\n", white, reset)
}

// ライトの点滅エフェクト
func printLights(frame int) {
	lights := []string{red, yellow, green, cyan, magenta, blue, white}
	result := white + "║ "
	for i := 0; i < 28; i++ {
		colorIdx := (i + frame) % len(lights)
		result += lights[colorIdx] + "★" + reset
	}
	result += white + " ║" + reset
	fmt.Println(result)
}

// アイドルを1フレーム描画
func printIdols(frameIdx int) {
	// 各行を組み立てて表示
	for row := 0; row < 7; row++ {
		line := white + "║  " + reset
		for i, idol := range idols {
			// フレームオフセット: アイドルごとに少しずらしてウェーブ効果
			adjustedFrame := (frameIdx + i) % 4
			cell := idol.color + danceFrames[adjustedFrame][row] + reset
			line += cell
			if i < len(idols)-1 {
				line += "  "
			}
		}
		line += white + "  ║" + reset
		fmt.Println(line)
	}
}

// アイドルの名前を表示
func printNames() {
	line := white + "║  " + reset
	for i, idol := range idols {
		nameDisplay := bold + idol.color + fmt.Sprintf("  %s   ", idol.name) + reset
		line += nameDisplay
		if i < len(idols)-1 {
			line += " "
		}
	}
	line += white + " ║" + reset
	fmt.Println(line)
}

// 観客の歓声
var crowdFrames = []string{
	"  ヾ(^▽^)ノ ヾ(≧∇≦)ノ ヾ(*°▽°*)ノ ヾ(^o^)ノ ヾ(≧▽≦)ノ",
	"  ヾ(≧∇≦)ノ ヾ(*°▽°*)ノ ヾ(^o^)ノ  ヾ(^▽^)ノ ヾ(≧∇≦)ノ",
}

func printCrowd(frame int) {
	crowdColor := yellow
	fmt.Printf("%s%s%s\n", crowdColor, crowdFrames[frame%2], reset)
}

// スコアボード
var scores = []int{98, 97, 99, 96, 98}

func printScores() {
	line := ""
	for i, idol := range idols {
		line += fmt.Sprintf("%s%s:%d点%s  ", bold+idol.color, idol.name, scores[i], reset)
	}
	fmt.Println(" " + line)
}

func animate() {
	fmt.Print(hideCursor)
	defer fmt.Print(showCursor)

	// Ctrl+C のハンドリング
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Print(showCursor)
		fmt.Printf("\n%s%sショーを終了します。ありがとうございました！%s\n", bold, yellow, reset)
		os.Exit(0)
	}()

	frame := 0
	for {
		fmt.Print(clearScreen)

		printStage(frame)
		printStageTop()
		printLights(frame)

		// ステージの空白行
		fmt.Printf("%s║%s%58s%s║%s\n", white, reset, "", white, reset)

		printIdols(frame)

		// 名前行
		fmt.Printf("%s║%s%58s%s║%s\n", white, reset, "", white, reset)
		printNames()

		// 音符行
		noteColors := []string{red, magenta, yellow, green, cyan}
		noteLine := white + "║  " + reset
		for j := 0; j < 14; j++ {
			noteIdx := (j + frame) % len(musicNotes)
			colorIdx := (j + frame) % len(noteColors)
			noteLine += noteColors[colorIdx] + musicNotes[noteIdx] + reset + " "
		}
		noteLine += white + "  ║" + reset
		fmt.Println(noteLine)

		printLights(frame + 3)
		printStageBottom()

		fmt.Println()
		printCrowd(frame)
		fmt.Println()
		printScores()
		fmt.Println()
		fmt.Printf("%s  ♪ Ctrl+C で終了 ♪%s\n", white, reset)

		frame++
		time.Sleep(300 * time.Millisecond)
	}
}

func main() {
	animate()
}
