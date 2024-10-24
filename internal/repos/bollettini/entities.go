package bollettini

import (
	"encoding/xml"
	"time"
)

// event rappresenta un evento di criticità idrogeologica o idraulica
// all'interno della applicazione. Viene costruito dopo il parsing XML
// ed è utilizzato per la visualizzazione dei dati.
type event struct {
	Date         time.Time
	Event        string
	Area         string
	Code         string
	OnSet        time.Time
	Expires      time.Time
	Category     string
	ResponseType string
	Urgency      string
	Severity     string
	Certainty    string
	SenderName   string
}

func (e event) CSV() []string {
	return []string{
		e.Date.Format("2006-01-02"),
		e.Event,
		e.Area,
		e.Code,
		e.OnSet.Format("2006-01-02 15:04:05"),
		e.Expires.Format("2006-01-02 15:04:05"),
		e.Category,
		e.ResponseType,
		e.Urgency,
		e.Severity,
		e.Certainty,
		e.SenderName,
	}
}

// result rappresenta il risultato del parsing XML del file
type result struct {
	Date   time.Time `xml:"sent"`
	Note   string    `xml:"note"`
	Alerts []alert   `xml:"info"`
}

// events restituisce una lista di eventi a partire da una lista di alert
// in modo da poter gestire più aree per lo stesso evento.
func (r *result) events() []event {
	var events []event
	for _, alert := range r.Alerts {
		ev := event{Date: r.Date}
		for _, area := range alert.Areas {
			ev.Area = area.Name
			ev.Code = area.Code
			ev.Event = alert.Event
			ev.OnSet = alert.OnSet
			ev.Expires = alert.Expires
			ev.Category = alert.Category
			ev.ResponseType = alert.ResponseType
			ev.Urgency = alert.Urgency
			ev.Severity = alert.Severity
			ev.Certainty = alert.Certainty
			ev.SenderName = alert.SenderName
			events = append(events, ev)
		}
	}
	return events
}

// alert rappresenta un alert all'interno del file XML
type alert struct {
	Category     string    `xml:"category"`
	Event        string    `xml:"event"`
	ResponseType string    `xml:"responseType"`
	Urgency      string    `xml:"urgency"`
	Severity     string    `xml:"severity"`
	Certainty    string    `xml:"certainty"`
	OnSet        time.Time `xml:"onset"`
	Expires      time.Time `xml:"expires"`
	SenderName   string    `xml:"senderName"`
	Areas        []area    `xml:"area"`
}

// area rappresenta un'area geografica all'interno di un alert
type area struct {
	Name string
	Code string
}

// UnmarshalXML implementa il metodo Unmarshaler per la struttura area
// in modo da poter gestire il tag Code come stringa
func (a *area) UnmarshalXML(e *xml.Decoder, start xml.StartElement) error {
	var temp struct {
		Name    string `xml:"areaDesc"`
		Geocode struct {
			Value string `xml:"value"`
		} `xml:"geocode"`
	}
	if err := e.DecodeElement(&temp, &start); err != nil {
		return err
	}
	a.Name = temp.Name
	a.Code = temp.Geocode.Value
	return nil
}
