package models

import (
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
)

type Repo struct {
	Id               int    `json:"-"`
	Name             string `json:"repo" orm:"size(100);null" description:"仓库全名"`
	Sig              string `json:"sig" orm:"size(100);null" description:"所属sig"`
	EnterpriseNumber int    `json:"enterprise_number" orm:"null" description:"企业仓库ID"`
}

func GetProjectIdByRepoName(repo string) int {
	var r Repo
	o := orm.NewOrm()
	if err := o.QueryTable("repo").Filter("name", repo).One(&r); err != nil {
		logs.Error("[GetProjectIdByRepoName] Fail to query repo by project id, err:", err)
		return 0
	}
	return r.EnterpriseNumber
}

func GetSigByRepo(repo string) string {
	var r Repo
	o := orm.NewOrm()
	if err := o.QueryTable("repo").Filter("name", repo).One(&r); err != nil {
		logs.Error("[GetSigByRepo] Fail to query sig by repo name, err:", err)
		return ""
	}
	return r.Sig
}

func SearchRepoByNumber(number int) (sig string, repo string) {
	var r Repo
	o := orm.NewOrm()
	if err := o.QueryTable("repo").Filter("enterprise_number", number).One(&r); err != nil {
		logs.Error("[SearchRepoByNumber] Fail to query repo by enterprise number, err:", err)
		return "", ""
	}
	return r.Sig, r.Name
}

func SearchRepoRecord(repo string) bool {
	var r Repo
	o := orm.NewOrm()
	if err := o.QueryTable("repo").Filter("name", repo).One(&r); err != nil {
		logs.Error(fmt.Sprintf("[SearchRepoRecord] Fail to query the repo, repo: %v, err: %v", repo, err))
		return false
	}
	return true
}
