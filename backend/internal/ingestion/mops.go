package ingestion

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// MOPS 靜態月報檔案：一個 GET 拿到全市場當月所有公司營收，不需要 session cookie
// URL 格式：https://mops.twse.com.tw/nas/t21/{sii|otc}/t21sc04_{民國年}_{月}_0.html
const mopsStaticBase = "https://mops.twse.com.tw/nas/t21"

// reTicker：台股代碼為 4 位數字（部分特殊股為 5-6 碼，但主流上市/上櫃為 4 碼）
var reTicker = regexp.MustCompile(`^\d{4,6}$`)

// FetchMOPSRevenueAll 抓取指定年月的全市場月營收（上市 + 上櫃），回傳以 ticker 為 key 的 map
// year/month 為西元年月（e.g., 2026, 5）
func FetchMOPSRevenueAll(year, month int) ([]RevenueRow, error) {
	minguo := year - 1911
	period := fmt.Sprintf("%04d-%02d", year, month)
	date := fmt.Sprintf("%04d-%02d-01", year, month)

	var rows []RevenueRow

	for _, boardType := range []string{"sii", "otc"} {
		url := fmt.Sprintf("%s/%s/t21sc04_%d_%d_0.html", mopsStaticBase, boardType, minguo, month)
		batch, err := fetchMOPSFile(url, period, date)
		if err != nil {
			// 非致命：月初尚未公佈或該板塊無資料
			continue
		}
		rows = append(rows, batch...)
	}
	return rows, nil
}

// FetchMOPSRevenueRecent 抓取最近 N 個月的月營收（通常 2 個月夠用）
func FetchMOPSRevenueRecent(n int) ([]RevenueRow, error) {
	loc, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		loc = time.FixedZone("CST", 8*60*60)
	}
	now := time.Now().In(loc)

	var all []RevenueRow
	for i := 0; i < n; i++ {
		t := now.AddDate(0, -i, 0)
		rows, err := FetchMOPSRevenueAll(t.Year(), int(t.Month()))
		if err != nil {
			continue
		}
		all = append(all, rows...)
	}
	return all, nil
}

func fetchMOPSFile(url, period, date string) ([]RevenueRow, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; AlphaLens/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("mops GET %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("mops file not found: %s", url)
	}

	return parseMOPSStaticHTML(resp.Body, period, date)
}

// parseMOPSStaticHTML 解析靜態月報 HTML
// 表格欄位（依序）：公司代號 | 公司名稱 | 當月營收 | 上月營收 | 去年同月 | ...
// 只取欄位 0（代號）和 欄位 2（當月營收）
func parseMOPSStaticHTML(r io.Reader, period, date string) ([]RevenueRow, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	var allRows [][]string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tr" {
			var cells []string
			var extractCells func(*html.Node)
			extractCells = func(c *html.Node) {
				if c.Type == html.ElementNode && (c.Data == "td" || c.Data == "th") {
					cells = append(cells, textContent(c))
				}
				for child := c.FirstChild; child != nil; child = child.NextSibling {
					extractCells(child)
				}
			}
			extractCells(n)
			if len(cells) > 0 {
				allRows = append(allRows, cells)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	var rows []RevenueRow
	for _, cells := range allRows {
		if len(cells) < 3 {
			continue
		}
		ticker := strings.TrimSpace(cells[0])
		if !reTicker.MatchString(ticker) {
			continue // 跳過 header 或非代碼行
		}

		revenueStr := strings.ReplaceAll(strings.TrimSpace(cells[2]), ",", "")
		revenue, err := strconv.ParseInt(revenueStr, 10, 64)
		if err != nil || revenue <= 0 {
			continue
		}

		rows = append(rows, RevenueRow{
			Date:   date,
			Ticker: ticker,
			Period: period,
			Amount: revenue,
		})
	}
	return rows, nil
}

// textContent 遞迴取出節點的純文字
func textContent(n *html.Node) string {
	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			sb.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.TrimSpace(sb.String())
}
