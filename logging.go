package main

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

var __log_lock sync.Mutex

var __log_counter = 0

const LOG_SHOW_FILENAMES = false

var __log_min_level = "debug"

func __write_log(level string, format string, args ...interface{}) {
	__log_lock.Lock()
	defer __log_lock.Unlock()
	msg := fmt.Sprintf(format, args...)
	now := time.Now()
	const layout = "MST 2006-01-02 15:04:05"
	nowstring := now.Format(layout)
	fmt.Printf("#%d\t%s >>> %s ::: %s\n", __log_counter, nowstring, level, msg)
	__log_counter++
}

func DEBUG(format string, args ...interface{}) {
	if __log_min_level != "debug" {
		return
	}
	_, file, line, _ := runtime.Caller(1)
	if LOG_SHOW_FILENAMES {
		format = fmt.Sprintf("%s <<< %s:%d", format, filepath.Base(file), line)
	}
	__write_log("DEBG", format, args...)
}

func INFO(format string, args ...interface{}) {
	if __log_min_level == "warning" || __log_min_level == "critical" {
		return
	}
	_, file, line, _ := runtime.Caller(1)
	if LOG_SHOW_FILENAMES {
		format = fmt.Sprintf("%s <<< %s:%d", format, filepath.Base(file), line)
	}
	__write_log("INFO", format, args...)
}

func WARNING(format string, args ...interface{}) {
	if __log_min_level == "critical" {
		return
	}
	_, file, line, _ := runtime.Caller(1)
	if LOG_SHOW_FILENAMES {
		format = fmt.Sprintf("%s <<< %s:%d", format, filepath.Base(file), line)
	}
	__write_log("WARN", format, args...)
}

func CRITICAL(format string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	if LOG_SHOW_FILENAMES {
		format = fmt.Sprintf("%s <<< %s:%d", format, filepath.Base(file), line)
	}
	__write_log("CRIT", format, args...)
}
