package main

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/GoCollyNovel/go-novel-scraper/configs"
	"github.com/GoCollyNovel/go-novel-scraper/internal/scraper"
	"github.com/GoCollyNovel/go-novel-scraper/pkg/db"
	"github.com/GoCollyNovel/go-novel-scraper/pkg/logger"
	"github.com/GoCollyNovel/go-novel-scraper/pkg/redis"
)

type PageData struct {
	Message string
}

func init() {
	configs.Setup()
	logger.Setup()
	db.ConnectMongoDB()
	redis.ConnectRedis()
}

func main() {
	// gui.RunFyneGUI()
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/scrape", handleScrape)

	log.Println("Starting server on :8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func handleScrape(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	url := r.FormValue("url")
	log.Println(url)
	outputDir := r.FormValue("outputDir")

	if url == "" || outputDir == "" {
		http.Error(w, "URL and output directory are required", http.StatusBadRequest)
		return
	}

	outputPath := filepath.Join(outputDir, scraper.OutputFileName)
	go scraper.ScrapeWebsite(url, outputPath)

	data := PageData{Message: "Scraping started. Please check the console for progress."}
	tmpl, _ := template.ParseFiles("index.html")
	tmpl.Execute(w, data)
}
