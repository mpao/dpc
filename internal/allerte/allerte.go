package allerte

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"dpc/internal/app"
	"dpc/internal/comuni"

	"github.com/spf13/cobra"
)

const (
	domain  = "https://api.github.com/repos/pcm-dpc/DPC-Bollettini-Criticita-Idrogeologica-Idraulica/"
	fileURL = "https://github.com/pcm-dpc/DPC-Bollettini-Criticita-Idrogeologica-Idraulica/raw/master/"
)

// tree è la struttura del json restituito dalla API di github che descrive
// i files presenti nel repository
type tree struct {
	SHA       string `json:"sha"`
	URL       string `json:"url"`
	Tree      []node `json:"tree"`
	Truncated bool   `json:"truncated"`
}

// node descrive le proprietà dei files nel tree che mi interessano
type node struct {
	date     time.Time
	Filename string `json:"path"`
	url      string // non usare la URL della API! vedi topojsonList()
}

// addDate aggiunge a node la data di pubblicazione
func (n *node) addDate() {
	rx := regexp.MustCompile(`files/topojson/(\d{8}_\d{4})_today\.json`)
	matches := rx.FindStringSubmatch(n.Filename)
	if len(matches) > 1 {
		date, _ := time.Parse("20060102_1504", matches[1])
		n.date = date
	}
}

// topojsonList scarica la lista di files topojson dal repository github
func topojsonList() ([]node, error) {
	url := domain + "git/trees/master?recursive=1"
	r, err := http.Get(url)
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
		if i != 0 && dayEqual(out[i-shift].date, n.date) {
			out[i-shift] = n
			shift++
			continue
		}
		out = append(out, n)
	}
	return slices.Clip(out)
}

// dayEqual stabilisce l'ugualianza tra due date senza tener conto dell'orario
func dayEqual(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// topojson scarica il topojson
func topojson(n node) ([]byte, error) {
	r, err := http.Get(n.url)
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

type event struct {
	data          time.Time
	Idrogeologico string
	Idraulico     string
	Temporali     string
	name          string
	zona          string
	prov          string
	sigla         string
	reg           string
	info          string
	lat           float64
	lon           float64
	pop           int
}

func (e *event) addInfo(c comuni.Comune) {
	e.name = c.Name // questo fixa i noi sbagliati dalla codifica
	e.prov = c.Prov
	e.sigla = c.Sigla
	e.reg = c.Reg
	e.info = c.Info
	e.lat = c.Lat
	e.lon = c.Lon
	e.pop = c.Pop
}

func (e *event) CSV() []string {
	return []string{
		e.data.Format("2006-01-02"),
		e.Idrogeologico,
		e.Idraulico,
		e.Temporali,
		e.name,
		e.prov,
		e.sigla,
		e.zona,
		e.reg,
		strconv.Itoa(e.pop),
		strconv.FormatFloat(e.lat, 'f', -1, 64),
		strconv.FormatFloat(e.lon, 'f', -1, 64),
		e.info,
	}
}

// Get comando per il download delle allerte DPC
func Get(cmd *cobra.Command, args []string) error {
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
		key := comuni.SetWrongUTF8(c.Name)
		if ev, ok := rawmap[key]; ok {
			ev.addInfo(c)
			out = append(out, ev)
		}
	}
	return out
}

func filterNodes(interval string, nodes []node) ([]node, error) {
	var out []node
	from, to, err := app.ParseDayParam(interval)
	if err != nil {
		return nil, err
	}
	if interval == "" {
		return append(out, nodes[len(nodes)-1]), nil
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

func writeJSON(nodes []node) error {
	for _, node := range nodes {
		b, err := topojson(node)
		if err != nil {
			slog.Error("")
		}
		name := "allerte-topojson-" + node.date.Format("20060102")
		if err := app.SaveBytes(name, b); err != nil {
			return err
		}
	}
	return nil
}

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
	for _, node := range nodes {
		payload := make([][]string, 0, 8_000)
		collection := events(node)
		for _, v := range collection {
			payload = append(payload, v.CSV())
		}
		joined = append(joined, payload...)
		if app.Join {
			name := "allerte-" + nodes[0].date.Format("20060102") + nodes[len(nodes)-1].date.Format("20060102")
			if err := app.SaveCSV(name, joined); err != nil {
				return err
			}
		} else {
			name := "allerte-" + node.date.Format("20060102")
			payload = slices.Insert(payload, 0, headers)
			if err := app.SaveCSV(name, payload); err != nil {
				return err
			}
		}
	}
	return nil
}
