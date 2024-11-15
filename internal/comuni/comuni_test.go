package comuni

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_InfoComuni(t *testing.T) {
	b, _ := os.ReadFile("assets/comuni.csv")
	list := comuni(b)
	assert.Equal(t, 7904, len(list))
}

func Test_Popolazione(t *testing.T) {
	b, _ := os.ReadFile("assets/popolazione_2021.csv")
	list := popolazione(b)
	assert.Equal(t, 7904, len(list))
}

func TestComuni(t *testing.T) {
	c := GetAll()
	assert.Equal(t, 7904, len(c))
	for _, v := range c {
		switch v.Name {
		case "Firenze":
			assert.Equal(t, 359_755, v.Pop)
		case "Milano":
			assert.Equal(t, 1_397_715, v.Pop)
		case "Forlì":
			assert.Equal(t, 117_479, v.Pop)
		case "Sassofeltrio": // in attesa dell'issue #11
			assert.Equal(t, 1_349, v.Pop)
		}
	}
}
