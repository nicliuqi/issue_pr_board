package controllers

import (
	"github.com/astaxie/beego"
)

type BaseController struct {
	beego.Controller
}

type JsonReturn struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func (c *BaseController) ApiJsonReturn(msg string, code int, data interface{}) {
	var JsonReturn JsonReturn
	JsonReturn.Msg = msg
	JsonReturn.Code = code
	JsonReturn.Data = data
	c.Data["json"] = JsonReturn
	c.ServeJSON()
	c.StopRun()
}

type DataReturn struct {
	Total   int64       `json:"total"`
	Page    int         `json:"page"`
	PerPage int         `json:"per_page"`
	Data    interface{} `json:"data"`
}

func (c *BaseController) ApiDataReturn(total int64, page int, per_page int, data interface{}) {
	var dr DataReturn
	dr.Total = total
	dr.Page = page
	dr.PerPage = per_page
	dr.Data = data
	c.Data["json"] = dr
	c.ServeJSON()
	c.StopRun()
}
