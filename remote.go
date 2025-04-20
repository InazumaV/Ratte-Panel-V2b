package v2b

import "github.com/InazumaV/Ratte-Interface/panel"

func (p *Panel) AddRemote(params *panel.AddRemoteParams) *panel.AddRemoteRsp {
	id := p.remotes.Count() + 1
	p.remotes.Set(KeyInt(id), &Remote{
		AddRemoteParams: params,
	})
	return &panel.AddRemoteRsp{
		RemoteId: id,
		Err:      nil,
	}
}

func (p *Panel) DelRemote(id int) error {
	p.remotes.Remove(KeyInt(id))
	return nil
}
