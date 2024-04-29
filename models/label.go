package models

import (
	"fmt"
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
	searchSql := fmt.Sprintf("select * from label where name='%s'", name)
	err := o.Raw(searchSql).QueryRow()
	if err == orm.ErrNoRows {
		return false
	}
	return true
}
