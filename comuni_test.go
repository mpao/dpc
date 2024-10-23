package main

import (
	"encoding/csv"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetPrecipitazioni(t *testing.T) {
	r, err := http.Get("https://raw.githubusercontent.com/pcm-dpc/DPC-Bollettini-Vigilanza-Meteorologica/master/files/topojson/20241019_oggi.json")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Body.Close()
	b, _ := io.ReadAll(r.Body)
	err = os.WriteFile("testdata/topo.json", b, 0666)
	assert.NoError(t, err)
}

func Test_GetInfoComuni(t *testing.T) {
	r, err := http.Get("https://github.com/opendatasicilia/comuni-italiani/raw/main/dati/main.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Body.Close()
	b, _ := io.ReadAll(r.Body)
	err = os.WriteFile("testdata/comuni.csv", b, 0666)
	assert.NoError(t, err)
}

func Test_Extract(t *testing.T) {
	b, _ := os.ReadFile("testdata/topo.json")
	list := extract(b)
	assert.Equal(t, 8048, len(list))
}

func Test_InfoComuni(t *testing.T) {
	b, _ := os.ReadFile("testdata/comuni.csv")
	list := infoComuni(b)
	assert.Equal(t, 7904, len(list))
}

func Test_Popolazione(t *testing.T) {
	b, _ := os.ReadFile("testdata/popolazione_2021.csv")
	list := popolazione(b)
	assert.Equal(t, 7904, len(list))
}

func Test_AddPopulation(t *testing.T) {
	b, _ := os.ReadFile("testdata/comuni.csv")
	comuni := infoComuni(b)
	b, _ = os.ReadFile("testdata/popolazione_2021.csv")
	pop := popolazione(b)
	for i, c := range comuni {
		c.addPopulation(pop)
		comuni[i] = c
	}
	for _, v := range comuni {
		if v.pop == 0 {
			t.Errorf("%s ha popolazione pari a zero", v.name)
			t.Fail()
		}
	}
}

func Test_OperationComuni(t *testing.T) {
	comuni := infoComuni(comuniData)
	pop := popolazione(popolazioneData)
	for i, c := range comuni {
		c.addPopulation(pop)
		comuni[i] = c
	}
	b, _ := os.ReadFile("testdata/topo.json")
	events := extract(b)

	for i, c := range comuni {
		c.addEvent(events)
		comuni[i] = c
	}
	file, _ := os.Create("testdata/out.csv")
	w := csv.NewWriter(file)
	w.UseCRLF = true
	w.Comma = separator
	_ = w.Write(headerComuni)
	for _, ev := range comuni {
		_ = w.Write(ev.CSV())
	}
	w.Flush()
	var notfound int
	for _, c := range comuni {
		if c.evento == "" {
			notfound++
		}
	}
	assert.Equal(t, 7904, len(comuni))
	assert.Equal(t, 111, notfound)
	t.Log("file creato correttamente in testdata")
}
