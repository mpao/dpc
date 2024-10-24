package comuni

import (
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"
	"time"

	"dpc/internal/ops"
)

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
		// TODO cancellare appena possibile questo IF
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

// JobComuni esegue le operazioni di recupero dei dati
func JobComuni(t ops.Target) error {
	comuni := infoComuni(comuniData)
	pop := popolazione(popolazioneData)
	b, err := ops.Download(t)
	if err != nil {
		return err
	}
	events := extract(b)
	for i, c := range comuni {
		c.addPopulation(pop)
		c.addEvent(events)
		comuni[i] = c
	}
	var headers = []string{
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
	return ops.Save(t.Name, headers, comuni)
}
