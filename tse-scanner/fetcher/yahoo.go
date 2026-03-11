// Package fetcher retrieves real-time stock quotes from Yahoo Finance.
// Yahoo Finance provides public JSON endpoints for Japanese stocks (*.T suffix).
package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"tse-scanner/model"
)

const (
	baseURL   = "https://query1.finance.yahoo.com/v7/finance/quote"
	batchSize = 50 // Yahoo Finance accepts up to ~100 symbols per request
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

// HTTPDoer is the interface satisfied by *http.Client, enabling test injection.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client wraps an HTTP client and fetches Yahoo Finance quotes.
type Client struct {
	http HTTPDoer
}

// New returns a Client with a production HTTP client (10s timeout).
func New() *Client {
	return &Client{
		http: &http.Client{Timeout: 10 * time.Second},
	}
}

// NewWithHTTP returns a Client using the provided HTTPDoer (for testing).
func NewWithHTTP(h HTTPDoer) *Client {
	return &Client{http: h}
}

// FetchQuotes fetches quotes for all stocks in the watchlist.
// Requests are batched to avoid hitting rate limits.
// Stocks that fail to fetch are returned as Quote{Valid: false}.
func (c *Client) FetchQuotes(ctx context.Context, stocks []model.Stock) ([]model.Quote, error) {
	// Build symbol→Stock lookup for fast merging
	lookup := make(map[string]model.Stock, len(stocks))
	for _, s := range stocks {
		lookup[s.Symbol] = s
	}

	results := make([]model.Quote, 0, len(stocks))
	for i := 0; i < len(stocks); i += batchSize {
		end := i + batchSize
		if end > len(stocks) {
			end = len(stocks)
		}
		batch := stocks[i:end]

		quotes, err := c.fetchBatch(ctx, batch, lookup)
		if err != nil {
			// Partial failure: fill batch as invalid and continue
			for _, s := range batch {
				results = append(results, model.Quote{
					Symbol: s.Symbol, Name: s.Name, Sector: s.Sector,
					Valid: false,
				})
			}
			continue
		}
		results = append(results, quotes...)
	}
	return results, nil
}

// fetchBatch fetches one batch of up to batchSize symbols.
func (c *Client) fetchBatch(ctx context.Context, batch []model.Stock, lookup map[string]model.Stock) ([]model.Quote, error) {
	symbols := make([]string, len(batch))
	for i, s := range batch {
		symbols[i] = s.Symbol
	}

	url := fmt.Sprintf("%s?symbols=%s&fields=shortName,regularMarketPrice,regularMarketChange,regularMarketChangePercent,regularMarketVolume,averageDailyVolume3Month,regularMarketDayHigh,regularMarketDayLow,regularMarketOpen,regularMarketPreviousClose,fiftyTwoWeekHigh,fiftyTwoWeekLow",
		baseURL, strings.Join(symbols, ","))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("リクエスト作成エラー: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTPリクエストエラー: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTPステータス %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("レスポンス読み込みエラー: %w", err)
	}

	return parseResponse(body, lookup, batch)
}

// ---- Yahoo Finance JSON response types ----

type yahooResponse struct {
	QuoteResponse struct {
		Result []yahooQuote `json:"result"`
		Error  *struct {
			Code        string `json:"code"`
			Description string `json:"description"`
		} `json:"error"`
	} `json:"quoteResponse"`
}

type yahooQuote struct {
	Symbol                    string  `json:"symbol"`
	ShortName                 string  `json:"shortName"`
	RegularMarketPrice        float64 `json:"regularMarketPrice"`
	RegularMarketChange       float64 `json:"regularMarketChange"`
	RegularMarketChangePercent float64 `json:"regularMarketChangePercent"`
	RegularMarketVolume       int64   `json:"regularMarketVolume"`
	AverageDailyVolume3Month  int64   `json:"averageDailyVolume3Month"`
	RegularMarketDayHigh      float64 `json:"regularMarketDayHigh"`
	RegularMarketDayLow       float64 `json:"regularMarketDayLow"`
	RegularMarketOpen         float64 `json:"regularMarketOpen"`
	RegularMarketPreviousClose float64 `json:"regularMarketPreviousClose"`
	FiftyTwoWeekHigh          float64 `json:"fiftyTwoWeekHigh"`
	FiftyTwoWeekLow           float64 `json:"fiftyTwoWeekLow"`
}

// parseResponse parses the Yahoo Finance JSON and merges sector info from lookup.
func parseResponse(body []byte, lookup map[string]model.Stock, batch []model.Stock) ([]model.Quote, error) {
	var yr yahooResponse
	if err := json.Unmarshal(body, &yr); err != nil {
		return nil, fmt.Errorf("JSONパースエラー: %w", err)
	}
	if yr.QuoteResponse.Error != nil {
		return nil, fmt.Errorf("Yahoo Finance エラー: %s", yr.QuoteResponse.Error.Description)
	}

	fetched := make(map[string]model.Quote, len(yr.QuoteResponse.Result))
	now := time.Now()
	for _, yq := range yr.QuoteResponse.Result {
		s := lookup[yq.Symbol]
		name := yq.ShortName
		if s.Name != "" {
			name = s.Name // prefer Japanese name
		}
		fetched[yq.Symbol] = model.Quote{
			Symbol:        yq.Symbol,
			Name:          name,
			Sector:        s.Sector,
			Price:         yq.RegularMarketPrice,
			Change:        yq.RegularMarketChange,
			ChangePercent: yq.RegularMarketChangePercent,
			Volume:        yq.RegularMarketVolume,
			AvgVolume3M:   yq.AverageDailyVolume3Month,
			DayHigh:       yq.RegularMarketDayHigh,
			DayLow:        yq.RegularMarketDayLow,
			Open:          yq.RegularMarketOpen,
			PrevClose:     yq.RegularMarketPreviousClose,
			WeekHigh52:    yq.FiftyTwoWeekHigh,
			WeekLow52:     yq.FiftyTwoWeekLow,
			FetchedAt:     now,
			Valid:          true,
		}
	}

	// Preserve order and fill missing symbols as invalid
	quotes := make([]model.Quote, 0, len(batch))
	for _, s := range batch {
		if q, ok := fetched[s.Symbol]; ok {
			quotes = append(quotes, q)
		} else {
			quotes = append(quotes, model.Quote{
				Symbol: s.Symbol, Name: s.Name, Sector: s.Sector,
				FetchedAt: now, Valid: false,
			})
		}
	}
	return quotes, nil
}
