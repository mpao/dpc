package main

import (
	_ "embed"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

//go:embed testdata/comuni.csv
var comuniData []byte

//go:embed testdata/popolazione_2021.csv
var popolazioneData []byte
var headerComuni = []string{
	"Data",
	"Evento",
	"Comune",
	"Provincia",
	"Sigla",
	"Zona",
	"Regione",
	"Popolazione",
	"Latitudine",
	"Longitudine",
	"Info",
}

type comune struct {
	data   time.Time
	id     string
	name   string
	prov   string
	sigla  string
	zone   string
	reg    string
	info   string
	evento string
	lat    float64
	lon    float64
	pop    int
}

func (c *comune) addPopulation(m map[string]int) {
	c.pop = m[c.id]
}

func (c *comune) addEvent(m map[string]evento) {
	// l'ugualianza tra i nomi dei comuni è problematica; i dati arrivano da due fonti diverse,
	// alcuni nomi sono completamente differenti (~200) e la fonte DPC ha problemi coi caratteri UTF8.
	// Uso quindi una stringa ricavata dal vero nome del comune per fare la ricerca nella map proveniente
	// da DPC
	s := strings.ToLower(c.name)
	for _, char := range []string{"à", "á", "ä", "è", "é", "ë", "ì", "í", "ï", "ò", "ó", "ö", "ù", "ú", "ü"} {
		s = strings.ReplaceAll(s, char, "")
	}
	c.zone = m[s].zona
	c.data = time.Now()
	c.evento = m[s].evento
	c.data = m[s].data
}

func (c *comune) CSV() []string {
	return []string{
		c.data.Format("2006-01-02"),
		c.evento,
		c.name,
		c.prov,
		c.sigla,
		c.zone,
		c.reg,
		strconv.Itoa(c.pop),
		strconv.FormatFloat(c.lat, 'f', -1, 64),
		strconv.FormatFloat(c.lon, 'f', -1, 64),
		c.info,
	}
}

type evento struct {
	data   time.Time
	name   string
	zona   string
	evento string
}

// extract estrae la lista dei nomi dei comuni con il relativo evento metereologico
func extract(b []byte) map[string]evento {
	type events struct {
		day                  time.Time
		NomeZona             string   `json:"Nome_Zona"`
		QuantitativiPrevisti string   `json:"Quantitativi_previsti"`
		Comuni               []string `json:"comuni"`
	}
	var jsonStruct struct {
		Objects map[string]struct {
			Geometries []struct {
				Properties events `json:"properties"`
			} `json:"geometries"`
		} `json:"objects"`
	}
	var data []events
	_ = json.Unmarshal(b, &jsonStruct)
	for k, v := range jsonStruct.Objects {
		for _, h := range v.Geometries {
			d, err := time.Parse("20060102150405", k)
			if err != nil {
				log.Fatal(err)
			}
			h.Properties.day = d
			data = append(data, h.Properties)
		}
	}
	slices.SortFunc(data, func(a, b events) int {
		return strings.Compare(a.NomeZona, b.NomeZona)
	})
	out := make(map[string]evento, 10_000)
	for _, ev := range data {
		for _, v := range ev.Comuni {
			out[strings.ToLower(v)] = evento{
				name:   v,
				zona:   ev.NomeZona,
				evento: ev.QuantitativiPrevisti,
				data:   ev.day,
			}
		}
	}
	return out
}

// infoComuni ricava la lista dei comuni italiani da un flusso di dati
func infoComuni(b []byte) (out []comune) {
	list := strings.Split(string(b), "\n")
	for i := 1; i < len(list); i++ {
		attrs := strings.Split(list[i], ",")
		if len(attrs) < 12 {
			continue
		}
		for i, v := range attrs {
			attrs[i] = strings.TrimSpace(v)
		}
		lt, err := strconv.ParseFloat(attrs[2], 64)
		if err != nil {
			lt = 0
		}
		lg, err := strconv.ParseFloat(attrs[3], 64)
		if err != nil {
			lt = 0
		}
		// https://github.com/opendatasicilia/comuni-italiani/issues/11#issuecomment-2426871148
		// Per un paese non è corretto l'ID ISTAT, va fatto a mano in attesa di fix a monte
		// TODO cancellare appena possibile
		if attrs[1] == "099031" {
			attrs[1] = "041060"
		}
		c := comune{
			id:    attrs[1],
			name:  attrs[0],
			prov:  attrs[4],
			sigla: attrs[5],
			reg:   attrs[6],
			info:  attrs[12],
			lat:   lt,
			lon:   lg,
		}
		out = append(out, c)
	}
	return
}

// popolazione ricava la popolazione dei comuni italiani da un flusso di dati
func popolazione(b []byte) map[string]int {
	list := strings.Split(string(b), "\n")
	var out = map[string]int{}
	for i := 1; i < len(list); i++ {
		attrs := strings.Split(list[i], ",")
		if len(attrs) < 2 {
			continue
		}
		for i, v := range attrs {
			attrs[i] = strings.TrimSpace(v)
		}
		value, err := strconv.Atoi(attrs[1])
		if err != nil {
			value = 0
		}
		s := fmt.Sprintf("%06s", attrs[0])
		out[s] = value
	}
	return out
}

func jobComuni(t target) error {
	suffix := time.Now().Format("200601021504")
	if local != "" {
		ss := strings.Split(local, "/")[len(strings.Split(local, "/"))-1]
		ss = strings.Split(ss, ".")[0]
		suffix = ss
	}
	filename := filepath.Join(dest, t.name) + "-" + suffix + ".txt"
	// prima di eseguire le richieste HTTP, assicurati di poter scrivere su disco.
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	comuni := infoComuni(comuniData)
	pop := popolazione(popolazioneData)
	for i, c := range comuni {
		c.addPopulation(pop)
		comuni[i] = c
	}
	b, err := download(t)
	if err != nil {
		return err
	}
	events := extract(b)
	for i, c := range comuni {
		c.addEvent(events)
		comuni[i] = c
	}
	// scrivi i dati
	w := csv.NewWriter(file)
	w.UseCRLF = true
	w.Comma = separator
	_ = w.Write(headerComuni)
	for _, ev := range comuni {
		_ = w.Write(ev.CSV())
	}
	w.Flush()
	return nil
}
