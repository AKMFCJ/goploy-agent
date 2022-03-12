package core

import (
	"fmt"
	"github.com/zhenorzz/goploy-agent/config"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LogLevel is log level
type LogLevel string

// log level
const (
	TRACE   LogLevel = "TRACE: "
	WARNING LogLevel = "WARNING: "
	INFO    LogLevel = "INFO: "
	ERROR   LogLevel = "ERROR: "
)

// Log information to file with logged day
func Log(lv LogLevel, content string) {
	var logFile io.Writer
	logPathEnv := config.Toml.Log.Path
	if strings.ToLower(logPathEnv) == "stdout" {
		logFile = os.Stdout
	} else {
		logPath, err := filepath.Abs(logPathEnv)
		if err != nil {
			fmt.Println(err.Error())
		}
		if _, err := os.Stat(logPath); err != nil && os.IsNotExist(err) {
			err := os.Mkdir(logPath, os.ModePerm)
			if nil != err {
				fmt.Println(err.Error())
			}
		}
		file := logPath + "/"
		if config.Toml.Log.Split {
			file += time.Now().Format("20060102") + ".log"
		} else {
			file += "runtime.log"
		}
		logFile, err = os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
		if nil != err {
			fmt.Println(err.Error())
		}
	}

	logger := log.New(logFile, string(lv), log.LstdFlags|log.Lshortfile)
	logger.Output(2, content)
}
