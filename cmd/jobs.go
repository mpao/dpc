package main

import (
	"log/slog"

	"dpc/internal/ops"
	b "dpc/internal/repos/bollettini"
	c "dpc/internal/repos/comuni"

	"github.com/robfig/cron"
)

var jobs = map[string]func(ops.Target) error{
	"allerte": b.JobAllarmi,
	"meteo":   b.JobAllarmi,
	"comuni":  c.JobComuni,
}

// jobManager esegue job() secondo i parametri passati
// a linea di comando dall'utente
func jobManager(t ops.Target) error {
	var job = jobs[t.Name]
	if ops.Service {
		f := func() {
			if err := job(t); err != nil {
				slog.Error(err.Error())
			}
		}
		j := cron.New()
		if err := j.AddFunc(ops.Round, f); err != nil {
			return err
		}
		j.Run() // è bloccante e lo voglio così!
	}
	if err := job(t); err != nil {
		slog.Error(err.Error())
	}
	return nil
}
