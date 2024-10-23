package main

import (
	_ "embed"
	"time"

	"github.com/spf13/cobra"
)

const (
	message = "Scarica dai repositories ufficiali del Dipartimento di Protezione Civile gli ultimi\n" +
		"dati disponibili sugli allarmi metereologici e criticità idrogeologica."
	messageAlert  = "Scarica i bollettini DPC di criticità idrogeologica e idraulica"
	messageMeteo  = "Scarica i bollettini DPC di vigilanza meteorologica"
	messageComuni = "Scarica i bollettini DPC meteo per ogni comune italiano"
	helpMessage   = "mostra queste informazioni"
	destMessage   = "indica la directory in cui salvare il risultato"
	roundMessage  = "specifica l'intervallo di tempo per la modalità'service';\n" +
		"di default usa '0 16 * * *', ovvero alle 16:00,\n" +
		"vedi le note del DPC github.com/pcm-dpc.\n" +
		"Il valore viene espresso con la grammatica per cron;\n" +
		"aiutati con https://crontab.guru in caso di necessità"
	serviceMessage = "rimane attivo dopo l'esecuzione, eseguendo un nuovo\n" +
		"download ad ogni intervallo specificato [vedi --round]"
	localMessage = "specifica un file zip locale, o una URL da cui estrarre i dati"
)

var (
	howto = applicationName + " meteo --help\n" +
		applicationName + " allerte --help\n" +
		applicationName + " comuni --help"
	service                   bool
	dest, local, round, proxy string
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
			url:  "https://github.com/pcm-dpc/DPC-Bollettini-Vigilanza-Meteorologica/raw/master/files/all/",
			filename: func() string {
				today := time.Now()
				return today.Format("20060102") + ".zip"
			},
			fallback: func() string {
				today := time.Now()
				return today.AddDate(0, 0, -1).Format("20060102") + ".zip"
			},
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
			url:  "https://github.com/pcm-dpc/DPC-Bollettini-Criticita-Idrogeologica-Idraulica/raw/master/files/all/",
			filename: func() string {
				return "latest_all.zip"
			},
			fallback: func() string {
				return "latest_all.zip"
			},
		}
		return jobManager(t)
	},
}

var comuni = &cobra.Command{
	Use:           "comuni",
	Short:         messageComuni,
	Long:          messageComuni,
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		t := target{
			name: "comuni",
			url:  "https://github.com/pcm-dpc/DPC-Bollettini-Vigilanza-Meteorologica/raw/master/files/topojson/",
			filename: func() string {
				today := time.Now()
				return today.Format("20060102") + "_oggi.json"
			},
			fallback: func() string {
				today := time.Now()
				return today.AddDate(0, 0, -1).Format("20060102") + "_oggi.json"
			},
		}
		return jobManager(t)
	},
}

func init() {
	root.PersistentFlags().BoolP("help", "h", false, helpMessage)
	root.PersistentFlags().BoolP("version", "v", false, "versione dell'applicazione")
	root.PersistentFlags().StringVarP(&proxy, "proxy", "p", "", "specifica il proxy da utilizzare")
	root.AddCommand(meteo)
	root.AddCommand(allerte)
	root.AddCommand(comuni)
	root.CompletionOptions.DisableDefaultCmd = true
	root.Example = howto
	root.SetHelpTemplate(helpTemplate)
	root.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})
}

func init() {
	meteo.Flags().StringVarP(&dest, "dest", "d", "./", destMessage)
	meteo.Flags().StringVarP(&local, "from", "f", "", localMessage)
	meteo.Flags().StringVarP(&round, "round", "r", "0 16 * * *", roundMessage)
	meteo.Flags().BoolVarP(&service, "service", "s", false, serviceMessage)
	meteo.MarkFlagsMutuallyExclusive("from", "round")
	meteo.MarkFlagsMutuallyExclusive("from", "service")
}

func init() {
	allerte.Flags().StringVarP(&dest, "dest", "d", "./", destMessage)
	allerte.Flags().StringVarP(&local, "from", "f", "", localMessage)
	allerte.Flags().StringVarP(&round, "round", "r", "0 16 * * *", roundMessage)
	allerte.Flags().BoolVarP(&service, "service", "s", false, serviceMessage)
	allerte.MarkFlagsMutuallyExclusive("from", "round")
	allerte.MarkFlagsMutuallyExclusive("from", "service")
}

func init() {
	comuni.Flags().StringVarP(&dest, "dest", "d", "./", destMessage)
	comuni.Flags().StringVarP(&round, "round", "r", "0 16 * * *", roundMessage)
	comuni.Flags().BoolVarP(&service, "service", "s", false, serviceMessage)
}
