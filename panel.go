package main

import (
	"github.com/InazumaV/Ratte-Interface/panel"
)

type Panel struct {
}

func (p *Panel) AddRemote(params *panel.AddRemoteParams) *panel.AddRemoteRsp {
	//TODO implement me
	panic("implement me")
}

func (p *Panel) DelRemote(id int) error {
	//TODO implement me
	panic("implement me")
}

func (p *Panel) GetNodeInfo(id int) *panel.GetNodeInfoRsp {
	//TODO implement me
	panic("implement me")
}

func (p *Panel) GetUserList(id int) *panel.GetUserListRsp {
	//TODO implement me
	panic("implement me")
}

func (p *Panel) ReportUserTraffic(pms *panel.ReportUserTrafficParams) error {
	//TODO implement me
	panic("implement me")
}

func NewPanel() *Panel {
	return &Panel{}
}
