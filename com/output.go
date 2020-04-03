package com

import "fmt"

// Global log level
var logLevel = NORMAL

// LogLevel is intended to set the verbosity of the output
type LogLevel int

const (
	// NAMEONLY prints out only the file paths of changed files
	// This is used for piping output to other programs
	NAMEONLY = -1

	// SILENT has no output
	SILENT LogLevel = iota
	// ERROR prints only error output
	ERROR
	// NORMAL prints out standard verbose output
	NORMAL
	// DEBUG prints out extra verbose output
	DEBUG
)

// String from level
func (level LogLevel) String() string {
	switch level {
	case NAMEONLY:
		return "NAME_ONLY"
	case SILENT:
		return "SILENT"
	case ERROR:
		return "ERROR"
	case NORMAL:
		return "NORMAL"
	case DEBUG:
		return "DEBUG"
	}
	return "UNKNOWN"
}

// LogLevelFrom string
func LogLevelFrom(level string) LogLevel {
	switch level {
	case "NAME_ONLY", "name-only", "o", "-1":
		return NAMEONLY
	case "SILENT", "silent", "s", "0":
		return SILENT
	case "ERROR", "error", "e", "1":
		return ERROR
	case "NORMAL", "normal", "n", "2":
		return NORMAL
	case "DEBUG", "debug", "d", "3":
		return DEBUG
	default:
		return NORMAL
	}
}

// SetLogLevel turns debug messaging on and off globally
func SetLogLevel(level LogLevel) {
	logLevel = level
}

// GetLogLevel returns the current log level
func GetLogLevel() LogLevel {
	return logLevel
}

// Outputln will println if level and setting match nameOnly, or if level is at or below logLevel
func Outputln(level LogLevel, a ...interface{}) (n int, err error) {
	if logLevel == SILENT {
		// Ignore
	} else if logLevel == NAMEONLY {
		// Only print NAMEONLY when level matches exact
		if logLevel == level {
			return fmt.Println(a...)
		}
	} else if level <= logLevel {
		return fmt.Println(a...)
	}

	err = fmt.Errorf("Log level <"+logLevel.String()+"> skips output at level:", level)
	return
}

// Errorln will print output at debug level
func Errorln(a ...interface{}) (n int, err error) {
	return Outputln(ERROR, a...)
}

// Println will print output at normal level
func Println(a ...interface{}) (n int, err error) {
	return Outputln(NORMAL, a...)
}

// Debugln will print output at debug level
func Debugln(a ...interface{}) (n int, err error) {
	return Outputln(DEBUG, a...)
}
