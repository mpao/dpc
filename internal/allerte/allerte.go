package allerte

// Scaricare le atterte comporta un lavoro aggiuntivo poiché non è garantita la
// presenza di uno e un solo file al giorno, e tale file si presenta con il nome
// uguale all'orario della pubblicazione, es. 20241101_1504
// Poiché l'orario è variabile, i nomi dei files non sono noti in partenza.
//
// Il processo parte quindi dal comando utente, eseguito dalla funzione Get
// che si occupa di ricavare i nomi files dalle API di github. Tali dati
// vengono filtrati in base a quanto richiesto dall'utente con il parametro --day
//
// Questa lista viene quindi passata al processo di download e salvataggio,
// utilizzando la funzione relativa al formato scelto dall'utente.
//
// - Formato JSON, `writeJSON`:
// Per ogni elemento della lista, viene scaricato il file e salvato.
// - Formato TSV, `writeCSV`:
// Per ogni elemento della lista, viene scaricato il file, da cui si ricavano
// gli eventi per ogni comune italiano attraverso la funzione `extract`. Questi dati
// vengono poi correlati ai comuni per avere le informazioni geografiche e di
// popolazione. Una volta creata questa "tabella", i dati sono salvati su file

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/mpao/dpc/internal/app"
	"github.com/mpao/dpc/internal/comuni"
)

const (
	domain         = "https://api.github.com/repos/pcm-dpc/DPC-Bollettini-Criticita-Idrogeologica-Idraulica/"
	fileURL        = "https://github.com/pcm-dpc/DPC-Bollettini-Criticita-Idrogeologica-Idraulica/raw/master/"
	dateLimit      = "01012020"         // data minima per richiesta dati, formato ddmmyyyy
	utf8Laceholder = "�"                // il dataset utilizza questo carattere come carattere UTF8 non identificato
	filenameCSV    = "allerte"          // nome del file salvato per i CSV
	filenameJSON   = "allerte-topojson" // nome del file salvato per i JSON
)

// Get comando per il download delle allerte DPC
func Get() error {
	list, err := topojsonList()
	if err != nil {
		return err
	}
	nodes, err := filterNodes(app.Interval, list)
	if err != nil {
		return err
	}
	if app.Original {
		return writeJSON(nodes)
	}
	return writeCSV(nodes)
}

// filterNodes restituisce un sottoinsieme della slice passata come argomento
// in base all'intervallo di date passate. Nello specifico, avendo scaricato
// tutte le date a partire dal 2020, serve a selezionare solo quelle decise
// dall'utente. Se nessuna data è stata specificata, restituisce la più recente.
func filterNodes(interval string, nodes []node) ([]node, error) {
	var out []node
	from, to, err := app.ParseDayParam(interval, dateLimit)
	if err != nil {
		return nil, err
	}
	if interval == "" {
		// restituisce gli ultimi due giorni disponibili,
		// ovvero la data più recente, e la sua successiva come previsione (tomorrow)
		return append(out, nodes[len(nodes)-2:]...), nil
	}
	for _, n := range nodes {
		// prima del check di ugualianza sulle date, elimina
		// eventuali valori di orario e timezone. trovo sia
		// più leggibile farlo come segue invece di usare time.Date
		d, _ := time.Parse("20060102", n.date.Format("20060102"))
		if d.Before(from) || d.After(to) {
			continue
		}
		out = append(out, n)
	}
	return out, nil
}

// topojsonList scarica la lista di files topojson dalle API github
func topojsonList() ([]node, error) {
	url := domain + "git/trees/master?recursive=1"
	r, err := app.HTTPClient().Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	b, _ := io.ReadAll(r.Body)
	if r.StatusCode != http.StatusOK {
		return nil, errors.New(string(b))
	}
	var t tree
	_ = json.Unmarshal(b, &t)
	if t.Truncated {
		slog.Warn("la lista dei files disponibili è stata scaricata parzialmente.")
	}
	rx := regexp.MustCompile(`files/topojson/\d{8}_\d{4}_today\.json`)
	files := make([]node, 0, len(t.Tree))
	for _, v := range t.Tree {
		if rx.MatchString(v.Filename) {
			v.addDate()
			// se importo la URL dal nodo originale, questo comporta un utilizzo delle API
			// che hanno un rate-limit spiegato qui: https://shorturl.at/Uk0E5
			// Poiché il numero di chiamate superano abbondantemente il limite imposto,
			// la soluzione è effettuare solo la chiamata per il tree main per ottenere
			// i nomi dei files, per poi accederci direttamente dal repository.
			// Ti ricordo che senza sapere il nome dei files non li troverai mai visto
			// che vengono pubblicati con nome variabile a seconda dell'orario;
			// ecco il perché ti tutto questo giro
			v.url = fileURL + v.Filename
			files = append(files, v)
		}
	}
	slices.SortFunc(files, func(a, b node) int {
		if a.date.Before(b.date) {
			return -1
		}
		return 1
	})
	// aggiungi l'elemento di "domani"
	latest := files[len(files)-1]
	tomorrow := node{
		date:     latest.date.AddDate(0, 0, 1),
		Filename: strings.ReplaceAll(latest.Filename, "today", "tomorrow"),
		url:      fileURL + strings.ReplaceAll(latest.Filename, "today", "tomorrow"),
	}
	files = append(files, tomorrow)
	out := deleteDuplicate(files)
	return out, nil
}

// deleteDuplicate elimina eventuali duplicati dalla slices in argomento.
// Può capitare che il tree contenga due o più pubblicazioni giornaliere, ma a me
// interessa solo l'ultima, la più recente in ordine temporale
func deleteDuplicate(in []node) []node {
	out := make([]node, 0, len(in))
	shift := 1 // conta quanti elementi rimango indietro nel loop, i duplicati quindi
	for i, n := range in {
		n.date = time.Date(n.date.Year(), n.date.Month(), n.date.Day(), 0, 0, 0, 0, time.Local)
		if i != 0 && app.DayEqual(out[i-shift].date, n.date) {
			out[i-shift] = n
			shift++
			continue
		}
		out = append(out, n)
	}
	return slices.Clip(out)
}

// topojson scarica il topojson
func topojson(n node) ([]byte, error) {
	r, err := app.HTTPClient().Get(n.url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	body, _ := io.ReadAll(r.Body)
	if r.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
	}
	// se non è un json restituisci slice vuota
	if body[0] != byte('{') {
		return []byte{}, nil
	}
	return body, nil
}

// writeJSON scarica i file json senza nessuna modifica
func writeJSON(nodes []node) error {
	var wg sync.WaitGroup
	for _, n := range nodes {
		wg.Add(1)
		go func(n node) error {
			defer wg.Done()
			b, err := topojson(n)
			if err != nil {
				slog.Error("fallito", "giorno", n.date.Format("02/01/2006"), "errore", err.Error())
			}
			name := filenameJSON + "-" + n.date.Format("20060102")
			if err := app.SaveBytes(name, b); err != nil {
				return err
			}
			return nil
		}(n) //nolint //errcheck ignora
	}
	wg.Wait()
	return nil
}

// writeCSV salva i dati elaborati su un file TSV
func writeCSV(nodes []node) error {
	headers := []string{
		"data",
		"idrogeologico",
		"idraulico",
		"temporali",
		"nome",
		"provincia",
		"sigla",
		"zona",
		"regione",
		"popolazione",
		"latitudine",
		"longitudine",
		"info",
	}
	joined := make([][]string, 0, 80_000)
	joined = slices.Insert(joined, 0, headers)
	var wg sync.WaitGroup
	var mutex sync.Mutex
	for _, n := range nodes {
		wg.Add(1)
		go func(n node) error {
			defer wg.Done()
			payload := make([][]string, 0, 8_000)
			collection := events(n)
			for _, v := range collection {
				payload = append(payload, v.CSV())
			}
			mutex.Lock()
			joined = append(joined, payload...)
			mutex.Unlock()
			if app.Join {
				name := filenameCSV + "-" + nodes[0].date.Format("20060102") + nodes[len(nodes)-1].date.Format("20060102")
				if err := app.SaveCSV(name, joined); err != nil {
					return err
				}
			} else {
				name := filenameCSV + "-" + n.date.Format("20060102")
				payload = slices.Insert(payload, 0, headers)
				if err := app.SaveCSV(name, payload); err != nil {
					return err
				}
			}
			return nil
		}(n) //nolint //errcheck ignora
	}
	wg.Wait()
	return nil
}

// events scarica i dati richiesti e li elabora in un formato adatto
// per essere salvati in un file TSV effettuando la join con i dati
// dei comuni italiani.
func events(n node) []event {
	b, err := topojson(n)
	if err != nil {
		return nil
	}
	rawmap := extract(b, n.date)
	cities := comuni.GetAll()
	out := make([]event, 0, len(rawmap))
	for _, c := range cities {
		// match con nomi corretti
		if ev, ok := rawmap[c.Name]; ok {
			ev.addInfo(c)
			out = append(out, ev)
			continue
		}
		// match con nomi accenti sbagliati
		key := comuni.SetWrongUTF8(c.Name, utf8Laceholder)
		if ev, ok := rawmap[key]; ok {
			ev.addInfo(c)
			out = append(out, ev)
		}
	}
	return out
}

// extract estrae gli eventi dal json
func extract(b []byte, d time.Time) map[string]event {
	type entry struct {
		NomeZona      string   `json:"Nome zona"`
		Idrogeologico string   `json:"Per rischio idrogeologico"`
		Idraulico     string   `json:"Per rischio idraulico"`
		Temporali     string   `json:"Per rischio temporali"`
		Comuni        []string `json:"comuni"`
	}
	var jsonStruct struct {
		Objects map[string]struct {
			Geometries []struct {
				Properties entry `json:"properties"`
			} `json:"geometries"`
		} `json:"objects"`
	}

	var entries []entry
	_ = json.Unmarshal(b, &jsonStruct)
	// accesso a una map con una sola entry, ma di cui non conosco la key
	// (è una data variabile). quindi niente paura, non è O(n²)
	for _, v := range jsonStruct.Objects {
		for _, h := range v.Geometries {
			entries = append(entries, h.Properties)
		}
	}
	slices.SortFunc(entries, func(a, b entry) int {
		return strings.Compare(a.NomeZona, b.NomeZona)
	})
	// la struttura map è l'unica che garantisce prestazioni decenti
	// nella join con le informazioni comunali, garantite dall'accesso
	// O(1) degli elementi.
	events := make(map[string]event, 10_000)
	// per ogni zona raccogli tutti i comuni. qui è obbligatorio un doppio loop
	// ma sono circa 7906 iterazioni costanti, ovvero il numero dei comuni italiani
	// qui suddivisi in sottoinsiemi che voglio aggregare in un unico contenitore.
	for _, entry := range entries {
		for _, c := range entry.Comuni {
			// i nomi dei comuni non hanno codifica corretta vedi issue
			// https://github.com/pcm-dpc/DPC-Bollettini-Criticita-Idrogeologica-Idraulica/issues/10
			events[c] = event{
				name:          c,
				zona:          entry.NomeZona,
				Temporali:     entry.Temporali,
				Idraulico:     entry.Idraulico,
				Idrogeologico: entry.Idrogeologico,
				data:          d,
			}
		}
	}
	return events
}
