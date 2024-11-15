package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDates(t *testing.T) {
	msg := "specificare il giorno nel formato ddmmyyyy oppure un intervallo ddmmyyyy-ddmmyyyy"
	cases := []struct {
		name string
		want string
	}{
		{name: ""},
		{name: "20101010-20101111", want: "specificare una data successiva al 01/01/2020"},
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
			_, _, err := ParseDayParam(test.name)
			if err != nil {
				got = err.Error()
			}
			assert.Equal(t, test.want, got)
		})
	}
}
