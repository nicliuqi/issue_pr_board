package controllers

import (
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"

	"issue_pr_board/models"
)

type LabelsColorsController struct {
	BaseController
}

func (c *LabelsColorsController) Get() {
	var labels []models.Label
	o := orm.NewOrm()
	if _, err := o.QueryTable(&models.Label{}).All(&labels); err != nil {
		logs.Error("查询label失败", 400, err)
	}
	c.ApiJsonReturn("请求成功", 200, labels)
}
