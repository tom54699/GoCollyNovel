package gui

import (
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/GoCollyNovel/go-novel-scraper/internal/scraper"
)

func RunFyneGUI() {
	a := app.New()
	w := a.NewWindow("Web Scraper")

	img := canvas.NewImageFromFile("Icon.png")
	img.FillMode = canvas.ImageFillContain

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
		outputPath := filepath.Join(dir, scraper.OutputFileName)
		result.SetText("Scraping in progress...")
		go scraper.ScrapeWebsite(url, outputPath)
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
