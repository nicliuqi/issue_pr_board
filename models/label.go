package models

import (
	"github.com/beego/beego/v2/client/orm"
)

type Label struct {
	Id       int     `json:"-"`
	Name     string  `json:"name"`
	Color    string  `json:"color"`
	UniqueId float64 `json:"-"`
}

func SearchLabel(name string) bool {
	o := orm.NewOrm()
	searchSql := "select * from label where name=?"
	err := o.Raw(searchSql, name).QueryRow()
	if err == orm.ErrNoRows {
		return false
	}
	return true
}
