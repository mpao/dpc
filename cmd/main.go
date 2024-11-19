package main

import (
	"fmt"
	"os"
	"time"
)

var (
	appName    = "dpc"  // Name valore di fallback, usa il Taskfile per la definizione
	appVersion = "v0.0" // Version valore di fallback, usa il Taskfile per la definizione
)

func main() {
	prefix := "avvio operazioni: "
	s := make(chan string)
	go func() {
		go spinner(prefix)
		if err := root.Execute(); err != nil {
			fmt.Printf("\rERRORE: %v\n", err)
			os.Exit(1)
		}
		s <- "fatto!"
	}()
	fmt.Printf("\r%s%v\n", prefix, <-s)
}

func spinner(msg string) {
	for {
		for _, r := range `-\|/` {
			// \r, carriage return posiziona il cursore ad inizio riga
			// e ricomincia a scrivere. Questa operazione simula la riscrittura
			// degli ultimi caratteri se si utilizza il medesimo prefisso.
			// prova a eliminare \r per vedere cosa intendo
			fmt.Printf("\r%s%c", msg, r)
			// per dare il senso rotatorio, scrivi i caratteri con
			// un intervallo di XXms uno dall'altro
			time.Sleep(70 * time.Millisecond)
		}
	}
}
