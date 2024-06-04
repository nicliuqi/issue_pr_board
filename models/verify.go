package models

import (
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
)

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
		logs.Error("[SearchEmailRecord] Fail to query verify by email address, err:", err)
		return false
	}
	return true
}

func CheckCode(addr string, code string) bool {
	var verify Verify
	o := orm.NewOrm()
	if err := o.QueryTable("verify").Filter("addr", addr).Filter("code", code).One(&verify); err != nil {
		logs.Error(fmt.Sprintf("[CheckCode] Fail to query verify by email address and code, addr: %v, code:"+
			"%v, err: %v", addr, code, err))
		return false
	}
	return true
}
