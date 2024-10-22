package main

import (
	"encoding/csv"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

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

func Test_Equal(t *testing.T) {
	cases := []struct {
		name   string
		comune string
		want   bool
	}{
		{name: "Citt Sant'Angelo", comune: "Città Sant'Angelo", want: true},
		{name: "Città Sant'Angelo", comune: "Città Sant'Angelo", want: true},
		{name: "foobar", comune: "Città Sant'Angelo", want: false},
		{name: "foobar", comune: "foobarbaz", want: false},
		{name: "foobarbaz", comune: "foobar", want: false},
		{name: "Ceva", comune: "Cave", want: false},
		{name: "Ctt Pi Bll ", comune: "Cïttà Più BèllÖ Ü", want: false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			c := comune{name: test.name}
			got := c.equal(test.comune)
			assert.Equal(t, test.want, got)
		})
	}
}

func Test_Extract(t *testing.T) {
	b, _ := os.ReadFile("testdata/topo.json")
	list := extract(b)
	assert.Equal(t, 10843, len(list))
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

func Test_Join(t *testing.T) {
	b, _ := os.ReadFile("testdata/comuni.csv")
	comuni := infoComuni(b)
	b, _ = os.ReadFile("testdata/popolazione_2021.csv")
	pop := popolazione(b)
	for i, c := range comuni {
		c.addPopulation(pop)
		comuni[i] = c
	}
	b, _ = os.ReadFile("testdata/topo.json")
	ee := extract(b)
	var foo []comune
	for _, c := range comuni {
		for _, e := range ee {
			if c.equal(e.name) {
				c.zone = e.zona
				ff := comune{
					data:   time.Now(),
					evento: e.evento,
				}
				foo = append(foo, ff)
				break
			}
		}
		if c.zone == "" {
			c.zone = "---"
			ff := comune{
				data:   time.Now(),
				evento: "ND",
			}
			foo = append(foo, ff)
		}
	}
	file, _ := os.Create("out.csv")
	w := csv.NewWriter(file)
	w.UseCRLF = true
	w.Comma = separator

	_ = w.Write(headerComuni)
	for _, ev := range foo {
		_ = w.Write(ev.CSV())
	}
	w.Flush()
	t.Log("done")
}
