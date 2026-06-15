package main

import (
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"
)

// Версионная информация, проставляется через ldflags при сборке (goreleaser).
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// Глобальный уровень логирования
var logLevel = "INFO"

// maxBodyBytes ограничивает размер читаемого тела ответа от недоверенного
// удалённого сервера (defense-in-depth против раздувания памяти).
const maxBodyBytes = 10 << 20 // 10 MiB

// maxDeadLinks ограничивает размер множества динамически забракованных URL,
// чтобы память долгоживущего процесса не росла неограниченно. При превышении
// множество сбрасывается — знание о мёртвых ссылках лишь оптимизация.
const maxDeadLinks = 100_000

// hrefRe извлекает значения href из HTML. Компилируется один раз: extractUrls
// вызывается на каждую страницу, а компиляция регэкспа дороже самого матчинга.
var hrefRe = regexp.MustCompile(`href\s*=\s*["']([^"']*)["']`)

// LogLevel определяет приоритет уровня логирования
var logLevelPriority = map[string]int{
	"DEBUG":   0,
	"INFO":    1,
	"WARNING": 2,
	"ERROR":   3,
}

// logMessage выводит сообщение только если его уровень достаточен
func logMessage(level, format string, v ...interface{}) {
	if logLevelPriority[level] >= logLevelPriority[logLevel] {
		log.Printf("["+level+"] "+format, v...)
	}
}

// TimeoutValue представляет значение таймаута, которое может быть int или bool
type TimeoutValue struct {
	Value *int
}

// UnmarshalJSON реализует json.Unmarshaler для обработки int или bool
func (t *TimeoutValue) UnmarshalJSON(data []byte) error {
	// Пробуем распарсить как int
	var i int
	if err := json.Unmarshal(data, &i); err == nil {
		t.Value = &i
		return nil
	}

	// Пробуем распарсить как bool
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		if b {
			// Если true, используем значение по умолчанию (0 = без таймаута)
			defaultTimeout := 0
			t.Value = &defaultTimeout
		} else {
			// Если false, nil означает без таймаута
			t.Value = nil
		}
		return nil
	}

	return fmt.Errorf("timeout must be an integer or boolean")
}

// Config представляет конфигурацию краулера
type Config struct {
	MaxDepth        int          `json:"max_depth"`
	MinSleep        int          `json:"min_sleep"`
	MaxSleep        int          `json:"max_sleep"`
	Timeout         TimeoutValue `json:"timeout"`
	RootURLs        []string     `json:"root_urls"`
	BlacklistedURLs []string     `json:"blacklisted_urls"`
	UserAgents      []string     `json:"user_agents"`
}

// Crawler - основной класс для обхода URL
type Crawler struct {
	config      *Config
	client      *http.Client
	links       []string
	blacklisted map[string]struct{}
	startTime   time.Time
}

// CrawlerTimedOut - ошибка превышения таймаута
type CrawlerTimedOut struct{}

func (e CrawlerTimedOut) Error() string {
	return "crawler timeout exceeded"
}

// NewCrawler создает новый экземпляр Crawler
func NewCrawler() *Crawler {
	return &Crawler{
		config:      &Config{},
		links:       []string{},
		blacklisted: make(map[string]struct{}),
		// Клиент создаётся один раз и переиспользуется между запросами
		// (единый конфиг, без аллокаций на каждый запрос).
		client: &http.Client{
			Timeout: 5 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

// readResponseBody читает тело HTTP ответа с поддержкой gzip
func (c *Crawler) readResponseBody(resp *http.Response) ([]byte, error) {
	var reader io.ReadCloser
	var err error

	// Проверяем Content-Encoding
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		defer reader.Close()
	default:
		reader = resp.Body
	}

	body, err := io.ReadAll(io.LimitReader(reader, maxBodyBytes))
	if err != nil {
		return nil, err
	}

	return body, nil
}

// LoadConfigFile загружает конфигурацию из JSON файла
func (c *Crawler) LoadConfigFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening config file: %w", err)
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return fmt.Errorf("error decoding config: %w", err)
	}

	c.config = &config
	return nil
}

// validateConfig проверяет, что обязательные поля конфига заданы корректно.
// Конфиг доверенный (контролируется оператором), но «битые» значения иначе
// приводят к panic во время обхода — даём понятную ошибку вместо краша.
func (c *Crawler) validateConfig() error {
	if len(c.config.RootURLs) == 0 {
		return fmt.Errorf("root_urls must not be empty")
	}
	if len(c.config.UserAgents) == 0 {
		return fmt.Errorf("user_agents must not be empty")
	}
	if c.config.MaxDepth <= 0 {
		return fmt.Errorf("max_depth must be greater than 0")
	}
	if c.config.MinSleep < 0 || c.config.MaxSleep < 0 {
		return fmt.Errorf("min_sleep and max_sleep must not be negative")
	}
	if c.config.MaxSleep <= c.config.MinSleep {
		return fmt.Errorf("max_sleep (%d) must be greater than min_sleep (%d)", c.config.MaxSleep, c.config.MinSleep)
	}
	return nil
}

// SetOption устанавливает конкретную опцию в конфигурации
func (c *Crawler) SetOption(option string, value interface{}) {
	switch option {
	case "timeout":
		if v, ok := value.(int); ok {
			c.config.Timeout.Value = &v
		}
	case "max_depth":
		if v, ok := value.(int); ok {
			c.config.MaxDepth = v
		}
	}
}

// request отправляет HTTP запрос со случайным User-Agent
func (c *Crawler) request(urlStr string) (*http.Response, error) {
	if len(c.config.UserAgents) == 0 {
		return nil, fmt.Errorf("no user agents configured")
	}

	randomUserAgent := c.config.UserAgents[rand.Intn(len(c.config.UserAgents))]

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", randomUserAgent)

	return c.client.Do(req)
}

// normalizeLink нормализует ссылку, делая её абсолютной
func (c *Crawler) normalizeLink(link, rootURL string) string {
	parsedURL, err := url.Parse(link)
	if err != nil {
		return ""
	}

	parsedRootURL, err := url.Parse(rootURL)
	if err != nil {
		return ""
	}

	// '//' означает сохранить текущий протокол
	if strings.HasPrefix(link, "//") {
		return fmt.Sprintf("%s://%s%s", parsedRootURL.Scheme, parsedURL.Host, parsedURL.Path)
	}

	// относительный путь
	if parsedURL.Scheme == "" {
		absoluteURL, err := url.Parse(rootURL)
		if err != nil {
			return ""
		}
		absoluteURL, err = absoluteURL.Parse(link)
		if err != nil {
			return ""
		}
		return absoluteURL.String()
	}

	return link
}

// isValidURL проверяет, является ли URL валидным
func (c *Crawler) isValidURL(urlStr string) bool {
	if urlStr == "" {
		return false
	}

	parsed, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	// Проверяем, что есть схема и хост
	if parsed.Scheme == "" || parsed.Host == "" {
		return false
	}

	// Проверяем, что схема http или https
	if parsed.Scheme != "http" && parsed.Scheme != "https" && parsed.Scheme != "ftp" {
		return false
	}

	return true
}

// isBlacklisted проверяет, находится ли URL в черном списке
func (c *Crawler) isBlacklisted(urlStr string) bool {
	// Точное совпадение с динамически забракованными URL — O(1).
	if _, ok := c.blacklisted[urlStr]; ok {
		return true
	}
	// Подстрочные паттерны из конфига (.css, bit.ly и т.п.).
	for _, blacklistedURL := range c.config.BlacklistedURLs {
		if strings.Contains(urlStr, blacklistedURL) {
			return true
		}
	}
	return false
}

// shouldAcceptURL проверяет, должен ли URL быть принят
func (c *Crawler) shouldAcceptURL(urlStr string) bool {
	return urlStr != "" && c.isValidURL(urlStr) && !c.isBlacklisted(urlStr)
}

// extractUrls извлекает ссылки из HTML тела
func (c *Crawler) extractUrls(body []byte, rootURL string) []string {
	matches := hrefRe.FindAllSubmatch(body, -1)

	urls := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			link := string(match[1])
			// Игнорируем ссылки начинающиеся с #
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

// removeAndBlacklist удаляет ссылку и добавляет в черный список
func (c *Crawler) removeAndBlacklist(link string) {
	// Сбрасываем множество при переполнении: знание о мёртвых ссылках —
	// лишь оптимизация против повторных заходов, его потеря безопасна и
	// ограничивает рост памяти долгоживущего процесса.
	if len(c.blacklisted) >= maxDeadLinks {
		c.blacklisted = make(map[string]struct{})
	}
	c.blacklisted[link] = struct{}{}

	// Удаляем ссылку из текущего списка (он невелик — ссылки одной страницы).
	newLinks := make([]string, 0, len(c.links))
	for _, l := range c.links {
		if l != link {
			newLinks = append(newLinks, l)
		}
	}
	c.links = newLinks
}

// isTimeoutReached проверяет, достигнут ли таймаут
func (c *Crawler) isTimeoutReached() bool {
	if c.config.Timeout.Value == nil || *c.config.Timeout.Value == 0 {
		return false
	}

	endTime := c.startTime.Add(time.Duration(*c.config.Timeout.Value) * time.Second)
	return time.Now().After(endTime)
}

// browseFromLinks выбирает случайную ссылку и переходит по ней
func (c *Crawler) browseFromLinks(depth int) error {
	if depth >= c.config.MaxDepth || len(c.links) == 0 {
		logMessage("DEBUG", "Hit a dead end, moving to the next root URL")
		return nil
	}

	if c.isTimeoutReached() {
		return CrawlerTimedOut{}
	}

	randomLink := c.links[rand.Intn(len(c.links))]

	logMessage("INFO", "Visiting %s", randomLink)

	resp, err := c.request(randomLink)
	if err != nil {
		logMessage("DEBUG", "Exception on URL: %s, removing from list and trying again! Error: %v", randomLink, err)
		c.removeAndBlacklist(randomLink)
		c.browseFromLinks(depth + 1)
		return nil
	}
	defer resp.Body.Close()

	// Читаем тело ответа
	body, err := c.readResponseBody(resp)
	if err != nil {
		logMessage("DEBUG", "Error reading response body: %v", err)
		c.removeAndBlacklist(randomLink)
		c.browseFromLinks(depth + 1)
		return nil
	}

	subLinks := c.extractUrls(body, randomLink)

	// Случайная задержка
	sleepDuration := rand.Intn(c.config.MaxSleep-c.config.MinSleep) + c.config.MinSleep
	time.Sleep(time.Duration(sleepDuration) * time.Second)

	// Убедимся, что у нас есть больше 1 ссылки для выбора
	if len(subLinks) > 1 {
		c.links = subLinks
	} else {
		c.removeAndBlacklist(randomLink)
	}

	return c.browseFromLinks(depth + 1)
}

// Crawl запускает процесс обхода
func (c *Crawler) Crawl() error {
	c.startTime = time.Now()

	for {
		if c.isTimeoutReached() {
			logMessage("INFO", "Timeout has exceeded, exiting")
			return nil
		}

		url := c.config.RootURLs[rand.Intn(len(c.config.RootURLs))]

		resp, err := c.request(url)
		if err != nil {
			logMessage("WARNING", "Error connecting to root url: %s - %v", url, err)
			time.Sleep(time.Duration(c.config.MinSleep) * time.Second)
			continue
		}

		// Читаем тела ответа
		body, err := c.readResponseBody(resp)
		if err != nil {
			logMessage("WARNING", "Error reading response body: %v", err)
			resp.Body.Close()
			time.Sleep(time.Duration(c.config.MinSleep) * time.Second)
			continue
		}
		resp.Body.Close()

		c.links = c.extractUrls(body, url)
		logMessage("DEBUG", "found %d links", len(c.links))

		err = c.browseFromLinks(0)
		if err != nil {
			if _, ok := err.(CrawlerTimedOut); ok {
				logMessage("INFO", "Timeout has exceeded, exiting")
				return nil
			}
			return err
		}
	}
}

func parseLogLevel(level string) error {
	level = strings.ToUpper(level)
	switch level {
	case "DEBUG", "INFO", "WARNING", "ERROR":
		logLevel = level
		return nil
	default:
		return fmt.Errorf("invalid log level: %s", level)
	}
}

func main() {
	logLevelFlag := flag.String("log", "info", "logging level (debug, info, warning, error)")
	configFile := flag.String("config", "", "config file path")
	timeout := flag.Int("timeout", 0, "timeout in seconds")
	verbose := flag.Bool("verbose", false, "enable verbose (debug) logging")
	showVersion := flag.Bool("version", false, "print version and exit")

	flag.Parse()

	// Поддержка `noisy version` как сабкоманды (для homebrew test и привычного UX).
	if *showVersion || (flag.NArg() == 1 && flag.Arg(0) == "version") {
		fmt.Printf("noisy %s (commit %s, built %s)\n", version, commit, date)
		return
	}

	if *configFile == "" {
		log.Fatal("Error: --config flag is required")
	}

	// Если указан --verbose, устанавливаем DEBUG уровень
	if *verbose {
		logLevel = "DEBUG"
	} else if err := parseLogLevel(*logLevelFlag); err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Настройка логирования с указанием уровня
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Обработка прерывания (Ctrl+C)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		logMessage("INFO", "Received interrupt signal, shutting down...")
		os.Exit(0)
	}()

	crawler := NewCrawler()

	if err := crawler.LoadConfigFile(*configFile); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	if err := crawler.validateConfig(); err != nil {
		log.Fatalf("Error: invalid config: %v", err)
	}

	if *timeout > 0 {
		crawler.SetOption("timeout", *timeout)
	}

	if err := crawler.Crawl(); err != nil {
		log.Fatalf("Error during crawl: %v", err)
	}
}
