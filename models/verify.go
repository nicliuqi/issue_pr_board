package models

import "github.com/beego/beego/v2/client/orm"

type Verify struct {
	Id      int
	Addr    string
	Code    string
	Created int64
}

func SearchEmailRecord(addr string) bool {
	var verify Verify
	o := orm.NewOrm()
	if err := o.QueryTable("verify").Filter("addr", addr).One(&verify); err != nil {
		return false
	}
	return true
}

func CheckCode(addr string, code string) bool {
	var verify Verify
	o := orm.NewOrm()
	if err := o.QueryTable("verify").Filter("addr", addr).Filter("code", code).One(&verify); err != nil {
		return false
	}
	return true
}
