package logger

import (
	"fmt"
	"log"
	"os"
)

type loggerLevel uint8

const (
	FatalLevel   loggerLevel = 0
	CLILevel     loggerLevel = 1
	WarningLevel loggerLevel = 2
	InfoLevel    loggerLevel = 3
	DebugLevel   loggerLevel = 4
)

const (
	reset  = "\033[0m"
	red    = "\033[31m"
	yellow = "\033[33m"
	green  = "\033[32m"
	cyan   = "\033[36m"
)

// sane default: 2
var level loggerLevel = 2

func Debug(msg string) {
	if DebugLevel <= level {
		log.Printf(yellow+"[%+v]"+reset+": %+v \n", DebugLevel.name(), msg)
	}
}

func Info(msg string) {
	if InfoLevel <= level {
		log.Printf(cyan+"[%+v]"+reset+": %+v \n", InfoLevel.name(), msg)
	}
}

func Warning(msg string) {
	if WarningLevel <= level {
		log.Printf(red+"[%+v]"+reset+": %+v \n", WarningLevel.name(), msg)
	}
}

func CLI(msg string) {
	if CLILevel <= level {
		log.Printf(cyan+"[%+v]"+reset+": %+v\n", CLILevel.name(), msg)
	}
}

func Fatal(err error) {
	if level >= FatalLevel {
		log.Printf("[%+v]: %+v \n", InfoLevel.name(), err)
		log.Printf("[%+v]: Fatal Error - Exiting... \n", InfoLevel.name())
		os.Exit(1)
	}
}

func Level(l loggerLevel) {
	level = l
	Debug(fmt.Sprint("Logger set to: ", l.name()))
}

func (l loggerLevel) name() string {
	switch l {
	case InfoLevel:
		return "INFO"
	case DebugLevel:
		return "DEBUG"
	case WarningLevel:
		return "WARNING"
	case CLILevel:
		return "CLI"
	case FatalLevel:
		return "FATAL"
	default:
		Warning("logger name: invalid level")
		return ""
	}
}
