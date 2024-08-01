package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

var baseURL = "https://69shuba.cx/book/36352/"
var logFormat = "%-8s%s"
var outputFileName = "novel"
var hrefs = make([]string, 0, 600)
var wg sync.WaitGroup
var mu sync.Mutex
var contentMap = make(map[int]string)

func main() {
	startTime := time.Now()

	// 创建和配置主爬虫实例
	mainCollector := createCollector()

	// 使用主爬虫实例爬取初始页面
	mainCollector.OnHTML("#catalog ul li", func(e *colly.HTMLElement) {
		fmt.Printf("%s\n", e.Text)
		href := e.ChildAttr("a", "href")
		fmt.Printf("Link: %s\n", href)
		hrefs = append(hrefs, href)
	})

	mainCollector.Visit(baseURL)

	// 对每个 href 使用 goroutine 并创建新的爬虫实例
	for idx, href := range hrefs {
		wg.Add(1)
		go func(idx int, href string) {
			defer wg.Done()
			startGoroutine := time.Now()
			fullURL := href

			// 创建和配置新的爬虫实例
			pageCollector := createCollector()
			pageCollector.OnHTML(".txtnav", func(e *colly.HTMLElement) {
				// 抓取内容
				content := e.Text
				mu.Lock()
				contentMap[idx] = content
				mu.Unlock()
			})

			pageCollector.Visit(fullURL)

			endGoroutine := time.Now()
			fmt.Printf("Goroutine ended for URL: %s at %v\n", href, endGoroutine)
			fmt.Printf("Duration for URL %s: %v\n", href, endGoroutine.Sub(startGoroutine))
		}(idx, href)
	}

	wg.Wait()

	// 创建输出文件
	file, err := os.Create(outputFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// 按顺序写入文件
	for idx := 0; idx < len(hrefs); idx++ {
		if content, ok := contentMap[idx]; ok {
			_, err := file.WriteString(content + "\n")
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)

	fmt.Printf("Finished scraping. Content saved to %s\n", outputFileName)
	fmt.Printf("Total execution time: %s\n", duration)
}

// 创建和配置 colly.Collector 的函数
func createCollector() *colly.Collector {
	c := colly.NewCollector()

	// c.Limit(&colly.LimitRule{
	// 	DomainGlob:  "*",
	// 	RandomDelay: 2 * time.Second,
	// })
	extensions.RandomUserAgent(c)

	c.DetectCharset = false
	c.AllowURLRevisit = true

	c.OnRequest(setReferer)

	c.OnResponse(func(r *colly.Response) {
		// 使用 GBK 解码器创建一个新的 Reader
		reader := transform.NewReader(bytes.NewReader(r.Body), simplifiedchinese.GBK.NewDecoder())

		// 读取转换后的内容
		body, err := io.ReadAll(reader)
		if err != nil {
			fmt.Println("Failed to read response body:", err)
			return
		}

		// 更新响应的 Body
		r.Body = body
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf(logFormat, "[ERROR]", fmt.Sprintf("Error: %s: Request URL: %s", err, r.Request.URL))
	})

	return c
}

// 设置 Referer
func setReferer(r *colly.Request) {
	r.Headers.Set("Connection", "keep-alive")
	r.Headers.Set("Content-Type", "text/html")
	r.Headers.Set("Accept-Language", "en-US,en;q=0.9")
}
