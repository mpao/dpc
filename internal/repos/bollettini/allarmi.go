package bollettini

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"os"
	"regexp"
	"strings"

	"dpc/internal/ops"
)

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

// JobAllarmi esegue le operazioni di recupero dei dati
func JobAllarmi(t ops.Target) (err error) {
	var events []event
	if ops.Local != "" {
		events, err = fromLocal(t)
	} else {
		events, err = fromNetwork(t)
	}
	if err != nil {
		return err
	}
	var header = []string{
		"Date",
		"Event",
		"Area",
		"Code",
		"OnSet",
		"Expires",
		"Category",
		"ResponseType",
		"Urgency",
		"Severity",
		"Certainty",
		"SenderName",
	}
	return ops.Save(t.Name, header, events)
}

func fromNetwork(t ops.Target) ([]event, error) {
	in, err := ops.Download(t)
	if err != nil {
		return nil, err
	}
	in, err = unzip(in)
	if err != nil {
		return nil, err
	}
	return parse(in)
}

func fromLocal(t ops.Target) ([]event, error) {
	in, err := os.ReadFile(ops.Local)
	if err != nil {
		t.URL = ops.Local
		return fromNetwork(t)
	}
	in, err = unzip(in)
	if err != nil {
		return nil, err
	}
	return parse(in)
}
