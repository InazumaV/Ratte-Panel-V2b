package main

import (
	"github.com/InazumaV/Ratte-Interface/panel"
	log "github.com/sirupsen/logrus"
)

func main() {
	p, err := panel.NewServer(nil, NewPanel())
	if err != nil {
		log.Fatalln(err)
	}
	if err = p.Run(); err != nil {
		log.Fatalln(err)
	}
}
