package allerte

import (
	"regexp"
	"strconv"
	"time"

	"github.com/mpao/dpc/internal/comuni"
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
