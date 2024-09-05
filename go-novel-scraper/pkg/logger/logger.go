package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/GoCollyNovel/go-novel-scraper/configs"
	"github.com/GoCollyNovel/go-novel-scraper/pkg/file"
)

type Level int

var (
	F *os.File

	DefaultPrefix      = ""
	DefaultCallerDepth = 2

	logger     *log.Logger
	logPrefix  = ""
	levelFlags = []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}

	currentDate string
	mutex       sync.Mutex
)

const (
	DEBUG Level = iota
	INFO
	WARNING
	ERROR
	FATAL
)

// Setup initialize the log instance
func Setup() {
	var err error
	filePath := configs.C.App.LogSavePath
	fileName := getLogFileName()
	F, err = file.MustOpen(fileName, filePath)
	if err != nil {
		log.Fatalf("logging.Setup err: %v", err)
	}

	logger = log.New(F, DefaultPrefix, log.LstdFlags)

	currentDate = time.Now().Format("2006-01-02")

	go watchLogFile() // Start watching for date changes
}

func getLogFileName() string {
	return fmt.Sprintf("%s%s.%s",
		configs.C.App.LogSaveName,
		time.Now().Format("20060102"),
		configs.C.App.LogFileExt,
	)
}

// Debug output logs at debug level
func Debug(v ...interface{}) {
	setPrefix(DEBUG)
	logger.Println(v)
}

// Info output logs at info level
func Info(v ...interface{}) {
	setPrefix(INFO)
	logger.Println(v)
}

// Warn output logs at warn level
func Warn(v ...interface{}) {
	setPrefix(WARNING)
	logger.Println(v)
}

// Error output logs at error level
func Error(v ...interface{}) {
	setPrefix(ERROR)
	logger.Println(v)
}

// Fatal output logs at fatal level
func Fatal(v ...interface{}) {
	setPrefix(FATAL)
	logger.Fatalln(v)
}

// setPrefix set the prefix of the log output
func setPrefix(level Level) {
	_, file, line, ok := runtime.Caller(DefaultCallerDepth)
	if ok {
		logPrefix = fmt.Sprintf("[%s][%s:%d]", levelFlags[level], filepath.Base(file), line)
	} else {
		logPrefix = fmt.Sprintf("[%s]", levelFlags[level])
	}

	logger.SetPrefix(logPrefix)
}

func watchLogFile() {
	for {
		time.Sleep(1 * time.Minute) // Check every minute
		if time.Now().Format("2006-01-02") != currentDate {
			mutex.Lock()
			updateLogFile()
			mutex.Unlock()
		}
	}
}

// updateLogFile closes the current log file and opens a new one with the new date
func updateLogFile() {
	if err := F.Close(); err != nil {
		Error("Update Log File,Failed to close log file: ", err)
	}

	filePath := configs.C.App.LogSavePath
	fileName := getLogFileName()
	var err error
	F, err = file.MustOpen(fileName, filePath)
	if err != nil {
		Error("Update Log File,Failed to open new log file: ", err)
	}

	logger.SetOutput(F)
	currentDate = time.Now().Format("2006-01-02")

	removeOldLogFiles(filePath)
}

func removeOldLogFiles(filePath string) {

	files, err := os.ReadDir(filePath)
	if err != nil {
		Error("Remove Old Log Files,Failed to read log directory: ", err)
	}

	if len(files) <= configs.C.App.MaxLogFiles {
		return
	}

	for i := 0; i < len(files)-configs.C.App.MaxLogFiles; i++ {
		err = os.Remove(filepath.Join(filePath, files[i].Name()))
		if err != nil {
			Error("Remove Old Log Files,Failed to remove old log file: ", err)
		}
	}
}
