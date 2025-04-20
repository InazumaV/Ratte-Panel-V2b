package main

import (
	"github.com/InazumaV/Ratte-Interface/panel"
	v2b "github.com/Yuzuki616/Ratte-Panel-V2b"
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
