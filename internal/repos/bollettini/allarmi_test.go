package bollettini

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_FixUTC(t *testing.T) {
	cases := []struct {
		name string
		in   []byte
		want string
	}{
		{name: "+01", in: []byte("+01</"), want: "+01:00</"},
		{name: "+02", in: []byte("+02</"), want: "+02:00</"},
		{name: "two", in: []byte("+01</foo+02</"), want: "+01:00</foo+02:00</"},
		{name: "+02:00", in: []byte("+02:00</"), want: "+02:00</"},
		{name: "invalid", in: []byte("+0A</"), want: "+0A</"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := fixUTC(c.in)
			assert.Contains(t, string(got), c.want)
		})
	}
}
