package models

import (
	"encoding/json"
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	"io"
	"issue_pr_board/config"
	"net/http"
)

type Label struct {
	Id       int     `json:"-"`
	Name     string  `json:"name"`
	Color    string  `json:"color"`
	UniqueId float64 `json:"-"`
}

func SearchLabel(name string) bool {
	var l Label
	o := orm.NewOrm()
	if err := o.QueryTable("label").Filter("name", name).One(&l); err != nil {
		logs.Error(fmt.Sprintf("[SearchLabel] Fail to search the label %v, err: %v", name, err))
		return false
	}
	return true
}

type ResponseEnterpriseLabel struct {
	Name  string  `json:"name"`
	Color string  `json:"color"`
	Id    float64 `json:"id"`
}

func InitLabels() {
	url := fmt.Sprintf("%v/enterprises/open_euler/labels?access_token=%v", config.AppConfig.GiteeV5ApiPrefix,
		config.AppConfig.AccessToken)
	resp, err := http.Get(url)
	if err != nil {
		logs.Error("[InitLabels] Fail to get enterprise labels colors")
	}
	if resp.StatusCode != http.StatusOK {
		logs.Error("[InitLabels] Get unexpected response when getting enterprise labels, detail:", resp.Status)
	}
	body, _ := io.ReadAll(resp.Body)
	if err = resp.Body.Close(); err != nil {
		logs.Error("[InitLabels] Fail to close response body of enterprise labels, err:", err)
	}
	var rels []ResponseEnterpriseLabel
	if err = json.Unmarshal(body, &rels); err != nil {
		logs.Error("[InitLabels] Fail to unmarshal response, err:", err)
		return
	}
	o := orm.NewOrm()
	for _, i := range rels {
		var lb Label
		lb.Name = i.Name
		lb.Color = i.Color
		lb.UniqueId = i.Id
		if !SearchLabel(lb.Name) {
			if _, err = o.Insert(&lb); err != nil {
				logs.Error("[InitLabels] Fail to create label, err:", err)
			}
		} else {
			if _, err = o.QueryTable("label").Filter("name", lb.Name).Update(orm.Params{
				"color":     lb.Color,
				"unique_id": lb.UniqueId,
			}); err != nil {
				logs.Error("[InitLabels] Fail to update label, err:", err)
			}
		}
	}
}
