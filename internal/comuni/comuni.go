package comuni

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"strings"
)

//go:embed assets/comuni.json
var comuniData []byte

// Comune descrive un comune italiano attraverso i dati ufficiali
// del censimento 2021
type Comune struct { //nolint //fieldalignment: struct with 136 pointer bytes could be 128
	Name  string   `json:"Comune"`
	Alias []string `json:"Alias,omitempty"`
	ID    string   `json:"ID"`
	Zone  []string `json:"Zone"`
	Prov  string
	Sigla string
	Reg   string
	Info  string
	Lat   float64
	Lon   float64
	Pop   int
}

// Keys genera una lista di chiavi che possono ricondurre univocamente allo stesso comune
func (c *Comune) Keys() []string {
	var yield []string
	for _, z := range c.Zone {
		yield = append(yield, Key(c.Name, z))
	}
	for _, a := range c.Alias {
		for _, z := range c.Zone {
			yield = append(yield, Key(a, z))
		}
	}
	return yield
}

// Get ricava le informazioni sul comune
func Get(id string) (Comune, bool) {
	var m = make(map[string]Comune)
	err := json.Unmarshal(comuniData, &m)
	if err != nil {
		return Comune{}, false
	}
	if val, ok := m[id]; ok {
		return val, ok
	}
	return Comune{}, false
}

// Amount restituisce il numero di comuni italiani definiti nella app
func Amount() int {
	var m = make(map[string]Comune)
	err := json.Unmarshal(comuniData, &m)
	if err != nil {
		return 0
	}
	return len(m)
}

// GetAll restituisce la lista di tutti i comuni italiani e i loro attributi
func GetAll() []Comune {
	var m = make(map[string]Comune)
	err := json.Unmarshal(comuniData, &m)
	if err != nil {
		return nil
	}
	var comuni = make([]Comune, 0, len(m))
	for _, c := range m {
		comuni = append(comuni, c)
	}
	return comuni
}

// FindEvent trova eventi atmosferici correlati al comune
func FindEvent[T any](c Comune, m map[string]T) (T, bool) {
	keys := c.Keys()
	for _, v := range keys {
		if val, ok := m[v]; ok {
			return val, ok
		}
	}
	return *new(T), false
}

func replaceChars(s string) string {
	var sb strings.Builder
	for _, c := range s {
		if 122 >= int(c) && int(c) >= 65 {
			sb.WriteRune(c)
		}
	}
	return sb.String()
}

// Key genera chiave identificativa
func Key(parts ...string) string {
	h := sha256.New()
	sep := []byte{0} // separatore tra le parti
	// in allerte i nomi dei comuni hanno bilinguismo
	// per comuni alpini: ignorali !
	bilingue := strings.Split(parts[0], "/")
	parts[0] = bilingue[0]
	for _, p := range parts {
		p = replaceChars(p)
		p = strings.ToLower(p) // previene case sensitive
		h.Write([]byte(p))
		h.Write(sep)
	}
	return hex.EncodeToString(h.Sum(nil))
}
