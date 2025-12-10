package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	appName    = "dpc"  // Name valore di fallback, usa il Taskfile per la definizione
	appVersion = "v0.0" // Version valore di fallback, usa il Taskfile per la definizione
)

func init() {
	l := &lumberjack.Logger{
		Filename:   fmt.Sprintf("./log/%s.log", time.Now().Format("2006-01-02")),
		MaxSize:    1, // megabytes
		MaxBackups: 5,
		MaxAge:     7, // days
	}
	multi := io.MultiWriter(l, os.Stdout)
	log.SetOutput(multi)
}

func main() {
	if err := root.Execute(); err != nil {
		fmt.Printf("\rErrore: %v\n", err)
		os.Exit(1)
	}
}
