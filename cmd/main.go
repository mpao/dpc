package main

import (
	"fmt"
	"os"
)

var (
	appName    = "dpc"  // Name valore di fallback, usa il Taskfile per la definizione
	appVersion = "v0.0" // Version valore di fallback, usa il Taskfile per la definizione
)

func main() {
	if err := root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
