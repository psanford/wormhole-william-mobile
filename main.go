package main

import (
	"log"
	"os"

	"gioui.org/app"
	"github.com/psanford/wormhole-william-mobile/ui"
)

func main() {
	gui := ui.New()
	go func() {
		if err := gui.Run(); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}
