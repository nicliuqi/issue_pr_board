package models

import (
	"github.com/beego/beego/v2/client/orm"
	_ "github.com/go-sql-driver/mysql"
)

type Pull struct {
	Id          int    `json:"-"`
	Org         string `json:"org" orm:"size(20);null" description:"组织"`
	Repo        string `json:"repo" orm:"size(100);null" description:"仓库"`
	Ref         string `json:"ref" orm:"size(100);null" description:"目标分支"`
	Sig         string `json:"sig" orm:"size(100);null" description:"所属sig组"`
	Link        string `json:"link" orm:"size(255);null;unique" description:"链接"`
	State       string `json:"state" orm:"size(20);null" description:"状态"`
	Author      string `json:"author" orm:"size(50);null" description:"提交者"`
	Assignees   string `json:"assignees" orm:"size(255);null" description:"指派者"`
	CreatedAt   string `json:"created_at" orm:"size(20);null" description:"PR创建时间"`
	UpdatedAt   string `json:"updated_at" orm:"size(20);null" description:"PR更新时间"`
	Title       string `json:"title" orm:"type(text);null" description:"标题"`
	Description string `json:"-" orm:"type(text);null" description:"描述"`
	Labels      string `json:"labels" orm:"type(text);null" description:"标签"`
	Draft       bool   `json:"draft" orm:"null" description:"是否是草稿"`
	Mergeable   bool   `json:"mergeable" orm:"null" description:"是否可合入"`
}

func SearchPullRecord(htmlUrl string) bool {
	var pull Pull
	o := orm.NewOrm()
	if err := o.QueryTable("pull").Filter("link", htmlUrl).One(&pull); err != nil {
		return false
	}
	return true
}
