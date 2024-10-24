package ops

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Download(t *testing.T) {
	tt := Target{
		Name: "test",
		URL:  "https://httpbin.org",
		Filename: func() string {
			return "/status/404"
		},
		Fallback: func() string {
			return "/anything"
		},
	}
	f, err := Download(tt)
	assert.Nil(t, err)
	assert.NotEmpty(t, f)
}

type foo string

func (f foo) CSV() []string {
	return []string{"foo"}
}

func Test_Save(t *testing.T) {
	cases := []struct {
		name string
		h    []string
		d    []foo
		want bool
	}{
		{name: "/KO", want: true},
		{name: "OK", want: false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			suffix := time.Now().Format("200601021504")
			got := Save(test.name, test.h, test.d)
			assert.Equal(t, test.want, got != nil)
			_ = os.Remove(test.name + "-" + suffix + ".txt")
		})
	}
}
