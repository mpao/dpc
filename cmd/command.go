package main

import (
	_ "embed"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/mpao/dpc/internal/allerte"
	"github.com/mpao/dpc/internal/app"
	"github.com/mpao/dpc/internal/comuni"
	"github.com/mpao/dpc/internal/meteo"
	"github.com/spf13/cobra"
)

const (
	message = "Scarica dai repositories ufficiali del Dipartimento di Protezione Civile i\n" +
		"dati sugli allarmi metereologici e criticità idrogeologica per ogni comune italiano."
	messageAlert = "Scarica i bollettini DPC di criticità idrogeologica e idraulica"
	messageMeteo = "Scarica i bollettini DPC di vigilanza meteorologica"
	helpMessage  = "mostra queste informazioni"
	destMessage  = "indica la directory in cui salvare il risultato"
	proxyMessage = "specifica il proxy da utilizzare"
	dayMessage   = "specifica un intervallo di date da scaricare:\nusa il formato ddmmyyyy per un singolo giorno\n" +
		"oppure ddmmyyyy-ddmmyyyy per un intervallo, estremi inclusi."
	joinMessage     = "in caso di richiesta di più giorni, salva in un unico file"
	originalMessage = "restituisce i topojson originali; è incompatibile con -j"
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

var alertCmd = &cobra.Command{
	Use:           "allerte",
	Short:         messageAlert,
	Long:          messageAlert,
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return uxWaitingMessage("allerte", allerte.Get)
	},
}

var meteoCmd = &cobra.Command{
	Use:           "meteo",
	Short:         messageMeteo,
	Long:          messageMeteo,
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return uxWaitingMessage("meteo", meteo.Get)
	},
}

func init() {
	root.PersistentFlags().BoolP("help", "h", false, helpMessage)
	root.PersistentFlags().BoolP("version", "v", false, "versione dell'applicazione")
	root.PersistentFlags().StringVarP(&app.Proxy, "proxy", "p", "", proxyMessage)
	root.PersistentFlags().StringVarP(&app.Dest, "dest", "d", "./", destMessage)
	root.AddCommand(alertCmd)
	root.AddCommand(meteoCmd)
	root.CompletionOptions.DisableDefaultCmd = true
	root.Example = howto
	root.SetHelpTemplate(helpTemplate)
	root.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})
}

func init() {
	alertCmd.Flags().StringVar(&app.Interval, "day", "", dayMessage)
	alertCmd.Flags().BoolVarP(&app.Original, "original", "o", false, originalMessage)
	alertCmd.Flags().BoolVarP(&app.Join, "join", "j", false, joinMessage)
	alertCmd.MarkFlagsMutuallyExclusive("join", "original")
}

func init() {
	meteoCmd.Flags().StringVar(&app.Interval, "day", "", dayMessage)
	meteoCmd.Flags().BoolVarP(&app.Original, "original", "o", false, originalMessage)
	meteoCmd.Flags().BoolVarP(&app.Join, "join", "j", false, joinMessage)
	meteoCmd.MarkFlagsMutuallyExclusive("join", "original")
}

// uxWaitingMessage manda in output messaggio di attesa per
// dare feedback all'utente sul procedere delle operazioni
func uxWaitingMessage(label string, f func() error) error {
	start := time.Now()
	var wg sync.WaitGroup
	done := make(chan struct{})
	wg.Add(1)
	go spinner(&wg, done)
	err := f()
	close(done)
	wg.Wait()
	if err != nil {
		return err
	}
	slog.Info(strings.ToUpper(label)+": dati salvati",
		"comuni", comuni.Amount(),
		"durata", fmt.Sprintf("%.2f secondi", time.Since(start).Seconds()),
	)
	return nil
}

func spinner(wg *sync.WaitGroup, done chan struct{}) {
	defer wg.Done()
	chars := `-\|/`
	i := 0
	ticker := time.NewTicker(70 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			fmt.Print("\r \r")
			return
		case <-ticker.C:
			fmt.Printf("\r %c", chars[i])
			i = (i + 1) % len(chars)
		}
	}
}
