package main

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

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

// equal l'ugualianza tra i nomi dei comuni è problematica; i dati arrivano da due fonti diverse,
// alcuni nomi sono completamente differenti (~200) e la fonte DPC ha problemi coi caratteri UTF8
func (c *comune) equal(s string) bool {
	s1 := strings.ToLower(c.name)
	s2 := strings.ToLower(s)
	for _, c := range []string{"à", "á", "ä", "è", "é", "ë", "ò", "ó", "ö", "ù", "ú", "ü"} {
		s1 = strings.ReplaceAll(s1, c, "")
		s2 = strings.ReplaceAll(s2, c, "")
	}
	return strings.EqualFold(s1, s2)
}

func (c *comune) addPopulation(m map[string]int) {
	c.pop = m[c.id]
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
	name   string
	zona   string
	evento string
}

// extract estrae la lista dei nomi dei comuni con il relativo evento metereologico
func extract(b []byte) (out []evento) {
	type events struct {
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
	for _, v := range jsonStruct.Objects {
		for _, h := range v.Geometries {
			data = append(data, h.Properties)
		}
	}
	slices.SortFunc(data, func(a, b events) int {
		return strings.Compare(a.NomeZona, b.NomeZona)
	})
	for _, hh := range data {
		for _, v := range hh.Comuni {
			out = append(out, evento{
				name:   v,
				zona:   hh.NomeZona,
				evento: hh.QuantitativiPrevisti,
			})
		}
	}
	slices.SortFunc(out, func(a, b evento) int {
		return strings.Compare(a.name, b.name)
	})
	return
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

func join() {}
