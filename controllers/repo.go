package controllers

import (
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/go-playground/validator/v10"
	"issue_pr_board/models"
	"issue_pr_board/utils"
	"net/http"
	"strconv"
	"strings"
)

type ReposController struct {
	BaseController
}

type QueryRepoParam struct {
	Keyword   string `validate:"max=100"`
	Sig       string `validate:"max=100"`
	Page      int    `validate:"min=1,max=100000"`
	PerPage   int    `validate:"min=1,max=100"`
	Direction string `validate:"max=4"`
}

func formQueryRepoSql(q QueryRepoParam) (int64, string, []string) {
	sqlParams := make([]string, 0, 0)
	rawSql := "select * from repo where enterprise_number != 0"
	keyword := q.Keyword
	sig := q.Sig
	page := q.Page
	perPage := q.PerPage
	direction := q.Direction
	sig = utils.CheckParams(sig)
	keyword = utils.CheckParams(keyword)
	if keyword != "" {
		rawSql += fmt.Sprintf(" and instr (name, ?)")
		sqlParams = append(sqlParams, strings.ToLower(keyword))
	}
	if sig != "" {
		rawSql += fmt.Sprintf(" and sig=?")
		sqlParams = append(sqlParams, sig)
	}
	if direction != "desc" {
		rawSql += " order by name"
	} else {
		rawSql += " order by name desc"
	}
	var repo []models.Repo
	o := orm.NewOrm()
	count, err := o.Raw(rawSql, sqlParams).QueryRows(&repo)
	if err != nil {
		return 0, "select * from repo", sqlParams
	}
	offset := perPage * (page - 1)
	rawSql += " limit ? offset ?"
	sqlParams = append(sqlParams, strconv.Itoa(perPage), strconv.Itoa(offset))
	return count, rawSql, sqlParams
}

type responseRepo struct {
	Name string `json:"repo"`
	Sig  string `json:"sig"`
}

func (c *ReposController) Get() {
	var repo []responseRepo
	page, _ := c.GetInt("page", 1)
	perPage, _ := c.GetInt("per_page", 10)
	qp := QueryRepoParam{
		Keyword:   c.GetString("keyword", ""),
		Sig:       c.GetString("sig", ""),
		Page:      page,
		PerPage:   perPage,
		Direction: c.GetString("direction", ""),
	}
	validate := validator.New()
	if validateErr := validate.Struct(qp); validateErr != nil {
		c.ApiJsonReturn("Invalid params", http.StatusBadRequest, nil)
	}
	count, sql, sqlParams := formQueryRepoSql(qp)
	o := orm.NewOrm()
	if _, err := o.Raw(sql, sqlParams).QueryRows(&repo); err == nil {
		c.ApiDataReturn(count, page, perPage, repo)
	} else {
		c.ApiJsonReturn("Query error", http.StatusBadRequest, nil)
	}
}
