package main

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

// benchmarkBody генерирует HTML с n ссылками для бенчмарков extractUrls.
func benchmarkBody(n int) []byte {
	var sb strings.Builder
	sb.WriteString("<html><body>")
	for i := range n {
		fmt.Fprintf(&sb, `<a href="https://example.com/page%d">link %d</a>`, i, i)
	}
	sb.WriteString("</body></html>")
	return []byte(sb.String())
}

func newBenchCrawler() *Crawler {
	c := NewCrawler()
	c.config = &Config{
		MaxDepth:        25,
		MinSleep:        3,
		MaxSleep:        6,
		RootURLs:        []string{"https://example.com"},
		BlacklistedURLs: []string{".css", ".png", ".ico", "bit.ly"},
		UserAgents:      []string{"test-agent"},
	}
	c.blacklisted = make(map[string]struct{})
	return c
}

// extractUrlsOld воспроизводит дооптимизационную реализацию (регэксп компилируется
// на каждый вызов, тело — string) для прямого сравнения в бенчмарке.
func extractUrlsOld(c *Crawler, body string, rootURL string) []string {
	pattern := `href\s*=\s*["']([^"']*)["']`
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(body, -1)

	var urls []string
	for _, match := range matches {
		if len(match) > 1 {
			link := match[1]
			if strings.HasPrefix(link, "#") {
				continue
			}
			normalizedURL := c.normalizeLink(link, rootURL)
			if c.shouldAcceptURL(normalizedURL) {
				urls = append(urls, normalizedURL)
			}
		}
	}
	return urls
}

func BenchmarkExtractUrlsOld(b *testing.B) {
	c := newBenchCrawler()
	body := string(benchmarkBody(200))
	root := "https://example.com"

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = extractUrlsOld(c, body, root)
	}
}

func BenchmarkExtractUrls(b *testing.B) {
	c := newBenchCrawler()
	body := benchmarkBody(200)
	root := "https://example.com"

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = c.extractUrls(body, root)
	}
}

// isBlacklistedOld воспроизводит дооптимизационную проверку: подстрочный скан
// по всему списку забракованных URL (включая динамически накопленные).
func isBlacklistedOld(patterns []string, urlStr string) bool {
	for _, p := range patterns {
		if strings.Contains(urlStr, p) {
			return true
		}
	}
	return false
}

func BenchmarkIsBlacklistedOld(b *testing.B) {
	patterns := make([]string, 0, 50000)
	for i := range 50000 {
		patterns = append(patterns, fmt.Sprintf("https://dead.example/%d", i))
	}
	target := "https://example.com/live-page"

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = isBlacklistedOld(patterns, target)
	}
}

// BenchmarkIsBlacklisted демонстрирует O(1)-проверку точных URL независимо от
// числа динамически забракованных ссылок (раньше был O(n) подстрочный скан).
func BenchmarkIsBlacklisted(b *testing.B) {
	c := newBenchCrawler()
	for i := range 50000 {
		c.blacklisted[fmt.Sprintf("https://dead.example/%d", i)] = struct{}{}
	}
	target := "https://example.com/live-page"

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = c.isBlacklisted(target)
	}
}

func TestExtractUrls(t *testing.T) {
	c := newBenchCrawler()
	body := []byte(`<a href="https://example.com/a">a</a>` +
		`<a href="/relative">b</a>` +
		`<a href="#anchor">skip</a>` +
		`<a href="https://example.com/style.css">skip-css</a>`)

	got := c.extractUrls(body, "https://example.com")

	want := map[string]bool{
		"https://example.com/a":        true,
		"https://example.com/relative": true,
	}
	if len(got) != len(want) {
		t.Fatalf("got %d urls %v, want %d", len(got), got, len(want))
	}
	for _, u := range got {
		if !want[u] {
			t.Errorf("unexpected url: %s", u)
		}
	}
}

func TestIsBlacklisted(t *testing.T) {
	c := newBenchCrawler()
	c.blacklisted["https://example.com/dead"] = struct{}{}

	cases := map[string]bool{
		"https://example.com/dead":      true,  // точное совпадение
		"https://example.com/style.css": true,  // паттерн .css
		"https://example.com/ok":        false, // ничего
	}
	for url, want := range cases {
		if got := c.isBlacklisted(url); got != want {
			t.Errorf("isBlacklisted(%q) = %v, want %v", url, got, want)
		}
	}
}
