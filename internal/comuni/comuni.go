package comuni

import (
	_ "embed"
	"fmt"
	"strconv"
	"strings"
)

//go:embed assets/comuni.csv
var comuniData []byte

//go:embed assets/popolazione_2021.csv
var popolazioneData []byte

// Comune descrive un comune italiano attraverso i dati ufficiali
// del censimento 2021
type Comune struct {
	ID    string
	Name  string
	Prov  string
	Sigla string
	Reg   string
	Info  string
	Lat   float64
	Lon   float64
	Pop   int
}

// SetWrongUTF8 estra un valore, considerato come chiave, che possa essere usato per
// aggregare i dati dei comuni italiani con i dati della protezione civile. Tale
// valore per ora può essere solo il nome del comune, facendo attenzione che i dati
// provenienti da DPC hanno parecchie criticità, primo su tutti la codifica non UTF8
func SetWrongUTF8(s, replacewith string) string {
	// l'ugualianza tra i nomi dei comuni è problematica; i dati arrivano da due fonti diverse,
	// alcuni nomi sono completamente differenti (~200) e la fonte DPC ha problemi coi caratteri UTF8.
	// Uso quindi una stringa ricavata dal vero nome del comune per fare la ricerca nella map proveniente
	// da DPC
	for _, char := range []string{
		"à", "á", "ä", "â",
		"è", "é", "ë", "ê",
		"ì", "í", "ï", "î",
		"ò", "ó", "ö", "ô",
		"ù", "ú", "ü", "û",
	} {
		s = strings.ReplaceAll(s, char, replacewith)
	}
	return s
}

// GetAll restituisce la lista di tutti i comuni italiani e i loro attributi
func GetAll() []Comune {
	comuni := comuni(comuniData)
	pop := popolazione(popolazioneData)
	for i, c := range comuni {
		c.Pop = pop[c.ID]
		comuni[i] = c
	}
	return comuni
}

// comuni ricava la lista dei comuni italiani da un flusso di dati
func comuni(b []byte) (out []Comune) {
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
		c := Comune{
			ID:    attrs[1],
			Name:  attrs[0],
			Prov:  attrs[4],
			Sigla: attrs[5],
			Reg:   attrs[6],
			Info:  attrs[12],
			Lat:   lt,
			Lon:   lg,
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
