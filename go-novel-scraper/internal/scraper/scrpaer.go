package scraper

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

var (
	wg             sync.WaitGroup
	mu             sync.Mutex
	contentMap     = make(map[int]string)
	goroutineLimit = 50
	retryLimit     = 3
	logFormat      = "%-8s%s"
	OutputFileName = "novel.txt"
)

func ScrapeWebsite(baseURL string, outputFileName string) {
	startTime := time.Now()

	mainCollector := createCollector()

	hrefs := make([]string, 0, 600)

	mainCollector.OnHTML("#catalog ul li", func(e *colly.HTMLElement) {
		href := e.ChildAttr("a", "href")
		if href != "" {
			hrefs = append(hrefs, href)
		}
	})

	if err := mainCollector.Visit(baseURL); err != nil {
		log.Printf("Failed to visit base URL: %v", err)
		return
	}

	sem := make(chan struct{}, goroutineLimit)
	for idx, href := range hrefs {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, href string) {
			defer wg.Done()
			defer func() { <-sem }()
			startGoroutine := time.Now()
			fullURL := href

			pageCollector := createCollector()
			pageCollector.OnHTML(".txtnav", func(e *colly.HTMLElement) {
				content := e.Text
				mu.Lock()
				contentMap[idx] = content
				mu.Unlock()
			})

			for attempt := 1; attempt <= retryLimit; attempt++ {
				err := pageCollector.Visit(fullURL)
				if err == nil {
					break
				}
				if attempt < retryLimit {
					log.Printf("Retrying URL %s (attempt %d/%d)", fullURL, attempt, retryLimit)
					time.Sleep(time.Second * time.Duration(attempt))
				} else {
					log.Printf("Failed to visit URL %s after %d attempts", fullURL, retryLimit)
				}
			}

			endGoroutine := time.Now()
			log.Printf("Goroutine ended for URL: %s at %v", href, endGoroutine)
			log.Printf("Duration for URL %s: %v", href, endGoroutine.Sub(startGoroutine))
		}(idx, href)
	}

	wg.Wait()

	file, err := os.Create(outputFileName)
	if err != nil {
		log.Printf("Failed to create output file: %v", err)
		return
	}
	defer file.Close()

	for idx := 0; idx < len(hrefs); idx++ {
		if content, ok := contentMap[idx]; ok {
			_, err := file.WriteString(content + "\n")
			if err != nil {
				log.Printf("Failed to write to file: %v", err)
				return
			}
		}
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)
	log.Printf("Finished scraping. Content saved to %s\nTotal execution time: %s", outputFileName, duration)
}

func createCollector() *colly.Collector {
	c := colly.NewCollector()

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: 2 * time.Second,
	})

	extensions.RandomUserAgent(c)

	c.DetectCharset = false
	c.AllowURLRevisit = true

	c.OnRequest(setReferer)

	c.OnResponse(func(r *colly.Response) {
		reader := transform.NewReader(bytes.NewReader(r.Body), simplifiedchinese.GBK.NewDecoder())
		body, err := io.ReadAll(reader)
		if err != nil {
			log.Println("Failed to read response body:", err)
			return
		}
		r.Body = body
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf(logFormat, "[ERROR]", fmt.Sprintf("Error: %s: Request URL: %s", err, r.Request.URL))
	})

	return c
}

func setReferer(r *colly.Request) {
	r.Headers.Set("Connection", "keep-alive")
	r.Headers.Set("Content-Type", "text/html")
	r.Headers.Set("Accept-Language", "en-US,en;q=0.9")
}
