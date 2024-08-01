package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
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
	outputFileName = "novel.txt"
)

func main() {
	a := app.New()
	w := a.NewWindow("Web Scraper")

	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("Enter URL here...")

	selectedDirLabel := widget.NewLabel("No directory selected")

	result := widget.NewLabel("")

	selectDirButton := widget.NewButton("Select Output Directory", func() {
		dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
			if err != nil {
				result.SetText(fmt.Sprintf("Failed to select directory: %v", err))
				return
			}
			if dir != nil {
				selectedDirLabel.SetText(dir.Path())
			}
		}, w)
	})

	startButton := widget.NewButton("Start Scraping", func() {
		url := urlEntry.Text
		dir := selectedDirLabel.Text
		if url == "" {
			result.SetText("Please enter a valid URL.")
			return
		}
		if dir == "No directory selected" {
			result.SetText("Please select a valid directory.")
			return
		}
		outputPath := filepath.Join(dir, outputFileName)
		result.SetText("Scraping in progress...")
		scrapeWebsite(url, outputPath, result)
	})

	content := container.NewVBox(
		widget.NewLabel("URL:"),
		urlEntry,
		selectDirButton,
		selectedDirLabel,
		startButton,
		result,
	)

	w.SetContent(content)
	w.Resize(fyne.NewSize(400, 200))
	w.ShowAndRun()
}

func scrapeWebsite(baseURL string, outputFileName string, result *widget.Label) {
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
		result.SetText(fmt.Sprintf("Failed to visit base URL: %v", err))
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
		result.SetText(fmt.Sprintf("Failed to create output file: %v", err))
		return
	}
	defer file.Close()

	for idx := 0; idx < len(hrefs); idx++ {
		if content, ok := contentMap[idx]; ok {
			_, err := file.WriteString(content + "\n")
			if err != nil {
				result.SetText(fmt.Sprintf("Failed to write to file: %v", err))
				return
			}
		}
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)
	result.SetText(fmt.Sprintf("Finished scraping. Content saved to %s\nTotal execution time: %s", outputFileName, duration))
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
