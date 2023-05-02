package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

var (
	debugLog    = log.New(os.Stdout, "\033[36m[debug]\033[0m ", log.LstdFlags|log.Lshortfile)
	infoLog     = log.New(os.Stdout, "\033[32m[info] \033[0m ", log.LstdFlags|log.Lshortfile)
	warningLog  = log.New(os.Stdout, "\033[33m[warn] \033[0m ", log.LstdFlags|log.Lshortfile)
	errorLog    = log.New(os.Stdout, "\033[31m[error]\033[0m ", log.LstdFlags|log.Lshortfile)
	loggers     = []*log.Logger{debugLog, infoLog, warningLog, errorLog}
	loggerLabel = []string{"debug", "info", "warning", "error"}
	mu          sync.Mutex
)

// log methods
var (
	Debug  = debugLog.Println
	Debugf = debugLog.Printf
	Info   = infoLog.Println
	Infof  = infoLog.Printf
	Warn   = warningLog.Println
	Warnf  = warningLog.Printf
	Error  = errorLog.Println
	Errorf = errorLog.Printf
)

const (
	DebugLevel = iota
	InfoLevel
	WarningLevel
	ErrorLevel
	Off
)

// SetLevel controls log level
func SetLevel(level int) {
	mu.Lock()
	defer mu.Unlock()
	for i, logger := range loggers {
		if i < level {
			logger.SetOutput(io.Discard)
		} else {
			logger.SetOutput(os.Stdout)
		}
	}
}

// LogToFile output log to file
// e.g. logFolder/yyyymmdd.<level>>.log
func LogToFile(logFolder string, level int) {
	mu.Lock()
	defer mu.Unlock()
	if _, err := os.Stat(logFolder); os.IsNotExist(err) {
		err2 := os.Mkdir(logFolder, 0777)
		if err2 != nil {
			log.Fatalf("error create log folder: %v", err)
		}
	}
	for i, logger := range loggers {
		if i >= level {
			fileName := fmt.Sprintf("%v.%v.log", time.Now().Format("20060102"), loggerLabel[i])
			f, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
			if err != nil {
				log.Fatalf("error writing log to file: %v", err)
			}
			//defer f.Close()
			logger.SetOutput(f)
		}
	}
}
