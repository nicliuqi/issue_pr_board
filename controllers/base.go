package controllers

import (
	beego "github.com/beego/beego/v2/server/web"
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
	var jsonReturn JsonReturn
	jsonReturn.Msg = msg
	jsonReturn.Code = code
	jsonReturn.Data = data
	c.Data["json"] = jsonReturn
	c.Ctx.ResponseWriter.WriteHeader(code)
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
