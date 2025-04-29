package utils

import (
	"log"
	"os"
	"time"
)

var (
	infoLogger  = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
)

func LogInfo(message string) {
	infoLogger.Println(message)
}

func LogError(err error) {
	errorLogger.Println(err)
}
