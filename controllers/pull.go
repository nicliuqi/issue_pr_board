package controllers

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	"issue_pr_board/models"
)

type PullsController struct {
	BaseController
}

type QueryPullParam struct {
	Org       string
	Repo      string
	Sig       string
	State     string
	Ref       string
	Author    string
	Assignee  string
	Sort      string
	Direction string
	Page      int
	PerPage   int
}

func formQueryPullSql(q QueryPullParam) (int64, string) {
	rawSql := "select * from pull"
	org := q.Org
	repo := q.Repo
	sig := q.Sig
	state := q.State
	ref := q.Ref
	assignee := q.Assignee
	author := q.Author
	sort := q.Sort
	direction := q.Direction
	page := q.Page
	perPage := q.PerPage
	if state != "" {
		if len(rawSql) == 18 {
			rawSql += fmt.Sprintf(" where state='%s'", state)
		} else {
			rawSql += fmt.Sprintf(" and state='%s'", state)
		}
	}
	if org != "" {
		if len(rawSql) == 18 {
			rawSql += fmt.Sprintf(" where org='%s'", org)
		} else {
			rawSql += fmt.Sprintf(" and org='%s'", org)
		}
	}
	if repo != "" {
		if len(rawSql) == 18 {
			rawSql += fmt.Sprintf(" where repo='%s'", repo)
		} else {
			rawSql += fmt.Sprintf(" and repo='%s'", repo)
		}
	}
	if sig != "" {
		if len(rawSql) == 18 {
			rawSql += fmt.Sprintf(" where sig='%s'", sig)
		} else {
			rawSql += fmt.Sprintf(" and sig='%s'", sig)
		}
	}
	if ref != "" {
		if len(rawSql) == 18 {
			rawSql += fmt.Sprintf(" where ref='%s'", ref)
		} else {
			rawSql += fmt.Sprintf(" and ref='%s'", ref)
		}
	}
	if author != "" {
		if len(rawSql) == 18 {
			rawSql += fmt.Sprintf(" where author='%s'", author)
		} else {
			rawSql += fmt.Sprintf(" and author='%s'", author)
		}
	}
	if assignee != "" {
		if len(rawSql) == 18 {
			rawSql += fmt.Sprintf(" where instr (assignees, '%s')", assignee)
		} else {
			rawSql += fmt.Sprintf(" and instr (assignees, '%s')", assignee)
		}
	}
	if sort != "updated_at" {
		sort = "created_at"
	}
	if direction == "asc" {
		rawSql += fmt.Sprintf(" order by %s asc", sort)
	} else {
		rawSql += fmt.Sprintf(" order by %s desc", sort)
	}
	var pull []models.Pull
	o := orm.NewOrm()
	count, _ := o.Raw(rawSql).QueryRows(&pull)
	offset := perPage * (page - 1)
	rawSql += fmt.Sprintf(" limit %v offset %v", perPage, offset)
	return count, rawSql
}

func (c *PullsController) Get() {
	var pull []models.Pull
	page, _ := c.GetInt("page", 1)
	perPage, _ := c.GetInt("per_page", 10)
	qp := QueryPullParam{
		Org:       c.GetString("org", ""),
		Repo:      c.GetString("repo", ""),
		Sig:       c.GetString("sig", ""),
		State:     c.GetString("state", ""),
		Ref:       c.GetString("ref", ""),
		Author:    c.GetString("author", ""),
		Assignee:  c.GetString("assignee", ""),
		Sort:      c.GetString("sort", ""),
		Direction: c.GetString("direction", ""),
		Page:      page,
		PerPage:   perPage,
	}
	count, sql := formQueryPullSql(qp)
	o := orm.NewOrm()
	_, err := o.Raw(sql).QueryRows(&pull)
	if err == nil {
		c.ApiDataReturn(count, page, perPage, pull)
	}
}

func SearchPullRecord(htmlUrl string) bool {
	o := orm.NewOrm()
	searchSql := fmt.Sprintf("select * from pull where link='%s'", htmlUrl)
	err := o.Raw(searchSql).QueryRow()
	if err == orm.ErrNoRows {
		return false
	}
	return true
}
