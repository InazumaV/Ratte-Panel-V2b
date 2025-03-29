package main

import (
	"encoding/json"
	"fmt"
	"github.com/InazumaV/Ratte-Interface/common/errors"
	"github.com/InazumaV/Ratte-Interface/panel"
	"github.com/InazumaV/Ratte-Interface/params"
	"strconv"
)

type UserInfo struct {
	Id         int    `json:"id"`
	Uuid       string `json:"uuid"`
	SpeedLimit int    `json:"speed_limit"`
}

type getUserListRsp struct {
	//Msg  string `json:"msg"`
	Users []UserInfo `json:"users"`
}

func (p *Panel) GetUserList(id int) (rsp *panel.GetUserListRsp) {
	defer func() {
		if rsp.Err != nil {
			rsp.Err = errors.NewStringFromErr(rsp.Err)
		}
	}()
	rm, _ := p.remotes.Get(KeyInt(id))
	r, err := p.client.R().
		SetHeader("If-None-Match", rm.userEtag).SetQueryParams(map[string]string{
		"token":     rm.Key,
		"node_type": rm.NodeType,
		"node_id":   strconv.Itoa(rm.NodeId),
	}).
		Get(rm.Baseurl + "/api/v1/server/UniProxy/user")
	if err != nil {
		return &panel.GetUserListRsp{
			Err: err,
		}
	}
	if r.StatusCode() == 304 {
		return &panel.GetUserListRsp{
			Hash: rm.userEtag,
		}
	}
	var userList getUserListRsp
	err = json.Unmarshal(r.Bytes(), &userList)
	if err != nil {
		return &panel.GetUserListRsp{
			Err: fmt.Errorf("unmarshal userlist error: %s", err),
		}
	}
	rm.userEtag = r.Header().Get("ETag")
	rsp = &panel.GetUserListRsp{
		Hash:  rm.userEtag,
		Users: make([]panel.UserInfo, 0, len(userList.Users)),
	}
	for _, user := range userList.Users {
		rsp.Users = append(rsp.Users, panel.UserInfo{
			HashOrKey: fmt.Sprintf("%s-%d", user.Uuid, user.SpeedLimit),
			UserInfo: params.UserInfo{
				Id:   user.Id,
				Name: user.Uuid,
				Key:  []string{user.Uuid},
			},
		})
	}
	return rsp
}

type UserTraffic struct {
	UID      int
	Upload   int64
	Download int64
}

func (p *Panel) ReportUserTraffic(pms *panel.ReportUserTrafficParams) (err error) {
	defer func() {
		if err != nil {
			err = errors.NewStringFromErr(err)
		}
	}()
	rm, _ := p.remotes.Get(KeyInt(pms.Id))
	data := make([]UserTraffic, 0, len(pms.Users))
	for _, user := range pms.Users {
		data = append(data, UserTraffic{
			UID:      user.Id,
			Upload:   user.Upload,
			Download: user.Download,
		})
	}
	r, err := p.client.R().
		SetBody(data).
		SetContentType("application/json").SetQueryParams(map[string]string{
		"token":     rm.Key,
		"node_type": rm.NodeType,
		"node_id":   strconv.Itoa(rm.NodeId),
	}).
		Post(rm.Baseurl + "/api/v1/server/UniProxy/user")
	if err != nil {
		return err
	}
	if r.StatusCode() != 200 {
		return fmt.Errorf("report user traffic error: %s", r.String())
	}
	return nil
}
