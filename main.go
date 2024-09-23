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

const (
	repoMeteo = "https://github.com/pcm-dpc/DPC-Bollettini-Vigilanza-Meteorologica"
	repoAlert = "https://github.com/pcm-dpc/DPC-Bollettini-Criticita-Idrogeologica-Idraulica"
	path      = "/raw/master/files/all/"
	separator = '\t' // separatore per il CSV
)

var (
	applicationName    = "dpc"  // valore di fallback, usa il Taskfile per la definizione
	applicationVersion = "v0.0" // valore di fallback, usa il Taskfile per la definizione
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
