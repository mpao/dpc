package app

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseDates(t *testing.T) {
	msg := "specificare il giorno nel formato ddmmyyyy oppure un intervallo ddmmyyyy-ddmmyyyy"
	cases := []struct {
		name string
		want string
	}{
		{name: ""},
		{name: "20101010-20101111", want: "specificare una data successiva al 01012020"},
		{name: "01012024", want: ""},
		{name: "01012024-32122024", want: msg},
		{name: "20241201-20241231", want: msg},
		{name: "01012024-01012024-foo", want: msg},
		{name: "01012024-31122024"},
		{name: "31122024-01012024", want: "specificare prima la data inferiore"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			var got string
			_, _, err := ParseDayParam(test.name, "01012020")
			if err != nil {
				got = err.Error()
			}
			assert.Equal(t, test.want, got)
		})
	}
}

func TestDayEqual(t *testing.T) {
	cases := []struct {
		name  string
		date1 string
		date2 string
		want  bool
	}{
		{date1: "20240101_0100", date2: "20240101_0100", want: true},
		{date1: "20240101_0100", date2: "20240101_0500", want: true},
		{date1: "20240101_0100", date2: "20240102_0100", want: false},
		{date1: "20230101_0100", date2: "20240101_0100", want: false},
		{date1: "20240201_0100", date2: "20240101_0100", want: false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			d1, _ := time.Parse("20060102_1504", test.date1)
			d2, _ := time.Parse("20060102_1504", test.date2)
			got := DayEqual(d1, d2)
			assert.Equal(t, test.want, got)
		})
	}
}
