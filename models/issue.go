package models

type Issue struct {
	Id         int    `json:"-"`
	Org        string `json:"org" orm:"size(20);null" description:"组织"`
	Repo       string `json:"repo" orm:"size(100);null"  description:"仓库"`
	Sig        string `json:"sig" orm:"size(100);null" description:"所属sig组"`
	Link       string `json:"link" orm:"size(255);null" description:"链接"`
	Number     string `json:"number" orm:"size(10);null;unique" description:"issue编号"`
	State      string `json:"state" orm:"size(20);null" description:"状态"`
	IssueType  string `json:"issue_type" orm:"size(20);null" description:"issue类型"`
	IssueState string `json:"issue_state" orm:"size(20);null" description:"issue状态"`
	Author     string `json:"author" orm:"size(50);null" description:"提交者"`
	Reporter   string `json:"reporter" orm:"size(100);null" description:"邮件认证提交者"`
	Assignee   string `json:"assignee" orm:"size(50);null" description:"指派者"`
	CreatedAt  string `json:"created_at" orm:"size(20);null" description:"issue创建时间"`
	UpdatedAt  string `json:"updated_at" orm:"size(20);null" description:"issue更新时间"`
	Title      string `json:"title" orm:"size(255);null" description:"标题"`
	Priority   string `json:"priority" orm:"size(10);null" description:"优先级"`
	Labels     string `json:"labels" orm:"type(text);null" description:"标签"`
	Branch     string `json:"branch" orm:"size(100);null" description:"指定分支"`
}
