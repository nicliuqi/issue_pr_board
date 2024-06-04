package models

import (
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
)

type Issue struct {
	Id          int    `json:"-"`
	Org         string `json:"org" orm:"size(20);null" description:"组织"`
	Repo        string `json:"repo" orm:"size(100);null"  description:"仓库"`
	Sig         string `json:"sig" orm:"size(100);null" description:"所属sig组"`
	Link        string `json:"link" orm:"size(255);null" description:"链接"`
	Number      string `json:"number" orm:"size(10);null;unique" description:"issue编号"`
	State       string `json:"state" orm:"size(20);null" description:"状态"`
	IssueType   string `json:"issue_type" orm:"size(20);null" description:"issue类型"`
	IssueState  string `json:"issue_state" orm:"size(20);null" description:"issue状态"`
	Author      string `json:"author" orm:"size(50);null" description:"提交者"`
	Reporter    string `json:"reporter" orm:"size(100);null" description:"邮件认证提交者"`
	Assignee    string `json:"assignee" orm:"size(50);null" description:"指派者"`
	CreatedAt   string `json:"created_at" orm:"size(20);null" description:"issue创建时间"`
	UpdatedAt   string `json:"updated_at" orm:"size(20);null" description:"issue更新时间"`
	Title       string `json:"title" orm:"size(191);null" description:"标题"`
	Description string `json:"-" orm:"type(text);null" description:"描述"`
	Priority    string `json:"priority" orm:"size(10);null" description:"优先级"`
	Labels      string `json:"labels" orm:"type(text);null" description:"标签"`
	Branch      string `json:"branch" orm:"size(100);null" description:"指定分支"`
	Milestone   string `json:"milestone" orm:"size(255);null" description:"里程碑"`
}

func SearchIssueRecord(number string) bool {
	var issue Issue
	o := orm.NewOrm()
	if err := o.QueryTable("issue").Filter("number", number).One(&issue); err != nil {
		logs.Error(fmt.Sprintf("[SearchIssueRecord] Fail to query issue record, issue number: %v, err: %v",
			number, err))
		return false
	}
	return true
}

func GetIssuePriority(priorityNum float64) string {
	switch priorityNum {
	case 0:
		return "不指定"
	case 1:
		return "不重要"
	case 2:
		return "次要"
	case 3:
		return "主要"
	case 4:
		return "严重"
	default:
		return "不指定"
	}
}
