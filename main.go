package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	repo       = "https://github.com/pcm-dpc/DPC-Bollettini-Criticita-Idrogeologica-Idraulica"
	path       = "/raw/master/files/all/latest_all.zip"
	separator  = '\t'      // separatore per il CSV
	fileprefix = "allerte" // prefisso per il nome del file
)

const (
	message      = "Scarica da github.com/pcm-dp i bollettini di criticità idrogeologica e idraulica"
	helpMessage  = "mostra queste informazioni"
	destMessage  = "indica la directory in cui salvare il risultato"
	roundMessage = "specifica l'intervallo di tempo per la modalità'service';\n" +
		"di default usa '0 16 * * *', ovvero alle 16:00,\n" +
		"vedi le note del DPC github.com/pcm-dpc/DPC-Bollettini-Criticita-Idrogeologica-Idraulica,\n" +
		"Il valore viene espresso con la grammatica per cron;\n" +
		"aiutati con https://crontab.guru in caso di necessità"
	serviceMessage = "rimane attivo dopo l'esecuzione, eseguendo un nuovo\n" +
		"download ad ogni intervallo specificato [vedi --round]"
	localMessage = "specifica un file zip locale da cui estrarre i dati"
)

var (
	service            bool
	dest, local, round string
	command            = &cobra.Command{
		Use:           "alert",
		Long:          message,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return jobManager()
		},
	}
)

func init() {
	command.Flags().BoolP("help", "h", false, helpMessage)
	command.Flags().StringVarP(&dest, "dest", "d", "./", destMessage)
	command.Flags().StringVarP(&local, "file", "f", "", localMessage)
	command.Flags().StringVarP(&round, "round", "r", "0 16 * * *", roundMessage)
	command.Flags().BoolVarP(&service, "service", "s", false, serviceMessage)
	command.MarkFlagsMutuallyExclusive("file", "round")
	command.MarkFlagsMutuallyExclusive("file", "service")
}

func init() {
	l := &lumberjack.Logger{
		Filename:   "alert.log",
		MaxSize:    3, // megabytes
		MaxBackups: 2,
	}
	multi := io.MultiWriter(l, os.Stdout)
	log.SetOutput(multi)
}

func main() {
	if err := command.Execute(); err != nil {
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
