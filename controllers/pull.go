package controllers

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/beego/beego/v2/client/orm"
	"github.com/go-playground/validator/v10"

	"issue_pr_board/models"
	"issue_pr_board/utils"
)

type PullsController struct {
	BaseController
}

type QueryPullParam struct {
	Org       string `validate:"max=20"`
	Repo      string `validate:"max=100"`
	Sig       string `validate:"max=100"`
	State     string `validate:"max=20"`
	Ref       string `validate:"max=100"`
	Author    string `validate:"max=50"`
	Assignee  string `validate:"max=50"`
	Label     string
	Exclusion string
	Search    string `validate:"max=255"`
	Sort      string `validate:"max=10"`
	Direction string `validate:"max=4"`
	Page      int    `validate:"min=1,max=1000000"`
	PerPage   int    `validate:"min=1,max=100"`
}

func formQueryPullSql(q QueryPullParam) (int64, string, []string) {
	sqlParams := make([]string, 0, 0)
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
				stateSql += fmt.Sprintf("state=?")
				sqlParams = append(sqlParams, stateStr)
			} else {
				stateSql += fmt.Sprintf(" or state=?")
				sqlParams = append(sqlParams, stateStr)
			}
		}
		rawSql += fmt.Sprintf(" and (%s)", stateSql)
	}
	if author != "" {
		authorSql := ""
		for index, atStr := range strings.Split(author, ",") {
			if index == 0 {
				authorSql += fmt.Sprintf("author=?")
				sqlParams = append(sqlParams, atStr)
			} else {
				authorSql += fmt.Sprintf(" or author=?")
				sqlParams = append(sqlParams, atStr)
			}
		}
		rawSql += fmt.Sprintf(" and (%s)", authorSql)
	}
	if assignee != "" {
		assigneeSql := ""
		for index, asStr := range strings.Split(assignee, ",") {
			if index == 0 {
				assigneeSql += fmt.Sprintf("find_in_set(?, assignees)")
				sqlParams = append(sqlParams, asStr)
			} else {
				assigneeSql += fmt.Sprintf(" or find_in_set(?, assignees)")
				sqlParams = append(sqlParams, asStr)
			}
		}
		rawSql += fmt.Sprintf(" and (%s)", assigneeSql)
	}
	if org != "" {
		rawSql += fmt.Sprintf(" and org=?")
		sqlParams = append(sqlParams, org)
	}
	if repo != "" {
		rawSql += fmt.Sprintf(" and repo=?")
		sqlParams = append(sqlParams, repo)
	}
	if sig != "" {
		rawSql += fmt.Sprintf(" and sig=?")
		sqlParams = append(sqlParams, sig)
	}
	if ref != "" {
		rawSql += fmt.Sprintf(" and ref=?")
		sqlParams = append(sqlParams, ref)
	}
	if label != "" {
		label = strings.Replace(label, "，", ",", -1)
		for _, labelStr := range strings.Split(label, ",") {
			rawSql += fmt.Sprintf(" and find_in_set(?, labels)")
			sqlParams = append(sqlParams, labelStr)
		}
	}
	if exclusion != "" {
		exclusion = strings.Replace(exclusion, "，", ",", -1)
		for _, exclusionStr := range strings.Split(exclusion, ",") {
			rawSql += fmt.Sprintf(" and !find_in_set(?, labels)")
			sqlParams = append(sqlParams, exclusionStr)
		}
	}
	if search != "" {
		rawSql += " and concat (repo, title, sig) like ?"
		search = "%" + search + "%"
		sqlParams = append(sqlParams, search)
	}
	if order != "updated_at" {
		if direction == "asc" {
			rawSql += fmt.Sprintf(" order by created_at asc")
		} else {
			rawSql += fmt.Sprintf(" order by created_at desc")
		}
	} else {
		if direction == "asc" {
			rawSql += fmt.Sprintf(" order by updated_at asc")
		} else {
			rawSql += fmt.Sprintf(" order by updated_at desc")
		}
	}
	o := orm.NewOrm()
	countSql := strings.Replace(rawSql, "*", "count(*)", -1)
	var sqlCount int
	_ = o.Raw(countSql, sqlParams).QueryRow(&sqlCount)
	offset := perPage * (page - 1)
	rawSql += fmt.Sprintf(" limit ? offset ?")
	sqlParams = append(sqlParams, strconv.Itoa(perPage), strconv.Itoa(offset))
	return int64(sqlCount), rawSql, sqlParams
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
		Label:     c.GetString("label", ""),
		Exclusion: c.GetString("exclusion", ""),
		Search:    c.GetString("search", ""),
		Page:      page,
		PerPage:   perPage,
	}
	validate := validator.New()
	if validateErr := validate.Struct(qp); validateErr != nil {
		c.ApiJsonReturn("Invalid params", http.StatusBadRequest, nil)
	}
	count, sql, sqlParams := formQueryPullSql(qp)
	o := orm.NewOrm()
	if _, err := o.Raw(sql, sqlParams).QueryRows(&pull); err == nil {
		c.ApiDataReturn(count, page, perPage, pull)
	} else {
		c.ApiJsonReturn("Query error", http.StatusBadRequest, nil)
	}
}

type PullsSigsController struct {
	BaseController
}

type PullsSigsParams struct {
	KeyWord string `validate:"max=50"`
}

func (c *PullsSigsController) Get() {
	var pull []models.Pull
	keyWord := c.GetString("keyword", "")
	keyWord = utils.CheckParams(keyWord)
	params := PullsSigsParams{
		KeyWord: keyWord,
	}
	validate := validator.New()
	if validateErr := validate.Struct(params); validateErr != nil {
		c.ApiJsonReturn("Invalid params", http.StatusBadRequest, nil)
	}
	o := orm.NewOrm()
	sql := "select distinct sig from pull where sig != 'Private' order by sig"
	if _, err := o.Raw(sql).QueryRows(&pull); err != nil {
		c.ApiJsonReturn("Query error", http.StatusBadRequest, nil)
	}
	res := make([]string, 0)
	for _, i := range pull {
		res = append(res, i.Sig)
	}
	if keyWord == "" {
		c.ApiJsonReturn("Success", http.StatusOK, res)
	} else {
		res2 := make([]string, 0)
		for _, j := range res {
			if strings.Contains(strings.ToLower(j), strings.ToLower(keyWord)) {
				res2 = append(res2, j)
			}
		}
		c.ApiJsonReturn("Success", http.StatusOK, res2)
	}
}

type PullsReposController struct {
	BaseController
}

type PullsReposParams struct {
	KeyWord string `validate:"max=100"`
	Page    int    `validate:"min=1,max=1000000"`
	PerPage int    `validate:"min=1,max=100"`
	Sig     string `validate:"max=100"`
}

func (c *PullsReposController) Get() {
	var pull []models.Pull
	var pull2 []models.Pull
	page, _ := c.GetInt("page", 1)
	perPage, _ := c.GetInt("per_page", 10)
	offset := perPage * (page - 1)
	sig := c.GetString("sig", "")
	keyWord := c.GetString("keyword", "")
	sig = utils.CheckParams(sig)
	keyWord = utils.CheckParams(keyWord)
	params := PullsReposParams{
		KeyWord: keyWord,
		Page:    page,
		PerPage: perPage,
		Sig:     sig,
	}
	validate := validator.New()
	if validateErr := validate.Struct(params); validateErr != nil {
		c.ApiJsonReturn("Invalid params", http.StatusBadRequest, nil)
	}
	o := orm.NewOrm()
	sql := ""
	sqlParams := make([]string, 0, 0)
	if sig == "" {
		sql = "select distinct repo from pull where sig != 'Private' order by repo"
	} else {
		sql = "select distinct repo from pull where sig != 'Private' and sig = ? order by repo"
		sqlParams = append(sqlParams, sig)
	}
	count, err := o.Raw(sql, sqlParams).QueryRows(&pull)
	if err != nil {
		c.ApiJsonReturn("Query error", http.StatusBadRequest, nil)
	}
	separateSql := sql + fmt.Sprintf(" limit ? offset ?")
	sqlParams = append(sqlParams, strconv.Itoa(perPage), strconv.Itoa(offset))
	if _, err = o.Raw(separateSql, sqlParams).QueryRows(&pull2); err != nil {
		c.ApiJsonReturn("Query error", http.StatusBadRequest, nil)
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
	offset := perPage * (page - 1)
	keyWord := c.GetString("keyword", "")
	keyWord = utils.CheckParams(keyWord)
	params := CommonParams{
		KeyWord: keyWord,
		Page:    page,
		PerPage: perPage,
	}
	validate := validator.New()
	if validateErr := validate.Struct(params); validateErr != nil {
		c.ApiJsonReturn("Invalid params", http.StatusBadRequest, nil)
	}
	o := orm.NewOrm()
	sql := "select distinct ref from pull where sig != 'Private' order by ref"
	count, err := o.Raw(sql).QueryRows(&pull)
	if err != nil {
		c.ApiJsonReturn("Query error", http.StatusBadRequest, nil)
	}
	separateSql := sql + fmt.Sprintf(" limit ? offset ?")
	if _, err = o.Raw(separateSql, perPage, offset).QueryRows(&pull2); err != nil {
		c.ApiJsonReturn("Query error", http.StatusBadRequest, nil)
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
	params := CommonParams{
		KeyWord: keyWord,
		Page:    page,
		PerPage: perPage,
	}
	validate := validator.New()
	if validateErr := validate.Struct(params); validateErr != nil {
		c.ApiJsonReturn("Invalid params", http.StatusBadRequest, nil)
	}
	offset := perPage * (page - 1)
	o := orm.NewOrm()
	sql := "select distinct author from pull where sig != 'Private' order by author"
	count, err := o.Raw(sql).QueryRows(&pull)
	if err != nil {
		c.ApiJsonReturn("Query error", http.StatusBadRequest, nil)
	}
	separateSql := sql + fmt.Sprintf(" limit ? offset ?")
	if _, err = o.Raw(separateSql, perPage, offset).QueryRows(&pull2); err != nil {
		c.ApiJsonReturn("Query error", http.StatusBadRequest, nil)
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
	params := CommonParams{
		KeyWord: keyWord,
		Page:    page,
		PerPage: perPage,
	}
	validate := validator.New()
	if validateErr := validate.Struct(params); validateErr != nil {
		c.ApiJsonReturn("Invalid params", http.StatusBadRequest, nil)
	}
	offset := perPage * (page - 1)
	o := orm.NewOrm()
	sql := "select distinct assignees from pull where sig != 'Private'"
	if _, err := o.Raw(sql).QueryRows(&pull); err != nil {
		c.ApiJsonReturn("Query error", http.StatusBadRequest, nil)
	}
	res := make([]string, 0)
	for _, i := range pull {
		if i.Assignees == "" {
			continue
		}
		for _, j := range strings.Split(i.Assignees, ",") {
			if utils.InMap(utils.ConvertStrSlice2Map(res), j) {
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
	params := CommonParams{
		KeyWord: keyWord,
		Page:    page,
		PerPage: perPage,
	}
	validate := validator.New()
	if validateErr := validate.Struct(params); validateErr != nil {
		c.ApiJsonReturn("Invalid params", http.StatusBadRequest, nil)
	}
	offset := perPage * (page - 1)
	o := orm.NewOrm()
	sql := "select distinct labels from pull where sig != 'Private'"
	if _, err := o.Raw(sql).QueryRows(&pull); err != nil {
		c.ApiJsonReturn("Query error", http.StatusBadRequest, nil)
	}
	res := make([]string, 0)
	for _, i := range pull {
		if i.Labels == "" {
			continue
		}
		for _, j := range strings.Split(i.Labels, ",") {
			if utils.InMap(utils.ConvertStrSlice2Map(res), j) {
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
