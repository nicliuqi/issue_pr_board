package models

type Repo struct {
	Id               int    `json:"-"`
	Name             string `json:"repo" orm:"size(100);null" description:"仓库全名"`
	Sig              string `json:"sig" orm:"size(100);null" description:"所属sig"`
	Branches         string `json:"branches" orm:"type(text);null" description:"分支"`
	Reviewers        string `json:"reviewers" orm:"type(text);null" description:"审查者"`
	EnterpriseNumber int    `json:"enterprise_number" orm:"null" description:"企业仓库ID"`
	Public           bool   `json:"public" orm:"null" description:"是否公开"`
	Status           string `json:"status" orm:"size(20);null" description:"仓库状态"`
	CreatedAt        string `json:"created_at" orm:"size(50);null" description:"创建时间"`
	UpdatedAt        string `json:"updated_at" orm:"size(50);null" description:"更新时间"`
}
