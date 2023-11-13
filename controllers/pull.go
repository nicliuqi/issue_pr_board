package controllers

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	"github.com/chenhg5/collection"
	"issue_pr_board/models"
	"issue_pr_board/utils"
	"sort"
	"strings"
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
	Label     string
	Exclusion string
	Search    string
	Sort      string
	Direction string
	Page      int
	PerPage   int
}

func formQueryPullSql(q QueryPullParam) (int64, string) {
	rawSql := "select * from pull where sig != 'Private'"
	org := q.Org
	repo := q.Repo
	sig := q.Sig
	state := q.State
	ref := q.Ref
	assignee := q.Assignee
	author := q.Author
	label := q.Label
	exclusion := q.Exclusion
	search := q.Search
	order := q.Sort
	direction := q.Direction
	page := q.Page
	perPage := q.PerPage
	org = utils.CheckParams(org)
	repo = utils.CheckParams(repo)
	sig = utils.CheckParams(sig)
	state = utils.CheckParams(state)
	ref = utils.CheckParams(ref)
	assignee = utils.CheckParams(assignee)
	author = utils.CheckParams(author)
	label = utils.CheckParams(label)
	exclusion = utils.CheckParams(exclusion)
	search = utils.CheckParams(search)
	if state != "" {
		stateSql := ""
		for index, stateStr := range strings.Split(state, ",") {
			if index == 0 {
				stateSql += fmt.Sprintf("state='%s'", stateStr)
			} else {
				stateSql += fmt.Sprintf(" or state='%s'", stateStr)
			}
		}
		rawSql += fmt.Sprintf(" and (%s)", stateSql)
	}
	if author != "" {
		authorSql := ""
		for index, atStr := range strings.Split(author, ",") {
			if index == 0 {
				authorSql += fmt.Sprintf("author='%s'", atStr)
			} else {
				authorSql += fmt.Sprintf(" or author='%s'", atStr)
			}
		}
		rawSql += fmt.Sprintf(" and (%s)", authorSql)
	}
	if assignee != "" {
		assigneeSql := ""
		for index, asStr := range strings.Split(assignee, ",") {
			if index == 0 {
				assigneeSql += fmt.Sprintf("find_in_set('%s', assignees)", asStr)
			} else {
				assigneeSql += fmt.Sprintf(" or find_in_set('%s', assignees)", asStr)
			}
		}
		rawSql += fmt.Sprintf(" and (%s)", assigneeSql)
	}
	if org != "" {
		rawSql += fmt.Sprintf(" and org='%s'", org)
	}
	if repo != "" {
		rawSql += fmt.Sprintf(" and repo='%s'", repo)
	}
	if sig != "" {
		rawSql += fmt.Sprintf(" and sig='%s'", sig)
	}
	if ref != "" {
		rawSql += fmt.Sprintf(" and ref='%s'", ref)
	}
	if label != "" {
		label = strings.Replace(label, "，", ",", -1)
		for _, labelStr := range strings.Split(label, ",") {
			rawSql += fmt.Sprintf(" and find_in_set('%s', labels)", labelStr)
		}
	}
	if exclusion != "" {
		exclusion = strings.Replace(exclusion, "，", ",", -1)
		for _, exclusionStr := range strings.Split(exclusion, ",") {
			rawSql += fmt.Sprintf(" and !find_in_set('%s', labels)", exclusionStr)
		}
	}
	if search != "" {
		searchSql := " and concat (repo, title, sig) like '%{search}%'"
		rawSql += strings.Replace(searchSql, "{search}", search, -1)
	}
	if order != "updated_at" {
		order = "created_at"
	}
	if direction == "asc" {
		rawSql += fmt.Sprintf(" order by %s asc", order)
	} else {
		rawSql += fmt.Sprintf(" order by %s desc", order)
	}
	o := orm.NewOrm()
	countSql := strings.Replace(rawSql, "*", "count(*)", -1)
	var sqlCount int
	_ = o.Raw(countSql).QueryRow(&sqlCount)
	offset := perPage * (page - 1)
	rawSql += fmt.Sprintf(" limit %v offset %v", perPage, offset)
	return int64(sqlCount), rawSql
}

func (c *PullsController) Get() {
	var pull []models.Pull
	page, _ := c.GetInt("page", 1)
	perPage, _ := c.GetInt("per_page", 10)
	if perPage > 100 {
		perPage = 100
	}
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
		Label:     c.GetString("label", ""),
		Exclusion: c.GetString("exclusion", ""),
		Search:    c.GetString("search", ""),
		Page:      page,
		PerPage:   perPage,
	}
	count, sql := formQueryPullSql(qp)
	o := orm.NewOrm()
	_, err := o.Raw(sql).QueryRows(&pull)
	if err == nil {
		c.ApiDataReturn(count, page, perPage, pull)
	} else {
		c.ApiJsonReturn("查询错误", 400, err)
	}
}

type PullsSigsController struct {
	BaseController
}

func (c *PullsSigsController) Get() {
	var pull []models.Pull
	keyWord := c.GetString("keyword", "")
	keyWord = utils.CheckParams(keyWord)
	o := orm.NewOrm()
	sql := "select distinct sig from pull where sig != 'Private' order by sig"
	_, err := o.Raw(sql).QueryRows(&pull)
	if err != nil {
		c.ApiJsonReturn("查询sigs错误", 400, err)
	}
	res := make([]string, 0)
	for _, i := range pull {
		res = append(res, i.Sig)
	}
	if keyWord == "" {
		c.ApiJsonReturn("请求成功", 200, res)
	} else {
		res2 := make([]string, 0)
		for _, j := range res {
			if strings.Contains(strings.ToLower(j), strings.ToLower(keyWord)) {
				res2 = append(res2, j)
			}
		}
		c.ApiJsonReturn("请求成功", 200, res2)
	}
}

type PullsReposController struct {
	BaseController
}

func (c *PullsReposController) Get() {
	var pull []models.Pull
	var pull2 []models.Pull
	page, _ := c.GetInt("page", 1)
	perPage, _ := c.GetInt("per_page", 10)
	if perPage > 100 {
		perPage = 100
	}
	offset := perPage * (page - 1)
	sig := c.GetString("sig", "")
	keyWord := c.GetString("keyword", "")
	sig = utils.CheckParams(sig)
	keyWord = utils.CheckParams(keyWord)
	o := orm.NewOrm()
	sql := ""
	if sig == "" {
		sql = "select distinct repo from pull where sig != 'Private' order by repo"
	} else {
		sql = fmt.Sprintf("select distinct repo from pull where sig != 'Private' and sig = '%s' order by repo", sig)
	}
	count, err := o.Raw(sql).QueryRows(&pull)
	if err != nil {
		c.ApiJsonReturn("查询repos错误", 400, err)
	}
	separateSql := sql + fmt.Sprintf(" limit %v offset %v", perPage, offset)
	_, err = o.Raw(separateSql).QueryRows(&pull2)
	if err != nil {
		c.ApiJsonReturn("分页查询repos错误", 400, err)
	}
	res := make([]string, 0)
	for _, i := range pull2 {
		res = append(res, i.Repo)
	}
	if keyWord == "" {
		c.ApiDataReturn(count, page, perPage, res)
	} else {
		newRes := make([]string, 0)
		for _, j := range pull {
			if strings.Contains(strings.ToLower(j.Repo), strings.ToLower(keyWord)) {
				newRes = append(newRes, j.Repo)
			}
		}
		count = int64(len(newRes))
		finalRes := make([]string, 0)
		if offset > int(count) {
			c.ApiDataReturn(count, page, perPage, finalRes)
		}
		if int(count) > offset && int(count) < perPage+offset {
			c.ApiDataReturn(count, page, perPage, newRes[offset:])
		}
		if int(count) == 0 {
			c.ApiDataReturn(count, page, perPage, finalRes)
		}
		c.ApiDataReturn(count, page, perPage, newRes[offset:offset+perPage])
	}
}

type PullsRefsController struct {
	BaseController
}

func (c *PullsRefsController) Get() {
	var pull []models.Pull
	var pull2 []models.Pull
	page, _ := c.GetInt("page", 1)
	perPage, _ := c.GetInt("per_page", 10)
	if perPage > 100 {
		perPage = 100
	}
	offset := perPage * (page - 1)
	keyWord := c.GetString("keyword", "")
	keyWord = utils.CheckParams(keyWord)
	o := orm.NewOrm()
	sql := "select distinct ref from pull where sig != 'Private' order by ref"
	count, err := o.Raw(sql).QueryRows(&pull)
	if err != nil {
		c.ApiJsonReturn("查询refs错误", 400, err)
	}
	separateSql := sql + fmt.Sprintf(" limit %v offset %v", perPage, offset)
	_, err = o.Raw(separateSql).QueryRows(&pull2)
	if err != nil {
		c.ApiJsonReturn("分页查询refs错误", 400, err)
	}
	res := make([]string, 0)
	for _, i := range pull2 {
		res = append(res, i.Ref)
	}
	if keyWord == "" {
		c.ApiDataReturn(count, page, perPage, res)
	} else {
		newRes := make([]string, 0)
		for _, j := range pull {
			if strings.Contains(strings.ToLower(j.Ref), strings.ToLower(keyWord)) {
				newRes = append(newRes, j.Ref)
			}
		}
		count = int64(len(newRes))
		finalRes := make([]string, 0)
		if offset > int(count) {
			c.ApiDataReturn(count, page, perPage, finalRes)
		}
		if int(count) > offset && int(count) < perPage+offset {
			c.ApiDataReturn(count, page, perPage, newRes[offset:])
		}
		if int(count) == 0 {
			c.ApiDataReturn(count, page, perPage, finalRes)
		}
		c.ApiDataReturn(count, page, perPage, newRes[offset:offset+perPage])
	}
}

type PullsAuthorsController struct {
	BaseController
}

func (c *PullsAuthorsController) Get() {
	var pull []models.Pull
	var pull2 []models.Pull
	page, _ := c.GetInt("page", 1)
	perPage, _ := c.GetInt("per_page", 20)
	keyWord := c.GetString("keyword", "")
	keyWord = utils.CheckParams(keyWord)
	if perPage > 100 {
		perPage = 100
	}
	offset := perPage * (page - 1)
	o := orm.NewOrm()
	sql := "select distinct author from pull where sig != 'Private' order by author"
	count, err := o.Raw(sql).QueryRows(&pull)
	if err != nil {
		c.ApiJsonReturn("查询错误", 400, err)
	}
	separateSql := sql + fmt.Sprintf(" limit %v offset %v", perPage, offset)
	_, err = o.Raw(separateSql).QueryRows(&pull2)
	if err != nil {
		c.ApiJsonReturn("分页查询错误", 400, err)
	}
	res := make([]string, 0)
	for _, i := range pull2 {
		author := i.Author
		res = append(res, author)
	}
	if keyWord == "" {
		c.ApiDataReturn(count, page, perPage, res)
	} else {
		newRes := make([]string, 0)
		for _, j := range pull {
			author := j.Author
			if strings.Contains(strings.ToLower(author), strings.ToLower(keyWord)) {
				newRes = append(newRes, author)
			}
		}
		count = int64(len(newRes))
		finalRes := make([]string, 0)
		if offset > int(count) {
			c.ApiDataReturn(count, page, perPage, finalRes)
		}
		if int(count) > offset && int(count) < perPage+offset {
			c.ApiDataReturn(count, page, perPage, newRes[offset:])
		}
		if int(count) == 0 {
			c.ApiDataReturn(count, page, perPage, finalRes)
		}
		c.ApiDataReturn(count, page, perPage, newRes[offset:offset+perPage])
	}
}

type PullsAssigneesController struct {
	BaseController
}

func (c *PullsAssigneesController) Get() {
	var pull []models.Pull
	page, _ := c.GetInt("page", 1)
	perPage, _ := c.GetInt("per_page", 20)
	keyWord := c.GetString("keyword", "")
	keyWord = utils.CheckParams(keyWord)
	if perPage > 100 {
		perPage = 100
	}
	offset := perPage * (page - 1)
	o := orm.NewOrm()
	sql := "select distinct assignees from pull where sig != 'Private'"
	_, err := o.Raw(sql).QueryRows(&pull)
	if err != nil {
		c.ApiJsonReturn("查询错误", 400, err)
	}
	res := make([]string, 0)
	for _, i := range pull {
		if i.Assignees == "" {
			continue
		}
		for _, j := range strings.Split(i.Assignees, ",") {
			if collection.Collect(res).Contains(j) {
				continue
			}
			if keyWord == "" {
				res = append(res, j)
			} else {
				if strings.Contains(strings.ToLower(j), strings.ToLower(keyWord)) {
					res = append(res, j)
				}
			}
		}
	}
	sort.Strings(res)
	count := int64(len(res))
	resp := make([]string, 0)
	if offset > int(count) {
		c.ApiDataReturn(count, page, perPage, resp)
	}
	if int(count) > offset && int(count) < perPage+offset {
		c.ApiDataReturn(count, page, perPage, res[offset:])
	}
	if int(count) == 0 {
		c.ApiDataReturn(count, page, perPage, resp)
	}
	c.ApiDataReturn(count, page, perPage, res[offset:offset+perPage])
}

type PullsLabelsController struct {
	BaseController
}

func (c *PullsLabelsController) Get() {
	var pull []models.Pull
	page, _ := c.GetInt("page", 1)
	perPage, _ := c.GetInt("per_page", 20)
	keyWord := c.GetString("keyword", "")
	keyWord = utils.CheckParams(keyWord)
	if perPage > 100 {
		perPage = 100
	}
	offset := perPage * (page - 1)
	o := orm.NewOrm()
	sql := "select distinct labels from pull where sig != 'Private'"
	_, err := o.Raw(sql).QueryRows(&pull)
	if err != nil {
		c.ApiJsonReturn("查询错误", 400, err)
	}
	res := make([]string, 0)
	for _, i := range pull {
		if i.Labels == "" {
			continue
		}
		for _, j := range strings.Split(i.Labels, ",") {
			if collection.Collect(res).Contains(j) {
				continue
			}
			if keyWord == "" {
				res = append(res, j)
			} else {
				if strings.Contains(strings.ToLower(j), strings.ToLower(keyWord)) {
					res = append(res, j)
				}
			}
		}
	}
	sort.Strings(res)
	count := int64(len(res))
	resp := make([]string, 0)
	if offset > int(count) {
		c.ApiDataReturn(count, page, perPage, resp)
	}
	if int(count) > offset && int(count) < perPage+offset {
		c.ApiDataReturn(count, page, perPage, res[offset:])
	}
	if int(count) == 0 {
		c.ApiDataReturn(count, page, perPage, resp)
	}
	c.ApiDataReturn(count, page, perPage, res[offset:offset+perPage])
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
