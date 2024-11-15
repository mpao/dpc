package main

import (
	_ "embed"

	"dpc/internal/allerte"
	"dpc/internal/app"

	"github.com/spf13/cobra"
)

const (
	message = "Scarica dai repositories ufficiali del Dipartimento di Protezione Civile gli ultimi\n" +
		"dati disponibili sugli allarmi metereologici e criticità idrogeologica per ogni comune italiano."
	messageAlert = "Scarica i bollettini DPC di criticità idrogeologica e idraulica"
	messageMeteo = "Scarica i bollettini DPC di vigilanza meteorologica"
	helpMessage  = "mostra queste informazioni"
	destMessage  = "indica la directory in cui salvare il risultato"
	proxyMessage = "specifica il proxy da utilizzare"
	dayMessage   = "specifica un intervallo di date da scaricare:\nusa il formato ddmmyyyy per un singolo giorno\n" +
		"oppure ddmmyyyy-ddmmyyyy per un intervallo, estremi inclusi."
	joinMessage     = "in caso di richiesta di più giorni, salva in un unico file"
	originalMessage = "restituisce i topojson originali"
)

var (
	//go:embed help.template
	helpTemplate string
	howto        = appName + " allerte --help\n" +
		appName + " meteo --help\n"
)

var root = &cobra.Command{
	Use:           appName,
	Long:          message,
	Version:       appVersion,
	SilenceErrors: true,
	SilenceUsage:  true,
}

var alert = &cobra.Command{
	Use:           "allerte",
	Short:         messageAlert,
	Long:          messageAlert,
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE:          allerte.Get,
}

func init() {
	root.PersistentFlags().BoolP("help", "h", false, helpMessage)
	root.PersistentFlags().BoolP("version", "v", false, "versione dell'applicazione")
	root.PersistentFlags().StringVarP(&app.Proxy, "proxy", "p", "", proxyMessage)
	root.PersistentFlags().StringVarP(&app.Dest, "dest", "d", "./", destMessage)
	root.AddCommand(alert)
	root.CompletionOptions.DisableDefaultCmd = true
	root.Example = howto
	root.SetHelpTemplate(helpTemplate)
	root.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})
}

func init() {
	alert.Flags().StringVar(&app.Interval, "day", "", dayMessage)
	alert.Flags().BoolVarP(&app.Original, "original", "o", false, originalMessage)
	alert.Flags().BoolVarP(&app.Join, "join", "j", false, joinMessage)
	alert.MarkFlagsMutuallyExclusive("join", "original")
}
