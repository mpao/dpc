package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

func init() {
	l := &lumberjack.Logger{
		Filename:   applicationName + ".log",
		MaxSize:    3, // megabytes
		MaxBackups: 2,
	}
	multi := io.MultiWriter(l, os.Stdout)
	log.SetOutput(multi)
}

func main() {
	if err := root.Execute(); err != nil {
		fmt.Println(overrideErrMessage(err))
		os.Exit(1)
	}
}

// overrideErrMessage sovrascrive i messaggi di errore generati da cobra.
// Per alcuni argomenti, l'output originale è piuttosto criptico per utenti
func overrideErrMessage(err error) error {
	original := err.Error()
	var msg string
	switch {
	case strings.Contains(original, "[target"):
		msg = "gli argomenti sono mutualmente esclusivi"
	case strings.Contains(original, "needs an argument"):
		msg = "è necessario specificare un valore per l'argomento"
	default:
		msg = original
	}
	return errors.New(msg + "\nusa -h per ulteriore aiuto")
}
