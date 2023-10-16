// Copyright (c) 2023 Julian Müller (ChaoticByte)

package main

import (
	"log"
	"os"
)

type Logger struct {
	LogLevel int
	ErrorLogger *log.Logger		// 0
	WarningLogger *log.Logger	// 1
	InfoLogger *log.Logger		// 2
	DebugLogger *log.Logger		// 3
}

func (l *Logger) error(msg any) {
	l.ErrorLogger.Println(msg)
}

func (l *Logger) warn(msg any) {
	if l.LogLevel >= 1 {
		l.WarningLogger.Println(msg)
	}
}

func (l *Logger) info(msg any) {
	if l.LogLevel >= 2 {
		l.InfoLogger.Println(msg)
	}
}

func (l *Logger) debug(msg any) {
	if l.LogLevel >= 3 {
		l.DebugLogger.Println(msg)
	}
}

func NewLogger(loglevel int) Logger {
	l := Logger{}
	l.LogLevel = loglevel
	l.ErrorLogger 	= log.New(os.Stderr, "ERROR ", log.Ldate|log.Ltime|log.Lmicroseconds)
	l.WarningLogger = log.New(os.Stderr, "WARN  ", log.Ldate|log.Ltime|log.Lmicroseconds)
	l.InfoLogger 	= log.New(os.Stderr, "INFO  ", log.Ldate|log.Ltime|log.Lmicroseconds)
	l.DebugLogger 	= log.New(os.Stderr, "DEBUG ", log.Ldate|log.Ltime|log.Lmicroseconds)
	return l
}
