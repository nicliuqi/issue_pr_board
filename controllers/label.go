package controllers

import (
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"issue_pr_board/models"
)

type LabelsColorsController struct {
	BaseController
}

func (c *LabelsColorsController) Get() {
	var labels []models.Label
	sql := "select * from label"
	o := orm.NewOrm()
	_, err := o.Raw(sql).QueryRows(&labels)
	if err != nil {
		logs.Error("查询label失败", 400, err)
	}
	c.ApiJsonReturn("请求成功", 200, labels)
}
