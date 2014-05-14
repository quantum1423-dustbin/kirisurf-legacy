package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/KirisurfProject/kilog"
)

var __log_lock sync.Mutex

var __log_counter = 0

const LOG_SHOW_FILENAMES = false

var __log_min_level = "debug"

func __write_log(level string, format string, args ...interface{}) {
	__log_lock.Lock()
	defer __log_lock.Unlock()
	msg := fmt.Sprintf(format, args...)
	now := time.Now().UTC()
	const layout = "MST 2006-01-02 15:04:05.000"
	nowstring := now.Format(layout)
	fmt.Printf("#%d\t%s >>> %s ::: %s\n", __log_counter, nowstring, level, msg)
	__log_counter++
}

func DEBUG(format string, args ...interface{}) {
	kilog.Debug(format, args...)
}

func INFO(format string, args ...interface{}) {
	kilog.Info(format, args...)
}

func WARNING(format string, args ...interface{}) {
	kilog.Warning(format, args...)
}

func CRITICAL(format string, args ...interface{}) {
	kilog.Critical(format, args...)
}
