package main

import (
	_ "embed"

	"github.com/spf13/cobra"
)

const (
	message = "Scarica dai repositories ufficiali del Dipartimento di Protezione Civile gli ultimi\n" +
		"dati disponibili sugli allarmi metereologici e criticità idrogeologica."
	messageAlert = "Scarica i bollettini DPC di criticità idrogeologica e idraulica"
	messageMeteo = "Scarica i bollettini DPC di vigilanza meteorologica"
	helpMessage  = "mostra queste informazioni"
	destMessage  = "indica la directory in cui salvare il risultato"
	roundMessage = "specifica l'intervallo di tempo per la modalità'service';\n" +
		"di default usa '0 16 * * *', ovvero alle 16:00,\n" +
		"vedi le note del DPC github.com/pcm-dpc.\n" +
		"Il valore viene espresso con la grammatica per cron;\n" +
		"aiutati con https://crontab.guru in caso di necessità"
	serviceMessage = "rimane attivo dopo l'esecuzione, eseguendo un nuovo\n" +
		"download ad ogni intervallo specificato [vedi --round]"
	localMessage = "specifica un file zip locale da cui estrarre i dati"
)

var (
	howto              = applicationName + " meteo --help\n" + applicationName + " allerte --help"
	service            bool
	dest, local, round string
	//go:embed help.template
	helpTemplate string
)

var root = &cobra.Command{
	Use:           applicationName,
	Long:          message,
	Version:       applicationVersion,
	SilenceErrors: true,
	SilenceUsage:  true,
}

var meteo = &cobra.Command{
	Use:           "meteo",
	Short:         messageMeteo,
	Long:          messageMeteo,
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		t := target{
			name: "meteo",
			repo: repoMeteo,
			url:  repoMeteo + path,
		}
		return jobManager(t)
	},
}

var allerte = &cobra.Command{
	Use:           "allerte",
	Short:         messageAlert,
	Long:          messageAlert,
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		t := target{
			name: "allerte",
			repo: repoAlert,
			url:  repoAlert + path + "latest_all.zip",
		}
		return jobManager(t)
	},
}

func init() {
	root.Flags().BoolP("help", "h", false, helpMessage)
	root.Flags().BoolP("version", "v", false, "versione dell'applicazione")
	root.AddCommand(meteo)
	root.AddCommand(allerte)
	root.CompletionOptions.DisableDefaultCmd = true
	root.Example = howto
	root.SetHelpTemplate(helpTemplate)
	root.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})
}

func init() {
	meteo.Flags().BoolP("help", "h", false, helpMessage)
	meteo.Flags().StringVarP(&dest, "dest", "d", "./", destMessage)
	meteo.Flags().StringVarP(&local, "file", "f", "", localMessage)
	meteo.Flags().StringVarP(&round, "round", "r", "0 16 * * *", roundMessage)
	meteo.Flags().BoolVarP(&service, "service", "s", false, serviceMessage)
	meteo.MarkFlagsMutuallyExclusive("file", "round")
	meteo.MarkFlagsMutuallyExclusive("file", "service")
}

func init() {
	allerte.Flags().BoolP("help", "h", false, helpMessage)
	allerte.Flags().StringVarP(&dest, "dest", "d", "./", destMessage)
	allerte.Flags().StringVarP(&local, "file", "f", "", localMessage)
	allerte.Flags().StringVarP(&round, "round", "r", "0 16 * * *", roundMessage)
	allerte.Flags().BoolVarP(&service, "service", "s", false, serviceMessage)
	allerte.MarkFlagsMutuallyExclusive("file", "round")
	allerte.MarkFlagsMutuallyExclusive("file", "service")
}
