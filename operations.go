package main

import (
	"archive/zip"
	"bytes"
	"crypto/tls"
	"encoding/csv"
	"encoding/xml"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/robfig/cron"
)

// target descrive le proprietà del file da scaricare.
// È necessario suddividerle poiché per le informazioni
// meteo non esiste un "latest" e ho bisogno di un fallback
// se le informazioni odierne non sono ancora presenti sul
// repository.
type target struct {
	name string
	repo string
	url  string
}

func httpClient() *http.Client {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	if proxy != "" {
		u, _ := url.Parse(proxy)
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy:           http.ProxyURL(u),
		}
	}
	return client
}

// download scarica i dati da DPC
func download(t target) ([]byte, error) {
	var url string
	today := time.Now()
	switch {
	case local != "":
		url = local
	case t.repo == repoMeteo:
		url = t.url + today.Format("20060102") + ".zip"
	default:
		url = t.url
	}
	resp, err := httpClient().Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		url = t.url + today.AddDate(0, 0, -1).Format("20060102") + ".zip"
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

// unzip scompatta il pacchetto ed estra l'unico file di interesse
func unzip(in []byte) ([]byte, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(in), int64(len(in)))
	if err != nil {
		return nil, err
	}
	var out []byte
	for _, zipFile := range zipReader.File {
		if !strings.HasPrefix(zipFile.Name, "Cap") {
			continue
		}
		f, err := zipFile.Open()
		if err != nil {
			return nil, err
		}
		defer f.Close()
		out, _ = io.ReadAll(f)
		break
	}
	return out, nil
}

// parse esegue parsing XML
func parse(in []byte) ([]event, error) {
	var rr result
	in = fixUTC(in)
	if err := xml.Unmarshal(in, &rr); err != nil {
		return nil, err
	}
	return rr.events(), nil
}

// fixUTC corregge il formato UTC, espresso in alcuni casi come +<numero> ed
// in altri come +<numero>:00. Questo metodo si occupa di uniformare il formato
// per utilizzare sempre +<numero>:00 nel MarhsalXML
func fixUTC(in []byte) []byte {
	re := regexp.MustCompile(`\+\d+</`)
	if re.Match(in) {
		return re.ReplaceAllFunc(in, func(match []byte) []byte {
			// Inserisce ":00" tra +<numero> e </
			return bytes.Replace(match, []byte("</"), []byte(":00</"), 1)
		})
	}
	return in
}

// jobManager esegue job() secondo i parametri passati
// a linea di comando dall'utente
func jobManager(t target) error {
	if service {
		f := func() {
			if err := job(t); err != nil {
				slog.Error(err.Error())
			}
		}
		j := cron.New()
		if err := j.AddFunc(round, f); err != nil {
			return err
		}
		j.Run() // è bloccante e lo voglio così!
	}
	if err := job(t); err != nil {
		slog.Error(err.Error())
	}
	return nil
}

// job esegue le operazioni di recupero dei dati e
// salva l'esito nel file generato con path e suffix
func job(t target) error {
	suffix := time.Now().Format("200601021504")
	if local != "" {
		ss := strings.Split(local, "/")[len(strings.Split(local, "/"))-1]
		ss = strings.Split(ss, ".")[0]
		suffix = ss
	}
	filename := filepath.Join(dest, t.name) + "-" + suffix + ".txt"
	// prima di eseguire le richieste HTTP, assicurati di poter scrivere su disco.
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	// ottieni i dati
	var events []event
	if local != "" {
		events, err = fromLocal(t)
	} else {
		events, err = fromNetwork(t)
	}
	if err != nil {
		return err
	}
	// scrivi i dati
	w := csv.NewWriter(file)
	w.UseCRLF = true
	w.Comma = separator
	_ = w.Write(header)
	for _, ev := range events {
		_ = w.Write(ev.CSV())
	}
	w.Flush()
	return nil
}

func fromNetwork(t target) ([]event, error) {
	in, err := download(t)
	if err != nil {
		return nil, err
	}
	in, err = unzip(in)
	if err != nil {
		return nil, err
	}
	return parse(in)
}

func fromLocal(t target) ([]event, error) {
	in, err := os.ReadFile(local)
	if err != nil {
		t.url = local
		return fromNetwork(t)
	}
	in, err = unzip(in)
	if err != nil {
		return nil, err
	}
	return parse(in)
}
