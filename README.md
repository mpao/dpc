# Bollettini Dipartimento della Protezione Civile

### Descrizione

Esegue il download  degli ultimi bollettini sugli allarmi metereologici e criticità idrogeologica
dal sito del Dipartimento della Protezione Civile, in formato **tsv** o **topojson**

* https://github.com/pcm-dpc/DPC-Bollettini-Vigilanza-Meteorologica
* https://github.com/pcm-dpc/DPC-Bollettini-Criticita-Idrogeologica-Idraulica
* https://github.com/opendatasicilia/comuni-italiani

### Compilazione (opzionale)

Installare [Go](https://go.dev/dl/) se non presente, per la compilazione dei sorgenti; è possibile utilizzare [task](https://taskfile.dev/)

```bash
$> go install github.com/go-task/task/v3/cmd/task@latest
$> task release
```

### Download

L'eseguibile è disponibile per le piattaforme più diffuse al link:

> https://github.com/mpao/dpc/releases


### Manuale d'uso

```text
Scarica dai repositories ufficiali del Dipartimento di Protezione Civile gli ultimi
dati disponibili sugli allarmi metereologici e criticità idrogeologica per ogni comune italiano.

Uso:
  dpc [comando]

Esempi:
dpc allerte --help
dpc meteo --help


Comandi disponibili:
  allerte     Scarica i bollettini DPC di criticità idrogeologica e idraulica

Opzioni:
  -d, --dest string    indica la directory in cui salvare il risultato (default "./")
  -h, --help           mostra queste informazioni
  -p, --proxy string   specifica il proxy da utilizzare
  -v, --version        versione dell'applicazione

Usa "dpc [comando] --help" per maggiori informazioni sul comando.
```

### Esempi d'uso

```bash
# mostra help
> dpc.exe -h

# utilizza proxy
> dpc.exe -p user:password@proxy.domain.com:port

# specifica la destinazione per salvare i files
> dpc.exe [comando] -d C:\tmp

# scarica i dati per il giorno passato come parametro ddmmyyyy
> dpc.exe [comando] --day 13082024

# scarica tutti i dati per i giorni contenuti nell'intervallo 
> dpc.exe [comando] --day 13082024-17112024

# scarica tutti i dati per i giorni contenuti nell'intervallo in un unico file
> dpc.exe [comando] --day 13082024-17112024 -j

# scarica i dati nel formato topojson originale (incompatibile con -j)
> dpc.exe [comando] -o
```
