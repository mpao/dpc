package main

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/xml"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/robfig/cron"
)

var url = repo + path

func download() ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

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
func parse(in []byte) ([]event, error) {
	var rr result
	if err := xml.Unmarshal(in, &rr); err != nil {
		return nil, err
	}
	return rr.events(), nil
}

// jobManager esegue job() secondo i parametri passati
// a linea di comando dall'utente
func jobManager() error {
	if service {
		f := func() {
			if err := job(); err != nil {
				slog.Error(err.Error())
			}
		}
		j := cron.New()
		if err := j.AddFunc(round, f); err != nil {
			return err
		}
		j.Run() // è bloccante e lo voglio così!
	}
	if err := job(); err != nil {
		slog.Error(err.Error())
	}
	return nil
}

// job esegue le operazioni di recupero dei dati e
// salva l'esito nel file generato con path e suffix
func job() error {
	// prima di eseguire le richieste HTTP, assicurati di poter scrivere su disco.
	suffix := time.Now().Format("200601021504")
	filename := filepath.Join(dest, fileprefix) + "-" + suffix + ".txt"
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	// ottieni i dati
	var events []event
	if local != "" {
		events, err = fromLocal(local)
	} else {
		events, err = fromNetwork()
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

func fromNetwork() ([]event, error) {
	in, err := download()
	if err != nil {
		return nil, err
	}
	in, err = unzip(in)
	if err != nil {
		return nil, err
	}
	return parse(in)
}

func fromLocal(filename string) ([]event, error) {
	in, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	in, err = unzip(in)
	if err != nil {
		return nil, err
	}
	return parse(in)
}
