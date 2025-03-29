package main

import "github.com/InazumaV/Ratte-Interface/panel"

import td "Ratte-Panel-V2b/test_data"

var p *Panel
var id int

func init() {
	p = NewPanel()
	rsp := p.AddRemote(&panel.AddRemoteParams{
		Name:     "test",
		Baseurl:  td.PanelUrl,
		NodeId:   td.PanelNodeId,
		NodeType: td.PanelNodeType,
		Key:      td.PanelKey,
	})
	if rsp.Err != nil {
		panic(rsp.Err)
	}
	id = rsp.RemoteId
}
