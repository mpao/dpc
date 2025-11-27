package allerte

import (
	"slices"
	"testing"
	"time"

	"github.com/mpao/dpc/internal/app"
	"github.com/stretchr/testify/assert"
)

// TestTopojsonList mancano geojson e topojson nel 2020, questo test lo
// dimostra ma riesce comunque a controllare che i dati siano sufficientemente presenti
func TestTopojsonList(t *testing.T) {
	n, err := topojsonList()
	if err != nil {
		t.Fatal(err)
	}
	// la slice è sufficientemente piena ?
	assert.True(t, len(n) > 1_000)
	var missing int
	for i, v := range n[:len(n)-1] {
		assert.False(t, v.date.IsZero())
		diff := int(n[i+1].date.Sub(v.date).Hours() / 24)
		if diff > 1 {
			missing++
			t.Log("mancante tra:", v.date.Format("02-01-2006"), n[i+1].date.Format("02-01-2006"), diff)
		}
	}
	// tra la prima e l'ultima data ci sono lo stesso numero di elementi
	// che la differenza di giorni tra le due date
	first := n[0].date
	last := n[len(n)-1].date
	diff := int(last.Sub(first).Hours()/24) - missing
	assert.Equal(t, diff, len(n)-1)
}

// TestGetTopojson lettura di un bollettino giornaliero
func TestTopojson(t *testing.T) {
	url := "https://github.com/pcm-dpc/DPC-Bollettini-Criticita-Idrogeologica-Idraulica/raw/master/"
	cases := []struct {
		name    string
		url     string
		content string
	}{
		{name: "OK", url: url + "files/topojson/20200101_1530_today.json", content: "Firenze"},
		{name: "KO", url: "", content: "unsupported protocol scheme"},
		{name: "Empty", url: "https://go.dev", content: "<!DOCTYPE html>"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			n := node{
				date: time.Now(),
				url:  test.url,
			}
			b, err := topojson(n)
			switch test.name {
			case "OK":
				assert.Contains(t, string(b), test.content)
			case "KO":
				assert.Contains(t, err.Error(), test.content)
			case "Empty":
				assert.Equal(t, "", string(b))
			}
		})
	}
}

func TestExtract(t *testing.T) {
	url := "https://github.com/pcm-dpc/DPC-Bollettini-Criticita-Idrogeologica-Idraulica/raw/master/"
	n := node{
		date: time.Now(),
		url:  url + "files/topojson/20200101_1530_today.json",
	}
	b, _ := topojson(n)
	// _ = os.WriteFile("topo.json", b, 0666)
	got := extract(b, n.date)
	assert.Equal(t, 7952, len(got))
}

func TestEvents(t *testing.T) {
	url := "https://github.com/pcm-dpc/DPC-Bollettini-Criticita-Idrogeologica-Idraulica/raw/master/"
	// il problema degli accenti è più grosso di quanto sembrasse:
	// nel 2020 - e quindi penso anche successivamente - la codifica era corretta
	// come UTF8, poi non si sa perché, la situazione cambia.
	cases := []struct {
		name    string
		date    time.Time
		file    string
		allarme string
	}{
		{
			name:    "01012020",
			date:    time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local),
			file:    "files/topojson/20200101_1530_today.json",
			allarme: "NESSUNA ALLERTA",
		},
		{
			name:    "13112024",
			date:    time.Date(2024, 11, 13, 0, 0, 0, 0, time.Local),
			file:    "files/topojson/20241113_1521_today.json",
			allarme: "ARANCIONE",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			n := node{
				date: test.date,
				url:  url + test.file,
			}
			out := events(n)
			var cornercases []event
			var emptyComuni int
			for _, v := range out {
				// controlla correzione lettere accentate, zona alluvionata il 13.11.2024
				// e un nome con due parole
				if slices.Contains([]string{"Forlì", "Giarre", "Nocera Umbra"}, v.name) {
					cornercases = append(cornercases, v)
				}
				// nel topojson ci sono molti comuni aggregati insieme, oppure
				// con il nome bilingue; mi perdo l'informazione di circa 140 comuni
				// Gli devo ignorare, l'aggregazione può essere fatta sulla zona di allerta
				if v.Idrogeologico == "" {
					emptyComuni++
				}
			}
			assert.Equal(t, 3, len(cornercases))
			assert.Equal(t, 0, emptyComuni)
			assert.Contains(t, cornercases[2].Idrogeologico, test.allarme)
		})
	}
}

func TestFilterNodes(t *testing.T) {
	cases := []struct {
		interval string
		size     int
	}{
		{interval: "", size: 2},
		{interval: "20200101", size: 0},
		{interval: "05012020", size: 1},
		{interval: "01012020-31012020", size: 31},
	}
	list, err := topojsonList()
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range cases {
		t.Run(test.interval, func(t *testing.T) {
			got, _ := filterNodes(test.interval, list)
			assert.Equal(t, test.size, len(got))
		})
	}
}

func TestDemo_FileCreation(t *testing.T) {
	cases := []struct {
		name     string
		original bool
		interval string
		join     bool
	}{
		{name: "single json", original: true, interval: "15082024"},
		{name: "some json", original: true, interval: "12112024-14112024"},
		{name: "single csv", interval: "15082024"},
		{name: "some csv", interval: "12112024-14112024"},
		{name: "joined csv", join: true, interval: "12112024-14112024"},
	}
	app.Dest = "../../bin"
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			app.Interval = test.interval
			app.Join = test.join
			app.Original = test.original
			if err := Get(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
