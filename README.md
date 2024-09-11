# Bollettini di criticità idrogeologica e idraulica

### Descrizione

```text
Scarica l'ultimo bollettino di criticità idrogeologica e idraulica dal sito del Dipartimento della Protezione Civile.
```

### Compilazione

```bash
$> go install github.com/go-task/task/v3/cmd/task@latest
$> task release
```

### Manuale d'uso

```text
Scarica da github.com/pcm-dp i bollettini di criticità idrogeologica e idraulica

Usage:
  alert [flags]

Flags:
  -d, --dest string    indica la directory in cui salvare il risultato (default "./")
  -f, --file string    specifica un file zip locale da cui estrarre i dati
  -h, --help           mostra queste informazioni
  -r, --round string   specifica l'intervallo di tempo per la modalità'service';
                       di default usa '0 16 * * *', ovvero alle 16:00,
                       vedi le note del DPC github.com/pcm-dpc/DPC-Bollettini-Criticita-Idrogeologica-Idraulica,
                       Il valore viene espresso con la grammatica per cron;
                       aiutati con https://crontab.guru in caso di necessità (default "0 16 * * *")
  -s, --service        rimane attivo dopo l'esecuzione, eseguendo un nuovo
                       download ad ogni intervallo specificato [vedi --round]
```

### Esempi d'uso

```bash
# mostra help
> alert.exe -h

# specifica la destinazione per salvare i files
> alert.exe -d C:\tmp

# viene eseguito come servizio, non termina alla fine del job
> alert.exe -s 

# specifica l'orario di esecuzione dei jobs
# con grammatica cron, usa https://crontab.guru per aiutarti
> alert.exe -s -r "* */4 * * *"

# scarica il file delle 14 per la giornata
> alert.exe -f 20240909_1459.zip

