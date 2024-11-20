package app

import (
	"encoding/csv"
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	// Dest config vedi -d
	Dest string
	// Proxy config vedi -p
	Proxy string
	// Original config vedi -o
	Original bool
	// Join config vedi -j
	Join bool
	// Interval config vedi -i
	Interval string
)

// ParseDayParam controlla e valida le date di input
func ParseDayParam(s, limit string) (from, to time.Time, err error) {
	msg := "specificare il giorno nel formato ddmmyyyy oppure un intervallo ddmmyyyy-ddmmyyyy"
	arr := strings.Split(s, "-")
	dateLimit, err := time.Parse("02012006", limit)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if s == "" {
		return time.Time{}, time.Time{}, nil
	}
	if len(arr) > 2 {
		return time.Time{}, time.Time{}, errors.New(msg)
	}
	if from, err = time.Parse("02012006", arr[0]); err != nil {
		return time.Time{}, time.Time{}, errors.New(msg)
	}
	if from.Before(dateLimit) {
		return time.Time{}, time.Time{}, errors.New("specificare una data successiva al " + limit)
	}
	if len(arr) == 1 {
		return from, from, nil
	}
	if to, err = time.Parse("02012006", arr[1]); err != nil {
		return time.Time{}, time.Time{}, errors.New(msg)
	}
	if from.After(to) {
		return time.Time{}, time.Time{}, errors.New("specificare prima la data inferiore")
	}
	return from, to, nil
}

// SaveCSV salva il risultato come file CSV
func SaveCSV(filename string, payload [][]string) error {
	filename = filepath.Join(Dest, filename) + ".txt"
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	w := csv.NewWriter(file)
	w.UseCRLF = true
	w.Comma = '\t'
	_ = w.WriteAll(payload)
	return nil
}

// SaveBytes salva il risultato esattamente come scaricato dal repository
func SaveBytes(filename string, payload []byte) error {
	filename = filepath.Join(Dest, filename) + ".json"
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	_, _ = file.Write(payload)
	return nil
}

// HTTPClient restituisce il client HTTP per l'applicazione
func HTTPClient() *http.Client {
	client := &http.Client{
		Timeout: 20 * time.Second,
	}
	if Proxy != "" {
		u, _ := url.Parse(Proxy)
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(u),
		}
	}
	return client
}

// DayEqual stabilisce l'ugualianza tra due date senza tener conto dell'orario
func DayEqual(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
