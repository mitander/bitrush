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

var l = log.New(os.Stderr, "", 0)

// sane default: 2
var level loggerLevel = 2

func Debug(msg string) {
	if DebugLevel <= level {
		l.Printf(yellow+"[%+v]"+reset+": %+v \n", DebugLevel.name(), msg)
	}
}

func Info(msg string) {
	if InfoLevel <= level {
		l.Printf(cyan+"[%+v]"+reset+": %+v \n", InfoLevel.name(), msg)
	}
}

func Warning(msg string) {
	if WarningLevel <= level {
		l.Printf(red+"[%+v]"+reset+": %+v \n", WarningLevel.name(), msg)
	}
}

func CLI(msg string) {
	if CLILevel <= level {
		l.Printf("%+v\n", msg)
	}
}

func Fatal(err error) {
	if level >= FatalLevel {
		l.Printf(red+"[%+v]: %d \n", InfoLevel.name(), err)
		log.Printf(red+"[%+v]: Fatal Error - Exiting... \n", InfoLevel.name())
		os.Exit(1)
	}
}

func Level(l loggerLevel) {
	level = l
	Debug(fmt.Sprint("Logger set to: ", l.name()))
}

func Help() {
	CLI("")
	CLI("BitRush")
	CLI("-------")
	CLI("-f [file] (required)")
	CLI("Info: torrent file you want to open")
	CLI("Usage: bitrush -f <torrent file>")
	CLI("")
	CLI("-o [out file] (optional)")
	CLI("Info: output file location - default '.' (current directory)")
	CLI("Usage: bitrush -o <output file>")
	CLI("")
	CLI("-h [help] (optional)")
	CLI("Info: show help menu")
	CLI("Usage: bitrush -h")
	CLI("")
	CLI("-d [debug] (optional")
	CLI("info: enable debug")
	CLI("Usage: bitrush -d")
	CLI("-------")
	CLI("")
}

func NoArgs() {
	CLI("")
	CLI("BitRush")
	CLI("-------")
	CLI("No .torrent file selected!")
	CLI("")
	CLI("Usage: bitrush -f <torrent file>")
	CLI("Help: bitrush -h")
	CLI("-------")
	CLI("")
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
