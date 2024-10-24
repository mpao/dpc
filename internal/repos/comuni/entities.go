package comuni

import (
	_ "embed"
	"strconv"
	"strings"
	"time"
)

//go:embed static/comuni.csv
var comuniData []byte

//go:embed static/popolazione_2021.csv
var popolazioneData []byte

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

func (c comune) CSV() []string {
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
