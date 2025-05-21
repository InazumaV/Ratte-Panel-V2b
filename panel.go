package v2b

import (
	"github.com/InazumaV/Ratte-Interface/panel"
	"github.com/orcaman/concurrent-map/v2"
	"resty.dev/v3"
	"strconv"
)

var _ panel.Panel = (*Panel)(nil)

type Panel struct {
	client  *resty.Client
	remotes cmap.ConcurrentMap[KeyInt, *Remote]
}

func (p *Panel) Type() string {
	return "Ratte-Panel-V2b"
}

func (p *Panel) CustomMethod(method string, args any, reply *any) error {
	return nil
}

type KeyInt int

func (k KeyInt) String() string {
	return strconv.Itoa(int(k))
}

type Remote struct {
	nodeEtag string
	userEtag string
	*panel.AddRemoteParams
}

func NewPanel() *Panel {
	return &Panel{
		client:  resty.New(),
		remotes: cmap.NewStringer[KeyInt, *Remote](),
	}
}
