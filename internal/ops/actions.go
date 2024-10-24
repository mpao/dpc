package ops

import (
	"encoding/csv"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const Separator = '\t' // Separator per il CSV

var (
	Service bool   // Service config vedi -s
	Dest    string // Dest config vedi -d
	Local   string // Local config vedi -f
	Round   string // Round config vedi -r
	Proxy   string // Proxy config vedi -p
)

// Target descrive le proprietà del file da scaricare.
// È necessario suddividerle poiché per le informazioni
// meteo non esiste un "latest" e ho bisogno di un fallback
// se le informazioni odierne non sono ancora presenti sul
// repository.
type Target struct {
	Filename func() string
	Fallback func() string
	Name     string
	URL      string
}

// CSVable interfaccia che descrive un dato che può essere salvato in CSV
type CSVable interface {
	CSV() []string
}

func httpClient() *http.Client {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	if Proxy != "" {
		u, _ := url.Parse(Proxy)
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(u),
		}
	}
	return client
}

// Download scarica i dati per il Target
func Download(t Target) ([]byte, error) {
	var url = t.URL + t.Filename()
	if Local != "" {
		url = Local
	}
	resp, err := httpClient().Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		url = t.URL + t.Fallback()
		resp, err = httpClient().Get(url) //nolint //bodyclose line 33
		if err != nil {
			return nil, err
		}
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// Save salva i dati ricavati in un file CSV
func Save[T CSVable](name string, headers []string, data []T) error {
	suffix := time.Now().Format("200601021504")
	if Local != "" {
		ss := strings.Split(Local, "/")[len(strings.Split(Local, "/"))-1]
		ss = strings.Split(ss, ".")[0]
		suffix = ss
	}
	filename := filepath.Join(Dest, name) + "-" + suffix + ".txt"
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	w := csv.NewWriter(file)
	w.UseCRLF = true
	w.Comma = Separator
	_ = w.Write(headers)
	for _, ev := range data {
		_ = w.Write(ev.CSV())
	}
	w.Flush()
	return nil
}
