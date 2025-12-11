package comuni

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetAll(t *testing.T) {
	got := GetAll()
	assert.Equal(t, 7904, len(got))

	var alias []Comune
	var zone []Comune
	for _, c := range got {
		switch c.Name {
		case "Firenze":
			assert.Equal(t, 359_755, c.Pop)
		case "Milano":
			assert.Equal(t, 1_397_715, c.Pop)
		case "Forlì":
			assert.Equal(t, 117_479, c.Pop)
		}
		if len(c.Alias) > 0 {
			alias = append(alias, c)
		}
		if len(c.Zone) < 2 {
			zone = append(zone, c) // non devono avere 0 o 1 zona: minimo due (allerta, meteo)
		}
	}
	assert.Equal(t, 163, len(alias))
	assert.Equal(t, 1, len(zone))
}

func Test_ReplaceChars(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{name: "Forlì", in: "Forlì", want: "Forl"},
		{name: "Malè", in: "Mal�", want: "Mal"},
		{name: "San Niccolò", in: "San Niccol", want: "SanNiccol"},
		{name: "Rhêmes-Notre-Dame", in: "Rhemes-Notre-Dame", want: "RhemesNotreDame"},
		{name: "Valle d'Aosta", in: "Valle d\u2019Aosta", want: "ValledAosta"},
		{name: "Valle d'Aosta", in: "Valle d’Aosta", want: "ValledAosta"},
		{name: "ZazA", in: "ZazA", want: "ZazA"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got := replaceChars(test.in)
			assert.Equal(t, test.want, got)
		})
	}
}

func Test_Key(t *testing.T) {
	cases := []struct {
		name string
		p1   string
		p2   string
	}{
		{name: "#1", p1: "Rhmes-Notre Dame", p2: "Valle d\u2019Aosta"},
		{name: "#1", p1: "Rhêmes-Notre-Dame", p2: "Valle d'Aosta"},
		{name: "#1", p1: "Rhmes-Notre-Dame", p2: "Valle d’Aosta"},
	}
	var got []string
	for _, test := range cases {
		got = append(got, Key(test.p1, test.p2))
	}
	for i := 1; i < len(got); i++ {
		if got[i] != got[0] {
			t.Fatalf("gli elementi non sono tutti uguali: %v", got)
		}
	}
}

func Test_Keys(t *testing.T) {
	sample := Comune{
		Name:  "Rhêmes-Notre-Dame",
		Alias: []string{"Rhmes-Notre-Dame", "Rhemes-Notre-Dame"},
		Zone:  []string{"Valle d'Aosta", "Valle d\u2019Aosta", "Valle d’Aosta"},
	}
	keys := sample.Keys()
	assert.Equal(t, 9, len(keys))
}

func Test_FindEvent(t *testing.T) {
	// mock comune
	sample := Comune{
		Name:  "Rhêmes-Notre-Dame",
		Alias: []string{"Rhmes-Notre-Dame", "Rhemes-Notre-Dame"},
		Zone:  []string{"Valle d'Aosta", "Valle d\u2019Aosta", "Valle d’Aosta"},
	}
	// mock mappa events
	m := map[string]int{
		Key(sample.Name, sample.Zone[2]):     1,
		Key(sample.Alias[1], sample.Zone[1]): 1,
	}
	cases := []struct {
		name string
		in   Comune
		m    map[string]int
		want int
		ok   bool
	}{
		{name: "ok name", in: sample, m: m, want: 1, ok: true},
		{name: "ok alias", in: sample, m: m, want: 1, ok: true},
		{name: "not found", in: Comune{Name: "foo"}, m: m},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got, ok := FindEvent(test.in, test.m)
			assert.Equal(t, test.want, got)
			assert.Equal(t, test.ok, ok)
		})
	}
}
