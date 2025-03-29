package main

import "github.com/InazumaV/Ratte-Interface/panel"

var p *Panel
var id int

func init() {
	p = NewPanel()
	rsp := p.AddRemote(&panel.AddRemoteParams{
		Name:     "test",
		Baseurl:  "https://apiv2b.sakuran.org",
		NodeId:   157,
		NodeType: "vmess",
		Key:      "29KUSBoq5g4SVF2L",
	})
	if rsp.Err != nil {
		panic(rsp.Err)
	}
	id = rsp.RemoteId
}
