package main

import (
	"Ratte-Panel-V2b/crypt"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/InazumaV/Ratte-Interface/common/errors"
	"github.com/InazumaV/Ratte-Interface/panel"
	"github.com/InazumaV/Ratte-Interface/params"
	"strconv"
	"strings"
)

const (
	None    = 0
	Tls     = 1
	Reality = 2
)

type CommonNode struct {
	Host       string      `json:"host"`
	ServerPort int         `json:"server_port"`
	ServerName string      `json:"server_name"`
	Routes     []Route     `json:"routes"`
	BaseConfig *BaseConfig `json:"base_config"`
}

type Route struct {
	Id          int         `json:"id"`
	Match       interface{} `json:"match"`
	Action      string      `json:"action"`
	ActionValue string      `json:"action_value"`
}
type BaseConfig struct {
	PushInterval any `json:"push_interval"`
	PullInterval any `json:"pull_interval"`
}

// VAllssNode is vmess and vless node info
type VAllssNode struct {
	CommonNode
	Tls                 int             `json:"tls"`
	TlsSettings         TlsSettings     `json:"tls_settings"`
	TlsSettingsBack     *TlsSettings    `json:"tlsSettings"`
	Network             string          `json:"network"`
	NetworkSettings     json.RawMessage `json:"network_settings"`
	NetworkSettingsBack json.RawMessage `json:"networkSettings"`
	ServerName          string          `json:"server_name"`

	// vless only
	Flow          string        `json:"flow"`
	RealityConfig RealityConfig `json:"-"`
}

type TlsSettings struct {
	ServerName string `json:"server_name"`
	ServerPort string `json:"server_port"`
	ShortId    string `json:"short_id"`
	PrivateKey string `json:"private_key"`
}

type RealityConfig struct {
	Xver         uint64 `json:"Xver"`
	MinClientVer string `json:"MinClientVer"`
	MaxClientVer string `json:"MaxClientVer"`
	MaxTimeDiff  string `json:"MaxTimeDiff"`
}

type ShadowsocksNode struct {
	CommonNode
	Cipher    string `json:"cipher"`
	ServerKey string `json:"server_key"`
}

type TrojanNode CommonNode

type HysteriaNode struct {
	CommonNode
	UpMbps   int    `json:"up_mbps"`
	DownMbps int    `json:"down_mbps"`
	Obfs     string `json:"obfs"`
}

type RawDNS struct {
	DNSMap  map[string]map[string]interface{}
	DNSJson []byte
}

type Rules struct {
	Regexp   []string
	Protocol []string
}

func (p *Panel) GetNodeInfo(id int) (rsp *panel.GetNodeInfoRsp) {
	defer func() {
		if rsp.Err != nil {
			rsp.Err = errors.NewStringFromErr(rsp.Err)
		}
	}()
	rsp = &panel.GetNodeInfoRsp{}
	rm, ok := p.remotes.Get(KeyInt(id))
	if !ok {
		return &panel.GetNodeInfoRsp{
			Err: fmt.Errorf("remote node not found"),
		}
	}
	r, err := p.client.
		R().
		SetHeader("If-None-Match", rm.nodeEtag).
		SetQueryParams(map[string]string{
			"token":     rm.Key,
			"node_type": rm.NodeType,
			"node_id":   strconv.Itoa(rm.NodeId),
		}).
		Get(rm.Baseurl + "/api/v1/server/UniProxy/config")
	if err != nil {
		return &panel.GetNodeInfoRsp{
			Err: err,
		}
	}
	if r.StatusCode() != 200 {
		return &panel.GetNodeInfoRsp{
			Err: fmt.Errorf("get node info error: %s", r.String()),
		}
	}
	if r.StatusCode() == 304 {
		return &panel.GetNodeInfoRsp{
			Err:  nil,
			Hash: rm.nodeEtag,
		}
	}

	var cm = &CommonNode{}
	cn := panel.NodeInfo{
		Type: rm.NodeType,
	}
	switch rm.NodeType {
	case "vmess", "vless":
		rsp := &VAllssNode{}
		err = json.Unmarshal(r.Bytes(), rsp)
		if err != nil {
			return &panel.GetNodeInfoRsp{
				Err: fmt.Errorf("decode v2ray params error: %s", err),
			}
		}
		if len(rsp.NetworkSettingsBack) > 0 {
			rsp.NetworkSettings = rsp.NetworkSettingsBack
			rsp.NetworkSettingsBack = nil
		}
		if rsp.TlsSettingsBack != nil {
			rsp.TlsSettings = *rsp.TlsSettingsBack
			rsp.TlsSettingsBack = nil
		}

		if len(rsp.NetworkSettings) > 0 {
			err = json.Unmarshal(rsp.NetworkSettings, &rsp.RealityConfig)
			if err != nil {
				return &panel.GetNodeInfoRsp{
					Err: fmt.Errorf("decode reality config error: %s", err),
				}
			}
		}
		if rsp.Tls == Reality {
			if rsp.TlsSettings.PrivateKey == "" {
				key := crypt.GenX25519Private([]byte("vless" + rm.Key))
				rsp.TlsSettings.PrivateKey = base64.RawURLEncoding.EncodeToString(key)
			}
		}
		cm = &rsp.CommonNode
		cn.VMess = &params.VMessNode{
			CommonNodeParams: params.CommonNodeParams{
				Host: cm.Host,
				Port: strconv.Itoa(cm.ServerPort),
			},
			TlsType: rsp.Tls,
			TlsSettings: params.TlsSettings{
				ServerName: rsp.TlsSettings.ServerName,
				ServerPort: rsp.TlsSettings.ServerPort,
				ShortId:    rsp.TlsSettings.ShortId,
				PrivateKey: rsp.TlsSettings.PrivateKey,
			},
			Network:         rsp.Network,
			NetworkSettings: rsp.NetworkSettings,
			ServerName:      rsp.ServerName,
		}
	case "shadowsocks":
		rsp := &ShadowsocksNode{}
		err = json.Unmarshal(r.Bytes(), rsp)
		if err != nil {
			return &panel.GetNodeInfoRsp{
				Err: fmt.Errorf("decode ss params error: %s", err),
			}
		}
		cm = &rsp.CommonNode
		cn.Shadowsocks = &params.ShadowsocksNode{
			CommonNodeParams: params.CommonNodeParams{
				Host: cm.Host,
				Port: strconv.Itoa(cm.ServerPort),
			},
			Cipher:    rsp.Cipher,
			ServerKey: rsp.ServerKey,
		}
	case "trojan":
		rsp := &TrojanNode{}
		err = json.Unmarshal(r.Bytes(), rsp)
		if err != nil {
			return &panel.GetNodeInfoRsp{
				Err: fmt.Errorf("decode trojan params error: %s", err),
			}
		}
		cm = (*CommonNode)(rsp)
		cn.Trojan = &params.TrojanNode{
			Host: cm.Host,
			Port: strconv.Itoa(cm.ServerPort),
		}
	case "hysteria":
		rsp := &HysteriaNode{}
		err = json.Unmarshal(r.Bytes(), rsp)
		if err != nil {
			return &panel.GetNodeInfoRsp{
				Err: fmt.Errorf("decode hysteria params error: %s", err),
			}
		}
		cm = &rsp.CommonNode
		cn.Hysteria = &params.HysteriaNode{
			CommonNodeParams: params.CommonNodeParams{
				Host: cm.Host,
				Port: strconv.Itoa(cm.ServerPort),
			},
			UpMbps:   rsp.UpMbps,
			DownMbps: rsp.DownMbps,
			Obfs:     rsp.Obfs,
		}
	}

	// parse rules and dns
	for i := range cm.Routes {
		var matchs []string
		if _, ok := cm.Routes[i].Match.(string); ok {
			matchs = strings.Split(cm.Routes[i].Match.(string), ",")
		} else if _, ok = cm.Routes[i].Match.([]string); ok {
			matchs = cm.Routes[i].Match.([]string)
		} else {
			temp := cm.Routes[i].Match.([]interface{})
			matchs = make([]string, len(temp))
			for i := range temp {
				matchs[i] = temp[i].(string)
			}
		}
		switch cm.Routes[i].Action {
		case "block":
			cn.Rules = matchs
		}
	}
	rm.nodeEtag = r.Header().Get("ETag")
	return &panel.GetNodeInfoRsp{
		NodeInfo: cn,
		Hash:     rm.nodeEtag,
	}
}
