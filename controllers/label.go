package controllers

import (
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	"net/http"

	"issue_pr_board/models"
)

type LabelsColorsController struct {
	BaseController
}

func (c *LabelsColorsController) Get() {
	var labels []models.Label
	o := orm.NewOrm()
	if _, err := o.QueryTable("label").All(&labels); err != nil {
		logs.Error("Fail to query label colors:", err)
		c.ApiJsonReturn("Fail to query label colors", http.StatusBadRequest, nil)
	}
	c.ApiJsonReturn("Success", http.StatusOK, labels)
}
