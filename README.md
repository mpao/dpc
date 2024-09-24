# Bollettini Dipartimento della Protezione Civile

### Descrizione

Esegue il download  degli ultimi bollettini sugli allarmi metereologici e criticità idrogeologica
dal sito del Dipartimento della Protezione Civile.

### Compilazione (opzionale)

```bash
$> go install github.com/go-task/task/v3/cmd/task@latest
$> task release
```

### Manuale d'uso

```text
Scarica dai repositories ufficiali del Dipartimento di Protezione Civile gli ultimi
dati disponibili sugli allarmi metereologici e criticità idrogeologica.

Uso:
  dpc [comando]

Esempi:
dpc meteo --help
dpc allerte --help

Comandi disponibili:
  allerte     Scarica i bollettini DPC di criticità idrogeologica e idraulica
  meteo       Scarica i bollettini DPC di vigilanza meteorologica

Opzioni:
  -h, --help           mostra queste informazioni
  -p, --proxy string   specifica il proxy da utilizzare
  -v, --version        versione dell'applicazione

Usa "dpc [comando] --help" per maggiori informazioni sul comando.
```

### Esempi d'uso

```bash
# mostra help
> dpc.exe -h

# specifica la destinazione per salvare i files
> dpc.exe [comando] -d C:\tmp

# viene eseguito come servizio, non termina alla fine del job
> dpc.exe [comando] -s 

# specifica l'orario di esecuzione dei jobs
# con grammatica cron, usa https://crontab.guru per aiutarti
> dpc.exe [comando] -s -r "* */4 * * *"

# elabora un file zip locale
# questi file provengono dal repository del DPC, e sono nella cartella files/xml/
> dpc.exe [comando] -f 20240909_1459.zip
```
