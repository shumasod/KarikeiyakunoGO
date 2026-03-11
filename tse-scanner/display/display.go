// Package display renders scan results to the terminal using ANSI escape codes.
package display

import (
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"tse-scanner/model"
)

// ANSI escape codes
const (
	reset     = "\033[0m"
	bold      = "\033[1m"
	red       = "\033[31m"
	green     = "\033[32m"
	yellow    = "\033[33m"
	blue      = "\033[34m"
	magenta   = "\033[35m"
	cyan      = "\033[36m"
	white     = "\033[37m"
	bgRed     = "\033[41m"
	bgYellow  = "\033[43m"
	clearLine = "\033[2K"
	clearScr  = "\033[2J\033[H"
)

// Render clears the screen and draws the full scan report.
func Render(candidates []model.Candidate, fetchedAt time.Time, interval time.Duration, totalScanned int) {
	fmt.Print(clearScr)
	printHeader(fetchedAt, interval, totalScanned, len(candidates))
	if len(candidates) == 0 {
		fmt.Printf("\n  %s急騰候補が見つかりませんでした。しばらくお待ちください。%s\n", yellow, reset)
	} else {
		printTable(candidates)
	}
	printFooter()
}

func printHeader(fetchedAt time.Time, interval time.Duration, total, found int) {
	jst := time.FixedZone("JST", 9*60*60)
	ts := fetchedAt.In(jst).Format("2006-01-02 15:04:05")
	status := marketStatus(fetchedAt.In(jst))
	nextIn := interval.Round(time.Second)

	fmt.Printf("%s%s", bold, cyan)
	fmt.Println("╔══════════════════════════════════════════════════════════════════════════════╗")
	fmt.Printf("║  🔥 東証急騰スキャナー  %-20s  次回更新: %-8v       ║\n", ts, nextIn)
	fmt.Printf("║  市場: %-10s  スキャン: %3d 銘柄  急騰候補: %3d 銘柄                  ║\n",
		status, total, found)
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════════╝")
	fmt.Print(reset)
}

func printTable(candidates []model.Candidate) {
	header := fmt.Sprintf("  %s%-6s  %-8s  %-18s  %-8s  %10s  %7s  %6s  %s%s",
		bold,
		"SCORE", "コード", "銘柄名", "業種", "現在値(円)", "騰落率", "出来高比", "シグナル",
		reset)
	fmt.Println()
	fmt.Println(header)
	fmt.Printf("  %s\n", strings.Repeat("─", 80))

	for _, c := range candidates {
		printRow(c)
	}
}

func printRow(c model.Candidate) {
	scoreColor := scoreToColor(c.SurgeScore)
	changeColor := green
	sign := "+"
	if c.ChangePercent < 0 {
		changeColor = red
		sign = ""
	}

	volStr := "  N/A "
	if c.AvgVolume3M > 0 {
		volStr = fmt.Sprintf("%5.1fx", c.VolumeRatio)
	}

	// Collect display-only signals (those whose Label starts with emoji)
	var sigLabels []string
	for _, s := range c.Signals {
		if isDisplaySignal(s.Label) {
			sigLabels = append(sigLabels, s.Label)
		}
	}
	sigStr := strings.Join(sigLabels, " ")

	// Pad Japanese name to fixed display width
	paddedName := padJP(c.Name, 18)

	codeStr := strings.TrimSuffix(c.Symbol, ".T")

	fmt.Printf("  %s%6.1f%s  %-8s  %s  %-8s  %10s  %s%s%7.2f%%%s  %6s  %s\n",
		scoreColor, c.SurgeScore, reset,
		codeStr,
		paddedName,
		c.Sector,
		formatPrice(c.Price),
		changeColor, sign, c.ChangePercent, reset,
		volStr,
		sigStr,
	)
}

func printFooter() {
	fmt.Printf("\n  %s%s⚠ 投資は自己責任です。このツールは情報提供のみを目的としています。%s\n",
		bold, yellow, reset)
}

// ---- helpers ----

func marketStatus(now time.Time) string {
	h, m, _ := now.Clock()
	mins := h*60 + m
	wd := now.Weekday()
	if wd == time.Saturday || wd == time.Sunday {
		return "🔴 休場（週末）"
	}
	// 前場 9:00–11:30、後場 12:30–15:30
	if (mins >= 9*60 && mins < 11*60+30) || (mins >= 12*60+30 && mins < 15*60+30) {
		return "🟢 取引中"
	}
	if mins >= 8*60 && mins < 9*60 {
		return "🟡 前場前"
	}
	if mins >= 15*60+30 {
		return "🔴 取引終了"
	}
	return "🟡 昼休み"
}

func scoreToColor(score float64) string {
	switch {
	case score >= 80:
		return bold + red
	case score >= 60:
		return bold + yellow
	case score >= 40:
		return bold + cyan
	default:
		return white
	}
}

func formatPrice(p float64) string {
	if p >= 10000 {
		return fmt.Sprintf("%10.0f", math.Round(p))
	}
	return fmt.Sprintf("%10.1f", p)
}

// isDisplaySignal returns true for emoji-prefixed signal labels (visual indicators).
func isDisplaySignal(label string) bool {
	if label == "" {
		return false
	}
	// Emoji code points are >= 0x1F000
	r, _ := utf8.DecodeRuneInString(label)
	return r > 0x1F000
}

// padJP pads a string containing Japanese characters to targetDisplayWidth.
// CJK characters count as 2 display columns.
func padJP(s string, targetDisplayWidth int) string {
	width := displayWidth(s)
	if width >= targetDisplayWidth {
		return s
	}
	return s + strings.Repeat(" ", targetDisplayWidth-width)
}

func displayWidth(s string) int {
	w := 0
	for _, r := range s {
		if r > 0x1000 { // CJK and emoji: 2 columns
			w += 2
		} else {
			w++
		}
	}
	return w
}
