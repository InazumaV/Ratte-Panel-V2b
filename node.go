package v2b

import (
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

func parseSecurityConfig(
	tp int,
	t *TlsSettings,
	r *RealityConfig,
	n *panel.NodeInfo,
) (sec *params.SecurityConfig) {
	switch tp {
	case Tls:
		n.Security = "tls"
		n.SecurityConfig = &params.SecurityConfig{
			TlsSettings: params.TlsSettings{
				ServerName: t.ServerName,
			},
		}
	case Reality:
		n.Security = "reality"
		sp, _ := strconv.Atoi(t.ServerPort)
		n.SecurityConfig = &params.SecurityConfig{
			RealityConfig: params.RealityConfig{
				Xver:         r.Xver,
				MinClientVer: r.MinClientVer,
				MaxClientVer: r.MaxClientVer,
				MaxTimeDiff:  r.MaxTimeDiff,
				ServerName:   t.ServerName,
				ServerPort:   sp,
			},
		}
	case None:
		n.Security = ""
	}
	return sec
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
	if r.StatusCode() == 304 {
		return &panel.GetNodeInfoRsp{
			Err:  nil,
			Hash: rm.nodeEtag,
		}
	}
	if r.StatusCode() != 200 {
		return &panel.GetNodeInfoRsp{
			Err: fmt.Errorf("get node info error: %s", r.String()),
		}
	}

	var cm = &CommonNode{}
	var cn panel.NodeInfo

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

		cn = panel.NodeInfo{
			Type: rm.NodeType,
			Port: rsp.ServerPort,
		}

		parseSecurityConfig(rsp.Tls, &rsp.TlsSettings, &rsp.RealityConfig, &cn)
		cm = &rsp.CommonNode
		switch rsp.Network {
		case "ws":
			cn.VMess = &params.VMess{
				Network: rsp.Network,
			}
			err := json.Unmarshal(rsp.NetworkSettings, &cn.VMess.Ws)
			if err != nil {
				return &panel.GetNodeInfoRsp{
					Err: fmt.Errorf("decode ws params error: %s", err),
				}
			}
		case "grpc":
			cn.VMess = &params.VMess{
				Network: rsp.Network,
			}
			err := json.Unmarshal(rsp.NetworkSettings, &cn.VMess.Grpc)
			if err != nil {
				return &panel.GetNodeInfoRsp{
					Err: fmt.Errorf("decode grpc params error: %s", err),
				}
			}
		}
	case "shadowsocks":
		rsp := &ShadowsocksNode{}
		err = json.Unmarshal(r.Bytes(), rsp)
		if err != nil {
			return &panel.GetNodeInfoRsp{
				Err: fmt.Errorf("decode ss params error: %s", err),
			}
		}
		cn = panel.NodeInfo{
			Type: rm.NodeType,
			Port: rsp.ServerPort,
		}
		cm = &rsp.CommonNode
		cn.Shadowsocks = &params.Shadowsocks{
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
		cn = panel.NodeInfo{
			Type: rm.NodeType,
			Port: rsp.ServerPort,
		}
		cn.Trojan = &params.Trojan{
			Host: cm.Host,
		}
		cm = (*CommonNode)(rsp)
	case "hysteria":
		rsp := &HysteriaNode{}
		err = json.Unmarshal(r.Bytes(), rsp)
		if err != nil {
			return &panel.GetNodeInfoRsp{
				Err: fmt.Errorf("decode hysteria params error: %s", err),
			}
		}
		cn = panel.NodeInfo{
			Type: rm.NodeType,
			Port: rsp.ServerPort,
		}
		cm = &rsp.CommonNode
		cn.Hysteria = &params.Hysteria{
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
