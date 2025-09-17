package main

import (
	"strings"

	"github.com/gofiber/fiber/v2/log"
)

var logLevelTable = map[string]log.Level{
	"trace": log.LevelTrace,
	"debug": log.LevelDebug,
	"info":  log.LevelInfo,
	"warn":  log.LevelWarn,
	"error": log.LevelError,
	"fatal": log.LevelFatal,
	"panic": log.LevelPanic,
}

func setLog(level string) {
	if level == "" {
		log.SetLevel(log.LevelInfo)
		return
	}

	lvName := strings.ToLower(level)

	lv, ok := logLevelTable[lvName]
	if ok {
		log.SetLevel(lv)
	} else {
		log.SetLevel(log.LevelInfo)
	}
}
