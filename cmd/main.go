package main

import (
	v2b "Ratte-Panel-V2b"
	"github.com/InazumaV/Ratte-Interface/panel"
	log "github.com/sirupsen/logrus"
)

func main() {
	p, err := panel.NewServer(nil, v2b.NewPanel())
	if err != nil {
		log.Fatalln(err)
	}
	if err = p.Run(); err != nil {
		log.Fatalln(err)
	}
}
